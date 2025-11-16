package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ValidateMintEligibility checks if a user is eligible to mint a specific VC type
// Returns: eligible (bool), missing requirements ([]string), error
func (k *Keeper) ValidateMintEligibility(holderAddress string, vcType vcregistrypb.VCType) (bool, []string, error) {
	if holderAddress == "" {
		return false, nil, types.ErrInvalidHolderAddress
	}

	// Get VC type name for policy lookup
	vcTypeName := fmt.Sprintf("%d", vcType)
	if vcType == vcregistrypb.VCType_VC_TYPE_UNSPECIFIED {
		return false, nil, types.ErrInvalidVCType
	}

	// Get policy for this VC type
	policy, ok := k.GetVCPolicy(vcTypeName)
	if !ok {
		return false, []string{"policy not found for VC type"}, types.ErrPolicyNotFound
	}

	// Check policy status
	if policy.Status != vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE {
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
	if err := k.CheckMintRateLimit(holderAddress); err != nil {
		if err == types.ErrRateLimitExceeded {
			missing = append(missing, "daily minting rate limit exceeded")
		} else {
			return false, missing, err
		}
	}

	// 6. Check singleton constraint
	if policy.Singleton {
		existingVCs := k.ListUserVCs(holderAddress, vcregistrypb.VCStatus_VC_STATUS_ACTIVE, vcType)
		if len(existingVCs) > 0 {
			missing = append(missing, "singleton VC of this type already exists and is active")
		}
	}

	// 7. Check max VCs per user
	params := k.GetParams()
	allVCs := k.ListUserVCs(holderAddress, vcregistrypb.VCStatus_VC_STATUS_UNSPECIFIED, vcregistrypb.VCType_VC_TYPE_UNSPECIFIED)
	if uint64(len(allVCs)) >= params.MaxVcsPerUser {
		missing = append(missing, fmt.Sprintf("maximum VCs per user exceeded (%d)", params.MaxVcsPerUser))
	}

	eligible := len(missing) == 0
	return eligible, missing, nil
}

// MintVC mints a new verifiable credential for a user
// Returns: VC ID (string), error
func (k *Keeper) MintVC(holderAddress, holderDID string, vcType vcregistrypb.VCType, vcTypeCustom string, metadata map[string]string) (string, error) {
	// 1. Validate inputs
	if holderAddress == "" {
		return "", types.ErrInvalidHolderAddress
	}
	if holderDID == "" {
		return "", types.ErrInvalidDID
	}
	if vcType == vcregistrypb.VCType_VC_TYPE_UNSPECIFIED {
		return "", types.ErrInvalidVCType
	}

	// 2. Check eligibility
	eligible, missingReqs, err := k.ValidateMintEligibility(holderAddress, vcType)
	if err != nil {
		return "", err
	}
	if !eligible {
		return "", fmt.Errorf("not eligible to mint VC: %v", missingReqs)
	}

	// 3. Get policy to determine VC parameters
	vcTypeName := vcTypeCustom
	if vcType != vcregistrypb.VCType_VC_TYPE_CUSTOM {
		vcTypeName = fmt.Sprintf("%d", vcType)
	}

	policy, ok := k.GetVCPolicy(vcTypeName)
	if !ok {
		return "", types.ErrPolicyNotFound
	}

	// 4. Generate unique VC ID
	vcID := k.generateVCID(holderAddress, vcType, vcTypeCustom)

	// 5. Calculate expiration
	var expiresAt *timestamppb.Timestamp
	if policy.ExpiryDurationDays > 0 {
		expiryTime := k.currentTime + (int64(policy.ExpiryDurationDays) * 86400) // days to seconds
		expiresAt = timestamppb.New(timestamppb.New(unixToTime(expiryTime)).AsTime())
	}

	// 6. Get current CS score
	currentCS := uint64(0)
	if k.csKeeper != nil {
		currentCS, _ = k.csKeeper.GetUserScore(holderAddress)
	}

	// 7. Generate credential hash
	credentialHash := k.generateCredentialHash(vcID, holderAddress, holderDID, vcType, vcTypeCustom)

	// 8. Create VC Record
	vcRecord := vcregistrypb.VCRecord{
		VcId:               vcID,
		VcType:             vcType,
		VcTypeCustom:       vcTypeCustom,
		HolderDid:          holderDID,
		HolderAddress:      holderAddress,
		Status:             vcregistrypb.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:           timestamppb.New(unixToTime(k.currentTime)),
		ExpiresAt:          expiresAt,
		IssuedHeight:       k.currentHeight,
		CredentialHash:     credentialHash,
		VerifierPluginHash: []byte{}, // Can be set by caller
		IssuerAssistant:    "",       // Can be set by caller
		PrerequisiteIrIds:  policy.RequiredIrIds,
		Metadata:           metadata,
		CsAtMint:           currentCS,
		PolicyVersion:      policy.Version,
	}

	// 9. Store VC record
	if err := k.SetVCRecord(vcRecord); err != nil {
		return "", fmt.Errorf("failed to store VC record: %w", err)
	}

	// 10. Add credential to DID document
	if err := k.AddCredentialToDID(holderDID, vcID); err != nil {
		// Log warning but don't fail - DID might not exist yet
		// In production, might want to handle this differently
	}

	// 11. Increment mint count for rate limiting
	k.IncrementMintCount(holderAddress)

	// 12. Emit event (would be done by msg_server in production)
	// Events would be emitted by the caller (msg_server)

	return vcID, nil
}

// generateVCID generates a unique VC ID using sha256(address+type+timestamp+height)
func (k *Keeper) generateVCID(holderAddress string, vcType vcregistrypb.VCType, vcTypeCustom string) string {
	h := sha256.New()
	h.Write([]byte(holderAddress))
	h.Write([]byte(fmt.Sprintf("%d", vcType)))
	if vcTypeCustom != "" {
		h.Write([]byte(vcTypeCustom))
	}
	h.Write([]byte(fmt.Sprintf("%d", k.currentTime)))
	h.Write([]byte(fmt.Sprintf("%d", k.currentHeight)))

	hashBytes := h.Sum(nil)
	return "vc:" + hex.EncodeToString(hashBytes)[:32]
}

// generateCredentialHash generates a hash of the full credential
func (k *Keeper) generateCredentialHash(vcID, holderAddress, holderDID string, vcType vcregistrypb.VCType, vcTypeCustom string) []byte {
	h := sha256.New()
	h.Write([]byte(vcID))
	h.Write([]byte(holderAddress))
	h.Write([]byte(holderDID))
	h.Write([]byte(fmt.Sprintf("%d", vcType)))
	if vcTypeCustom != "" {
		h.Write([]byte(vcTypeCustom))
	}
	h.Write([]byte(fmt.Sprintf("%d", k.currentTime)))
	return h.Sum(nil)
}

// unixToTime converts unix timestamp to time.Time
func unixToTime(unix int64) time.Time {
	if unix == 0 {
		return time.Time{}
	}
	return time.Unix(unix, 0)
}
