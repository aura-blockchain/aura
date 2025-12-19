package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
)

// MaxAuthAttemptsPerBlock caps repeated auth attempts per signer per block to mitigate brute-force.
const MaxAuthAttemptsPerBlock = 5

// CheckAuthRateLimit increments and enforces per-block auth attempt limits for the signer.
func (k Keeper) CheckAuthRateLimit(ctx sdk.Context, signer sdk.AccAddress) error {
	if signer.Empty() {
		return types.ErrUnauthorized
	}

	store := k.getStore(ctx)
	key := types.GetAuthRateLimitKey(ctx.BlockHeight(), signer.String())

	countBz, err := store.Get(key)
	if err != nil {
		return fmt.Errorf("auth rate-limit store error: %w", err)
	}

	var count uint64
	if countBz != nil {
		count = sdk.BigEndianToUint64(countBz)
	}

	if count >= MaxAuthAttemptsPerBlock {
		return types.ErrAuthRateLimited
	}

	count++
	if err := store.Set(key, sdk.Uint64ToBigEndian(count)); err != nil {
		return fmt.Errorf("auth rate-limit store set error: %w", err)
	}

	return nil
}
