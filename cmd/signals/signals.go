package main

/*
The Go program serves as an example for managing context and handling system signals.
At its core, it consists of three main routines: `main`, `handleSignals`, and `doSomething`.

1. The `main` function sets up a cancellable context and initiates the other two routines as goroutines.
2. `handleSignals` listens for system signals like `SIGTERM` and `SIGINT` to gracefully shut down the application by cancelling the context.
3. `doSomething` simulates a long-running task that can be interrupted based on the context's state.

The program utilizes a `WaitGroup` to ensure that all goroutines complete their execution before the program exits.
This example effectively demonstrates how to use context for cancellation and how to handle system signals for a graceful shutdown.
*/

// Import required packages
import (
	"context" // For managing timeouts and cancellation signals
	// For formatted I/O operations
	// For interacting with operating system functionalities
	// For signal handling
	"sync" // For synchronization primitives
	// For system call invocations
	"time" // For time operations

	"github.com/go-logr/logr"
	"github.com/tilseiffert/go-experiments/internal/logging"
	"github.com/tilseiffert/go-experiments/lib/signals"
)

// Entry point of the program
func main() {

	logger := logging.CreateLogger()

	// Log the start of the main function
	logger.Info("Hello 👋")
	defer logger.Info("Bye 👋")

	// Create a new context with cancellation capabilities
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize a WaitGroup for managing goroutines
	var wg sync.WaitGroup

	// Defer the cancel function to ensure resources are freed
	defer func() {
		logger.Info("deferred cleanup, trigger cancel of context")
		cancel() // Cancel the context
	}()

	// Start the signal handling function
	logger.Info("Calling handleSignals()")
	signals.HandleSignals(ctx, cancel, &logger)

	// Start the doSomething function as a goroutine
	logger.Info("Calling doSomething()")
	wg.Add(1) // Increment the WaitGroup counter
	go doSomething(ctx, &wg, &logger)

	// Wait for the context to be done (e.g., cancelled)
	logger.Info("Waiting for context to be done...")
	<-ctx.Done()

	// Wait for all goroutines to finish
	logger.Info("Context done. Waiting for goroutines to finish...")
	wg.Wait()

}

// doSomething simulates a long-running task
func doSomething(ctx context.Context, wg *sync.WaitGroup, logger *logr.Logger) {

	logger = logging.LoggerAddName(logger, "doSomething")

	logger.V(1).Info("doSomething started.")

	// Decrement the WaitGroup counter when the function exits
	defer wg.Done()
	defer logger.V(1).Info("doSomething exited.")

	// Infinite loop to simulate work
	for {
		select {
		case <-ctx.Done():
			// Log and exit if the context is done
			logger.Info("doSomething received context done signal.")
			return
		default:
			// Log and continue working
			logger.V(2).Info("doSomething running...")
			time.Sleep(1 * time.Second)
		}
	}
}
