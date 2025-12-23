package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil"
	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

func TestInitGenesisDefault(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Default genesis should initialize cleanly
	defaultGenesis := types.DefaultGenesis()
	err := k.InitGenesis(ctx, defaultGenesis)
	require.NoError(t, err)

	// Verify params were set
	params := k.GetParams(ctx)
	require.NotNil(t, params)
}

func TestInitGenesisWithData(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()
	genesis := &securitypb.GenesisState{
		Params: types.DefaultParams(),
		NetworkSecurity: securitypb.NetworkSecurityState{
			RateLimits: []securitypb.RateLimit{
				{
					PeerId:         "peer-1",
					RequestsPerMin: 100,
				},
			},
			Reputations: []securitypb.PeerReputation{
				{
					PeerId:     "peer-1",
					Score:      80,
					LastUpdate: now,
				},
			},
			TrustedPeers: []securitypb.TrustedPeer{
				{
					PeerId:    "peer-trusted",
					PublicKey: []byte("pubkey"),
				},
			},
		},
		ValidatorSecurity: securitypb.ValidatorSecurityState{
			Validators: []securitypb.ValidatorSecurityInfo{
				{
					ValidatorAddress: "auravaloper1test",
					SecurityScore:    90,
				},
			},
		},
		WalletSecurity: securitypb.WalletSecurityState{
			HardwareWallets: []securitypb.HardwareWallet{
				{
					WalletId: "hw-1",
					DeviceId: "device-123",
				},
			},
			MultisigWallets: []securitypb.MultiSigWallet{
				{
					WalletId:  "multisig-1",
					Threshold: 2,
					Signers:   []string{"signer1", "signer2", "signer3"},
				},
			},
		},
		IncidentResponse: securitypb.IncidentResponseState{
			Incidents: []securitypb.Incident{
				{
					IncidentId: "incident-1",
					Severity:   securitypb.IncidentSeverity_CRITICAL,
					Status:     securitypb.IncidentStatus_OPEN,
				},
			},
		},
		Cryptography: securitypb.CryptographyState{
			KeyRotationSchedules: []securitypb.KeyRotationSchedule{
				{
					Id:       "rotation-1",
					Interval: 86400,
				},
			},
			ThresholdSchemes: []securitypb.ThresholdScheme{
				{
					SchemeId:  "scheme-1",
					Threshold: 2,
					Total:     3,
				},
			},
		},
		Privacy: securitypb.PrivacyState{
			MixingPools: []securitypb.MixingPool{
				{
					PoolId:          "pool-1",
					MinParticipants: 2,
					MaxParticipants: 10,
					Denomination:    sdkmath.NewInt(1000000),
					Status:          "active",
				},
			},
		},
	}

	err := k.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify data was imported
	rateLimits := k.GetAllRateLimits(ctx)
	require.Len(t, rateLimits, 1)

	reputations := k.GetAllPeerReputations(ctx)
	require.Len(t, reputations, 1)

	hwWallets := k.GetAllHardwareWallets(ctx)
	require.Len(t, hwWallets, 1)

	incidents := k.GetAllIncidents(ctx)
	require.Len(t, incidents, 1)
}

func TestExportGenesis(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Initialize with default genesis
	defaultGenesis := types.DefaultGenesis()
	err := k.InitGenesis(ctx, defaultGenesis)
	require.NoError(t, err)

	// Export genesis
	exported := k.ExportGenesis(ctx)
	require.NotNil(t, exported)
	require.NotNil(t, exported.Params)
}

