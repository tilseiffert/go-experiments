package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

// Constants for configuring OAuth2 and OIDC
const (
	DefaultPort  = 8080
	ProviderURL  = "https://auth.tils-hub.de/realms/go-experiments"
	clientID     = "go-experiments-general"
	clientSecret = "OP09JUCn0NaP1pyIRbyDKAT8RVe0Xhlp"
)

// session struct holds OAuth2 and OIDC configuration
type session struct {
	config   oauth2.Config
	provider *oidc.Provider
	ctx      context.Context
}

// handleAuthCallback handles the OAuth2 callback
func (s *session) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("### Callback called...")

	// Extract authorization code from URL query
	code := r.URL.Query().Get("code")
	if code == "" {
		// Handle missing code
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Code not found in request"))
		if err != nil {
			log.Println("Error writing response:", err)
		}
		return
	}

	log.Println("Code:", code)

	// Perform OAuth2 code exchange
	token, err := s.config.Exchange(s.ctx, code)
	if err != nil {
		// Handle exchange error
		log.Fatalln("token exchange failed:", err)
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Token exchange failed"))
		if err != nil {
			log.Println("Error writing response:", err)
		}
		return
	}

	log.Println("Token:", token)

	// Extract id_token from the OAuth2 token
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		// Handle missing id_token
		log.Fatalln("id_token not found in token")
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("id_token not found in token"))
		if err != nil {
			log.Println("Error writing response:", err)
		}
		return
	}

	// Verify id_token
	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	idTokenObj, err := s.provider.Verifier(oidcConfig).Verify(s.ctx, idToken)
	if err != nil {
		// Handle verification error
		return
	}

	// Extract user claims from id_token
	var claims struct {
		Email string `json:"email"`
		// Add more fields as needed
	}
	if err := idTokenObj.Claims(&claims); err != nil {
		// Handle claim extraction error
		return
	}

	// Respond with a greeting
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Hello " + claims.Email))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

// main is the entry point of the application
func main() {
	log.Println("Starting webserver...")
	defer log.Println("Stopping webserver...")

	// Create a background context
	ctx := context.Background()

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
			RedirectURL:  "http://localhost:" + fmt.Sprint(DefaultPort) + "/callback",
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
	}

	// Define route for OAuth2 authorization
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, sess.config.AuthCodeURL(""), http.StatusFound)
		log.Println("Redirecting to auth endpoint...")
	})

	// Define route for OAuth2 callback
	http.HandleFunc("/callback", sess.handleAuthCallback)

	// Start the web server
	log.Printf("Listening on http://localhost:%d/\n", DefaultPort)
	log.Fatal(http.ListenAndServe(":"+fmt.Sprint(DefaultPort), nil))
}
