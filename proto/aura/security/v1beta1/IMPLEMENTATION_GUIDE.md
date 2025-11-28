# Security Module Implementation Guide

This guide provides step-by-step instructions for implementing the consolidated security module based on the protobuf definitions.

## Quick Start

### 1. Generate Go Code

```bash
cd /home/decri/blockchain-projects/aura/proto
buf generate
```

This will generate:
- `aura/security/v1beta1/security.pb.go` - Types
- `aura/security/v1beta1/genesis.pb.go` - Genesis state
- `aura/security/v1beta1/query.pb.go` - Query messages
- `aura/security/v1beta1/query_grpc.pb.go` - gRPC query service
- `aura/security/v1beta1/tx.pb.go` - Transaction messages
- `aura/security/v1beta1/tx_grpc.pb.go` - gRPC transaction service

### 2. Directory Structure

Create the following directory structure:

```
chain/x/security/
├── keeper/
│   ├── keeper.go                    # Main keeper
│   ├── genesis.go                   # Genesis import/export
│   ├── msg_server.go                # Transaction message handlers
│   ├── query_server.go              # Query handlers
│   ├── network_security.go          # Network security logic
│   ├── validator_security.go        # Validator security logic
│   ├── wallet_security.go           # Wallet security logic
│   ├── incident_response.go         # Incident response logic
│   ├── cryptography.go              # Cryptography logic
│   └── privacy.go                   # Privacy logic
├── types/
│   ├── keys.go                      # KVStore keys
│   ├── errors.go                    # Error definitions
│   ├── events.go                    # Event types
│   ├── codec.go                     # Amino/Proto codec
│   └── expected_keepers.go          # Keeper interfaces
├── client/cli/
│   ├── query.go                     # CLI query commands
│   └── tx.go                        # CLI transaction commands
├── module.go                         # Module definition
└── README.md                         # Module documentation
```

## Implementation Steps

### Step 1: Define Store Keys

Create `chain/x/security/types/keys.go`:

```go
package types

const (
    ModuleName = "security"
    StoreKey   = ModuleName
    RouterKey  = ModuleName
)

// KVStore key prefixes
var (
    // Network Security
    TrustedPeerPrefix      = []byte{0x01}
    NodeReputationPrefix   = []byte{0x02}
    RateLimitPrefix        = []byte{0x03}
    ForkAlertPrefix        = []byte{0x04}
    PartitionAlertPrefix   = []byte{0x05}

    // Validator Security
    ValidatorSecurityPrefix     = []byte{0x10}
    DoubleSignEvidencePrefix    = []byte{0x11}
    DowntimeInfractionPrefix    = []byte{0x12}
    ValidatorAlertPrefix        = []byte{0x13}
    SentryNodePrefix            = []byte{0x14}

    // Wallet Security
    HardwareWalletPrefix        = []byte{0x20}
    MultiSigWalletPrefix        = []byte{0x21}
    PendingMultiSigTxPrefix     = []byte{0x22}
    SocialRecoveryPrefix        = []byte{0x23}
    RecoveryRequestPrefix       = []byte{0x24}
    SpendingLimitPrefix         = []byte{0x25}
    BiometricAuthPrefix         = []byte{0x26}

    // Incident Response
    IncidentPrefix              = []byte{0x30}
    AuditLogPrefix              = []byte{0x31}
    ResponseActionPrefix        = []byte{0x32}

    // Cryptography
    KeyRotationSchedulePrefix   = []byte{0x40}
    ThresholdSchemePrefix       = []byte{0x41}
    ZKProofConfigPrefix         = []byte{0x42}
    QuantumResistantKeyPrefix   = []byte{0x43}

    // Privacy
    MixingPoolPrefix            = []byte{0x50}
    StealthAddressPrefix        = []byte{0x51}
    RingSignaturePrefix         = []byte{0x52}
    ConfidentialTxPrefix        = []byte{0x53}
)
```

