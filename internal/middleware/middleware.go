package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
)

type MiddlewareOption func(*middleware)

type middleware struct {
	logger  *zerolog.Logger
	handler http.Handler
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// logger with log to std null
	logger := zerolog.New(nil)

	if m.logger != nil {
		logger = m.logger.With().
			Str("request", ulid.Make().String()).
			Logger()
	}

	logger.Trace().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Str("remote", r.RemoteAddr).
		Msg("request received")

	/* handle actual request */
	m.handler.ServeHTTP(w, r)

	logger.Info().Str("duration-seconds", fmt.Sprint(time.Since(start).Seconds())).Msg("request served")
}
