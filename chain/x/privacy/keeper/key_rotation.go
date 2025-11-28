package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/common/determinism"
	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// RotateViewKey rotates a view key for enhanced security
func (k Keeper) RotateViewKey(ctx context.Context, owner string) error {
	// Get existing view keys
	existingKeys := k.GetViewKeys(ctx, owner)

	// Generate new key pair
	newPublicKey := k.generateNewPublicKey(ctx, owner)
	newPrivateKey := k.generateNewPrivateKey(ctx, owner)

	// Parse owner address
	ownerAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return err
	}

	// Create new view key
	newViewKey := &privacyproto.ViewKey{
		Address:        ownerAddr,
		PublicViewKey:  newPublicKey,
		PrivateViewKey: newPrivateKey,
	}

	// Store new view key
	if err := k.SetViewKey(ctx, owner, newViewKey); err != nil {
		return err
	}

	// Mark old keys as rotated (keep for audit trail)
	for _, oldKey := range existingKeys {
		k.markKeyAsRotated(ctx, owner, oldKey.PublicViewKey)
	}

	return nil
}

func (k Keeper) generateNewPublicKey(ctx context.Context, owner string) []byte {
	// In production, use proper key generation
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	data := fmt.Sprintf("%s_%d_public", owner, determinism.GetBlockTime(sdkCtx).UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

func (k Keeper) generateNewPrivateKey(ctx context.Context, owner string) []byte {
	// In production, use proper key generation
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	data := fmt.Sprintf("%s_%d_private", owner, determinism.GetBlockTime(sdkCtx).UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

func (k Keeper) markKeyAsRotated(ctx context.Context, owner string, publicKey []byte) {
	store := k.getStore(ctx)
	key := append([]byte("rotated_key_"), publicKey...)
	store.Set(key, []byte(owner))
}

// ScheduleKeyRotation schedules automatic key rotation
func (k Keeper) ScheduleKeyRotation(ctx context.Context, owner string, interval time.Duration) error {
	store := k.getStore(ctx)
	key := []byte(fmt.Sprintf("rotation_schedule_%s", owner))

	schedule := []byte(fmt.Sprintf("%.0f", interval.Seconds()))
	store.Set(key, schedule)

	return nil
}
