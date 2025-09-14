# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o user-service ./cmd/user-service

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary
COPY --from=builder /app/user-service /user-service

# Copy config files
COPY --from=builder /app/configs /configs

# Expose port
EXPOSE 8082

# Add health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/user-service", "health"]

# Run as non-root user
USER 65534:65534

CMD ["/user-service"]
