package keeper

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)


// ============================
// VESTING OPERATIONS
// ============================

// CreateVestingSchedule creates a new vesting schedule with given parameters
func (k Keeper) CreateVestingSchedule(
	ctx context.Context,
	creator string,
	beneficiaryAddress string,
	totalAmount sdk.Coin,
	startTime time.Time,
	cliffDuration uint64,
	vestingDuration uint64,
	vestingType economicspb.VestingType,
	scheduleType economicspb.ScheduleType,
) (string, error) {
	// Validate amount > 0
	if totalAmount.Amount.IsZero() || totalAmount.Amount.IsNegative() {
		return "", errorsmod.Wrap(types.ErrInvalidAmount, "vesting amount must be greater than zero")
	}

	// Validate vesting duration > 0
	if vestingDuration == 0 {
		return "", errorsmod.Wrap(types.ErrInvalidRequest, "invalid vesting duration")
	}

	// Generate unique schedule ID using counter
	scheduleNum, err := k.GetNextScheduleID(ctx)
	if err != nil {
		return "", err
	}
	scheduleID := fmt.Sprintf("schedule-%s-%d", beneficiaryAddress, scheduleNum)

	// Calculate end time
	endTime := startTime.Add(time.Duration(vestingDuration) * time.Second)

	// Create schedule
	vestedCoin := sdk.NewCoin(totalAmount.Denom, sdkmath.ZeroInt())
	schedule := &economicspb.VestingSchedule{
		Id:              scheduleID,
		Address:         beneficiaryAddress,
		OriginalAmount:  totalAmount,
		VestedAmount:    vestedCoin,
		StartTime:       startTime,
		EndTime:         endTime,
		CliffDuration:   cliffDuration,
		VestingType:     vestingType,
		ScheduleType:    scheduleType,
		Revoked:         false,
	}

	// Store schedule
	if err := k.SetVestingSchedule(ctx, schedule); err != nil {
		return "", err
	}

	// Update user index
	if err := k.AddUserVestingSchedule(ctx, beneficiaryAddress, scheduleID); err != nil {
		return "", err
	}

	return scheduleID, nil
}

// ReleaseVestedTokens releases vested tokens for a beneficiary (wrapper for msg_server)
func (k Keeper) ReleaseVestedTokens(ctx context.Context, beneficiary sdk.AccAddress, scheduleID string) (sdk.Coin, error) {
	// Get the schedule
	schedule, err := k.GetVestingSchedule(ctx, scheduleID)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrScheduleNotFound, "schedule not found")
	}

	// Verify beneficiary
	if schedule.Address != beneficiary.String() {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrUnauthorized, "not the beneficiary of this schedule")
	}

	// Calculate vested amount
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime()
	vestedAmount, err := k.CalculateVestedAmount(schedule, currentTime)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Calculate amount to release (vested - already released)
	alreadyVested := schedule.VestedAmount.Amount
	toRelease := vestedAmount.Sub(alreadyVested)

	if toRelease.IsNegative() || toRelease.IsZero() {
		return sdk.NewCoin(schedule.OriginalAmount.Denom, sdkmath.ZeroInt()), nil
	}

	// Update schedule
	newVestedCoin := sdk.NewCoin(schedule.OriginalAmount.Denom, vestedAmount)
	schedule.VestedAmount = newVestedCoin
	if err := k.SetVestingSchedule(ctx, schedule); err != nil {
		return sdk.Coin{}, err
	}

	return sdk.NewCoin(schedule.OriginalAmount.Denom, toRelease), nil
}

