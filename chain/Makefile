# Aura Blockchain Makefile

# Build variables
BINARY_NAME=aurad
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "development")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

SHORT_TEST_PACKAGES := ./app ./x/vcregistry
SHORT_TEST_PARALLEL ?= 4
SHORT_TEST_TIMEOUT ?= 8m

# Go build flags
LDFLAGS=-ldflags "\
	-X github.com/aequitas/aura/chain/cmd/aurad/cmd.Version=$(VERSION) \
	-X github.com/aequitas/aura/chain/cmd/aurad/cmd.Commit=$(COMMIT) \
	-X github.com/aequitas/aura/chain/cmd/aurad/cmd.BuildDate=$(BUILD_DATE)"

.PHONY: all build install clean test lint help

all: build

## build: Build the aurad binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/aurad
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install the aurad binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) ./cmd/aurad
	@echo "Binary installed to: $$(go env GOPATH)/bin/$(BINARY_NAME)"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v -parallel=4 -timeout=30m ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

## lint: Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

## init-node: Initialize a local node
init-node:
	@echo "Initializing local node..."
	@$(BUILD_DIR)/$(BINARY_NAME) init my-node --chain-id aura-localnet

## start-node: Start the local node
start-node:
	@echo "Starting local node..."
	@$(BUILD_DIR)/$(BINARY_NAME) start

## version: Display version information
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

## help: Display this help message
help:
	@echo "Aura Blockchain Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^##/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print "  " substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t

## test-fast: Run tests in parallel with optimizations (8 parallel workers)
test-fast:
	@echo "Running tests with optimizations..."
	@go test -v -parallel=8 -timeout=20m -count=1 ./...

## test-race: Run tests with race detection (4 parallel workers)
test-race:
	@echo "Running tests with race detection..."
	@go test -v -race -parallel=4 -timeout=30m ./...

## test-short: Run only short tests (skip long-running integration tests)
test-short:
	@echo "Running short tests only (focused package list)..."
	@for pkg in $(SHORT_TEST_PACKAGES); do \
		echo "  -> $$pkg"; \
		go test -v -short -parallel=$(SHORT_TEST_PARALLEL) -timeout=$(SHORT_TEST_TIMEOUT) $$pkg; \
	done

## test-cached: Run tests using cache for faster reruns
test-cached:
	@echo "Running tests with cache..."
	@go test -v -parallel=8 -timeout=20m ./...

## test-pkg: Run tests for specific package (usage: make test-pkg PKG=./x/aura)
test-pkg:
	@echo "Running tests for package: $(PKG)"
	@go test -v -parallel=4 -timeout=10m $(PKG)

## test-verbose-less: Run tests with less verbose output
test-verbose-less:
	@echo "Running tests (minimal output)..."
	@go test -parallel=8 -timeout=20m ./...

###############################################################################
###                           WASM Integration                              ###
###############################################################################

## build-wasm: Compile WASM contracts (requires Rust toolchain)
build-wasm:
	@echo "Building WASM contracts..."
	@if [ -d "../contracts" ]; then \
		cd ../contracts && \
		for contract in */; do \
			if [ -f "$$contract/Cargo.toml" ]; then \
				echo "  Building $$contract..."; \
				cd "$$contract" && \
				cargo wasm && \
				cd ..; \
			fi; \
		done; \
	else \
		echo "Contracts directory not found. Run 'make init-wasm-structure' first."; \
	fi

## optimize-wasm: Optimize WASM contracts using rust-optimizer
optimize-wasm:
	@echo "Optimizing WASM contracts..."
	@if command -v docker >/dev/null 2>&1; then \
		if [ -d "../contracts" ]; then \
			cd ../contracts && \
			docker run --rm -v "$$(pwd)":/code \
				--mount type=volume,source="$$(basename $$(pwd))_cache",target=/code/target \
				--mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
				cosmwasm/rust-optimizer:0.15.0; \
		else \
			echo "Contracts directory not found."; \
		fi; \
	else \
		echo "Docker is required for WASM optimization."; \
		exit 1; \
	fi

## test-wasm: Run WASM module tests
test-wasm:
	@echo "Running WASM module tests..."
	@go test -v -parallel=4 -timeout=10m ./x/wasm/... ./x/aura-bindings/... ./x/contractregistry/...

## test-wasm-coverage: Run WASM tests with coverage
test-wasm-coverage:
	@echo "Running WASM tests with coverage..."
	@go test -v -coverprofile=coverage_wasm.out ./x/wasm/... ./x/aura-bindings/... ./x/contractregistry/...
	@go tool cover -func=coverage_wasm.out

## test-contracts: Run Rust contract tests
test-contracts:
	@echo "Running contract tests..."
	@if [ -d "../contracts" ]; then \
		cd ../contracts && \
		for contract in */; do \
			if [ -f "$$contract/Cargo.toml" ]; then \
				echo "  Testing $$contract..."; \
				cd "$$contract" && \
				cargo test && \
				cd ..; \
			fi; \
		done; \
	else \
		echo "Contracts directory not found."; \
	fi

## schema-contracts: Generate JSON schemas for contracts
schema-contracts:
	@echo "Generating contract schemas..."
	@if [ -d "../contracts" ]; then \
		cd ../contracts && \
		for contract in */; do \
			if [ -f "$$contract/Cargo.toml" ]; then \
				echo "  Generating schema for $$contract..."; \
				cd "$$contract" && \
				cargo schema && \
				cd ..; \
			fi; \
		done; \
	else \
		echo "Contracts directory not found."; \
	fi

## check-rust: Check Rust toolchain and dependencies
check-rust:
	@echo "Checking Rust toolchain..."
	@if ! command -v rustc >/dev/null 2>&1; then \
		echo "Error: Rust is not installed. Please install from https://rustup.rs/"; \
		exit 1; \
	fi
	@echo "  Rust version: $$(rustc --version)"
	@if ! rustup target list --installed | grep -q wasm32-unknown-unknown; then \
		echo "  Installing wasm32-unknown-unknown target..."; \
		rustup target add wasm32-unknown-unknown; \
	else \
		echo "  wasm32-unknown-unknown target: installed"; \
	fi
	@if ! command -v cargo-generate >/dev/null 2>&1; then \
		echo "  Note: cargo-generate not found. Install with: cargo install cargo-generate"; \
	else \
		echo "  cargo-generate: $$(cargo-generate --version)"; \
	fi

## init-wasm-structure: Create initial WASM project structure
init-wasm-structure:
	@echo "Creating WASM project structure..."
	@mkdir -p x/wasm x/aura-bindings x/contractregistry
	@mkdir -p ../contracts/packages/aura-bindings
	@mkdir -p ../proto/aura/contractregistry/v1beta1
	@mkdir -p ../proto/aura/aura-bindings/v1beta1
	@echo "WASM project structure created."
