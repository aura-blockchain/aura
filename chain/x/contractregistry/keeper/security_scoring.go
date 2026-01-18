// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

// SecurityScoreWeights defines the weights for different scoring factors
// All weights use sdkmath.LegacyDec for deterministic cross-platform calculations
type SecurityScoreWeights struct {
	AuditWeight            sdkmath.LegacyDec
	MetricsWeight          sdkmath.LegacyDec
	ComplianceWeight       sdkmath.LegacyDec
	PolicyWeight           sdkmath.LegacyDec
	HistoryWeight          sdkmath.LegacyDec
	SourceCodeWeight       sdkmath.LegacyDec
	VulnerabilityWeight    sdkmath.LegacyDec
	TimeDecayFactor        sdkmath.LegacyDec
	AuditAgePenaltyDays    int64
	MaxAuditAgeDays        int64
	MinExecutionsForBonus  uint64
	SuccessRateThreshold   sdkmath.LegacyDec
	UniqueUsersThreshold   uint64
}

// DefaultSecurityScoreWeights returns the default scoring weights
// All decimal values use sdkmath.LegacyDec for deterministic calculations
func DefaultSecurityScoreWeights() SecurityScoreWeights {
	return SecurityScoreWeights{
		AuditWeight:            sdkmath.LegacyNewDec(25),         // 25 points max for audit status
		MetricsWeight:          sdkmath.LegacyNewDec(20),         // 20 points max for execution metrics
		ComplianceWeight:       sdkmath.LegacyNewDec(20),         // 20 points max for compliance configuration
		PolicyWeight:           sdkmath.LegacyNewDec(15),         // 15 points max for security policy
		HistoryWeight:          sdkmath.LegacyNewDec(10),         // 10 points max for historical behavior
		SourceCodeWeight:       sdkmath.LegacyNewDec(5),          // 5 points max for source code availability
		VulnerabilityWeight:    sdkmath.LegacyNewDec(5),          // 5 points deduction for known issues
		TimeDecayFactor:        sdkmath.LegacyNewDecWithPrec(1, 2), // 0.01 - Daily decay factor for stale audits
		AuditAgePenaltyDays:    90,                               // Start penalizing after 90 days
		MaxAuditAgeDays:        365,                              // Maximum audit age before minimum score
		MinExecutionsForBonus:  1000,                             // Minimum executions for reliability bonus
		SuccessRateThreshold:   sdkmath.LegacyNewDecWithPrec(95, 2), // 0.95 - 95% success rate for full metrics score
		UniqueUsersThreshold:   100,                              // Unique users threshold for adoption score
	}
}