// RevokeVestingSchedule revokes a vesting schedule (wrapper for msg_server)
func (k Keeper) RevokeVestingSchedule(ctx context.Context, revoker sdk.AccAddress, scheduleID string, reason string) (sdk.Coin, error) {
	// Get the schedule
	schedule, err := k.GetVestingSchedule(ctx, scheduleID)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrScheduleNotFound, "schedule not found")
	}

	// Check if already revoked
	if schedule.Revoked {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrScheduleRevoked, "schedule already revoked")
	}

	// Calculate unvested amount
	vestedAmount := schedule.VestedAmount.Amount
	totalAmount := schedule.OriginalAmount.Amount
	unvestedAmount := totalAmount.Sub(vestedAmount)

	// Mark as revoked
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	schedule.Revoked = true
	revokedTime := sdkCtx.BlockTime()
	schedule.RevokedAt = &revokedTime
	schedule.RevokedReason = reason

	// Update schedule
	if err := k.SetVestingSchedule(ctx, schedule); err != nil {
		return sdk.Coin{}, err
	}

	return sdk.NewCoin(schedule.OriginalAmount.Denom, unvestedAmount), nil
}

// ============================
// GOVERNANCE OPERATIONS
// ============================

// SubmitProposal submits a new governance proposal
func (k Keeper) SubmitProposal(
	ctx context.Context,
	title string,
	description string,
	category economicspb.ProposalCategory,
	proposer sdk.AccAddress,
	initialDeposit sdk.Coins,
	isEmergency bool,
) (uint64, error) {
	// Validate title not empty
	if title == "" {
		return 0, errorsmod.Wrap(types.ErrInvalidRequest, "invalid title")
	}

	// Validate description not empty
	if description == "" {
		return 0, errorsmod.Wrap(types.ErrInvalidRequest, "invalid description")
	}

	// Get next proposal ID
	proposalID, err := k.GetNextProposalID(ctx)
	if err != nil {
		return 0, err
	}

	// Get params
	params, err := k.GetParams(ctx)
	if err != nil {
		return 0, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()

	// Calculate timeframes
	var votingPeriod time.Duration
	if isEmergency {
		votingPeriod = params.Governance.EmergencyVotingPeriod
	} else {
		votingPeriod = params.Governance.VotingPeriod
	}

	depositPeriod := params.Governance.MaxDepositPeriod

	// Create proposal
	proposal := &economicspb.Proposal{
		Id:              proposalID,
		Title:           title,
		Description:     description,
		Category:        category,
		Proposer:        proposer.String(),
		Status:          economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
		SubmitTime:      now,
		DepositEndTime:  now.Add(depositPeriod),
		TotalDeposit:    initialDeposit,
		IsEmergency:     isEmergency,
		ExecutionDelay:  &params.Governance.ExecutionDelay,
	}

	// Check if initial deposit meets minimum
	if initialDeposit.IsAllGTE(params.Governance.MinDeposit) {
		// Move to voting period
		proposal.Status = economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
		votingStart := now
		votingEnd := now.Add(votingPeriod)
		proposal.VotingStartTime = &votingStart
		proposal.VotingEndTime = &votingEnd
	}

	// Store proposal
	if err := k.SetProposal(ctx, proposal); err != nil {
		return 0, err
	}

	// Store initial deposit if provided
	if !initialDeposit.IsZero() {
		deposit := &economicspb.Deposit{
			ProposalId: proposalID,
			Depositor:  proposer.String(),
			Amount:     initialDeposit,
			Timestamp:  now,
		}
		if err := k.SetDeposit(ctx, deposit); err != nil {
			return 0, err
		}
	}

	// Increment proposal ID
	if err := k.SetNextProposalID(ctx, proposalID+1); err != nil {
		return 0, err
	}

	return proposalID, nil
}

// AddDeposit adds a deposit to a proposal
func (k Keeper) AddDeposit(ctx context.Context, proposalID uint64, depositor sdk.AccAddress, amount sdk.Coins) error {
	// Get proposal
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return errorsmod.Wrap(types.ErrProposalNotFound, "proposal not found")
	}

	// Check if proposal is in deposit period
	if proposal.Status != economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD {
		return errorsmod.Wrap(types.ErrInvalidProposalStatus, "proposal not in deposit period")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()

	// Add deposit
	deposit := &economicspb.Deposit{
		ProposalId: proposalID,
		Depositor:  depositor.String(),
		Amount:     amount,
		Timestamp:  now,
	}

	if err := k.SetDeposit(ctx, deposit); err != nil {
		return err
	}

	// Update total deposit
	totalDepositCoins := proposal.TotalDeposit.Add(amount...)
	proposal.TotalDeposit = totalDepositCoins

	// Check if minimum deposit met
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if totalDepositCoins.IsAllGTE(params.Governance.MinDeposit) {
		// Move to voting period
		votingPeriod := params.Governance.VotingPeriod
		if proposal.IsEmergency {
			votingPeriod = params.Governance.EmergencyVotingPeriod
		}

		proposal.Status = economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
		votingStart := now
		votingEnd := now.Add(votingPeriod)
		proposal.VotingStartTime = &votingStart
		proposal.VotingEndTime = &votingEnd
	}

	return k.SetProposal(ctx, proposal)
}

// AddVote casts a vote on a proposal
func (k Keeper) AddVote(ctx context.Context, proposalID uint64, voter sdk.AccAddress, option economicspb.VoteOption, isSecret bool, voteCommitment string) error {
	// Validate vote option is not unspecified
	if option == economicspb.VoteOption_VOTE_OPTION_UNSPECIFIED {
		return errorsmod.Wrap(types.ErrInvalidVote, "invalid vote option")
	}

	// Get proposal
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return errorsmod.Wrap(types.ErrProposalNotFound, "proposal not found")
	}

	// Check if proposal is in voting period
	if proposal.Status != economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
		return errorsmod.Wrap(types.ErrInvalidProposalStatus, "proposal not in voting period")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Create vote
	votingPower := sdkmath.NewInt(1000000) // TODO: Calculate actual voting power
	vote := &economicspb.Vote{
		ProposalId:     proposalID,
		Voter:          voter.String(),
		Option:         option,
		Timestamp:      sdkCtx.BlockTime(),
		VotingPower:    votingPower,
		IsSecret:       isSecret,
		VoteCommitment: voteCommitment,
	}

	return k.SetVote(ctx, vote)
}

