package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/tilseiffert/go-experiments/internal/logging"
)

// HandleSignals listens for OS signals and cancels the context if an interrupt is received.
// logger may be nil.
func HandleSignals(ctx context.Context, cancel context.CancelFunc, logger *logr.Logger) {

	logger = logging.LoggerAddName(logger, "HandleSignals")

	// Create a channel to receive OS signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to handle incoming signals
	go func() {
		select {
		case sig := <-c:
			// Log and cancel the context if an interrupt signal is received
			logger.Info("Received signal. Initiating shutdown...", "signal", sig)
			cancel()
		case <-ctx.Done():
			// Exit if the context is done
			logger.V(logging.LvlTrace + 1).Info("Context done. Exiting signal handler.")
		}
	}()

	logger.V(logging.LvlTrace).Info("Signal handler set up.")

}
