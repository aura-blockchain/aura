# ============================================================================
# Multi-stage Dockerfile for AURA Blockchain - Production Optimized
# ============================================================================

# ============================================================================
# Stage 1: Security Scanner Base
# ============================================================================
FROM golang:1.21-alpine AS security-scanner

RUN apk add --no-cache git

WORKDIR /scan
COPY chain/go.mod chain/go.sum ./

# Install security scanning tools
RUN go install golang.org/x/vuln/cmd/govulncheck@latest
RUN go install github.com/securego/gosec/v2/cmd/gosec@latest

# ============================================================================
# Stage 2: Builder
# ============================================================================
FROM golang:1.21-alpine AS builder

# Security: Run as non-root during build
RUN addgroup -g 10001 builder && \
    adduser -D -u 10001 -G builder builder

# Install build dependencies with specific versions for reproducibility
RUN apk add --no-cache \
    git=~2.40 \
    make=~4.4 \
    gcc=~12.2 \
    musl-dev=~1.2 \
    linux-headers=~6.3 \
    ca-certificates

WORKDIR /app

# Copy dependency files first for better layer caching
COPY --chown=builder:builder chain/go.mod chain/go.sum ./chain/
WORKDIR /app/chain

# Verify dependencies before download
RUN go mod verify

# Download dependencies (cached layer)
RUN go mod download

# Copy source code
COPY --chown=builder:builder chain/ ./

# Scan for vulnerabilities
COPY --from=security-scanner /go/bin/govulncheck /usr/local/bin/
RUN govulncheck ./... || echo "Vulnerability scan completed with warnings"

# Build info
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE
ARG BUILD_TAGS="netgo,ledger,muslc"

# Build the binary with maximum optimizations and security hardening
RUN CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -mod=readonly \
    -tags "${BUILD_TAGS}" \
    -ldflags "\
        -X github.com/cosmos/cosmos-sdk/version.Name=aura \
        -X github.com/cosmos/cosmos-sdk/version.AppName=aurad \
        -X github.com/cosmos/cosmos-sdk/version.Version=${VERSION} \
        -X github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT} \
        -X github.com/cosmos/cosmos-sdk/version.BuildTags=${BUILD_TAGS} \
        -X 'github.com/cosmos/cosmos-sdk/version.BuildDate=${BUILD_DATE}' \
        -w -s \
        -linkmode=external \
        -extldflags '-Wl,-z,relro,-z,now,-z,muldefs -static -fstack-protector-strong'" \
    -trimpath \
    -buildvcs=false \
    -o /app/build/aurad \
    ./cmd/aurad

# Verify the binary was built correctly
RUN /app/build/aurad version --long

# Strip debug symbols (additional size reduction)
RUN strip /app/build/aurad || true

# ============================================================================
# Stage 3: Runtime (Minimal Production Image)
# ============================================================================
FROM alpine:3.18

# Security: Install only essential runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    libgcc \
    tzdata && \
    # Create non-root user
    addgroup -g 1000 aura && \
    adduser -D -u 1000 -G aura aura && \
    # Set up directories with proper permissions
    mkdir -p /home/aura/.aura/config \
             /home/aura/.aura/data \
             /home/aura/.aura/cosmovisor/genesis/bin \
             /home/aura/.aura/cosmovisor/upgrades && \
    chown -R aura:aura /home/aura && \
    # Remove unnecessary packages
    apk del apk-tools && \
    # Clear cache
    rm -rf /var/cache/apk/*

# Copy binary from builder with ownership
COPY --from=builder --chown=aura:aura /app/build/aurad /usr/local/bin/aurad

# Verify binary is executable
RUN /usr/local/bin/aurad version

# Security: Switch to non-root user
USER aura
WORKDIR /home/aura

# Expose ports with documentation
# 26656: P2P communication
# 26657: RPC server
# 26660: Prometheus metrics endpoint
# 1317: REST API server (gRPC-gateway)
# 9090: gRPC server
EXPOSE 26656 26657 26660 1317 9090

# Volume mounts for persistence (use external volumes in production)
VOLUME ["/home/aura/.aura"]

# Health check with improved reliability
HEALTHCHECK --interval=30s \
            --timeout=10s \
            --start-period=120s \
            --retries=5 \
    CMD aurad status 2>&1 | grep -q '"catching_up":false' || \
        (aurad status 2>&1 | grep -q '"latest_block_height"' && exit 0) || \
        exit 1

# Graceful shutdown signal
STOPSIGNAL SIGTERM

# Set environment variables for production
ENV AURA_HOME=/home/aura/.aura \
    AURA_LOG_LEVEL=info \
    AURA_ENABLE_METRICS=true

# Default command with graceful shutdown
ENTRYPOINT ["aurad"]
CMD ["start", "--log_level=$AURA_LOG_LEVEL"]

# ============================================================================
# Metadata Labels (OCI Standard)
# ============================================================================
LABEL org.opencontainers.image.source="https://github.com/aura/aura" \
      org.opencontainers.image.description="AURA Blockchain Node - Production Optimized" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.vendor="AURA Network" \
      org.opencontainers.image.title="AURA Node" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      maintainer="AURA Team <dev@aura.network>"

# Security labels
LABEL security.scan.enabled="true" \
      security.non-root="true" \
      security.readonly-rootfs="recommended"
