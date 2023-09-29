# go-experiments


## [signals.go](cmd/signals/signals.go)

The Go program serves as an example for managing context and handling system signals. 
At its core, it consists of three main routines: `main`, `handleSignals`, and `doSomething`.

1. The `main` function sets up a cancellable context and initiates the other two routines as goroutines.
2. `handleSignals` listens for system signals like `SIGTERM` and `SIGINT` to gracefully shut down the application by cancelling the context.
3. `doSomething` simulates a long-running task that can be interrupted based on the context's state.

The program utilizes a `WaitGroup` to ensure that all goroutines complete their execution before the program exits. 
This example effectively demonstrates how to use context for cancellation and how to handle system signals for a graceful shutdown.