### Step 2: Create the Keeper

Create `chain/x/security/keeper/keeper.go`:

```go
package keeper

import (
    "cosmossdk.io/core/store"
    "cosmossdk.io/log"
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"

    "github.com/aequitas/aura/chain/x/security/types"
)

type Keeper struct {
    cdc          codec.BinaryCodec
    storeService store.KVStoreService
    authority    string

    // Expected keepers
    bankKeeper    types.BankKeeper
    stakingKeeper types.StakingKeeper
    slashingKeeper types.SlashingKeeper
}

func NewKeeper(
    cdc codec.BinaryCodec,
    storeService store.KVStoreService,
    authority string,
    bankKeeper types.BankKeeper,
    stakingKeeper types.StakingKeeper,
    slashingKeeper types.SlashingKeeper,
) *Keeper {
    return &Keeper{
        cdc:          cdc,
        storeService: storeService,
        authority:    authority,
        bankKeeper:    bankKeeper,
        stakingKeeper: stakingKeeper,
        slashingKeeper: slashingKeeper,
    }
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
    return ctx.Logger().With("module", "x/"+types.ModuleName)
}
```

### Step 3: Implement Message Server

Create `chain/x/security/keeper/msg_server.go`:

```go
package keeper

import (
    "context"

    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

type msgServer struct {
    Keeper
}

// NewMsgServerImpl returns an implementation of the security MsgServer interface
func NewMsgServerImpl(keeper Keeper) v1beta1.MsgServer {
    return &msgServer{Keeper: keeper}
}

var _ v1beta1.MsgServer = msgServer{}

// Example: AddTrustedPeer implementation
func (ms msgServer) AddTrustedPeer(
    goCtx context.Context,
    msg *v1beta1.MsgAddTrustedPeer,
) (*v1beta1.MsgAddTrustedPeerResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Validate authority
    if ms.authority != msg.Authority {
        return nil, errors.Wrapf(types.ErrUnauthorized, "expected %s, got %s", ms.authority, msg.Authority)
    }

    // Create trusted peer
    peer := &v1beta1.TrustedPeer{
        PeerId:      msg.PeerId,
        Address:     msg.Address,
        PublicKey:   msg.PublicKey,
        Description: msg.Description,
        AddedAt:     ctx.BlockTime(),
    }

    // Store trusted peer
    if err := ms.SetTrustedPeer(ctx, peer); err != nil {
        return nil, err
    }

    // Emit event
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeAddTrustedPeer,
            sdk.NewAttribute(types.AttributeKeyPeerID, msg.PeerId),
            sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
        ),
    )

    return &v1beta1.MsgAddTrustedPeerResponse{}, nil
}

// Implement remaining 50+ message handlers...
```

### Step 4: Implement Query Server

Create `chain/x/security/keeper/query_server.go`:

