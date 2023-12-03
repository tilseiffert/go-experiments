// Package logging provides utilities for logging.
// It uses the logr interface and Zerolog as the logging backend.
package logging

import (
	"os"

	"github.com/go-logr/logr"     // Import logr interface for logging
	"github.com/go-logr/zerologr" // Import Zerologr adapter for logr
	"github.com/rs/zerolog"       // Import Zerolog for logging
)

const (
	LvlInfo  = 0
	LvlDebug = 1
	LvlTrace = 2
)

// CreateLogger initializes and returns a new logger based on the logr interface.
// It configures Zerolog as the backend for logging.
//
// Logging Levels:
// The function sets up mapping between logr V-levels and Zerolog levels.
//   - logr.V(0) maps to zerolog.InfoLevel
//   - logr.V(1) maps to zerolog.DebugLevel
//   - logr.V(2) maps to zerolog.TraceLevel
//
// Notes:
//   - V-levels higher than 2 can be used in logr but may not have Zerolog features like Hooks and Sampling.
//   - V-level values are only logged when using logr's Info() method, not Error().
//
// Returns:
//   - A logr.Logger instance configured to use Zerolog.
func CreateLogger() logr.Logger {
	// Customize the field name and separator for loggers created by Zerologr.
	zerologr.NameFieldName = "logger"
	zerologr.NameSeparator = "/"

	// Set the maximum V-level for logging with Zerologr.
	// V-levels higher than this won't be logged.
	zerologr.SetMaxV(1)

	// Create and return a new logr.Logger instance using Zerolog as the backend.
	return zerologr.New(CreateLoggerZerolog())
}

// CreateLoggerZerolog initializes and returns a new Zerolog logger.
//
// Returns:
// A pointer to a configured zerolog.Logger instance.
func CreateLoggerZerolog() *zerolog.Logger {
	// Set the time field format for Zerolog to Unix time in milliseconds.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	// Set the global logging level for Zerolog to Trace.
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	zerolog.SetGlobalLevel(-100)

	// Create a new Zerolog logger instance with a console writer that outputs to STDERR.
	zl := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})

	// Enhance the logger by adding caller and timestamp fields.
	zl = zl.With().Caller().Timestamp().Logger()

	// Return the enhanced Zerolog logger.
	return &zl
}

// LoggerAddName adds a name to a logr.Logger instance.
// If the logger is nil, a new logger is created that discards all messages.
func LoggerAddName(logger *logr.Logger, name string) *logr.Logger {

	// If logger is nil, create a new logger that discards all messages.
	if logger == nil {
		l := logr.Discard()
		return &l
	}

	// Add the name to the logger.
	l := logger.WithName(name)
	return &l
}