// SecurityScore represents the computed security score for a contract
// All score values use sdkmath.LegacyDec for deterministic calculations
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
	ScoreBreakdown     map[string]sdkmath.LegacyDec
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
		ScoreBreakdown:  make(map[string]sdkmath.LegacyDec),
		Recommendations: make([]string, 0),
	}

	// 1. Audit Score (25 points max)
	auditScore := k.calculateAuditScore(ctx, &info, weights)
	score.AuditScore = auditScore.TruncateInt().Uint64()
	score.ScoreBreakdown["audit"] = auditScore

	// 2. Metrics Score (20 points max)
	metricsScore := k.calculateMetricsScore(metrics, weights)
	score.MetricsScore = metricsScore.TruncateInt().Uint64()
	score.ScoreBreakdown["metrics"] = metricsScore

	// 3. Compliance Score (20 points max)
	complianceScore := k.calculateComplianceScore(&info, weights)
	score.ComplianceScore = complianceScore.TruncateInt().Uint64()
	score.ScoreBreakdown["compliance"] = complianceScore

	// 4. Policy Score (15 points max)
	policyScore := k.calculatePolicyScore(&info, weights)
	score.PolicyScore = policyScore.TruncateInt().Uint64()
	score.ScoreBreakdown["policy"] = policyScore

	// 5. History Score (10 points max)
	historyScore := k.calculateHistoryScore(ctx, contractAddr, metrics, weights)
	score.HistoryScore = historyScore.TruncateInt().Uint64()
	score.ScoreBreakdown["history"] = historyScore

	// 6. Source Code Score (5 points max)
	sourceCodeScore := k.calculateSourceCodeScore(&info, weights)
	score.SourceCodeScore = sourceCodeScore.TruncateInt().Uint64()
	score.ScoreBreakdown["source_code"] = sourceCodeScore

	// 7. Vulnerability Deductions (up to -5 points)
	vulnerabilityPenalty := k.calculateVulnerabilityPenalty(ctx, contractAddr, metrics, weights)
	score.VulnerabilityScore = vulnerabilityPenalty.TruncateInt64()
	score.ScoreBreakdown["vulnerability_penalty"] = vulnerabilityPenalty

	// Calculate total score (0-100)
	totalScore := auditScore.Add(metricsScore).Add(complianceScore).Add(policyScore).
		Add(historyScore).Add(sourceCodeScore).Add(vulnerabilityPenalty)
	zero := sdkmath.LegacyZeroDec()
	hundred := sdkmath.LegacyNewDec(100)
	if totalScore.LT(zero) {
		totalScore = zero
	}
	if totalScore.GT(hundred) {
		totalScore = hundred
	}
	score.TotalScore = totalScore.TruncateInt().Uint64()

	// Determine risk level and recommendations
	score.RiskLevel = k.determineRiskLevel(score.TotalScore)
	score.Recommendations = k.generateRecommendations(ctx, &info, metrics, &score)

	// Store the computed score
	k.setSecurityScore(ctx, contractAddr, &score)

	return score, nil
}

// calculateAuditScore computes the audit component of the security score
// Uses sdkmath.LegacyDec for deterministic calculations
func (k Keeper) calculateAuditScore(ctx sdk.Context, info *pb.ContractInfo, weights SecurityScoreWeights) sdkmath.LegacyDec {
	compliance := info.Compliance
	score := sdkmath.LegacyZeroDec()

	// Precompute common multipliers for determinism
	pointTwo := sdkmath.LegacyNewDecWithPrec(2, 1)     // 0.2
	pointThree := sdkmath.LegacyNewDecWithPrec(3, 1)   // 0.3
	pointFive := sdkmath.LegacyNewDecWithPrec(5, 1)    // 0.5
	one := sdkmath.LegacyOneDec()

	// Check if audit is configured as required
	if compliance.RequireAudit {
		score = score.Add(weights.AuditWeight.Mul(pointTwo)) // 20% for having audit requirement
	}

	// Check if audit report exists
	if compliance.AuditReportUri != "" {
		score = score.Add(weights.AuditWeight.Mul(pointThree)) // 30% for having audit report

		// Check audit freshness
		if compliance.LastAuditDate != nil && !compliance.LastAuditDate.IsZero() {
			auditAge := ctx.BlockTime().Sub(*compliance.LastAuditDate)
			auditAgeDays := int64(auditAge.Hours() / 24)

			if auditAgeDays <= weights.AuditAgePenaltyDays {
				// Full points for recent audit
				score = score.Add(weights.AuditWeight.Mul(pointFive))
			} else if auditAgeDays <= weights.MaxAuditAgeDays {
				// Decay points for older audits using deterministic integer math
				// decayRatio = (auditAgeDays - AuditAgePenaltyDays) / (MaxAuditAgeDays - AuditAgePenaltyDays)
				numerator := sdkmath.LegacyNewDec(auditAgeDays - weights.AuditAgePenaltyDays)
				denominator := sdkmath.LegacyNewDec(weights.MaxAuditAgeDays - weights.AuditAgePenaltyDays)
				decayRatio := numerator.Quo(denominator)
				decayMultiplier := one.Sub(decayRatio)
				score = score.Add(weights.AuditWeight.Mul(pointFive).Mul(decayMultiplier))
			}
			// No points if audit is too old
		}
	}

	return score
}

