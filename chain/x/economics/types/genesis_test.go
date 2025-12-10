package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	economicstypes "github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// TestDefaultGenesisState verifies default genesis state is valid
func TestDefaultGenesisState(t *testing.T) {
	genesis := economicstypes.DefaultGenesisState()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.Params)
	require.Empty(t, genesis.VestingSchedules)
	require.Empty(t, genesis.Proposals)
	require.Empty(t, genesis.Votes)
	require.Empty(t, genesis.Deposits)
	require.Empty(t, genesis.VoteLocks)
	require.Empty(t, genesis.VoteDelegations)
	require.Empty(t, genesis.PendingTreasuryTxs)
	require.NotNil(t, genesis.UserMevBalances)
	require.NotNil(t, genesis.LastLargeTxTimes)
	require.Equal(t, uint64(1), genesis.NextProposalId)
	require.Equal(t, uint64(1), genesis.NextVestingScheduleId)
	require.Equal(t, uint64(1), genesis.NextVoteLockId)
	require.Equal(t, uint64(1), genesis.NextTreasuryTxId)

	// Should be valid
	err := economicstypes.ValidateGenesisState(genesis)
	require.NoError(t, err)
}

// TestValidateGenesisState tests genesis state validation
func TestValidateGenesisState(t *testing.T) {
	tests := []struct {
		name      string
		genesis   *economicspb.GenesisState
		expectErr bool
		errMsg    string
	}{
		{
			name:      "nil genesis",
			genesis:   nil,
			expectErr: true,
			errMsg:    "genesis state cannot be nil",
		},
		{
			name:      "valid default genesis",
			genesis:   economicstypes.DefaultGenesisState(),
			expectErr: false,
		},
		{
			name: "invalid params",
			genesis: &economicspb.GenesisState{
				Params: economicspb.Params{
					Fees: economicspb.FeeParams{
						MinFeeMultiplier: 200,
						MaxFeeMultiplier: 100, // Invalid: min > max
					},
					Vesting:         economicspb.VestingParams{},
					Treasury:        economicspb.TreasuryParams{MultisigThreshold: 1},
					Governance:      economicspb.GovernanceParams{},
					Mev:             economicspb.MEVParams{UserRedistributionPercentage: 5000, ValidatorPercentage: 3000, TreasuryPercentage: 1000, BurnPercentage: 1000},
					WhaleProtection: economicspb.WhaleProtectionParams{},
					LiquidityMining: economicspb.LiquidityMiningParams{EpochDurationBlocks: 100},
					Tokenomics:      economicspb.TokenomicsParams{InflationCheckInterval: 100},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "invalid params",
		},
		{
			name: "empty vesting schedule ID",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				VestingSchedules: []economicspb.VestingSchedule{
					{
						Id:      "", // Empty ID
						Address: "aura1test",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "ID cannot be empty",
		},
		{
			name: "duplicate vesting schedule IDs",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				VestingSchedules: []economicspb.VestingSchedule{
					{
						Id:      "schedule1",
						Address: "aura1test1",
					},
					{
						Id:      "schedule1", // Duplicate
						Address: "aura1test2",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "duplicate ID",
		},
		{
			name: "vesting schedule with empty address",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				VestingSchedules: []economicspb.VestingSchedule{
					{
						Id:      "schedule1",
						Address: "", // Empty address
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "address cannot be empty",
		},
		{
			name: "empty vote lock ID",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				VoteLocks: []economicspb.VoteLock{
					{
						Id:    "", // Empty ID
						Owner: "aura1test",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "ID cannot be empty",
		},
		{
			name: "duplicate vote lock IDs",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				VoteLocks: []economicspb.VoteLock{
					{
						Id:    "lock1",
						Owner: "aura1test1",
					},
					{
						Id:    "lock1", // Duplicate
						Owner: "aura1test2",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "duplicate ID",
		},
		{
			name: "vote lock with empty owner",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				VoteLocks: []economicspb.VoteLock{
					{
						Id:    "lock1",
						Owner: "", // Empty owner
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "owner cannot be empty",
		},
		{
			name: "zero proposal ID",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				Proposals: []economicspb.Proposal{
					{
						Id:       0, // Zero ID
						Proposer: "aura1test",
						Title:    "Test Proposal",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "ID cannot be zero",
		},
		{
			name: "duplicate proposal IDs",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				Proposals: []economicspb.Proposal{
					{
						Id:       1,
						Proposer: "aura1test1",
						Title:    "Proposal 1",
					},
					{
						Id:       1, // Duplicate
						Proposer: "aura1test2",
						Title:    "Proposal 2",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "duplicate ID",
		},
		{
			name: "proposal with empty proposer",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				Proposals: []economicspb.Proposal{
					{
						Id:       1,
						Proposer: "", // Empty proposer
						Title:    "Test",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "proposer cannot be empty",
		},
		{
			name: "proposal with empty title",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				Proposals: []economicspb.Proposal{
					{
						Id:       1,
						Proposer: "aura1test",
						Title:    "", // Empty title
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "title cannot be empty",
		},
		{
			name: "vote with zero proposal ID",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				Votes: []economicspb.Vote{
					{
						ProposalId: 0, // Zero proposal ID
						Voter:      "aura1test",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "proposal ID cannot be zero",
		},
		{
			name: "vote with empty voter",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				Votes: []economicspb.Vote{
					{
						ProposalId: 1,
						Voter:      "", // Empty voter
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "voter cannot be empty",
		},
		{
			name: "deposit with zero proposal ID",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				Deposits: []economicspb.Deposit{
					{
						ProposalId: 0, // Zero proposal ID
						Depositor:  "aura1test",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "proposal ID cannot be zero",
		},
		{
			name: "deposit with empty depositor",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				Deposits: []economicspb.Deposit{
					{
						ProposalId: 1,
						Depositor:  "", // Empty depositor
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "depositor cannot be empty",
		},
		{
			name: "pending treasury tx with empty ID",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				PendingTreasuryTxs: []economicspb.PendingTreasuryTx{
					{
						TxId:      "", // Empty ID
						Recipient: "aura1test",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "ID cannot be empty",
		},
		{
			name: "duplicate pending treasury tx IDs",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				PendingTreasuryTxs: []economicspb.PendingTreasuryTx{
					{
						TxId:      "tx1",
						Recipient: "aura1test1",
					},
					{
						TxId:      "tx1", // Duplicate
						Recipient: "aura1test2",
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "duplicate ID",
		},
		{
			name: "pending treasury tx with empty recipient",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				PendingTreasuryTxs: []economicspb.PendingTreasuryTx{
					{
						TxId:      "tx1",
						Recipient: "", // Empty recipient
					},
				},
				UserMevBalances:  make(map[string]string),
				LastLargeTxTimes: make(map[string]int64),
			},
			expectErr: true,
			errMsg:    "recipient cannot be empty",
		},
		{
			name: "valid genesis with all data",
			genesis: &economicspb.GenesisState{
				Params: *economicstypes.DefaultParams(),
				VestingSchedules: []economicspb.VestingSchedule{
					{
						Id:      "schedule1",
						Address: "aura1test1",
					},
					{
						Id:      "schedule2",
						Address: "aura1test2",
					},
				},
				Proposals: []economicspb.Proposal{
					{
						Id:       1,
						Proposer: "aura1test",
						Title:    "Test Proposal",
					},
				},
				Votes: []economicspb.Vote{
					{
						ProposalId: 1,
						Voter:      "aura1voter",
					},
				},
				Deposits: []economicspb.Deposit{
					{
						ProposalId: 1,
						Depositor:  "aura1depositor",
					},
				},
				VoteLocks: []economicspb.VoteLock{
					{
						Id:    "lock1",
						Owner: "aura1owner",
					},
				},
				VoteDelegations: []economicspb.VoteDelegation{
					{
						Delegator: "aura1delegator",
						Delegate: "aura1delegate",
					},
				},
				PendingTreasuryTxs: []economicspb.PendingTreasuryTx{
					{
						TxId:      "tx1",
						Recipient: "aura1recipient",
					},
				},
				UserMevBalances:       map[string]string{"aura1user": "1000"},
				LastLargeTxTimes:      map[string]int64{"aura1user": 12345},
				NextProposalId:        2,
				NextVestingScheduleId: 3,
				NextVoteLockId:        2,
				NextTreasuryTxId:      2,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := economicstypes.ValidateGenesisState(tt.genesis)

			if tt.expectErr {
				require.Error(t, err, "expected error for %s", tt.name)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg, "error message mismatch")
				}
			} else {
				require.NoError(t, err, "unexpected error for %s", tt.name)
			}
		})
	}
}

// TestValidateParamsProto tests the ValidateParamsProto function
func TestValidateParamsProto(t *testing.T) {
	tests := []struct {
		name      string
		params    *economicspb.Params
		expectErr bool
		errMsg    string
	}{
		{
			name:      "nil params",
			params:    nil,
			expectErr: true,
			errMsg:    "params cannot be nil",
		},
		{
			name:      "valid default params",
			params:    economicstypes.DefaultParams(),
			expectErr: false,
		},
		{
			name: "invalid target block utilization",
			params: &economicspb.Params{
				Fees: economicspb.FeeParams{
					TargetBlockUtilization: 10001,
				},
				Vesting:         economicspb.VestingParams{},
				Treasury:        economicspb.TreasuryParams{},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{},
				Tokenomics:      economicspb.TokenomicsParams{},
			},
			expectErr: true,
			errMsg:    "target block utilization cannot exceed 100%",
		},
		{
			name: "invalid fee multipliers",
			params: &economicspb.Params{
				Fees: economicspb.FeeParams{
					MinFeeMultiplier: 200,
					MaxFeeMultiplier: 100,
				},
				Vesting:         economicspb.VestingParams{},
				Treasury:        economicspb.TreasuryParams{},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{},
				Tokenomics:      economicspb.TokenomicsParams{},
			},
			expectErr: true,
			errMsg:    "max fee multiplier must be >= min fee multiplier",
		},
		{
			name: "invalid vesting duration",
			params: &economicspb.Params{
				Fees: economicspb.FeeParams{},
				Vesting: economicspb.VestingParams{
					MinVestingDuration: 100,
					MaxVestingDuration: 50,
				},
				Treasury:        economicspb.TreasuryParams{},
				Governance:      economicspb.GovernanceParams{},
				Mev:             economicspb.MEVParams{},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{},
				Tokenomics:      economicspb.TokenomicsParams{},
			},
			expectErr: true,
			errMsg:    "max vesting duration must be >= min vesting duration",
		},
		{
			name: "invalid governance quorum",
			params: &economicspb.Params{
				Fees:    economicspb.FeeParams{},
				Vesting: economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{},
				Governance: economicspb.GovernanceParams{
					Quorum: 10001,
				},
				Mev:             economicspb.MEVParams{},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{},
				Tokenomics:      economicspb.TokenomicsParams{},
			},
			expectErr: true,
			errMsg:    "governance quorum cannot exceed 100%",
		},
		{
			name: "invalid governance threshold",
			params: &economicspb.Params{
				Fees:     economicspb.FeeParams{},
				Vesting:  economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{},
				Governance: economicspb.GovernanceParams{
					Threshold: 10001,
				},
				Mev:             economicspb.MEVParams{},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{},
				Tokenomics:      economicspb.TokenomicsParams{},
			},
			expectErr: true,
			errMsg:    "governance threshold cannot exceed 100%",
		},
		{
			name: "invalid governance veto threshold",
			params: &economicspb.Params{
				Fees:     economicspb.FeeParams{},
				Vesting:  economicspb.VestingParams{},
				Treasury: economicspb.TreasuryParams{},
				Governance: economicspb.GovernanceParams{
					VetoThreshold: 10001,
				},
				Mev:             economicspb.MEVParams{},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{},
				Tokenomics:      economicspb.TokenomicsParams{},
			},
			expectErr: true,
			errMsg:    "governance veto threshold cannot exceed 100%",
		},
		{
			name: "invalid tokenomics inflation rates",
			params: &economicspb.Params{
				Fees:       economicspb.FeeParams{},
				Vesting:    economicspb.VestingParams{},
				Treasury:   economicspb.TreasuryParams{},
				Governance: economicspb.GovernanceParams{},
				Mev:        economicspb.MEVParams{},
				WhaleProtection: economicspb.WhaleProtectionParams{},
				LiquidityMining: economicspb.LiquidityMiningParams{},
				Tokenomics: economicspb.TokenomicsParams{
					MinInflationRate: 1000,
					MaxInflationRate: 500,
				},
			},
			expectErr: true,
			errMsg:    "max inflation rate must be >= min inflation rate",
		},
		{
			name: "invalid whale protection max holding",
			params: &economicspb.Params{
				Fees:       economicspb.FeeParams{},
				Vesting:    economicspb.VestingParams{},
				Treasury:   economicspb.TreasuryParams{},
				Governance: economicspb.GovernanceParams{},
				Mev:        economicspb.MEVParams{},
				WhaleProtection: economicspb.WhaleProtectionParams{
					MaxHoldingPercentage: 10001,
				},
				LiquidityMining: economicspb.LiquidityMiningParams{},
				Tokenomics:      economicspb.TokenomicsParams{},
			},
			expectErr: true,
			errMsg:    "max holding percentage cannot exceed 100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := economicstypes.ValidateParamsProto(tt.params)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
