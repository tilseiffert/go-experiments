# Define arguments for base images and source directory
ARG BUILD_IMAGE=golang:1.21@sha256:c416ceeec1cdf037b80baef1ccb402c230ab83a9134b34c0902c542eb4539c82
ARG FINAL_IMAGE=gcr.io/distroless/static@sha256:e7e79fb2947f38ce0fab6061733f7e1959c12b843079042fe13f56ca7b9d178c

# Build Stage: Using a specific version of the Go image with SHA256 for security and reproducibility
FROM ${BUILD_IMAGE} AS builder

# Set the working directory inside the container
WORKDIR /usr/src/app

# Copy go.mod and go.sum files and download dependencies
# This allows for caching and only re-downloads if these files change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the entire source code
COPY . .

# Compile the Go application with specific flags for optimization
# - GOOS=linux: Sets the target operating system to Linux
# - GOARCH=amd64: Sets the target architecture to AMD64
# - -ldflags: Allows you to set linker options
#   - '-w': Disables DWARF debugging information to reduce size
#   - '-s': Disables generation of the Go symbol table to reduce size
#   - '-extldflags "-static"': Makes the output binary statically linked
# - -a: Forces rebuilding of packages that are already up-to-date
# - -v: Enables verbose output to show what is being compiled
# - -o /app: Specifies the output path for the compiled binary
RUN GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags "-static"' -a -v -o /app ./cmd/webserver/webserver.go





# Final Stage: Using a static non-root distroless image for security
FROM ${FINAL_IMAGE}

# Copy the compiled application from the builder stage
COPY --from=builder /app /usr/local/bin/app

# Run the application as a non-root user for added security
USER nobody:nobody

# Command to run the application
CMD ["app"]
