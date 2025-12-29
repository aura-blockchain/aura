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

// Query limits to prevent DoS via unbounded iteration
const (
	// maxTokenomicsScheduleIterations limits vesting schedule iteration in TokenomicsStats
	maxTokenomicsScheduleIterations = 10000
	// maxTokenomicsLockIterations limits vote lock iteration in TokenomicsStats
	maxTokenomicsLockIterations = 10000
	// maxVestingSchedulesPerAddress limits schedules loaded per address query
	maxVestingSchedulesPerAddress = 1000
	// maxProposalScanLimit limits proposals scanned when using voter/depositor filters
	maxProposalScanLimit = 500
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

// VestingSchedulesByAddress queries all vesting schedules for an address.
// Limits iteration to maxVestingSchedulesPerAddress to prevent DoS.
// Accumulates totals during iteration to avoid loading all schedules into memory.
func (qs queryServer) VestingSchedulesByAddress(goCtx context.Context, req *economicspb.QueryVestingSchedulesByAddressRequest) (*economicspb.QueryVestingSchedulesByAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate address
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid address: %s", err)
	}

	// Get current time for vested amount calculation
	currentTimeUnix, err := qs.keeper.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}
	currentTime := time.Unix(currentTimeUnix, 0)

	// Parse pagination params
	var offset, limit uint64
	if req.Pagination != nil {
		offset = req.Pagination.Offset
		limit = req.Pagination.Limit
		if limit == 0 {
			limit = query.DefaultLimit
		}
	} else {
		limit = query.DefaultLimit
	}

	// Stream through schedules, accumulating totals and collecting paginated results
	totalVested := math.ZeroInt()
	totalVesting := math.ZeroInt()
	var schedules []economicspb.VestingSchedule
	var total uint64
	count := 0

	// Use bounded iteration - accumulate totals for all (up to limit), but only return paginated window
	allSchedules, err := qs.keeper.GetVestingSchedulesByAddress(ctx, addr)
	if err != nil {
		return nil, err
	}

	for i, schedule := range allSchedules {
		// Hard limit to prevent DoS
		if count >= maxVestingSchedulesPerAddress {
			break
		}
		count++
		total++

		// Calculate vested amounts for totals
		vestedAmount, calcErr := qs.keeper.CalculateVestedAmount(schedule, currentTime)
		if calcErr != nil {
			continue // Skip invalid schedules in total calculation
		}
		totalVested = totalVested.Add(vestedAmount)
		totalVesting = totalVesting.Add(schedule.OriginalAmount.Amount.Sub(vestedAmount))

		// Only add to results if within pagination window
		idx := uint64(i)
		if idx >= offset && idx < offset+limit && schedule != nil {
			schedules = append(schedules, *schedule)
		}
	}

	// Build pagination response
	pageRes := &query.PageResponse{Total: total}

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