// calculateMetricsScore computes the execution metrics component
// Uses sdkmath.LegacyDec for deterministic calculations
func (k Keeper) calculateMetricsScore(metrics *pb.ContractMetrics, weights SecurityScoreWeights) sdkmath.LegacyDec {
	if metrics == nil || metrics.TotalExecutions == 0 {
		return sdkmath.LegacyZeroDec()
	}

	score := sdkmath.LegacyZeroDec()

	// Precompute common multipliers for determinism
	pointFour := sdkmath.LegacyNewDecWithPrec(4, 1)   // 0.4
	pointThree := sdkmath.LegacyNewDecWithPrec(3, 1)  // 0.3

	// Success rate component (40% of metrics weight)
	// successRate = SuccessfulExecutions / TotalExecutions
	successfulExecs := sdkmath.LegacyNewDec(int64(metrics.SuccessfulExecutions))
	totalExecs := sdkmath.LegacyNewDec(int64(metrics.TotalExecutions))
	successRate := successfulExecs.Quo(totalExecs)

	if successRate.GTE(weights.SuccessRateThreshold) {
		score = score.Add(weights.MetricsWeight.Mul(pointFour))
	} else {
		// score += MetricsWeight * 0.4 * (successRate / SuccessRateThreshold)
		rateRatio := successRate.Quo(weights.SuccessRateThreshold)
		score = score.Add(weights.MetricsWeight.Mul(pointFour).Mul(rateRatio))
	}

	// Execution volume component (30% of metrics weight)
	minExecsForBonus := sdkmath.LegacyNewDec(int64(weights.MinExecutionsForBonus))
	if metrics.TotalExecutions >= weights.MinExecutionsForBonus {
		score = score.Add(weights.MetricsWeight.Mul(pointThree))
	} else {
		// ratio = TotalExecutions / MinExecutionsForBonus
		ratio := totalExecs.Quo(minExecsForBonus)
		score = score.Add(weights.MetricsWeight.Mul(pointThree).Mul(ratio))
	}

	// User adoption component (30% of metrics weight)
	uniqueUsersThreshold := sdkmath.LegacyNewDec(int64(weights.UniqueUsersThreshold))
	if metrics.UniqueUsers >= weights.UniqueUsersThreshold {
		score = score.Add(weights.MetricsWeight.Mul(pointThree))
	} else if metrics.UniqueUsers > 0 {
		// ratio = UniqueUsers / UniqueUsersThreshold
		uniqueUsers := sdkmath.LegacyNewDec(int64(metrics.UniqueUsers))
		ratio := uniqueUsers.Quo(uniqueUsersThreshold)
		score = score.Add(weights.MetricsWeight.Mul(pointThree).Mul(ratio))
	}

	return score
}

// calculateComplianceScore computes the compliance configuration component
// Uses sdkmath.LegacyDec for deterministic calculations
func (k Keeper) calculateComplianceScore(info *pb.ContractInfo, weights SecurityScoreWeights) sdkmath.LegacyDec {
	compliance := info.Compliance
	score := sdkmath.LegacyZeroDec()

	// Precompute common multipliers for determinism
	pointOneFive := sdkmath.LegacyNewDecWithPrec(15, 2)   // 0.15
	pointTwoFive := sdkmath.LegacyNewDecWithPrec(25, 2)   // 0.25
	pointTwo := sdkmath.LegacyNewDecWithPrec(2, 1)        // 0.2
	three := sdkmath.LegacyNewDec(3)
	one := sdkmath.LegacyOneDec()

	// KYC enforcement (30% of compliance weight)
	if compliance.EnforceKyc {
		score = score.Add(weights.ComplianceWeight.Mul(pointOneFive))
		// Bonus for higher KYC levels: min(MinKycLevel/3.0, 1.0)
		kycLevel := sdkmath.LegacyNewDec(int64(compliance.MinKycLevel))
		kycRatio := kycLevel.Quo(three)
		if kycRatio.GT(one) {
			kycRatio = one
		}
		score = score.Add(weights.ComplianceWeight.Mul(pointOneFive).Mul(kycRatio))
	}

	// Sanctions check enforcement (25% of compliance weight)
	if compliance.EnforceSanctionsCheck {
		score = score.Add(weights.ComplianceWeight.Mul(pointTwoFive))
	}

	// Spending limits enforcement (20% of compliance weight)
	if compliance.EnforceSpendingLimits {
		score = score.Add(weights.ComplianceWeight.Mul(pointTwo))
	}

	// Audit requirement (25% of compliance weight)
	if compliance.RequireAudit {
		score = score.Add(weights.ComplianceWeight.Mul(pointTwoFive))
	}

	return score
}

