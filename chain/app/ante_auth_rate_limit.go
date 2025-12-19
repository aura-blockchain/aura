package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	walletsecuritykeeper "github.com/aequitas/aura/chain/x/walletsecurity/keeper"
	walletsecuritytypes "github.com/aequitas/aura/chain/x/walletsecurity/types"
)

// AuthRateLimitDecorator limits repeated authorization attempts per signer per block.
// This is a lightweight guard that relies on walletsecurity keeper storage to avoid
// in-memory nondeterminism.
type AuthRateLimitDecorator struct {
	walletKeeper *walletsecuritykeeper.Keeper
}

// NewAuthRateLimitDecorator creates a new rate limit decorator.
func NewAuthRateLimitDecorator(walletKeeper *walletsecuritykeeper.Keeper) sdk.AnteDecorator {
	return AuthRateLimitDecorator{walletKeeper: walletKeeper}
}

// AnteHandle enforces rate limits before signature verification.
func (d AuthRateLimitDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if d.walletKeeper == nil {
		return next(ctx, tx, simulate)
	}

	signers := make(map[string]struct{})
	for _, msg := range tx.GetMsgs() {
		msgWithSigners, ok := msg.(interface{ GetSigners() []sdk.AccAddress })
		if !ok {
			continue
		}
		for _, addr := range msgWithSigners.GetSigners() {
			signers[addr.String()] = struct{}{}
		}
	}

	for addrStr := range signers {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			return ctx, walletsecuritytypes.ErrUnauthorized
		}
		if err := d.walletKeeper.CheckAuthRateLimit(ctx, addr); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}
