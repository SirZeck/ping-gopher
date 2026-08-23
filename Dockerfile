# Stage 1: Build the Go binary using Go 1.22 builder image
FROM golang:1.22-alpine AS builder

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

RUN apk add --no-cache ca-certificates tzdata wget

# Create non-root user for security hardening
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/pinggopher /app/pinggopher

RUN chown -R appuser:appgroup /app

USER appuser

# Expose HTTP API port
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/v1/status/public || exit 1

# Default entrypoint & command
ENTRYPOINT ["/app/pinggopher"]
CMD ["--role=all"]
