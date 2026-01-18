// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"regexp"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BiasDetectionResult represents the result of bias detection
type BiasDetectionResult struct {
	HasBias            bool
	BiasScore          sdkmath.LegacyDec // 0.0 to 1.0 using deterministic decimal
	BiasTypes          []string
	ProblematicPhrases []string
	Recommendations    []string
	Severity           BiasSeverity
}

// BiasSeverity defines bias severity levels
type BiasSeverity string

const (
	SeverityNone   BiasSeverity = "none"
	SeverityLow    BiasSeverity = "low"
	SeverityMedium BiasSeverity = "medium"
	SeverityHigh   BiasSeverity = "high"
)

// BiasCategory defines types of bias to detect
type BiasCategory string

const (
	BiasGender       BiasCategory = "gender"
	BiasRace         BiasCategory = "race"
	BiasAge          BiasCategory = "age"
	BiasReligion     BiasCategory = "religion"
	BiasNationality  BiasCategory = "nationality"
	BiasDisability   BiasCategory = "disability"
	BiasSocioeconomic BiasCategory = "socioeconomic"
)

// BiasPattern represents a bias detection pattern
type BiasPattern struct {
	Category    BiasCategory
	Pattern     *regexp.Regexp
	Severity    BiasSeverity
	Description string
}

// GetBiasPatterns returns predefined bias detection patterns
func GetBiasPatterns() []BiasPattern {
	return []BiasPattern{
		// Gender bias patterns
		{
			Category:    BiasGender,
			Pattern:     regexp.MustCompile(`(?i)\b(men are better|women should|ladies only|guys only)\b`),
			Severity:    SeverityHigh,
			Description: "Gender stereotyping detected",
		},
		{
			Category:    BiasGender,
			Pattern:     regexp.MustCompile(`(?i)\b(male dominated|female dominated|boys club|girls club)\b`),
			Severity:    SeverityMedium,
			Description: "Gender-based generalization",
		},
		// Race/ethnicity bias patterns
		{
			Category:    BiasRace,
			Pattern:     regexp.MustCompile(`(?i)\b(all [race] are|typical [race]|you people)\b`),
			Severity:    SeverityHigh,
			Description: "Racial stereotyping detected",
		},
		// Age bias patterns
		{
			Category:    BiasAge,
			Pattern:     regexp.MustCompile(`(?i)\b(too old|too young|elderly people|millennials are)\b`),
			Severity:    SeverityMedium,
			Description: "Age-based stereotyping",
		},
		// Disability bias patterns
		{
			Category:    BiasDisability,
			Pattern:     regexp.MustCompile(`(?i)\b(handicapped|crippled|suffers from|victim of)\b`),
			Severity:    SeverityHigh,
			Description: "Ableist language detected",
		},
		// Socioeconomic bias patterns
		{
			Category:    BiasSocioeconomic,
			Pattern:     regexp.MustCompile(`(?i)\b(poor people are|rich people are|lower class|upper class)\b`),
			Severity:    SeverityMedium,
			Description: "Socioeconomic stereotyping",
		},
	}
}

// DetectBias analyzes text for potential bias
func (k Keeper) DetectBias(ctx sdk.Context, text string) BiasDetectionResult {
	result := BiasDetectionResult{
		BiasTypes:          []string{},
		ProblematicPhrases: []string{},
		Recommendations:    []string{},
	}

	patterns := GetBiasPatterns()
	biasScores := make(map[BiasSeverity]int)

	for _, pattern := range patterns {
		matches := pattern.Pattern.FindAllString(text, -1)
		if len(matches) > 0 {
			result.HasBias = true
			result.BiasTypes = append(result.BiasTypes, string(pattern.Category))
			result.ProblematicPhrases = append(result.ProblematicPhrases, matches...)

			// Increment severity count
			biasScores[pattern.Severity]++

			// Add recommendation
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Review %s: %s", pattern.Category, pattern.Description))
		}
	}

	// Calculate overall bias score
	result.BiasScore = calculateBiasScore(biasScores)
	result.Severity = determineSeverity(result.BiasScore)

	// Log bias detection
	if result.HasBias {
		k.logBiasDetection(ctx, result)
	}

	return result
}

