package economicsecurity

import (
	"context"

	"github.com/aequitas/aura/chain/x/economicsecurity/keeper"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// QueryServer implements the economic security module's Query service
type QueryServer struct {
	keeper *keeper.Keeper
	economicsecuritypb.UnimplementedQueryServer
}

// NewQueryServer creates a new QueryServer instance
func NewQueryServer(k *keeper.Keeper) economicsecuritypb.QueryServer {
	return &QueryServer{keeper: k}
}
// Params queries the module parameters
func (s *QueryServer) Params(ctx context.Context, req *economicsecuritypb.QueryParamsRequest) (*economicsecuritypb.QueryParamsResponse, error) {
	params := s.keeper.GetParams()
	return &economicsecuritypb.QueryParamsResponse{
		Params: &params,
	}, nil
}

// VestingSchedule queries a vesting schedule by ID
func (s *QueryServer) VestingSchedule(ctx context.Context, req *economicsecuritypb.QueryVestingScheduleRequest) (*economicsecuritypb.QueryVestingScheduleResponse, error) {
	schedule, ok := s.keeper.GetVestingSchedule(req.ScheduleId)
	if !ok {
		return nil, types.ErrVestingScheduleNotFound
	}

	// Calculate vested and remaining amounts
	// This is simplified - in production would use proper calculation
	return &economicsecuritypb.QueryVestingScheduleResponse{
		Schedule:        schedule,
		VestedAmount:    schedule.VestedAmount,
		RemainingAmount: "0", // Would calculate properly
		NextVestTime:    schedule.StartTime,
	}, nil
}

// VestingSchedulesByBeneficiary queries all vesting schedules for a beneficiary
func (s *QueryServer) VestingSchedulesByBeneficiary(ctx context.Context, req *economicsecuritypb.QueryVestingSchedulesByBeneficiaryRequest) (*economicsecuritypb.QueryVestingSchedulesByBeneficiaryResponse, error) {
	schedules := s.keeper.GetUserVestingSchedules(req.BeneficiaryAddress)
	totalVested, totalVesting := s.keeper.GetTotalVesting()

	return &economicsecuritypb.QueryVestingSchedulesByBeneficiaryResponse{
		Schedules:    schedules,
		TotalVested:  totalVested,
		TotalVesting: totalVesting,
	}, nil
}

// VoteLock queries a vote lock by ID
func (s *QueryServer) VoteLock(ctx context.Context, req *economicsecuritypb.QueryVoteLockRequest) (*economicsecuritypb.QueryVoteLockResponse, error) {
	lock, ok := s.keeper.GetVoteLock(req.LockId)
	if !ok {
		return nil, types.ErrVoteLockNotFound
	}

	return &economicsecuritypb.QueryVoteLockResponse{
		Lock: lock,
	}, nil
}

// VoteLocksByOwner queries all vote locks for an owner
func (s *QueryServer) VoteLocksByOwner(ctx context.Context, req *economicsecuritypb.QueryVoteLocksByOwnerRequest) (*economicsecuritypb.QueryVoteLocksByOwnerResponse, error) {
	locks := s.keeper.GetUserVoteLocks(req.Owner)
	votingPower, totalLocked, _ := s.keeper.GetVotingPower(req.Owner)

	return &economicsecuritypb.QueryVoteLocksByOwnerResponse{
		Locks:            locks,
		TotalLocked:      totalLocked,
		TotalVotingPower: votingPower,
	}, nil
}

// VotingPower queries the voting power for an address
func (s *QueryServer) VotingPower(ctx context.Context, req *economicsecuritypb.QueryVotingPowerRequest) (*economicsecuritypb.QueryVotingPowerResponse, error) {
	votingPower, lockedAmount, activeLocks := s.keeper.GetVotingPower(req.Address)

	return &economicsecuritypb.QueryVotingPowerResponse{
		VotingPower:  votingPower,
		LockedAmount: lockedAmount,
		ActiveLocks:  activeLocks,
	}, nil
}

// PendingTreasuryTx queries a pending treasury transaction
func (s *QueryServer) PendingTreasuryTx(ctx context.Context, req *economicsecuritypb.QueryPendingTreasuryTxRequest) (*economicsecuritypb.QueryPendingTreasuryTxResponse, error) {
	tx, ok := s.keeper.GetPendingTreasuryTx(req.TxId)
	if !ok {
		return nil, types.ErrTxNotFound
	}

	return &economicsecuritypb.QueryPendingTreasuryTxResponse{
		Transaction: tx,
	}, nil
}

// PendingTreasuryTxs queries all pending treasury transactions
func (s *QueryServer) PendingTreasuryTxs(ctx context.Context, req *economicsecuritypb.QueryPendingTreasuryTxsRequest) (*economicsecuritypb.QueryPendingTreasuryTxsResponse, error) {
	txs := s.keeper.GetAllPendingTreasuryTxs()

	return &economicsecuritypb.QueryPendingTreasuryTxsResponse{
		Transactions: txs,
	}, nil
}

// InflationMetrics queries current inflation metrics
func (s *QueryServer) InflationMetrics(ctx context.Context, req *economicsecuritypb.QueryInflationMetricsRequest) (*economicsecuritypb.QueryInflationMetricsResponse, error) {
	params := s.keeper.GetParams()

	return &economicsecuritypb.QueryInflationMetricsResponse{
		CurrentInflationRate: params.Tokenomics.InflationRate,
		TargetInflationRate:  params.Tokenomics.TargetInflationRate,
		InflationChange_24H:   0, // Would calculate from historical data
		LastAdjustment:       params.Tokenomics.LastInflationAdjustment,
		NextCheck:            params.Tokenomics.LastInflationCheck,
	}, nil
}

// InflationAlerts queries inflation alerts
func (s *QueryServer) InflationAlerts(ctx context.Context, req *economicsecuritypb.QueryInflationAlertsRequest) (*economicsecuritypb.QueryInflationAlertsResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}

	alerts := s.keeper.GetInflationAlerts(limit)

	return &economicsecuritypb.QueryInflationAlertsResponse{
		Alerts: alerts,
	}, nil
}