func TestGenesisRoundTrip(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()

	// Create a comprehensive genesis state
	genesis := &securitypb.GenesisState{
		Params: types.DefaultParams(),
		NetworkSecurity: securitypb.NetworkSecurityState{
			RateLimits: []securitypb.RateLimit{
				{
					PeerId:         "peer-1",
					RequestsPerMin: 100,
				},
				{
					PeerId:         "peer-2",
					RequestsPerMin: 50,
				},
			},
			Reputations: []securitypb.PeerReputation{
				{
					PeerId:     "peer-1",
					Score:      80,
					LastUpdate: now,
				},
				{
					PeerId:     "peer-2",
					Score:      90,
					LastUpdate: now,
				},
			},
			TrustedPeers: []securitypb.TrustedPeer{
				{
					PeerId:    "peer-trusted-1",
					PublicKey: []byte("pubkey1"),
				},
				{
					PeerId:    "peer-trusted-2",
					PublicKey: []byte("pubkey2"),
				},
			},
			ForkAlerts: []securitypb.ForkAlert{
				{
					AlertId: "fork-1",
					Height:  100,
				},
			},
			PartitionAlerts: []securitypb.PartitionAlert{
				{
					AlertId: "partition-1",
				},
			},
		},
		ValidatorSecurity: securitypb.ValidatorSecurityState{
			Validators: []securitypb.ValidatorSecurityInfo{
				{
					ValidatorAddress: "auravaloper1test1",
					SecurityScore:    90,
				},
				{
					ValidatorAddress: "auravaloper1test2",
					SecurityScore:    85,
				},
			},
			Alerts: []securitypb.ValidatorSecurityAlert{
				{
					Id:               "alert-1",
					ValidatorAddress: "auravaloper1test1",
					Severity:         securitypb.AlertSeverity_ALERT_SEVERITY_HIGH,
				},
			},
		},
		WalletSecurity: securitypb.WalletSecurityState{
			HardwareWallets: []securitypb.HardwareWallet{
				{
					WalletId: "hw-1",
					DeviceId: "device-123",
				},
				{
					WalletId: "hw-2",
					DeviceId: "device-456",
				},
			},
			MultisigWallets: []securitypb.MultiSigWallet{
				{
					WalletId:  "multisig-1",
					Threshold: 2,
					Signers:   []string{"signer1", "signer2", "signer3"},
				},
			},
			PendingMultisigTxs: []securitypb.PendingMultiSigTx{
				{
					TxId:     "tx-1",
					WalletId: "multisig-1",
				},
			},
			RecoveryRequests: []securitypb.RecoveryRequest{
				{
					RequestId: "recovery-1",
					Address:   "aura1test",
				},
			},
		},
		IncidentResponse: securitypb.IncidentResponseState{
			Incidents: []securitypb.Incident{
				{
					IncidentId: "incident-1",
					Severity:   securitypb.IncidentSeverity_CRITICAL,
					Status:     securitypb.IncidentStatus_OPEN,
				},
				{
					IncidentId: "incident-2",
					Severity:   securitypb.IncidentSeverity_HIGH,
					Status:     securitypb.IncidentStatus_INVESTIGATING,
				},
			},
			AuditLogs: []securitypb.AuditLogEntry{
				{
					LogId:  "log-1",
					Action: "action-1",
				},
			},
		},
		Cryptography: securitypb.CryptographyState{
			KeyRotationSchedules: []securitypb.KeyRotationSchedule{
				{
					Id:       "rotation-1",
					Interval: 86400,
				},
			},
			ThresholdSchemes: []securitypb.ThresholdScheme{
				{
					SchemeId:  "scheme-1",
					Threshold: 2,
					Total:     3,
				},
				{
					SchemeId:  "scheme-2",
					Threshold: 3,
					Total:     5,
				},
			},
			ZkProofConfigs: []securitypb.ZKProofConfig{
				{
					ProofId: "proof-1",
				},
			},
			QuantumResistantKeys: []securitypb.QuantumResistantKey{
				{
					KeyId:     "qr-key-1",
					Algorithm: "CRYSTALS-Kyber",
				},
			},
		},
		Privacy: securitypb.PrivacyState{
			MixingPools: []securitypb.MixingPool{
				{
					PoolId:          "pool-1",
					MinParticipants: 2,
					MaxParticipants: 10,
					Denomination:    sdkmath.NewInt(1000000),
					Status:          "active",
				},
				{
					PoolId:          "pool-2",
					MinParticipants: 5,
					MaxParticipants: 20,
					Denomination:    sdkmath.NewInt(5000000),
					Status:          "active",
				},
			},
		},
	}

	// Import genesis
	err := k.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Export genesis (first export)
	exported1 := k.ExportGenesis(ctx)
	require.NotNil(t, exported1)

	// Verify exported data matches original counts
	require.Equal(t, len(genesis.NetworkSecurity.RateLimits), len(exported1.NetworkSecurity.RateLimits))
	require.Equal(t, len(genesis.NetworkSecurity.Reputations), len(exported1.NetworkSecurity.Reputations))
	require.Equal(t, len(genesis.NetworkSecurity.TrustedPeers), len(exported1.NetworkSecurity.TrustedPeers))
	require.Equal(t, len(genesis.ValidatorSecurity.Validators), len(exported1.ValidatorSecurity.Validators))
	require.Equal(t, len(genesis.WalletSecurity.HardwareWallets), len(exported1.WalletSecurity.HardwareWallets))
	require.Equal(t, len(genesis.WalletSecurity.MultisigWallets), len(exported1.WalletSecurity.MultisigWallets))
	require.Equal(t, len(genesis.IncidentResponse.Incidents), len(exported1.IncidentResponse.Incidents))
	require.Equal(t, len(genesis.Cryptography.KeyRotationSchedules), len(exported1.Cryptography.KeyRotationSchedules))
	require.Equal(t, len(genesis.Cryptography.ThresholdSchemes), len(exported1.Cryptography.ThresholdSchemes))
	require.Equal(t, len(genesis.Privacy.MixingPools), len(exported1.Privacy.MixingPools))

	// Create a fresh keeper for re-import
	ctx2, k2 := testutil.NewSecurityKeeperForTest(t)

	// Re-import the exported genesis
	err = k2.InitGenesis(ctx2, exported1)
	require.NoError(t, err)

	// Export again (second export)
	exported2 := k2.ExportGenesis(ctx2)
	require.NotNil(t, exported2)

	// The two exports should be identical
	require.Equal(t, len(exported1.NetworkSecurity.RateLimits), len(exported2.NetworkSecurity.RateLimits))
	require.Equal(t, len(exported1.NetworkSecurity.Reputations), len(exported2.NetworkSecurity.Reputations))
	require.Equal(t, len(exported1.NetworkSecurity.TrustedPeers), len(exported2.NetworkSecurity.TrustedPeers))
	require.Equal(t, len(exported1.NetworkSecurity.ForkAlerts), len(exported2.NetworkSecurity.ForkAlerts))
	require.Equal(t, len(exported1.NetworkSecurity.PartitionAlerts), len(exported2.NetworkSecurity.PartitionAlerts))
	require.Equal(t, len(exported1.ValidatorSecurity.Validators), len(exported2.ValidatorSecurity.Validators))
	require.Equal(t, len(exported1.ValidatorSecurity.Alerts), len(exported2.ValidatorSecurity.Alerts))
	require.Equal(t, len(exported1.WalletSecurity.HardwareWallets), len(exported2.WalletSecurity.HardwareWallets))
	require.Equal(t, len(exported1.WalletSecurity.MultisigWallets), len(exported2.WalletSecurity.MultisigWallets))
	require.Equal(t, len(exported1.WalletSecurity.PendingMultisigTxs), len(exported2.WalletSecurity.PendingMultisigTxs))
	require.Equal(t, len(exported1.WalletSecurity.RecoveryRequests), len(exported2.WalletSecurity.RecoveryRequests))
	require.Equal(t, len(exported1.IncidentResponse.Incidents), len(exported2.IncidentResponse.Incidents))
	require.Equal(t, len(exported1.IncidentResponse.AuditLogs), len(exported2.IncidentResponse.AuditLogs))
	require.Equal(t, len(exported1.Cryptography.KeyRotationSchedules), len(exported2.Cryptography.KeyRotationSchedules))
	require.Equal(t, len(exported1.Cryptography.ThresholdSchemes), len(exported2.Cryptography.ThresholdSchemes))
	require.Equal(t, len(exported1.Cryptography.ZkProofConfigs), len(exported2.Cryptography.ZkProofConfigs))
	require.Equal(t, len(exported1.Cryptography.QuantumResistantKeys), len(exported2.Cryptography.QuantumResistantKeys))
	require.Equal(t, len(exported1.Privacy.MixingPools), len(exported2.Privacy.MixingPools))

	// Verify individual records match
	for i := range exported1.NetworkSecurity.RateLimits {
		require.Equal(t, exported1.NetworkSecurity.RateLimits[i].PeerId, exported2.NetworkSecurity.RateLimits[i].PeerId)
		require.Equal(t, exported1.NetworkSecurity.RateLimits[i].RequestsPerMin, exported2.NetworkSecurity.RateLimits[i].RequestsPerMin)
	}

	for i := range exported1.ValidatorSecurity.Validators {
		require.Equal(t, exported1.ValidatorSecurity.Validators[i].ValidatorAddress, exported2.ValidatorSecurity.Validators[i].ValidatorAddress)
		require.Equal(t, exported1.ValidatorSecurity.Validators[i].SecurityScore, exported2.ValidatorSecurity.Validators[i].SecurityScore)
	}

	for i := range exported1.WalletSecurity.HardwareWallets {
		require.Equal(t, exported1.WalletSecurity.HardwareWallets[i].WalletId, exported2.WalletSecurity.HardwareWallets[i].WalletId)
		require.Equal(t, exported1.WalletSecurity.HardwareWallets[i].DeviceId, exported2.WalletSecurity.HardwareWallets[i].DeviceId)
	}

	for i := range exported1.IncidentResponse.Incidents {
		require.Equal(t, exported1.IncidentResponse.Incidents[i].IncidentId, exported2.IncidentResponse.Incidents[i].IncidentId)
		require.Equal(t, exported1.IncidentResponse.Incidents[i].Severity, exported2.IncidentResponse.Incidents[i].Severity)
	}

	for i := range exported1.Privacy.MixingPools {
		require.Equal(t, exported1.Privacy.MixingPools[i].PoolId, exported2.Privacy.MixingPools[i].PoolId)
		require.Equal(t, exported1.Privacy.MixingPools[i].Status, exported2.Privacy.MixingPools[i].Status)
	}
}

func TestGenesisValidation(t *testing.T) {
	// Valid default genesis
	require.NoError(t, types.ValidateGenesisState(types.DefaultGenesis()))

	// Nil genesis should error
	require.Error(t, types.ValidateGenesisState(nil))
}
