// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import "fmt"

// Event types for the aiassistant module
const (
	EventTypeRegisterAssistant  = "register_assistant"
	EventTypeUpdateLocales      = "update_locales"
	EventTypeHeartbeat          = "heartbeat"
	EventTypeHeartbeatFailure   = "heartbeat_failure"
	EventTypeReportMisbehavior  = "report_misbehavior"
	EventTypeSlashAssistant     = "slash_assistant"
	EventTypeJailAssistant      = "jail_assistant"
	EventTypeTombstoneAssistant = "tombstone_assistant"
	EventTypeStakeDeposit       = "stake_deposit"
	EventTypeStakeWithdrawal    = "stake_withdrawal"
	EventTypeSponsorshipAdded   = "sponsorship_added"
	EventTypeStatusChange       = "status_change"
	EventTypeParamsUpdate       = "params_update"
)

// Event attribute keys
const (
	AttributeKeyAssistantAddress  = "assistant_address"
	AttributeKeyOwnerAddress      = "owner_address"
	AttributeKeyOperatorAddress   = "operator_address"
	AttributeKeyReporterAddress   = "reporter_address"
	AttributeKeyStakeAmount       = "stake_amount"
	AttributeKeyStakeDenom        = "stake_denom"
	AttributeKeySponsorshipAmount = "sponsorship_amount"
	AttributeKeyLocales           = "locales"
	AttributeKeyLocaleCount       = "locale_count"
	AttributeKeyModelHash         = "model_hash"
	AttributeKeyApiKeyFingerprint = "api_key_fingerprint"
	AttributeKeyLastHeartbeat     = "last_heartbeat"
	AttributeKeyNextSlashTime     = "next_slash_time"
	AttributeKeyHeartbeatLatency  = "heartbeat_latency_seconds"
	AttributeKeyHeartbeatFailures = "heartbeat_failures"
	AttributeKeySlashFraction     = "slash_fraction"
	AttributeKeySlashedAmount     = "slashed_amount"
	AttributeKeySlashCount        = "slash_count"
	AttributeKeyMisbehaviorReports = "misbehavior_reports"
	AttributeKeyMisbehaviorReason = "misbehavior_reason"
	AttributeKeyOldStatus         = "old_status"
	AttributeKeyNewStatus         = "new_status"
	AttributeKeyRemainingStake    = "remaining_stake"
	AttributeKeyBlockHeight       = "block_height"
	AttributeKeyBlockTime         = "block_time"
	AttributeKeyMinStake          = "min_stake"
	AttributeKeyMaxLocales        = "max_locales"
	AttributeKeyHeartbeatWindow   = "heartbeat_window_seconds"
	AttributeKeyHeartbeatGrace    = "heartbeat_grace_seconds"
)

// Helper functions to create event attributes

// NewRegisterAssistantEvent creates attributes for assistant registration
func NewRegisterAssistantEvent(
	assistantAddr, ownerAddr, modelHash, apiKeyFingerprint string,
	stakeAmount, stakeDenom, sponsorshipAmount string,
	locales []string,
	blockHeight int64,
	blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAssistantAddress:  assistantAddr,
		AttributeKeyOwnerAddress:      ownerAddr,
		AttributeKeyModelHash:         modelHash,
		AttributeKeyApiKeyFingerprint: apiKeyFingerprint,
		AttributeKeyStakeAmount:       stakeAmount,
		AttributeKeyStakeDenom:        stakeDenom,
		AttributeKeySponsorshipAmount: sponsorshipAmount,
		AttributeKeyLocales:           formatLocales(locales),
		AttributeKeyLocaleCount:       formatInt(len(locales)),
		AttributeKeyBlockHeight:       formatInt64(blockHeight),
		AttributeKeyBlockTime:         blockTime,
	}
}