```go
package keeper

import (
    "context"

    sdk "github.com/cosmos/cosmos-sdk/types"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

type queryServer struct {
    Keeper
}

// NewQueryServerImpl returns an implementation of the security QueryServer interface
func NewQueryServerImpl(keeper Keeper) v1beta1.QueryServer {
    return &queryServer{Keeper: keeper}
}

var _ v1beta1.QueryServer = queryServer{}

// Params returns the consolidated security parameters
func (qs queryServer) Params(
    goCtx context.Context,
    req *v1beta1.QueryParamsRequest,
) (*v1beta1.QueryParamsResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

    ctx := sdk.UnwrapSDKContext(goCtx)
    params := qs.GetParams(ctx)

    return &v1beta1.QueryParamsResponse{Params: params}, nil
}

// SecurityStatus returns overall security status
func (qs queryServer) SecurityStatus(
    goCtx context.Context,
    req *v1beta1.QuerySecurityStatusRequest,
) (*v1beta1.QuerySecurityStatusResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

    ctx := sdk.UnwrapSDKContext(goCtx)

    // Collect status from all domains
    networkStatus := qs.GetNetworkSecurityStatus(ctx)
    validatorStatus := qs.GetValidatorSecurityStatus(ctx)
    walletStatus := qs.GetWalletSecurityStatus(ctx)
    incidentStatus := qs.GetIncidentResponseStatus(ctx)

    // Calculate overall security score
    securityScore := qs.CalculateOverallSecurityScore(ctx)

    return &v1beta1.QuerySecurityStatusResponse{
        NetworkHealthy:        networkStatus.Healthy,
        ConnectedPeers:        networkStatus.ConnectedPeers,
        HasActiveForks:        networkStatus.HasActiveForks,
        IsPartitioned:         networkStatus.IsPartitioned,
        TotalValidators:       validatorStatus.TotalValidators,
        JailedValidators:      validatorStatus.JailedValidators,
        ActiveValidatorAlerts: validatorStatus.ActiveAlerts,
        TotalWallets:          walletStatus.TotalWallets,
        HardwareWallets:       walletStatus.HardwareWallets,
        MultisigWallets:       walletStatus.MultisigWallets,
        ActiveIncidents:       incidentStatus.ActiveIncidents,
        CriticalIncidents:     incidentStatus.CriticalIncidents,
        OverallSecurityScore:  securityScore,
    }, nil
}

// Implement remaining 40+ query handlers...
```

### Step 5: Implement Genesis

Create `chain/x/security/keeper/genesis.go`:

