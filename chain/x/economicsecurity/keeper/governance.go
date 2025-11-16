package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// GOVERNANCE (Features 7, 8, 9)
// ============================

// CheckProposalStake validates if an address has sufficient stake to create a proposal (Feature 7)
func (k *Keeper) CheckProposalStake(proposer, stakeAmount string) error {
	params := k.GetParams()

	minStake := new(big.Int)
	minStake.SetString(params.Governance.MinProposalStake, 10)

	stake := new(big.Int)
	if _, ok := stake.SetString(stakeAmount, 10); !ok {
		return types.ErrInvalidAmount
	}

	if stake.Cmp(minStake) < 0 {
		return types.ErrInsufficientStake
	}

	return nil
}

// CalculateQuadraticVotingPower calculates voting power using quadratic voting (Feature 8)
func (k *Keeper) CalculateQuadraticVotingPower(stakeAmount string) (string, error) {
	params := k.GetParams()

	if !params.Governance.QuadraticVotingEnabled {
		// If quadratic voting disabled, return linear voting power
		return stakeAmount, nil
	}

	stake := new(big.Int)
	if _, ok := stake.SetString(stakeAmount, 10); !ok {
		return "0", types.ErrInvalidAmount
	}

	// Quadratic voting: voting power = sqrt(stake)
	// We use a simplified integer square root
	votingPower := sqrt(stake)

	return votingPower.String(), nil
}

// LockVotingTokens locks tokens for voting power boost (Feature 9)
func (k *Keeper) LockVotingTokens(owner, amount string, lockDuration uint64) (string, string, error) {
	params := k.GetParams()

	if !params.Governance.VoteLockingEnabled {
		return "", "0", errors.New("vote locking is disabled")
	}

	if lockDuration < params.Governance.MinLockDuration {
		return "", "0", types.ErrLockDurationTooShort
	}

	if lockDuration > params.Governance.MaxLockDuration {
		return "", "0", types.ErrLockDurationTooLong
	}

	lockAmt := new(big.Int)
	if _, ok := lockAmt.SetString(amount, 10); !ok || lockAmt.Cmp(big.NewInt(0)) <= 0 {
		return "", "0", types.ErrInvalidAmount
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Generate lock ID
	lockID := k.generateLockID(owner, amount, lockDuration)

	// Calculate voting power multiplier based on lock duration
	// multiplier = 1 + (duration / 1 year) * multiplier_per_year
	oneYear := uint64(31536000) // seconds in a year
	yearsLocked := float64(lockDuration) / float64(oneYear)
	multiplierBasisPoints := params.Governance.LockMultiplierPerYear

	totalMultiplier := 1.0 + (yearsLocked * float64(multiplierBasisPoints) / float64(types.BasisPoints))

	votingPower := new(big.Float).SetInt(lockAmt)
	votingPower.Mul(votingPower, big.NewFloat(totalMultiplier))

	votingPowerInt, _ := votingPower.Int(nil)

	lock := &types.VoteLock{
		LockId:      lockID,
		Owner:       owner,
		Amount:      amount,
		LockStart:   timestamppb.New(time.Unix(k.currentTime, 0)),
		LockEnd:     timestamppb.New(time.Unix(k.currentTime+int64(lockDuration), 0)),
		VotingPower: votingPowerInt.String(),
		Withdrawn:   false,
	}

	k.voteLocks[lockID] = lock
	k.userVoteLocks[owner] = append(k.userVoteLocks[owner], lockID)

	return lockID, votingPowerInt.String(), nil
}

// UnlockVotingTokens unlocks voting tokens after lock period
func (k *Keeper) UnlockVotingTokens(owner, lockID string) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	lock, ok := k.voteLocks[lockID]
	if !ok {
		return "0", types.ErrVoteLockNotFound
	}

	if lock.Owner != owner {
		return "0", types.ErrUnauthorized
	}

	if lock.Withdrawn {
		return "0", types.ErrVoteLockAlreadyWithdrawn
	}

	lockEnd := lock.LockEnd.AsTime().Unix()
	if k.currentTime < lockEnd {
		return "0", types.ErrVoteLockNotExpired
	}

	// Mark as withdrawn
	lock.Withdrawn = true
	k.voteLocks[lockID] = lock

	return lock.Amount, nil
}

// GetVotingPower calculates total voting power for an address
func (k *Keeper) GetVotingPower(address string) (string, string, uint64) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	lockIDs := k.userVoteLocks[address]
	totalVotingPower := big.NewInt(0)
	totalLocked := big.NewInt(0)
	activeLocks := uint64(0)

	for _, lockID := range lockIDs {
		lock, ok := k.voteLocks[lockID]
		if !ok || lock.Withdrawn {
			continue
		}

		// Check if lock is still active
		lockEnd := lock.LockEnd.AsTime().Unix()
		if k.currentTime >= lockEnd {
			continue
		}

		activeLocks++

		votingPower := new(big.Int)
		votingPower.SetString(lock.VotingPower, 10)
		totalVotingPower.Add(totalVotingPower, votingPower)

		locked := new(big.Int)
		locked.SetString(lock.Amount, 10)
		totalLocked.Add(totalLocked, locked)
	}

	return totalVotingPower.String(), totalLocked.String(), activeLocks
}

// GetVoteLock retrieves a vote lock by ID
func (k *Keeper) GetVoteLock(lockID string) (*types.VoteLock, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	lock, ok := k.voteLocks[lockID]
	return lock, ok
}

// GetUserVoteLocks returns all vote locks for a user
func (k *Keeper) GetUserVoteLocks(owner string) []*types.VoteLock {
	k.mu.RLock()
	defer k.mu.RUnlock()

	lockIDs := k.userVoteLocks[owner]
	locks := make([]*types.VoteLock, 0, len(lockIDs))

	for _, lockID := range lockIDs {
		if lock, ok := k.voteLocks[lockID]; ok {
			locks = append(locks, lock)
		}
	}

	return locks
}

func (k *Keeper) generateLockID(owner, amount string, duration uint64) string {
	h := sha256.New()
	h.Write([]byte(owner))
	h.Write([]byte(amount))
	h.Write([]byte(fmt.Sprintf("%d", duration)))
	h.Write([]byte(fmt.Sprintf("%d", k.currentTime)))
	return "vl:" + hex.EncodeToString(h.Sum(nil))[:32]
}

// sqrt calculates integer square root using Newton's method
func sqrt(n *big.Int) *big.Int {
	if n.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}

	// Convert to float for sqrt calculation
	nFloat := new(big.Float).SetInt(n)
	sqrtFloat := new(big.Float).Sqrt(nFloat)

	// Convert back to int
	result, _ := sqrtFloat.Int(nil)
	return result
}

// GetTotalLockedGovernance returns total amount locked for governance
func (k *Keeper) GetTotalLockedGovernance() string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	total := big.NewInt(0)

	for _, lock := range k.voteLocks {
		if lock.Withdrawn {
			continue
		}

		lockEnd := lock.LockEnd.AsTime().Unix()
		if k.currentTime >= lockEnd {
			continue
		}

		amount := new(big.Int)
		amount.SetString(lock.Amount, 10)
		total.Add(total, amount)
	}

	return total.String()
}
