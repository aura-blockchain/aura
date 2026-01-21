package main

import (
	"errors"
	"math"
	"sync"
	"time"
)

// RewardsManager manages verifier reward distribution
type RewardsManager struct {
	mu               sync.RWMutex
	rewardPool       int64
	distributedToday int64
	baseReward       int64
	distributions    map[string]*RewardDistribution
	config           RewardConfig
}

// RewardConfig configures the reward system
type RewardConfig struct {
	BaseReward        int64
	QualityMultiplier float64
	SpeedMultiplier   float64
	TierMultipliers   map[VerifierTier]float64
	MaxDailyReward    int64
	MinimumPayout     int64
}

// RewardDistribution tracks a reward distribution
type RewardDistribution struct {
	ID         string
	VerifierID string
	TaskID     string
	Amount     int64
	BaseAmount int64
	Bonus      int64
	Reason     string
	CreatedAt  time.Time
	Status     string
	TxHash     string
}

// RewardCalculation contains reward calculation details
type RewardCalculation struct {
	BaseReward   int64
	QualityBonus int64
	SpeedBonus   int64
	TierBonus    int64
	TotalReward  int64
	Multiplier   float64
}

// NewRewardsManager creates a new rewards manager
func NewRewardsManager(config RewardConfig) *RewardsManager {
	if config.BaseReward == 0 {
		config.BaseReward = 100
	}
	if config.QualityMultiplier == 0 {
		config.QualityMultiplier = 1.5
	}
	if config.SpeedMultiplier == 0 {
		config.SpeedMultiplier = 1.2
	}
	if config.TierMultipliers == nil {
		config.TierMultipliers = map[VerifierTier]float64{
			TierBronze:   1.0,
			TierSilver:   1.2,
			TierGold:     1.5,
			TierPlatinum: 2.0,
		}
	}
	if config.MaxDailyReward == 0 {
		config.MaxDailyReward = 10000
	}
	if config.MinimumPayout == 0 {
		config.MinimumPayout = 50
	}

	return &RewardsManager{
		rewardPool:    1000000, // Initial pool
		distributions: make(map[string]*RewardDistribution),
		config:        config,
	}
}

// CalculateReward calculates reward for a completed task
func (rm *RewardsManager) CalculateReward(
	verifier *Verifier,
	task *Task,
	quality float64,
	completionTime time.Duration,
) *RewardCalculation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	calc := &RewardCalculation{
		BaseReward: rm.config.BaseReward,
	}

	// Quality bonus (0-100 quality score)
	if quality >= 90 {
		calc.QualityBonus = int64(float64(rm.config.BaseReward) * rm.config.QualityMultiplier)
	} else if quality >= 80 {
		calc.QualityBonus = int64(float64(rm.config.BaseReward) * 0.5)
	}

	// Speed bonus (if completed faster than average)
	expectedTime := 24 * time.Hour
	if completionTime < expectedTime/2 {
		calc.SpeedBonus = int64(float64(rm.config.BaseReward) * rm.config.SpeedMultiplier)
	}

	// Tier bonus
	tierMultiplier := rm.config.TierMultipliers[verifier.Tier]
	calc.TierBonus = int64(float64(rm.config.BaseReward) * (tierMultiplier - 1.0))

	// Calculate total
	calc.TotalReward = calc.BaseReward + calc.QualityBonus + calc.SpeedBonus + calc.TierBonus
	calc.Multiplier = float64(calc.TotalReward) / float64(calc.BaseReward)

	return calc
}

// DistributeReward distributes reward to a verifier
func (rm *RewardsManager) DistributeReward(
	verifierID string,
	taskID string,
	amount int64,
	reason string,
) (*RewardDistribution, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Check minimum payout
	if amount < rm.config.MinimumPayout {
		return nil, errors.New("amount below minimum payout threshold")
	}

	// Check pool availability
	if rm.rewardPool < amount {
		return nil, errors.New("insufficient reward pool")
	}

	// Check daily limit
	if rm.distributedToday+amount > rm.config.MaxDailyReward {
		return nil, errors.New("daily reward limit exceeded")
	}

	// Create distribution
	dist := &RewardDistribution{
		ID:         generateID("reward"),
		VerifierID: verifierID,
		TaskID:     taskID,
		Amount:     amount,
		BaseAmount: rm.config.BaseReward,
		Bonus:      amount - rm.config.BaseReward,
		Reason:     reason,
		CreatedAt:  time.Now(),
		Status:     "pending",
	}

	rm.distributions[dist.ID] = dist

	// Update pool and daily total
	rm.rewardPool -= amount
	rm.distributedToday += amount

	return dist, nil
}

// ProcessReward processes a reward distribution (simulate blockchain tx)
func (rm *RewardsManager) ProcessReward(distributionID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	dist, exists := rm.distributions[distributionID]
	if !exists {
		return errors.New("distribution not found")
	}

	if dist.Status != "pending" {
		return errors.New("distribution already processed")
	}

	// Simulate blockchain transaction
	// In production, this would interact with the blockchain
	dist.TxHash = generateTxHash()
	dist.Status = "completed"

	return nil
}

