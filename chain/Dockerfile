# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache make git gcc musl-dev linux-headers

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT}" \
    -trimpath \
    -o /app/aurad \
    ./cmd/aurad

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S aura && adduser -S aura -G aura

# Set working directory
WORKDIR /home/aura

# Copy binary from builder
COPY --from=builder /app/aurad /usr/local/bin/aurad

# Set ownership
RUN chown -R aura:aura /home/aura

# Switch to non-root user
USER aura

# Expose ports
# P2P
EXPOSE 26656
# RPC
EXPOSE 26657
# REST API
EXPOSE 1317
# gRPC
EXPOSE 9090
# gRPC-web
EXPOSE 9091

# Default command
ENTRYPOINT ["aurad"]
CMD ["start"]
