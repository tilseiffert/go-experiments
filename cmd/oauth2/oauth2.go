package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/go-logr/logr"
	"github.com/tilseiffert/go-experiments/internal/logging"
	"github.com/tilseiffert/go-experiments/lib/auth"
	"github.com/tilseiffert/go-experiments/lib/signals"
	"golang.org/x/oauth2"
)

// Constants for configuring OAuth2 and OIDC
const (
	DefaultPort  = 8080
	ProviderURL  = "https://auth.tils-hub.de/realms/go-experiments"
	clientID     = "go-experiments-general"
	clientSecret = "OP09JUCn0NaP1pyIRbyDKAT8RVe0Xhlp"
	authPath     = "/auth"
)

var (
	ErrAuthCallbackError  = errors.New("auth callback returned error")
	ErrAuthRedirectToAuth = errors.New("redirecting to auth endpoint")
)

// session struct holds OAuth2 and OIDC configuration
type session struct {
	config   oauth2.Config  // deprecated
	provider *oidc.Provider // deprecated
	ctx      context.Context
	logger   logr.Logger
	auth     auth.AuthProvider
}

// Spezialisiertes Struct für Claims mit Gruppeninformationen
type customClaims struct {
	Email          string                 `json:"email"`
	RealmAccess    map[string]interface{} `json:"realm_access"`
	ResourceAccess map[string]interface{} `json:"resource_access"`
}

func printDebugLine() {
	fmt.Printf("\n\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")
}

func (s *session) handleRoot(w http.ResponseWriter, r *http.Request) {

	printDebugLine()

	logger := logging.LoggerAddName(&s.logger, "handleRoot")
	fmt.Printf("handleRoot called...")
	logger.V(logging.LvlTrace).Info("handleRoot called...")

	token, err, redirect := auth.CheckAuth(&s.auth, w, r, logger)

	if redirect {
		logger.V(logging.LvlInfo).Info("Redirect was performed, exiting handler...")
		return
	}

	if err != nil {
		logger.V(logging.LvlInfo).Info("Error checking auth: " + err.Error())
		http.Error(w, "Error checking auth: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if token == "" {
		logger.V(logging.LvlInfo).Info("No token found :(")
		http.Error(w, "No token found :(", http.StatusUnauthorized)
		return
	}

	logger.V(logging.LvlInfo).Info("Token recieved from auth.CheckAuth()")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf("Received Token: %v", token)

	if _, err := w.Write([]byte(response)); err != nil {
		logger.V(0).Info("Error writing response: %v", err)
	}

	// // Schritt 1: Authorization Code anfordern
	// http.Redirect(w, r, s.config.AuthCodeURL(""), http.StatusFound)
	// logger.V(2).Info("Redirecting to auth endpoint...")
}

// handleAuthCallback handles the OAuth2 callback
func (s *session) handleAuthCallback(w http.ResponseWriter, r *http.Request) {

	logger := logging.LoggerAddName(&s.logger, "handleAuthCallback")
	logger.V(3).Info("Callback called...")

	code := r.URL.Query().Get("code")
	if code == "" {

		error := r.URL.Query().Get("error")

		if error != "" {
			http.Error(w, "Code not found in request\nKeycloak error: "+error+"\nKeycloak error message: "+r.URL.Query().Get("error_description"), http.StatusBadRequest)
			logger.V(0).Info("Code not found in request", "Keycloak error", error, "Keycloak error message", r.URL.Query().Get("error_description"))
			return
		}

		http.Redirect(w, r, s.config.AuthCodeURL(""), http.StatusFound)
		logger.V(2).Info("Redirecting to auth endpoint...")

		return
	}

	logger.V(1).Info("Authorization Code: %s", code)

	// Schritt 2: Code gegen Token tauschen
	token, err := s.config.Exchange(s.ctx, code)
	if err != nil {
		http.Error(w, "Token exchange failed", http.StatusBadRequest)
		logger.V(0).Info("Token exchange failed: %v", err)
		return
	}
	logger.V(2).Info("Received Token: %v", token)

	// Schritt 3: id_token extrahieren und überprüfen
	idTokenStr, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "id_token not found in token", http.StatusBadRequest)
		logger.V(1).Info("id_token not found in token")
		return
	}

	oidcConfig := &oidc.Config{ClientID: clientID}
	idToken, err := s.provider.Verifier(oidcConfig).Verify(s.ctx, idTokenStr)
	if err != nil {
		http.Error(w, "ID token verification failed", http.StatusBadRequest)
		logger.V(1).Info("ID token verification failed: %v", err)
		return
	}

	// Schritt 4: Alle Claims extrahieren und anzeigen
	var allClaims map[string]interface{}
	if err := idToken.Claims(&allClaims); err != nil {
		http.Error(w, "Failed to extract claims", http.StatusBadRequest)
		logger.V(1).Info("Failed to extract claims: %v", err)
		return
	}

	claimsJSON, err := json.MarshalIndent(allClaims, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal claims", http.StatusInternalServerError)
		logger.V(1).Info("Failed to marshal claims: %v", err)
		return
	}

	logger.V(3).Info("All Claims: %s", claimsJSON)

	var claims customClaims
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "Failed to extract claims", http.StatusBadRequest)
		logger.V(1).Info("Failed to extract claims: %v", err)
		return
	}

	// Gruppeninformationen aus RealmAccess oder ResourceAccess extrahieren
	var groups []string
	if roles, found := claims.RealmAccess["roles"]; found {
		groups = append(groups, roles.([]string)...)
	}
	if clientRoles, found := claims.ResourceAccess["client_roles"]; found {
		groups = append(groups, clientRoles.([]string)...)
	}

	logger.V(1).Info("User Email: " + claims.Email)
	logger.V(1).Info("User Groups: " + fmt.Sprint(groups))

	// Schritt 5: Antwort an den Client senden
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf("Hello %s, you belong to these groups: %v\n\n Your claims:\n%s", claims.Email, groups, string(claimsJSON))
	if _, err := w.Write([]byte(response)); err != nil {
		logger.V(0).Info("Error writing response: %v", err)
	}
}

