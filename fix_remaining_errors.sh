#!/bin/bash

set -e

echo "=== Fixing governance keeper syntax error (EOF) ==="
# The keeper.go has syntax errors from sed, let's check the end
cd chain/x/governance/keeper
# Count braces
OPEN=$(grep -o '{' keeper.go | wc -l)
CLOSE=$(grep -o '}' keeper.go | wc -l)
echo "Open braces: $OPEN, Close braces: $CLOSE"
if [ $OPEN -gt $CLOSE ]; then
  echo "Adding missing closing brace"
  echo "}" >> keeper.go
fi
cd -

echo "=== Removing unused import from benchmark ==="
sed -i '/^\s*"context"/d' chain/testing/benchmark/benchmark.go

echo "=== Fixing compliance keeper duplicate methods ==="
cd chain/x/compliance/keeper
# Remove the actual duplicate method declarations
sed -i '90,110{/^func (k \*Keeper) GetKYCRecord/,/^}/d}' kyc_aml.go 2>/dev/null || true
sed -i '180,200{/^func (k \*Keeper) GetAMLProfile/,/^}/d}' kyc_aml.go 2>/dev/null || true
sed -i '260,280{/^func (k \*Keeper) GetSanctionsResult/,/^}/d}' sanctions.go 2>/dev/null || true
sed -i '350,370{/^func (k \*Keeper) GetTransactionAlerts/,/^}/d}' transaction_monitoring.go 2>/dev/null || true
cd -

echo "=== Fixing auth proto embeddings (make methods exported) ==="
sed -i 's/func (m msgServer) mustEmbedUnimplementedMsgServer/func (msgServer) mustEmbedUnimplementedMsgServer/' chain/x/auth/keeper/msg_server.go
sed -i 's/func (q queryServer) mustEmbedUnimplementedQueryServer/func (queryServer) mustEmbedUnimplementedQueryServer/' chain/x/auth/keeper/query_server.go

echo "=== Fixing vcregistry proto embeddings (make methods exported) ==="
sed -i 's/func (m \*MsgServer) mustEmbedUnimplementedMsgServer/func (\*MsgServer) mustEmbedUnimplementedMsgServer/' chain/x/vcregistry/keeper/msg_server.go
sed -i 's/func (q \*QueryServer) mustEmbedUnimplementedQueryServer/func (\*QueryServer) mustEmbedUnimplementedQueryServer/' chain/x/vcregistry/keeper/query.go

echo "=== Fixing prevalidation invalid indirection ==="
sed -i 's/return &\*DefaultParams()/p := DefaultParams(); return \&p/g' chain/x/prevalidation/types/types.go

echo "=== Fixing economicsecurity params ==="
sed -i 's/types.ValidateParams(params)/types.ValidateParams(\&params)/g' chain/x/economicsecurity/params/params.go

echo "=== Fixing bridge types ==="
# Fix sdk.OneDec
sed -i 's/sdk\.OneDec/math.LegacyOneDec/g' chain/x/bridge/types/params_security.go
# Remove unused math import from expected_keepers.go
sed -i '/^\s*"cosmossdk.io\/math"$/d' chain/x/bridge/types/expected_keepers.go

echo "=== Fixing dex types ==="
# Remove unused sdk import
sed -i '/sdk "github.com\/cosmos\/cosmos-sdk\/types"/d' chain/x/dex/types/security_types.go
# Remove unused math import
sed -i '/^\s*"cosmossdk.io\/math"$/d' chain/x/dex/types/expected_keepers.go

echo "=== Fixing networksecurity types pointer issues ==="
sed -i 's/Params: \*DefaultParams()/p := DefaultParams(); Params: \&p/g' chain/x/networksecurity/types/genesis.go
sed -i 's/RateLimit: \*DefaultRateLimitConfig()/r := DefaultRateLimitConfig(); RateLimit: \&r/g' chain/x/networksecurity/types/params.go
sed -i 's/Connection: \*DefaultConnectionConfig()/c := DefaultConnectionConfig(); Connection: \&c/g' chain/x/networksecurity/types/params.go
sed -i 's/Mempool: \*DefaultMempoolConfig()/m := DefaultMempoolConfig(); Mempool: \&m/g' chain/x/networksecurity/types/params.go
sed -i 's/Reputation: \*DefaultReputationConfig()/r := DefaultReputationConfig(); Reputation: \&r/g' chain/x/networksecurity/types/params.go
sed -i 's/Gossip: \*DefaultGossipConfig()/g := DefaultGossipConfig(); Gossip: \&g/g' chain/x/networksecurity/types/params.go
sed -i 's/ForkDetection: \*DefaultForkDetectionConfig()/f := DefaultForkDetectionConfig(); ForkDetection: \&f/g' chain/x/networksecurity/types/params.go
sed -i 's/PartitionDetection: \*DefaultPartitionDetectionConfig()/p := DefaultPartitionDetectionConfig(); PartitionDetection: \&p/g' chain/x/networksecurity/types/params.go

echo "Script completed!"
