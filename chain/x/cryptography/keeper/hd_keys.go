package keeper

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

const (
	// BIP32 constants
	hardenedKeyStart = uint32(0x80000000)
	seedBytes        = 64
	chainCodeBytes   = 32
)

// DeriveHDKey derives a hierarchical deterministic key using BIP32/BIP44
func (k Keeper) DeriveHDKey(
	ctx context.Context,
	masterKeyID string,
	seedHash []byte, // Hash of seed for verification (never store actual seed)
	derivationPath string,
) (*cryptoproto.HDKeyDerivation, error) {
	if masterKeyID == "" {
		return nil, types.ErrInvalidKeyID
	}
	if len(seedHash) != 32 {
		return nil, types.ErrInvalidSeed
	}
	if derivationPath == "" {
		return nil, types.ErrInvalidDerivationPath
	}

	// Parse derivation path (e.g., "m/44'/118'/0'/0/0")
	pathParts, err := parseDerivationPath(derivationPath)
	if err != nil {
		return nil, err
	}

	// Extract metadata from BIP44 path
	metadata, err := extractHDKeyMetadata(pathParts)
	if err != nil {
		return nil, err
	}

	// In a real implementation, this would:
	// 1. Use the actual seed (from secure storage) to derive the key
	// 2. Apply BIP32 derivation for each path component
	// 3. Generate chain codes for each derivation level
	// For this implementation, we'll create a placeholder chain code

	// Generate a deterministic chain code based on the path
	chainCode, err := k.generateChainCode(masterKeyID, derivationPath)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	hdKey := &cryptoproto.HDKeyDerivation{
		MasterKeyId:    masterKeyID,
		SeedHash:       seedHash,
		DerivationPath: derivationPath,
		Depth:          int32(len(pathParts)),
		ChainCode:      chainCode,
		CreatedAt:      now,
		Metadata:       metadata,
	}

	// Store in state
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(hdKey)
	store.Set(types.GetHDKeyDerivationKey(masterKeyID), bz)

	k.Logger(ctx).Info("derived HD key",
		"master_key_id", masterKeyID,
		"path", derivationPath,
		"depth", len(pathParts),
	)

	return hdKey, nil
}

// DeriveChildKey derives a child key from a parent key using BIP32
func (k Keeper) DeriveChildKey(
	parentKey []byte,
	parentChainCode []byte,
	index uint32,
	hardened bool,
) ([]byte, []byte, error) {
	if len(parentKey) != 32 {
		return nil, nil, fmt.Errorf("invalid parent key length")
	}
	if len(parentChainCode) != chainCodeBytes {
		return nil, nil, fmt.Errorf("invalid chain code length")
	}

	// Apply hardened key derivation if requested
	if hardened {
		index |= hardenedKeyStart
	}

	// Prepare data for HMAC
	data := make([]byte, 37)
	if hardened {
		// For hardened keys: 0x00 || parent_key || index
		data[0] = 0
		copy(data[1:33], parentKey)
	} else {
		// For normal keys: parent_public_key || index
		// In a real implementation, derive public key from private key
		copy(data[0:33], parentKey) // Simplified
	}
	binary.BigEndian.PutUint32(data[33:37], index)

	// Generate HMAC-SHA512
	mac := hmac.New(sha512.New, parentChainCode)
	mac.Write(data)
	hmacResult := mac.Sum(nil)

	// Split the result
	childKey := hmacResult[:32]
	childChainCode := hmacResult[32:]

	return childKey, childChainCode, nil
}

// parseDerivationPath parses a BIP32/44 derivation path
func parseDerivationPath(path string) ([]uint32, error) {
	if !strings.HasPrefix(path, "m/") && !strings.HasPrefix(path, "M/") {
		return nil, types.ErrInvalidDerivationPath
	}

	// Remove the "m/" prefix
	path = path[2:]
	if path == "" {
		return []uint32{}, nil
	}

	parts := strings.Split(path, "/")
	indices := make([]uint32, len(parts))

	for i, part := range parts {
		hardened := false
		if strings.HasSuffix(part, "'") || strings.HasSuffix(part, "h") {
			hardened = true
			part = part[:len(part)-1]
		}

		index, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, types.ErrInvalidDerivationPath
		}

		if hardened {
			indices[i] = uint32(index) | hardenedKeyStart
		} else {
			indices[i] = uint32(index)
		}
	}

	return indices, nil
}

// extractHDKeyMetadata extracts BIP44 metadata from derivation path
func extractHDKeyMetadata(pathParts []uint32) (*cryptoproto.HDKeyMetadata, error) {
	metadata := &cryptoproto.HDKeyMetadata{}

	// BIP44 format: m / purpose' / coin_type' / account' / change / address_index
	if len(pathParts) >= 1 {
		purpose := pathParts[0] & ^hardenedKeyStart
		metadata.Purpose = fmt.Sprintf("%d", purpose)
	}
	if len(pathParts) >= 2 {
		coinType := pathParts[1] & ^hardenedKeyStart
		metadata.CoinType = fmt.Sprintf("%d", coinType)
	}
	if len(pathParts) >= 3 {
		account := pathParts[2] & ^hardenedKeyStart
		metadata.Account = int32(account)
	}
	if len(pathParts) >= 4 {
		change := pathParts[3]
		metadata.Change = int32(change)
	}
	if len(pathParts) >= 5 {
		addressIndex := pathParts[4]
		metadata.AddressIndex = int32(addressIndex)
	}

	return metadata, nil
}

// generateChainCode generates a deterministic chain code
func (k Keeper) generateChainCode(masterKeyID, derivationPath string) ([]byte, error) {
	// Use HMAC-SHA256 to generate a deterministic chain code
	key := []byte("aura_chain_code_v1")
	data := []byte(masterKeyID + ":" + derivationPath)

	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	chainCode := mac.Sum(nil)

	return chainCode, nil
}

// GetHDKeyDerivation retrieves an HD key derivation
func (k Keeper) GetHDKeyDerivation(ctx context.Context, masterKeyID string) (*cryptoproto.HDKeyDerivation, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetHDKeyDerivationKey(masterKeyID))
	if bz == nil {
		return nil, types.ErrHDKeyDerivationFailed
	}

	var hdKey cryptoproto.HDKeyDerivation
	k.cdc.MustUnmarshal(bz, &hdKey)

	return &hdKey, nil
}

// ValidateBIP44Path validates a BIP44 derivation path
func (k Keeper) ValidateBIP44Path(path string) error {
	pathParts, err := parseDerivationPath(path)
	if err != nil {
		return err
	}

	// BIP44 requires at least purpose and coin_type
	if len(pathParts) < 2 {
		return types.ErrInvalidDerivationPath
	}

	// Verify purpose is 44' (hardened)
	if pathParts[0] != (44 | hardenedKeyStart) {
		return fmt.Errorf("BIP44 requires purpose 44'")
	}

	// Verify coin_type is hardened
	if pathParts[1] < hardenedKeyStart {
		return fmt.Errorf("BIP44 requires hardened coin_type")
	}

	// If account is present, verify it's hardened
	if len(pathParts) >= 3 && pathParts[2] < hardenedKeyStart {
		return fmt.Errorf("BIP44 requires hardened account")
	}

	return nil
}
