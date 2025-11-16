# Aura Blockchain - Production Dockerfile
# Multi-stage build for optimized image size (Go + PHP)

# ============================================================================
# Stage 1: Go Builder - Compile Cosmos SDK blockchain
# ============================================================================
FROM golang:1.25-alpine as go-builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev linux-headers

# Set working directory
WORKDIR /build

# Copy go modules
COPY chain/go.mod chain/go.sum ./
RUN go mod download

# Copy proto files and dependencies
COPY proto/go.mod proto/go.sum ./proto/
COPY chain . .

# Build the blockchain binary
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o aura-chain ./cmd/aurad/main.go

# ============================================================================
# Stage 2: PHP Builder - Prepare PHP dependencies
# ============================================================================
FROM php:8.2-cli-alpine as php-builder

# Install PHP build dependencies
RUN apk add --no-cache \
    composer \
    git \
    libxml2-dev \
    curl-dev \
    openssl-dev

# Set working directory
WORKDIR /build

# Copy composer files
COPY composer.json composer.lock ./

# Install PHP dependencies
RUN composer install --no-dev --optimize-autoloader --no-interaction

# ============================================================================
# Stage 3: Runtime - Minimal production image
# ============================================================================
FROM alpine:3.18

# Metadata labels
LABEL maintainer="Aequitas Blockchain Team"
LABEL description="Aura Blockchain - Cosmos SDK with PHP Integration"
LABEL version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/aequitas/aura"

# Set environment variables
ENV AURA_ENV=production \
    AURA_HOME=/app \
    AURA_DATA_DIR=/data \
    AURA_LOG_DIR=/logs \
    PATH="/app/bin:$PATH"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    libssl3 \
    libcrypto3 \
    tini \
    php82 \
    php82-curl \
    php82-json \
    php82-xml \
    php82-pdo \
    php82-pdo_pgsql \
    php82-openssl

# Create non-root user for security
RUN addgroup -S aura -g 1000 && \
    adduser -S aura -G aura -u 1000 -h /home/aura && \
    mkdir -p /app /data /logs && \
    chown -R aura:aura /app /data /logs

# Set working directory
WORKDIR /app

# Copy Go binary from builder
COPY --from=go-builder /build/aura-chain /app/bin/aura-chain

# Copy PHP dependencies from builder
COPY --from=php-builder /build/vendor /app/vendor

# Copy application code
COPY --chown=aura:aura chain/ ./chain/
COPY --chown=aura:aura wallet/ ./wallet/
COPY --chown=aura:aura ai-assistant/ ./ai-assistant/
COPY --chown=aura:aura verifier-portal/ ./verifier-portal/
COPY --chown=aura:aura config/ ./config/

# Create necessary directories
RUN mkdir -p \
    /data/blockchain \
    /data/state \
    /logs/node \
    /logs/api && \
    chown -R aura:aura /data /logs

# Switch to non-root user
USER aura

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:${AURA_API_PORT:-26657}/health || exit 1

# Expose ports
# 26656 - P2P network port
# 26657 - RPC API port
# 1317  - REST API port
# 9090  - gRPC port
# 9091  - gRPC-Web port
EXPOSE 26656 26657 1317 9090 9091

# Volumes
VOLUME ["/data", "/logs"]

# Use tini to handle signals properly
ENTRYPOINT ["/sbin/tini", "--"]

# Default command - start blockchain node
CMD ["/app/bin/aura-chain", "start"]
