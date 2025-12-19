package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// AuthKeeper defines the subset of auth keeper methods walletsecurity depends on.
type AuthKeeper interface {
	GetSession(ctx sdk.Context, sessionID string) (*authproto.Session, error)
}
