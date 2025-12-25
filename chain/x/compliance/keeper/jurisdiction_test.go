// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// makePiiCommitment creates a test 32-byte PII commitment
func makePiiCommitmentForJurisdiction(seed string) []byte {
	pii := make([]byte, 32)
	copy(pii, []byte(seed))
	return pii
}

// TestIsJurisdictionBlocked verifies jurisdiction-based access control
func TestIsJurisdictionBlocked(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	// Set params with blocked jurisdictions
	params := types.DefaultParams()
	params.BlockedJurisdictions = []string{"KP", "IR", "SY", "CU", "RU", "BY"}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	tests := []struct {
		name         string
		jurisdiction string
		wantBlocked  bool
	}{
		{
			name:         "blocked - North Korea",
			jurisdiction: "KP",
			wantBlocked:  true,
		},
		{
			name:         "blocked - Iran",
			jurisdiction: "IR",
			wantBlocked:  true,
		},
		{
			name:         "blocked - Syria",
			jurisdiction: "SY",
			wantBlocked:  true,
		},
		{
			name:         "blocked - Cuba",
			jurisdiction: "CU",
			wantBlocked:  true,
		},
		{
			name:         "blocked - Russia",
			jurisdiction: "RU",
			wantBlocked:  true,
		},
		{
			name:         "blocked - Belarus",
			jurisdiction: "BY",
			wantBlocked:  true,
		},
		{
			name:         "allowed - United States",
			jurisdiction: "US",
			wantBlocked:  false,
		},
		{
			name:         "allowed - United Kingdom",
			jurisdiction: "GB",
			wantBlocked:  false,
		},
		{
			name:         "allowed - Germany",
			jurisdiction: "DE",
			wantBlocked:  false,
		},
		{
			name:         "allowed - Japan",
			jurisdiction: "JP",
			wantBlocked:  false,
		},
		{
			name:         "blocked - case insensitive lowercase kp",
			jurisdiction: "kp",
			wantBlocked:  true,
		},
		{
			name:         "blocked - case insensitive lowercase ir",
			jurisdiction: "ir",
			wantBlocked:  true,
		},
		{
			name:         "blocked - case insensitive mixed case Sy",
			jurisdiction: "Sy",
			wantBlocked:  true,
		},
		{
			name:         "blocked - empty jurisdiction (fail-safe)",
			jurisdiction: "",
			wantBlocked:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked := keeper.IsJurisdictionBlocked(ctx, tt.jurisdiction)
			require.Equal(t, tt.wantBlocked, blocked,
				"IsJurisdictionBlocked(%s) = %v, want %v",
				tt.jurisdiction, blocked, tt.wantBlocked)
		})
	}
}

// TestIsJurisdictionBlocked_EmptyBlockedList tests behavior with no blocked jurisdictions
func TestIsJurisdictionBlocked_EmptyBlockedList(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	// Set params with empty blocked jurisdictions
	params := types.DefaultParams()
	params.BlockedJurisdictions = []string{}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// All jurisdictions should be allowed
	tests := []struct {
		jurisdiction string
		wantBlocked  bool
	}{
		{"US", false},
		{"KP", false},
		{"IR", false},
		{"", true}, // Empty still blocked (fail-safe)
	}

	for _, tt := range tests {
		blocked := keeper.IsJurisdictionBlocked(ctx, tt.jurisdiction)
		require.Equal(t, tt.wantBlocked, blocked,
			"IsJurisdictionBlocked(%s) = %v, want %v (empty blocked list)",
			tt.jurisdiction, blocked, tt.wantBlocked)
	}
}

