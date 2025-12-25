// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

// Package auth provides comprehensive authentication and authorization functionality for the AURA blockchain.
//
// The auth module implements role-based access control (RBAC), multi-signature wallets, time-locked
// actions, emergency admin capabilities, validator key rotation, session management, rate limiting,
// and comprehensive audit logging.
//
// # Overview
//
// The auth module is a critical security component that manages:
//   - Role definitions and assignments with expiration
//   - Multi-signature wallet creation and proposal management
//   - Time-locked actions for governance and security
//   - Emergency admin privileges for crisis response
//   - Validator key rotation for enhanced security
//   - User session management with device tracking
//   - Rate limiting to prevent abuse
//   - Comprehensive audit logs for compliance
//
// # Architecture
//
// The module follows the standard Cosmos SDK module pattern with:
//   - Keeper: Core business logic and state management
//   - Types: Message and query type definitions
//   - Client: CLI and gRPC query interfaces
//   - Proto: Protocol buffer definitions
//
// # Key Components
//
// ## Roles and Permissions
//
// The module provides a flexible RBAC system where roles can be created with
// specific permissions. Default roles include:
//   - Admin: Full system access
//   - Moderator: User management and moderation
//   - Validator: Validator-specific operations
//   - User: Basic user permissions
//
// Example usage:
//
//	// Create a custom role
//	role := &authproto.Role{
//	    Name: "auditor",
//	    Permissions: []string{types.PermissionViewAuditLogs},
//	    Description: "Read-only audit access",
//	}
//	keeper.SetRole(ctx, role)
//
//	// Assign role to user with expiration
//	assignment := &authproto.RoleAssignment{
//	    Address: userAddress,
//	    RoleName: "auditor",
//	    ExpiresAt: timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
//	}
//	keeper.SetRoleAssignment(ctx, assignment)
//
// ## Multi-Signature Wallets
//
// Multi-signature wallets require M-of-N signatures for transactions:
//
//	wallet := &authproto.MultisigWallet{
//	    Id: "treasury-wallet",
//	    Signers: []string{addr1, addr2, addr3},
//	    Threshold: 2, // Requires 2 of 3 signatures
//	}
//	keeper.SetMultisigWallet(ctx, wallet)
//
// ## Time-Locked Actions
//
// Critical actions can be time-locked to allow for review and cancellation:
//
//	action := &authproto.TimeLockedAction{
//	    Id: "upgrade-proposal",
//	    Action: "upgrade-chain",
//	    ExecuteAt: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
//	    Status: authproto.ActionStatus_ACTION_PENDING,
//	}
//	keeper.SetTimeLockedAction(ctx, action)
//
// ## Session Management
//
// User sessions track authentication state across devices:
//
//	session := &authproto.Session{
//	    SessionId: uuid.New().String(),
//	    UserAddress: userAddress,
//	    DeviceInfo: "Mozilla/5.0...",
//	    IpAddress: "192.168.1.1",
//	    ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
//	}
//	keeper.SetSession(ctx, session)
//
// ## Rate Limiting
//
// Rate limiting prevents abuse at per-minute, per-hour, and per-day levels:
//
//	config := &authproto.RateLimitConfig{
//	    UserAddress: userAddress,
//	    MaxPerMinute: 60,
//	    MaxPerHour: 1000,
//	    MaxPerDay: 10000,
//	}
//	keeper.SetRateLimitConfig(ctx, config)
//
// ## Audit Logging
//
// All security-relevant actions are logged for compliance:
//
//	keeper.LogAudit(ctx, userAddress, "role_assigned", "auditor", "success", nil, "")
//
// # Security Considerations
//
// The auth module implements multiple security layers:
//   - Input validation on all parameters
//   - Permission checks before state changes
//   - Audit logging for all sensitive operations
//   - Automatic cleanup of expired data
//   - Rate limiting to prevent abuse
//   - Time-based role expiration
//
// # Performance Optimizations
//
// The module includes several performance optimizations:
//   - Indexed lookups for roles and assignments
//   - Efficient iteration with prefix iterators
//   - Automatic cleanup of expired sessions (max 10000 audit logs)
//   - Batched role assignment queries
//
// # Integration
//
// To integrate the auth module in your application:
//
//	import (
//	    authkeeper "github.com/aequitas/aura/chain/x/auth/keeper"
//	    authtypes "github.com/aequitas/aura/chain/x/auth/types"
//	)
//
//	// In app.go
//	app.AuthKeeper = authkeeper.NewKeeper(
//	    appCodec,
//	    keys[authtypes.StoreKey],
//	)
//
// # State Structure
//
// The module stores data with the following key prefixes:
//   - 0x01: Roles
//   - 0x02: Role assignments
//   - 0x03: Multisig wallets
//   - 0x04: Multisig proposals
//   - 0x05: Time-locked actions
//   - 0x06: Emergency admins
//   - 0x07: Validator rotations
//   - 0x08: Sessions
//   - 0x09: User session index
//   - 0x0A: Rate limits
//   - 0x0B: Audit logs
//   - 0x0C: Module parameters
//   - 0x0D: Audit log counter
//
// # Events
//
// The module emits events for important operations:
//   - role_created
//   - role_assigned
//   - role_revoked
//   - multisig_created
//   - proposal_created
//   - proposal_signed
//   - session_created
//   - session_invalidated
//
// # Queries
//
// Available queries:
//   - Roles: Get all roles or specific role
//   - RoleAssignments: Get assignments for an address
//   - MultisigWallets: Query multisig wallet configuration
//   - Sessions: List active sessions
//   - AuditLogs: Query audit trail
//
// # Transactions
//
// Available transactions:
//   - CreateRole: Define a new role
//   - AssignRole: Grant role to address
//   - RevokeRole: Remove role from address
//   - CreateMultisigWallet: Setup multi-sig wallet
//   - CreateProposal: Propose multisig action
//   - SignProposal: Sign a multisig proposal
//   - CreateTimeLockedAction: Schedule delayed action
//   - RotateValidatorKey: Update validator key
//
// # Module Parameters
//
// Configurable parameters:
//   - MaxSessionDuration: Maximum session lifetime
//   - DefaultRateLimit: Default rate limit values
//   - AuditRetentionPeriod: How long to keep audit logs
//   - EmergencyAdminTimeout: Emergency admin privilege duration
//
// # Compliance Features
//
// The module supports regulatory compliance with:
//   - Comprehensive audit trails
//   - Role-based access control
//   - Time-based access expiration
//   - Multi-signature approvals
//   - Emergency response capabilities
//
// For more information, see the module documentation at:
// https://github.com/aequitas/aura/tree/main/docs/modules/auth
package auth
