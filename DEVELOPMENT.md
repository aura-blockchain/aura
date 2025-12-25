# Development Guide

## Prerequisites

- Go 1.24+
- Make
- Docker (optional, for testnet)
- Protocol Buffers compiler

## Quick Start

```bash
# Clone
git clone https://github.com/aequitas/aura.git
cd aura/chain

# Build
make build

# Run tests
make test

# Start local node
./aurad init mynode --chain-id aura-local-1
./aurad start
```

## Project Structure

```
chain/
├── cmd/aurad/        # CLI entry point
├── app/              # Application wiring
├── x/                # 27 Cosmos SDK modules
├── proto/            # Protobuf definitions
├── testing/          # Test utilities
└── docs/             # Module documentation
```

## Building

```bash
make build          # Build binary
make install        # Install to $GOPATH/bin
make proto-gen      # Regenerate protobuf
make lint           # Run linters
make test           # Run all tests
```

## Testing

```bash
# Unit tests
go test ./x/... -v

# Integration tests
go test ./testing/integration/... -v

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Module Development

1. Scaffold: `ignite scaffold module mymodule`
2. Define types in `proto/aura/mymodule/v1/`
3. Implement keeper in `x/mymodule/keeper/`
4. Add tests in `x/mymodule/keeper/*_test.go`
5. Register in `app/app.go`

## Code Standards

- Follow [Cosmos SDK patterns](https://docs.cosmos.network/)
- Use `errorsmod.Wrap` for errors
- Table-driven tests required
- 80%+ test coverage for new code
- Run `make lint` before committing

## Commit Convention

```
type(scope): description

feat(bridge): add cross-chain transfer
fix(dex): resolve orderbook race condition
docs(identity): update DID spec
test(compliance): add AML edge cases
```

## Resources

- [Architecture](ARCHITECTURE.md)
- [Contributing](CONTRIBUTING.md)
- [API Docs](docs/api/)
