package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdksecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ripemd160"
)

// TestSimpleSignatureFlow tests the signature flow step by step
func TestSimpleSignatureFlow(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInput(t)
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, &ps, nil, nil, nil, nil)
	ctx := input.Ctx

	// Generate key pair
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	// Derive address manually (same logic as keeper)
	pubKeyBytes := pubKey.SerializeCompressed()
	fmt.Printf("Public key (compressed): %x\n", pubKeyBytes)

	sha256Hash := sha256.Sum256(pubKeyBytes)
	fmt.Printf("SHA256 hash: %x\n", sha256Hash)

	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	addressHash := ripemd160Hasher.Sum(nil)
	pawAddress := hex.EncodeToString(addressHash)
	fmt.Printf("PAW Address: %s\n", pawAddress)

	auraAddress := "aura1test"

	// Create message
	message := fmt.Sprintf("Link PAW address %s to Aura address %s", pawAddress, auraAddress)
	fmt.Printf("Message: %s\n", message)

	msgHash := sha256.Sum256([]byte(message))
	fmt.Printf("Message hash: %x\n", msgHash)

	// Sign using SignCompact - try with different settings
	compactSig := secp256k1ecdsa.SignCompact(privKey, msgHash[:], true) // compressed
	fmt.Printf("Compact sig (compressed) length: %d, recovery byte: %d\n", len(compactSig), compactSig[0])

	// Rearrange to [R][S][recovery_id] format
	signature := make([]byte, 65)
	copy(signature[0:64], compactSig[1:65]) // R and S
	signature[64] = compactSig[0] - 27      // Convert recovery byte

	fmt.Printf("Signature: R=%x S=%x RecoveryID=%d\n", signature[0:32], signature[32:64], signature[64])

	// Try recovery manually to debug
	recoveryID := signature[64]
	if recoveryID >= 27 {
		recoveryID -= 27
	}
	fmt.Printf("Recovery ID after adjustment: %d\n", recoveryID)

	// Try all recovery IDs to see which one works
	for recID := byte(0); recID <= 7; recID++ {
		fmt.Printf("\nTrying recovery ID: %d\n", recID)

		// Manually build compact signature for recovery
		testCompactSig := make([]byte, 65)
		testCompactSig[0] = byte(27 + recID)
		copy(testCompactSig[1:33], signature[0:32])   // R
		copy(testCompactSig[33:65], signature[32:64]) // S

		recoveredPubKey, _, err := secp256k1ecdsa.RecoverCompact(testCompactSig, msgHash[:])
		if err != nil {
			fmt.Printf("  Recovery failed: %v\n", err)
			continue
		}

		recoveredPubKeyBytes := recoveredPubKey.SerializeCompressed()
		fmt.Printf("  Recovered public key: %x\n", recoveredPubKeyBytes)

		// Derive address from recovered key
		recSha256 := sha256.Sum256(recoveredPubKeyBytes)
		recRipemd160 := ripemd160.New()
		recRipemd160.Write(recSha256[:])
		recAddressHash := recRipemd160.Sum(nil)
		recAddress := hex.EncodeToString(recAddressHash)

		fmt.Printf("  Recovered address: %s\n", recAddress)
		fmt.Printf("  Original address:  %s\n", pawAddress)
		fmt.Printf("  Match: %v\n", recAddress == pawAddress)

		if recAddress == pawAddress {
			fmt.Printf("  ✓ FOUND CORRECT RECOVERY ID: %d\n", recID)
		}
	}

	// Test the verification step manually
	fmt.Printf("\n=== Testing verification step ===\n")

	recoveredPubKeyBytes := pubKeyBytes // We know this is the correct key
	pubKeyObj := &sdksecp256k1.PubKey{Key: recoveredPubKeyBytes}
	fmt.Printf("Public key for verification: %x\n", recoveredPubKeyBytes)
	fmt.Printf("Signature for verification: %x\n", signature[:64])
	fmt.Printf("Message hash for verification: %x\n", msgHash)

	verified := pubKeyObj.VerifySignature(msgHash[:], signature[:64])
	fmt.Printf("VerifySignature result: %v\n", verified)

	// Now test with keeper
	fmt.Printf("\n=== Testing with keeper ===\n")
	valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, signature)
	fmt.Printf("Keeper verification result: %v\n", valid)

	require.True(t, valid, "Signature should verify")
}
