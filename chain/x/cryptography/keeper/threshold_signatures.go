package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// CreateThresholdScheme creates a new threshold signature scheme.
//
// Security considerations:
//   - Threshold must be > 0 and <= total participants (enforced)
//   - Scheme ID is derived from creator and timestamp (collision-resistant)
//   - Public key is placeholder (production systems would derive from real DKG)
//   - Participant IDs must be unique (validated)
//
// This implementation provides the state management layer. Actual cryptographic
// threshold signature generation (Shamir Secret Sharing, BLS signatures, etc.)
// would be implemented in a production system.
//
// Parameters:
//   - ctx: SDK context for state access
//   - creator: Address creating the scheme
//   - threshold: Minimum signatures required (t in t-of-n)
//   - totalParticipants: Total number of participants (n in t-of-n)
//   - participantIDs: List of participant identifiers
//   - schemeType: Type of threshold signature scheme
//
// Returns:
//   - schemeID: Unique identifier for the scheme
//   - publicKey: Group public key (placeholder in this implementation)
//   - error: ErrInvalidInput if parameters are invalid
func (k Keeper) CreateThresholdScheme(
	ctx context.Context,
	creator string,
	threshold uint32,
	totalParticipants uint32,
	participantIDs []string,
	schemeType cryptoproto.ThresholdSchemeType,
) (string, []byte, error) {
	// Validate threshold parameters
	if threshold == 0 || totalParticipants == 0 {
		return "", nil, types.ErrInvalidInput.Wrap("threshold and totalParticipants must be > 0")
	}
	if threshold > totalParticipants {
		return "", nil, types.ErrInvalidInput.Wrap("threshold cannot exceed totalParticipants")
	}
	if uint32(len(participantIDs)) != totalParticipants {
		return "", nil, types.ErrInvalidInput.Wrapf("participantIDs length (%d) must equal totalParticipants (%d)",
			len(participantIDs), totalParticipants)
	}

	// Validate participant IDs are unique
	seen := make(map[string]bool)
	for _, id := range participantIDs {
		if id == "" {
			return "", nil, types.ErrInvalidInput.Wrap("participant ID cannot be empty")
		}
		if seen[id] {
			return "", nil, types.ErrInvalidInput.Wrapf("duplicate participant ID: %s", id)
		}
		seen[id] = true
	}

	// Generate scheme ID
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	schemeID := fmt.Sprintf("threshold_%s_%d", creator, blockTime.Unix())

	// Generate placeholder group public key
	// Production implementation would perform Distributed Key Generation (DKG)
	// and derive the public key from the combined participant contributions
	publicKey := k.generatePlaceholderThresholdPublicKey(schemeID, threshold, totalParticipants)

	// Create and store scheme
	scheme := &cryptoproto.ThresholdSignatureScheme{
		SchemeId:          schemeID,
		Threshold:         int32(threshold),
		TotalParticipants: int32(totalParticipants),
		ParticipantIds:    participantIDs,
		PublicKey:         publicKey,
		SchemeType:        schemeType,
		Status:            cryptoproto.ThresholdSchemeStatus_THRESHOLD_SCHEME_STATUS_ACTIVE,
		CreatedAt:         timestamppb.New(blockTime),
	}

	if err := k.SetThresholdScheme(ctx, scheme); err != nil {
		return "", nil, err
	}

	k.Logger(ctx).Info("created threshold signature scheme",
		"scheme_id", schemeID,
		"threshold", threshold,
		"total_participants", totalParticipants,
		"scheme_type", schemeType.String(),
	)

	return schemeID, publicKey, nil
}

