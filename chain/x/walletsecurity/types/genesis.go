package types

import (
	"fmt"
)

// GenesisState represents the state that must be persisted when the chain starts from genesis
type GenesisState struct {
	Params                []byte          `json:"params"`
	HardwareWallets       [][]byte        `json:"hardware_wallets"`
	MultiSigWallets       [][]byte        `json:"multisig_wallets"`
	PendingMultiSigTxs    [][]byte        `json:"pending_multisig_txs"`
	SocialRecoveryConfigs [][]byte        `json:"social_recovery_configs"`
	RecoveryRequests      [][]byte        `json:"recovery_requests"`
	DeviceFingerprints    [][]byte        `json:"device_fingerprints"`
	Sessions              [][]byte        `json:"sessions"`
	AnomalyDetections     [][]byte        `json:"anomaly_detections"`
	WalletAnalytics       [][]byte        `json:"wallet_analytics"`
	InsurancePolicies     [][]byte        `json:"insurance_policies"`
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                nil,
		HardwareWallets:       [][]byte{},
		MultiSigWallets:       [][]byte{},
		PendingMultiSigTxs:    [][]byte{},
		SocialRecoveryConfigs: [][]byte{},
		RecoveryRequests:      [][]byte{},
		DeviceFingerprints:    [][]byte{},
		Sessions:              [][]byte{},
		AnomalyDetections:     [][]byte{},
		WalletAnalytics:       [][]byte{},
		InsurancePolicies:     [][]byte{},
	}
}

// DefaultGenesis returns a default genesis state (alias for DefaultGenesisState)
func DefaultGenesis() *GenesisState {
	return DefaultGenesisState()
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Basic validation - arrays should not be nil after initialization
	return nil
}

// ValidateGenesis validates the genesis state (module-level function)
func ValidateGenesis(gen *GenesisState) error {
	if gen == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}
	return gen.Validate()
}
