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

	// Store params
	if genState.Params != nil {
		k.SetParams(ctx, *genState.Params)
	}

	// Initialize network security state
	if genState.NetworkSecurity != nil {
		ns := genState.NetworkSecurity

		// Track seen IDs for duplicate detection
		seenPeerIDs := make(map[string]bool)
		seenAlertIDs := make(map[string]bool)

		for _, rl := range ns.RateLimits {
			// CRITICAL SECURITY: Detect duplicate peer IDs in rate limits
			if seenPeerIDs[rl.PeerId] {
				return fmt.Errorf("duplicate peer ID in rate limits: %s", rl.PeerId)
			}
			seenPeerIDs[rl.PeerId] = true
			k.SetRateLimit(ctx, rl)
		}

		seenPeerIDs = make(map[string]bool)
		for _, rep := range ns.Reputations {
			// CRITICAL SECURITY: Detect duplicate peer IDs in reputations
			if seenPeerIDs[rep.PeerId] {
				return fmt.Errorf("duplicate peer ID in reputations: %s", rep.PeerId)
			}
			seenPeerIDs[rep.PeerId] = true
			k.SetPeerReputation(ctx, rep)
		}

		seenPeerIDs = make(map[string]bool)
		for _, peer := range ns.TrustedPeers {
			// CRITICAL SECURITY: Detect duplicate peer IDs in trusted peers
			if seenPeerIDs[peer.PeerId] {
				return fmt.Errorf("duplicate peer ID in trusted peers: %s", peer.PeerId)
			}
			seenPeerIDs[peer.PeerId] = true
			k.SetTrustedPeer(ctx, peer)
		}

		for _, alert := range ns.ForkAlerts {
			// CRITICAL SECURITY: Detect duplicate alert IDs
			if seenAlertIDs[alert.AlertId] {
				return fmt.Errorf("duplicate fork alert ID: %s", alert.AlertId)
			}
			seenAlertIDs[alert.AlertId] = true
			k.SetForkAlert(ctx, alert)
		}

		seenAlertIDs = make(map[string]bool)
		for _, alert := range ns.PartitionAlerts {
			// CRITICAL SECURITY: Detect duplicate alert IDs
			if seenAlertIDs[alert.AlertId] {
				return fmt.Errorf("duplicate partition alert ID: %s", alert.AlertId)
			}
			seenAlertIDs[alert.AlertId] = true
			k.SetPartitionAlert(ctx, alert)
		}
	}

	// Initialize validator security state
	if genState.ValidatorSecurity != nil {
		vs := genState.ValidatorSecurity

		seenValidators := make(map[string]bool)
		seenAlertIDs := make(map[string]bool)

		for _, vi := range vs.Validators {
			// CRITICAL SECURITY: Detect duplicate validator addresses
			if seenValidators[vi.ValidatorAddress] {
				return fmt.Errorf("duplicate validator address in security info: %s", vi.ValidatorAddress)
			}
			seenValidators[vi.ValidatorAddress] = true
			k.SetValidatorSecurityInfo(ctx, vi)
		}

		for _, ev := range vs.DoubleSignEvidences {
			k.SetDoubleSignEvidence(ctx, ev)
		}

		for _, inf := range vs.DowntimeInfractions {
			k.SetDowntimeInfraction(ctx, inf)
		}

		for _, alert := range vs.Alerts {
			// CRITICAL SECURITY: Detect duplicate alert IDs
			if seenAlertIDs[alert.Id] {
				return fmt.Errorf("duplicate validator alert ID: %s", alert.Id)
			}
			seenAlertIDs[alert.Id] = true
			k.SetValidatorAlert(ctx, alert)
		}

		for _, sn := range vs.SentryNodes {
			k.SetSentryNode(ctx, sn)
		}
	}

	// Initialize wallet security state
	if genState.WalletSecurity != nil {
		ws := genState.WalletSecurity

		seenWalletIDs := make(map[string]bool)
		seenTxIDs := make(map[string]bool)
		seenRequestIDs := make(map[string]bool)

		for _, hw := range ws.HardwareWallets {
			// CRITICAL SECURITY: Detect duplicate wallet IDs
			if seenWalletIDs[hw.WalletId] {
				return fmt.Errorf("duplicate hardware wallet ID: %s", hw.WalletId)
			}
			seenWalletIDs[hw.WalletId] = true
			k.SetHardwareWallet(ctx, hw)
		}

		seenWalletIDs = make(map[string]bool)
		for _, ms := range ws.MultisigWallets {
			// CRITICAL SECURITY: Detect duplicate wallet IDs
			if seenWalletIDs[ms.WalletId] {
				return fmt.Errorf("duplicate multisig wallet ID: %s", ms.WalletId)
			}
			seenWalletIDs[ms.WalletId] = true
			k.SetMultiSigWallet(ctx, ms)
		}

		for _, tx := range ws.PendingMultisigTxs {
			// CRITICAL SECURITY: Detect duplicate transaction IDs
			if seenTxIDs[tx.TxId] {
				return fmt.Errorf("duplicate pending multisig tx ID: %s", tx.TxId)
			}
			seenTxIDs[tx.TxId] = true
			k.SetPendingMultiSigTx(ctx, tx)
		}

		for _, src := range ws.SocialRecoveryConfigs {
			k.SetSocialRecoveryConfig(ctx, src)
		}

		for _, req := range ws.RecoveryRequests {
			// CRITICAL SECURITY: Detect duplicate request IDs
			if seenRequestIDs[req.RequestId] {
				return fmt.Errorf("duplicate recovery request ID: %s", req.RequestId)
			}
			seenRequestIDs[req.RequestId] = true
			k.SetRecoveryRequest(ctx, req)
		}

		for _, sl := range ws.SpendingLimits {
			k.SetSpendingLimit(ctx, sl)
		}
	}

	// Initialize incident response state
	if genState.IncidentResponse != nil {
		ir := genState.IncidentResponse

		seenIncidentIDs := make(map[string]bool)
		seenLogIDs := make(map[string]bool)

		for _, inc := range ir.Incidents {
			// CRITICAL SECURITY: Detect duplicate incident IDs
			if seenIncidentIDs[inc.IncidentId] {
				return fmt.Errorf("duplicate incident ID: %s", inc.IncidentId)
			}
			seenIncidentIDs[inc.IncidentId] = true
			k.SetIncident(ctx, inc)
		}

		for _, entry := range ir.AuditLogs {
			// CRITICAL SECURITY: Detect duplicate log IDs
			if seenLogIDs[entry.LogId] {
				return fmt.Errorf("duplicate audit log ID: %s", entry.LogId)
			}
			seenLogIDs[entry.LogId] = true
			k.SetAuditLogEntry(ctx, entry)
		}
		// Note: NextIncidentId is stored separately in KV store, not imported from genesis
	}

	// Initialize cryptography state
	if genState.Cryptography != nil {
		crypto := genState.Cryptography

		seenScheduleIDs := make(map[string]bool)
		seenSchemeIDs := make(map[string]bool)
		seenProofIDs := make(map[string]bool)
		seenKeyIDs := make(map[string]bool)

		for _, krs := range crypto.KeyRotationSchedules {
			// CRITICAL SECURITY: Detect duplicate schedule IDs
			if seenScheduleIDs[krs.Id] {
				return fmt.Errorf("duplicate key rotation schedule ID: %s", krs.Id)
			}
			seenScheduleIDs[krs.Id] = true
			k.SetKeyRotationSchedule(ctx, krs)
		}

		for _, ts := range crypto.ThresholdSchemes {
			// CRITICAL SECURITY: Detect duplicate scheme IDs
			if seenSchemeIDs[ts.SchemeId] {
				return fmt.Errorf("duplicate threshold scheme ID: %s", ts.SchemeId)
			}
			seenSchemeIDs[ts.SchemeId] = true
			k.SetThresholdScheme(ctx, ts)
		}

		for _, zk := range crypto.ZkProofConfigs {
			// CRITICAL SECURITY: Detect duplicate proof IDs
			if seenProofIDs[zk.ProofId] {
				return fmt.Errorf("duplicate ZK proof config ID: %s", zk.ProofId)
			}
			seenProofIDs[zk.ProofId] = true
			k.SetZKProofConfig(ctx, zk)
		}

		for _, qrk := range crypto.QuantumResistantKeys {
			// CRITICAL SECURITY: Detect duplicate key IDs
			if seenKeyIDs[qrk.KeyId] {
				return fmt.Errorf("duplicate quantum resistant key ID: %s", qrk.KeyId)
			}
			seenKeyIDs[qrk.KeyId] = true
			k.SetQuantumResistantKey(ctx, qrk)
		}
	}

	// Initialize privacy state
	if genState.Privacy != nil {
		priv := genState.Privacy

		seenPoolIDs := make(map[string]bool)

		for _, mp := range priv.MixingPools {
			// CRITICAL SECURITY: Detect duplicate pool IDs
			if seenPoolIDs[mp.PoolId] {
				return fmt.Errorf("duplicate mixing pool ID: %s", mp.PoolId)
			}
			seenPoolIDs[mp.PoolId] = true
			k.SetMixingPool(ctx, mp)
		}

		for _, sa := range priv.StealthAddresses {
			// Use hex encoding of OneTimeAddress as key since StealthAddress has no id field
			k.SetStealthAddress(ctx, sa)
		}

		for _, rs := range priv.RingSignatures {
			// Use hex encoding of KeyImage as key since RingSignature has no id field
			k.SetRingSignature(ctx, rs)
		}
	}

	k.Logger(ctx).Info("security module genesis initialized")
	return nil
}

