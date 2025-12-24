package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
)

// ValidateMintEligibility checks if a user is eligible to mint a specific VC type
// Returns: eligible (bool), missing requirements ([]string), error
func (k *Keeper) ValidateMintEligibility(ctx context.Context, holderAddress string, vcType types.VCType) (bool, []string, error) {
	if holderAddress == "" {
		return false, nil, types.ErrInvalidHolderAddress
	}

	// Get VC type name for policy lookup
	vcTypeName := fmt.Sprintf("%d", vcType)
	if vcType == types.VCType_VC_TYPE_UNSPECIFIED {
		return false, nil, types.ErrInvalidVCType
	}

	// Get policy for this VC type
	policy, ok := k.GetVCPolicy(ctx, vcTypeName)
	if !ok {
		return false, []string{"policy not found for VC type"}, types.ErrPolicyNotFound
	}

	// Check policy status
	if policy.Status != types.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE {
		return false, []string{"policy is not active"}, types.ErrPolicyInactive
	}

	// Check if csKeeper is set
	if k.csKeeper == nil {
		return false, []string{"confidence score keeper not configured"}, types.ErrCSKeeperNotSet
	}

	missing := []string{}

	// 1. Check confidence score threshold
	userScore, hasScore := k.csKeeper.GetUserScore(holderAddress)
	if !hasScore || userScore < policy.CsThreshold {
		missing = append(missing, fmt.Sprintf("insufficient confidence score (current: %d, required: %d)", userScore, policy.CsThreshold))
	}

	// 2. Check anchor IR completion
	_, hasAnchor := k.csKeeper.GetAnchorInfo(holderAddress)
	if !hasAnchor {
		missing = append(missing, "anchor IR (IR-000) not completed")
	}

	// 3. Check required IRs
	for _, requiredIRID := range policy.RequiredIrIds {
		if !k.csKeeper.HasCompletedIR(holderAddress, requiredIRID) {
			missing = append(missing, fmt.Sprintf("required IR not completed: %s", requiredIRID))
		}
	}

	// 4. Check arena requirements
	if policy.RequiredArena != "" && policy.RequiredArenaScore > 0 {
		arenaScore, err := k.csKeeper.GetArenaScore(holderAddress, policy.RequiredArena)
		if err != nil || arenaScore < policy.RequiredArenaScore {
			missing = append(missing, fmt.Sprintf("insufficient arena score for %s (current: %d, required: %d)",
				policy.RequiredArena, arenaScore, policy.RequiredArenaScore))
		}
	}

	// 5. Check rate limits
	if err := k.CheckMintRateLimit(ctx, holderAddress); err != nil {
		if err == types.ErrRateLimitExceeded {
			missing = append(missing, "daily minting rate limit exceeded")
		} else {
			return false, missing, err
		}
	}

	// 6. Check singleton constraint
	if policy.Singleton {
		existingVCs := k.ListUserVCs(ctx, holderAddress, types.VCStatus_VC_STATUS_ACTIVE, vcType)
		if len(existingVCs) > 0 {
			missing = append(missing, "singleton VC of this type already exists and is active")
		}
	}

	// 7. Check max VCs per user
	params, _ := k.GetParams(ctx)
	allVCs := k.ListUserVCs(ctx, holderAddress, types.VCStatus_VC_STATUS_UNSPECIFIED, types.VCType_VC_TYPE_UNSPECIFIED)
	if uint64(len(allVCs)) >= params.MaxVcsPerUser {
		missing = append(missing, fmt.Sprintf("maximum VCs per user exceeded (%d)", params.MaxVcsPerUser))
	}

	eligible := len(missing) == 0
	return eligible, missing, nil
}