// AddWeightedVote casts a weighted vote on a proposal
func (k Keeper) AddWeightedVote(ctx context.Context, proposalID uint64, voter sdk.AccAddress, options []*economicspb.WeightedVoteOption) error {
	// Validate options not empty
	if len(options) == 0 {
		return errorsmod.Wrap(types.ErrInvalidVote, "invalid vote options")
	}

	// Get proposal
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return errorsmod.Wrap(types.ErrProposalNotFound, "proposal not found")
	}

	// Check if proposal is in voting period
	if proposal.Status != economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
		return errorsmod.Wrap(types.ErrInvalidProposalStatus, "proposal not in voting period")
	}

	// Validate weights sum to 1
	totalWeight := sdkmath.LegacyZeroDec()
	for _, opt := range options {
		totalWeight = totalWeight.Add(opt.Weight)
	}
	if !totalWeight.Equal(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(types.ErrInvalidVote, "weights must sum to 1.0")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// For weighted votes, we store one vote record with the first option
	// In a real implementation, you'd want to store all weighted options
	votingPower := sdkmath.NewInt(1000000) // TODO: Calculate actual voting power
	vote := &economicspb.Vote{
		ProposalId:  proposalID,
		Voter:       voter.String(),
		Option:      options[0].Option, // Store first option as primary
		Timestamp:   sdkCtx.BlockTime(),
		VotingPower: votingPower,
	}

	return k.SetVote(ctx, vote)
}

