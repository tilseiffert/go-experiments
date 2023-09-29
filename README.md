# go-experiments

Overview:
- [signals.go](#signalsgo)
- [Dockerfile.TEMPLATE](#dockerfiletemplate)


## [signals.go](cmd/signals/signals.go)


The Go code demonstrates how to manage context and system signals in a concurrent application. It shows how to initiate, monitor, and gracefully shut down long-running tasks using context and signal handling. At its core, it consists of three main routines: `main`, `handleSignals`, and `doSomething`.

1. The `main` function sets up a cancellable context and initiates the other two routines as goroutines.
2. `handleSignals` listens for system signals like `SIGTERM` and `SIGINT` to gracefully shut down the application by cancelling the context.
3. `doSomething` simulates a long-running task that can be interrupted based on the context's state.

### Core Functionalities:

1. **Context Management:**
    
    - Creates a cancellable context to manage the lifecycle of goroutines.
    - Example: `ctx, cancel := context.WithCancel(context.Background())`

2. **Signal Handling:**
    
    - Listens for system signals like `SIGTERM` and `SIGINT` and cancels the context when received.
    - Example: `signal.Notify(c, os.Interrupt, syscall.SIGTERM)`

3. **Concurrent Execution:**
    
    - Uses goroutines to run functions concurrently.
    - Manages goroutines using a `WaitGroup`.
    - Example: `go doSomething(ctx, &wg)`

4. **Graceful Shutdown:**
    
    - Waits for all goroutines to finish before exiting the program.
    - Example: `wg.Wait()`

5. **Logging:**
    
    - Provides a logging function to print time-stamped messages.
    - Example: `log("Hello 👋", "main")`

6. **Resource Cleanup:**
    
    - Uses `defer` to ensure that resources like context are properly released.
    - Example: `defer cancel()`

### Use-Case Example:

This code would be applicable in a server application where you have multiple long-running tasks that need to be managed and could be interrupted by system signals. For instance, a web scraper that runs multiple scraping tasks in parallel but needs to shut down gracefully when receiving a termination signal.



## [Dockerfile.TEMPLATE](Dockerfile.TEMPLATE)


The Dockerfile is designed to build a secure and optimized containerized Go application. It employs a multi-stage build process to separate the build environment from the runtime environment, enhancing security and reducing the final image size.

### Core Functionalities:

1. **Multi-Stage Build:**
    
    - Utilizes two stages: `builder` for building the Go application and `FINAL_IMAGE` for the runtime.
    - Example: `FROM ${BUILD_IMAGE} AS builder` and `FROM ${FINAL_IMAGE}`

2. **Argument Definitions:**
    
    - Defines build arguments for base images and source directory.
    - Example: `ARG BUILD_IMAGE=golang:1.21@sha256:...`

3. **Dependency Management:**
    
    - Copies `go.mod` and `go.sum` files and downloads dependencies.
    - Caches dependencies unless `go.mod` or `go.sum` changes.
    - Example: `COPY go.mod go.sum ./` and `RUN go mod download && go mod verify`

4. **Source Code Copy:**
    
    - Copies the entire source code into the container.
    - Example: `COPY . .`

5. **Optimized Compilation:**
    
    - Compiles the Go application with flags for optimization and static linking.
    - Example: `RUN GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags "-static"' -a -v -o /app ./cmd/signals/signals.go`

6. **Security Measures:**
    
    - Uses a distroless image for the final stage.
    - Runs the application as a non-root user.
    - Example: `USER nobody:nobody` and `CMD ["app"]`

### Use-Case Example:

This Dockerfile template would be particularly useful for deploying Go applications in a Kubernetes cluster where security and optimization are critical. The multi-stage build ensures that only the necessary components are included in the final image, and running as a non-root user adds an extra layer of security.
