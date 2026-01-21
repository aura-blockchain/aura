// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import pb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"

// Re-export all proto types
type (
	// Enums
	TransactionType  = pb.TransactionType
	ValidationStatus = pb.ValidationStatus
	CacheStrategy    = pb.CacheStrategy

	// Core types
	PreValidatedTransaction = pb.PreValidatedTransaction
	ValidationTemplate      = pb.ValidationTemplate
	ValidationMetadata      = pb.ValidationMetadata
	PreValidationMetrics    = pb.PreValidationMetrics
	TypeMetrics             = pb.TypeMetrics
	HourlyMetrics           = pb.HourlyMetrics
	TemplateStats           = pb.TemplateStats
	ControlGroupMetrics     = pb.ControlGroupMetrics
	SchedulerConfig         = pb.SchedulerConfig
	AutoScalingConfig       = pb.AutoScalingConfig
	Params                  = pb.Params

	// Genesis types
	GenesisState = pb.GenesisState

	// Event types
	EventPreValidationCreated  = pb.EventPreValidationCreated
	EventPreValidationExecuted = pb.EventPreValidationExecuted
	EventPreValidationExpired  = pb.EventPreValidationExpired
	EventSchedulerRun          = pb.EventSchedulerRun
	EventAutoScaling           = pb.EventAutoScaling
	EventCacheHit              = pb.EventCacheHit
	EventCacheMiss             = pb.EventCacheMiss
	EventMetricsUpdate         = pb.EventMetricsUpdate
)

// Transaction represents a transaction for pre-validation
type Transaction struct {
	Sender    string
	Recipient string
	Amount    string
	Nonce     uint64
	Data      []byte
	GasPrice  string
	Signature []byte
}

// ValidationResult represents the result of transaction validation
type ValidationResult struct {
	TxHash      string
	Valid       bool
	GasEstimate uint64
	Error       string
}

// Re-export enum values for TransactionType
const (
	TransactionType_TX_TYPE_UNSPECIFIED             = pb.TransactionType_TX_TYPE_UNSPECIFIED
	TransactionType_TX_TYPE_IR_COMPLETION           = pb.TransactionType_TX_TYPE_IR_COMPLETION
	TransactionType_TX_TYPE_DEX_SWAP                = pb.TransactionType_TX_TYPE_DEX_SWAP
	TransactionType_TX_TYPE_LP_DEPOSIT              = pb.TransactionType_TX_TYPE_LP_DEPOSIT
	TransactionType_TX_TYPE_LP_WITHDRAWAL           = pb.TransactionType_TX_TYPE_LP_WITHDRAWAL
	TransactionType_TX_TYPE_VC_MINT                 = pb.TransactionType_TX_TYPE_VC_MINT
	TransactionType_TX_TYPE_BRIDGE_TRANSFER         = pb.TransactionType_TX_TYPE_BRIDGE_TRANSFER
	TransactionType_TX_TYPE_CONFIDENCE_SCORE_UPDATE = pb.TransactionType_TX_TYPE_CONFIDENCE_SCORE_UPDATE
	TransactionType_TX_TYPE_IDENTITY_CHANGE         = pb.TransactionType_TX_TYPE_IDENTITY_CHANGE
)

// Re-export enum values for ValidationStatus
const (
	ValidationStatus_VALIDATION_STATUS_UNSPECIFIED = pb.ValidationStatus_VALIDATION_STATUS_UNSPECIFIED
	ValidationStatus_VALIDATION_STATUS_PENDING     = pb.ValidationStatus_VALIDATION_STATUS_PENDING
	ValidationStatus_VALIDATION_STATUS_VALIDATED   = pb.ValidationStatus_VALIDATION_STATUS_VALIDATED
	ValidationStatus_VALIDATION_STATUS_EXPIRED     = pb.ValidationStatus_VALIDATION_STATUS_EXPIRED
	ValidationStatus_VALIDATION_STATUS_EXECUTED    = pb.ValidationStatus_VALIDATION_STATUS_EXECUTED
	ValidationStatus_VALIDATION_STATUS_FAILED      = pb.ValidationStatus_VALIDATION_STATUS_FAILED
)

// Re-export enum values for CacheStrategy
const (
	CacheStrategy_CACHE_STRATEGY_UNSPECIFIED = pb.CacheStrategy_CACHE_STRATEGY_UNSPECIFIED
	CacheStrategy_CACHE_STRATEGY_LRU         = pb.CacheStrategy_CACHE_STRATEGY_LRU
	CacheStrategy_CACHE_STRATEGY_LFU         = pb.CacheStrategy_CACHE_STRATEGY_LFU
	CacheStrategy_CACHE_STRATEGY_FIFO        = pb.CacheStrategy_CACHE_STRATEGY_FIFO
	CacheStrategy_CACHE_STRATEGY_ADAPTIVE    = pb.CacheStrategy_CACHE_STRATEGY_ADAPTIVE
)