// DelegateVote delegates voting power to another address
func (k Keeper) DelegateVote(ctx context.Context, delegator sdk.AccAddress, delegate sdk.AccAddress, categories []economicspb.ProposalCategory) error {
	// Prevent self-delegation
	if delegator.String() == delegate.String() {
		return errorsmod.Wrap(types.ErrInvalidRequest, "cannot delegate to self")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Create delegation
	delegatedPower := sdkmath.NewInt(1000000) // TODO: Calculate actual power
	delegation := &economicspb.VoteDelegation{
		Delegator:      delegator.String(),
		Delegate:       delegate.String(),
		DelegationTime: sdkCtx.BlockTime(),
		DelegatedPower: delegatedPower,
		Categories:     categories,
	}

	return k.SetVoteDelegation(ctx, delegation)
}

// UndelegateVote removes vote delegation
func (k Keeper) UndelegateVote(ctx context.Context, delegator sdk.AccAddress, delegate sdk.AccAddress, categories []economicspb.ProposalCategory) error {
	// Check if delegation exists
	_, err := k.GetVoteDelegation(ctx, delegator.String(), delegate.String())
	if err != nil {
		return errorsmod.Wrap(types.ErrDelegationNotFound, "delegation not found")
	}

	// Delete delegation
	return k.DeleteVoteDelegation(ctx, delegator.String(), delegate.String())
}

// ExecuteProposal executes a passed proposal
func (k Keeper) ExecuteProposal(ctx context.Context, proposalID uint64, executor sdk.AccAddress) error {
	// Get proposal
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return errorsmod.Wrap(types.ErrProposalNotFound, "proposal not found")
	}

	// Check if proposal passed
	if proposal.Status != economicspb.ProposalStatus_PROPOSAL_STATUS_PASSED {
		return errorsmod.Wrap(types.ErrInvalidProposalStatus, "proposal not in passed status")
	}

	// Check if execution delay has passed
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if proposal.VotingEndTime != nil && proposal.ExecutionDelay != nil {
		executionTime := proposal.VotingEndTime.Add(*proposal.ExecutionDelay)
		if sdkCtx.BlockTime().Before(executionTime) {
			return errorsmod.Wrap(types.ErrExecutionDelayNotMet, "execution delay not met")
		}
	}

	// Mark as executed
	proposal.Status = economicspb.ProposalStatus_PROPOSAL_STATUS_EXECUTED
	executionTime := sdkCtx.BlockTime()
	proposal.ExecutionTime = &executionTime

	return k.SetProposal(ctx, proposal)
}

// RevealSecretVote reveals a secret ballot vote
func (k Keeper) RevealSecretVote(ctx context.Context, proposalID uint64, voter sdk.AccAddress, option economicspb.VoteOption, revealKey string) error {
	// Validate reveal key not empty
	if revealKey == "" {
		return errorsmod.Wrap(types.ErrInvalidRequest, "invalid reveal key")
	}

	// Get existing vote
	vote, err := k.GetVote(ctx, proposalID, voter.String())
	if err != nil {
		return errorsmod.Wrap(types.ErrVoteNotFound, "vote not found")
	}

	// Verify it's a secret vote
	if !vote.IsSecret {
		return errorsmod.Wrap(types.ErrInvalidVote, "not a secret vote")
	}

	// TODO: Verify reveal key matches commitment
	// For now, just update the vote with the revealed option
	vote.Option = option
	vote.EncryptedVote = "" // Clear encrypted data after reveal

	return k.SetVote(ctx, vote)
}

// ============================
// VOTE LOCK OPERATIONS
// ============================

