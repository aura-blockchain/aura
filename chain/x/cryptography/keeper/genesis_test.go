package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

func TestGenesis(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("InitGenesis - nil genesis", func(t *testing.T) {
		err := k.InitGenesis(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("InitGenesis - empty genesis", func(t *testing.T) {
		genesis := &cryptoproto.GenesisState{}
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})

	t.Run("InitGenesis - full genesis", func(t *testing.T) {
		params := types.DefaultParams()
		params.DefaultRotationIntervalDays = 120

		now := time.Now()

		// Key rotation schedule
		schedule := &cryptoproto.KeyRotationSchedule{
			Id:                      "schedule-1",
			KeyId:                   "key-1",
			NextRotationTime:        now.Add(24 * time.Hour),
			RotationIntervalSeconds: 86400,
			Enabled:                 true,
			CreatedBy:               "creator",
			Policy: &cryptoproto.KeyRotationPolicy{
				MaxAgeDays:              90,
				WarningDaysBeforeExpiry: 7,
				AutoRotate:              true,
				MaxRotationAttempts:     3,
			},
		}

		// Secure enclave
		enclave := &cryptoproto.SecureEnclaveConfig{
			EnclaveId:       "enclave-1",
			EnclaveType:     cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX,
			AttestationData: make([]byte, 432),
			AttestationTime: now,
			Status:          cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_READY,
			EnclaveMetadata: map[string]string{"version": "1.0"},
		}

		// Quantum resistant key
		expiresAt := now.Add(365 * 24 * time.Hour)
		qrKey := &cryptoproto.QuantumResistantKey{
			KeyId:       "qr-key-1",
			Algorithm:   cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM,
			PublicKey:   make([]byte, 1312),
			KeyMetadata: []byte("dilithium2"),
			CreatedAt:   now,
			ExpiresAt:   &expiresAt,
		}

		// Random source
		randomSource := &cryptoproto.CryptoRandomSource{
			SourceId:        "rng-1",
			SourceType:      cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM,
			EntropyPoolHash: make([]byte, 32),
			EntropyBits:     512,
			LastSeeded:      now,
			Status:          cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY,
		}

		// Key stretching config
		keyStretchConfig := &cryptoproto.KeyStretchingConfig{
			ConfigId:    "ksc-1",
			Algorithm:   cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2ID,
			Iterations:  3,
			MemoryCost:  65536,
			Parallelism: 4,
			KeyLength:   32,
			Salt:        make([]byte, 16),
			CreatedAt:   now,
		}

		// Certificate pin
		certPinExpiresAt := now.Add(365 * 24 * time.Hour)
		certPin := &cryptoproto.CertificatePin{
			PinId:             "pin-1",
			Hostname:          "genesis-test.com",
			CertificateHashes: [][]byte{make([]byte, 32)},
			PinType:           cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI,
			CreatedAt:         now,
			ExpiresAt:         &certPinExpiresAt,
			Enabled:           true,
		}

		genesis := &cryptoproto.GenesisState{
			Params:               params,
			KeyRotationSchedules: []*cryptoproto.KeyRotationSchedule{schedule},
			ThresholdSchemes:     []*cryptoproto.ThresholdSignatureScheme{},
			ZkProofConfigs:       []*cryptoproto.ZKProofConfig{},
			SecureEnclaves:       []*cryptoproto.SecureEnclaveConfig{enclave},
			QuantumResistantKeys: []*cryptoproto.QuantumResistantKey{qrKey},
			RandomSources:        []*cryptoproto.CryptoRandomSource{randomSource},
			KeyStretchingConfigs: []*cryptoproto.KeyStretchingConfig{keyStretchConfig},
			CertificatePins:      []*cryptoproto.CertificatePin{certPin},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify data was loaded
		loadedParams, err := k.GetParams(ctx)
		require.NoError(t, err)
		require.Equal(t, int32(120), loadedParams.DefaultRotationIntervalDays)

		loadedSchedule, err := k.GetKeyRotationSchedule(ctx, "schedule-1")
		require.NoError(t, err)
		require.Equal(t, "key-1", loadedSchedule.KeyId)

		loadedEnclave, err := k.GetSecureEnclave(ctx, "enclave-1")
		require.NoError(t, err)
		require.Equal(t, cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, loadedEnclave.EnclaveType)

		loadedQRKey, err := k.GetQuantumResistantKey(ctx, "qr-key-1")
		require.NoError(t, err)
		require.Equal(t, cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM, loadedQRKey.Algorithm)

		loadedRandomSource, err := k.GetRandomSource(ctx, "rng-1")
		require.NoError(t, err)
		require.Equal(t, int64(512), loadedRandomSource.EntropyBits)

		loadedCertPin, err := k.GetCertificatePin(ctx, "genesis-test.com")
		require.NoError(t, err)
		require.Equal(t, "genesis-test.com", loadedCertPin.Hostname)
	})

	t.Run("InitGenesis - invalid params", func(t *testing.T) {
		invalidParams := types.DefaultParams()
		invalidParams.MinThresholdParticipants = 0 // Invalid

		genesis := &cryptoproto.GenesisState{
			Params: invalidParams,
		}

		err := k.InitGenesis(ctx, genesis)
		require.Error(t, err)
	})

	t.Run("ExportGenesis", func(t *testing.T) {
		// Set up some data
		entropy := make([]byte, 64)
		_, err := k.InitializeRandomSource(ctx, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM, entropy)
		require.NoError(t, err)

		_, err = k.CreateKeyRotationSchedule(ctx, "creator", "export-key", 86400, nil)
		require.NoError(t, err)

		_, _, err = k.CreateThresholdScheme(ctx, "creator", 2, 3, []string{"p1", "p2", "p3"}, cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA)
		require.NoError(t, err)

		attestation := make([]byte, 432)
		_, err = k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, attestation, nil)
		require.NoError(t, err)

		_, _, err = k.GenerateQuantumResistantKey(ctx, "creator", cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER, nil)
		require.NoError(t, err)

		hash := make([]byte, 32)
		_, err = k.AddCertificatePin(ctx, "creator", "export-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		// Export genesis
		exported := k.ExportGenesis(ctx)
		require.NotNil(t, exported)
		require.NotNil(t, exported.Params)
		require.NotEmpty(t, exported.KeyRotationSchedules)
		require.NotEmpty(t, exported.ThresholdSchemes)
		require.NotEmpty(t, exported.SecureEnclaves)
		require.NotEmpty(t, exported.QuantumResistantKeys)
		require.NotEmpty(t, exported.RandomSources)
		require.NotEmpty(t, exported.CertificatePins)
	})

	t.Run("InitGenesis and ExportGenesis - round trip", func(t *testing.T) {
		// Create initial genesis state
		params := types.DefaultParams()
		params.DefaultRotationIntervalDays = 180

		now := time.Now()
		schedule := &cryptoproto.KeyRotationSchedule{
			Id:                      "roundtrip-schedule",
			KeyId:                   "roundtrip-key",
			NextRotationTime:        now.Add(24 * time.Hour),
			RotationIntervalSeconds: 86400,
			Enabled:                 true,
			CreatedBy:               "creator",
		}

		genesis := &cryptoproto.GenesisState{
			Params:               params,
			KeyRotationSchedules: []*cryptoproto.KeyRotationSchedule{schedule},
		}

		// Init genesis
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Export genesis
		exported := k.ExportGenesis(ctx)
		require.NotNil(t, exported)
		require.Equal(t, int32(180), exported.Params.DefaultRotationIntervalDays)

		// Verify schedule was exported
		found := false
		for _, sched := range exported.KeyRotationSchedules {
			if sched.Id == "roundtrip-schedule" {
				found = true
				require.Equal(t, "roundtrip-key", sched.KeyId)
				break
			}
		}
		require.True(t, found)
	})

	t.Run("SetKeyRotationSchedule - nil schedule", func(t *testing.T) {
		err := k.SetKeyRotationSchedule(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("SetSecureEnclaveConfig - nil config", func(t *testing.T) {
		err := k.SetSecureEnclaveConfig(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("SetQuantumResistantKey - nil key", func(t *testing.T) {
		err := k.SetQuantumResistantKey(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("SetRandomSource - nil source", func(t *testing.T) {
		err := k.SetRandomSource(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("SetKeyStretchingConfig - nil config", func(t *testing.T) {
		err := k.SetKeyStretchingConfig(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("SetCertificatePin - nil pin", func(t *testing.T) {
		err := k.SetCertificatePin(ctx, nil)
		require.NoError(t, err)
	})
}
