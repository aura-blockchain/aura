package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/aequitas/aura/chain/x/bridge/types"
	"github.com/aequitas/aura/chain/x/common/security"
)

// Keeper of the bridge store
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	// Dependencies
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	vcKeeper      types.VCRegistryKeeper // For shared identity verification

	// Security features
	reentrancyGuard *security.ReentrancyGuard
	pauseGuard      *security.PauseGuard
	inputValidator  *security.InputValidator
	safeMath        *security.SafeMath
	gasLimitGuard   *security.GasLimitGuard
	accessControl   *security.AccessControl
}

// NewKeeper creates a new bridge Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	vcKeeper types.VCRegistryKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		paramstore:    ps,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		vcKeeper:      vcKeeper,
		// Initialize security features
		reentrancyGuard: security.NewReentrancyGuard(),
		pauseGuard:      security.NewPauseGuard(""), // Admin will be set via governance
		inputValidator:  security.NewInputValidator(),
		safeMath:        security.NewSafeMath(),
		gasLimitGuard:   security.NewGasLimitGuard(1000000),    // 1M gas limit per tx
		accessControl:   security.NewAccessControl([]string{}), // Admins set via governance
	}
}

// GetParams returns the total set of bridge parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the bridge parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	// Only admins can set params
	caller := ctx.GetMsgSender()
	if !k.accessControl.IsAdmin(caller.String()) {
		return security.ErrUnauthorized
	}

	k.paramstore.SetParamSet(ctx, &params)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_params_updated",
			sdk.NewAttribute("updated_by", caller.String()),
		),
	)

	return nil
}

// Pause pauses the bridge module (emergency use only)
func (k Keeper) Pause(ctx sdk.Context, caller string) error {
	return k.pauseGuard.Pause(ctx, caller)
}

// Unpause unpauses the bridge module
func (k Keeper) Unpause(ctx sdk.Context, caller string) error {
	return k.pauseGuard.Unpause(ctx, caller)
}

// IsPaused checks if the bridge module is paused
func (k Keeper) IsPaused() bool {
	return k.pauseGuard.IsPaused()
}

// ============================================================================
// SHARED IDENTITY (Super-Compatibility with PAW & XAI)
// ============================================================================

// LinkCrossChainIdentity links an AURA address with PAW/XAI addresses
// Enables shared identity verification across all three chains
func (k Keeper) LinkCrossChainIdentity(
	ctx sdk.Context,
	auraAddress string,
	pawAddress string,
	xaiAddress string,
	signature []byte,
) error {
	// Check if module is paused
	if err := k.pauseGuard.CheckNotPaused(); err != nil {
		return err
	}

	// Reentrancy protection
	return k.reentrancyGuard.WithReentrancyGuard(func() error {
		// Input validation
		if err := k.inputValidator.ValidateAddress(auraAddress); err != nil {
			return fmt.Errorf("invalid aura address: %w", err)
		}
		if err := k.inputValidator.ValidateAddress(pawAddress); err != nil {
			return fmt.Errorf("invalid paw address: %w", err)
		}
		if err := k.inputValidator.ValidateAddress(xaiAddress); err != nil {
			return fmt.Errorf("invalid xai address: %w", err)
		}
		if err := k.inputValidator.ValidateSliceNotEmpty(signature, "signature"); err != nil {
			return err
		}

		// Verify user owns all addresses (signature verification)
		// In production, would verify signatures from all chains

		// Get or create shared identity
		identity := k.GetSharedIdentity(ctx, auraAddress)
		if identity == nil {
			identity = &types.SharedIdentity{
				AuraAddress:     auraAddress,
				VerifiedAura:    false,
				VerifiedPaw:     false,
				VerifiedXai:     false,
				LinkedAddresses: make(map[string]string),
			}
		}

		// Link addresses
		identity.LinkedAddresses["paw"] = pawAddress
		identity.LinkedAddresses["xai"] = xaiAddress

		// Check AURA verification status
		irScore := k.vcKeeper.GetIRScore(ctx, auraAddress)
		identity.AuraIrScore = irScore
		identity.VerifiedAura = irScore >= 100

		// Store
		k.SetSharedIdentity(ctx, identity)

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"identity_linked",
				sdk.NewAttribute("aura_address", auraAddress),
				sdk.NewAttribute("paw_address", pawAddress),
				sdk.NewAttribute("xai_address", xaiAddress),
			),
		)

		return nil
	})
}

// SyncVerificationStatus syncs verification status from PAW/XAI to AURA
// If user is verified on PAW or XAI, it can boost their AURA status
func (k Keeper) SyncVerificationStatus(
	ctx sdk.Context,
	auraAddress string,
	sourceChain string, // "paw" or "xai"
	verified bool,
	proof []byte, // Merkle proof of verification on source chain
) error {
	// Check if module is paused
	if err := k.pauseGuard.CheckNotPaused(); err != nil {
		return err
	}

	// Reentrancy protection
	return k.reentrancyGuard.WithReentrancyGuard(func() error {
		// Input validation
		if err := k.inputValidator.ValidateAddress(auraAddress); err != nil {
			return fmt.Errorf("invalid aura address: %w", err)
		}
		if err := k.inputValidator.ValidateString(sourceChain, "sourceChain"); err != nil {
			return err
		}
		if err := k.inputValidator.ValidateSliceNotEmpty(proof, "proof"); err != nil {
			return err
		}

		identity := k.GetSharedIdentity(ctx, auraAddress)
		if identity == nil {
			return fmt.Errorf("no linked identity found for address: %s", auraAddress)
		}

		// Verify proof (in production, verify Merkle proof against source chain state root)
		// For now, trust the input

		// Update verification status
		if sourceChain == "paw" {
			identity.VerifiedPaw = verified
		} else if sourceChain == "xai" {
			identity.VerifiedXai = verified
		} else {
			return fmt.Errorf("unknown source chain: %s", sourceChain)
		}

		// Store
		k.SetSharedIdentity(ctx, identity)

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"verification_synced",
				sdk.NewAttribute("aura_address", auraAddress),
				sdk.NewAttribute("source_chain", sourceChain),
				sdk.NewAttribute("verified", fmt.Sprintf("%t", verified)),
			),
		)

		// If verified on any chain, user gets benefits on AURA
		// (e.g., reduced fees, higher trust score, etc.)

		return nil
	})
}

// ============================================================================
// SHARED IDENTITY STORAGE
// ============================================================================

// GetSharedIdentity returns a shared identity by AURA address
func (k Keeper) GetSharedIdentity(ctx sdk.Context, auraAddress string) *types.SharedIdentity {
	store := ctx.KVStore(k.storeKey)
	key := types.SharedIdentityKey(auraAddress)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var identity types.SharedIdentity
	k.cdc.MustUnmarshal(bz, &identity)
	return &identity
}

// SetSharedIdentity stores a shared identity
func (k Keeper) SetSharedIdentity(ctx sdk.Context, identity *types.SharedIdentity) {
	store := ctx.KVStore(k.storeKey)
	key := types.SharedIdentityKey(identity.AuraAddress)

	bz := k.cdc.MustMarshal(identity)
	store.Set(key, bz)
}

// IsVerifiedOnAnyChain checks if user is verified on AURA, PAW, or XAI
func (k Keeper) IsVerifiedOnAnyChain(ctx sdk.Context, auraAddress string) bool {
	identity := k.GetSharedIdentity(ctx, auraAddress)
	if identity == nil {
		return false
	}

	return identity.VerifiedAura || identity.VerifiedPaw || identity.VerifiedXai
}
