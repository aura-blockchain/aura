package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	gs := DefaultGenesisState()

	// Verify we have 16 default VC policies
	require.Len(t, gs.VCPolicies, 16, "should have 16 default VC policies")

	// Verify genesis state is valid
	err := gs.Validate()
	require.NoError(t, err, "default genesis state should be valid")

	// Verify all policies have required fields
	policyTypesSeen := make(map[string]bool)
	for i, policy := range gs.VCPolicies {
		// Each policy should have a unique type name
		require.NotEmpty(t, policy.VcTypeName, "policy %d should have a type name", i)
		require.False(t, policyTypesSeen[policy.VcTypeName], "policy type %s should be unique", policy.VcTypeName)
		policyTypesSeen[policy.VcTypeName] = true

		// Each policy should have required metadata
		require.NotNil(t, policy.CreatedAt, "policy %d should have created_at timestamp", i)
		require.Equal(t, VCPolicyStatusActive, policy.Status, "policy %d should be active", i)
		require.Equal(t, "1.0.0", policy.Version, "policy %d should have version 1.0.0", i)
		require.Equal(t, "genesis", policy.Creator, "policy %d should be created by genesis", i)
	}
}

func TestDefaultVCPolicies(t *testing.T) {
	policies := DefaultVCPolicies()

	// Verify we get exactly 16 policies
	require.Len(t, policies, 16, "should return 16 default policies")

	// Verify policy names match expected types
	expectedNames := []string{
		"VerifiedHuman",
		"AgeOver18",
		"AgeOver21",
		"ResidentOf",
		"BiometricAuth",
		"KYCVerification",
		"NotaryPublic",
		"ProfessionalLicense",
		"BiometricFocus",
		"SocialFocus",
		"GeolocationFocus",
		"HighAssuranceFocus",
		"PossessionFocus",
		"KnowledgeFocus",
		"PersistenceFocus",
		"SpecializedFocus",
	}

	for i, expectedName := range expectedNames {
		require.Equal(t, expectedName, policies[i].VcTypeName, "policy %d should have name %s", i, expectedName)
	}

	// Verify core credential policies (indices 0-7)
	coreCredentials := []struct {
		index           int
		name            string
		csThreshold     uint64
		singleton       bool
		requiresRenewal bool
		expiryDays      uint64
	}{
		{0, "VerifiedHuman", 50, true, true, 365},
		{1, "AgeOver18", 60, true, false, 1825},
		{2, "AgeOver21", 60, true, false, 1825},
		{3, "ResidentOf", 70, false, true, 365},
		{4, "BiometricAuth", 80, true, false, 730},
		{5, "KYCVerification", 90, true, true, 365},
		{6, "NotaryPublic", 95, true, true, 365},
		{7, "ProfessionalLicense", 90, false, true, 365},
	}

	for _, tc := range coreCredentials {
		t.Run(tc.name, func(t *testing.T) {
			policy := policies[tc.index]
			require.Equal(t, tc.name, policy.VcTypeName)
			require.Equal(t, tc.csThreshold, policy.CsThreshold)
			require.Equal(t, tc.singleton, policy.Singleton)
			require.Equal(t, tc.requiresRenewal, policy.RequiresAnnualRenewal)
			require.Equal(t, tc.expiryDays, policy.ExpiryDurationDays)
		})
	}

	// Verify arena focus policies (indices 8-15)
	arenaFocusPolicies := []struct {
		index      int
		name       string
		arena      string
		arenaScore uint64
	}{
		{8, "BiometricFocus", "biometric", 200},
		{9, "SocialFocus", "social", 150},
		{10, "GeolocationFocus", "geolocation", 150},
		{11, "HighAssuranceFocus", "high_assurance", 300},
		{12, "PossessionFocus", "possession", 100},
		{13, "KnowledgeFocus", "knowledge", 150},
		{14, "PersistenceFocus", "persistence", 100},
		{15, "SpecializedFocus", "specialized", 250},
	}

	for _, tc := range arenaFocusPolicies {
		t.Run(tc.name, func(t *testing.T) {
			policy := policies[tc.index]
			require.Equal(t, tc.name, policy.VcTypeName)
			require.Equal(t, tc.arena, policy.RequiredArena)
			require.Equal(t, tc.arenaScore, policy.RequiredArenaScore)
		})
	}
}

func TestGenesisStateValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid default genesis",
			setup: func() GenesisState {
				return DefaultGenesisState()
			},
			wantErr: false,
		},
		{
			name: "duplicate policy type",
			setup: func() GenesisState {
				gs := DefaultGenesisState()
				// Add a duplicate policy
				gs.VCPolicies = append(gs.VCPolicies, gs.VCPolicies[0])
				return gs
			},
			wantErr: true,
			errMsg:  "duplicate vc policy",
		},
		{
			name: "policy with empty type name",
			setup: func() GenesisState {
				gs := DefaultGenesisState()
				gs.VCPolicies[0].VcTypeName = ""
				return gs
			},
			wantErr: true,
			errMsg:  "vc_type_name cannot be empty",
		},
		{
			name: "policy with nil created_at",
			setup: func() GenesisState {
				gs := DefaultGenesisState()
				gs.VCPolicies[0].CreatedAt = nil
				return gs
			},
			wantErr: true,
			errMsg:  "created_at cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := tt.setup()
			err := gs.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