// LiquidityMiningStats queries liquidity mining statistics
func (s *QueryServer) LiquidityMiningStats(ctx context.Context, req *economicsecuritypb.QueryLiquidityMiningStatsRequest) (*economicsecuritypb.QueryLiquidityMiningStatsResponse, error) {
	enabled, allocated, distributed, remaining, epoch, nextDistHeight := s.keeper.GetLiquidityMiningStats()

	return &economicsecuritypb.QueryLiquidityMiningStatsResponse{
		Enabled:                enabled,
		TotalAllocated:         allocated,
		TotalDistributed:       distributed,
		RemainingRewards:       remaining,
		CurrentEpoch:           epoch,
		RewardsThisEpoch:       "0", // Would track per-epoch
		NextDistributionHeight: nextDistHeight,
	}, nil
}

// MEVStats queries MEV redistribution statistics
func (s *QueryServer) MEVStats(ctx context.Context, req *economicsecuritypb.QueryMEVStatsRequest) (*economicsecuritypb.QueryMEVStatsResponse, error) {
	enabled, captured, redistributed, pending, userPct, strategy := s.keeper.GetMEVStats()

	return &economicsecuritypb.QueryMEVStatsResponse{
		Enabled:                      enabled,
		TotalCaptured:                captured,
		TotalRedistributed:           redistributed,
		PendingRedistribution:        pending,
		UserRedistributionPercentage: userPct,
		Strategy:                     strategy,
	}, nil
}

// UserMEVBalance queries a user's MEV redistribution balance
func (s *QueryServer) UserMEVBalance(ctx context.Context, req *economicsecuritypb.QueryUserMEVBalanceRequest) (*economicsecuritypb.QueryUserMEVBalanceResponse, error) {
	balance := s.keeper.GetUserMEVBalance(req.Address)

	return &economicsecuritypb.QueryUserMEVBalanceResponse{
		Balance:          balance,
		LifetimeReceived: "0", // Would track lifetime totals
	}, nil
}

// TokenomicsStats queries overall tokenomics statistics
func (s *QueryServer) TokenomicsStats(ctx context.Context, req *economicsecuritypb.QueryTokenomicsStatsRequest) (*economicsecuritypb.QueryTokenomicsStatsResponse, error) {
	params := s.keeper.GetParams()
	totalVested, totalVesting := s.keeper.GetTotalVesting()
	totalLocked := s.keeper.GetTotalLockedGovernance()
	whaleTriggers := s.keeper.GetWhaleProtectionTriggers_24H()
	taxCollected := s.keeper.GetTaxCollected24h()

	return &economicsecuritypb.QueryTokenomicsStatsResponse{
		MaxSupply:                  params.Tokenomics.MaxSupply,
		CirculatingSupply:          params.Tokenomics.CirculatingSupply,
		TotalVested:                totalVested,
		TotalVesting:               totalVesting,
		TotalLockedGovernance:      totalLocked,
		TreasuryBalance:            "0", // Would fetch from bank module
		CurrentInflationRate:       params.Tokenomics.InflationRate,
		TotalBurned:                "0", // Would track
		WhaleProtectionTriggers_24H: whaleTriggers,
		TransferTaxCollected_24H:    taxCollected,
	}, nil
}
