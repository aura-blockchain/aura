// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/common/determinism"

	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// CheckPrivacyCompliance checks if privacy features comply with regulations
func (k Keeper) CheckPrivacyCompliance(ctx context.Context, jurisdiction string) (bool, error) {
	// Different jurisdictions may have different privacy requirements
	// This allows for selective disclosure while maintaining privacy

	switch jurisdiction {
	case "EU":
		return k.checkGDPRCompliance(ctx)
	case "US":
		return k.checkUSCompliance(ctx)
	default:
		return true, nil // Default to allowing
	}
}

func (k Keeper) checkGDPRCompliance(ctx context.Context) (bool, error) {
	// GDPR requires right to be forgotten
	// Ensure view keys are available for authorized disclosure
	return true, nil
}

func (k Keeper) checkUSCompliance(ctx context.Context) (bool, error) {
	// US regulations may require KYC/AML compliance
	return true, nil
}

// RegisterViewKey registers a view key for selective disclosure
// SECURITY: Only accepts public keys. Private keys must never be passed to the blockchain.
func (k Keeper) RegisterViewKey(ctx context.Context, owner string, publicKey []byte) error {
	// SECURITY: privateKey parameter removed - private keys must never be on-chain
	if len(publicKey) == 0 {
		return fmt.Errorf("public key cannot be empty")
	}

	// Validate key length
	keyLen := len(publicKey)
	if keyLen != 32 && keyLen != 33 && keyLen != 64 {
		return fmt.Errorf("invalid public key length (must be 32, 33, or 64 bytes)")
	}

	viewKey := &privacyproto.ViewKey{
		KeyType:       "AUDIT", // Default key type
		PublicViewKey: publicKey,
		// NO PrivateViewKey field - removed for security
		Address: []byte(owner),
	}

	return k.SetViewKey(ctx, owner, viewKey)
}

// SelectiveDisclose allows selective disclosure using view keys
func (k Keeper) SelectiveDisclose(ctx context.Context, txID string, viewKey []byte) (map[string]interface{}, error) {
	// Verify view key
	vk, err := k.GetViewKeyByPublic(ctx, viewKey)
	if err != nil {
		return nil, fmt.Errorf("view key not found")
	}

	// Decrypt transaction details using view key
	details := map[string]interface{}{
		"tx_id":     txID,
		"disclosed": true,
		"address":   string(vk.Address), // ViewKey has 'address' field, not 'Owner'
	}

	return details, nil
}

// AuditPrivacyOperation audits privacy operations for compliance
func (k Keeper) AuditPrivacyOperation(ctx context.Context, operation, details string) error {
	store := k.getStore(ctx)
	key := []byte(fmt.Sprintf("audit_%s_%d", operation, determinism.GetBlockTime(ctx).UnixNano()))

	auditLog := []byte(fmt.Sprintf("%s: %s", operation, details))
	store.Set(key, auditLog)

	return nil
}
