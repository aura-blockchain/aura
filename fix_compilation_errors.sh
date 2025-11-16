#!/bin/bash

# Script to fix common compilation errors across the Aura blockchain

set -e

echo "=== Fixing Bridge Types SDK Migration ==="
find chain/x/bridge/types -name "*.go" -type f -exec sed -i '
/^import (/,/^)/{
	/sdk "github.com\/cosmos\/cosmos-sdk\/types"/a\	"cosmossdk.io/math"
}
s/\([^.]\)sdk\.Int\([^(]\)/\1math.Int\2/g
s/^sdk\.Int\([^(]\)/math.Int\1/g
s/sdk\.NewInt/math.NewInt/g
s/sdk\.ZeroInt/math.ZeroInt/g
s/sdk\.Dec/math.LegacyDec/g
s/sdk\.NewDec(/math.LegacyNewDec(/g
s/sdk\.NewDecWithPrec/math.LegacyNewDecWithPrec/g
s/sdk\.ZeroDec/math.LegacyZeroDec/g
' {} \;

echo "=== Fixing DEX Types SDK Migration ==="
find chain/x/dex/types -name "*.go" -type f -exec sed -i '
/^import (/,/^)/{
	/sdk "github.com\/cosmos\/cosmos-sdk\/types"/a\	"cosmossdk.io/math"
}
s/\([^.]\)sdk\.Int\([^(]\)/\1math.Int\2/g
s/^sdk\.Int\([^(]\)/math.Int\1/g
s/sdk\.NewInt/math.NewInt/g
s/sdk\.ZeroInt/math.ZeroInt/g
s/sdk\.Dec/math.LegacyDec/g
s/sdk\.NewDec(/math.LegacyNewDec(/g
s/sdk\.NewDecWithPrec/math.LegacyNewDecWithPrec/g
s/sdk\.ZeroDec/math.LegacyZeroDec/g
' {} \;

echo "=== Fixing Auth Keeper - Remove Duplicate Methods ==="
cd chain/x/auth/keeper
# Remove duplicate GetEmergencyAdmin from keeper.go
sed -i '/^func (k \*Keeper) GetEmergencyAdmin/,/^}/d' keeper.go 2>/dev/null || true
# Remove duplicate GetValidatorKeyRotation from keeper.go
sed -i '/^\/\/ GetValidatorKeyRotation/,/^}/d' keeper.go 2>/dev/null || true
# Remove duplicate GetMultisigWallet from keeper.go
sed -i '/^\/\/ GetMultisigWallet/,/^}/d' keeper.go 2>/dev/null || true
# Remove duplicate GetMultisigProposal from keeper.go
sed -i '/^\/\/ GetMultisigProposal/,/^}/d' keeper.go 2>/dev/null || true
# Remove duplicate GetRateLimitConfig from keeper.go
sed -i '/^\/\/ GetRateLimitConfig/,/^}/d' keeper.go 2>/dev/null || true
# Remove duplicate GetSession from keeper.go
sed -i '/^\/\/ GetSession/,/^}/d' keeper.go 2>/dev/null || true
# Remove duplicate GetTimeLockedAction from keeper.go
sed -i '/^\/\/ GetTimeLockedAction/,/^}/d' keeper.go 2>/dev/null || true
cd -

echo "=== Fixing Auth Keeper - Add Proto Embeddings ==="
cat >> chain/x/auth/keeper/msg_server.go << 'EOFMSG'

// mustEmbedUnimplementedMsgServer implements the proto interface requirement
func (m msgServer) mustEmbedUnimplementedMsgServer() {}
EOFMSG

cat >> chain/x/auth/keeper/query_server.go << 'EOFQUERY'

// mustEmbedUnimplementedQueryServer implements the proto interface requirement
func (q queryServer) mustEmbedUnimplementedQueryServer() {}
EOFQUERY

echo "=== Fixing Cryptography Keeper - Add Proto Embeddings ==="
cat >> chain/x/cryptography/keeper/msg_server.go << 'EOFMSG2'

// mustEmbedUnimplementedMsgServer implements the proto interface requirement
func (m msgServer) mustEmbedUnimplementedMsgServer() {}
EOFMSG2

cat >> chain/x/cryptography/keeper/query_server.go << 'EOFQUERY2'

// mustEmbedUnimplementedQueryServer implements the proto interface requirement
func (q queryServer) mustEmbedUnimplementedQueryServer() {}
EOFQUERY2

echo "=== Fixing Cryptography Keeper - Remove Duplicate Methods ==="
cd chain/x/cryptography/keeper
sed -i '/^\/\/ GetCertificatePin retrieves/,/^}/d' keeper.go 2>/dev/null || true
sed -i '/^\/\/ GetRandomSource retrieves/,/^}/d' keeper.go 2>/dev/null || true
cd -

echo "=== Fixing Cryptography Keeper - Timestamp Conversions ==="
find chain/x/cryptography/keeper -name "*.go" -type f -exec sed -i '
s/ExpiresAt: now,/ExpiresAt: timestamppb.New(now),/g
s/ExpiresAt: expiresAt,/ExpiresAt: timestamppb.New(*expiresAt),/g
s/pin\.ExpiresAt\.Before/pin.ExpiresAt.AsTime().Before/g
' {} \;

echo "=== Fixing Compliance Keeper - Remove Duplicate Methods ==="
cd chain/x/compliance/keeper
sed -i '/^func (k \*Keeper) GetKYCRecord.*already declared/,/^}/d' kyc_aml.go 2>/dev/null || true
sed -i '/^func (k \*Keeper) GetAMLProfile.*already declared/,/^}/d' kyc_aml.go 2>/dev/null || true  
sed -i '/^func (k \*Keeper) GetSanctionsResult.*already declared/,/^}/d' sanctions.go 2>/dev/null || true
sed -i '/^func (k \*Keeper) GetTransactionAlerts.*already declared/,/^}/d' transaction_monitoring.go 2>/dev/null || true
cd -

echo "=== Fixing VCRegistry Keeper - Add Proto Embeddings ==="
cat >> chain/x/vcregistry/keeper/msg_server.go << 'EOFMSG3'

// mustEmbedUnimplementedMsgServer implements the proto interface requirement
func (m *MsgServer) mustEmbedUnimplementedMsgServer() {}
EOFMSG3

cat >> chain/x/vcregistry/keeper/query.go << 'EOFQUERY3'

// mustEmbedUnimplementedQueryServer implements the proto interface requirement
func (q *QueryServer) mustEmbedUnimplementedQueryServer() {}
EOFQUERY3

echo "=== Fixing Prevalidation Types ==="
sed -i 's/return &\*DefaultParams()/return DefaultParams()/g' chain/x/prevalidation/types/types.go 2>/dev/null || true

echo "Script completed!"
