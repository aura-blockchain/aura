# Contributor FAQ

Common questions and answers for Aura contributors.

## Getting Started

### Q: How do I set up my development environment?

```bash
# Clone and enter the repository
git clone https://github.com/aura-blockchain/aura.git
cd aura/chain

# Install Go 1.24+ (required)
go version  # Should show 1.24+

# Install dependencies and build
go mod download
go build -o aurad ./cmd/aurad

# Install pre-commit hooks
cd .. && pre-commit install
```

### Q: How do I run a local testnet?

```bash
cd chain
./aurad init my-node --chain-id aura-local-1
./aurad start
```

Or use Docker:
```bash
cd k8s && docker-compose up
```

### Q: Where is the documentation?

| Resource | Location |
|----------|----------|
| API Reference | `docs/api/openapi.json` |
| CLI Commands | `docs/development/CLI_*.md` |
| Module Docs | `docs/modules/<module>/` |
| Architecture | `chain/docs/` |

---

## Development

### Q: How do I add a new module?

1. Scaffold with Ignite: `ignite scaffold module <name>`
2. Define proto types in `proto/aura/<name>/v1beta1/`
3. Generate Go code: `make proto-gen`
4. Implement keeper in `chain/x/<name>/keeper/`
5. Register in `chain/app/app.go`
6. Add tests in `chain/x/<name>/keeper/*_test.go`

### Q: How do I run tests?

```bash
cd chain

# All tests
go test ./...

# Specific module
go test ./x/identity/...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Quick authz tests
make test-authz
```

### Q: How do I regenerate proto files?

```bash
cd chain
make proto-gen        # Go code
make proto-swagger    # OpenAPI specs
```

### Q: What's the commit message format?

Follow [Conventional Commits](https://conventionalcommits.org):
- `feat(module): add new feature`
- `fix(module): correct bug`
- `docs: update README`
- `test(module): add tests`
- `refactor(module): restructure code`

---

## Architecture

### Q: What are the core modules?

| Module | Purpose |
|--------|---------|
| `identity` | DID management, multisig, RBAC |
| `privacy` | Confidential transactions, ZK proofs |
| `compliance` | KYC/AML, sanctions, GDPR |
| `dex` | Order book, liquidity pools, swaps |
| `bridge` | Cross-chain transfers |
| `vcregistry` | Verifiable Credentials |

### Q: How do modules communicate?

Modules use keeper references passed in `app.go`. Example:
```go
app.IdentityKeeper = identitykeeper.NewKeeper(
    app.ComplianceKeeper,  // Cross-module dependency
    ...
)
```

### Q: Where are the proto definitions?

Proto files are in `proto/aura/<module>/v1beta1/`:
- `*.proto` - Type definitions
- `query.proto` - Query service (with HTTP annotations)
- `tx.proto` - Transaction messages
- `genesis.proto` - Genesis state

---

## Testing

### Q: How do I write table-driven tests?

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "foo", "bar", false},
        {"empty input", "", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Q: How do I set up a test keeper?

See `chain/testutil/` for helpers. Example:
```go
func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
    // Use testutil helpers
    return testutil.NewTestKeeper(t)
}
```

---

## Common Issues

### Q: Build fails with "undefined" errors

Run `go mod tidy` and ensure you're using Go 1.24+.

### Q: Proto generation fails

1. Install `buf`: `go install github.com/bufbuild/buf/cmd/buf@latest`
2. Update deps: `cd proto && buf dep update`
3. Regenerate: `make proto-gen`

### Q: Tests fail with type mismatches

Check `chain/docs/GOGOPROTO_TYPES.md` for proto type handling. Common fixes:
- Use `time.Time` not `timestamppb.New()`
- Use `sdkmath.NewInt()` not string amounts
- Use value types when `nullable=false` in proto

### Q: How do I debug a failing invariant?

Add logging in the invariant function:
```go
func MyInvariant(k Keeper) sdk.Invariant {
    return func(ctx sdk.Context) (string, bool) {
        k.Logger(ctx).Info("checking invariant", "key", value)
        // ...
    }
}
```

---

## SDKs

### Q: What SDKs are available?

| Language | Location | Status |
|----------|----------|--------|
| Go | `sdk/go/` | Production |
| JavaScript | `sdk/javascript/` | Production |
| Python | `sdk/python/` | Production |

### Q: How do I use the Go SDK?

```go
import "github.com/aura-blockchain/aura/sdk/go/client"

c, _ := client.New("http://localhost:1317")
identity, _ := c.Identity.GetRecord(ctx, "did:aura:123")
```

---

## Getting Help

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and ideas
- **Documentation**: `docs/` directory
- **Code Comments**: Most modules have detailed docstrings
