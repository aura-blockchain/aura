package privacy

// This file implements OFF-CHAIN coin mixing and tumbler utilities.
//
// IMPORTANT: All mixing and shuffling functions are OFF-CHAIN utilities for external
// mixing coordinators. They use crypto/rand and MUST NOT be called from consensus code.
//
// Coin mixing (tumbling) breaks transaction linkability by mixing coins from multiple
// users through a coordination process:
// - Users join mixing pools with fixed denominations
// - Off-chain coordinator shuffles participants anonymously
// - Outputs are distributed to unlinkable addresses
// - On-chain transactions appear unrelated
//
// ON-CHAIN VS OFF-CHAIN SEPARATION:
// - OFF-CHAIN: CreatePool(), ExecuteMixing(), shuffleParticipants(), ScheduleTumbling()
//   These functions use crypto/rand for pool IDs, shuffling, and scheduling.
//   They are utilities for building external mixing coordinator services.
//   The actual mixing coordination happens OFF-CHAIN.
//
// - ON-CHAIN: MsgCreateMixingPool, MsgJoinMixingPool handlers
//   Message handlers create pool records with deterministic IDs (creator + block height).
//   No randomness is used during consensus - pools are just participant registries.
//   Actual shuffling and distribution happens OFF-CHAIN.
//
// The blockchain only tracks pool membership. The mixing coordinator (external service)
// performs the actual mixing off-chain, then submits distribution transactions separately.
//
// These utility functions are for building external mixing services, NOT for use in
// consensus-critical message handlers.

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

const (
	// Mixing pool statuses
	PoolStatusPending   = "PENDING"
	PoolStatusActive    = "ACTIVE"
	PoolStatusMixing    = "MIXING"
	PoolStatusCompleted = "COMPLETED"
	PoolStatusCancelled = "CANCELLED"
)

// MixingService implements a coin mixing service (tumbler)
type MixingService struct {
	minParticipants int
	pools           map[string]*MixingPool
	mu              sync.RWMutex
}

// NewMixingService creates a new mixing service
func NewMixingService(minParticipants int) *MixingService {
	if minParticipants < 2 {
		minParticipants = 2
	}

	return &MixingService{
		minParticipants: minParticipants,
		pools:           make(map[string]*MixingPool),
	}
}

// MixingPool represents a mixing pool
type MixingPool struct {
	ID              string
	Denomination    *big.Int
	MinParticipants int
	MaxParticipants int
	MinMixRounds    int
	MaxMixRounds    int
	Deadline        time.Time
	Fee             *big.Int
	Status          string
	Participants    []*MixingParticipant
	CreatedAt       time.Time
}

// MixingParticipant represents a participant in a mixing pool
type MixingParticipant struct {
	ID             string
	Commitment     []byte
	OutputAddress  []byte
	BlindingFactor *big.Int
	JoinedAt       time.Time
}

// CreatePool creates a new mixing pool
func (ms *MixingService) CreatePool(
	denomination *big.Int,
	minParticipants, maxParticipants int,
	minRounds, maxRounds int,
	deadline time.Duration,
	fee *big.Int,
	now time.Time, // Accept time parameter for determinism
) (*MixingPool, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if denomination == nil || denomination.Sign() <= 0 {
		return nil, errors.New("denomination must be positive")
	}
	if minParticipants < 2 {
		return nil, errors.New("minimum participants must be at least 2")
	}
	if maxParticipants < minParticipants {
		return nil, errors.New("max participants must be >= min participants")
	}
	if fee == nil || fee.Sign() < 0 {
		return nil, errors.New("fee must be non-negative")
	}

	// Generate pool ID
	poolID, err := generatePoolID(now)
	if err != nil {
		return nil, err
	}

	pool := &MixingPool{
		ID:              poolID,
		Denomination:    denomination,
		MinParticipants: minParticipants,
		MaxParticipants: maxParticipants,
		MinMixRounds:    minRounds,
		MaxMixRounds:    maxRounds,
		Deadline:        now.Add(deadline),
		Fee:             fee,
		Status:          PoolStatusPending,
		Participants:    make([]*MixingParticipant, 0),
		CreatedAt:       now,
	}

	ms.pools[poolID] = pool
	return pool, nil
}

// JoinPool allows a participant to join a mixing pool
func (ms *MixingService) JoinPool(
	poolID string,
	participantID string,
	commitment []byte,
	outputAddress []byte,
	blindingFactor *big.Int,
	now time.Time, // Accept time parameter for determinism
) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	pool, exists := ms.pools[poolID]
	if !exists {
		return errors.New("pool not found")
	}

	if pool.Status != PoolStatusPending && pool.Status != PoolStatusActive {
		return fmt.Errorf("pool is not accepting participants (status: %s)", pool.Status)
	}

	if len(pool.Participants) >= pool.MaxParticipants {
		return errors.New("pool is full")
	}

	if now.After(pool.Deadline) {
		pool.Status = PoolStatusCancelled
		return errors.New("pool deadline has passed")
	}

	// Check for duplicate participant
	for _, p := range pool.Participants {
		if p.ID == participantID {
			return errors.New("participant already in pool")
		}
	}

	participant := &MixingParticipant{
		ID:             participantID,
		Commitment:     commitment,
		OutputAddress:  outputAddress,
		BlindingFactor: blindingFactor,
		JoinedAt:       now,
	}

	pool.Participants = append(pool.Participants, participant)

	// Update pool status
	if len(pool.Participants) >= pool.MinParticipants {
		pool.Status = PoolStatusActive
	}

	return nil
}

