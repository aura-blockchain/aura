package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

// SecurityScoreWeights defines the weights for different scoring factors
type SecurityScoreWeights struct {
	AuditWeight            float64
	MetricsWeight          float64
	ComplianceWeight       float64
	PolicyWeight           float64
	HistoryWeight          float64
	SourceCodeWeight       float64
	VulnerabilityWeight    float64
	TimeDecayFactor        float64
	AuditAgePenaltyDays    int64
	MaxAuditAgeDays        int64
	MinExecutionsForBonus  uint64
	SuccessRateThreshold   float64
	UniqueUsersThreshold   uint64
}

// DefaultSecurityScoreWeights returns the default scoring weights
func DefaultSecurityScoreWeights() SecurityScoreWeights {
	return SecurityScoreWeights{
		AuditWeight:            25.0, // 25 points max for audit status
		MetricsWeight:          20.0, // 20 points max for execution metrics
		ComplianceWeight:       20.0, // 20 points max for compliance configuration
		PolicyWeight:           15.0, // 15 points max for security policy
		HistoryWeight:          10.0, // 10 points max for historical behavior
		SourceCodeWeight:       5.0,  // 5 points max for source code availability
		VulnerabilityWeight:    5.0,  // 5 points deduction for known issues
		TimeDecayFactor:        0.01, // Daily decay factor for stale audits
		AuditAgePenaltyDays:    90,   // Start penalizing after 90 days
		MaxAuditAgeDays:        365,  // Maximum audit age before minimum score
		MinExecutionsForBonus:  1000, // Minimum executions for reliability bonus
		SuccessRateThreshold:   0.95, // 95% success rate for full metrics score
		UniqueUsersThreshold:   100,  // Unique users threshold for adoption score
	}
}

// SecurityScore represents the computed security score for a contract
type SecurityScore struct {
	ContractAddress    string
	TotalScore         uint64
	AuditScore         uint64
	MetricsScore       uint64
	ComplianceScore    uint64
	PolicyScore        uint64
	HistoryScore       uint64
	SourceCodeScore    uint64
	VulnerabilityScore int64
	RiskLevel          string
	Recommendations    []string
	ComputedAt         time.Time
	ExpiresAt          time.Time
	ScoreBreakdown     map[string]float64
}

// SecurityScoreKey returns the KV store key for security scores
func SecurityScoreKey(contractAddr string) []byte {
	return append([]byte("security_score/"), []byte(contractAddr)...)
}