// NewUpdateLocalesEvent creates attributes for locale update
func NewUpdateLocalesEvent(
	assistantAddr, ownerAddr string,
	oldLocales, newLocales []string,
	blockHeight int64,
	blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAssistantAddress: assistantAddr,
		AttributeKeyOwnerAddress:     ownerAddr,
		"old_locales":                formatLocales(oldLocales),
		"new_locales":                formatLocales(newLocales),
		AttributeKeyLocaleCount:      formatInt(len(newLocales)),
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewHeartbeatEvent creates attributes for heartbeat
func NewHeartbeatEvent(
	assistantAddr, operatorAddr string,
	latencySeconds float64,
	nextSlashTime int64,
	blockHeight int64,
	blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAssistantAddress: assistantAddr,
		AttributeKeyOperatorAddress:  operatorAddr,
		AttributeKeyHeartbeatLatency: formatFloat64(latencySeconds),
		AttributeKeyNextSlashTime:    formatInt64(nextSlashTime),
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewHeartbeatFailureEvent creates attributes for heartbeat failure
func NewHeartbeatFailureEvent(
	assistantAddr string,
	failureCount uint64,
	slashedAmount, slashFraction string,
	blockHeight int64,
	blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAssistantAddress:  assistantAddr,
		AttributeKeyHeartbeatFailures: formatUint64(failureCount),
		AttributeKeySlashedAmount:     slashedAmount,
		AttributeKeySlashFraction:     slashFraction,
		AttributeKeyBlockHeight:       formatInt64(blockHeight),
		AttributeKeyBlockTime:         blockTime,
	}
}

// NewReportMisbehaviorEvent creates attributes for misbehavior report
func NewReportMisbehaviorEvent(
	assistantAddr, reporterAddr, reason string,
	misbehaviorReports uint64,
	slashedAmount, slashFraction string,
	newStatus string,
	blockHeight int64,
	blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAssistantAddress:   assistantAddr,
		AttributeKeyReporterAddress:    reporterAddr,
		AttributeKeyMisbehaviorReason:  reason,
		AttributeKeyMisbehaviorReports: formatUint64(misbehaviorReports),
		AttributeKeySlashedAmount:      slashedAmount,
		AttributeKeySlashFraction:      slashFraction,
		AttributeKeyNewStatus:          newStatus,
		AttributeKeyBlockHeight:        formatInt64(blockHeight),
		AttributeKeyBlockTime:          blockTime,
	}
}

// NewSlashAssistantEvent creates attributes for slashing
func NewSlashAssistantEvent(
	assistantAddr string,
	slashFraction, slashedAmount, remainingStake string,
	slashCount uint64,
	blockHeight int64,
	blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAssistantAddress: assistantAddr,
		AttributeKeySlashFraction:    slashFraction,
		AttributeKeySlashedAmount:    slashedAmount,
		AttributeKeyRemainingStake:   remainingStake,
		AttributeKeySlashCount:       formatUint64(slashCount),
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewStatusChangeEvent creates attributes for status change
func NewStatusChangeEvent(
	assistantAddr, oldStatus, newStatus string,
	blockHeight int64,
	blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyAssistantAddress: assistantAddr,
		AttributeKeyOldStatus:        oldStatus,
		AttributeKeyNewStatus:        newStatus,
		AttributeKeyBlockHeight:      formatInt64(blockHeight),
		AttributeKeyBlockTime:        blockTime,
	}
}

// NewParamsUpdateEvent creates attributes for params update
func NewParamsUpdateEvent(
	minStake, maxLocales string,
	heartbeatWindow, heartbeatGrace string,
	slashFractionDowntime, slashFractionMisbehavior string,
	blockHeight int64,
	blockTime string,
) map[string]string {
	return map[string]string{
		AttributeKeyMinStake:          minStake,
		AttributeKeyMaxLocales:        maxLocales,
		AttributeKeyHeartbeatWindow:   heartbeatWindow,
		AttributeKeyHeartbeatGrace:    heartbeatGrace,
		"slash_fraction_downtime":     slashFractionDowntime,
		"slash_fraction_misbehavior":  slashFractionMisbehavior,
		AttributeKeyBlockHeight:       formatInt64(blockHeight),
		AttributeKeyBlockTime:         blockTime,
	}
}

// Helper formatting functions

func formatInt(i int) string {
	return formatInt64(int64(i))
}

func formatInt64(i int64) string {
	return fmt.Sprintf("%d", i)
}

func formatUint64(u uint64) string {
	return fmt.Sprintf("%d", u)
}

func formatFloat64(f float64) string {
	return fmt.Sprintf("%.6f", f)
}

func formatLocales(locales []string) string {
	result := ""
	for i, loc := range locales {
		if i > 0 {
			result += ","
		}
		result += loc
	}
	return result
}
