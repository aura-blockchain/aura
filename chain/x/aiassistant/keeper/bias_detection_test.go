// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import "testing"

func TestDetectBias(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name       string
		text       string
		expectBias bool
	}{
		{
			name:       "neutral text",
			text:       "What is the weather like today?",
			expectBias: false,
		},
		{
			name:       "gender bias detected",
			text:       "men are better than women at science",
			expectBias: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := k.DetectBias(ctx, tt.text)
			if tt.expectBias && !result.HasBias {
				t.Errorf("expected bias to be detected for input %q", tt.text)
			}
			if !tt.expectBias && result.HasBias {
				t.Errorf("expected no bias for input %q", tt.text)
			}
		})
	}
}

func TestBiasSeverityLevels(t *testing.T) {
	k, ctx := setupKeeper(t)

	cases := []struct {
		text     string
		expected BiasSeverity
	}{
		{
			text:     "men are better than women at science",
			expected: SeverityHigh,
		},
		{
			text:     "millennials are always entitled",
			expected: SeverityMedium,
		},
		{
			text:     "Everyone deserves equal respect.",
			expected: SeverityNone,
		},
	}

	for _, tc := range cases {
		result := k.DetectBias(ctx, tc.text)
		if result.Severity != tc.expected {
			t.Errorf("expected severity %s for text %q, got %s", tc.expected, tc.text, result.Severity)
		}
		if result.BiasScore < 0 || result.BiasScore > 1 {
			t.Errorf("bias score should be normalized, got %f", result.BiasScore)
		}
	}
}

func TestValidateResponseForBias(t *testing.T) {
	k, ctx := setupKeeper(t)

	// High severity response should be rejected
	if err := k.ValidateResponseForBias(ctx, "men are better than women at science"); err == nil {
		t.Fatalf("expected validation error for biased response")
	}

	// Neutral response should pass
	if err := k.ValidateResponseForBias(ctx, "This option is suitable for many users."); err != nil {
		t.Fatalf("unexpected error for neutral response: %v", err)
	}
}

func TestBiasDetectionEmitsAuditLog(t *testing.T) {
	k, ctx := setupKeeper(t)

	result := k.DetectBias(ctx, "men are better than women at science")
	if !result.HasBias {
		t.Fatalf("expected bias to be detected")
	}

	// Ensure a bias_detection audit entry was recorded
	logs := k.GetAuditLogs(ctx, 5)
	found := false
	for _, log := range logs {
		if log.OperationType == "bias_detection" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected bias_detection audit log to be recorded")
	}
}
