package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/go-logr/logr"
	"github.com/tilseiffert/go-experiments/internal/logging"
	"golang.org/x/oauth2"
)

var (
	ErrAuthCallbackError  = errors.New("auth callback returned error")
	ErrAuthRedirectToAuth = errors.New("redirecting to auth endpoint")
	CookieNameToken       = "auth-token"
)

type AuthProvider struct {
	Provider *oidc.Provider
	Config   oauth2.Config
}

func (a *AuthProvider) validateToken(ctx context.Context, token *oauth2.Token, logger *logr.Logger) error {

	logger = logging.LoggerAddName(logger, "validateToken")
	logger.V(logging.LvlTrace).Info("validateToken called...")

	_, err := a.Provider.UserInfo(ctx, oauth2.StaticTokenSource(token))

	return err
}

// CheckAuth checks if the user is authenticated and returns the token if so.
// If the user is not authenticated, it will redirect to the auth endpoint.
// If the user is authenticated but the token is invalid, it will return an error.
//
// Returns:
//   - token: The token if the user is authenticated
//   - error: An error if the user is not authenticated or the token is invalid
//   - bool: if a redirect to the auth endpoint was performed
func CheckAuth(a *AuthProvider, w http.ResponseWriter, r *http.Request, logger *logr.Logger) (string, error, bool) {

	logger = logging.LoggerAddName(logger, "checkAuth")
	logger.V(logging.LvlTrace).Info("checkAuth called...")

	// check for existing token in cookie
	cookie, err := r.Cookie(CookieNameToken)

	if err == nil {
		// token found in cookie
		logger.V(logging.LvlTrace).Info("token found in cookie '" + CookieNameToken + "#")
		err = a.validateToken(r.Context(), &oauth2.Token{AccessToken: cookie.Value}, logger)
		return cookie.Value, err, false
	}

	// check if ErrNoCookie
	if err != http.ErrNoCookie {
		// some other error occured
		logger.V(logging.LvlInfo).Info("Unexpected error reading auth-token cookie: " + err.Error())
		return "", err, false
	}

	// check for callback

	code := r.URL.Query().Get("code")

	if code == "" {

		error := r.URL.Query().Get("error")

		// check for error
		if error != "" {
			logger.V(logging.LvlInfo).Info("Error occured on auth callback", "error", error, "error-message", r.URL.Query().Get("error_description"))
			return "", fmt.Errorf("Got error in callback '%s' wiht message '%s': %w", error, r.URL.Query().Get("error_description"), ErrAuthCallbackError), false
		}

		http.Redirect(w, r, a.Config.AuthCodeURL(""), http.StatusFound)
		logger.V(logging.LvlDebug).Info("Redirecting to auth endpoint...")

		return "", ErrAuthRedirectToAuth, true
	}

	logger.V(logging.LvlTrace + 3).Info("Code found in request: " + code)

	// exchange code for token
	token, err := a.Config.Exchange(r.Context(), code)
	if err != nil {
		logger.V(logging.LvlInfo).Info("Error exchanging code for token: " + err.Error())
		return "", fmt.Errorf("Unexpected error exchanging code for token: %w", err), false
	}

	logger.V(logging.LvlTrace + 3).Info("Token recieved from CodeExchange")

	// validate token
	// _, err = a.Provider.UserInfo(r.Context(), oauth2.StaticTokenSource(token))
	err = a.validateToken(r.Context(), token, logger)

	if err != nil {
		logger.V(logging.LvlInfo).Info("Unexpected error validating token: " + err.Error())
		return "", err, false
	}

	// set cookie
	http.SetCookie(w, &http.Cookie{
		Name:  CookieNameToken,
		Value: token.AccessToken,
	})

	return token.AccessToken, nil, false
}
