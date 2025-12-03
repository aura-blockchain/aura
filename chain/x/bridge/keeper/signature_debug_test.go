package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ripemd160"
)

func TestDebugSignatureVerification(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInput(t)
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, &ps, nil, nil, nil, nil)
	ctx := input.Ctx

	// Generate key pair
	privKey, _ := secp256k1.GeneratePrivateKey()
	pubKey := privKey.PubKey()

	// Derive address
	pubKeyBytes := pubKey.SerializeCompressed()
	sha256Hash := sha256.Sum256(pubKeyBytes)
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	addressHash := ripemd160Hasher.Sum(nil)
	pawAddress := hex.EncodeToString(addressHash)
	auraAddress := "aura1test"

	fmt.Printf("PAW Address: %s\n", pawAddress)

	// Sign message
	message := fmt.Sprintf("Link PAW address %s to Aura address %s", pawAddress, auraAddress)
	msgHash := sha256.Sum256([]byte(message))
	fmt.Printf("Message: %s\n", message)

	// Sign using SignCompact
	compactSig := secp256k1ecdsa.SignCompact(privKey, msgHash[:], true)
	fmt.Printf("Compact sig length: %d, recovery byte: %d\n", len(compactSig), compactSig[0])

	// Rearrange to [R][S][recovery_id]
	signature := make([]byte, 65)
	copy(signature[0:64], compactSig[1:65])
	signature[64] = compactSig[0] - 27
	fmt.Printf("Signature recovery ID: %d\n", signature[64])

	// Call keeper method
	valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, signature)
	fmt.Printf("Verification result: %v\n", valid)

	require.True(t, valid, "Signature should verify")
}
