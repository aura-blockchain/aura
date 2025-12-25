// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
	"github.com/aequitas/aura/chain/x/economics/types"
)

// Ensure QueryServer implements the QueryServer interface
var _ economicspb.QueryServer = queryServer{}

// queryServer is the query server implementation
type queryServer struct {
	economicspb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServer creates a new query server
func NewQueryServer(keeper *Keeper) economicspb.QueryServer {
	return &queryServer{keeper: keeper}
}

// Params queries the economics module parameters
func (qs queryServer) Params(goCtx context.Context, req *economicspb.QueryParamsRequest) (*economicspb.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := qs.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryParamsResponse{
		Params: *params,
	}, nil
}

// VestingSchedule queries a vesting schedule by ID
func (qs queryServer) VestingSchedule(goCtx context.Context, req *economicspb.QueryVestingScheduleRequest) (*economicspb.QueryVestingScheduleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	schedule, err := qs.keeper.GetVestingSchedule(ctx, req.ScheduleId)
	if err != nil {
		return nil, err
	}

	// Calculate vested and remaining amounts
	currentTimeUnix, err := qs.keeper.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}
	currentTime := time.Unix(currentTimeUnix, 0)

	vestedAmount, err := qs.keeper.CalculateVestedAmount(schedule, currentTime)
	if err != nil {
		return nil, err
	}

	remainingAmount := schedule.OriginalAmount.Amount.Sub(vestedAmount)

	return &economicspb.QueryVestingScheduleResponse{
		Schedule:        *schedule,
		VestedAmount:    vestedAmount,
		RemainingAmount: remainingAmount,
	}, nil
}

