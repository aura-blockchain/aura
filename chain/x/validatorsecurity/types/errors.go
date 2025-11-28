package types

import (
	"cosmossdk.io/errors"
)

var (
	ErrInvalidValidator           = errors.Register(ModuleName, 1, "invalid validator")
	ErrValidatorNotFound          = errors.Register(ModuleName, 2, "validator not found")
	ErrValidatorAlreadyRegistered = errors.Register(ModuleName, 3, "validator already registered")
	ErrValidatorJailed            = errors.Register(ModuleName, 4, "validator is jailed")
	ErrValidatorTombstoned        = errors.Register(ModuleName, 5, "validator is tombstoned")
	ErrInvalidDoubleSignEvidence  = errors.Register(ModuleName, 6, "invalid double sign evidence")
	ErrInsufficientStake          = errors.Register(ModuleName, 7, "insufficient stake amount")
	ErrInvalidSentryNode          = errors.Register(ModuleName, 8, "invalid sentry node")
	ErrSentryNodeNotFound         = errors.Register(ModuleName, 9, "sentry node not found")
	ErrInsufficientSentryNodes    = errors.Register(ModuleName, 10, "insufficient sentry nodes")
	ErrInvalidGeographicLocation  = errors.Register(ModuleName, 11, "invalid geographic location")
	ErrRegionCapacityExceeded     = errors.Register(ModuleName, 12, "region capacity exceeded")
	ErrInvalidKeys                = errors.Register(ModuleName, 13, "invalid hot/cold keys")
	ErrKeysNotSeparated           = errors.Register(ModuleName, 14, "keys not separated")
	ErrAlertNotFound              = errors.Register(ModuleName, 15, "alert not found")
	ErrInvalidAlert               = errors.Register(ModuleName, 16, "invalid alert")
	ErrCannotUnjail               = errors.Register(ModuleName, 17, "cannot unjail validator")
	ErrDowntimeViolation          = errors.Register(ModuleName, 18, "downtime violation detected")
	ErrFailoverFailed             = errors.Register(ModuleName, 19, "failover operation failed")
	ErrNoBackupValidators         = errors.Register(ModuleName, 20, "no backup validators available")
	ErrInvalidBackupValidator     = errors.Register(ModuleName, 21, "invalid backup validator")
	ErrEvidenceNotFound           = errors.Register(ModuleName, 22, "double sign evidence not found")
	ErrInfractionNotFound         = errors.Register(ModuleName, 23, "downtime infraction not found")
)
