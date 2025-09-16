# Multi-stage build for hostsctl

# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for go modules
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o hostsctl \
    ./cmd/hostsctl

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 hostsctl && \
    adduser -D -u 1001 -G hostsctl hostsctl

# Set working directory
WORKDIR /home/hostsctl

# Copy binary from builder stage
COPY --from=builder /app/hostsctl /usr/local/bin/hostsctl

# Make binary executable
RUN chmod +x /usr/local/bin/hostsctl

# Copy example configs
COPY --from=builder /app/configs ./configs

# Change ownership
RUN chown -R hostsctl:hostsctl /home/hostsctl

# Switch to non-root user
USER hostsctl

# Set entrypoint
ENTRYPOINT ["hostsctl"]

# Default command
CMD ["--help"]