// LockVotingTokens locks tokens for voting power boost
func (k Keeper) LockVotingTokens(ctx context.Context, owner sdk.AccAddress, amount sdk.Coin, lockDuration uint64) (string, sdkmath.Int, error) {
	// Validate amount > 0
	if amount.Amount.IsZero() || amount.Amount.IsNegative() {
		return "", sdkmath.ZeroInt(), errorsmod.Wrap(types.ErrInvalidAmount, "lock amount must be greater than zero")
	}

	// Validate lock duration > 0
	if lockDuration == 0 {
		return "", sdkmath.ZeroInt(), errorsmod.Wrap(types.ErrInvalidRequest, "invalid lock duration")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()

	// Generate lock ID
	lockID := fmt.Sprintf("lock-%s-%d", owner.String(), now.Unix())

	// Calculate voting power (amount * time multiplier)
	votingPowerStr := k.CalculateVotingPowerFromDuration(amount.Amount.String(), int64(lockDuration))
	votingPower, ok := sdkmath.NewIntFromString(votingPowerStr)
	if !ok {
		return "", sdkmath.ZeroInt(), errorsmod.Wrap(types.ErrInvalidAmount, "failed to calculate voting power")
	}

	// Create vote lock
	lock := &economicspb.VoteLock{
		Id:          lockID,
		Owner:       owner.String(),
		Amount:      amount,
		LockStart:   now,
		LockEnd:     now.Add(time.Duration(lockDuration) * time.Second),
		VotingPower: votingPower,
		Withdrawn:   false,
	}

	// Store lock
	if err := k.SetVoteLock(ctx, lock); err != nil {
		return "", sdkmath.ZeroInt(), err
	}

	// Update user index
	if err := k.AddUserVoteLock(ctx, owner.String(), lockID); err != nil {
		return "", sdkmath.ZeroInt(), err
	}

	return lockID, votingPower, nil
}

// UnlockVotingTokens unlocks voting tokens after lock period
func (k Keeper) UnlockVotingTokens(ctx context.Context, owner sdk.AccAddress, lockID string) (sdk.Coin, error) {
	// Get lock
	lock, err := k.GetVoteLock(ctx, lockID)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrLockNotFound, "lock not found")
	}

	// Verify owner
	if lock.Owner != owner.String() {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrUnauthorized, "not the owner of this lock")
	}

	// Check if already withdrawn
	if lock.Withdrawn {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrLockWithdrawn, "tokens already withdrawn")
	}

	// Check if lock period ended
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockTime().Before(lock.LockEnd) {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrLockNotEnded, "lock period not ended")
	}

	// Mark as withdrawn
	lock.Withdrawn = true
	if err := k.SetVoteLock(ctx, lock); err != nil {
		return sdk.Coin{}, err
	}

	return lock.Amount, nil
}

// ============================
// TREASURY OPERATIONS
// ============================

