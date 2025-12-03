package keeper

import (
	"context"
	"fmt"
	"time"
)

// RotateViewKey rotates a view key for enhanced security
// SECURITY: This method is DEPRECATED and should not be used.
// Key generation must happen client-side, not on the blockchain.
// The blockchain cannot securely generate private keys because:
// 1. All blockchain state is public and deterministic
// 2. Private keys generated on-chain would be visible to all validators
// 3. Anyone could regenerate the "private" key using the same deterministic process
//
// Instead, clients should:
// 1. Generate new key pairs client-side using secure random number generators
// 2. Register only the new public key on-chain via MsgRegisterViewKey
// 3. Revoke old public keys via MsgRevokeViewKey
func (k Keeper) RotateViewKey(ctx context.Context, owner string) error {
	return fmt.Errorf("RotateViewKey is deprecated: key generation must be client-side only")
}

// DEPRECATED: generateNewPublicKey should never be called.
// Public keys must be generated client-side from secure private keys.
func (k Keeper) generateNewPublicKey(ctx context.Context, owner string) []byte {
	// SECURITY VIOLATION: This function should not exist.
	// Key generation on-chain is fundamentally insecure.
	panic("generateNewPublicKey is deprecated: keys must be generated client-side")
}

// DEPRECATED: generateNewPrivateKey should never be called.
// Private keys must NEVER be generated on the blockchain.
func (k Keeper) generateNewPrivateKey(ctx context.Context, owner string) []byte {
	// SECURITY VIOLATION: This function should not exist.
	// Generating private keys on-chain is a critical security vulnerability.
	// All blockchain state is public and anyone can see these "private" keys.
	panic("generateNewPrivateKey is deprecated: private keys must NEVER be generated on-chain")
}

func (k Keeper) markKeyAsRotated(ctx context.Context, owner string, publicKey []byte) {
	store := k.getStore(ctx)
	key := append([]byte("rotated_key_"), publicKey...)
	store.Set(key, []byte(owner))
}

// ScheduleKeyRotation schedules automatic key rotation
// DEPRECATED: This function is deprecated because key rotation must be client-side.
// Clients should implement their own key rotation schedules and call MsgRegisterViewKey
// with newly generated keys when rotation is needed.
func (k Keeper) ScheduleKeyRotation(ctx context.Context, owner string, interval time.Duration) error {
	return fmt.Errorf("ScheduleKeyRotation is deprecated: key rotation must be client-side")
}