// ComputeSecurityScore calculates a comprehensive security score for a contract
func (k Keeper) ComputeSecurityScore(ctx sdk.Context, contractAddr string) (SecurityScore, error) {
	weights := DefaultSecurityScoreWeights()

	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return SecurityScore{}, types.ErrContractNotFound
	}

	metrics, _ := k.GetContractMetrics(ctx, contractAddr)
	if metrics == nil {
		metrics = &pb.ContractMetrics{ContractAddress: contractAddr}
	}

	score := SecurityScore{
		ContractAddress: contractAddr,
		ComputedAt:      ctx.BlockTime(),
		ExpiresAt:       ctx.BlockTime().Add(24 * time.Hour), // Score valid for 24 hours
		ScoreBreakdown:  make(map[string]float64),
		Recommendations: make([]string, 0),
	}

	// 1. Audit Score (25 points max)
	auditScore := k.calculateAuditScore(ctx, &info, weights)
	score.AuditScore = uint64(auditScore)
	score.ScoreBreakdown["audit"] = auditScore

	// 2. Metrics Score (20 points max)
	metricsScore := k.calculateMetricsScore(metrics, weights)
	score.MetricsScore = uint64(metricsScore)
	score.ScoreBreakdown["metrics"] = metricsScore

	// 3. Compliance Score (20 points max)
	complianceScore := k.calculateComplianceScore(&info, weights)
	score.ComplianceScore = uint64(complianceScore)
	score.ScoreBreakdown["compliance"] = complianceScore

	// 4. Policy Score (15 points max)
	policyScore := k.calculatePolicyScore(&info, weights)
	score.PolicyScore = uint64(policyScore)
	score.ScoreBreakdown["policy"] = policyScore

	// 5. History Score (10 points max)
	historyScore := k.calculateHistoryScore(ctx, contractAddr, metrics, weights)
	score.HistoryScore = uint64(historyScore)
	score.ScoreBreakdown["history"] = historyScore

	// 6. Source Code Score (5 points max)
	sourceCodeScore := k.calculateSourceCodeScore(&info, weights)
	score.SourceCodeScore = uint64(sourceCodeScore)
	score.ScoreBreakdown["source_code"] = sourceCodeScore

	// 7. Vulnerability Deductions (up to -5 points)
	vulnerabilityPenalty := k.calculateVulnerabilityPenalty(ctx, contractAddr, metrics, weights)
	score.VulnerabilityScore = int64(vulnerabilityPenalty)
	score.ScoreBreakdown["vulnerability_penalty"] = vulnerabilityPenalty

	// Calculate total score (0-100)
	totalScore := auditScore + metricsScore + complianceScore + policyScore +
		historyScore + sourceCodeScore + vulnerabilityPenalty
	if totalScore < 0 {
		totalScore = 0
	}
	if totalScore > 100 {
		totalScore = 100
	}
	score.TotalScore = uint64(totalScore)

	// Determine risk level and recommendations
	score.RiskLevel = k.determineRiskLevel(score.TotalScore)
	score.Recommendations = k.generateRecommendations(ctx, &info, metrics, &score)

	// Store the computed score
	k.setSecurityScore(ctx, contractAddr, &score)

	return score, nil
}

// calculateAuditScore computes the audit component of the security score
func (k Keeper) calculateAuditScore(ctx sdk.Context, info *pb.ContractInfo, weights SecurityScoreWeights) float64 {
	if info.Compliance == nil {
		return 0
	}

	compliance := info.Compliance
	var score float64 = 0

	// Check if audit is configured as required
	if compliance.RequireAudit {
		score += weights.AuditWeight * 0.2 // 20% for having audit requirement
	}

	// Check if audit report exists
	if compliance.AuditReportUri != "" {
		score += weights.AuditWeight * 0.3 // 30% for having audit report

		// Check audit freshness
		if compliance.LastAuditDate != nil {
			auditAge := ctx.BlockTime().Sub(compliance.LastAuditDate.AsTime())
			auditAgeDays := int64(auditAge.Hours() / 24)

			if auditAgeDays <= weights.AuditAgePenaltyDays {
				// Full points for recent audit
				score += weights.AuditWeight * 0.5
			} else if auditAgeDays <= weights.MaxAuditAgeDays {
				// Decay points for older audits
				decayRatio := float64(auditAgeDays-weights.AuditAgePenaltyDays) /
					float64(weights.MaxAuditAgeDays-weights.AuditAgePenaltyDays)
				score += weights.AuditWeight * 0.5 * (1 - decayRatio)
			}
			// No points if audit is too old
		}
	}

	return score
}

// calculateMetricsScore computes the execution metrics component
func (k Keeper) calculateMetricsScore(metrics *pb.ContractMetrics, weights SecurityScoreWeights) float64 {
	if metrics == nil || metrics.TotalExecutions == 0 {
		return 0
	}

	var score float64 = 0

	// Success rate component (40% of metrics weight)
	successRate := float64(metrics.SuccessfulExecutions) / float64(metrics.TotalExecutions)
	if successRate >= weights.SuccessRateThreshold {
		score += weights.MetricsWeight * 0.4
	} else {
		score += weights.MetricsWeight * 0.4 * (successRate / weights.SuccessRateThreshold)
	}

	// Execution volume component (30% of metrics weight)
	if metrics.TotalExecutions >= weights.MinExecutionsForBonus {
		score += weights.MetricsWeight * 0.3
	} else {
		ratio := float64(metrics.TotalExecutions) / float64(weights.MinExecutionsForBonus)
		score += weights.MetricsWeight * 0.3 * ratio
	}

	// User adoption component (30% of metrics weight)
	if metrics.UniqueUsers >= weights.UniqueUsersThreshold {
		score += weights.MetricsWeight * 0.3
	} else if metrics.UniqueUsers > 0 {
		ratio := float64(metrics.UniqueUsers) / float64(weights.UniqueUsersThreshold)
		score += weights.MetricsWeight * 0.3 * ratio
	}

	return score
}

