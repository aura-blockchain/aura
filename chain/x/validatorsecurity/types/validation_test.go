package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidateValidatorInfo_Valid(t *testing.T) {
	info := &ValidatorSecurityInfo{
		ValidatorAddress: "auravaloper1test",
		Latitude:         40.7128,
		Longitude:        -74.0060,
	}

	err := ValidateValidatorInfo(info)
	require.NoError(t, err)
}

func TestValidateValidatorInfo_Nil(t *testing.T) {
	err := ValidateValidatorInfo(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validator info cannot be nil")
}

func TestValidateValidatorInfo_EmptyAddress(t *testing.T) {
	info := &ValidatorSecurityInfo{
		ValidatorAddress: "",
		Latitude:         40.7128,
		Longitude:        -74.0060,
	}

	err := ValidateValidatorInfo(info)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validator address cannot be empty")
}

func TestValidateValidatorInfo_InvalidLatitude(t *testing.T) {
	tests := []struct {
		name     string
		latitude float64
	}{
		{"below minimum", -91.0},
		{"above maximum", 91.0},
		{"way below", -200.0},
		{"way above", 200.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ValidatorSecurityInfo{
				ValidatorAddress: "auravaloper1test",
				Latitude:         tt.latitude,
				Longitude:        -74.0060,
			}

			err := ValidateValidatorInfo(info)
			require.Error(t, err)
			require.Contains(t, err.Error(), "latitude must be between -90 and 90")
		})
	}
}

func TestValidateValidatorInfo_ValidLatitude(t *testing.T) {
	tests := []struct {
		name     string
		latitude float64
	}{
		{"minimum", -90.0},
		{"maximum", 90.0},
		{"zero", 0.0},
		{"positive", 45.0},
		{"negative", -45.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ValidatorSecurityInfo{
				ValidatorAddress: "auravaloper1test",
				Latitude:         tt.latitude,
				Longitude:        0.0,
			}

			err := ValidateValidatorInfo(info)
			require.NoError(t, err)
		})
	}
}

func TestValidateValidatorInfo_InvalidLongitude(t *testing.T) {
	tests := []struct {
		name      string
		longitude float64
	}{
		{"below minimum", -181.0},
		{"above maximum", 181.0},
		{"way below", -500.0},
		{"way above", 500.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ValidatorSecurityInfo{
				ValidatorAddress: "auravaloper1test",
				Latitude:         40.7128,
				Longitude:        tt.longitude,
			}

			err := ValidateValidatorInfo(info)
			require.Error(t, err)
			require.Contains(t, err.Error(), "longitude must be between -180 and 180")
		})
	}
}

func TestValidateValidatorInfo_ValidLongitude(t *testing.T) {
	tests := []struct {
		name      string
		longitude float64
	}{
		{"minimum", -180.0},
		{"maximum", 180.0},
		{"zero", 0.0},
		{"positive", 90.0},
		{"negative", -90.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ValidatorSecurityInfo{
				ValidatorAddress: "auravaloper1test",
				Latitude:         0.0,
				Longitude:        tt.longitude,
			}

			err := ValidateValidatorInfo(info)
			require.NoError(t, err)
		})
	}
}

func TestValidateDoubleSignEvidence_Valid(t *testing.T) {
	evidence := &DoubleSignEvidence{
		ValidatorAddress: "auravaloper1test",
		Height:           1000,
		Time:             timestamppb.Now(),
	}

	err := ValidateDoubleSignEvidence(evidence)
	require.NoError(t, err)
}

func TestValidateDoubleSignEvidence_Nil(t *testing.T) {
	err := ValidateDoubleSignEvidence(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "evidence cannot be nil")
}

func TestValidateDoubleSignEvidence_EmptyAddress(t *testing.T) {
	evidence := &DoubleSignEvidence{
		ValidatorAddress: "",
		Height:           1000,
	}

	err := ValidateDoubleSignEvidence(evidence)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validator address cannot be empty")
}

