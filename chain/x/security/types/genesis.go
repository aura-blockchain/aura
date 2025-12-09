package types

import (
	"fmt"

	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// DefaultGenesis returns the default genesis state for the security module
func DefaultGenesis() *securitypb.GenesisState {
	return &securitypb.GenesisState{
		Params:            securitypb.Params{},
		NetworkSecurity:   securitypb.NetworkSecurityState{},
		ValidatorSecurity: securitypb.ValidatorSecurityState{},
		WalletSecurity:    securitypb.WalletSecurityState{},
		IncidentResponse:  securitypb.IncidentResponseState{},
		Cryptography:      securitypb.CryptographyState{},
		Privacy:           securitypb.PrivacyState{},
	}
}

// ValidateGenesisState validates the genesis state
func ValidateGenesisState(gs *securitypb.GenesisState) error {
	if gs == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	// Validate params
	if err := ValidateParams(&gs.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate network security state
	if err := validateNetworkSecurityState(&gs.NetworkSecurity); err != nil {
		return fmt.Errorf("invalid network security state: %w", err)
	}

	// Validate validator security state
	if err := validateValidatorSecurityState(&gs.ValidatorSecurity); err != nil {
		return fmt.Errorf("invalid validator security state: %w", err)
	}

	// Validate wallet security state
	if err := validateWalletSecurityState(&gs.WalletSecurity); err != nil {
		return fmt.Errorf("invalid wallet security state: %w", err)
	}

	// Validate incident response state
	if err := validateIncidentResponseState(&gs.IncidentResponse); err != nil {
		return fmt.Errorf("invalid incident response state: %w", err)
	}

	// Validate cryptography state
	if err := validateCryptographyState(&gs.Cryptography); err != nil {
		return fmt.Errorf("invalid cryptography state: %w", err)
	}

	// Validate privacy state
	if err := validatePrivacyState(&gs.Privacy); err != nil {
		return fmt.Errorf("invalid privacy state: %w", err)
	}

	return nil
}

// ValidateParams validates the security module parameters
func ValidateParams(params *securitypb.Params) error {
	// Add param validation logic as needed
	return nil
}

func validateNetworkSecurityState(ns *securitypb.NetworkSecurityState) error {
	// Validate rate limits
	for i, rl := range ns.RateLimits {
		if rl.PeerId == "" {
			return fmt.Errorf("rate limit %d: peer_id cannot be empty", i)
		}
	}

	// Validate trusted peers
	for i, tp := range ns.TrustedPeers {
		if tp.PeerId == "" {
			return fmt.Errorf("trusted peer %d: peer_id cannot be empty", i)
		}
	}

	// Validate reputations
	for i, rep := range ns.Reputations {
		if rep.PeerId == "" {
			return fmt.Errorf("reputation %d: peer_id cannot be empty", i)
		}
	}

	// Validate fork alerts
	for i, alert := range ns.ForkAlerts {
		if alert.AlertId == "" {
			return fmt.Errorf("fork alert %d: alert_id cannot be empty", i)
		}
	}

	// Validate partition alerts
	for i, alert := range ns.PartitionAlerts {
		if alert.AlertId == "" {
			return fmt.Errorf("partition alert %d: alert_id cannot be empty", i)
		}
	}

	return nil
}

func validateValidatorSecurityState(vs *securitypb.ValidatorSecurityState) error {
	// Validate validator infos
	for i, vi := range vs.Validators {
		if vi.ValidatorAddress == "" {
			return fmt.Errorf("validator %d: validator_address cannot be empty", i)
		}
	}

	// Validate double sign evidences
	for i, ev := range vs.DoubleSignEvidences {
		if ev.ValidatorAddress == "" {
			return fmt.Errorf("double sign evidence %d: validator_address cannot be empty", i)
		}
	}

	// Validate downtime infractions
	for i, inf := range vs.DowntimeInfractions {
		if inf.ValidatorAddress == "" {
			return fmt.Errorf("downtime infraction %d: validator_address cannot be empty", i)
		}
	}

	// Validate alerts
	for i, alert := range vs.Alerts {
		if alert.Id == "" {
			return fmt.Errorf("validator alert %d: id cannot be empty", i)
		}
	}

	return nil
}

func validateWalletSecurityState(ws *securitypb.WalletSecurityState) error {
	// Validate hardware wallets
	for i, hw := range ws.HardwareWallets {
		if hw.WalletId == "" {
			return fmt.Errorf("hardware wallet %d: wallet_id cannot be empty", i)
		}
	}

	// Validate multisig wallets
	for i, ms := range ws.MultisigWallets {
		if ms.WalletId == "" {
			return fmt.Errorf("multisig wallet %d: wallet_id cannot be empty", i)
		}
		if ms.Threshold == 0 {
			return fmt.Errorf("multisig wallet %d: threshold must be greater than 0", i)
		}
		if len(ms.Signers) < int(ms.Threshold) {
			return fmt.Errorf("multisig wallet %d: not enough signers for threshold", i)
		}
	}

	// Validate pending multisig transactions
	for i, tx := range ws.PendingMultisigTxs {
		if tx.TxId == "" {
			return fmt.Errorf("pending multisig tx %d: tx_id cannot be empty", i)
		}
	}

	// Validate social recovery configs
	for i, src := range ws.SocialRecoveryConfigs {
		if src.WalletId == "" {
			return fmt.Errorf("social recovery config %d: wallet_id cannot be empty", i)
		}
	}

	// Validate recovery requests
	for i, req := range ws.RecoveryRequests {
		if req.RequestId == "" {
			return fmt.Errorf("recovery request %d: request_id cannot be empty", i)
		}
	}

	// Validate spending limits
	for i, sl := range ws.SpendingLimits {
		if sl.WalletId == "" {
			return fmt.Errorf("spending limit %d: wallet_id cannot be empty", i)
		}
	}

	return nil
}

func validateIncidentResponseState(ir *securitypb.IncidentResponseState) error {
	// Validate incidents
	for i, inc := range ir.Incidents {
		if inc.IncidentId == "" {
			return fmt.Errorf("incident %d: incident_id cannot be empty", i)
		}
	}

	// Validate audit logs
	for i, entry := range ir.AuditLogs {
		if entry.LogId == "" {
			return fmt.Errorf("audit log entry %d: log_id cannot be empty", i)
		}
	}

	return nil
}

func validateCryptographyState(c *securitypb.CryptographyState) error {
	// Validate key rotation schedules
	for i, krs := range c.KeyRotationSchedules {
		if krs.Id == "" {
			return fmt.Errorf("key rotation schedule %d: id cannot be empty", i)
		}
		if krs.KeyId == "" {
			return fmt.Errorf("key rotation schedule %d: key_id cannot be empty", i)
		}
	}

	// Validate threshold schemes
	for i, ts := range c.ThresholdSchemes {
		if ts.SchemeId == "" {
			return fmt.Errorf("threshold scheme %d: scheme_id cannot be empty", i)
		}
	}

	// Validate ZK proof configs
	for i, zk := range c.ZkProofConfigs {
		if zk.ProofId == "" {
			return fmt.Errorf("zk proof config %d: proof_id cannot be empty", i)
		}
	}

	// Validate quantum resistant keys
	for i, qrk := range c.QuantumResistantKeys {
		if qrk.KeyId == "" {
			return fmt.Errorf("quantum resistant key %d: key_id cannot be empty", i)
		}
	}

	return nil
}

func validatePrivacyState(p *securitypb.PrivacyState) error {
	// Validate mixing pools
	for i, mp := range p.MixingPools {
		if mp.PoolId == "" {
			return fmt.Errorf("mixing pool %d: pool_id cannot be empty", i)
		}
	}

	// Validate stealth addresses
	for i, sa := range p.StealthAddresses {
		if len(sa.OneTimeAddress) == 0 {
			return fmt.Errorf("stealth address %d: one_time_address cannot be empty", i)
		}
	}

	// Validate ring signatures
	for i, rs := range p.RingSignatures {
		if len(rs.KeyImage) == 0 {
			return fmt.Errorf("ring signature %d: key_image cannot be empty", i)
		}
	}

	return nil
}