// calculateComplianceScore computes the compliance configuration component
func (k Keeper) calculateComplianceScore(info *pb.ContractInfo, weights SecurityScoreWeights) float64 {
	if info.Compliance == nil {
		return 0
	}

	compliance := info.Compliance
	var score float64 = 0

	// KYC enforcement (30% of compliance weight)
	if compliance.EnforceKyc {
		score += weights.ComplianceWeight * 0.15
		// Bonus for higher KYC levels
		score += weights.ComplianceWeight * 0.15 * math.Min(float64(compliance.MinKycLevel)/3.0, 1.0)
	}

	// Sanctions check enforcement (25% of compliance weight)
	if compliance.EnforceSanctionsCheck {
		score += weights.ComplianceWeight * 0.25
	}

	// Spending limits enforcement (20% of compliance weight)
	if compliance.EnforceSpendingLimits {
		score += weights.ComplianceWeight * 0.2
	}

	// Audit requirement (25% of compliance weight)
	if compliance.RequireAudit {
		score += weights.ComplianceWeight * 0.25
	}

	return score
}

// calculatePolicyScore computes the security policy configuration component
func (k Keeper) calculatePolicyScore(info *pb.ContractInfo, weights SecurityScoreWeights) float64 {
	if info.SecurityPolicy == nil {
		return 0
	}

	policy := info.SecurityPolicy
	var score float64 = 0

	// Pausability (20% of policy weight) - ability to pause is a security feature
	if policy.AllowPause {
		score += weights.PolicyWeight * 0.2
	}

	// Gas limits (25% of policy weight)
	if policy.MaxGasPerTx > 0 {
		score += weights.PolicyWeight * 0.25
	}

	// Rate limiting (25% of policy weight)
	if policy.RateLimitPerUser > 0 {
		score += weights.PolicyWeight * 0.25
	}

	// Access control via whitelist (15% of policy weight)
	if len(policy.WhitelistedAddresses) > 0 {
		score += weights.PolicyWeight * 0.15
	}

	// Blacklist usage (15% of policy weight)
	if len(policy.BlacklistedAddresses) > 0 {
		score += weights.PolicyWeight * 0.15
	}

	return score
}

// calculateHistoryScore computes the historical behavior component
func (k Keeper) calculateHistoryScore(ctx sdk.Context, contractAddr string, metrics *pb.ContractMetrics, weights SecurityScoreWeights) float64 {
	var score float64 = 0

	// No rate limit violations (50% of history weight)
	if metrics != nil && metrics.RateLimitViolations == 0 {
		score += weights.HistoryWeight * 0.5
	} else if metrics != nil && metrics.TotalExecutions > 0 {
		violationRatio := float64(metrics.RateLimitViolations) / float64(metrics.TotalExecutions)
		if violationRatio < 0.01 { // Less than 1% violations
			score += weights.HistoryWeight * 0.5 * (1 - violationRatio*100)
		}
	}

	// No compliance failures (50% of history weight)
	if metrics != nil && metrics.ComplianceFailures == 0 {
		score += weights.HistoryWeight * 0.5
	} else if metrics != nil && metrics.TotalExecutions > 0 {
		failureRatio := float64(metrics.ComplianceFailures) / float64(metrics.TotalExecutions)
		if failureRatio < 0.01 { // Less than 1% failures
			score += weights.HistoryWeight * 0.5 * (1 - failureRatio*100)
		}
	}

	return score
}

