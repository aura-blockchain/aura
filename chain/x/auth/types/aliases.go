// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"

// Re-export frequently used protobuf types for convenience.
type (
	Role                 = authproto.Role
	RoleAssignment       = authproto.RoleAssignment
	MultisigWallet       = authproto.MultisigWallet
	MultisigProposal     = authproto.MultisigProposal
	TimeLockedAction     = authproto.TimeLockedAction
	EmergencyAdmin       = authproto.EmergencyAdmin
	ValidatorKeyRotation = authproto.ValidatorKeyRotation
	Session              = authproto.Session
	RateLimitConfig      = authproto.RateLimitConfig
	AuditLog             = authproto.AuditLog
)
