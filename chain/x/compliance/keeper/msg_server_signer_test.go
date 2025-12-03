package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)


func makePiiCommitment(seed string) []byte {
	pii := make([]byte, 32)
	copy(pii, []byte(seed))
	return pii
}

// TestSubmitKYC_SignerVerification tests that KYC submission requires proper signer verification
func TestSubmitKYC_SignerVerification(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Create valid bech32 addresses
	providerAddr := sdk.AccAddress([]byte("provider_address_12")).String()
	unauthorizedAddr := sdk.AccAddress([]byte("unauthorized_addr_1")).String()

	// Set up approved provider
	params := keeper.GetParams(ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	t.Run("unauthorized provider rejected", func(t *testing.T) {
		req := &types.MsgSubmitKYC{
			Address:       sdk.AccAddress([]byte("user_address_12345")).String(),
			KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:      unauthorizedAddr,
			PiiCommitment: makePiiCommitment("test"),
		}
		_, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "provider not authorized")
	})

	t.Run("empty provider rejected", func(t *testing.T) {
		req := &types.MsgSubmitKYC{
			Address:       "aura1user000000000000000000000000000000000",
			KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:      "",
			PiiCommitment: makePiiCommitment("test"),
		}
		_, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "provider is required")
	})

	t.Run("invalid provider address rejected", func(t *testing.T) {
		// Invalid bech32 address should fail at GetSigners
		req := &types.MsgSubmitKYC{
			Address:       "aura1user000000000000000000000000000000000",
			KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:      "invalid-bech32-address",
			PiiCommitment: makePiiCommitment("test"),
		}
		_, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		// Should fail at signer verification with invalid address
		require.Contains(t, err.Error(), "no signers")
	})
}

