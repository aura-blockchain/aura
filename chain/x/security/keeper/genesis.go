// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
	"github.com/aequitas/aura/chain/x/security/types"
)

// InitGenesis initializes the module's state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState *securitypb.GenesisState) error {
	// Validate genesis state
	if err := types.ValidateGenesisState(genState); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Store params (Params is a value type, always present)
	k.SetParams(ctx, genState.Params)

	// Initialize network security state (value type, always present)
	ns := genState.NetworkSecurity

	// Track seen IDs for duplicate detection
	seenPeerIDs := make(map[string]bool)
	seenAlertIDs := make(map[string]bool)

	for i := range ns.RateLimits {
		// CRITICAL SECURITY: Detect duplicate peer IDs in rate limits
		if seenPeerIDs[ns.RateLimits[i].PeerId] {
			return fmt.Errorf("duplicate peer ID in rate limits: %s", ns.RateLimits[i].PeerId)
		}
		seenPeerIDs[ns.RateLimits[i].PeerId] = true
		k.SetRateLimit(ctx, &ns.RateLimits[i])
	}

	seenPeerIDs = make(map[string]bool)
	for i := range ns.Reputations {
		// CRITICAL SECURITY: Detect duplicate peer IDs in reputations
		if seenPeerIDs[ns.Reputations[i].PeerId] {
			return fmt.Errorf("duplicate peer ID in reputations: %s", ns.Reputations[i].PeerId)
		}
		seenPeerIDs[ns.Reputations[i].PeerId] = true
		k.SetPeerReputation(ctx, &ns.Reputations[i])
	}

	seenPeerIDs = make(map[string]bool)
	for i := range ns.TrustedPeers {
		// CRITICAL SECURITY: Detect duplicate peer IDs in trusted peers
		if seenPeerIDs[ns.TrustedPeers[i].PeerId] {
			return fmt.Errorf("duplicate peer ID in trusted peers: %s", ns.TrustedPeers[i].PeerId)
		}
		seenPeerIDs[ns.TrustedPeers[i].PeerId] = true
		k.SetTrustedPeer(ctx, &ns.TrustedPeers[i])
	}

	for i := range ns.ForkAlerts {
		// CRITICAL SECURITY: Detect duplicate alert IDs
		if seenAlertIDs[ns.ForkAlerts[i].AlertId] {
			return fmt.Errorf("duplicate fork alert ID: %s", ns.ForkAlerts[i].AlertId)
		}
		seenAlertIDs[ns.ForkAlerts[i].AlertId] = true
		k.SetForkAlert(ctx, &ns.ForkAlerts[i])
	}

	seenAlertIDs = make(map[string]bool)
	for i := range ns.PartitionAlerts {
		// CRITICAL SECURITY: Detect duplicate alert IDs
		if seenAlertIDs[ns.PartitionAlerts[i].AlertId] {
			return fmt.Errorf("duplicate partition alert ID: %s", ns.PartitionAlerts[i].AlertId)
		}
		seenAlertIDs[ns.PartitionAlerts[i].AlertId] = true
		k.SetPartitionAlert(ctx, &ns.PartitionAlerts[i])
	}

	// Initialize validator security state (value type, always present)
	vs := genState.ValidatorSecurity

	seenValidators := make(map[string]bool)
	seenAlertIDs = make(map[string]bool)

	for i := range vs.Validators {
		// CRITICAL SECURITY: Detect duplicate validator addresses
		if seenValidators[vs.Validators[i].ValidatorAddress] {
			return fmt.Errorf("duplicate validator address in security info: %s", vs.Validators[i].ValidatorAddress)
		}
		seenValidators[vs.Validators[i].ValidatorAddress] = true
		k.SetValidatorSecurityInfo(ctx, &vs.Validators[i])
	}

	for i := range vs.DoubleSignEvidences {
		k.SetDoubleSignEvidence(ctx, &vs.DoubleSignEvidences[i])
	}

	for i := range vs.DowntimeInfractions {
		k.SetDowntimeInfraction(ctx, &vs.DowntimeInfractions[i])
	}

	for i := range vs.Alerts {
		// CRITICAL SECURITY: Detect duplicate alert IDs
		if seenAlertIDs[vs.Alerts[i].Id] {
			return fmt.Errorf("duplicate validator alert ID: %s", vs.Alerts[i].Id)
		}
		seenAlertIDs[vs.Alerts[i].Id] = true
		k.SetValidatorAlert(ctx, &vs.Alerts[i])
	}

	for i := range vs.SentryNodes {
		k.SetSentryNode(ctx, &vs.SentryNodes[i])
	}

	// Initialize wallet security state (value type, always present)
	ws := genState.WalletSecurity

	seenWalletIDs := make(map[string]bool)
	seenTxIDs := make(map[string]bool)
	seenRequestIDs := make(map[string]bool)

	for i := range ws.HardwareWallets {
		// CRITICAL SECURITY: Detect duplicate wallet IDs
		if seenWalletIDs[ws.HardwareWallets[i].WalletId] {
			return fmt.Errorf("duplicate hardware wallet ID: %s", ws.HardwareWallets[i].WalletId)
		}
		seenWalletIDs[ws.HardwareWallets[i].WalletId] = true
		k.SetHardwareWallet(ctx, &ws.HardwareWallets[i])
	}

	seenWalletIDs = make(map[string]bool)
	for i := range ws.MultisigWallets {
		// CRITICAL SECURITY: Detect duplicate wallet IDs
		if seenWalletIDs[ws.MultisigWallets[i].WalletId] {
			return fmt.Errorf("duplicate multisig wallet ID: %s", ws.MultisigWallets[i].WalletId)
		}
		seenWalletIDs[ws.MultisigWallets[i].WalletId] = true
		k.SetMultiSigWallet(ctx, &ws.MultisigWallets[i])
	}

	for i := range ws.PendingMultisigTxs {
		// CRITICAL SECURITY: Detect duplicate transaction IDs
		if seenTxIDs[ws.PendingMultisigTxs[i].TxId] {
			return fmt.Errorf("duplicate pending multisig tx ID: %s", ws.PendingMultisigTxs[i].TxId)
		}
		seenTxIDs[ws.PendingMultisigTxs[i].TxId] = true
		k.SetPendingMultiSigTx(ctx, &ws.PendingMultisigTxs[i])
	}

	for i := range ws.SocialRecoveryConfigs {
		k.SetSocialRecoveryConfig(ctx, &ws.SocialRecoveryConfigs[i])
	}

	for i := range ws.RecoveryRequests {
		// CRITICAL SECURITY: Detect duplicate request IDs
		if seenRequestIDs[ws.RecoveryRequests[i].RequestId] {
			return fmt.Errorf("duplicate recovery request ID: %s", ws.RecoveryRequests[i].RequestId)
		}
		seenRequestIDs[ws.RecoveryRequests[i].RequestId] = true
		k.SetRecoveryRequest(ctx, &ws.RecoveryRequests[i])
	}

	for i := range ws.SpendingLimits {
		k.SetSpendingLimit(ctx, &ws.SpendingLimits[i])
	}

	// Initialize incident response state (value type, always present)
	ir := genState.IncidentResponse

	seenIncidentIDs := make(map[string]bool)
	seenLogIDs := make(map[string]bool)

	for i := range ir.Incidents {
		// CRITICAL SECURITY: Detect duplicate incident IDs
		if seenIncidentIDs[ir.Incidents[i].IncidentId] {
			return fmt.Errorf("duplicate incident ID: %s", ir.Incidents[i].IncidentId)
		}
		seenIncidentIDs[ir.Incidents[i].IncidentId] = true
		k.SetIncident(ctx, &ir.Incidents[i])
	}

	for i := range ir.AuditLogs {
		// CRITICAL SECURITY: Detect duplicate log IDs
		if seenLogIDs[ir.AuditLogs[i].LogId] {
			return fmt.Errorf("duplicate audit log ID: %s", ir.AuditLogs[i].LogId)
		}
		seenLogIDs[ir.AuditLogs[i].LogId] = true
		k.SetAuditLogEntry(ctx, &ir.AuditLogs[i])
	}
	// Note: NextIncidentId is stored separately in KV store, not imported from genesis

	// Initialize cryptography state (value type, always present)
	crypto := genState.Cryptography

	seenScheduleIDs := make(map[string]bool)
	seenSchemeIDs := make(map[string]bool)
	seenProofIDs := make(map[string]bool)
	seenKeyIDs := make(map[string]bool)

	for i := range crypto.KeyRotationSchedules {
		// CRITICAL SECURITY: Detect duplicate schedule IDs
		if seenScheduleIDs[crypto.KeyRotationSchedules[i].Id] {
			return fmt.Errorf("duplicate key rotation schedule ID: %s", crypto.KeyRotationSchedules[i].Id)
		}
		seenScheduleIDs[crypto.KeyRotationSchedules[i].Id] = true
		k.SetKeyRotationSchedule(ctx, &crypto.KeyRotationSchedules[i])
	}

	for i := range crypto.ThresholdSchemes {
		// CRITICAL SECURITY: Detect duplicate scheme IDs
		if seenSchemeIDs[crypto.ThresholdSchemes[i].SchemeId] {
			return fmt.Errorf("duplicate threshold scheme ID: %s", crypto.ThresholdSchemes[i].SchemeId)
		}
		seenSchemeIDs[crypto.ThresholdSchemes[i].SchemeId] = true
		k.SetThresholdScheme(ctx, &crypto.ThresholdSchemes[i])
	}

	for i := range crypto.ZkProofConfigs {
		// CRITICAL SECURITY: Detect duplicate proof IDs
		if seenProofIDs[crypto.ZkProofConfigs[i].ProofId] {
			return fmt.Errorf("duplicate ZK proof config ID: %s", crypto.ZkProofConfigs[i].ProofId)
		}
		seenProofIDs[crypto.ZkProofConfigs[i].ProofId] = true
		k.SetZKProofConfig(ctx, &crypto.ZkProofConfigs[i])
	}

	for i := range crypto.QuantumResistantKeys {
		// CRITICAL SECURITY: Detect duplicate key IDs
		if seenKeyIDs[crypto.QuantumResistantKeys[i].KeyId] {
			return fmt.Errorf("duplicate quantum resistant key ID: %s", crypto.QuantumResistantKeys[i].KeyId)
		}
		seenKeyIDs[crypto.QuantumResistantKeys[i].KeyId] = true
		k.SetQuantumResistantKey(ctx, &crypto.QuantumResistantKeys[i])
	}

	// Initialize privacy state (value type, always present)
	priv := genState.Privacy

	seenPoolIDs := make(map[string]bool)

	for i := range priv.MixingPools {
		// CRITICAL SECURITY: Detect duplicate pool IDs
		if seenPoolIDs[priv.MixingPools[i].PoolId] {
			return fmt.Errorf("duplicate mixing pool ID: %s", priv.MixingPools[i].PoolId)
		}
		seenPoolIDs[priv.MixingPools[i].PoolId] = true
		k.SetMixingPool(ctx, &priv.MixingPools[i])
	}

	for i := range priv.StealthAddresses {
		// Use hex encoding of OneTimeAddress as key since StealthAddress has no id field
		k.SetStealthAddress(ctx, &priv.StealthAddresses[i])
	}

	for i := range priv.RingSignatures {
		// Use hex encoding of KeyImage as key since RingSignature has no id field
		k.SetRingSignature(ctx, &priv.RingSignatures[i])
	}

	k.Logger(ctx).Info("security module genesis initialized")
	return nil
}

