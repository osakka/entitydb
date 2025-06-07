# EntityDB Production Container
# Multi-stage build for optimal security and size

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy source code
COPY src/ .

# Build arguments
ARG VERSION=unknown
ARG BUILD_DATE=unknown

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE}" \
    -a -installsuffix cgo \
    -o entitydb main.go

# Verify the binary
RUN ./entitydb --version

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 entitydb && \
    adduser -D -u 1000 -G entitydb entitydb && \
    mkdir -p /app/var /app/logs && \
    chown -R entitydb:entitydb /app

# Copy binary from builder
COPY --from=builder /build/entitydb /app/entitydb

# Copy static files
COPY share/htdocs/ /app/share/htdocs/
COPY share/config/ /app/share/config/

# Set proper permissions
RUN chown -R entitydb:entitydb /app

# Switch to non-root user
USER entitydb

# Set working directory
WORKDIR /app

# Expose ports
EXPOSE 8085 8443

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ./entitydb --entitydb-port 8085 --help > /dev/null || exit 1

# Set environment defaults
ENV ENTITYDB_DATA_PATH=/app/var \
    ENTITYDB_STATIC_DIR=/app/share/htdocs \
    ENTITYDB_LOG_LEVEL=info \
    ENTITYDB_PORT=8085 \
    ENTITYDB_SSL_PORT=8443

# Default command
CMD ["./entitydb", "--entitydb-data-path", "/app/var", "--entitydb-static-dir", "/app/share/htdocs"]

# Labels for metadata
LABEL org.opencontainers.image.title="EntityDB" \
      org.opencontainers.image.description="High-performance temporal database with nanosecond precision" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/your-org/entitydb" \
      org.opencontainers.image.licenses="MIT"