```go
package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// InitGenesis initializes the security module's state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState *v1beta1.GenesisState) {
    // Set parameters
    k.SetParams(ctx, genState.Params)

    // Initialize network security state
    for _, peer := range genState.NetworkSecurity.TrustedPeers {
        k.SetTrustedPeer(ctx, &peer)
    }
    for _, rep := range genState.NetworkSecurity.Reputations {
        k.SetNodeReputation(ctx, &rep)
    }
    for _, limit := range genState.NetworkSecurity.RateLimits {
        k.SetRateLimitEntry(ctx, &limit)
    }
    for _, alert := range genState.NetworkSecurity.ForkAlerts {
        k.SetForkAlert(ctx, &alert)
    }
    for _, alert := range genState.NetworkSecurity.PartitionAlerts {
        k.SetPartitionAlert(ctx, &alert)
    }

    // Initialize validator security state
    for _, val := range genState.ValidatorSecurity.Validators {
        k.SetValidatorSecurityInfo(ctx, &val)
    }
    for _, evidence := range genState.ValidatorSecurity.DoubleSignEvidences {
        k.SetDoubleSignEvidence(ctx, &evidence)
    }
    // ... initialize remaining validator security state

    // Initialize wallet security state
    for _, hw := range genState.WalletSecurity.HardwareWallets {
        k.SetHardwareWalletConfig(ctx, &hw)
    }
    for _, ms := range genState.WalletSecurity.MultisigWallets {
        k.SetMultiSigWallet(ctx, &ms)
    }
    // ... initialize remaining wallet security state

    // Initialize incident response state
    for _, incident := range genState.IncidentResponse.Incidents {
        k.SetIncident(ctx, &incident)
    }
    // ... initialize remaining incident response state

    // Initialize cryptography state
    for _, schedule := range genState.Cryptography.KeyRotationSchedules {
        k.SetKeyRotationSchedule(ctx, &schedule)
    }
    // ... initialize remaining cryptography state

    // Initialize privacy state
    for _, pool := range genState.Privacy.MixingPools {
        k.SetMixingPool(ctx, &pool)
    }
    // ... initialize remaining privacy state
}

// ExportGenesis exports the security module's state to genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *v1beta1.GenesisState {
    return &v1beta1.GenesisState{
        Params: k.GetParams(ctx),
        NetworkSecurity: &v1beta1.NetworkSecurityState{
            TrustedPeers:    k.GetAllTrustedPeers(ctx),
            Reputations:     k.GetAllNodeReputations(ctx),
            RateLimits:      k.GetAllRateLimitEntries(ctx),
            ForkAlerts:      k.GetAllForkAlerts(ctx),
            PartitionAlerts: k.GetAllPartitionAlerts(ctx),
        },
        ValidatorSecurity: &v1beta1.ValidatorSecurityState{
            Validators:           k.GetAllValidatorSecurityInfo(ctx),
            DoubleSignEvidences:  k.GetAllDoubleSignEvidences(ctx),
            DowntimeInfractions:  k.GetAllDowntimeInfractions(ctx),
            Alerts:               k.GetAllValidatorAlerts(ctx),
            SentryNodes:          k.GetAllSentryNodes(ctx),
        },
        WalletSecurity: &v1beta1.WalletSecurityState{
            HardwareWallets:      k.GetAllHardwareWallets(ctx),
            MultisigWallets:      k.GetAllMultiSigWallets(ctx),
            PendingMultisigTxs:   k.GetAllPendingMultiSigTxs(ctx),
            SocialRecoveryConfigs: k.GetAllSocialRecoveryConfigs(ctx),
            RecoveryRequests:     k.GetAllRecoveryRequests(ctx),
            SpendingLimits:       k.GetAllSpendingLimits(ctx),
            BiometricAuths:       k.GetAllBiometricAuths(ctx),
        },
        IncidentResponse: &v1beta1.IncidentResponseState{
            Incidents:       k.GetAllIncidents(ctx),
            AuditLogs:       k.GetAllAuditLogs(ctx),
            ResponseActions: k.GetAllResponseActions(ctx),
        },
        Cryptography: &v1beta1.CryptographyState{
            KeyRotationSchedules:  k.GetAllKeyRotationSchedules(ctx),
            ThresholdSchemes:      k.GetAllThresholdSchemes(ctx),
            ZkProofConfigs:        k.GetAllZKProofConfigs(ctx),
            QuantumResistantKeys:  k.GetAllQuantumResistantKeys(ctx),
        },
        Privacy: &v1beta1.PrivacyState{
            MixingPools:              k.GetAllMixingPools(ctx),
            StealthAddresses:         k.GetAllStealthAddresses(ctx),
            RingSignatures:           k.GetAllRingSignatures(ctx),
            ConfidentialTransactions: k.GetAllConfidentialTransactions(ctx),
        },
    }
}
```

### Step 6: Define Module

Create `chain/x/security/module.go`:

```go
package security

import (
    "context"
    "encoding/json"
    "fmt"

    "cosmossdk.io/core/appmodule"
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/codec"
    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/spf13/cobra"

    "github.com/aequitas/aura/chain/x/security/keeper"
    "github.com/aequitas/aura/chain/x/security/types"
    securityv1 "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

var (
    _ module.AppModule      = AppModule{}
    _ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module
type AppModuleBasic struct {
    cdc codec.Codec
}

// Name returns the module's name
func (AppModuleBasic) Name() string {
    return types.ModuleName
}

// RegisterLegacyAminoCodec registers the module's types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
    types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
    types.RegisterInterfaces(reg)
}

// DefaultGenesis returns default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
    return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis validates the genesis state
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
    var genState securityv1.GenesisState
    if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
        return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
    }
    return types.ValidateGenesis(&genState)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
    if err := securityv1.RegisterQueryHandlerClient(context.Background(), mux, securityv1.NewQueryClient(clientCtx)); err != nil {
        panic(err)
    }
}

// GetTxCmd returns the root tx command
func (AppModuleBasic) GetTxCmd() *cobra.Command {
    return cli.GetTxCmd()
}

// GetQueryCmd returns the root query command
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
    return cli.GetQueryCmd()
}

// AppModule implements an application module
type AppModule struct {
    AppModuleBasic

    keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
    return AppModule{
        AppModuleBasic: AppModuleBasic{cdc: cdc},
        keeper:         keeper,
    }
}

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
    securityv1.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
    securityv1.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs genesis initialization
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
    var genState securityv1.GenesisState
    cdc.MustUnmarshalJSON(data, &genState)
    am.keeper.InitGenesis(ctx, &genState)
    return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
    genState := am.keeper.ExportGenesis(ctx)
    return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }
```