// ExportGenesis exports the module's state
func (k Keeper) ExportGenesis(ctx sdk.Context) *securitypb.GenesisState {
	params := k.GetParams(ctx)

	// Convert pointer slices to value slices for genesis export
	return &securitypb.GenesisState{
		Params: params,
		NetworkSecurity: securitypb.NetworkSecurityState{
			RateLimits:      dereferenceSlice(k.GetAllRateLimits(ctx)),
			Reputations:     dereferenceSlice(k.GetAllPeerReputations(ctx)),
			TrustedPeers:    dereferenceSlice(k.GetAllTrustedPeers(ctx)),
			ForkAlerts:      dereferenceSlice(k.GetAllForkAlerts(ctx)),
			PartitionAlerts: dereferenceSlice(k.GetAllPartitionAlerts(ctx)),
		},
		ValidatorSecurity: securitypb.ValidatorSecurityState{
			Validators:          dereferenceSlice(k.GetAllValidatorSecurityInfos(ctx)),
			DoubleSignEvidences: dereferenceSlice(k.GetAllDoubleSignEvidence(ctx)),
			DowntimeInfractions: dereferenceSlice(k.GetAllDowntimeInfractions(ctx)),
			Alerts:              dereferenceSlice(k.GetAllValidatorAlerts(ctx)),
			SentryNodes:         dereferenceSlice(k.GetAllSentryNodes(ctx)),
		},
		WalletSecurity: securitypb.WalletSecurityState{
			HardwareWallets:       dereferenceSlice(k.GetAllHardwareWallets(ctx)),
			MultisigWallets:       dereferenceSlice(k.GetAllMultiSigWallets(ctx)),
			PendingMultisigTxs:    dereferenceSlice(k.GetAllPendingMultiSigTxs(ctx)),
			SocialRecoveryConfigs: dereferenceSlice(k.GetAllSocialRecoveryConfigs(ctx)),
			RecoveryRequests:      dereferenceSlice(k.GetAllRecoveryRequests(ctx)),
			SpendingLimits:        dereferenceSlice(k.GetAllSpendingLimits(ctx)),
		},
		IncidentResponse: securitypb.IncidentResponseState{
			Incidents: dereferenceSlice(k.GetAllIncidents(ctx)),
			AuditLogs: dereferenceSlice(k.GetAllAuditLogEntries(ctx)),
			// Note: NextIncidentId is not part of IncidentResponseState proto
			// It's stored separately in KV store
		},
		Cryptography: securitypb.CryptographyState{
			KeyRotationSchedules: dereferenceSlice(k.GetAllKeyRotationSchedules(ctx)),
			ThresholdSchemes:     dereferenceSlice(k.GetAllThresholdSchemes(ctx)),
			ZkProofConfigs:       dereferenceSlice(k.GetAllZKProofConfigs(ctx)),
			QuantumResistantKeys: dereferenceSlice(k.GetAllQuantumResistantKeys(ctx)),
		},
		Privacy: securitypb.PrivacyState{
			MixingPools:      dereferenceSlice(k.GetAllMixingPools(ctx)),
			StealthAddresses: dereferenceSlice(k.GetAllStealthAddresses(ctx)),
			RingSignatures:   dereferenceSlice(k.GetAllRingSignatures(ctx)),
		},
	}
}

// dereferenceSlice converts a slice of pointers []*T to a slice of values []T
func dereferenceSlice[T any](ptrs []*T) []T {
	result := make([]T, len(ptrs))
	for i, ptr := range ptrs {
		if ptr != nil {
			result[i] = *ptr
		}
	}
	return result
}
