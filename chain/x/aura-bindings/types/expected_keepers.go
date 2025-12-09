package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VCKeeper defines the expected interface for the VC registry keeper
type VCKeeper interface {
	// VerifyCredential verifies a verifiable credential
	VerifyCredential(ctx sdk.Context, vc interface{}) (bool, error)

	// IssueCredential issues a new verifiable credential
	IssueCredential(ctx sdk.Context, holder string, credentialType string, claims map[string]interface{}) (string, error)

	// RevokeCredential revokes a previously issued credential
	RevokeCredential(ctx sdk.Context, credentialID string) error

	// GetCredential retrieves a credential by ID
	GetCredential(ctx sdk.Context, credentialID string) (interface{}, error)
}
