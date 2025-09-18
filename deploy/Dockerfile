# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker's build cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
# CGO_ENABLED=0 disables CGo, resulting in a statically linked binary
# -o specifies the output binary name
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fetchtracker cmd/fetchtracker/*.go

# Runtime Stage
FROM alpine:latest

WORKDIR /srv/

# Copy the built binary from the builder stage
COPY --from=builder /app/fetchtracker .

# Expose the port your application listens on (if applicable)
EXPOSE 8080

# Set the entry point to run the application
CMD ["./fetchtracker", "-c", "config.yml"]
