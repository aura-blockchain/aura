package types

import "time"

// Stub types until protobuf definitions are created
type ValidatorSecurityInfo struct {
	IsTombstoned bool
	IsJailed     bool
	JailedUntil  time.Time
}

type ValidatorSecurityParams struct{}
type ValidatorAlert struct{}
type SentryNodeInfo struct{}
type DoubleSignEvidence struct{}
type DowntimeInfraction struct{}
