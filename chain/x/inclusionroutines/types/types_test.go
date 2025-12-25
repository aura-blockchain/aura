// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
)

func TestTypeExports(t *testing.T) {
	var _ IRStatus
	var _ PrivacyTier
	var _ Arena
	var _ IRDefinition
	var _ IRPrerequisite
	var _ IRGraphNode
	var _ IRRateLimit
	var _ Params
	var _ GenesisState
}

func TestMessageTypeExports(t *testing.T) {
	var _ MsgCreateIR
	var _ MsgUpdateIR
	var _ MsgDeleteIR
	var _ MsgActivateIR
	var _ MsgSuspendIR
	var _ MsgSetIRPrerequisites
	var _ MsgSetIRRateLimit
}

func TestQueryTypeExports(t *testing.T) {
	var _ QueryIRRequest
	var _ QueryIRResponse
	var _ QueryListIRsRequest
	var _ QueryListIRsResponse
	var _ QueryIRGraphRequest
	var _ QueryIRGraphResponse
	var _ QueryRateLimitRequest
	var _ QueryRateLimitResponse
	var _ QueryParamsRequest
	var _ QueryParamsResponse
}

func TestIRStatusEnums(t *testing.T) {
	statuses := []IRStatus{
		IRStatus_IR_STATUS_UNSPECIFIED,
		IRStatus_IR_STATUS_DRAFT,
		IRStatus_IR_STATUS_REVIEWING,
		IRStatus_IR_STATUS_APPROVED,
		IRStatus_IR_STATUS_ACTIVE,
		IRStatus_IR_STATUS_SUSPENDED,
		IRStatus_IR_STATUS_DEPRECATED,
		IRStatus_IR_STATUS_RETIRED,
	}

	seen := make(map[IRStatus]bool)
	for _, status := range statuses {
		if seen[status] {
			t.Errorf("duplicate IRStatus value: %v", status)
		}
		seen[status] = true
	}
}

func TestPrivacyTierEnums(t *testing.T) {
	tiers := []PrivacyTier{
		PrivacyTier_PRIVACY_TIER_UNSPECIFIED,
		PrivacyTier_PRIVACY_TIER_LOW,
		PrivacyTier_PRIVACY_TIER_MEDIUM,
		PrivacyTier_PRIVACY_TIER_HIGH,
	}

	seen := make(map[PrivacyTier]bool)
	for _, tier := range tiers {
		if seen[tier] {
			t.Errorf("duplicate PrivacyTier value: %v", tier)
		}
		seen[tier] = true
	}
}

func TestArenaEnums(t *testing.T) {
	arenas := []Arena{
		Arena_ARENA_UNSPECIFIED,
		Arena_ARENA_ANCHOR,
		Arena_ARENA_BIOMETRIC,
		Arena_ARENA_POSSESSION,
		Arena_ARENA_KNOWLEDGE,
		Arena_ARENA_SOCIAL,
		Arena_ARENA_GEOLOCATION,
		Arena_ARENA_HIGH_ASSURANCE,
		Arena_ARENA_PERSISTENCE,
		Arena_ARENA_SPECIALIZED,
	}

	seen := make(map[Arena]bool)
	for _, arena := range arenas {
		if seen[arena] {
			t.Errorf("duplicate Arena value: %v", arena)
		}
		seen[arena] = true
	}
}