// ExecuteMixing executes the mixing process for a pool
func (ms *MixingService) ExecuteMixing(poolID string, now time.Time) (*MixingResult, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	pool, exists := ms.pools[poolID]
	if !exists {
		return nil, errors.New("pool not found")
	}

	if pool.Status != PoolStatusActive {
		return nil, fmt.Errorf("pool is not ready for mixing (status: %s)", pool.Status)
	}

	if len(pool.Participants) < pool.MinParticipants {
		return nil, fmt.Errorf("not enough participants (%d < %d)", len(pool.Participants), pool.MinParticipants)
	}

	pool.Status = PoolStatusMixing

	// Shuffle participants to anonymize the mapping
	shuffled := make([]*MixingParticipant, len(pool.Participants))
	copy(shuffled, pool.Participants)
	shuffleParticipants(shuffled)

	// Create output mappings
	outputs := make([]*MixingOutput, len(shuffled))
	for i, participant := range shuffled {
		outputs[i] = &MixingOutput{
			OutputAddress: participant.OutputAddress,
			Amount:        new(big.Int).Sub(pool.Denomination, pool.Fee),
			Round:         i % pool.MaxMixRounds,
		}
	}

	pool.Status = PoolStatusCompleted

	return &MixingResult{
		PoolID:            poolID,
		Outputs:           outputs,
		TotalParticipants: len(pool.Participants),
		ExecutedAt:        now,
	}, nil
}

// GetPoolStatus retrieves the status of a mixing pool
func (ms *MixingService) GetPoolStatus(poolID string) (*MixingPool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	pool, exists := ms.pools[poolID]
	if !exists {
		return nil, errors.New("pool not found")
	}

	return pool, nil
}

// MixingOutput represents an output from the mixing process
type MixingOutput struct {
	OutputAddress []byte
	Amount        *big.Int
	Round         int
}

// MixingResult represents the result of a mixing operation
type MixingResult struct {
	PoolID            string
	Outputs           []*MixingOutput
	TotalParticipants int
	ExecutedAt        time.Time
}

// Helper functions

func generatePoolID(now time.Time) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	hasher := sha256.New()
	hasher.Write(randomBytes)
	hasher.Write([]byte(now.String()))
	return fmt.Sprintf("pool_%x", hasher.Sum(nil)[:16]), nil
}

func shuffleParticipants(participants []*MixingParticipant) {
	n := len(participants)
	for i := n - 1; i > 0; i-- {
		// Generate random index
		jBytes := make([]byte, 8)
		if _, err := rand.Read(jBytes); err != nil {
			// If randomness fails, keep participant in place to avoid panics
			continue
		}
		j := int(new(big.Int).SetBytes(jBytes).Int64()) % (i + 1)
		if j < 0 {
			j = -j
		}
		participants[i], participants[j] = participants[j], participants[i]
	}
}

// TumblerService implements a tumbler (coin mixing) service
type TumblerService struct {
	minMixRounds int
	schedules    map[string]*TumblerSchedule
	mu           sync.RWMutex
}

// NewTumblerService creates a new tumbler service
func NewTumblerService(minMixRounds int) *TumblerService {
	if minMixRounds < 1 {
		minMixRounds = 1
	}

	return &TumblerService{
		minMixRounds: minMixRounds,
		schedules:    make(map[string]*TumblerSchedule),
	}
}

// TumblerSchedule represents a tumbling schedule
type TumblerSchedule struct {
	ID           string
	InputAddress string
	OutputAddrs  []string
	TotalAmount  *big.Int
	Splits       []*big.Int
	Delays       []time.Duration
	Status       string
	ScheduledAt  time.Time
	CompletedAt  *time.Time
}

// ScheduleTumbling schedules a tumbling operation
func (ts *TumblerService) ScheduleTumbling(
	inputAddr string,
	outputAddrs []string,
	totalAmount *big.Int,
	splits []*big.Int,
	delays []time.Duration,
	now time.Time, // Accept time parameter for determinism
) (*TumblerSchedule, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if len(outputAddrs) != len(splits) || len(splits) != len(delays) {
		return nil, errors.New("output addresses, splits, and delays must have the same length")
	}

	// Verify splits sum to total amount
	sum := big.NewInt(0)
	for _, split := range splits {
		if split.Sign() <= 0 {
			return nil, errors.New("all splits must be positive")
		}
		sum.Add(sum, split)
	}

	if sum.Cmp(totalAmount) != 0 {
		return nil, errors.New("splits do not sum to total amount")
	}

	// Generate schedule ID
	scheduleID, err := generateScheduleID()
	if err != nil {
		return nil, err
	}

	schedule := &TumblerSchedule{
		ID:           scheduleID,
		InputAddress: inputAddr,
		OutputAddrs:  outputAddrs,
		TotalAmount:  totalAmount,
		Splits:       splits,
		Delays:       delays,
		Status:       "SCHEDULED",
		ScheduledAt:  now,
	}

	ts.schedules[scheduleID] = schedule
	return schedule, nil
}

func generateScheduleID() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	hasher := sha256.New()
	hasher.Write(randomBytes)
	return fmt.Sprintf("schedule_%x", hasher.Sum(nil)[:16]), nil
}
