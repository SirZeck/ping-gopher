# Stage 1: Build the Go binary using latest Go 1.25 builder image
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy module definitions
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build production binary (CGO disabled for static linking)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/pinggopher ./cmd/pinggopher

# Stage 2: Lightweight runtime image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/pinggopher /app/pinggopher

# Expose HTTP API port
EXPOSE 8080

# Default entrypoint & command
ENTRYPOINT ["/app/pinggopher"]
CMD ["--role=all"]
