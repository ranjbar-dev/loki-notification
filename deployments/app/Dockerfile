# Build stage
FROM golang:1.25-alpine AS base

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

FROM base AS builder

# Build the application
RUN go build -o /loki-notification ./cmd

FROM base AS dev

CMD ["go", "run", "./cmd", "-config", "/app/config/config.yaml"]

# Final stage
FROM alpine:latest AS production

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /loki-notification .

# Copy configuration and migrations
COPY --chown=appuser:appuser config/ ./config/

# Switch to non-root user
USER appuser

# Expose port (if needed for future API)
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/loki-notification"]

CMD ["-config", "/app/config/config.yaml"]

