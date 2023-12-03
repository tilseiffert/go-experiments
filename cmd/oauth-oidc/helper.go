package main

import (
	"crypto/rand"
	"math/big"
	"os"

	"github.com/rs/zerolog"
)

// initLogger sets up the application-wide logger.
// returns:
//   - logger: A zerolog.Logger instance configured to use Zerolog.
//   - bye: A function to be called at the end of the application (e.g. via defer) to send goodbye message.
func initLogger() (zerolog.Logger, func()) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Fetch and set the log level from environment variables.
	logLevel := os.Getenv(ENV_LOGLEVEL)
	if logLevel == "" {
		logLevel = DefaultLoglevel
	}
	lvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}
	zerolog.SetGlobalLevel(lvl)

	// Create the logger with a timestamp.
	logger := zerolog.New(os.Stdout).With().Timestamp().
		Str("version", VERSION).Str("application", NAME).
		Logger()

	// Enable pretty printing if specified in environment variables.
	if os.Getenv(ENV_PRETTYPRINT) == "true" {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Caller().Logger()
	}

	// Add hostname to logger context.
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	logger = logger.With().Str("hostname", hostname).Logger()

	logger.Debug().Str("log-level", logger.GetLevel().String()).Str("build", BUILD).Msg("Hello 👋")

	return logger, func() {
		logger.Debug().Str("log-level", logger.GetLevel().String()).Str("build", BUILD).Msg("Bye 👋")
	}
}

// GenerateRandomString generates a random string of length n using crypto/rand.
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	random := make([]byte, n)
	letters_len := big.NewInt(int64(len(letters)))
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, letters_len)
		if err != nil {
			return "", err
		}
		random[i] = letters[num.Int64()]
	}

	return string(random), nil
}
