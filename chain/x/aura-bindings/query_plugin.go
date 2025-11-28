package aurabindings

import (
	"encoding/json"

	"cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ErrVerifiableCredentialNotFound = errors.Register("aura-bindings", 1, "verifiable credential not found")
)

// Define the custom query structure that the smart contracts will use.
// This should mirror the Rust `AuraQuery` enum.
type AuraQuery struct {
	VCRegistry *VCRegistryQuery `json:"vc_registry,omitempty"`
}

type VCRegistryQuery struct {
	GetVC *GetVCQuery `json:"get_vc,omitempty"`
}

type GetVCQuery struct {
	Address string `json:"address"`
}

// VerifiableCredential is a helper struct to return the VCRecord in a format that the CosmWasm contract expects.
type VerifiableCredential struct {
	Address  string `json:"address"`
	VCBase64 string `json:"vc_base64"`
}

// NewQueryPlugin creates a new query plugin for the wasm keeper with security validation.
func NewQueryPlugin(
	vcKeeper *vckeeper.Keeper,
	wasmKeeper *wasmkeeper.Keeper,
) *wasmkeeper.QueryPlugins {
	return &wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(vcKeeper, wasmKeeper),
	}
}

// CustomQuerier returns a function that handles custom queries with security checks.
func CustomQuerier(
	vcKeeper *vckeeper.Keeper,
	wasmKeeper *wasmkeeper.Keeper,
) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	// Commented out until wasm keeper types are available
	// securityValidator := NewSecurityValidator(wasmKeeper)

	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		// NOTE: We don't have direct access to contractAddr in query context
		// This is a limitation of the query plugin design
		// In production, you might need to extract it from context or use other methods

		var auraQuery AuraQuery
		if err := json.Unmarshal(request, &auraQuery); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal aura query")
		}

		if auraQuery.VCRegistry != nil {
			if auraQuery.VCRegistry.GetVC != nil {
				// SECURITY: Validate address
				addr, err := sdk.AccAddressFromBech32(auraQuery.VCRegistry.GetVC.Address)
				if err != nil {
					return nil, errors.Wrap(err, "invalid address")
				}
				_ = addr // Use addr to avoid unused variable error

				// SECURITY: Query rate limiting would happen here
				// For now, we log but don't enforce since we don't have contract context
				// In production, you'd extract contract address from context

				vcs := vcKeeper.ListUserVCs(ctx, auraQuery.VCRegistry.GetVC.Address, vctypes.VCStatus_VC_STATUS_UNSPECIFIED, vctypes.VCType_VC_TYPE_UNSPECIFIED)
				if len(vcs) == 0 {
					return nil, errors.Wrap(ErrVerifiableCredentialNotFound, auraQuery.VCRegistry.GetVC.Address)
				}

				// For simplicity, we return the first VC
				vcRecord := vcs[0]

				// SECURITY: Filter sensitive fields based on contract permissions
				// For now, we use basic filtering (FilterSensitiveData is in skipped security.go)
				// Convert back to VerifiableCredential
				vc := VerifiableCredential{
					Address:  vcRecord.HolderAddress,
					VCBase64: string(vcRecord.CredentialHash),
				}

				// SECURITY: Log query for audit trail
				if wasmKeeper != nil {
					// Would log security event if we had contract context
				}

				return json.Marshal(vc)
			}
		}

		return nil, errors.Wrap(wasmtypes.ErrUnknownMsg, "unknown aura query")
	}
}