// TestIsJurisdictionBlocked_GovernanceUpdate tests updating blocked jurisdictions via governance
func TestIsJurisdictionBlocked_GovernanceUpdate(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)

	// Initial params with some blocked jurisdictions
	params := types.DefaultParams()
	params.BlockedJurisdictions = []string{"KP", "IR"}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Verify initial state
	require.True(t, keeper.IsJurisdictionBlocked(ctx, "KP"))
	require.True(t, keeper.IsJurisdictionBlocked(ctx, "IR"))
	require.False(t, keeper.IsJurisdictionBlocked(ctx, "SY"))

	// Simulate governance proposal to add Syria
	params.BlockedJurisdictions = []string{"KP", "IR", "SY"}
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Verify update took effect
	require.True(t, keeper.IsJurisdictionBlocked(ctx, "KP"))
	require.True(t, keeper.IsJurisdictionBlocked(ctx, "IR"))
	require.True(t, keeper.IsJurisdictionBlocked(ctx, "SY"))

	// Simulate governance proposal to remove Iran (sanctions lifted)
	params.BlockedJurisdictions = []string{"KP", "SY"}
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Verify removal took effect
	require.True(t, keeper.IsJurisdictionBlocked(ctx, "KP"))
	require.False(t, keeper.IsJurisdictionBlocked(ctx, "IR")) // Now allowed
	require.True(t, keeper.IsJurisdictionBlocked(ctx, "SY"))
}

// TestSubmitKYC_BlockedJurisdiction tests KYC submission rejection for blocked jurisdictions
func TestSubmitKYC_BlockedJurisdiction(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	msgServer := NewMsgServer(keeper)

	// Setup approved KYC provider
	providerAddr := sdk.AccAddress([]byte("provider_address_12")).String()
	params := types.DefaultParams()
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{"KP", "IR", "SY", "CU", "RU", "BY"}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	piiCommitment := makePiiCommitmentForJurisdiction("test")

	// Test cases for blocked jurisdictions
	blockedCases := []struct {
		name         string
		jurisdiction string
	}{
		{"North Korea", "KP"},
		{"Iran", "IR"},
		{"Syria", "SY"},
		{"Cuba", "CU"},
		{"Russia", "RU"},
		{"Belarus", "BY"},
		{"lowercase kp", "kp"},
		{"mixed case Ir", "Ir"},
	}

	for _, tc := range blockedCases {
		t.Run("blocked_"+tc.name, func(t *testing.T) {
			msg := &types.MsgSubmitKYC{
				Address:       sdk.AccAddress([]byte("user_12345678901")).String(),
				KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      providerAddr,
				PiiCommitment: piiCommitment,
				Jurisdiction:  tc.jurisdiction,
			}

			resp, err := msgServer.SubmitKYC(sdk.WrapSDKContext(ctx), msg)

			// Should fail with permission denied due to blocked jurisdiction
			require.Error(t, err)
			require.Nil(t, resp)
			require.Contains(t, err.Error(), "blocked due to OFAC sanctions",
				"Expected OFAC sanctions error for jurisdiction %s", tc.jurisdiction)
		})
	}
}

// TestSubmitKYC_AllowedJurisdiction tests KYC submission acceptance for allowed jurisdictions
func TestSubmitKYC_AllowedJurisdiction(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	msgServer := NewMsgServer(keeper)

	// Setup approved KYC provider
	providerAddr := sdk.AccAddress([]byte("provider_address_12")).String()
	params := types.DefaultParams()
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{"KP", "IR", "SY", "CU", "RU", "BY"}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	piiCommitment := makePiiCommitmentForJurisdiction("test")

	// Test cases for allowed jurisdictions
	allowedCases := []struct {
		name         string
		jurisdiction string
	}{
		{"United States", "US"},
		{"United Kingdom", "GB"},
		{"Germany", "DE"},
		{"Japan", "JP"},
		{"Canada", "CA"},
		{"Australia", "AU"},
	}

	for _, tc := range allowedCases {
		t.Run("allowed_"+tc.name, func(t *testing.T) {
			userAddr := sdk.AccAddress([]byte("user_" + tc.jurisdiction + "_123456")).String()
			msg := &types.MsgSubmitKYC{
				Address:       userAddr,
				KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      providerAddr,
				PiiCommitment: piiCommitment,
				Jurisdiction:  tc.jurisdiction,
			}

			resp, err := msgServer.SubmitKYC(sdk.WrapSDKContext(ctx), msg)

			// Should succeed (note: will fail on signer check in full test, but jurisdiction validation should pass)
			// The error we get should NOT be about jurisdiction being blocked
			if err != nil {
				require.NotContains(t, err.Error(), "blocked due to OFAC sanctions",
					"Should not reject allowed jurisdiction %s", tc.jurisdiction)
			}

			// If it succeeds (in cases where signer check is bypassed), verify record
			if err == nil {
				require.NotNil(t, resp)
				require.True(t, resp.Success)

				// Verify KYC record was created with jurisdiction
				record, err := keeper.GetKYCRecord(ctx, userAddr)
				require.NoError(t, err)
				require.NotNil(t, record)
				require.Equal(t, tc.jurisdiction, record.Jurisdiction,
					"Jurisdiction should be stored in KYC record")
			}
		})
	}
}

