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
CONTRACT_PACKAGES := binding-tester vc-issuer

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

## test-authz: Run auth-focused PermissionDenied mapping tests
test-authz:
	@echo "Running auth-focused tests (PermissionDenied mapping)..."
	@go test ./x/identity/keeper ./x/incidentresponse/keeper ./x/economicsecurity/keeper ./x/vcregistry/keeper

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
	@echo "Running tests with reduced verbosity..."
	@go test -v -run Test -count=1 ./...

## deploy-contracts: Deploy contracts using foundry (requires forge)
deploy-contracts:
	@echo "Deploying contracts..."
	@cd contracts && forge script script/Deploy.s.sol:DeployScript --rpc-url $${RPC_URL} --broadcast --verify

## test-contracts: Run contract tests (requires foundry)
test-contracts:
	@echo "Running contract tests..."
	@cd contracts && forge test

## proto-gen: Generate Go code from proto files
proto-gen:
	@echo "Generating Go code from proto files..."
	@cd ../proto && buf generate
	@echo "Proto generation complete"

## proto-swagger: Generate OpenAPI/Swagger specs from proto files
proto-swagger:
	@echo "Generating OpenAPI specs from proto files..."
	@cd ../proto && buf generate --template buf.gen.openapi.yaml
	@python3 ../scripts/merge-swagger.py
	@echo "OpenAPI spec generated: docs/api/openapi.json"

## proto-all: Generate all proto outputs (Go code + OpenAPI)
proto-all: proto-gen proto-swagger
	@echo "All proto outputs generated"