// TestReportSuspiciousActivity_SignerVerification tests SAR filing authorization
func TestReportSuspiciousActivity_SignerVerification(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	t.Run("empty reporter rejected", func(t *testing.T) {
		req := &types.MsgReportSuspiciousActivity{
			Reporter:        "",
			Address:         "aura1user000000000000000000000000000000000",
			TransactionHash: "hash123",
			ActivityType:    "structuring",
			Description:     "suspicious pattern",
		}
		_, err := server.ReportSuspiciousActivity(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "reporter is required")
	})

	t.Run("no signers rejected", func(t *testing.T) {
		req := &types.MsgReportSuspiciousActivity{
			Reporter:        "aura1reporter000000000000000000000000000000",
			Address:         "aura1user000000000000000000000000000000000",
			TransactionHash: "hash123",
			ActivityType:    "structuring",
			Description:     "suspicious pattern",
		}
		_, err := server.ReportSuspiciousActivity(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		// Should fail at signer verification or permission check
		require.True(t, err.Error() == "rpc error: code = Unauthenticated desc = no signers" ||
			err.Error() == "rpc error: code = PermissionDenied desc = reporter must be transaction signer")
	})
}

// TestScreenSanctions_SignerVerification tests sanctions screening authorization
func TestScreenSanctions_SignerVerification(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	t.Run("empty address rejected", func(t *testing.T) {
		req := &types.MsgScreenSanctions{
			Address: "",
		}
		_, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address is required")
	})

	t.Run("no signers rejected", func(t *testing.T) {
		req := &types.MsgScreenSanctions{
			Address: "aura1user000000000000000000000000000000000",
		}
		_, err := server.ScreenSanctions(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		// Should fail at signer verification or permission check
		require.True(t, err.Error() == "rpc error: code = Unauthenticated desc = no signers" ||
			err.Error() == "rpc error: code = PermissionDenied desc = address must match transaction signer")
	})
}

// TestRecordGDPRConsent_SignerVerification tests GDPR consent authorization
func TestRecordGDPRConsent_SignerVerification(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	t.Run("empty address rejected", func(t *testing.T) {
		req := &types.MsgRecordGDPRConsent{
			Address:        "",
			ConsentType:    "data_processing",
			Consented:      true,
			ConsentVersion: "v1",
		}
		_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address and consent type required")
	})

	t.Run("no signers rejected", func(t *testing.T) {
		req := &types.MsgRecordGDPRConsent{
			Address:        "aura1user000000000000000000000000000000000",
			ConsentType:    "data_processing",
			Consented:      true,
			ConsentVersion: "v1",
		}
		_, err := server.RecordGDPRConsent(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		// Should fail at signer verification or permission check
		require.True(t, err.Error() == "rpc error: code = Unauthenticated desc = no signers" ||
			err.Error() == "rpc error: code = PermissionDenied desc = address must match transaction signer")
	})
}

// TestRequestGDPRData_SignerVerification tests GDPR data request authorization
func TestRequestGDPRData_SignerVerification(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	t.Run("empty address rejected", func(t *testing.T) {
		req := &types.MsgRequestGDPRData{
			Address:     "",
			RequestType: "access",
		}
		_, err := server.RequestGDPRData(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address and request type required")
	})

	t.Run("no signers rejected", func(t *testing.T) {
		req := &types.MsgRequestGDPRData{
			Address:     "aura1user000000000000000000000000000000000",
			RequestType: "access",
		}
		_, err := server.RequestGDPRData(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		// Should fail at signer verification or permission check
		require.True(t, err.Error() == "rpc error: code = Unauthenticated desc = no signers" ||
			err.Error() == "rpc error: code = PermissionDenied desc = address must match transaction signer")
	})
}

// TestGenerateTaxReport_SignerVerification tests tax report generation authorization
func TestGenerateTaxReport_SignerVerification(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	t.Run("empty address rejected", func(t *testing.T) {
		req := &types.MsgGenerateTaxReport{
			Address:      "",
			TaxYear:      "2024",
			Jurisdiction: "US",
			ReportType:   "1099",
		}
		_, err := server.GenerateTaxReport(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address, tax year, and jurisdiction required")
	})

	t.Run("no signers rejected", func(t *testing.T) {
		req := &types.MsgGenerateTaxReport{
			Address:      "aura1user000000000000000000000000000000000",
			TaxYear:      "2024",
			Jurisdiction: "US",
			ReportType:   "1099",
		}
		_, err := server.GenerateTaxReport(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		// Should fail at signer verification or permission check
		require.True(t, err.Error() == "rpc error: code = Unauthenticated desc = no signers" ||
			err.Error() == "rpc error: code = PermissionDenied desc = address must match transaction signer")
	})
}

// TestGetSigners_Implementation tests that GetSigners is implemented for all message types
func TestGetSigners_Implementation(t *testing.T) {
	// Use real valid bech32 addresses with correct length
	validAddr := sdk.AccAddress([]byte("test_address_123456"))
	validBech32 := validAddr.String()

	t.Run("MsgSubmitKYC GetSigners", func(t *testing.T) {
		msg := &types.MsgSubmitKYC{
			Provider: validBech32,
		}
		signers := msg.GetSigners()
		require.NotNil(t, signers)
		// Should return the provider address
		require.Len(t, signers, 1)
		require.Equal(t, validAddr, signers[0])
	})

	t.Run("MsgReportSuspiciousActivity GetSigners", func(t *testing.T) {
		msg := &types.MsgReportSuspiciousActivity{
			Reporter: validBech32,
		}
		signers := msg.GetSigners()
		require.NotNil(t, signers)
		require.Len(t, signers, 1)
		require.Equal(t, validAddr, signers[0])
	})

	t.Run("MsgScreenSanctions GetSigners", func(t *testing.T) {
		msg := &types.MsgScreenSanctions{
			Address: validBech32,
		}
		signers := msg.GetSigners()
		require.NotNil(t, signers)
		require.Len(t, signers, 1)
		require.Equal(t, validAddr, signers[0])
	})

	t.Run("MsgRecordGDPRConsent GetSigners", func(t *testing.T) {
		msg := &types.MsgRecordGDPRConsent{
			Address: validBech32,
		}
		signers := msg.GetSigners()
		require.NotNil(t, signers)
		require.Len(t, signers, 1)
		require.Equal(t, validAddr, signers[0])
	})

	t.Run("MsgRequestGDPRData GetSigners", func(t *testing.T) {
		msg := &types.MsgRequestGDPRData{
			Address: validBech32,
		}
		signers := msg.GetSigners()
		require.NotNil(t, signers)
		require.Len(t, signers, 1)
		require.Equal(t, validAddr, signers[0])
	})

	t.Run("MsgGenerateTaxReport GetSigners", func(t *testing.T) {
		msg := &types.MsgGenerateTaxReport{
			Address: validBech32,
		}
		signers := msg.GetSigners()
		require.NotNil(t, signers)
		require.Len(t, signers, 1)
		require.Equal(t, validAddr, signers[0])
	})
}

// TestGetSigners_InvalidAddress tests handling of invalid addresses
func TestGetSigners_InvalidAddress(t *testing.T) {
	t.Run("MsgSubmitKYC invalid provider", func(t *testing.T) {
		msg := &types.MsgSubmitKYC{
			Provider: "invalid-address",
		}
		signers := msg.GetSigners()
		// Should return empty list for invalid address
		require.Len(t, signers, 0)
	})

	t.Run("MsgReportSuspiciousActivity empty reporter", func(t *testing.T) {
		msg := &types.MsgReportSuspiciousActivity{
			Reporter: "",
		}
		signers := msg.GetSigners()
		require.Len(t, signers, 0)
	})
}

// TestKYCProviderAuthorization_Integration tests the complete authorization flow
func TestKYCProviderAuthorization_Integration(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewMsgServer(keeper)

	// Create valid bech32 addresses
	providerAddr := sdk.AccAddress([]byte("provider_address_12")).String()
	unauthorizedAddr := sdk.AccAddress([]byte("unauthorized_addr_1")).String()
	userAddr := sdk.AccAddress([]byte("user_address_12345")).String()

	t.Run("setup approved providers", func(t *testing.T) {
		params := keeper.GetParams(ctx)
		params.ApprovedKycProviders = []string{providerAddr}
		err := keeper.SetParams(ctx, params)
		require.NoError(t, err)

		// Verify params were set
		retrieved := keeper.GetParams(ctx)
		require.Len(t, retrieved.ApprovedKycProviders, 1)
		require.Equal(t, providerAddr, retrieved.ApprovedKycProviders[0])
	})

	t.Run("unauthorized provider cannot submit KYC", func(t *testing.T) {
		req := &types.MsgSubmitKYC{
			Address:       userAddr,
			KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:      unauthorizedAddr,
			PiiCommitment: makePiiCommitment("test"),
		}
		_, err := server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "provider not authorized")
	})

	t.Run("empty providers list blocks all submissions", func(t *testing.T) {
		params := keeper.GetParams(ctx)
		params.ApprovedKycProviders = []string{}
		err := keeper.SetParams(ctx, params)
		require.NoError(t, err)

		req := &types.MsgSubmitKYC{
			Address:       userAddr,
			KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:      providerAddr,
			PiiCommitment: makePiiCommitment("test"),
		}
		_, err = server.SubmitKYC(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "provider not authorized")
	})
}
