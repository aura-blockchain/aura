# Aura Blockchain Daemon Implementation

## Overview

This document describes the implementation of the Aura blockchain daemon (`aurad`), the main entry point for running an Aura blockchain node.

## File Structure

```
cmd/aurad/
├── main.go                 # Main entry point
├── README.md               # User documentation
├── IMPLEMENTATION.md       # This file
└── cmd/                    # Command implementations
    ├── root.go             # Root command and configuration
    ├── init.go             # Node initialization command
    ├── start.go            # Node start command
    ├── status.go           # Status query command
    ├── version.go          # Version information command
    ├── keys.go             # Key management commands
    ├── query.go            # Query commands
    └── tx.go               # Transaction commands
```

## Implementation Details

### main.go

**Location**: `C:\Users\decri\GitClones\aura\chain\cmd\aurad\main.go`

**Purpose**: Entry point for the Aura blockchain daemon.

**Key Features**:
- Clean, minimal entry point
- Delegates all functionality to the cmd package
- Proper error handling and exit codes
- Professional error messaging

**Code Structure**:
```go
func main() {
    rootCmd := cmd.NewRootCmd()
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error executing aurad: %v\n", err)
        os.Exit(1)
    }
}
```

### cmd/root.go

**Purpose**: Root command implementation and global configuration.

**Key Features**:
- Root command setup with Cobra
- Global flag definitions (home, config, log-level, log-format)
- Configuration initialization using Viper
- Environment variable support with AURA_ prefix
- Subcommand registration
- Lazy app initialization

**Configuration**:
- Default home directory: `~/.aura`
- Config file: `~/.aura/config/config.toml`
- Environment prefix: `AURA_`

### cmd/init.go

**Purpose**: Initialize a new Aura blockchain node.

**Key Features**:
- Creates directory structure (config/, data/, keys/)
- Generates config.toml with node settings
- Generates app.toml with application settings
- Supports custom chain ID and moniker
- Skip existing files to prevent overwriting

**Usage**:
```bash
aurad init my-node --chain-id aura-1
```

**Generated Files**:
1. `config/config.toml` - Node configuration (RPC, P2P, Consensus)
2. `config/app.toml` - Application configuration (gRPC, API, Modules)

### cmd/start.go

**Purpose**: Start the Aura blockchain node.

**Key Features**:
- Starts gRPC server on port 9090
- Starts API server on port 1317
- Lazy app initialization
- Graceful shutdown handling (SIGINT/SIGTERM)
- Concurrent server execution
- Error propagation from servers
- Health check endpoint
- Module registration

**Architecture**:
```
startNode()
  ├─> startGRPCServer()  (port 9090)
  ├─> startAPIServer()   (port 1317)
  └─> Signal Handler     (SIGINT/SIGTERM)
```

### cmd/status.go

**Purpose**: Query the running node for status.

**Key Features**:
- Connects to API server
- Retrieves node status
- Displays chain information
- Shows endpoint addresses
- Connection error handling
- Timeout protection (5 seconds)

### cmd/version.go

**Purpose**: Display version information.

**Key Features**:
- Shows version, commit hash, build date
- Displays Go version and OS/Arch
- Lists all enabled modules
- Build-time variable injection via ldflags

**Version Variables**:
- `Version` - Set via ldflags during build
- `Commit` - Git commit hash
- `BuildDate` - Build timestamp

### cmd/keys.go

**Purpose**: Key management commands.

**Commands**:
- `aurad keys add [name]` - Add new key
- `aurad keys list` - List all keys
- `aurad keys show [name]` - Show key details
- `aurad keys delete [name]` - Delete key

**Note**: Full implementation requires Cosmos SDK keyring integration.

### cmd/query.go

**Purpose**: Query blockchain state and module data.

**Query Hierarchy**:
```
aurad query
├── block [height]
├── tx [hash]
├── account [address]
├── identitychange
│   ├── params
│   └── record [address]
├── inclusionroutines
│   ├── params
│   └── routine [id]
├── confidencescore
│   ├── params
│   └── score [address]
├── vcregistry
│   ├── params
│   └── credential [id]
├── dataregistry
│   ├── params
│   └── data [id]
└── governance
    ├── params
    ├── proposal [id]
    └── proposals
```

### cmd/tx.go

**Purpose**: Create and broadcast transactions.

**Transaction Hierarchy**:
```
aurad tx
├── broadcast [file]
├── sign [file]
├── identitychange
│   ├── register [did]
│   ├── update [did]
│   └── deactivate [did]
├── inclusionroutines
│   ├── propose [routine-json]
│   └── vote [routine-id] [vote]
├── confidencescore
│   └── update [address] [score]
├── vcregistry
│   ├── issue [credential-json]
│   └── revoke [credential-id]
├── dataregistry
│   ├── register [data-json]
│   └── update [data-id] [data-json]
└── governance
    ├── submit-proposal [proposal-json]
    ├── vote [proposal-id] [vote]
    └── deposit [proposal-id] [amount]
```

