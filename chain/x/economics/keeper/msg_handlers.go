package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
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
		Id:             scheduleID,
		Address:        beneficiaryAddress,
		OriginalAmount: totalAmount,
		VestedAmount:   vestedCoin,
		StartTime:      startTime,
		EndTime:        endTime,
		CliffDuration:  cliffDuration,
		VestingType:    vestingType,
		ScheduleType:   scheduleType,
		Revoked:        false,
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
	// Validate schedule ID not empty
	if scheduleID == "" {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrInvalidScheduleID, "invalid schedule ID")
	}

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
		Id:             proposalID,
		Title:          title,
		Description:    description,
		Category:       category,
		Proposer:       proposer.String(),
		Status:         economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
		SubmitTime:     now,
		DepositEndTime: now.Add(depositPeriod),
		TotalDeposit:   initialDeposit,
		IsEmergency:    isEmergency,
		ExecutionDelay: &params.Governance.ExecutionDelay,
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
		return fmt.Errorf("error in AddDeposit for ProposalId: %w", err)
	}

	// Update total deposit
	totalDepositCoins := proposal.TotalDeposit.Add(amount...)
	proposal.TotalDeposit = totalDepositCoins

	// Check if minimum deposit met
	params, err := k.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
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

	// Calculate actual voting power based on locked tokens and delegations
	votingPower, _, _, _, err := k.CalculateVotingPower(ctx, voter, proposalID)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidRequest, "failed to calculate voting power")
	}

	// Require minimum voting power
	if votingPower.IsZero() {
		return errorsmod.Wrap(types.ErrInsufficientVotingPower, "no voting power available")
	}

	// Create vote
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

	// Calculate actual voting power based on locked tokens and delegations
	votingPower, _, _, _, err := k.CalculateVotingPower(ctx, voter, proposalID)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidRequest, "failed to calculate voting power")
	}

	// Require minimum voting power
	if votingPower.IsZero() {
		return errorsmod.Wrap(types.ErrInsufficientVotingPower, "no voting power available")
	}

	// For weighted votes, we store one vote record with the first option
	// In a real implementation, you'd want to store all weighted options
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

	// Calculate actual delegated power based on delegator's locked tokens
	// Note: We use proposalID = 0 as this is a standing delegation, not vote-specific
	delegatedPower, _, _, _, err := k.CalculateVotingPower(ctx, delegator, 0)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidRequest, "failed to calculate delegated power")
	}

	// Require minimum voting power to delegate
	if delegatedPower.IsZero() {
		return errorsmod.Wrap(types.ErrInsufficientVotingPower, "no voting power available to delegate")
	}

	// Create delegation
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

	// Verify the commitment hasn't been revealed yet
	if vote.VoteCommitment == "" {
		return errorsmod.Wrap(types.ErrVoteAlreadyRevealed, "vote has already been revealed")
	}

	// Verify reveal key matches the original commitment
	// The commitment should be a hash of (option || revealKey)
	// Using SHA256 as the hash function
	commitmentData := fmt.Sprintf("%d:%s", option, revealKey)
	computedCommitment := fmt.Sprintf("%x", sha256.Sum256([]byte(commitmentData)))

	if computedCommitment != vote.VoteCommitment {
		return errorsmod.Wrap(types.ErrInvalidCommitment, "reveal key does not match vote commitment")
	}

	// Update the vote with the revealed option
	vote.Option = option
	vote.EncryptedVote = ""  // Clear encrypted data after reveal
	vote.VoteCommitment = "" // Clear commitment to mark as revealed

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

	// Execute fund transfer from treasury module account to recipient
	// NOTE: This requires the economics keeper to have a BankKeeper reference.
	// For now, we emit an event that the treasury module can listen to and execute.
	// In production, add bankKeeper to the Keeper struct and inject it during initialization.
	//
	// Production implementation would be:
	// treasuryAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	// recipientAddr, err := sdk.AccAddressFromBech32(tx.Recipient)
	// if err != nil {
	//     return false, errorsmod.Wrap(types.ErrInvalidAddress, "invalid recipient address")
	// }
	// if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, tx.Amount); err != nil {
	//     return false, errorsmod.Wrap(types.ErrInsufficientTreasuryBalance, err.Error())
	// }

	// Emit event for treasury spend execution
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_spend_executed",
			sdk.NewAttribute("tx_id", tx.TxId),
			sdk.NewAttribute("recipient", tx.Recipient),
			sdk.NewAttribute("amount", tx.Amount.String()),
			sdk.NewAttribute("proposer", tx.Proposer),
			sdk.NewAttribute("signatures", fmt.Sprintf("%d", len(tx.Signatures))),
		),
	)

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

	// Update inflation metrics
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime()

	// Get or create inflation metrics
	metrics, err := k.GetInflationMetrics(ctx)
	if err != nil {
		// Create new metrics if not found
		metrics = &economicspb.InflationMetrics{
			CurrentRate:       newRate,
			CirculatingSupply: sdkmath.ZeroInt(),
			TotalVested:       sdkmath.ZeroInt(),
			TotalVesting:      sdkmath.ZeroInt(),
			LastAdjustment:    currentTime,
			NextCheck:         currentTime,
		}
	} else {
		// Update existing metrics
		metrics.CurrentRate = newRate
		metrics.LastAdjustment = currentTime
	}

	// Store updated metrics
	if err := k.SetInflationMetrics(ctx, metrics); err != nil {
		return 0, 0, err
	}

	// Emit event for inflation adjustment
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"inflation_rate_adjusted",
			sdk.NewAttribute("old_rate_bps", fmt.Sprintf("%d", oldRate)),
			sdk.NewAttribute("new_rate_bps", fmt.Sprintf("%d", newRate)),
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("authority", authority.String()),
			sdk.NewAttribute("timestamp", currentTime.Format(time.RFC3339)),
		),
	)

	return oldRate, newRate, nil
}