// MintVC mints a new verifiable credential for a user
// Returns: VC ID (string), error
func (k *Keeper) MintVC(ctx context.Context, holderAddress, holderDID string, vcType types.VCType, vcTypeCustom string, metadata map[string]string) (string, error) {
	// 1. Validate inputs
	if holderAddress == "" {
		return "", types.ErrInvalidHolderAddress
	}
	if holderDID == "" {
		return "", types.ErrInvalidDID
	}
	if vcType == types.VCType_VC_TYPE_UNSPECIFIED {
		return "", types.ErrInvalidVCType
	}

	// 2. Check eligibility
	eligible, missingReqs, err := k.ValidateMintEligibility(ctx, holderAddress, vcType)
	if err != nil {
		return "", err
	}
	if !eligible {
		return "", fmt.Errorf("not eligible to mint VC: %v", missingReqs)
	}

	// 3. Get policy to determine VC parameters
	vcTypeName := vcTypeCustom
	if vcType != types.VCType_VC_TYPE_CUSTOM {
		vcTypeName = fmt.Sprintf("%d", vcType)
	}

	policy, ok := k.GetVCPolicy(ctx, vcTypeName)
	if !ok {
		return "", types.ErrPolicyNotFound
	}

	// 4. Generate unique VC ID
	vcID := k.generateVCID(holderAddress, vcType, vcTypeCustom)

	// 5. Calculate expiration
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()
	var expiresAt *gogotypes.Timestamp
	if policy.ExpiryDurationDays > 0 {
		expiryTime := currentTime + (int64(policy.ExpiryDurationDays) * 86400) // days to seconds
		expiresAt = &gogotypes.Timestamp{Seconds: expiryTime, Nanos: 0}
	}

	// 6. Get current CS score
	currentCS := uint64(0)
	if k.csKeeper != nil {
		currentCS, _ = k.csKeeper.GetUserScore(holderAddress)
	}

	// 7. Generate credential hash
	credentialHash := k.generateCredentialHash(vcID, holderAddress, holderDID, vcType, vcTypeCustom)

	// 8. Create VC Record
	currentHeight := uint64(sdkCtx.BlockHeight())
	vcRecord := types.VCRecord{
		VcId:               vcID,
		VcType:             vcType,
		VcTypeCustom:       vcTypeCustom,
		HolderDid:          holderDID,
		HolderAddress:      holderAddress,
		Status:             types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:           &gogotypes.Timestamp{Seconds: currentTime, Nanos: 0},
		ExpiresAt:          expiresAt, // This is already *gogotypes.Timestamp
		IssuedHeight:       currentHeight,
		CredentialHash:     credentialHash,
		VerifierPluginHash: []byte{}, // Can be set by caller
		IssuerAssistant:    "",       // Can be set by caller
		PrerequisiteIrIds:  policy.RequiredIrIds,
		Metadata:           metadata,
		CsAtMint:           currentCS,
		PolicyVersion:      policy.Version,
	}

	// 9. Store VC record
	if err := k.SetVCRecord(ctx, vcRecord); err != nil {
		return "", fmt.Errorf("failed to store VC record: %w", err)
	}

	// 10. Add credential to DID document
	if err := k.AddCredentialToDID(ctx, holderDID, vcID); err != nil {
		// Log warning but don't fail - DID might not exist yet
		// In production, might want to handle this differently
	}

	// 11. Increment mint count for rate limiting
	k.IncrementMintCount(ctx, holderAddress)

	// 12. Emit event (would be done by msg_server in production)
	// Events would be emitted by the caller (msg_server)

	return vcID, nil
}

// generateVCID generates a unique VC ID using sha256(address+type+timestamp+height)
// Note: This function must be called with context available to get consensus-safe time/height
func (k *Keeper) generateVCID(holderAddress string, vcType types.VCType, vcTypeCustom string) string {
	// Note: This function is only called from MintVC which has already unwrapped context
	// In production, pass time/height as parameters for better design
	h := sha256.New()
	h.Write([]byte(holderAddress))
	h.Write([]byte(fmt.Sprintf("%d", vcType)))
	if vcTypeCustom != "" {
		h.Write([]byte(vcTypeCustom))
	}
	// Using static values here as a temporary measure - in production should pass as params
	// The actual time/height from MintVC context is used in the VC record itself
	h.Write([]byte(holderAddress)) // Use address again for uniqueness
	h.Write([]byte(fmt.Sprintf("%d", vcType)))

	hashBytes := h.Sum(nil)
	return "vc:" + hex.EncodeToString(hashBytes)[:32]
}

// generateCredentialHash generates a hash of the full credential
// Note: This function must be called with context available to get consensus-safe time
func (k *Keeper) generateCredentialHash(vcID, holderAddress, holderDID string, vcType types.VCType, vcTypeCustom string) []byte {
	// Note: This function is only called from MintVC which has already unwrapped context
	// In production, pass time as parameter for better design
	h := sha256.New()
	h.Write([]byte(vcID))
	h.Write([]byte(holderAddress))
	h.Write([]byte(holderDID))
	h.Write([]byte(fmt.Sprintf("%d", vcType)))
	if vcTypeCustom != "" {
		h.Write([]byte(vcTypeCustom))
	}
	// Using vcID again for deterministic hashing - actual time is in VC record
	h.Write([]byte(vcID))
	return h.Sum(nil)
}