// calculateSourceCodeScore computes the source code availability component
func (k Keeper) calculateSourceCodeScore(info *pb.ContractInfo, weights SecurityScoreWeights) float64 {
	if info.Metadata == nil {
		return 0
	}

	metadata := info.Metadata
	var score float64 = 0

	// Source code URL available (60% of source code weight)
	if metadata.SourceCodeUrl != "" {
		score += weights.SourceCodeWeight * 0.6
	}

	// Homepage URL available (20% of source code weight)
	if metadata.Homepage != "" {
		score += weights.SourceCodeWeight * 0.2
	}

	// Description available (20% of source code weight)
	if metadata.Description != "" {
		score += weights.SourceCodeWeight * 0.2
	}

	return score
}

// calculateVulnerabilityPenalty computes deductions for known issues
func (k Keeper) calculateVulnerabilityPenalty(ctx sdk.Context, contractAddr string, metrics *pb.ContractMetrics, weights SecurityScoreWeights) float64 {
	var penalty float64 = 0

	// Check if contract is blacklisted (maximum penalty)
	if k.IsBlacklisted(ctx, contractAddr) {
		return -weights.VulnerabilityWeight
	}

	// Penalize high failure rates
	if metrics != nil && metrics.TotalExecutions > 0 {
		failureRate := float64(metrics.FailedExecutions) / float64(metrics.TotalExecutions)
		if failureRate > 0.1 { // More than 10% failure rate
			penalty -= weights.VulnerabilityWeight * 0.5 * math.Min(failureRate, 1.0)
		}
	}

	// Penalize high rate limit violations
	if metrics != nil && metrics.TotalExecutions > 0 {
		violationRate := float64(metrics.RateLimitViolations) / float64(metrics.TotalExecutions)
		if violationRate > 0.05 { // More than 5% violation rate
			penalty -= weights.VulnerabilityWeight * 0.3 * math.Min(violationRate*10, 1.0)
		}
	}

	// Penalize compliance failures
	if metrics != nil && metrics.TotalExecutions > 0 {
		complianceFailureRate := float64(metrics.ComplianceFailures) / float64(metrics.TotalExecutions)
		if complianceFailureRate > 0.02 { // More than 2% compliance failures
			penalty -= weights.VulnerabilityWeight * 0.2 * math.Min(complianceFailureRate*50, 1.0)
		}
	}

	return penalty
}

// determineRiskLevel determines the risk classification based on score
func (k Keeper) determineRiskLevel(score uint64) string {
	switch {
	case score >= 80:
		return "LOW"
	case score >= 60:
		return "MEDIUM"
	case score >= 40:
		return "HIGH"
	default:
		return "CRITICAL"
	}
}

// generateRecommendations creates actionable recommendations to improve score
func (k Keeper) generateRecommendations(ctx sdk.Context, info *pb.ContractInfo, metrics *pb.ContractMetrics, score *SecurityScore) []string {
	recommendations := make([]string, 0)

	// Audit recommendations
	if score.AuditScore < 20 {
		if info.Compliance == nil || info.Compliance.AuditReportUri == "" {
			recommendations = append(recommendations, "Add a security audit report to improve trust")
		}
		if info.Compliance == nil || !info.Compliance.RequireAudit {
			recommendations = append(recommendations, "Enable audit requirement in compliance settings")
		}
		if info.Compliance != nil && info.Compliance.LastAuditDate != nil {
			auditAge := ctx.BlockTime().Sub(info.Compliance.LastAuditDate.AsTime())
			if auditAge.Hours()/24 > 90 {
				recommendations = append(recommendations, "Update security audit - current audit is more than 90 days old")
			}
		}
	}

	// Compliance recommendations
	if score.ComplianceScore < 15 {
		if info.Compliance == nil || !info.Compliance.EnforceKyc {
			recommendations = append(recommendations, "Enable KYC enforcement for improved compliance")
		}
		if info.Compliance == nil || !info.Compliance.EnforceSanctionsCheck {
			recommendations = append(recommendations, "Enable sanctions screening for regulatory compliance")
		}
	}

	// Policy recommendations
	if score.PolicyScore < 10 {
		if info.SecurityPolicy == nil || info.SecurityPolicy.MaxGasPerTx == 0 {
			recommendations = append(recommendations, "Set maximum gas per transaction to prevent abuse")
		}
		if info.SecurityPolicy == nil || info.SecurityPolicy.RateLimitPerUser == 0 {
			recommendations = append(recommendations, "Enable rate limiting to prevent spam attacks")
		}
	}

	// Metrics recommendations
	if metrics != nil {
		if metrics.TotalExecutions > 0 {
			failureRate := float64(metrics.FailedExecutions) / float64(metrics.TotalExecutions)
			if failureRate > 0.05 {
				recommendations = append(recommendations, "High failure rate detected - review contract logic")
			}
		}
		if metrics.RateLimitViolations > 0 {
			recommendations = append(recommendations, "Rate limit violations detected - consider adjusting limits or investigating abuse")
		}
	}

	// Source code recommendations
	if score.SourceCodeScore < 3 {
		if info.Metadata == nil || info.Metadata.SourceCodeUrl == "" {
			recommendations = append(recommendations, "Publish source code to improve transparency and trust")
		}
	}

	return recommendations
}