// calculateBiasScore calculates overall bias score using deterministic decimal arithmetic
func calculateBiasScore(scores map[BiasSeverity]int) sdkmath.LegacyDec {
	// Weights as basis points (25 = 0.25, 50 = 0.50, 100 = 1.00)
	weightsBasisPoints := map[BiasSeverity]int64{
		SeverityLow:    25,
		SeverityMedium: 50,
		SeverityHigh:   100,
	}

	totalScore := sdkmath.LegacyZeroDec()
	totalCount := int64(0)

	// Process severities in deterministic order
	for _, severity := range []BiasSeverity{SeverityLow, SeverityMedium, SeverityHigh} {
		count, exists := scores[severity]
		if !exists || count == 0 {
			continue
		}
		weight := sdkmath.LegacyNewDec(weightsBasisPoints[severity]).QuoInt64(100)
		contribution := weight.MulInt64(int64(count))
		totalScore = totalScore.Add(contribution)
		totalCount += int64(count)
	}

	if totalCount == 0 {
		return sdkmath.LegacyZeroDec()
	}

	// Normalize to 0-1 range: score = totalScore / totalCount
	score := totalScore.QuoInt64(totalCount)
	if score.GT(sdkmath.LegacyOneDec()) {
		score = sdkmath.LegacyOneDec()
	}

	return score
}

// determineSeverity determines overall severity from score using deterministic decimal comparison
func determineSeverity(score sdkmath.LegacyDec) BiasSeverity {
	lowThreshold := sdkmath.LegacyNewDecWithPrec(3, 1)  // 0.3
	medThreshold := sdkmath.LegacyNewDecWithPrec(7, 1)  // 0.7

	switch {
	case score.IsZero():
		return SeverityNone
	case score.LT(lowThreshold):
		return SeverityLow
	case score.LT(medThreshold):
		return SeverityMedium
	default:
		return SeverityHigh
	}
}

// ValidateResponseForBias validates AI response before returning to user
func (k Keeper) ValidateResponseForBias(ctx sdk.Context, response string) error {
	result := k.DetectBias(ctx, response)

	// Reject responses with high bias
	if result.Severity == SeverityHigh {
		return fmt.Errorf("response contains high-severity bias and cannot be returned")
	}

	// Warn for medium bias
	if result.Severity == SeverityMedium {
		// Format score with 2 decimal places using deterministic string conversion
		scoreStr := result.BiasScore.MulInt64(100).TruncateInt64()
		formattedScore := fmt.Sprintf("%d.%02d", scoreStr/100, scoreStr%100)
		if scoreStr < 0 {
			formattedScore = "0.00"
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"bias_warning",
				sdk.NewAttribute("severity", string(result.Severity)),
				sdk.NewAttribute("score", formattedScore),
			),
		)
	}

	return nil
}

// logBiasDetection logs bias detection for monitoring
func (k Keeper) logBiasDetection(ctx sdk.Context, result BiasDetectionResult) {
	// Format score with 2 decimal places using deterministic string conversion
	scoreStr := result.BiasScore.MulInt64(100).TruncateInt64()
	formattedScore := fmt.Sprintf("%d.%02d", scoreStr/100, scoreStr%100)
	if scoreStr < 0 {
		formattedScore = "0.00"
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bias_detected",
			sdk.NewAttribute("score", formattedScore),
			sdk.NewAttribute("severity", string(result.Severity)),
			sdk.NewAttribute("types", strings.Join(result.BiasTypes, ",")),
		),
	)

	// Could also store in audit log
	audit := AuditLog{
		OperationType: "bias_detection",
		Success:       true,
		Metadata: map[string]string{
			"bias_score": formattedScore,
			"severity":   string(result.Severity),
			"types":      strings.Join(result.BiasTypes, ","),
		},
	}
	k.LogAIOperation(ctx, audit)
}

// GetBiasStatistics returns bias detection statistics
func (k Keeper) GetBiasStatistics(ctx sdk.Context) BiasStatistics {
	// Would query audit logs for bias detection events
	return BiasStatistics{
		TotalDetections: 0,
		ByCategoryCount: make(map[BiasCategory]uint64),
		BySeverityCount: make(map[BiasSeverity]uint64),
	}
}

// BiasStatistics represents bias detection statistics
type BiasStatistics struct {
	TotalDetections uint64
	ByCategoryCount map[BiasCategory]uint64
	BySeverityCount map[BiasSeverity]uint64
	AverageSeverity sdkmath.LegacyDec // Deterministic decimal for consensus safety
}