// calculatePolicyScore computes the security policy configuration component
// Uses sdkmath.LegacyDec for deterministic calculations
func (k Keeper) calculatePolicyScore(info *pb.ContractInfo, weights SecurityScoreWeights) sdkmath.LegacyDec {
	policy := info.SecurityPolicy
	score := sdkmath.LegacyZeroDec()

	// Precompute common multipliers for determinism
	pointTwo := sdkmath.LegacyNewDecWithPrec(2, 1)        // 0.2
	pointTwoFive := sdkmath.LegacyNewDecWithPrec(25, 2)   // 0.25
	pointOneFive := sdkmath.LegacyNewDecWithPrec(15, 2)   // 0.15

	// Pausability (20% of policy weight) - ability to pause is a security feature
	if policy.AllowPause {
		score = score.Add(weights.PolicyWeight.Mul(pointTwo))
	}

	// Gas limits (25% of policy weight)
	if policy.MaxGasPerTx > 0 {
		score = score.Add(weights.PolicyWeight.Mul(pointTwoFive))
	}

	// Rate limiting (25% of policy weight)
	if policy.RateLimitPerUser > 0 {
		score = score.Add(weights.PolicyWeight.Mul(pointTwoFive))
	}

	// Access control via whitelist (15% of policy weight)
	if len(policy.WhitelistedAddresses) > 0 {
		score = score.Add(weights.PolicyWeight.Mul(pointOneFive))
	}

	// Blacklist usage (15% of policy weight)
	if len(policy.BlacklistedAddresses) > 0 {
		score = score.Add(weights.PolicyWeight.Mul(pointOneFive))
	}

	return score
}

// calculateHistoryScore computes the historical behavior component
// Uses sdkmath.LegacyDec for deterministic calculations
func (k Keeper) calculateHistoryScore(ctx sdk.Context, contractAddr string, metrics *pb.ContractMetrics, weights SecurityScoreWeights) sdkmath.LegacyDec {
	score := sdkmath.LegacyZeroDec()

	// Precompute common multipliers for determinism
	pointFive := sdkmath.LegacyNewDecWithPrec(5, 1)       // 0.5
	pointZeroOne := sdkmath.LegacyNewDecWithPrec(1, 2)    // 0.01
	hundred := sdkmath.LegacyNewDec(100)
	one := sdkmath.LegacyOneDec()

	// No rate limit violations (50% of history weight)
	if metrics != nil && metrics.RateLimitViolations == 0 {
		score = score.Add(weights.HistoryWeight.Mul(pointFive))
	} else if metrics != nil && metrics.TotalExecutions > 0 {
		// violationRatio = RateLimitViolations / TotalExecutions
		violations := sdkmath.LegacyNewDec(int64(metrics.RateLimitViolations))
		totalExecs := sdkmath.LegacyNewDec(int64(metrics.TotalExecutions))
		violationRatio := violations.Quo(totalExecs)
		if violationRatio.LT(pointZeroOne) { // Less than 1% violations
			// score += HistoryWeight * 0.5 * (1 - violationRatio*100)
			penalty := one.Sub(violationRatio.Mul(hundred))
			if penalty.IsPositive() {
				score = score.Add(weights.HistoryWeight.Mul(pointFive).Mul(penalty))
			}
		}
	}

	// No compliance failures (50% of history weight)
	if metrics != nil && metrics.ComplianceFailures == 0 {
		score = score.Add(weights.HistoryWeight.Mul(pointFive))
	} else if metrics != nil && metrics.TotalExecutions > 0 {
		// failureRatio = ComplianceFailures / TotalExecutions
		failures := sdkmath.LegacyNewDec(int64(metrics.ComplianceFailures))
		totalExecs := sdkmath.LegacyNewDec(int64(metrics.TotalExecutions))
		failureRatio := failures.Quo(totalExecs)
		if failureRatio.LT(pointZeroOne) { // Less than 1% failures
			// score += HistoryWeight * 0.5 * (1 - failureRatio*100)
			penalty := one.Sub(failureRatio.Mul(hundred))
			if penalty.IsPositive() {
				score = score.Add(weights.HistoryWeight.Mul(pointFive).Mul(penalty))
			}
		}
	}

	return score
}

