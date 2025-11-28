package types

import "errors"

// Module-specific errors for the inclusionroutines module
var (
	// IR management errors
	ErrIRNotFound          = errors.New("inclusion routine not found")
	ErrIRAlreadyExists     = errors.New("inclusion routine already exists")
	ErrIRInvalidID         = errors.New("invalid inclusion routine id")
	ErrIRInvalidStatus     = errors.New("invalid inclusion routine status")
	ErrIRInvalidScore      = errors.New("invalid score: must be non-negative")
	ErrIRInvalidReward     = errors.New("invalid poi reward: must be non-negative")
	ErrIRSuspended         = errors.New("inclusion routine is suspended")
	ErrIRRetired           = errors.New("inclusion routine is retired")
	ErrIRNotActive         = errors.New("inclusion routine is not active")
	ErrInvalidIRDefinition = errors.New("invalid inclusion routine definition")
	ErrIRSunset            = errors.New("inclusion routine has reached sunset height")

	// Prerequisite errors
	ErrPrerequisiteNotMet   = errors.New("prerequisite not met")
	ErrCircularDependency   = errors.New("circular dependency detected")
	ErrInvalidPrerequisite  = errors.New("invalid prerequisite")
	ErrPrerequisiteNotFound = errors.New("prerequisite not found")
	ErrSelfPrerequisite     = errors.New("ir cannot be its own prerequisite")

	// Rate limit errors
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrInvalidRateLimit  = errors.New("invalid rate limit configuration")
	ErrRateLimitNotFound = errors.New("rate limit configuration not found")

	// Authorization errors
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidAuthority = errors.New("invalid authority address")

	// Validation errors
	ErrInvalidName            = errors.New("invalid name: cannot be empty")
	ErrInvalidDescription     = errors.New("invalid description: cannot be empty")
	ErrInvalidArena           = errors.New("invalid arena")
	ErrInvalidPrivacyTier     = errors.New("invalid privacy tier")
	ErrInvalidVersion         = errors.New("invalid version")
	ErrInvalidMetadataHash    = errors.New("invalid metadata hash")
	ErrInvalidHeight          = errors.New("invalid height")
	ErrSunsetBeforeActivation = errors.New("sunset height must be after activation height")
	ErrEmptyLocaleTag         = errors.New("locale tags cannot be empty")
)
