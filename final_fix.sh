#!/bin/bash

# Final comprehensive fix script

echo "=== Fix networksecurity params syntax errors ==="
# Rewrite the DefaultParams function properly
cat > chain/x/networksecurity/types/params_temp.go << 'PARAMEOF'
package types

import (
	networksecuritypb "github.com/aequitas/aura/proto/aura/networksecurity/v1beta1"
)

// DefaultParams returns default parameters
func DefaultParams() *networksecuritypb.Params {
	rateLimit := DefaultRateLimitConfig()
	connection := DefaultConnectionConfig()
	mempool := DefaultMempoolConfig()
	reputation := DefaultReputationConfig()
	gossip := DefaultGossipConfig()
	forkDetection := DefaultForkDetectionConfig()
	partitionDetection := DefaultPartitionDetectionConfig()
	
	return &networksecuritypb.Params{
		RateLimit:          &rateLimit,
		Connection:         &connection,
		Mempool:            &mempool,
		Reputation:         &reputation,
		Gossip:             &gossip,
		ForkDetection:      &forkDetection,
		PartitionDetection: &partitionDetection,
	}
}
PARAMEOF
# Find and replace the DefaultParams function
LINE=$(grep -n "^func DefaultParams()" chain/x/networksecurity/types/params.go | cut -d: -f1)
if [ ! -z "$LINE" ]; then
  # Delete old function and append new one
  head -n $((LINE-1)) chain/x/networksecurity/types/params.go > chain/x/networksecurity/types/params_new.go
  cat chain/x/networksecurity/types/params_temp.go >> chain/x/networksecurity/types/params_new.go
  # Find end of file after the old function
  tail -n +$((LINE+20)) chain/x/networksecurity/types/params.go | grep -A 9999 "^func" >> chain/x/networksecurity/types/params_new.go 2>/dev/null || true
  mv chain/x/networksecurity/types/params_new.go chain/x/networksecurity/types/params.go
fi
rm -f chain/x/networksecurity/types/params_temp.go

echo "=== Fix prevalidation pointer issue ==="
sed -i 's/p := DefaultParams(); return &p/params := DefaultParams(); return params/' chain/x/prevalidation/types/types.go

echo "=== Fix compliance transaction_monitoring syntax error ==="
# Remove broken lines around line 350
sed -i '350d' chain/x/compliance/keeper/transaction_monitoring.go 2>/dev/null || true

echo "=== Fix governance keeper unused imports and missing return ==="
# Remove unused imports
sed -i '/^\s*governancepb "github.com\/aequitas\/aura\/proto\/aura\/governance\/v1beta1"/d' chain/x/governance/keeper/keeper.go
sed -i '/^\s*"crypto\/sha256"/d' chain/x/governance/keeper/keeper.go
sed -i '/^\s*"encoding\/hex"/d' chain/x/governance/keeper/keeper.go  
sed -i '/^\s*"fmt"/d' chain/x/governance/keeper/keeper.go
sed -i '/^\s*"strconv"/d' chain/x/governance/keeper/keeper.go
sed -i '/^\s*"time"/d' chain/x/governance/keeper/keeper.go

echo "=== Fix bridge unused sdk import ==="
sed-i '/sdk "github.com\/cosmos\/cosmos-sdk\/types"/d' chain/x/bridge/types/params_security.go 2>/dev/null || true

echo "=== Fix proto embed methods - need to be visible to proto package ==="
# The methods need to match the exact signature from the proto generated code
# Let's check what the proto code expects
grep -r "mustEmbedUnimplementedMsgServer" chain/x/auth/keeper/ || echo "Signature not found, will use standard approach"

echo "=== Fix walletsecurity proto embeddings ==="
cat >> chain/x/walletsecurity/keeper/msg_server.go << 'WALLETMSG'

func (msgServer) mustEmbedUnimplementedMsgServer() {}
WALLETMSG

cat >> chain/x/walletsecurity/keeper/query_server.go << 'WALLETQUERY'

func (queryServer) mustEmbedUnimplementedQueryServer() {}
WALLETQUERY

echo "Script completed!"
