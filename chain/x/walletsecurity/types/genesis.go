package types

import (
	"fmt"

	pb "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// GenesisState re-exports the protobuf genesis type for canonical codec handling.
type GenesisState = pb.GenesisState

// DefaultGenesisState returns a deterministic, protobuf-backed genesis state.
func DefaultGenesisState() *GenesisState {
	return &pb.GenesisState{
		Params: &pb.WalletSecurityParams{
			HardwareWalletEnabled:        true,
			SupportedHardwareTypes:       []int32{},
			MaxSigners:                   5,
			MinThreshold:                 2,
			MaxThreshold:                 5,
			SocialRecoveryEnabled:        true,
			MaxGuardians:                 5,
			MinRecoveryThreshold:         2,
			DefaultRecoveryDelaySeconds:  86400,
			DefaultSessionTimeoutSeconds: 3600,
			MaxSessionDurationSeconds:    86400,
			SpendingLimitsEnabled:        true,
			DefaultDailyLimit:            "0",
			BiometricEnabled:             false,
			MaxBiometricAttempts:         5,
			LockoutDurationSeconds:       300,
			DustFilterEnabled:            true,
			MinDustAmount:                "1",
			PhishingProtectionEnabled:    true,
			RequireDomainVerification:    false,
		},
		HardwareWallets:     []*pb.HardwareWalletConfig{},
		MultisigWallets:     []*pb.MultiSigWallet{},
		PendingTransactions: []*pb.PendingMultiSigTransaction{},
		RecoveryConfigs:     []*pb.SocialRecoveryConfig{},
		RecoveryRequests:    []*pb.RecoveryRequest{},
		DomainVerifications: []*pb.DomainVerification{},
		PhishingConfigs:     []*pb.PhishingProtectionConfig{},
		SpendingLimits:      []*pb.SpendingLimit{},
		SessionConfigs:      []*pb.SessionConfig{},
		BiometricConfigs:    []*pb.BiometricAuth{},
		EnclaveConfigs:      []*pb.SecureEnclaveConfig{},
		EncryptedBackups:    []*pb.EncryptedBackup{},
		DustFilters:         []*pb.DustAttackFilter{},
		DustTransactions:    []*pb.DustTransaction{},
		SecurityMetrics:     []*pb.WalletSecurityMetrics{},
	}
}

// DefaultGenesis returns a default genesis state (alias for DefaultGenesisState)
func DefaultGenesis() *GenesisState {
	return DefaultGenesisState()
}

// ValidateGenesis performs sanity checks on the protobuf genesis.
func ValidateGenesis(gen *GenesisState) error {
	if gen == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}
	if gen.Params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	if gen.Params.MinThreshold < 1 || gen.Params.MaxThreshold < gen.Params.MinThreshold {
		return fmt.Errorf("invalid multisig thresholds: min %d max %d", gen.Params.MinThreshold, gen.Params.MaxThreshold)
	}
	if gen.Params.MaxSigners < gen.Params.MinThreshold {
		return fmt.Errorf("max_signers must be >= min_threshold")
	}
	return nil
}