// Proposals queries all proposals.
// When voter or depositor filters are specified, limits scan to maxProposalScanLimit
// proposals to prevent N+1 query patterns from causing timeouts.
func (qs queryServer) Proposals(goCtx context.Context, req *economicspb.QueryProposalsRequest) (*economicspb.QueryProposalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.keeper.storeKey)

	// Create prefix store for proposals
	proposalStore := prefix.NewStore(store, types.ProposalPrefix)

	// Track if we have expensive filters that cause N+1 queries
	hasExpensiveFilters := req.Voter != "" || req.Depositor != ""

	// Pre-validate addresses to avoid repeated parsing
	var voterAddr, depositorAddr sdk.AccAddress
	var voterErr, depositorErr error
	if req.Voter != "" {
		voterAddr, voterErr = sdk.AccAddressFromBech32(req.Voter)
	}
	if req.Depositor != "" {
		depositorAddr, depositorErr = sdk.AccAddressFromBech32(req.Depositor)
	}

	var proposals []economicspb.Proposal
	scannedCount := 0

	pageRes, err := query.Paginate(proposalStore, req.Pagination, func(key, value []byte) error {
		// Limit scan count when using expensive filters to prevent N+1 DoS
		if hasExpensiveFilters {
			scannedCount++
			if scannedCount > maxProposalScanLimit {
				return nil // Skip remaining proposals
			}
		}

		var proposal economicspb.Proposal
		if err := qs.keeper.cdc.Unmarshal(value, &proposal); err != nil {
			return fmt.Errorf("failed to unmarshal proposal: %w", err)
		}

		// Filter by status if specified (cheap filter - no extra lookups)
		if req.Status != economicspb.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED && proposal.Status != req.Status {
			return nil
		}

		// Filter by voter if specified (expensive - requires store lookup)
		if req.Voter != "" {
			if voterErr != nil {
				return nil
			}
			hasVoted, err := qs.keeper.HasVoted(ctx, proposal.Id, voterAddr)
			if err != nil || !hasVoted {
				return nil
			}
		}

		// Filter by depositor if specified (expensive - requires store lookup)
		if req.Depositor != "" {
			if depositorErr != nil {
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

	// Get all locks for this owner
	allLocks, err := qs.keeper.GetVoteLocksByOwner(ctx, owner)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	totalLocked := math.ZeroInt()
	totalVotingPower := math.ZeroInt()

	for _, lock := range allLocks {
		totalLocked = totalLocked.Add(lock.Amount.Amount)
		totalVotingPower = totalVotingPower.Add(lock.VotingPower)
	}

	// Apply in-memory pagination
	var locks []economicspb.VoteLock
	total := uint64(len(allLocks))

	// Parse pagination params
	var offset, limit uint64
	if req.Pagination != nil {
		offset = req.Pagination.Offset
		limit = req.Pagination.Limit
		if limit == 0 {
			limit = query.DefaultLimit
		}
	} else {
		limit = query.DefaultLimit
	}

	// Apply offset and limit
	start := offset
	end := offset + limit
	if end > total {
		end = total
	}

	for i := start; i < end; i++ {
		if allLocks[i] != nil {
			locks = append(locks, *allLocks[i])
		}
	}

	// Build pagination response
	pageRes := &query.PageResponse{Total: total}

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

	// Calculate 24h change based on previous rate stored during inflation adjustments
	// This tracks actual rate changes, not time-based sampling
	inflationChange24h := uint64(0)

	// Get the previous inflation rate (stored when rate was last adjusted)
	previousRate, err := qs.keeper.GetPreviousInflation(ctx)
	if err == nil && previousRate > 0 {
		// Check if adjustment was within last 24 hours
		if !metrics.LastAdjustment.IsZero() {
			blockTime := sdk.UnwrapSDKContext(goCtx).BlockTime()
			timeSinceAdjustment := blockTime.Sub(metrics.LastAdjustment)

			if timeSinceAdjustment <= 24*time.Hour {
				// Calculate absolute change in basis points
				currentRate := metrics.CurrentRate
				if currentRate >= previousRate {
					inflationChange24h = currentRate - previousRate
				} else {
					inflationChange24h = previousRate - currentRate
				}
			}
		}
	}

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

// TokenomicsStats queries overall tokenomics statistics.
// Uses bounded iteration to prevent DoS attacks - limits to maxTokenomicsScheduleIterations
// vesting schedules and maxTokenomicsLockIterations vote locks.
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

	// Calculate total vested and vesting using bounded iteration
	totalVested := math.ZeroInt()
	totalVesting := math.ZeroInt()

	currentTimeUnix, timeErr := qs.keeper.GetCurrentTime(ctx)
	if timeErr == nil {
		currentTime := time.Unix(currentTimeUnix, 0)
		scheduleCount := 0
		// Use bounded iteration instead of loading all schedules into memory
		_ = qs.keeper.IterateVestingSchedules(ctx, func(schedule *economicspb.VestingSchedule) bool {
			if scheduleCount >= maxTokenomicsScheduleIterations {
				return true // Stop iteration
			}
			scheduleCount++
			vestedAmount, err := qs.keeper.CalculateVestedAmount(schedule, currentTime)
			if err != nil {
				return false // Continue to next
			}
			totalVested = totalVested.Add(vestedAmount)
			totalVesting = totalVesting.Add(schedule.OriginalAmount.Amount.Sub(vestedAmount))
			return false
		})
	}

	// Calculate total locked governance using bounded iteration
	totalLockedGovernance := math.ZeroInt()
	lockCount := 0
	_ = qs.keeper.IterateVoteLocks(ctx, func(lock *economicspb.VoteLock) bool {
		if lockCount >= maxTokenomicsLockIterations {
			return true // Stop iteration
		}
		lockCount++
		totalLockedGovernance = totalLockedGovernance.Add(lock.Amount.Amount)
		return false
	})

	// Get treasury balance from stored stats
	treasuryBalanceStr, _ := qs.keeper.GetTreasuryBalance(ctx)
	treasuryBalance, ok := math.NewIntFromString(treasuryBalanceStr)
	if !ok {
		treasuryBalance = math.ZeroInt()
	}

	// Get total burned from stored stats
	totalBurnedStr, _ := qs.keeper.GetTotalBurned(ctx)
	totalBurned, ok := math.NewIntFromString(totalBurnedStr)
	if !ok {
		totalBurned = math.ZeroInt()
	}

	// Calculate circulating supply = maxSupply - totalVesting - totalLockedGovernance - treasuryBalance
	circulatingSupply := maxSupply.Sub(totalVesting).Sub(totalLockedGovernance).Sub(treasuryBalance)
	if circulatingSupply.IsNegative() {
		circulatingSupply = math.ZeroInt()
	}

	// Get 24h stats from stored metrics
	whaleProtectionTriggers24h := qs.keeper.GetWhaleProtectionTriggers24h(ctx)
	transferTaxCollected24hStr, _ := qs.keeper.GetTransferTaxCollected24h(ctx)
	transferTaxCollected24h, ok := math.NewIntFromString(transferTaxCollected24hStr)
	if !ok {
		transferTaxCollected24h = math.ZeroInt()
	}

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
