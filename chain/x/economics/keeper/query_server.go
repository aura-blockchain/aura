package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
		Params: params,
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
		Schedule:        schedule,
		VestedAmount:    vestedAmount.String(),
		RemainingAmount: remainingAmount.String(),
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

	schedules, err := qs.keeper.GetVestingSchedulesByAddress(ctx, addr)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	currentTimeUnix, err := qs.keeper.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}
	currentTime := time.Unix(currentTimeUnix, 0)

	totalVested := math.ZeroInt()
	totalVesting := math.ZeroInt()

	for _, schedule := range schedules {
		vestedAmount, err := qs.keeper.CalculateVestedAmount(schedule, currentTime)
		if err != nil {
			return nil, err
		}
		totalVested = totalVested.Add(vestedAmount)

		totalVesting = totalVesting.Add(schedule.OriginalAmount.Amount.Sub(vestedAmount))
	}

	return &economicspb.QueryVestingSchedulesByAddressResponse{
		Schedules:     schedules,
		TotalVested:   totalVested.String(),
		TotalVesting:  totalVesting.String(),
	}, nil
}

// AllVestingSchedules queries all vesting schedules
func (qs queryServer) AllVestingSchedules(goCtx context.Context, req *economicspb.QueryAllVestingSchedulesRequest) (*economicspb.QueryAllVestingSchedulesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	schedules, err := qs.keeper.GetAllVestingSchedules(ctx)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryAllVestingSchedulesResponse{
		Schedules: schedules,
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
		Proposal: proposal,
	}, nil
}

// Proposals queries all proposals
func (qs queryServer) Proposals(goCtx context.Context, req *economicspb.QueryProposalsRequest) (*economicspb.QueryProposalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	proposals := []*economicspb.Proposal{}
	err := qs.keeper.IterateProposals(ctx, func(proposal *economicspb.Proposal) bool {
		// Filter by status if specified
		if req.Status != economicspb.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED && proposal.Status != req.Status {
			return false
		}

		// Filter by voter if specified
		if req.Voter != "" {
			// Check if voter has voted
			voterAddr, err := sdk.AccAddressFromBech32(req.Voter)
			if err != nil {
				return false
			}
			hasVoted, err := qs.keeper.HasVoted(ctx, proposal.Id, voterAddr)
			if err != nil || !hasVoted {
				return false
			}
		}

		// Filter by depositor if specified
		if req.Depositor != "" {
			// Check if depositor has deposited
			depositorAddr, err := sdk.AccAddressFromBech32(req.Depositor)
			if err != nil {
				return false
			}
			hasDeposited, err := qs.keeper.HasDeposited(ctx, proposal.Id, depositorAddr)
			if err != nil || !hasDeposited {
				return false
			}
		}

		proposals = append(proposals, proposal)
		return false
	})
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryProposalsResponse{
		Proposals: proposals,
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
		Vote: vote,
	}, nil
}

// Votes queries all votes for a proposal
func (qs queryServer) Votes(goCtx context.Context, req *economicspb.QueryVotesRequest) (*economicspb.QueryVotesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	votes, err := qs.keeper.GetVotes(ctx, req.ProposalId)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryVotesResponse{
		Votes: votes,
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

	return &economicspb.QueryDepositResponse{
		Deposit: deposit,
	}, nil
}

// Deposits queries all deposits for a proposal
func (qs queryServer) Deposits(goCtx context.Context, req *economicspb.QueryDepositsRequest) (*economicspb.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	deposits, err := qs.keeper.GetDeposits(ctx, req.ProposalId)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryDepositsResponse{
		Deposits: deposits,
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
		Tally: tally,
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
		Lock: lock,
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

	locks, err := qs.keeper.GetVoteLocksByOwner(ctx, owner)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	totalLocked := math.ZeroInt()
	totalVotingPower := math.ZeroInt()

	for _, lock := range locks {
		totalLocked = totalLocked.Add(lock.Amount.Amount)

		// Parse voting power from string (customtype field)
		lockVotingPower, ok := math.NewIntFromString(lock.VotingPower)
		if ok {
			totalVotingPower = totalVotingPower.Add(lockVotingPower)
		}
	}

	return &economicspb.QueryVoteLocksByOwnerResponse{
		Locks:             locks,
		TotalLocked:       totalLocked.String(),
		TotalVotingPower:  totalVotingPower.String(),
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
		VotingPower:    votingPower.String(),
		LockedAmount:   lockedAmount.String(),
		DelegatedPower: delegatedPower.String(),
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

	delegations, err := qs.keeper.GetVoteDelegations(ctx, delegator)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryVoteDelegationsResponse{
		Delegations: delegations,
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
		Transaction: tx,
	}, nil
}

// PendingTreasuryTxs queries all pending treasury transactions
func (qs queryServer) PendingTreasuryTxs(goCtx context.Context, req *economicspb.QueryPendingTreasuryTxsRequest) (*economicspb.QueryPendingTreasuryTxsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	txs, err := qs.keeper.GetAllPendingTreasuryTxs(ctx)
	if err != nil {
		return nil, err
	}

	return &economicspb.QueryPendingTreasuryTxsResponse{
		Transactions: txs,
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
		Metrics:             metrics,
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

	enabled := params.Mev != nil && params.Mev.Enabled
	// Use default strategy (proportional to stake is most common)
	strategy := economicspb.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_STAKE
	if !enabled {
		strategy = economicspb.MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED
	}

	return &economicspb.QueryMEVStatsResponse{
		Stats:    stats,
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
		Balance:          balance.String(),
		LifetimeReceived: lifetimeReceived.String(),
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

	enabled := params.LiquidityMining != nil && params.LiquidityMining.Enabled

	return &economicspb.QueryLiquidityMiningStatsResponse{
		Stats:   stats,
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

	// Get max supply from params (it's a math.Int stored as string in proto)
	maxSupply := math.ZeroInt()
	currentInflationRate := uint64(0)
	if params.Tokenomics != nil {
		// MaxSupply is customtype math.Int (string in proto)
		var ok bool
		maxSupply, ok = math.NewIntFromString(params.Tokenomics.MaxSupply)
		if !ok {
			maxSupply = math.ZeroInt()
		}
		currentInflationRate = params.Tokenomics.TargetInflationRate
	}

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
		MaxSupply:                  maxSupply.String(),
		CirculatingSupply:          circulatingSupply.String(),
		TotalVested:                totalVested.String(),
		TotalVesting:               totalVesting.String(),
		TotalLockedGovernance:      totalLockedGovernance.String(),
		TreasuryBalance:            treasuryBalance.String(),
		CurrentInflationRate:       currentInflationRate,
		TotalBurned:                totalBurned.String(),
		WhaleProtectionTriggers_24H: whaleProtectionTriggers24h,
		TransferTaxCollected_24H:   transferTaxCollected24h.String(),
	}, nil
}