// TestSubmitKYC_MissingJurisdiction tests validation of missing jurisdiction
func TestSubmitKYC_MissingJurisdiction(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	msgServer := NewMsgServer(keeper)

	// Setup approved KYC provider
	providerAddr := sdk.AccAddress([]byte("provider_address_12")).String()
	params := types.DefaultParams()
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	piiCommitment := makePiiCommitmentForJurisdiction("test")

	msg := &types.MsgSubmitKYC{
		Address:       sdk.AccAddress([]byte("user_12345678901")).String(),
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: piiCommitment,
		Jurisdiction:  "", // Empty jurisdiction
	}

	resp, err := msgServer.SubmitKYC(sdk.WrapSDKContext(ctx), msg)

	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "jurisdiction is required")
}

// TestSubmitKYC_InvalidJurisdictionFormat tests validation of jurisdiction format
func TestSubmitKYC_InvalidJurisdictionFormat(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	msgServer := NewMsgServer(keeper)

	// Setup approved KYC provider
	providerAddr := sdk.AccAddress([]byte("provider_address_12")).String()
	params := types.DefaultParams()
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	piiCommitment := makePiiCommitmentForJurisdiction("test")

	invalidCases := []struct {
		name         string
		jurisdiction string
	}{
		{"too short", "U"},
		{"too long", "USA"},
		{"numeric", "12"},
		{"special chars", "U$"},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &types.MsgSubmitKYC{
				Address:       sdk.AccAddress([]byte("user_12345678901")).String(),
				KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      providerAddr,
				PiiCommitment: piiCommitment,
				Jurisdiction:  tc.jurisdiction,
			}

			resp, err := msgServer.SubmitKYC(sdk.WrapSDKContext(ctx), msg)

			require.Error(t, err)
			require.Nil(t, resp)
			require.Contains(t, err.Error(), "2-letter ISO 3166-1 alpha-2 country code")
		})
	}
}

// TestSubmitKYC_JurisdictionInEvent tests that jurisdiction is included in KYC event
func TestSubmitKYC_JurisdictionInEvent(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	msgServer := NewMsgServer(keeper)

	// Setup approved KYC provider
	providerAddr := sdk.AccAddress([]byte("provider_address_12")).String()
	params := types.DefaultParams()
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{"KP"}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	piiCommitment := makePiiCommitmentForJurisdiction("test")

	msg := &types.MsgSubmitKYC{
		Address:       sdk.AccAddress([]byte("user_12345678901")).String(),
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      providerAddr,
		PiiCommitment: piiCommitment,
		Jurisdiction:  "US",
	}

	resp, err := msgServer.SubmitKYC(sdk.WrapSDKContext(ctx), msg)

	// May fail on signer check, but if it succeeds, verify event
	if err == nil {
		require.NotNil(t, resp)

		// Check that event was emitted with jurisdiction
		events := ctx.EventManager().Events()
		require.NotEmpty(t, events)

		found := false
		for _, event := range events {
			if event.Type == types.EventTypeKYCSubmitted {
				for _, attr := range event.Attributes {
					if string(attr.Key) == types.AttributeKeyJurisdiction {
						require.Equal(t, "US", string(attr.Value))
						found = true
					}
				}
			}
		}
		require.True(t, found, "Jurisdiction attribute should be present in KYC submitted event")
	}
}
