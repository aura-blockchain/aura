package keeper

import (
	"errors"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// MockConfidenceScoreKeeper is a mock implementation of ConfidenceScoreKeeper for testing
type MockConfidenceScoreKeeper struct {
	userScores      map[string]uint64
	completedIRs    map[string]map[string]bool   // address -> ir_id -> bool
	anchorInfo      map[string]bool              // address -> bool
	arenaScores     map[string]map[string]uint64 // address -> arena -> score
	verifiedUsers   map[string]bool
	shouldFailArena bool
}

func NewMockConfidenceScoreKeeper() *MockConfidenceScoreKeeper {
	return &MockConfidenceScoreKeeper{
		userScores:    make(map[string]uint64),
		completedIRs:  make(map[string]map[string]bool),
		anchorInfo:    make(map[string]bool),
		arenaScores:   make(map[string]map[string]uint64),
		verifiedUsers: make(map[string]bool),
	}
}

func (m *MockConfidenceScoreKeeper) GetUserScore(walletAddr string) (uint64, bool) {
	score, ok := m.userScores[walletAddr]
	return score, ok
}

func (m *MockConfidenceScoreKeeper) SetUserScore(walletAddr string, score uint64) {
	m.userScores[walletAddr] = score
}

func (m *MockConfidenceScoreKeeper) HasCompletedIR(walletAddr, irID string) bool {
	if irs, ok := m.completedIRs[walletAddr]; ok {
		return irs[irID]
	}
	return false
}

func (m *MockConfidenceScoreKeeper) CompleteIR(walletAddr, irID string) {
	if m.completedIRs[walletAddr] == nil {
		m.completedIRs[walletAddr] = make(map[string]bool)
	}
	m.completedIRs[walletAddr][irID] = true
}

func (m *MockConfidenceScoreKeeper) GetAnchorInfo(walletAddr string) (interface{}, bool) {
	hasAnchor := m.anchorInfo[walletAddr]
	return nil, hasAnchor
}

func (m *MockConfidenceScoreKeeper) SetAnchorInfo(walletAddr string, hasAnchor bool) {
	m.anchorInfo[walletAddr] = hasAnchor
}

func (m *MockConfidenceScoreKeeper) GetArenaScore(walletAddr, arena string) (uint64, error) {
	if m.shouldFailArena {
		return 0, errors.New("arena lookup failed")
	}
	if scores, ok := m.arenaScores[walletAddr]; ok {
		return scores[arena], nil
	}
	return 0, nil
}

func (m *MockConfidenceScoreKeeper) SetArenaScore(walletAddr, arena string, score uint64) {
	if m.arenaScores[walletAddr] == nil {
		m.arenaScores[walletAddr] = make(map[string]uint64)
	}
	m.arenaScores[walletAddr][arena] = score
}

func (m *MockConfidenceScoreKeeper) IsVerified(walletAddr string) bool {
	return m.verifiedUsers[walletAddr]
}

func (m *MockConfidenceScoreKeeper) SetVerified(walletAddr string, verified bool) {
	m.verifiedUsers[walletAddr] = verified
}

// Helper function to create a test keeper with mock CS keeper
func setupTestKeeperWithMockCS(t *testing.T) (*Keeper, sdk.Context, *MockConfidenceScoreKeeper) {
	keeper, ctx := setupKeeperForTest(t)
	mockCS := NewMockConfidenceScoreKeeper()
	keeper.SetConfidenceScoreKeeper(mockCS)
	return keeper, ctx, mockCS
}

// Helper function to setup a valid VC policy
func setupTestPolicyWithCtx(keeper *Keeper, ctx sdk.Context, policyName string, status vcregistrypb.VCPolicyStatus) *vcregistrypb.VCPolicy {
	policy := &vcregistrypb.VCPolicy{
		VcTypeName:         policyName,
		Status:             status,
		Version:            "1",
		CsThreshold:        100,
		RequiredIrIds:      []string{"IR-001", "IR-002"},
		RequiredArena:      "degen_games",
		RequiredArenaScore: 50,
		Singleton:          false,
		ExpiryDurationDays: 365,
		CreatedAt:          timestamppb.Now(),
	}
	keeper.SetVCPolicy(ctx, *policy)
	return policy
}

// ============================================================================
// ValidateMintEligibility Tests
// ============================================================================

// TestValidateMintEligibility_Success tests when all requirements are met
func TestValidateMintEligibility_Success(t *testing.T) {
	keeper, ctx, mockCS := setupTestKeeperWithMockCS(t)

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	// Setup CS keeper with required data
	mockCS.SetUserScore(userAddr, 150)
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	// Setup policy
	setupTestPolicyWithCtx(keeper, ctx, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Test
	eligible, missing, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	require.NoError(t, err)
	assert.True(t, eligible)
	assert.Empty(t, missing)
}

// TestValidateMintEligibility_InsufficientCS tests CS below threshold
func TestValidateMintEligibility_InsufficientCS(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	// Setup with insufficient CS (50 < 100 required)
	mockCS.SetUserScore(userAddr, 50)
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Test
	eligible, missing, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	require.NoError(t, err)
	assert.False(t, eligible)
	assert.NotEmpty(t, missing)
	assert.Len(t, missing, 1)
	assert.Contains(t, missing[0], "insufficient confidence score")
}

// TestValidateMintEligibility_MissingIR tests required IR not completed
func TestValidateMintEligibility_MissingIR(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	mockCS.SetUserScore(userAddr, 150)
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	// Missing IR-002
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Test
	eligible, missing, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	require.NoError(t, err)
	assert.False(t, eligible)
	assert.NotEmpty(t, missing)
	found := false
	for _, req := range missing {
		if containsString(req, "IR-002") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected missing IR-002 requirement")
}

// TestValidateMintEligibility_AnchorNotCompleted tests anchor IR missing
func TestValidateMintEligibility_AnchorNotCompleted(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	mockCS.SetUserScore(userAddr, 150)
	// No anchor info
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Test
	eligible, missing, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	require.NoError(t, err)
	assert.False(t, eligible)
	assert.NotEmpty(t, missing)
	found := false
	for _, req := range missing {
		if containsString(req, "anchor IR") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected anchor IR requirement")
}

// TestValidateMintEligibility_ArenaScoreTooLow tests arena score requirement not met
func TestValidateMintEligibility_ArenaScoreTooLow(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	mockCS.SetUserScore(userAddr, 150)
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 25) // 25 < 50 required

	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Test
	eligible, missing, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	require.NoError(t, err)
	assert.False(t, eligible)
	assert.NotEmpty(t, missing)
	found := false
	for _, req := range missing {
		if containsString(req, "arena") && containsString(req, "degen_games") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected insufficient arena score requirement")
}

// TestValidateMintEligibility_SingletonViolation tests singleton constraint
func TestValidateMintEligibility_SingletonViolation(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	mockCS.SetUserScore(userAddr, 150)
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	// Setup policy with singleton constraint
	policy := setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)
	policy.Singleton = true
	keeper.SetVCPolicy(ctx, *policy)

	// Add existing active VC of same type for user
	existingVC := &vcregistrypb.VCRecord{
		VcId:          "vc:existing123",
		VcType:        vcType,
		HolderAddress: userAddr,
		Status:        vcregistrypb.VCStatus_VC_STATUS_ACTIVE,
	}
	keeper.SetVCRecord(ctx, *existingVC)

	// Test
	eligible, missing, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	require.NoError(t, err)
	assert.False(t, eligible)
	assert.NotEmpty(t, missing)
	found := false
	for _, req := range missing {
		if containsString(req, "singleton") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected singleton VC violation")
}

// TestValidateMintEligibility_RateLimitExceeded tests rate limit enforcement
func TestValidateMintEligibility_RateLimitExceeded(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	mockCS.SetUserScore(userAddr, 150)
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Set up rate limit: max 5 per day (default)
	// Increment count to hit limit
	for i := 0; i < 5; i++ {
		keeper.IncrementMintCount(ctx, userAddr)
	}

	// Test
	eligible, missing, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	require.NoError(t, err)
	assert.False(t, eligible)
	assert.NotEmpty(t, missing)
	found := false
	for _, req := range missing {
		if containsString(req, "rate limit") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected rate limit exceeded")
}

// ============================================================================
// MintVC Tests
// ============================================================================

// TestMintVC_Success tests successful VC minting
func TestMintVC_Success(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	userDID := "did:aura:user123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	// Setup user with all requirements met
	mockCS.SetUserScore(userAddr, 150)
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Register DID
	keeper.RegisterDID(ctx, userDID, userAddr, []*vcregistrypb.VerificationMethod{}, "")

	// Test
	metadata := map[string]string{
		"issued_at": "test",
	}
	vcID, err := keeper.MintVC(ctx, userAddr, userDID, vcType, "", metadata)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, vcID)
	assert.True(t, len(vcID) > 0)

	// Verify VC was stored
	storedVC, ok := keeper.GetVCRecord(ctx, vcID)
	assert.True(t, ok)
	assert.Equal(t, userAddr, storedVC.HolderAddress)
	assert.Equal(t, vcType, storedVC.VcType)
	assert.Equal(t, vcregistrypb.VCStatus_VC_STATUS_ACTIVE, storedVC.Status)
	assert.Equal(t, uint64(150), storedVC.CsAtMint)
}

// TestMintVC_NotEligible tests minting fails when user not eligible
func TestMintVC_NotEligible(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	userDID := "did:aura:user123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	// Setup user with insufficient CS
	mockCS.SetUserScore(userAddr, 50) // Below 100 threshold
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)
	keeper.RegisterDID(ctx, userDID, userAddr, []*vcregistrypb.VerificationMethod{}, "")

	// Test
	vcID, err := keeper.MintVC(ctx, userAddr, userDID, vcType, "", map[string]string{})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, vcID)
	assert.Contains(t, err.Error(), "not eligible")
}

// TestMintVC_PolicyNotFound tests minting fails when policy doesn't exist
func TestMintVC_PolicyNotFound(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	userDID := "did:aura:user123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	// Setup user
	mockCS.SetUserScore(userAddr, 150)
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	// Don't setup policy
	keeper.RegisterDID(ctx, userDID, userAddr, []*vcregistrypb.VerificationMethod{}, "")

	// Test
	vcID, err := keeper.MintVC(ctx, userAddr, userDID, vcType, "", map[string]string{})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, vcID)
	assert.Equal(t, types.ErrPolicyNotFound, err)
}

// TestMintVC_PolicyInactive tests minting fails when policy is inactive
func TestMintVC_PolicyInactive(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	userDID := "did:aura:user123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	mockCS.SetUserScore(userAddr, 150)
	mockCS.SetAnchorInfo(userAddr, true)
	mockCS.CompleteIR(userAddr, "IR-001")
	mockCS.CompleteIR(userAddr, "IR-002")
	mockCS.SetArenaScore(userAddr, "degen_games", 75)

	// Setup policy with DEPRECATED status
	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED)
	keeper.RegisterDID(ctx, userDID, userAddr, []*vcregistrypb.VerificationMethod{}, "")

	// Test
	vcID, err := keeper.MintVC(ctx, userAddr, userDID, vcType, "", map[string]string{})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, vcID)
	assert.Equal(t, types.ErrPolicyInactive, err)
}