func TestValidateParams(t *testing.T) {
	params := DefaultParams()
	require.NoError(t, ValidateParams(params))

	invalidSlash := DefaultParams()
	invalidSlash.DoubleSignSlashFraction = "1.5"
	require.Error(t, ValidateParams(invalidSlash))

	invalidDowntime := DefaultParams()
	invalidDowntime.DowntimeSlashFraction = "-0.1"
	require.Error(t, ValidateParams(invalidDowntime))

	invalidWindow := DefaultParams()
	invalidWindow.SignedBlocksWindow = 0
	require.Error(t, ValidateParams(invalidWindow))

	invalidMonitoring := DefaultParams()
	invalidMonitoring.MonitoringInterval = durationpb.New(0)
	require.Error(t, ValidateParams(invalidMonitoring))

	invalidFailover := DefaultParams()
	invalidFailover.FailoverTimeout = durationpb.New(0)
	require.Error(t, ValidateParams(invalidFailover))

	paramsNoGeo := DefaultParams()
	paramsNoGeo.EnableGeoDistribution = false
	paramsNoGeo.RequireSentryNodes = false
	paramsNoGeo.EnableAutoFailover = false
	paramsNoGeo.MonitoringInterval = durationpb.New(30 * time.Second)
	require.NoError(t, ValidateParams(paramsNoGeo))
}

func TestValidateDowntimeInfraction_Valid(t *testing.T) {
	infraction := &DowntimeInfraction{
		ValidatorAddress: "auravaloper1test",
		MissedBlocks:     50,
	}

	err := ValidateDowntimeInfraction(infraction)
	require.NoError(t, err)
}

func TestValidateDowntimeInfraction_Nil(t *testing.T) {
	err := ValidateDowntimeInfraction(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "infraction cannot be nil")
}

func TestValidateDowntimeInfraction_EmptyAddress(t *testing.T) {
	infraction := &DowntimeInfraction{
		ValidatorAddress: "",
		MissedBlocks:     50,
	}

	err := ValidateDowntimeInfraction(infraction)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validator address cannot be empty")
}

func TestValidateValidatorAlert_Valid(t *testing.T) {
	alert := &ValidatorAlert{
		Id:               "alert1",
		ValidatorAddress: "auravaloper1test",
		AlertType:        ValidatorAlert_DOWNTIME,
		Severity:         ValidatorAlert_WARNING,
		Message:          "Test alert",
	}

	err := ValidateValidatorAlert(alert)
	require.NoError(t, err)
}

func TestValidateValidatorAlert_Nil(t *testing.T) {
	err := ValidateValidatorAlert(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "alert cannot be nil")
}

func TestValidateValidatorAlert_EmptyAddress(t *testing.T) {
	alert := &ValidatorAlert{
		Id:               "alert1",
		ValidatorAddress: "",
		AlertType:        ValidatorAlert_DOWNTIME,
	}

	err := ValidateValidatorAlert(alert)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validator address cannot be empty")
}

func TestValidateSentryNodeInfo_Valid(t *testing.T) {
	info := &SentryNodeInfo{
		ValidatorAddress: "auravaloper1test",
		Address:          "192.168.1.1:26656",
	}

	err := ValidateSentryNodeInfo(info)
	require.NoError(t, err)
}

func TestValidateSentryNodeInfo_Nil(t *testing.T) {
	err := ValidateSentryNodeInfo(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sentry node info cannot be nil")
}

func TestValidateSentryNodeInfo_EmptyValidatorAddress(t *testing.T) {
	info := &SentryNodeInfo{
		ValidatorAddress: "",
		Address:          "192.168.1.1:26656",
	}

	err := ValidateSentryNodeInfo(info)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validator address cannot be empty")
}

func TestValidateSentryNodeInfo_EmptyNodeAddress(t *testing.T) {
	info := &SentryNodeInfo{
		ValidatorAddress: "auravaloper1test",
		Address:          "",
	}

	err := ValidateSentryNodeInfo(info)
	require.Error(t, err)
	require.Contains(t, err.Error(), "node address cannot be empty")
}