// calculateSourceCodeScore computes the source code availability component
// Uses sdkmath.LegacyDec for deterministic calculations
func (k Keeper) calculateSourceCodeScore(info *pb.ContractInfo, weights SecurityScoreWeights) sdkmath.LegacyDec {
	metadata := info.Metadata
	score := sdkmath.LegacyZeroDec()

	// Precompute common multipliers for determinism
	pointSix := sdkmath.LegacyNewDecWithPrec(6, 1)   // 0.6
	pointTwo := sdkmath.LegacyNewDecWithPrec(2, 1)   // 0.2

	// Source code URL available (60% of source code weight)
	if metadata.SourceCodeUrl != "" {
		score = score.Add(weights.SourceCodeWeight.Mul(pointSix))
	}

	// Homepage URL available (20% of source code weight)
	if metadata.Homepage != "" {
		score = score.Add(weights.SourceCodeWeight.Mul(pointTwo))
	}

	// Description available (20% of source code weight)
	if metadata.Description != "" {
		score = score.Add(weights.SourceCodeWeight.Mul(pointTwo))
	}

	return score
}

// calculateVulnerabilityPenalty computes deductions for known issues
// Uses sdkmath.LegacyDec for deterministic calculations
func (k Keeper) calculateVulnerabilityPenalty(ctx sdk.Context, contractAddr string, metrics *pb.ContractMetrics, weights SecurityScoreWeights) sdkmath.LegacyDec {
	penalty := sdkmath.LegacyZeroDec()

	// Check if contract is blacklisted (maximum penalty)
	if k.IsBlacklisted(ctx, contractAddr) {
		return weights.VulnerabilityWeight.Neg()
	}

	// Precompute common multipliers for determinism
	pointFive := sdkmath.LegacyNewDecWithPrec(5, 1)       // 0.5
	pointThree := sdkmath.LegacyNewDecWithPrec(3, 1)      // 0.3
	pointTwo := sdkmath.LegacyNewDecWithPrec(2, 1)        // 0.2
	pointOneZero := sdkmath.LegacyNewDecWithPrec(1, 1)    // 0.1
	pointZeroFive := sdkmath.LegacyNewDecWithPrec(5, 2)   // 0.05
	pointZeroTwo := sdkmath.LegacyNewDecWithPrec(2, 2)    // 0.02
	ten := sdkmath.LegacyNewDec(10)
	fifty := sdkmath.LegacyNewDec(50)
	one := sdkmath.LegacyOneDec()

	// Penalize high failure rates
	if metrics != nil && metrics.TotalExecutions > 0 {
		// failureRate = FailedExecutions / TotalExecutions
		failed := sdkmath.LegacyNewDec(int64(metrics.FailedExecutions))
		totalExecs := sdkmath.LegacyNewDec(int64(metrics.TotalExecutions))
		failureRate := failed.Quo(totalExecs)
		if failureRate.GT(pointOneZero) { // More than 10% failure rate
			// penalty -= VulnerabilityWeight * 0.5 * min(failureRate, 1.0)
			cappedRate := failureRate
			if cappedRate.GT(one) {
				cappedRate = one
			}
			penalty = penalty.Sub(weights.VulnerabilityWeight.Mul(pointFive).Mul(cappedRate))
		}
	}

	// Penalize high rate limit violations
	if metrics != nil && metrics.TotalExecutions > 0 {
		// violationRate = RateLimitViolations / TotalExecutions
		violations := sdkmath.LegacyNewDec(int64(metrics.RateLimitViolations))
		totalExecs := sdkmath.LegacyNewDec(int64(metrics.TotalExecutions))
		violationRate := violations.Quo(totalExecs)
		if violationRate.GT(pointZeroFive) { // More than 5% violation rate
			// penalty -= VulnerabilityWeight * 0.3 * min(violationRate*10, 1.0)
			scaledRate := violationRate.Mul(ten)
			if scaledRate.GT(one) {
				scaledRate = one
			}
			penalty = penalty.Sub(weights.VulnerabilityWeight.Mul(pointThree).Mul(scaledRate))
		}
	}

	// Penalize compliance failures
	if metrics != nil && metrics.TotalExecutions > 0 {
		// complianceFailureRate = ComplianceFailures / TotalExecutions
		failures := sdkmath.LegacyNewDec(int64(metrics.ComplianceFailures))
		totalExecs := sdkmath.LegacyNewDec(int64(metrics.TotalExecutions))
		complianceFailureRate := failures.Quo(totalExecs)
		if complianceFailureRate.GT(pointZeroTwo) { // More than 2% compliance failures
			// penalty -= VulnerabilityWeight * 0.2 * min(complianceFailureRate*50, 1.0)
			scaledRate := complianceFailureRate.Mul(fifty)
			if scaledRate.GT(one) {
				scaledRate = one
			}
			penalty = penalty.Sub(weights.VulnerabilityWeight.Mul(pointTwo).Mul(scaledRate))
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
		if info.Compliance.AuditReportUri == "" {
			recommendations = append(recommendations, "Add a security audit report to improve trust")
		}
		if !info.Compliance.RequireAudit {
			recommendations = append(recommendations, "Enable audit requirement in compliance settings")
		}
		if info.Compliance.LastAuditDate != nil && !info.Compliance.LastAuditDate.IsZero() {
			auditAge := ctx.BlockTime().Sub(*info.Compliance.LastAuditDate)
			if auditAge.Hours()/24 > 90 {
				recommendations = append(recommendations, "Update security audit - current audit is more than 90 days old")
			}
		}
	}

	// Compliance recommendations
	if score.ComplianceScore < 15 {
		if !info.Compliance.EnforceKyc {
			recommendations = append(recommendations, "Enable KYC enforcement for improved compliance")
		}
		if !info.Compliance.EnforceSanctionsCheck {
			recommendations = append(recommendations, "Enable sanctions screening for regulatory compliance")
		}
	}

	// Policy recommendations
	if score.PolicyScore < 10 {
		if info.SecurityPolicy.MaxGasPerTx == 0 {
			recommendations = append(recommendations, "Set maximum gas per transaction to prevent abuse")
		}
		if info.SecurityPolicy.RateLimitPerUser == 0 {
			recommendations = append(recommendations, "Enable rate limiting to prevent spam attacks")
		}
	}

	// Metrics recommendations
	if metrics != nil {
		if metrics.TotalExecutions > 0 {
			// Use LegacyDec for deterministic comparison
			// failureRate = FailedExecutions / TotalExecutions
			failed := sdkmath.LegacyNewDec(int64(metrics.FailedExecutions))
			totalExecs := sdkmath.LegacyNewDec(int64(metrics.TotalExecutions))
			failureRate := failed.Quo(totalExecs)
			threshold := sdkmath.LegacyNewDecWithPrec(5, 2) // 0.05
			if failureRate.GT(threshold) {
				recommendations = append(recommendations, "High failure rate detected - review contract logic")
			}
		}
		if metrics.RateLimitViolations > 0 {
			recommendations = append(recommendations, "Rate limit violations detected - consider adjusting limits or investigating abuse")
		}
	}

	// Source code recommendations
	if score.SourceCodeScore < 3 {
		if info.Metadata.SourceCodeUrl == "" {
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
