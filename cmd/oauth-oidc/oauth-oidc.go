// help: quickly load env via command-line: `export $(grep -v '^#' cmd/oauth-oidc/config.env | xargs)` (https://stackoverflow.com/a/20909045)

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/exp/slog"

	"github.com/oklog/ulid/v2"
	zitadel_logging "github.com/zitadel/logging"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

// OAuth2 and OIDC example application; following https://github.com/zitadel/oidc/blob/main/example/client/app/app.go

//nolint:gochecknoglobals
var (
	VERSION = "undefined"                 // version
	BUILD   = "undefined"                 // build datetime
	NAME    = "go-experiments/oauth-oidc" // application name
)

// Global variables to hold configuration settings.
//
//nolint:gochecknoglobals
var (
	DefaultPort         = 8080
	DefaultLoglevel     = "info"
	DefaultCallbackPath = "/auth/callback"
	ENV_LOGLEVEL        = "LOG_LEVEL"
	ENV_PRETTYPRINT     = "LOG_PRETTY"
)

func main() {

	logger, bye := initLogger()
	defer bye()

	zitadel_logger := slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}),
	)

	port := DefaultPort
	callbackPath := DefaultCallbackPath

	issuer := os.Getenv("ISSUER")
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	scopes := strings.Split(os.Getenv("SCOPES"), " ")

	redirectURI := fmt.Sprintf("http://localhost:%v%v", port, callbackPath)

	// CHECK ENV

	if issuer == "" {
		logger.Fatal().Msg("ENV 'ISSUER' empty or not set")
	}

	if clientID == "" {
		logger.Fatal().Msg("ENV 'CLIENT_ID' empty or not set")
	}

	if len(scopes) == 0 {
		logger.Warn().Msg("ENV 'SCOPES' empty or not set")
	}

	// INIT

	key, err := GenerateRandomString(32)

	if err != nil {
		logger.Panic().Err(err).Msg("Failed to generate random string")
	}

	logger.Trace().Str("key", key).Msg("Generated random string for cookie encryption")

	cookieHandler := httphelper.NewCookieHandler([]byte(key), []byte(key), httphelper.WithUnsecure())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpclient := &http.Client{
		Timeout: time.Minute,
	}

	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
		rp.WithHTTPClient(httpclient),
	}

	if clientSecret == "" {
		options = append(options, rp.WithPKCE(cookieHandler))
	}

	provider, err := rp.NewRelyingPartyOIDC(ctx, issuer, clientID, clientSecret, redirectURI, scopes, options...)

	if err != nil {
		logger.Panic().Err(err).Msg("Failed to create OIDC provider")
	}

	// generate some state (representing the state of the user in your application,
	// e.g. the page where he was before sending him to login
	state := func() string {
		return ulid.Make().String()
	}

	// register the AuthURLHandler at your preferred path.
	// the AuthURLHandler creates the auth request and redirects the user to the auth server.
	// including state handling with secure cookie and the possibility to use PKCE.
	// Prompts can optionally be set to inform the server of
	// any messages that need to be prompted back to the user.
	http.Handle("/login", rp.AuthURLHandler(state, provider, rp.WithPromptURLParam("Welcome back!")))

	// for demonstration purposes the returned userinfo response is written as JSON object onto response
	marshalUserinfo := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *oidc.UserInfo) {
		data, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}

	// register the CodeExchangeHandler at the callbackPath
	// the CodeExchangeHandler handles the auth response, creates the token request and calls the callback function
	// with the returned tokens from the token endpoint
	// in this example the callback function itself is wrapped by the UserinfoCallback which
	// will call the Userinfo endpoint, check the sub and pass the info into the callback function
	http.Handle(callbackPath, rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), provider))

	// simple counter for request IDs
	var counter atomic.Int64
	// enable incomming request logging
	mw := zitadel_logging.Middleware(
		zitadel_logging.WithLogger(zitadel_logger),
		zitadel_logging.WithGroup("server"),
		zitadel_logging.WithIDFunc(
			func() slog.Attr {
				return slog.Int64("id", counter.Add(1))
			}),
	)

	lis := fmt.Sprintf("127.0.0.1:%d", port)
	logger.Info().Msgf("server listening, press ctrl+c to stop, addr: %v", lis)
	err = http.ListenAndServe(lis, mw(http.DefaultServeMux))
	if err != http.ErrServerClosed {
		logger.Error().Err(err).Msg("server terminated")
		os.Exit(1)
	}
}
