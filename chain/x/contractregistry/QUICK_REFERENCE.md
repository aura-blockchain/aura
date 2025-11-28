# ContractRegistry Module - Quick Reference

## Server Interface Usage

### Importing the Interfaces
```go
import (
    "github.com/aequitas/aura/chain/x/contractregistry/types"
)

// Use the interfaces
var msgServer types.MsgServer
var queryServer types.QueryServer
```

### Creating Server Implementations
```go
import (
    "github.com/aequitas/aura/chain/x/contractregistry/keeper"
    pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

// In your test or application code
msgServer := keeper.NewMsgServerImpl(keeper)
queryServer := keeper.NewQueryServerImpl(keeper)

// Or use the protobuf interface directly
var msgServer pb.MsgServer = keeper.NewMsgServerImpl(keeper)
var queryServer pb.QueryServer = keeper.NewQueryServerImpl(keeper)
```

### Type Alias Pattern
The types package aliases the protobuf-generated interfaces:

**File**: `/home/decri/blockchain-projects/aura/chain/x/contractregistry/types/expected_keepers.go`
```go
type (
    // MsgServer is the server API for Msg service
    MsgServer = pb.MsgServer

    // QueryServer is the server API for Query service
    QueryServer = pb.QueryServer
)
```

### Available Methods

#### MsgServer Methods
```go
RegisterContract(ctx, *MsgRegisterContract) (*MsgRegisterContractResponse, error)
UpdateContractMetadata(ctx, *MsgUpdateContractMetadata) (*MsgUpdateContractMetadataResponse, error)
UpdateSecurityPolicy(ctx, *MsgUpdateSecurityPolicy) (*MsgUpdateSecurityPolicyResponse, error)
PauseContract(ctx, *MsgPauseContract) (*MsgPauseContractResponse, error)
UnpauseContract(ctx, *MsgUnpauseContract) (*MsgUnpauseContractResponse, error)
DeprecateContract(ctx, *MsgDeprecateContract) (*MsgDeprecateContractResponse, error)
WhitelistContract(ctx, *MsgWhitelistContract) (*MsgWhitelistContractResponse, error)
BlacklistContract(ctx, *MsgBlacklistContract) (*MsgBlacklistContractResponse, error)
RemoveFromBlacklist(ctx, *MsgRemoveFromBlacklist) (*MsgRemoveFromBlacklistResponse, error)
AuditContract(ctx, *MsgAuditContract) (*MsgAuditContractResponse, error)
VerifyContract(ctx, *MsgVerifyContract) (*MsgVerifyContractResponse, error)
```

#### QueryServer Methods
```go
ContractInfo(ctx, *QueryContractInfoRequest) (*QueryContractInfoResponse, error)
ContractsByCreator(ctx, *QueryContractsByCreatorRequest) (*QueryContractsByCreatorResponse, error)
ContractsByTag(ctx, *QueryContractsByTagRequest) (*QueryContractsByTagResponse, error)
RegisteredContracts(ctx, *QueryRegisteredContractsRequest) (*QueryRegisteredContractsResponse, error)
ContractMetrics(ctx, *QueryContractMetricsRequest) (*QueryContractMetricsResponse, error)
```

## Key Files

### Interface Definitions
- **Type Aliases**: `/home/decri/blockchain-projects/aura/chain/x/contractregistry/types/expected_keepers.go`
- **Proto Msg Interface**: `/home/decri/blockchain-projects/aura/proto/aura/contractregistry/v1beta1/msg_grpc.pb.go`
- **Proto Query Interface**: `/home/decri/blockchain-projects/aura/proto/aura/contractregistry/v1beta1/query_grpc.pb.go`

### Implementations
- **MsgServer**: `/home/decri/blockchain-projects/aura/chain/x/contractregistry/keeper/msg_server.go`
- **QueryServer**: `/home/decri/blockchain-projects/aura/chain/x/contractregistry/keeper/query_server.go`

### Test Files
- `/home/decri/blockchain-projects/aura/chain/x/contractregistry/keeper/msg_server_test.go`
- `/home/decri/blockchain-projects/aura/chain/x/contractregistry/keeper/query_server_test.go`
- `/home/decri/blockchain-projects/aura/chain/x/contractregistry/keeper/msg_server_comprehensive_test.go`
- `/home/decri/blockchain-projects/aura/chain/x/contractregistry/keeper/query_server_comprehensive_test.go`

## Testing Pattern
```go
type MsgServerTestSuite struct {
    suite.Suite
    keeper    *keeper.Keeper
    msgServer types.MsgServer  // ✅ Uses type alias
    ctx       *testutil.TestContext
}

func (s *MsgServerTestSuite) SetupTest() {
    s.keeper = &keeper.Keeper{}
    s.msgServer = keeper.NewMsgServerImpl(s.keeper)
}
```

## Status
✅ All interfaces properly defined and aliased
✅ All test files compile successfully
✅ Follows Cosmos SDK v0.50 best practices
