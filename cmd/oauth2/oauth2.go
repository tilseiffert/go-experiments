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
}

// handleAuthCallback handles the OAuth2 callback
func (s *session) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Implementiere den Callback für OAuth2
	log.Println("### Callback called...")

	code := r.URL.Query().Get("code")
	if code == "" {
		// Fehlerbehandlung
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Code not found in request"))

		if err != nil {
			log.Println("Error writing response:", err)
		}
	}

	log.Println("Code:", code)

	ctx := context.Background()

	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		// Fehlerbehandlung
		log.Fatalln("token exchange failed:", err)

		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Token exchange failed"))

		if err != nil {
			log.Println("Error writing response:", err)
		}
	}

	log.Println("Token:", token)

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		// Fehlerbehandlung
		log.Fatalln("id_token not found in token")

		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("id_token not found in token"))

		if err != nil {
			log.Println("Error writing response:", err)
		}
	}

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	idTokenObj, err := s.provider.Verifier(oidcConfig).Verify(ctx, idToken)
	if err != nil {
		// Fehlerbehandlung
		return
	}

	// Nutzerdaten extrahieren
	var claims struct {
		Email string `json:"email"`
		// weitere Felder
	}
	if err := idTokenObj.Claims(&claims); err != nil {
		// Fehlerbehandlung
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Hello " + claims.Email))

	if err != nil {
		log.Println("Error writing response:", err)
	}
}

func main() {

	log.Println("Starting webserver...")
	defer log.Println("Stopping webserver...")

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, ProviderURL)
	if err != nil {
		panic(err)
	}

	sess := session{
		provider: provider,
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  "http://localhost:" + fmt.Sprint(DefaultPort) + "/callback",
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
	}

	// config := oauth2.Config{
	// 	ClientID:     clientID,
	// 	ClientSecret: clientSecret,
	// 	Endpoint:     provider.Endpoint(),
	// 	RedirectURL:  "http://localhost:" + fmt.Sprint(DefaultPort) + "/callback",
	// 	Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	// }

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, sess.config.AuthCodeURL(""), http.StatusFound)
		log.Println("Redirecting to auth endpoint...")
	})

	http.HandleFunc("/callback", sess.handleAuthCallback)

	// Start listening for HTTP requests
	log.Printf("Listening on http://localhost:%d/\n", DefaultPort)
	log.Fatal(http.ListenAndServe(":"+fmt.Sprint(DefaultPort), nil))
}

// http://localhost:8080/callback
//	?session_state=38ae7cba-1588-4671-83e0-4d123cd90bf6
//	&code=0444e102-a71d-47ef-857a-8e6baa25e6fe.38ae7cba-1588-4671-83e0-4d123cd90bf6.228ac7b3-ceb7-4106-8a74-56d6708325c8