// ExportGenesis exports the module's state
func (k Keeper) ExportGenesis(ctx sdk.Context) *securitypb.GenesisState {
	params := k.GetParams(ctx)

	return &securitypb.GenesisState{
		Params: &params,
		NetworkSecurity: &securitypb.NetworkSecurityState{
			RateLimits:      k.GetAllRateLimits(ctx),
			Reputations:     k.GetAllPeerReputations(ctx),
			TrustedPeers:    k.GetAllTrustedPeers(ctx),
			ForkAlerts:      k.GetAllForkAlerts(ctx),
			PartitionAlerts: k.GetAllPartitionAlerts(ctx),
		},
		ValidatorSecurity: &securitypb.ValidatorSecurityState{
			Validators:          k.GetAllValidatorSecurityInfos(ctx),
			DoubleSignEvidences: k.GetAllDoubleSignEvidence(ctx),
			DowntimeInfractions: k.GetAllDowntimeInfractions(ctx),
			Alerts:              k.GetAllValidatorAlerts(ctx),
			SentryNodes:         k.GetAllSentryNodes(ctx),
		},
		WalletSecurity: &securitypb.WalletSecurityState{
			HardwareWallets:       k.GetAllHardwareWallets(ctx),
			MultisigWallets:       k.GetAllMultiSigWallets(ctx),
			PendingMultisigTxs:    k.GetAllPendingMultiSigTxs(ctx),
			SocialRecoveryConfigs: k.GetAllSocialRecoveryConfigs(ctx),
			RecoveryRequests:      k.GetAllRecoveryRequests(ctx),
			SpendingLimits:        k.GetAllSpendingLimits(ctx),
		},
		IncidentResponse: &securitypb.IncidentResponseState{
			Incidents: k.GetAllIncidents(ctx),
			AuditLogs: k.GetAllAuditLogEntries(ctx),
			// Note: NextIncidentId is not part of IncidentResponseState proto
			// It's stored separately in KV store
		},
		Cryptography: &securitypb.CryptographyState{
			KeyRotationSchedules: k.GetAllKeyRotationSchedules(ctx),
			ThresholdSchemes:     k.GetAllThresholdSchemes(ctx),
			ZkProofConfigs:       k.GetAllZKProofConfigs(ctx),
			QuantumResistantKeys: k.GetAllQuantumResistantKeys(ctx),
		},
		Privacy: &securitypb.PrivacyState{
			MixingPools:      k.GetAllMixingPools(ctx),
			StealthAddresses: k.GetAllStealthAddresses(ctx),
			RingSignatures:   k.GetAllRingSignatures(ctx),
		},
	}
}
