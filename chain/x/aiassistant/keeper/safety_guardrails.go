// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"regexp"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// SafetyGuardrail defines a safety check for AI operations
type SafetyGuardrail struct {
	Name        string
	Category    SafetyCategory
	Enabled     bool
	Severity    SafetySeverity
	CheckFunc   func(ctx sdk.Context, input, output string) SafetyViolation
	Description string
}

// SafetyCategory defines categories of safety checks
type SafetyCategory string

const (
	SafetyContentFilter SafetyCategory = "content_filter"
	SafetyPII           SafetyCategory = "pii_detection"
	SafetyToxicity      SafetyCategory = "toxicity"
	SafetyViolence      SafetyCategory = "violence"
	SafetyMisinformation SafetyCategory = "misinformation"
	SafetySecurity      SafetyCategory = "security"
	SafetyPrivacy       SafetyCategory = "privacy"
)

// SafetySeverity defines severity of safety violations
type SafetySeverity string

const (
	SafetySeverityLow      SafetySeverity = "low"
	SafetySeverityMedium   SafetySeverity = "medium"
	SafetySeverityHigh     SafetySeverity = "high"
	SafetySeverityCritical SafetySeverity = "critical"
)

// SafetyViolation represents a safety violation
type SafetyViolation struct {
	Detected     bool
	Category     SafetyCategory
	Severity     SafetySeverity
	Reason       string
	Details      string
	Blocked      bool
	Suggestions  []string
}

// SafetyCheckResult represents the result of all safety checks
type SafetyCheckResult struct {
	Passed       bool
	Violations   []SafetyViolation
	OverallScore float64
	Blocked      bool
}

// GetSafetyGuardrails returns all safety guardrails
func (k Keeper) GetSafetyGuardrails() []SafetyGuardrail {
	return []SafetyGuardrail{
		{
			Name:        "PII Detection",
			Category:    SafetyPII,
			Enabled:     true,
			Severity:    SafetySeverityHigh,
			CheckFunc:   checkPII,
			Description: "Detects and blocks personally identifiable information",
		},
		{
			Name:        "Toxicity Filter",
			Category:    SafetyToxicity,
			Enabled:     true,
			Severity:    SafetySeverityCritical,
			CheckFunc:   checkToxicity,
			Description: "Filters toxic and harmful content",
		},
		{
			Name:        "Violence Detection",
			Category:    SafetyViolence,
			Enabled:     true,
			Severity:    SafetySeverityHigh,
			CheckFunc:   checkViolence,
			Description: "Detects violent or harmful content",
		},
		{
			Name:        "Security Check",
			Category:    SafetySecurity,
			Enabled:     true,
			Severity:    SafetySeverityCritical,
			CheckFunc:   checkSecurity,
			Description: "Prevents security vulnerabilities",
		},
		{
			Name:        "Privacy Protection",
			Category:    SafetyPrivacy,
			Enabled:     true,
			Severity:    SafetySeverityHigh,
			CheckFunc:   checkPrivacy,
			Description: "Protects user privacy",
		},
	}
}

// CheckSafety runs all safety guardrails on input and output
func (k Keeper) CheckSafety(ctx sdk.Context, input, output string) (SafetyCheckResult, error) {
	result := SafetyCheckResult{
		Passed:     true,
		Violations: []SafetyViolation{},
	}

	guardrails := k.GetSafetyGuardrails()

	for _, guardrail := range guardrails {
		if !guardrail.Enabled {
			continue
		}

		violation := guardrail.CheckFunc(ctx, input, output)
		if violation.Detected {
			result.Passed = false
			result.Violations = append(result.Violations, violation)

			// Critical violations block the operation
			if violation.Severity == SafetySeverityCritical {
				result.Blocked = true
			}
		}
	}

	// Calculate overall safety score
	result.OverallScore = k.calculateSafetyScore(result.Violations)

	// Log safety check
	k.logSafetyCheck(ctx, result)

	// Return error if blocked
	if result.Blocked {
		return result, sdkerrors.Wrap(types.ErrInvalidParams, "safety check failed: critical violation detected")
	}

	return result, nil
}

// checkPII detects personally identifiable information
func checkPII(ctx sdk.Context, input, output string) SafetyViolation {
	violation := SafetyViolation{
		Category: SafetyPII,
		Severity: SafetySeverityHigh,
	}

	// Patterns for common PII
	patterns := []struct {
		pattern *regexp.Regexp
		name    string
	}{
		{regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`), "SSN"},
		{regexp.MustCompile(`\b\d{16}\b`), "Credit Card"},
		{regexp.MustCompile(`\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`), "Email"},
		{regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`), "Phone Number"},
	}

	text := input + " " + output
	detected := []string{}

	for _, p := range patterns {
		if p.pattern.MatchString(text) {
			detected = append(detected, p.name)
		}
	}

	if len(detected) > 0 {
		violation.Detected = true
		violation.Reason = "PII detected in content"
		violation.Details = fmt.Sprintf("Found: %s", strings.Join(detected, ", "))
		violation.Blocked = true
		violation.Suggestions = []string{"Remove or redact personally identifiable information"}
	}

	return violation
}

