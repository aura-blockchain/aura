.PHONY: help build test proto-gen clean install lint fmt test-integration test-e2e test-stress \
	test-chaos test-coverage test-bench test-unit test-vc test-cs test-ir test-id test-race \
	mod-tidy check ci security-scan docker-up docker-down docker-logs init-testnet start-node reset-testnet

# Build variables
BUILD_DIR := build
CHAIN_DIR := chain
PROTO_DIR := proto
BINARY_NAME := aurad
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html
NODE_HOME := $(HOME)/.aura

# Help target - Show all available targets
help:
	@echo "============================================================================"
	@echo "AURA Blockchain - Makefile Targets"
	@echo "============================================================================"
	@echo ""
	@echo "Installation & Setup:"
	@echo "  install               ## Install binaries to GOPATH"
	@echo ""
	@echo "Development:"
	@echo "  build                 ## Build all modules and create binaries"
	@echo "  fmt                   ## Format code with gofmt and buf"
	@echo "  lint                  ## Run linters (golangci-lint)"
	@echo "  proto-gen             ## Generate protobuf bindings"
	@echo "  mod-tidy              ## Tidy Go modules"
	@echo ""
	@echo "Testing:"
	@echo "  test                  ## Run all tests"
	@echo "  test-unit             ## Run unit tests only"
	@echo "  test-integration      ## Run integration tests"
	@echo "  test-e2e              ## Run end-to-end tests"
	@echo "  test-stress           ## Run stress tests"
	@echo "  test-chaos            ## Run chaos engineering tests"
	@echo "  test-coverage         ## Run tests with coverage report"
	@echo "  test-race             ## Run tests with race detector"
	@echo "  test-bench            ## Run benchmark tests"
	@echo "  test-vc               ## Run VC registry tests"
	@echo "  test-cs               ## Run confidence score tests"
	@echo "  test-ir               ## Run inclusion routines tests"
	@echo "  test-id               ## Run identity change tests"
	@echo ""
	@echo "Docker & Services:"
	@echo "  docker-up             ## Start Docker services"
	@echo "  docker-down           ## Stop Docker services"
	@echo "  docker-logs           ## View Docker service logs"
	@echo ""
	@echo "Blockchain Operations:"
	@echo "  init-testnet          ## Initialize testnet node"
	@echo "  start-node            ## Start blockchain node"
	@echo "  reset-testnet         ## Reset testnet data"
	@echo ""
	@echo "Quality & Security:"
	@echo "  check                 ## Run all checks (fmt, lint, test)"
	@echo "  security-scan         ## Run security vulnerability scan"
	@echo "  ci                    ## Run CI pipeline"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean                 ## Remove build artifacts and caches"
	@echo ""

# Build all modules
build:
	@echo "Building AURA blockchain..."
	cd $(CHAIN_DIR) && go build ./...
	cd $(CHAIN_DIR) && go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/aurad
	@echo "Build complete!"

# Run all tests
test:
	@echo "Running all tests..."
	cd $(CHAIN_DIR) && go test ./... -v

# Run unit tests only (excluding integration tests)
test-unit:
	@echo "Running unit tests..."
	cd $(CHAIN_DIR) && go test ./... -short -v

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	cd $(CHAIN_DIR) && go test ./testing/integration/... -v -tags=integration -timeout 45m

# Run end-to-end tests
test-e2e:
	@echo "Running end-to-end tests..."
	cd $(CHAIN_DIR) && go test ./testing/e2e/... -v -tags=e2e -timeout 60m

# Run stress tests
test-stress:
	@echo "Running stress tests..."
	cd $(CHAIN_DIR) && go test ./testing/stress/... -v -tags=stress -timeout 90m

# Run chaos engineering tests
test-chaos:
	@echo "Running chaos engineering tests..."
	cd $(CHAIN_DIR) && go test ./testing/chaos/... -v -tags=chaos -timeout 60m

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	cd $(CHAIN_DIR) && go test ./... -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic
	cd $(CHAIN_DIR) && go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(CHAIN_DIR)/$(COVERAGE_HTML)"
	cd $(CHAIN_DIR) && go tool cover -func=$(COVERAGE_FILE) | grep total

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	cd $(CHAIN_DIR) && go test ./... -v -race

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	cd $(CHAIN_DIR) && go test ./testing/benchmark/... -bench=. -benchmem -timeout 30m

# Run VC registry tests
test-vc:
	@echo "Running VC registry tests..."
	cd $(CHAIN_DIR) && go test ./x/vcregistry/... -v

# Run confidence score tests
test-cs:
	@echo "Running confidence score tests..."
	cd $(CHAIN_DIR) && go test ./x/confidencescore/... -v

# Run inclusion routines tests
test-ir:
	@echo "Running inclusion routines tests..."
	cd $(CHAIN_DIR) && go test ./x/inclusionroutines/... -v

# Run identity change tests
test-id:
	@echo "Running identity change tests..."
	cd $(CHAIN_DIR) && go test ./x/identitychange/... -v

# Generate protobuf bindings
proto-gen:
	@echo "Generating protobuf bindings..."
	cd $(PROTO_DIR) && buf generate --template buf.gen.yaml
	@echo "Protobuf generation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(CHAIN_DIR)/$(BUILD_DIR)
	rm -f $(CHAIN_DIR)/$(COVERAGE_FILE)
	rm -f $(CHAIN_DIR)/$(COVERAGE_HTML)
	rm -f $(PROTO_DIR)/aura/**/*.pb.go
	@echo "Clean complete!"

# Install binaries to GOPATH
install:
	@echo "Installing binaries..."
	cd $(CHAIN_DIR) && go install ./cmd/aurad
	@echo "Install complete!"

# Run linters
lint:
	@echo "Running linters..."
	cd $(CHAIN_DIR) && golangci-lint run --timeout=10m
	@echo "Lint complete!"

# Format code
fmt:
	@echo "Formatting code..."
	cd $(CHAIN_DIR) && go fmt ./...
	cd $(PROTO_DIR) && buf format -w
	@echo "Format complete!"

# Tidy Go modules
mod-tidy:
	@echo "Tidying Go modules..."
	cd $(CHAIN_DIR) && go mod tidy
	cd $(PROTO_DIR) && go mod tidy
	@echo "Mod tidy complete!"

# Run security scan
security-scan:
	@echo "Running security scan..."
	cd $(CHAIN_DIR) && govulncheck ./...
	@echo "Security scan complete!"

# Run all checks
check: fmt lint test
	@echo "All checks passed!"

# Run CI pipeline
ci: build test-coverage lint
	@echo "CI pipeline complete!"

# ============================================================================
# Docker & Services
# ============================================================================

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "✓ Docker services started"
	@docker-compose ps

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "✓ Docker services stopped"

docker-logs:
	@echo "Viewing Docker service logs (press Ctrl+C to exit)..."
	docker-compose logs -f

# ============================================================================
# Blockchain Node Operations
# ============================================================================

init-testnet:
	@echo "Initializing testnet node..."
	@mkdir -p $(NODE_HOME)
	cd $(CHAIN_DIR) && go run ./cmd/aurad init-testnet --home $(NODE_HOME)
	@echo "✓ Testnet node initialized at $(NODE_HOME)"

start-node:
	@echo "Starting blockchain node..."
	cd $(CHAIN_DIR) && go run ./cmd/aurad start --home $(NODE_HOME)

reset-testnet:
	@echo "Resetting testnet data..."
	@rm -rf $(NODE_HOME)
	@echo "✓ Testnet data reset"
	@$(MAKE) init-testnet