## Module Integration

The daemon integrates with 8 Aura blockchain modules:

1. **Identity Change** - DID lifecycle management
2. **Inclusion Routines** - Routine proposals and voting
3. **Confidence Score** - Address confidence scoring
4. **VC Registry** - Verifiable credential management
5. **Data Registry** - On-chain data registration
6. **Governance** - On-chain governance
7. **DEX** - Decentralized exchange (disabled, pending SDK integration)
8. **Bridge** - Cross-chain bridge (disabled, pending SDK integration)

## Configuration Management

### Viper Integration

All configuration is managed through Viper with the following precedence:
1. Command-line flags (highest priority)
2. Environment variables (AURA_ prefix)
3. Config files (config.toml, app.toml)
4. Default values (lowest priority)

### Environment Variables

All flags can be set via environment variables:
```bash
export AURA_HOME=/custom/path
export AURA_LOG_LEVEL=debug
export AURA_LOG_FORMAT=json
```

## Server Architecture

### gRPC Server
- Port: 9090 (default)
- Protocol: gRPC
- Purpose: Module service registration and gRPC queries
- Configuration: `--grpc-address`

### API Server
- Port: 1317 (default)
- Protocol: HTTP/REST
- Endpoints:
  - `/` - Basic status
  - `/health` - Health check
- Configuration: `--api-address`

### RPC Server
- Port: 26657 (default)
- Protocol: Tendermint RPC
- Purpose: Blockchain RPC queries
- Configuration: `--rpc-address`

## Error Handling

### Graceful Shutdown
- Listens for SIGINT and SIGTERM signals
- Allows running goroutines to complete
- Proper resource cleanup

### Error Propagation
- Errors from servers are propagated to main error channel
- First error stops the node
- Clear error messages to stderr

## Build System

### Makefile

**Location**: `C:\Users\decri\GitClones\aura\chain\Makefile`

**Targets**:
- `make build` - Build the binary
- `make install` - Install to $GOPATH/bin
- `make clean` - Remove build artifacts
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage
- `make lint` - Run linters
- `make fmt` - Format code

**Version Injection**:
```makefile
LDFLAGS=-ldflags "\
    -X github.com/aequitas/aura/chain/cmd/aurad/cmd.Version=$(VERSION) \
    -X github.com/aequitas/aura/chain/cmd/aurad/cmd.Commit=$(COMMIT) \
    -X github.com/aequitas/aura/chain/cmd/aurad/cmd.BuildDate=$(BUILD_DATE)"
```

## Dependencies

### Core Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `cosmossdk.io/log` - Logging
- `google.golang.org/grpc` - gRPC server

### Application Dependencies
- `github.com/aequitas/aura/chain/app` - Aura app
- All Aura module packages (x/identitychange, x/vcregistry, etc.)

## Security Considerations

### Key Management
- Keys stored in configurable keyring backend
- Support for os, file, test, memory backends
- Secure key generation and storage

### Configuration
- Sensitive data not logged
- Config files have appropriate permissions (0644)
- Directories created with 0755 permissions

## Future Enhancements

### Planned Features
1. Full Cosmos SDK integration
2. State sync support
3. Snapshot functionality
4. Advanced debugging tools
5. Metrics and monitoring endpoints
6. WebSocket support
7. Complete DEX module integration
8. Complete Bridge module integration

### Command Additions
1. Export/import genesis
2. Tendermint commands
3. Debug commands
4. Validate genesis command

## Testing

### Unit Tests
- Each command should have unit tests
- Mock app and logger for testing
- Test configuration loading
- Test error handling

### Integration Tests
- Test full node initialization
- Test server startup
- Test query/tx workflows
- Test graceful shutdown

## Documentation

### User Documentation
- `README.md` - Comprehensive user guide
- Command help text (via Cobra)
- Example commands
- Troubleshooting guide

### Developer Documentation
- `IMPLEMENTATION.md` - This file
- Code comments
- Architecture diagrams

## Compliance

### Cosmos SDK Patterns
- Follows standard Cosmos daemon structure
- Compatible with standard Cosmos tools
- Uses standard port numbers

### Best Practices
- Clean code architecture
- Separation of concerns
- Comprehensive error handling
- Proper logging
- Security-first approach

## Maintenance

### Version Updates
Update version variables in Makefile or use git tags:
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
make build
```

### Configuration Updates
When adding new configuration:
1. Add to config.toml/app.toml templates
2. Update Viper bindings
3. Document in README.md

## Contact

For questions or contributions:
- GitHub: https://github.com/aequitas/aura
- Issues: https://github.com/aequitas/aura/issues
