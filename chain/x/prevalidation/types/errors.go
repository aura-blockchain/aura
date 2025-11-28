package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Prevalidation module errors
var (
	ErrInvalidInput                = errorsmod.Register("prevalidation", 1, "invalid input")
	ErrInvalidTransaction          = errorsmod.Register("prevalidation", 2, "invalid transaction")
	ErrTransactionNotFound         = errorsmod.Register("prevalidation", 3, "transaction not found")
	ErrTransactionExpired          = errorsmod.Register("prevalidation", 4, "transaction expired")
	ErrInvalidTemplate             = errorsmod.Register("prevalidation", 5, "invalid validation template")
	ErrTemplateNotFound            = errorsmod.Register("prevalidation", 6, "validation template not found")
	ErrInvalidTransactionType      = errorsmod.Register("prevalidation", 7, "invalid transaction type")
	ErrValidationFailed            = errorsmod.Register("prevalidation", 8, "validation failed")
	ErrCacheNotFound               = errorsmod.Register("prevalidation", 9, "cache entry not found")
	ErrInvalidCacheStrategy        = errorsmod.Register("prevalidation", 10, "invalid cache strategy")
	ErrSchedulerDisabled           = errorsmod.Register("prevalidation", 11, "scheduler is disabled")
	ErrAutoScalingDisabled         = errorsmod.Register("prevalidation", 12, "auto-scaling is disabled")
	ErrInvalidMetrics              = errorsmod.Register("prevalidation", 13, "invalid metrics")
	ErrInsufficientConfidenceScore = errorsmod.Register("prevalidation", 14, "insufficient confidence score")
	ErrMaxValidationAttempts       = errorsmod.Register("prevalidation", 15, "maximum validation attempts exceeded")
	ErrInvalidStatus               = errorsmod.Register("prevalidation", 16, "invalid transaction status")
	ErrUnauthorized                = errorsmod.Register("prevalidation", 17, "unauthorized operation")
)
