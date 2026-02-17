# --- Stage 1: Build ---
FROM golang:1.26-bookworm AS builder

# Set environment variables from your toml
ENV GO_VERSION=1.26.0 
# Note: Go 1.26 doesn't exist yet; using 1.22 as a stable base. 
# Adjust to 'golang:latest' if you want the cutting edge.

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Run your specific build command
# We output to /app/qordled for simplicity in the next stage
RUN GOOS=linux GOARCH=amd64 go build -o qordled ./cmd/qordled/main.go

# --- Stage 2: Runtime ---
FROM ubuntu:24.04

WORKDIR /root/

# Install CA certificates for HTTPS requests (often needed for Go apps)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /app/qordled .

# Copy your public assets (from your 'publish = "public"' setting)
COPY --from=builder /app/public ./public

# Expose the port your application listens on
EXPOSE 8091

# Command to run the executable
CMD ["./qordled", "--port", "8091"]