// checkToxicity detects toxic content
func checkToxicity(ctx sdk.Context, input, output string) SafetyViolation {
	violation := SafetyViolation{
		Category: SafetyToxicity,
		Severity: SafetySeverityCritical,
	}

	toxicPatterns := []string{
		`(?i)\b(hate|racist|sexist|bigot|offensive)\b`,
		`(?i)\b(profanity|curse|swear)\b`,
		// Add more patterns
	}

	text := strings.ToLower(input + " " + output)
	detected := false

	for _, pattern := range toxicPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			detected = true
			break
		}
	}

	if detected {
		violation.Detected = true
		violation.Reason = "Toxic content detected"
		violation.Details = "Content contains potentially harmful or offensive language"
		violation.Blocked = true
		violation.Suggestions = []string{"Rephrase using respectful language", "Remove offensive terms"}
	}

	return violation
}

// checkViolence detects violent content
func checkViolence(ctx sdk.Context, input, output string) SafetyViolation {
	violation := SafetyViolation{
		Category: SafetyViolence,
		Severity: SafetySeverityHigh,
	}

	violencePatterns := []string{
		`(?i)\b(kill|murder|attack|assault|harm|hurt)\b`,
		`(?i)\b(weapon|gun|knife|bomb|explosive)\b`,
	}

	text := strings.ToLower(input + " " + output)
	detected := false

	for _, pattern := range violencePatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			detected = true
			break
		}
	}

	if detected {
		violation.Detected = true
		violation.Reason = "Violent content detected"
		violation.Details = "Content references violence or harmful actions"
		violation.Blocked = false // Warning only
		violation.Suggestions = []string{"Review content for violent themes"}
	}

	return violation
}

// checkSecurity detects security vulnerabilities
func checkSecurity(ctx sdk.Context, input, output string) SafetyViolation {
	violation := SafetyViolation{
		Category: SafetySecurity,
		Severity: SafetySeverityCritical,
	}

	securityPatterns := []string{
		`(?i)(private key|secret key|api key|password)`,
		`(?i)(sql injection|xss|csrf)`,
		`(?i)(exploit|vulnerability|hack)`,
	}

	text := strings.ToLower(input + " " + output)
	detected := false

	for _, pattern := range securityPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			detected = true
			break
		}
	}

	if detected {
		violation.Detected = true
		violation.Reason = "Security concern detected"
		violation.Details = "Content may contain security-sensitive information"
		violation.Blocked = true
		violation.Suggestions = []string{"Remove security-sensitive information", "Use secure alternatives"}
	}

	return violation
}

// checkPrivacy checks for privacy violations
func checkPrivacy(ctx sdk.Context, input, output string) SafetyViolation {
	violation := SafetyViolation{
		Category: SafetyPrivacy,
		Severity: SafetySeverityHigh,
	}

	privacyPatterns := []string{
		`(?i)(private|confidential|secret)\s+(data|information)`,
		`(?i)(personal|sensitive)\s+(information|data)`,
	}

	text := strings.ToLower(input + " " + output)
	detected := false

	for _, pattern := range privacyPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			detected = true
			break
		}
	}

	if detected {
		violation.Detected = true
		violation.Reason = "Privacy concern detected"
		violation.Details = "Content may contain private or confidential information"
		violation.Blocked = false
		violation.Suggestions = []string{"Review privacy implications", "Ensure proper consent"}
	}

	return violation
}

// calculateSafetyScore calculates overall safety score
func (k Keeper) calculateSafetyScore(violations []SafetyViolation) float64 {
	if len(violations) == 0 {
		return 1.0 // Perfect score
	}

	severityWeights := map[SafetySeverity]float64{
		SafetySeverityLow:      0.1,
		SafetySeverityMedium:   0.3,
		SafetySeverityHigh:     0.6,
		SafetySeverityCritical: 1.0,
	}

	totalPenalty := 0.0
	for _, v := range violations {
		totalPenalty += severityWeights[v.Severity]
	}

	// Score from 0 to 1, where 1 is safest
	score := 1.0 - (totalPenalty / float64(len(violations)))
	if score < 0 {
		score = 0
	}

	return score
}

// logSafetyCheck logs safety check results
func (k Keeper) logSafetyCheck(ctx sdk.Context, result SafetyCheckResult) {
	if len(result.Violations) > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"safety_check",
				sdk.NewAttribute("passed", fmt.Sprintf("%t", result.Passed)),
				sdk.NewAttribute("violations", fmt.Sprintf("%d", len(result.Violations))),
				sdk.NewAttribute("score", fmt.Sprintf("%.2f", result.OverallScore)),
				sdk.NewAttribute("blocked", fmt.Sprintf("%t", result.Blocked)),
			),
		)
	}
}