// setSecurityScore stores a computed security score
func (k Keeper) setSecurityScore(ctx sdk.Context, contractAddr string, score *SecurityScore) {
	store := ctx.KVStore(k.storeKey)
	key := SecurityScoreKey(contractAddr)

	// Create a deterministic hash of the score for storage
	scoreData := make([]byte, 8)
	binary.BigEndian.PutUint64(scoreData, score.TotalScore)
	hash := sha256.Sum256(scoreData)

	// Store minimal data to track score validity
	store.Set(key, hash[:])
}

// GetSecurityScore retrieves a previously computed security score, or computes a new one
func (k Keeper) GetSecurityScore(ctx sdk.Context, contractAddr string) (SecurityScore, error) {
	// Always compute fresh score for now - could add caching later
	return k.ComputeSecurityScore(ctx, contractAddr)
}

// BatchComputeSecurityScores computes security scores for multiple contracts
func (k Keeper) BatchComputeSecurityScores(ctx sdk.Context, contractAddrs []string) ([]SecurityScore, error) {
	scores := make([]SecurityScore, 0, len(contractAddrs))

	for _, addr := range contractAddrs {
		score, err := k.ComputeSecurityScore(ctx, addr)
		if err != nil {
			// Skip contracts that error, but continue processing
			continue
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// GetContractsByRiskLevel retrieves all contracts matching a specific risk level
func (k Keeper) GetContractsByRiskLevel(ctx sdk.Context, riskLevel string) ([]string, error) {
	contracts := make([]string, 0)

	k.IterateContractInfo(ctx, func(info *pb.ContractInfo) bool {
		score, err := k.ComputeSecurityScore(ctx, info.Address)
		if err == nil && score.RiskLevel == riskLevel {
			contracts = append(contracts, info.Address)
		}
		return false // Continue iteration
	})

	return contracts, nil
}

// GetHighRiskContracts retrieves all contracts with HIGH or CRITICAL risk
func (k Keeper) GetHighRiskContracts(ctx sdk.Context) ([]SecurityScore, error) {
	scores := make([]SecurityScore, 0)

	k.IterateContractInfo(ctx, func(info *pb.ContractInfo) bool {
		score, err := k.ComputeSecurityScore(ctx, info.Address)
		if err == nil && (score.RiskLevel == "HIGH" || score.RiskLevel == "CRITICAL") {
			scores = append(scores, score)
		}
		return false // Continue iteration
	})

	return scores, nil
}

// iterateSecurityScores iterates over all security score entries (internal helper)
func (k Keeper) iterateSecurityScores(ctx sdk.Context, cb func(contractAddr string, bz []byte) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SecurityScoreKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		if len(key) <= len(types.SecurityScoreKeyPrefix) {
			continue
		}
		contractAddr := string(key[len(types.SecurityScoreKeyPrefix):])
		if cb(contractAddr, iterator.Value()) {
			break
		}
	}
}
