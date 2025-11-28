package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.NotNil(t, params)
	require.Equal(t, "0.05", params.DoubleSignSlashFraction)
	require.Equal(t, "0.01", params.DowntimeSlashFraction)
	require.Equal(t, int64(1000), params.SignedBlocksWindow)
	require.Equal(t, "0.5", params.MinSignedPerWindow)
	require.Equal(t, "1000000", params.MinimumStakeAmount)

	// Validate default params
	require.NoError(t, ValidateParams(params))
}

func TestValidateParams_Valid(t *testing.T) {
	params := DefaultParams()
	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidateParams_NilParams(t *testing.T) {
	err := ValidateParams(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "params cannot be nil")
}

func TestValidateParams_EmptyDoubleSignSlashFraction(t *testing.T) {
	params := DefaultParams()
	params.DoubleSignSlashFraction = ""
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "double_sign_slash_fraction")
}

func TestValidateParams_EmptyDowntimeSlashFraction(t *testing.T) {
	params := DefaultParams()
	params.DowntimeSlashFraction = ""
	err := ValidateParams(params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "downtime_slash_fraction")
}

func TestValidateParams_CustomValues(t *testing.T) {
	params := &ValidatorSecurityParams{
		DoubleSignSlashFraction: "0.10",
		DowntimeSlashFraction:   "0.02",
		SignedBlocksWindow:      200,
		MinSignedPerWindow:      "0.75",
		MinimumStakeAmount:      "5000000",
		DowntimeJailDuration:    durationpb.New(30 * time.Minute),
		MonitoringInterval:      durationpb.New(2 * time.Minute),
		FailoverTimeout:         durationpb.New(10 * time.Minute),
	}

	err := ValidateParams(params)
	require.NoError(t, err)
}

func TestValidatorAlert_Severity_Constants(t *testing.T) {
	// Test that severity constants are properly defined
	require.NotEqual(t, ValidatorAlert_INFO, ValidatorAlert_WARNING)
	require.NotEqual(t, ValidatorAlert_WARNING, ValidatorAlert_CRITICAL)
	require.NotEqual(t, ValidatorAlert_INFO, ValidatorAlert_CRITICAL)
}

func TestValidatorAlert_AlertType_Constants(t *testing.T) {
	// Test that alert type constants are properly defined
	alertTypes := []ValidatorAlert_AlertType{
		ValidatorAlert_DOWNTIME,
		ValidatorAlert_DOUBLE_SIGN,
		ValidatorAlert_LOW_STAKE,
		ValidatorAlert_SENTRY_NODE_OFFLINE,
		ValidatorAlert_GEOGRAPHIC_VIOLATION,
		ValidatorAlert_KEY_COMPROMISE,
		ValidatorAlert_FAILOVER_TRIGGERED,
	}

	// Ensure all alert types are unique
	for i, at1 := range alertTypes {
		for j, at2 := range alertTypes {
			if i != j {
				require.NotEqual(t, at1, at2)
			}
		}
	}
}

func TestValidatorSecurityInfo_Fields(t *testing.T) {
	info := &ValidatorSecurityInfo{
		ValidatorAddress: "auravaloper1test",
		Latitude:         40.7128,
		Longitude:        -74.0060,
	}

	require.Equal(t, "auravaloper1test", info.ValidatorAddress)
	require.Equal(t, 40.7128, info.Latitude)
	require.Equal(t, -74.0060, info.Longitude)
}

func TestSentryNodeInfo_Fields(t *testing.T) {
	node := &SentryNodeInfo{
		ValidatorAddress: "auravaloper1test",
		Address:          "192.168.1.1:26656",
	}

	require.Equal(t, "auravaloper1test", node.ValidatorAddress)
	require.Equal(t, "192.168.1.1:26656", node.Address)
}

func TestDoubleSignEvidence_Fields(t *testing.T) {
	evidence := &DoubleSignEvidence{
		ValidatorAddress: "auravaloper1test",
		Height:           1000,
		Time:             nil,
	}

	require.Equal(t, "auravaloper1test", evidence.ValidatorAddress)
	require.Equal(t, int64(1000), evidence.Height)
}

func TestDowntimeInfraction_Fields(t *testing.T) {
	infraction := &DowntimeInfraction{
		ValidatorAddress: "auravaloper1test",
		MissedBlocks:     50,
	}

	require.Equal(t, "auravaloper1test", infraction.ValidatorAddress)
	require.Equal(t, int64(50), infraction.MissedBlocks)
}

func TestValidatorAlert_Fields(t *testing.T) {
	alert := &ValidatorAlert{
		Id:               "alert1",
		ValidatorAddress: "auravaloper1test",
		AlertType:        ValidatorAlert_DOWNTIME,
		Severity:         ValidatorAlert_WARNING,
		Message:          "Validator missed blocks",
		Acknowledged:     false,
	}

	require.Equal(t, "alert1", alert.Id)
	require.Equal(t, "auravaloper1test", alert.ValidatorAddress)
	require.Equal(t, ValidatorAlert_DOWNTIME, alert.AlertType)
	require.Equal(t, ValidatorAlert_WARNING, alert.Severity)
	require.Equal(t, "Validator missed blocks", alert.Message)
	require.False(t, alert.Acknowledged)
}

func TestValidatorSecurityParams_AllFields(t *testing.T) {
	params := DefaultParams()

	require.NotEmpty(t, params.DoubleSignSlashFraction)
	require.NotEmpty(t, params.DowntimeSlashFraction)
	require.NotZero(t, params.SignedBlocksWindow)
	require.NotEmpty(t, params.MinSignedPerWindow)
	require.NotEmpty(t, params.MinimumStakeAmount)
}