// SubmitThresholdSignatureShare submits a signature share for threshold aggregation.
//
// Security considerations:
//   - Validates scheme exists and is active
//   - Checks participant is authorized for the scheme
//   - Prevents duplicate shares from same participant
//   - Aggregates shares when threshold is reached
//   - Verifies combined signature (placeholder in this implementation)
//
// This implementation provides state management. Production systems would:
//  1. Verify the share using the participant's verification key
//  2. Perform Lagrange interpolation when threshold is reached
//  3. Verify the combined signature against the group public key
//
// Parameters:
//   - ctx: SDK context for state access
//   - submitter: Address submitting the share
//   - schemeID: Identifier of the threshold scheme
//   - signatureShare: The signature share bytes
//   - messageHash: Hash of the message being signed
//
// Returns:
//   - sharesCollected: Total number of shares collected for this message
//   - thresholdReached: Whether the threshold has been reached
//   - combinedSignature: The combined signature (if threshold reached, else nil)
//   - error: Various errors for invalid inputs or state
func (k Keeper) SubmitThresholdSignatureShare(
	ctx context.Context,
	submitter string,
	schemeID string,
	signatureShare []byte,
	messageHash []byte,
) (uint32, bool, []byte, error) {
	// Retrieve scheme
	scheme, err := k.GetThresholdScheme(ctx, schemeID)
	if err != nil {
		return 0, false, nil, err
	}

	// Validate scheme is active
	if scheme.Status != cryptoproto.ThresholdSchemeStatus_THRESHOLD_SCHEME_STATUS_ACTIVE {
		return 0, false, nil, types.ErrInvalidInput.Wrapf("scheme %s is not active (status: %s)",
			schemeID, scheme.Status.String())
	}

	// Validate submitter is a participant
	isParticipant := false
	for _, pid := range scheme.ParticipantIds {
		if pid == submitter {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return 0, false, nil, types.ErrInvalidInput.Wrapf("submitter %s is not a participant in scheme %s",
			submitter, schemeID)
	}

	// Check if this participant has already submitted for this message
	existingShares := k.GetThresholdSignatureSharesForScheme(ctx, schemeID, messageHash)
	for _, share := range existingShares {
		if share.ParticipantId == submitter {
			return 0, false, nil, types.ErrInvalidInput.Wrapf("participant %s has already submitted a share for this message",
				submitter)
		}
	}

	// Create and store the signature share
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	share := &cryptoproto.ThresholdSignatureShare{
		SchemeId:       schemeID,
		ParticipantId:  submitter,
		SignatureShare: signatureShare,
		MessageHash:    messageHash,
		SignedAt:       timestamppb.New(blockTime),
	}

	if err := k.SetThresholdSignatureShare(ctx, share); err != nil {
		return 0, false, nil, err
	}

	// Get updated share count
	allShares := k.GetThresholdSignatureSharesForScheme(ctx, schemeID, messageHash)
	sharesCollected := uint32(len(allShares))

	k.Logger(ctx).Info("threshold signature share submitted",
		"scheme_id", schemeID,
		"participant", submitter,
		"shares_collected", sharesCollected,
		"threshold", scheme.Threshold,
	)

	// Check if threshold is reached
	thresholdReached := sharesCollected >= uint32(scheme.Threshold)
	var combinedSignature []byte

	if thresholdReached {
		// Combine shares into final signature
		// Production implementation would perform Lagrange interpolation
		// and verify the combined signature
		combinedSignature = k.combineThresholdSignatures(allShares, scheme)

		k.Logger(ctx).Info("threshold reached - signature combined",
			"scheme_id", schemeID,
			"shares_used", sharesCollected,
			"threshold", scheme.Threshold,
		)
	}

	return sharesCollected, thresholdReached, combinedSignature, nil
}

// generatePlaceholderThresholdPublicKey generates a deterministic placeholder public key.
// Production systems would perform Distributed Key Generation (DKG) and derive this
// from participant contributions.
func (k Keeper) generatePlaceholderThresholdPublicKey(schemeID string, threshold, totalParticipants uint32) []byte {
	// Create deterministic but unique key based on scheme parameters
	h := sha256.New()
	h.Write([]byte(schemeID))
	binary.Write(h, binary.BigEndian, threshold)
	binary.Write(h, binary.BigEndian, totalParticipants)
	return h.Sum(nil)
}

// combineThresholdSignatures combines signature shares into a final signature.
// This is a placeholder implementation. Production systems would:
//  1. Perform Lagrange interpolation on the shares
//  2. Combine shares according to the scheme type (e.g., BLS signature aggregation)
//  3. Verify the combined signature against the group public key
func (k Keeper) combineThresholdSignatures(
	shares []*cryptoproto.ThresholdSignatureShare,
	scheme *cryptoproto.ThresholdSignatureScheme,
) []byte {
	// Placeholder: Concatenate all shares and hash
	// Production implementation would perform proper cryptographic combination
	h := sha256.New()
	h.Write([]byte(scheme.SchemeId))
	for _, share := range shares {
		h.Write(share.SignatureShare)
		h.Write(share.MessageHash)
	}
	return h.Sum(nil)
}