// GetDistribution retrieves a reward distribution
func (rm *RewardsManager) GetDistribution(distributionID string) (*RewardDistribution, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	dist, exists := rm.distributions[distributionID]
	if !exists {
		return nil, errors.New("distribution not found")
	}

	return dist, nil
}

// GetVerifierRewards gets all rewards for a verifier
func (rm *RewardsManager) GetVerifierRewards(verifierID string) []*RewardDistribution {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rewards := make([]*RewardDistribution, 0)
	for _, dist := range rm.distributions {
		if dist.VerifierID == verifierID {
			rewards = append(rewards, dist)
		}
	}

	return rewards
}

// GetVerifierTotalEarned calculates total earned by a verifier
func (rm *RewardsManager) GetVerifierTotalEarned(verifierID string) int64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	total := int64(0)
	for _, dist := range rm.distributions {
		if dist.VerifierID == verifierID && dist.Status == "completed" {
			total += dist.Amount
		}
	}

	return total
}

// AddToPool adds funds to the reward pool
func (rm *RewardsManager) AddToPool(amount int64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.rewardPool += amount
}

// GetPoolBalance returns the current reward pool balance
func (rm *RewardsManager) GetPoolBalance() int64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.rewardPool
}

// GetStatistics returns reward system statistics
func (rm *RewardsManager) GetStatistics() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	totalDistributed := int64(0)
	pendingDistributions := 0
	completedDistributions := 0

	for _, dist := range rm.distributions {
		totalDistributed += dist.Amount
		if dist.Status == "pending" {
			pendingDistributions++
		} else if dist.Status == "completed" {
			completedDistributions++
		}
	}

	return map[string]interface{}{
		"reward_pool":             rm.rewardPool,
		"total_distributed":       totalDistributed,
		"distributed_today":       rm.distributedToday,
		"pending_distributions":   pendingDistributions,
		"completed_distributions": completedDistributions,
		"total_distributions":     len(rm.distributions),
		"avg_reward":              rm.calculateAverageReward(),
	}
}

// ResetDailyLimit resets the daily distribution limit (call at midnight)
func (rm *RewardsManager) ResetDailyLimit() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.distributedToday = 0
}

// calculateAverageReward calculates the average reward amount
func (rm *RewardsManager) calculateAverageReward() int64 {
	if len(rm.distributions) == 0 {
		return 0
	}

	total := int64(0)
	for _, dist := range rm.distributions {
		total += dist.Amount
	}

	return total / int64(len(rm.distributions))
}

// ReputationTracker tracks and updates verifier reputation
type ReputationTracker struct {
	mu      sync.RWMutex
	scores  map[string]*ReputationScore
	history map[string][]*ReputationEvent
}

// ReputationScore tracks a verifier's reputation
type ReputationScore struct {
	VerifierID       string
	Score            int
	Level            string
	TotalPositive    int
	TotalNegative    int
	ConsistencyScore float64
	AccuracyScore    float64
	LastUpdated      time.Time
}

// ReputationEvent represents a reputation change event
type ReputationEvent struct {
	Timestamp time.Time
	Delta     int
	Reason    string
	TaskID    string
}

// NewReputationTracker creates a new reputation tracker
func NewReputationTracker() *ReputationTracker {
	return &ReputationTracker{
		scores:  make(map[string]*ReputationScore),
		history: make(map[string][]*ReputationEvent),
	}
}

// UpdateReputation updates a verifier's reputation
func (rt *ReputationTracker) UpdateReputation(verifierID string, delta int, reason string, taskID string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	score, exists := rt.scores[verifierID]
	if !exists {
		score = &ReputationScore{
			VerifierID: verifierID,
			Score:      100, // Starting score
		}
		rt.scores[verifierID] = score
	}

	// Update score
	score.Score += delta
	if score.Score < 0 {
		score.Score = 0
	}
	if score.Score > 1000 {
		score.Score = 1000
	}

	// Update counters
	if delta > 0 {
		score.TotalPositive++
	} else if delta < 0 {
		score.TotalNegative++
	}

	// Calculate consistency
	total := score.TotalPositive + score.TotalNegative
	if total > 0 {
		score.ConsistencyScore = float64(score.TotalPositive) / float64(total) * 100
	}

	// Update level
	score.Level = rt.calculateLevel(score.Score)
	score.LastUpdated = time.Now()

	// Record event
	event := &ReputationEvent{
		Timestamp: time.Now(),
		Delta:     delta,
		Reason:    reason,
		TaskID:    taskID,
	}

	if rt.history[verifierID] == nil {
		rt.history[verifierID] = make([]*ReputationEvent, 0)
	}
	rt.history[verifierID] = append(rt.history[verifierID], event)
}

// GetReputation gets a verifier's reputation score
func (rt *ReputationTracker) GetReputation(verifierID string) *ReputationScore {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	return rt.scores[verifierID]
}

// GetHistory gets reputation history for a verifier
func (rt *ReputationTracker) GetHistory(verifierID string) []*ReputationEvent {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	return rt.history[verifierID]
}

func (rt *ReputationTracker) calculateLevel(score int) string {
	if score >= 800 {
		return "Platinum"
	} else if score >= 600 {
		return "Gold"
	} else if score >= 400 {
		return "Silver"
	}
	return "Bronze"
}

func generateTxHash() string {
	return "0x" + generateRandomString(64)
}