// main is the entry point of the application
func main() {
	logger := logging.CreateLogger().WithName("main")

	logger.V(logging.LvlTrace).Info("Hello 👋")
	defer logger.V(logging.LvlTrace).Info("Stopping webserver...")

	// Create a new context with cancellation capabilities
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		logger.V(logging.LvlTrace).Info("deferred cleanup, trigger cancel of context")
		cancel() // Cancel the context
	}()

	// Start the signal handling function
	logger.V(logging.LvlTrace + 1).Info("Calling handleSignals()")
	signals.HandleSignals(ctx, cancel, &logger)

	// Initialize OIDC provider
	provider, err := oidc.NewProvider(ctx, ProviderURL)
	if err != nil {
		panic(err)
	}

	// Initialize session with OAuth2 and OIDC configuration
	sess := session{
		ctx:      ctx,
		provider: provider,
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  "http://localhost:" + fmt.Sprint(DefaultPort) + authPath,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		logger: logger.WithName("session"),
		auth: auth.AuthProvider{
			Provider: provider,
			Config: oauth2.Config{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				Endpoint:     provider.Endpoint(),
				RedirectURL:  "http://localhost:" + fmt.Sprint(DefaultPort),
				Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			},
		},
	}

	// // Define route for OAuth2 authorization
	// http.HandleFunc(authPath, func(w http.ResponseWriter, r *http.Request) {
	// 	http.Redirect(w, r, sess.config.AuthCodeURL(""), http.StatusFound)
	// 	logger.V(2).Info("Redirecting to auth endpoint...")
	// })

	http.HandleFunc("/", sess.handleRoot)
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	http.Redirect(w, r, authPath, http.StatusFound)
	// 	logger.V(logging.LvlInfo).Info("Redirecting to auth endpoint...")
	// })

	// Define route for OAuth2 callback
	http.HandleFunc(authPath, sess.handleAuthCallback)

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		printDebugLine()
		logger.V(logging.LvlTrace).Info("favicon called...")
		http.Error(w, "No favicon here", http.StatusNotFound)
		logger.V(logging.LvlInfo).Info("No favicon here")
	})

	// // Start the web server
	// panic(http.ListenAndServe(":"+fmt.Sprint(DefaultPort), nil))

	srv := &http.Server{
		Addr:    ":" + fmt.Sprint(DefaultPort),
		Handler: nil, // Use http.DefaultServeMux
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	logger.V(logging.LvlInfo).Info(fmt.Sprintf("Listening on http://localhost:%d/", DefaultPort))

	// Wait for a signal to shutdown
	<-ctx.Done()

	logger.V(logging.LvlDebug).Info("Shutting down the server")
	ctxShutDown, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		logger.V(logging.LvlInfo).Info("Server Shutdown Failed:", err)
	}
}