// ProposeTreasurySpend proposes a treasury spend
func (k Keeper) ProposeTreasurySpend(ctx context.Context, proposer sdk.AccAddress, recipient sdk.AccAddress, amount sdk.Coins, description string) (string, time.Time, error) {
	// Validate amount > 0
	if amount.IsZero() || !amount.IsValid() {
		return "", time.Time{}, errorsmod.Wrap(types.ErrInvalidAmount, "treasury spend amount must be greater than zero")
	}

	// Validate description not empty
	if description == "" {
		return "", time.Time{}, errorsmod.Wrap(types.ErrInvalidRequest, "invalid description")
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return "", time.Time{}, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()

	// Generate tx ID
	txID := fmt.Sprintf("treasury-tx-%d", now.Unix())

	// Calculate executable time (after timelock)
	executableAt := now.Add(params.Treasury.TimelockDuration)

	// Create pending transaction
	tx := &economicspb.PendingTreasuryTx{
		TxId:         txID,
		Recipient:    recipient.String(),
		Amount:       amount,
		Description:  description,
		Proposer:     proposer.String(),
		Signatures:   []string{},
		CreatedAt:    now,
		ExecutableAt: executableAt,
		Executed:     false,
		Rejected:     false,
	}

	if err := k.SetPendingTreasuryTx(ctx, tx); err != nil {
		return "", time.Time{}, err
	}

	return txID, executableAt, nil
}

// SignTreasurySpend signs a treasury spend proposal
func (k Keeper) SignTreasurySpend(ctx context.Context, signer sdk.AccAddress, txID string) (uint32, uint32, error) {
	// Get transaction
	tx, err := k.GetPendingTreasuryTx(ctx, txID)
	if err != nil {
		return 0, 0, errorsmod.Wrap(types.ErrTreasuryTxNotFound, "transaction not found")
	}

	// Check if already executed or rejected
	if tx.Executed || tx.Rejected {
		return 0, 0, errorsmod.Wrap(types.ErrTreasuryTxExecuted, "transaction already executed or rejected")
	}

	// Check if signer already signed
	signerStr := signer.String()
	for _, sig := range tx.Signatures {
		if sig == signerStr {
			return uint32(len(tx.Signatures)), 0, nil // Already signed
		}
	}

	// Add signature
	tx.Signatures = append(tx.Signatures, signerStr)

	// Update transaction
	if err := k.SetPendingTreasuryTx(ctx, tx); err != nil {
		return 0, 0, err
	}

	// Get params to check threshold
	params, err := k.GetParams(ctx)
	if err != nil {
		return 0, 0, err
	}

	return uint32(len(tx.Signatures)), params.Treasury.MultisigThreshold, nil
}

// ExecuteTreasurySpend executes an approved treasury spend
func (k Keeper) ExecuteTreasurySpend(ctx context.Context, executor sdk.AccAddress, txID string) (bool, error) {
	// Get transaction
	tx, err := k.GetPendingTreasuryTx(ctx, txID)
	if err != nil {
		return false, errorsmod.Wrap(types.ErrTreasuryTxNotFound, "transaction not found")
	}

	// Check if already executed
	if tx.Executed {
		return false, errorsmod.Wrap(types.ErrTreasuryTxExecuted, "transaction already executed")
	}

	// Check if rejected
	if tx.Rejected {
		return false, errorsmod.Wrap(types.ErrTreasuryTxRejected, "transaction rejected")
	}

	// Get params
	params, err := k.GetParams(ctx)
	if err != nil {
		return false, err
	}

	// Check if enough signatures
	if uint32(len(tx.Signatures)) < params.Treasury.MultisigThreshold {
		return false, errorsmod.Wrap(types.ErrInsufficientSignatures, "insufficient signatures")
	}

	// Check if timelock passed
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockTime().Before(tx.ExecutableAt) {
		return false, errorsmod.Wrap(types.ErrTimelockNotMet, "timelock not met")
	}

	// Mark as executed
	tx.Executed = true

	// Update transaction
	if err := k.SetPendingTreasuryTx(ctx, tx); err != nil {
		return false, err
	}

	// TODO: Actually transfer funds from treasury to recipient
	// This would involve bank keeper operations

	return true, nil
}

// ============================
// ADMIN OPERATIONS
// ============================

// UpdateParams updates module parameters (wrapper for msg_server)
func (k Keeper) UpdateParams(ctx context.Context, authority sdk.AccAddress, params *economicspb.Params) error {
	// Verify authority
	if authority.String() != k.GetAuthority() {
		return errorsmod.Wrapf(types.ErrUnauthorized, "unauthorized: expected %s, got %s", k.GetAuthority(), authority.String())
	}

	// Update params
	return k.SetParams(ctx, params)
}

// AdjustInflationRate manually adjusts inflation rate (wrapper for msg_server)
func (k Keeper) AdjustInflationRate(ctx context.Context, authority sdk.AccAddress, newRate uint64, reason string) (uint64, uint64, error) {
	// Verify authority
	if authority.String() != k.GetAuthority() {
		return 0, 0, errorsmod.Wrapf(types.ErrUnauthorized, "unauthorized: expected %s, got %s", k.GetAuthority(), authority.String())
	}

	// Validate reason
	if reason == "" {
		return 0, 0, errorsmod.Wrap(types.ErrInvalidRequest, "invalid reason: reason cannot be empty")
	}

	// Validate inflation rate (basis points: max 10000 = 100%)
	// Allowing up to 10000 basis points (100%) as maximum reasonable inflation rate
	if newRate > 10000 {
		return 0, 0, errorsmod.Wrapf(types.ErrInvalidInflationRate, "invalid inflation rate: %d basis points exceeds maximum of 10000 (100%%)", newRate)
	}

	// Get current rate
	oldRate, err := k.GetPreviousInflation(ctx)
	if err != nil {
		oldRate = 0 // Default to 0 if not set
	}

	// Store new rate
	if err := k.SetPreviousInflation(ctx, newRate); err != nil {
		return 0, 0, err
	}

	// TODO: Update inflation metrics with reason
	// For now, just return the old and new rates

	return oldRate, newRate, nil
}
