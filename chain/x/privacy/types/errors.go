package types

import (
	"errors"
)

var (
	ErrInvalidRingSize        = errors.New("invalid ring size")
	ErrInvalidMixingParams    = errors.New("invalid mixing parameters")
	ErrInvalidCommitment      = errors.New("invalid commitment")
	ErrInvalidProof           = errors.New("invalid proof")
	ErrNullifierExists        = errors.New("nullifier already exists")
	ErrInvalidNullifier       = errors.New("invalid nullifier")
	ErrKeyImageAlreadyUsed    = errors.New("key image already used")
)
