package keeper

import (
	"crypto/sha256"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// OffChainKYCData represents PII that should be stored off-chain
type OffChainKYCData struct {
	VerificationID string   `json:"verification_id"`
	Documents      []string `json:"documents"`
	Jurisdiction   string   `json:"jurisdiction"`
	RiskScore      string   `json:"risk_score"`
}

// ComputePIICommitment creates a SHA-256 hash of the PII data
func ComputePIICommitment(data OffChainKYCData) ([]byte, error) {
	// Serialize to canonical JSON (deterministic order via sort_keys)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(jsonBytes)
	return hash[:], nil
}

// TestSubmitKYC_WithPIICommitment tests GDPR-compliant KYC submission
func TestSubmitKYC_WithPIICommitment(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Simulate off-chain PII data
	offChainPII := OffChainKYCData{
		VerificationID: "KYC-2024-001234",
		Documents:      []string{"passport", "utility_bill"},
		Jurisdiction:   "US-CA",
		RiskScore:      "low",
	}

	// Compute commitment hash
	commitment, err := ComputePIICommitment(offChainPII)
	require.NoError(t, err)
	require.Len(t, commitment, 32, "SHA-256 hash must be 32 bytes")

	// Setup provider
	providerAddr := sdk.AccAddress([]byte("provider_address_1234")).String()
	params := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Submit KYC with commitment
	msg := &types.MsgSubmitKYC{
		Address:       "aura1test",
		KycLevel:      types.KYCLevel_KYC_LEVEL_ADVANCED,
		Provider:      providerAddr,
		PiiCommitment: commitment,
	}

	resp, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.True(t, resp.Success)
	require.Contains(t, resp.Message, "PII commitment")

	// Verify record was stored with commitment
	record, err := keeper.GetKYCRecord(ctx, msg.Address)
	require.NoError(t, err)
	require.Equal(t, commitment, record.PiiCommitment)
	require.Equal(t, types.KYCLevel_KYC_LEVEL_ADVANCED, record.KycLevel)
}

// TestSubmitKYC_InvalidCommitmentSize tests validation of commitment size
func TestSubmitKYC_InvalidCommitmentSize(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	providerAddr := sdk.AccAddress([]byte("provider_address_1234")).String()
	params := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	tests := []struct {
		name       string
		commitment []byte
	}{
		{
			name:       "empty commitment",
			commitment: []byte{},
		},
		{
			name:       "too short commitment",
			commitment: make([]byte, 16),
		},
		{
			name:       "too long commitment",
			commitment: make([]byte, 64),
		},
		{
			name:       "nil commitment",
			commitment: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &types.MsgSubmitKYC{
				Address:       "aura1test",
				KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      providerAddr,
				PiiCommitment: tt.commitment,
			}

			resp, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), msg)
			require.Error(t, err)
			require.Nil(t, resp)
			require.Contains(t, err.Error(), "pii_commitment must be 32 bytes")
		})
	}
}

// TestPIICommitmentVerification tests that commitments can be verified
func TestPIICommitmentVerification(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Original PII data
	originalPII := OffChainKYCData{
		VerificationID: "TEST-001",
		Documents:      []string{"passport", "driver_license"},
		Jurisdiction:   "EU-DE",
		RiskScore:      "medium",
	}

	// Compute commitment
	commitment, err := ComputePIICommitment(originalPII)
	require.NoError(t, err)

	// Setup provider
	providerAddr := sdk.AccAddress([]byte("provider_address_1234")).String()
	params := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Submit to blockchain
	msg := &types.MsgSubmitKYC{
		Address:       "aura1test",
		KycLevel:      types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		Provider:      providerAddr,
		PiiCommitment: commitment,
	}

	_, err = server.SubmitKYC(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)

	// Retrieve from blockchain
	record, err := keeper.GetKYCRecord(ctx, msg.Address)
	require.NoError(t, err)

	// Verify we can recompute the same commitment
	recomputedCommitment, err := ComputePIICommitment(originalPII)
	require.NoError(t, err)
	require.Equal(t, commitment, recomputedCommitment)
	require.Equal(t, record.PiiCommitment, recomputedCommitment)

	// Verify different PII produces different commitment
	differentPII := originalPII
	differentPII.RiskScore = "high"
	differentCommitment, err := ComputePIICommitment(differentPII)
	require.NoError(t, err)
	require.NotEqual(t, commitment, differentCommitment)
}

// TestEraseGDPRData tests GDPR Article 17 "Right to Erasure"
func TestEraseGDPRData(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Use valid bech32 addresses
	userAddr := sdk.AccAddress([]byte("user_address_12345")).String()
	providerAddr := sdk.AccAddress([]byte("provider_address_1234")).String()

	// First, submit a KYC record
	piiData := OffChainKYCData{
		VerificationID: "KYC-ERASE-TEST",
		Documents:      []string{"passport"},
		Jurisdiction:   "EU-FR",
		RiskScore:      "low",
	}
	commitment, err := ComputePIICommitment(piiData)
	require.NoError(t, err)

	params := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)

	submitMsg := &types.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: commitment,
		Jurisdiction:  "US", // Required: ISO 3166-1 alpha-2 country code
	}

	_, err = server.SubmitKYC(sdk.WrapSDKContext(ctx), submitMsg)
	require.NoError(t, err)

	// Now request data erasure
	eraseMsg := &types.MsgEraseGDPRData{
		Address:       userAddr,
		ErasureReason: "GDPR Article 17 - User requested data deletion",
	}

	resp, err := server.EraseGDPRData(sdk.WrapSDKContext(ctx), eraseMsg)
	require.NoError(t, err)
	require.True(t, resp.Success)
	require.NotEmpty(t, resp.ErasureEventId)
	require.Contains(t, resp.ErasureEventId, "gdpr-erasure")

	// Verify erasure event was emitted
	events := ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == "gdpr_data_erased" {
			found = true
			break
		}
	}
	require.True(t, found, "gdpr_data_erased event not emitted")

	// Verify on-chain record still exists (commitment preserved for audit)
	record, err := keeper.GetKYCRecord(ctx, userAddr)
	require.NoError(t, err)
	require.Equal(t, commitment, record.PiiCommitment, "commitment should remain for audit trail")
}

// TestCommitmentDeterminism verifies that commitment computation is deterministic
func TestCommitmentDeterminism(t *testing.T) {
	piiData := OffChainKYCData{
		VerificationID: "DET-TEST-001",
		Documents:      []string{"passport", "bill", "id"},
		Jurisdiction:   "US-NY",
		RiskScore:      "low",
	}

	// Compute commitment multiple times
	commitment1, err := ComputePIICommitment(piiData)
	require.NoError(t, err)

	commitment2, err := ComputePIICommitment(piiData)
	require.NoError(t, err)

	commitment3, err := ComputePIICommitment(piiData)
	require.NoError(t, err)

	// All should be identical
	require.Equal(t, commitment1, commitment2)
	require.Equal(t, commitment2, commitment3)

	// Should be exactly 32 bytes (SHA-256)
	require.Len(t, commitment1, 32)
}

// BenchmarkCommitmentComputation benchmarks the PII commitment computation
func BenchmarkCommitmentComputation(b *testing.B) {
	piiData := OffChainKYCData{
		VerificationID: "BENCH-001",
		Documents:      []string{"passport", "utility_bill", "bank_statement"},
		Jurisdiction:   "US-CA",
		RiskScore:      "medium",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ComputePIICommitment(piiData)
	}
}
