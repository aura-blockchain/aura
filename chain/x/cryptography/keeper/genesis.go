package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// InitGenesis initializes the cryptography module state from genesis
func (k *Keeper) InitGenesis(ctx context.Context, data *cryptoproto.GenesisState) error {
	if data == nil {
		return nil
	}

	// Unwrap SDK context for methods that need it
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Set parameters
	if data.Params != nil {
		if err := k.SetParams(ctx, data.Params); err != nil {
			k.Logger(ctx).Error("failed to set params", "error", err)
			return err
		}
	}

	// Initialize key rotation schedules
	for _, schedule := range data.KeyRotationSchedules {
		if err := k.SetKeyRotationSchedule(ctx, schedule); err != nil {
			k.Logger(ctx).Error("failed to initialize key rotation schedule", "key_id", schedule.KeyId, "error", err)
		}
	}

	// Initialize threshold signature schemes
	for _, scheme := range data.ThresholdSchemes {
		if err := k.SetThresholdScheme(ctx, scheme); err != nil {
			k.Logger(ctx).Error("failed to initialize threshold scheme", "scheme_id", scheme.SchemeId, "error", err)
		}
	}

	// Initialize ZK proof configs
	for _, config := range data.ZkProofConfigs {
		if err := k.SetZKProofConfig(sdkCtx, config); err != nil {
			k.Logger(ctx).Error("failed to initialize ZK proof config", "proof_id", config.ProofId, "error", err)
		}
	}

	// Initialize secure enclaves
	for _, enclave := range data.SecureEnclaves {
		if err := k.SetSecureEnclaveConfig(sdkCtx, enclave); err != nil {
			k.Logger(ctx).Error("failed to initialize secure enclave", "enclave_id", enclave.EnclaveId, "error", err)
		}
	}

	// Initialize quantum-resistant keys
	for _, key := range data.QuantumResistantKeys {
		if err := k.SetQuantumResistantKey(sdkCtx, key); err != nil {
			k.Logger(ctx).Error("failed to initialize quantum-resistant key", "key_id", key.KeyId, "error", err)
		}
	}

	// Initialize random sources
	for _, source := range data.RandomSources {
		if err := k.SetRandomSource(ctx, source); err != nil {
			k.Logger(ctx).Error("failed to initialize random source", "source_id", source.SourceId, "error", err)
		}
	}

	// Initialize key stretching configs
	for _, config := range data.KeyStretchingConfigs {
		if err := k.SetKeyStretchingConfig(ctx, config); err != nil {
			k.Logger(ctx).Error("failed to initialize key stretching config", "config_id", config.ConfigId, "error", err)
		}
	}

	// Initialize certificate pins
	for _, pin := range data.CertificatePins {
		if err := k.SetCertificatePin(ctx, pin); err != nil {
			k.Logger(ctx).Error("failed to initialize certificate pin", "domain", pin.Hostname, "error", err)
		}
	}

	// Initialize ZK proof verifications
	for _, verification := range data.ZkProofVerifications {
		if err := k.SetZKProofVerification(ctx, verification); err != nil {
			k.Logger(ctx).Error("failed to initialize ZK proof verification", "verification_id", verification.VerificationId, "error", err)
		}
	}

	k.Logger(ctx).Info("cryptography module initialized from genesis")
	return nil
}

// ExportGenesis exports the cryptography module state to genesis
func (k *Keeper) ExportGenesis(ctx context.Context) *cryptoproto.GenesisState {
	// Unwrap SDK context for methods that need it
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get parameters
	params, err := k.GetParams(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to get params during export", "error", err)
		params = &cryptoproto.Params{}
	}

	// Export key rotation schedules
	keyRotationSchedules := k.GetAllKeyRotationSchedules(ctx)

	// Export threshold signature schemes
	thresholdSchemes := k.GetAllThresholdSchemes(ctx)

	// Export ZK proof configs
	zkProofConfigs := k.GetAllZKProofConfigs(sdkCtx)

	// Export secure enclaves
	secureEnclaves := make([]*cryptoproto.SecureEnclaveConfig, 0)
	_ = k.IterateSecureEnclaves(ctx, func(enclave *cryptoproto.SecureEnclaveConfig) bool {
		secureEnclaves = append(secureEnclaves, enclave)
		return false
	})

	// Export quantum-resistant keys
	quantumResistantKeys := make([]*cryptoproto.QuantumResistantKey, 0)
	_ = k.IterateQuantumKeys(ctx, func(key *cryptoproto.QuantumResistantKey) bool {
		quantumResistantKeys = append(quantumResistantKeys, key)
		return false
	})

	// Export random sources
	randomSources := make([]*cryptoproto.CryptoRandomSource, 0)
	_ = k.IterateRandomSources(ctx, func(source *cryptoproto.CryptoRandomSource) bool {
		randomSources = append(randomSources, source)
		return false
	})

	// Export key stretching configs
	keyStretchingConfigs := make([]*cryptoproto.KeyStretchingConfig, 0)
	_ = k.IterateKeyStretchingConfigs(ctx, func(config *cryptoproto.KeyStretchingConfig) bool {
		keyStretchingConfigs = append(keyStretchingConfigs, config)
		return false
	})

	// Export certificate pins
	certificatePins := make([]*cryptoproto.CertificatePin, 0)
	_ = k.IterateCertificatePins(ctx, func(pin *cryptoproto.CertificatePin) bool {
		certificatePins = append(certificatePins, pin)
		return false
	})

	// Export ZK proof verifications
	zkProofVerifications := make([]*cryptoproto.ZKProofVerification, 0)
	_ = k.IterateZKProofVerifications(ctx, func(verification *cryptoproto.ZKProofVerification) bool {
		zkProofVerifications = append(zkProofVerifications, verification)
		return false
	})

	return &cryptoproto.GenesisState{
		Params:               params,
		KeyRotationSchedules: keyRotationSchedules,
		ThresholdSchemes:     thresholdSchemes,
		ZkProofConfigs:       zkProofConfigs,
		SecureEnclaves:       secureEnclaves,
		QuantumResistantKeys: quantumResistantKeys,
		RandomSources:        randomSources,
		KeyStretchingConfigs: keyStretchingConfigs,
		CertificatePins:      certificatePins,
		ZkProofVerifications: zkProofVerifications,
	}
}