## Testing

Create comprehensive tests for each component:

```go
// keeper_test.go
func TestKeeper(t *testing.T) {
    // Test keeper initialization
    // Test parameter setting/getting
    // Test state persistence
}

// msg_server_test.go
func TestMsgServer(t *testing.T) {
    // Test each message handler
    // Test authorization
    // Test error cases
}

// query_server_test.go
func TestQueryServer(t *testing.T) {
    // Test each query handler
    // Test pagination
    // Test edge cases
}

// genesis_test.go
func TestGenesis(t *testing.T) {
    // Test genesis initialization
    // Test genesis export
    // Test round-trip consistency
}
```

## Integration

Add the security module to your app:

```go
// In app/app.go

import (
    securitykeeper "github.com/aequitas/aura/chain/x/security/keeper"
    securitytypes "github.com/aequitas/aura/chain/x/security/types"
    security "github.com/aequitas/aura/chain/x/security"
)

// Add to ModuleBasics
ModuleBasics = module.NewBasicManager(
    // ... other modules
    security.AppModuleBasic{},
)

// Add keeper to App struct
type App struct {
    // ... other keepers
    SecurityKeeper securitykeeper.Keeper
}

// Initialize keeper
app.SecurityKeeper = securitykeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[securitytypes.StoreKey]),
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    app.BankKeeper,
    app.StakingKeeper,
    app.SlashingKeeper,
)

// Register module
app.ModuleManager = module.NewManager(
    // ... other modules
    security.NewAppModule(appCodec, app.SecurityKeeper),
)

// Set order
app.ModuleManager.SetOrderBeginBlockers(
    // ... other modules
    securitytypes.ModuleName,
)

app.ModuleManager.SetOrderEndBlockers(
    // ... other modules
    securitytypes.ModuleName,
)

app.ModuleManager.SetOrderInitGenesis(
    // ... other modules
    securitytypes.ModuleName,
)
```

## CLI Commands

Examples of CLI commands to implement:

```bash
# Query commands
aurad query security params
aurad query security status
aurad query security peer <peer-id>
aurad query security validator <validator-address>
aurad query security wallet <wallet-id>
aurad query security incident <incident-id>

# Transaction commands
aurad tx security add-trusted-peer <peer-id> <address> <pubkey>
aurad tx security register-hardware-wallet <type> <device-id>
aurad tx security create-multisig <signers> <threshold>
aurad tx security create-incident <title> <description> <severity>
aurad tx security rotate-key <key-id>
```

## Monitoring and Observability

Add metrics, logs, and traces:

```go
// Add Prometheus metrics
var (
    activePeersMetric = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "security_active_peers",
        Help: "Number of active peers",
    })

    activeIncidentsMetric = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "security_active_incidents",
        Help: "Number of active security incidents",
    })
)

// Add structured logging
k.Logger(ctx).Info("trusted peer added",
    "peer_id", peerID,
    "address", address,
)

// Add tracing
ctx, span := tracer.Start(ctx, "security.AddTrustedPeer")
defer span.End()
```

## Next Steps

1. Implement keeper methods for each domain
2. Add comprehensive unit tests
3. Add integration tests
4. Add CLI commands
5. Update documentation
6. Add migration scripts from old modules
7. Deploy and test on testnet
8. Audit security implementation
9. Deploy to mainnet

## Support

For questions or issues:
- Review existing module implementations
- Check the proto definitions in this directory
- Consult Cosmos SDK documentation
- Review security best practices