// VestingSchedulesByAddress queries all vesting schedules for an address
func (qs queryServer) VestingSchedulesByAddress(goCtx context.Context, req *economicspb.QueryVestingSchedulesByAddressRequest) (*economicspb.QueryVestingSchedulesByAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate address
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid address: %s", err)
	}

	store := ctx.KVStore(qs.keeper.storeKey)

	// Create prefix store for this address's vesting schedules
	userVestingPrefix := append(types.UserVestingIndexPrefix, addr.Bytes()...)
	userVestingStore := prefix.NewStore(store, userVestingPrefix)

	// Calculate totals
	currentTimeUnix, err := qs.keeper.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}
	currentTime := time.Unix(currentTimeUnix, 0)

	totalVested := math.ZeroInt()
	totalVesting := math.ZeroInt()
	var schedules []economicspb.VestingSchedule

	pageRes, err := query.Paginate(userVestingStore, req.Pagination, func(key, value []byte) error {
		// Get the schedule ID from the index
		scheduleID := string(value)

		// Retrieve the actual schedule
		schedule, err := qs.keeper.GetVestingSchedule(ctx, scheduleID)
		if err != nil {
			return fmt.Errorf("failed to get vesting schedule %s: %w", scheduleID, err)
		}

		// Calculate vested amount for this schedule
		vestedAmount, err := qs.keeper.CalculateVestedAmount(schedule, currentTime)
		if err != nil {
			return fmt.Errorf("failed to calculate vested amount: %w", err)
		}
		totalVested = totalVested.Add(vestedAmount)
		totalVesting = totalVesting.Add(schedule.OriginalAmount.Amount.Sub(vestedAmount))

		schedules = append(schedules, *schedule)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryVestingSchedulesByAddressResponse{
		Schedules:    schedules,
		TotalVested:  totalVested,
		TotalVesting: totalVesting,
		Pagination:   pageRes,
	}, nil
}

// AllVestingSchedules queries all vesting schedules
func (qs queryServer) AllVestingSchedules(goCtx context.Context, req *economicspb.QueryAllVestingSchedulesRequest) (*economicspb.QueryAllVestingSchedulesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.keeper.storeKey)

	// Create prefix store for vesting schedules
	vestingStore := prefix.NewStore(store, types.VestingSchedulePrefix)

	var schedules []economicspb.VestingSchedule
	pageRes, err := query.Paginate(vestingStore, req.Pagination, func(key, value []byte) error {
		var schedule economicspb.VestingSchedule
		if err := qs.keeper.cdc.Unmarshal(value, &schedule); err != nil {
			return fmt.Errorf("failed to unmarshal vesting schedule: %w", err)
		}
		schedules = append(schedules, schedule)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryAllVestingSchedulesResponse{
		Schedules:  schedules,
		Pagination: pageRes,
	}, nil
}

// Proposal queries a proposal by ID
func (qs queryServer) Proposal(goCtx context.Context, req *economicspb.QueryProposalRequest) (*economicspb.QueryProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	proposal, err := qs.keeper.GetProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryProposalResponse{
		Proposal: *proposal,
	}, nil
}

// Proposals queries all proposals
func (qs queryServer) Proposals(goCtx context.Context, req *economicspb.QueryProposalsRequest) (*economicspb.QueryProposalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.keeper.storeKey)

	// Create prefix store for proposals
	proposalStore := prefix.NewStore(store, types.ProposalPrefix)

	var proposals []economicspb.Proposal
	pageRes, err := query.Paginate(proposalStore, req.Pagination, func(key, value []byte) error {
		var proposal economicspb.Proposal
		if err := qs.keeper.cdc.Unmarshal(value, &proposal); err != nil {
			return fmt.Errorf("failed to unmarshal proposal: %w", err)
		}

		// Filter by status if specified
		if req.Status != economicspb.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED && proposal.Status != req.Status {
			return nil
		}

		// Filter by voter if specified
		if req.Voter != "" {
			voterAddr, err := sdk.AccAddressFromBech32(req.Voter)
			if err != nil {
				return nil
			}
			hasVoted, err := qs.keeper.HasVoted(ctx, proposal.Id, voterAddr)
			if err != nil || !hasVoted {
				return nil
			}
		}

		// Filter by depositor if specified
		if req.Depositor != "" {
			depositorAddr, err := sdk.AccAddressFromBech32(req.Depositor)
			if err != nil {
				return nil
			}
			hasDeposited, err := qs.keeper.HasDeposited(ctx, proposal.Id, depositorAddr)
			if err != nil || !hasDeposited {
				return nil
			}
		}

		proposals = append(proposals, proposal)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryProposalsResponse{
		Proposals:  proposals,
		Pagination: pageRes,
	}, nil
}

// Vote queries a vote by proposal ID and voter
func (qs queryServer) Vote(goCtx context.Context, req *economicspb.QueryVoteRequest) (*economicspb.QueryVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate voter address
	_, err := sdk.AccAddressFromBech32(req.Voter)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid voter address: %s", err)
	}

	vote, err := qs.keeper.GetVote(ctx, req.ProposalId, req.Voter)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryVoteResponse{
		Vote: *vote,
	}, nil
}

// Votes queries all votes for a proposal
func (qs queryServer) Votes(goCtx context.Context, req *economicspb.QueryVotesRequest) (*economicspb.QueryVotesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.keeper.storeKey)

	// Create prefix store for votes on this proposal
	proposalPrefix := make([]byte, 9)
	copy(proposalPrefix, types.VotePrefix)
	binary.BigEndian.PutUint64(proposalPrefix[1:], req.ProposalId)
	voteStore := prefix.NewStore(store, proposalPrefix)

	var votes []economicspb.Vote
	pageRes, err := query.Paginate(voteStore, req.Pagination, func(key, value []byte) error {
		var vote economicspb.Vote
		if err := qs.keeper.cdc.Unmarshal(value, &vote); err != nil {
			return fmt.Errorf("failed to unmarshal vote: %w", err)
		}
		votes = append(votes, vote)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryVotesResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

// Deposit queries a deposit by proposal ID and depositor
func (qs queryServer) Deposit(goCtx context.Context, req *economicspb.QueryDepositRequest) (*economicspb.QueryDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate depositor address
	depositor, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid depositor address: %s", err)
	}

	deposit, err := qs.keeper.GetDeposit(ctx, req.ProposalId, depositor.String())
	if err != nil {
		return nil, err
	}

	// Convert from *Deposit to Deposit
	var depositValue economicspb.Deposit
	if deposit != nil {
		depositValue = *deposit
	}

	return &economicspb.QueryDepositResponse{
		Deposit: depositValue,
	}, nil
}

// Deposits queries all deposits for a proposal
func (qs queryServer) Deposits(goCtx context.Context, req *economicspb.QueryDepositsRequest) (*economicspb.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.keeper.storeKey)

	// Create prefix store for deposits on this proposal
	proposalPrefix := make([]byte, 9)
	copy(proposalPrefix, types.DepositPrefix)
	binary.BigEndian.PutUint64(proposalPrefix[1:], req.ProposalId)
	depositStore := prefix.NewStore(store, proposalPrefix)

	var deposits []economicspb.Deposit
	pageRes, err := query.Paginate(depositStore, req.Pagination, func(key, value []byte) error {
		var deposit economicspb.Deposit
		if err := qs.keeper.cdc.Unmarshal(value, &deposit); err != nil {
			return fmt.Errorf("failed to unmarshal deposit: %w", err)
		}
		deposits = append(deposits, deposit)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryDepositsResponse{
		Deposits:   deposits,
		Pagination: pageRes,
	}, nil
}

// TallyResult queries the tally of a proposal
func (qs queryServer) TallyResult(goCtx context.Context, req *economicspb.QueryTallyResultRequest) (*economicspb.QueryTallyResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tally, err := qs.keeper.GetTallyResult(ctx, req.ProposalId)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryTallyResultResponse{
		Tally: *tally,
	}, nil
}

// VoteLock queries a vote lock by ID
func (qs queryServer) VoteLock(goCtx context.Context, req *economicspb.QueryVoteLockRequest) (*economicspb.QueryVoteLockResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lock, err := qs.keeper.GetVoteLock(ctx, req.LockId)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryVoteLockResponse{
		Lock: *lock,
	}, nil
}

// VoteLocksByOwner queries all vote locks for an owner
func (qs queryServer) VoteLocksByOwner(goCtx context.Context, req *economicspb.QueryVoteLocksByOwnerRequest) (*economicspb.QueryVoteLocksByOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid owner address: %s", err)
	}

	store := ctx.KVStore(qs.keeper.storeKey)

	// Create prefix store for this owner's vote locks
	userLockPrefix := append(types.UserVoteLockIndexPrefix, owner.Bytes()...)
	userLockStore := prefix.NewStore(store, userLockPrefix)

	// Calculate totals
	totalLocked := math.ZeroInt()
	totalVotingPower := math.ZeroInt()
	var locks []economicspb.VoteLock

	pageRes, err := query.Paginate(userLockStore, req.Pagination, func(key, value []byte) error {
		// Get the lock ID from the index
		lockID := string(value)

		// Retrieve the actual lock
		lock, err := qs.keeper.GetVoteLock(ctx, lockID)
		if err != nil {
			return fmt.Errorf("failed to get vote lock %s: %w", lockID, err)
		}

		totalLocked = totalLocked.Add(lock.Amount.Amount)
		totalVotingPower = totalVotingPower.Add(lock.VotingPower)

		locks = append(locks, *lock)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryVoteLocksByOwnerResponse{
		Locks:            locks,
		TotalLocked:      totalLocked,
		TotalVotingPower: totalVotingPower,
		Pagination:       pageRes,
	}, nil
}

// VotingPower queries the voting power for an address
func (qs queryServer) VotingPower(goCtx context.Context, req *economicspb.QueryVotingPowerRequest) (*economicspb.QueryVotingPowerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate address
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid address: %s", err)
	}

	votingPower, lockedAmount, delegatedPower, activeLocks, err := qs.keeper.CalculateVotingPower(ctx, addr, req.ProposalId)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryVotingPowerResponse{
		VotingPower:    votingPower,
		LockedAmount:   lockedAmount,
		DelegatedPower: delegatedPower,
		ActiveLocks:    activeLocks,
	}, nil
}

// VoteDelegations queries all vote delegations for a delegator
func (qs queryServer) VoteDelegations(goCtx context.Context, req *economicspb.QueryVoteDelegationsRequest) (*economicspb.QueryVoteDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate delegator address
	delegator, err := sdk.AccAddressFromBech32(req.Delegator)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid delegator address: %s", err)
	}

	store := ctx.KVStore(qs.keeper.storeKey)

	// Create prefix store for this delegator's vote delegations
	delegatorPrefix := append(types.VoteDelegationPrefix, []byte(delegator.String())...)
	delegationStore := prefix.NewStore(store, delegatorPrefix)

	var delegations []economicspb.VoteDelegation
	pageRes, err := query.Paginate(delegationStore, req.Pagination, func(key, value []byte) error {
		var delegation economicspb.VoteDelegation
		if err := qs.keeper.cdc.Unmarshal(value, &delegation); err != nil {
			return fmt.Errorf("failed to unmarshal vote delegation: %w", err)
		}
		delegations = append(delegations, delegation)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryVoteDelegationsResponse{
		Delegations: delegations,
		Pagination:  pageRes,
	}, nil
}

// PendingTreasuryTx queries a pending treasury transaction
func (qs queryServer) PendingTreasuryTx(goCtx context.Context, req *economicspb.QueryPendingTreasuryTxRequest) (*economicspb.QueryPendingTreasuryTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	tx, err := qs.keeper.GetPendingTreasuryTx(ctx, req.TxId)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryPendingTreasuryTxResponse{
		Transaction: *tx,
	}, nil
}

// PendingTreasuryTxs queries all pending treasury transactions
func (qs queryServer) PendingTreasuryTxs(goCtx context.Context, req *economicspb.QueryPendingTreasuryTxsRequest) (*economicspb.QueryPendingTreasuryTxsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.keeper.storeKey)

	// Create prefix store for pending treasury transactions
	treasuryStore := prefix.NewStore(store, types.PendingTreasuryTxPrefix)

	var txs []economicspb.PendingTreasuryTx
	pageRes, err := query.Paginate(treasuryStore, req.Pagination, func(key, value []byte) error {
		var tx economicspb.PendingTreasuryTx
		if err := qs.keeper.cdc.Unmarshal(value, &tx); err != nil {
			return fmt.Errorf("failed to unmarshal pending treasury tx: %w", err)
		}
		txs = append(txs, tx)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryPendingTreasuryTxsResponse{
		Transactions: txs,
		Pagination:   pageRes,
	}, nil
}

// InflationMetrics queries current inflation metrics
func (qs queryServer) InflationMetrics(goCtx context.Context, req *economicspb.QueryInflationMetricsRequest) (*economicspb.QueryInflationMetricsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	metrics, err := qs.keeper.GetInflationMetrics(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate 24h change (placeholder implementation)
	inflationChange24h := uint64(0)

	return &economicspb.QueryInflationMetricsResponse{
		Metrics:             *metrics,
		InflationChange_24H: inflationChange24h,
	}, nil
}

// MEVStats queries MEV redistribution statistics
func (qs queryServer) MEVStats(goCtx context.Context, req *economicspb.QueryMEVStatsRequest) (*economicspb.QueryMEVStatsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	stats, err := qs.keeper.GetMEVStats(ctx)
	if err != nil {
		return nil, err
	}

	params, err := qs.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	enabled := params.Mev.Enabled
	// Use default strategy (proportional to stake is most common)
	strategy := economicspb.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE
	if !enabled {
		strategy = economicspb.MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED
	}

	// Convert from *MEVStats to MEVStats
	var statsValue economicspb.MEVStats
	if stats != nil {
		statsValue = *stats
	}

	return &economicspb.QueryMEVStatsResponse{
		Stats:    statsValue,
		Enabled:  enabled,
		Strategy: strategy,
	}, nil
}

// UserMEVBalance queries a user's MEV redistribution balance
func (qs queryServer) UserMEVBalance(goCtx context.Context, req *economicspb.QueryUserMEVBalanceRequest) (*economicspb.QueryUserMEVBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate address
	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid address: %s", err)
	}

	// Get current MEV balance (GetUserMEVBalance returns string)
	balanceStr, err := qs.keeper.GetUserMEVBalance(ctx, req.Address)
	if err != nil {
		return nil, err
	}

	// Parse balance from string
	balance, ok := math.NewIntFromString(balanceStr)
	if !ok {
		balance = math.ZeroInt()
	}

	// Lifetime received is currently not tracked separately
	// For now, return the same as balance (would need separate tracking in future)
	lifetimeReceived := balance

	return &economicspb.QueryUserMEVBalanceResponse{
		Balance:          balance,
		LifetimeReceived: lifetimeReceived,
	}, nil
}

// LiquidityMiningStats queries liquidity mining statistics
func (qs queryServer) LiquidityMiningStats(goCtx context.Context, req *economicspb.QueryLiquidityMiningStatsRequest) (*economicspb.QueryLiquidityMiningStatsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	stats, err := qs.keeper.GetLiquidityMiningStats(ctx)
	if err != nil {
		return nil, err
	}

	params, err := qs.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	enabled := params.LiquidityMining.Enabled

	// Convert from *LiquidityMiningStats to LiquidityMiningStats
	var statsValue economicspb.LiquidityMiningStats
	if stats != nil {
		statsValue = *stats
	}

	return &economicspb.QueryLiquidityMiningStatsResponse{
		Stats:   statsValue,
		Enabled: enabled,
	}, nil
}

// TokenomicsStats queries overall tokenomics statistics
func (qs queryServer) TokenomicsStats(goCtx context.Context, req *economicspb.QueryTokenomicsStatsRequest) (*economicspb.QueryTokenomicsStatsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get params for max supply and other config
	params, err := qs.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// Get max supply from params (already a math.Int due to customtype)
	maxSupply := params.Tokenomics.MaxSupply
	currentInflationRate := params.Tokenomics.TargetInflationRate

	// Calculate circulating supply (placeholder)
	circulatingSupply := math.ZeroInt()

	// Calculate total vested and vesting
	totalVested := math.ZeroInt()
	totalVesting := math.ZeroInt()

	schedules, err := qs.keeper.GetAllVestingSchedules(ctx)
	if err == nil {
		currentTimeUnix, err := qs.keeper.GetCurrentTime(ctx)
		if err == nil {
			currentTime := time.Unix(currentTimeUnix, 0)
			for _, schedule := range schedules {
				vestedAmount, err := qs.keeper.CalculateVestedAmount(schedule, currentTime)
				if err != nil {
					continue
				}
				totalVested = totalVested.Add(vestedAmount)

				totalVesting = totalVesting.Add(schedule.OriginalAmount.Amount.Sub(vestedAmount))
			}
		}
	}

	// Calculate total locked governance
	totalLockedGovernance := math.ZeroInt()
	allLocks, err := qs.keeper.GetAllVoteLocks(ctx)
	if err == nil {
		for _, lock := range allLocks {
			totalLockedGovernance = totalLockedGovernance.Add(lock.Amount.Amount)
		}
	}

	// Get treasury balance (placeholder)
	treasuryBalance := math.ZeroInt()

	// Get total burned (placeholder)
	totalBurned := math.ZeroInt()

	// Get whale protection triggers 24h (placeholder)
	whaleProtectionTriggers24h := uint64(0)

	// Get transfer tax collected 24h (placeholder)
	transferTaxCollected24h := math.ZeroInt()

	return &economicspb.QueryTokenomicsStatsResponse{
		MaxSupply:                   maxSupply,
		CirculatingSupply:           circulatingSupply,
		TotalVested:                 totalVested,
		TotalVesting:                totalVesting,
		TotalLockedGovernance:       totalLockedGovernance,
		TreasuryBalance:             treasuryBalance,
		CurrentInflationRate:        currentInflationRate,
		TotalBurned:                 totalBurned,
		WhaleProtectionTriggers_24H: whaleProtectionTriggers24h,
		TransferTaxCollected_24H:    transferTaxCollected24h,
		Pagination:                  nil, // Not applicable for aggregate stats query
	}, nil
}
