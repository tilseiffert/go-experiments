package main

// Import required packages
import (
	"context"   // For managing timeouts and cancellation signals
	"fmt"       // For formatted I/O operations
	"os"        // For interacting with operating system functionalities
	"os/signal" // For signal handling
	"sync"      // For synchronization primitives
	"syscall"   // For system call invocations
	"time"      // For time operations
)

// log function prints a formatted log message with the current time and function name
func log(msg string, function string) {
	fmt.Printf("%s: %s() %s\n", time.Now().Format(time.RFC3339), function, msg)
}

// Entry point of the program
func main() {
	// Log the start of the main function
	log("Hello 👋", "main")
	defer log("Bye 👋", "main")

	// Create a new context with cancellation capabilities
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize a WaitGroup for managing goroutines
	var wg sync.WaitGroup

	// Defer the cancel function to ensure resources are freed
	defer func() {
		log("deferred cleanup, trigger cancel of context", "main")
		cancel() // Cancel the context
	}()

	// Start the signal handling function
	log("Calling handleSignals()", "main")
	handleSignals(ctx, cancel)

	// Start the doSomething function as a goroutine
	log("Calling doSomething()", "main")
	wg.Add(1) // Increment the WaitGroup counter
	go doSomething(ctx, &wg)

	// Wait for the context to be done (e.g., cancelled)
	log("Waiting for context to be done...", "main")
	<-ctx.Done()

	// Wait for all goroutines to finish
	log("Waiting for goroutines to finish...", "main")
	wg.Wait()

}

// handleSignals listens for OS signals and cancels the context if an interrupt is received
func handleSignals(ctx context.Context, cancel context.CancelFunc) {
	log("handleSignals started.", "handleSignals")

	// Create a channel to receive OS signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to handle incoming signals
	go func() {
		select {
		case sig := <-c:
			// Log and cancel the context if an interrupt signal is received
			log(fmt.Sprintf("Received signal: %s. Initiating shutdown...", sig), "handleSignals")
			cancel()
		case <-ctx.Done():
			// Exit if the context is done
			log("Context done. Exiting signal handler.", "handleSignals")
		}
	}()
}

// doSomething simulates a long-running task
func doSomething(ctx context.Context, wg *sync.WaitGroup) {
	log("doSomething started.", "doSomething")

	// Decrement the WaitGroup counter when the function exits
	defer wg.Done()
	defer log("doSomething exited.", "doSomething")

	// Infinite loop to simulate work
	for {
		select {
		case <-ctx.Done():
			// Log and exit if the context is done
			log("doSomething received context done signal.", "doSomething")
			return
		default:
			// Log and continue working
			log("doSomething running...", "doSomething")
			time.Sleep(1 * time.Second)
		}
	}
}
