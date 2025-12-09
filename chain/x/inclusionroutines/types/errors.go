package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Inclusion routines module error codes
var (
	// IR management errors (1-19)
	ErrIRNotFound          = errorsmod.Register(ModuleName, 1, "inclusion routine not found")
	ErrIRAlreadyExists     = errorsmod.Register(ModuleName, 2, "inclusion routine already exists")
	ErrIRInvalidID         = errorsmod.Register(ModuleName, 3, "invalid inclusion routine id")
	ErrIRInvalidStatus     = errorsmod.Register(ModuleName, 4, "invalid inclusion routine status")
	ErrIRInvalidScore      = errorsmod.Register(ModuleName, 5, "invalid score: must be non-negative")
	ErrIRInvalidReward     = errorsmod.Register(ModuleName, 6, "invalid poi reward: must be non-negative")
	ErrIRSuspended         = errorsmod.Register(ModuleName, 7, "inclusion routine is suspended")
	ErrIRRetired           = errorsmod.Register(ModuleName, 8, "inclusion routine is retired")
	ErrIRNotActive         = errorsmod.Register(ModuleName, 9, "inclusion routine is not active")
	ErrInvalidIRDefinition = errorsmod.Register(ModuleName, 10, "invalid inclusion routine definition")
	ErrIRSunset            = errorsmod.Register(ModuleName, 11, "inclusion routine has reached sunset height")

	// Prerequisite errors (20-29)
	ErrPrerequisiteNotMet   = errorsmod.Register(ModuleName, 20, "prerequisite not met")
	ErrCircularDependency   = errorsmod.Register(ModuleName, 21, "circular dependency detected")
	ErrInvalidPrerequisite  = errorsmod.Register(ModuleName, 22, "invalid prerequisite")
	ErrPrerequisiteNotFound = errorsmod.Register(ModuleName, 23, "prerequisite not found")
	ErrSelfPrerequisite     = errorsmod.Register(ModuleName, 24, "ir cannot be its own prerequisite")

	// Rate limit errors (30-39)
	ErrRateLimitExceeded = errorsmod.Register(ModuleName, 30, "rate limit exceeded")
	ErrInvalidRateLimit  = errorsmod.Register(ModuleName, 31, "invalid rate limit configuration")
	ErrRateLimitNotFound = errorsmod.Register(ModuleName, 32, "rate limit configuration not found")

	// Authorization errors (40-49)
	ErrUnauthorized     = errorsmod.Register(ModuleName, 40, "unauthorized")
	ErrInvalidAuthority = errorsmod.Register(ModuleName, 41, "invalid authority address")

	// Validation errors (50-69)
	ErrInvalidName            = errorsmod.Register(ModuleName, 50, "invalid name: cannot be empty")
	ErrInvalidDescription     = errorsmod.Register(ModuleName, 51, "invalid description: cannot be empty")
	ErrInvalidArena           = errorsmod.Register(ModuleName, 52, "invalid arena")
	ErrInvalidPrivacyTier     = errorsmod.Register(ModuleName, 53, "invalid privacy tier")
	ErrInvalidVersion         = errorsmod.Register(ModuleName, 54, "invalid version")
	ErrInvalidMetadataHash    = errorsmod.Register(ModuleName, 55, "invalid metadata hash")
	ErrInvalidHeight          = errorsmod.Register(ModuleName, 56, "invalid height")
	ErrSunsetBeforeActivation = errorsmod.Register(ModuleName, 57, "sunset height must be after activation height")
	ErrEmptyLocaleTag         = errorsmod.Register(ModuleName, 58, "locale tags cannot be empty")
)