// ============================================================================
// Edge Cases and Additional Tests
// ============================================================================

// TestValidateMintEligibility_InvalidHolderAddress tests with empty holder address
func TestValidateMintEligibility_InvalidHolderAddress(t *testing.T) {
	keeper, _ := setupTestKeeper()

	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	// Test
	eligible, _, err := keeper.ValidateMintEligibility(ctx, "", vcType)

	// Assert
	assert.Error(t, err)
	assert.False(t, eligible)
	assert.Equal(t, types.ErrInvalidHolderAddress, err)
}

// TestValidateMintEligibility_InvalidVCType tests with unspecified VC type
func TestValidateMintEligibility_InvalidVCType(t *testing.T) {
	keeper, _ := setupTestKeeper()

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_UNSPECIFIED

	// Test
	eligible, _, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	assert.Error(t, err)
	assert.False(t, eligible)
	assert.Equal(t, types.ErrInvalidVCType, err)
}

// TestValidateMintEligibility_NoCSKeeper tests when CS keeper is not set
func TestValidateMintEligibility_NoCSKeeper(t *testing.T) {
	store := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(store, "authority")
	// Don't set CS keeper
	keeper.SetCurrentHeight(1)
	keeper.SetCurrentTime(time.Now().Unix())

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Test
	eligible, _, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	assert.Error(t, err)
	assert.False(t, eligible)
	assert.Equal(t, types.ErrCSKeeperNotSet, err)
}

