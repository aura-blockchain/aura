package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrAssistantExists      = sdkerrors.Register(ModuleName, 1, "assistant already registered")
	ErrAssistantNotFound    = sdkerrors.Register(ModuleName, 2, "assistant not found")
	ErrInvalidLocale        = sdkerrors.Register(ModuleName, 3, "invalid locale")
	ErrUnauthorizedOperator = sdkerrors.Register(ModuleName, 4, "unauthorized operator")
	ErrHeartbeatExpired     = sdkerrors.Register(ModuleName, 5, "heartbeat missed the safety window")
	ErrInsufficientStake    = sdkerrors.Register(ModuleName, 6, "insufficient stake")
	ErrLocaleCapacity       = sdkerrors.Register(ModuleName, 7, "locale list exceeds capacity")
	ErrInvalidParams        = sdkerrors.Register(ModuleName, 8, "invalid ai-assistant params")
)