// TestMintVC_InvalidHolderAddress tests minting with empty holder address
func TestMintVC_InvalidHolderAddress(t *testing.T) {
	keeper, _ := setupTestKeeper()

	userDID := "did:aura:user123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	// Test
	vcID, err := keeper.MintVC(ctx, "", userDID, vcType, "", map[string]string{})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, vcID)
	assert.Equal(t, types.ErrInvalidHolderAddress, err)
}

// TestMintVC_InvalidDID tests minting with empty DID
func TestMintVC_InvalidDID(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	mockCS.SetUserScore(userAddr, 150)
	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Test
	vcID, err := keeper.MintVC(ctx, userAddr, "", vcType, "", map[string]string{})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, vcID)
	assert.Equal(t, types.ErrInvalidDID, err)
}

// TestMintVC_InvalidVCType tests minting with unspecified VC type
func TestMintVC_InvalidVCType(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	userDID := "did:aura:user123"

	mockCS.SetUserScore(userAddr, 150)
	keeper.RegisterDID(ctx, userDID, userAddr, []*vcregistrypb.VerificationMethod{}, "")

	// Test
	vcID, err := keeper.MintVC(ctx, userAddr, userDID, vcregistrypb.VCType_VC_TYPE_UNSPECIFIED, "", map[string]string{})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, vcID)
	assert.Equal(t, types.ErrInvalidVCType, err)
}

// TestValidateMintEligibility_MultipleFailures tests when multiple requirements are missing
func TestValidateMintEligibility_MultipleFailures(t *testing.T) {
	keeper, mockCS := setupTestKeeper()

	userAddr := "aura1testuser123"
	vcType := vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN

	// Setup user missing multiple requirements
	mockCS.SetUserScore(userAddr, 50) // Below threshold
	// No anchor
	mockCS.CompleteIR(userAddr, "IR-001")
	// Missing IR-002
	mockCS.SetArenaScore(userAddr, "degen_games", 25) // Below threshold

	setupTestPolicy(keeper, fmt.Sprintf("%d", vcType), vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE)

	// Test
	eligible, missing, err := keeper.ValidateMintEligibility(ctx, userAddr, vcType)

	// Assert
	require.NoError(t, err)
	assert.False(t, eligible)
	assert.NotEmpty(t, missing)
	// Should have multiple missing requirements
	assert.GreaterOrEqual(t, len(missing), 3)
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
