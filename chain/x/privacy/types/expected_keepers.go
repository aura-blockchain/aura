// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper interface for privacy operations
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected bank keeper interface for privacy operations
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// ZKProofSystem defines the interface for zero-knowledge proof operations
// This interface provides cryptographic primitives for privacy features
type ZKProofSystem interface {
	// VerifyProof verifies a zero-knowledge proof
	VerifyProof(proof []byte, publicInputs []byte, verificationKey []byte) (bool, error)

	// GenerateProof generates a zero-knowledge proof (for testing/simulation)
	GenerateProof(witness []byte, publicInputs []byte, provingKey []byte) ([]byte, error)

	// VerifyRangeProof verifies a range proof for confidential transactions
	VerifyRangeProof(commitment []byte, proof []byte, min, max uint64) (bool, error)
}

// MixingService defines the interface for coin mixing operations
// Provides privacy through transaction mixing and anonymity sets
type MixingService interface {
	// CreateMixingPool creates a new mixing pool
	CreateMixingPool(denomination string, minParticipants uint32) (string, error)

	// JoinMixingPool adds a participant to a mixing pool
	JoinMixingPool(poolID string, participant string, commitment []byte) error

	// ExecuteMixing executes the mixing process for a pool
	ExecuteMixing(poolID string) error

	// GetPoolStatus returns the status of a mixing pool
	GetPoolStatus(poolID string) (string, error)
}

// ViewKeyManager defines the interface for managing view keys in privacy transactions
// View keys allow selective disclosure of transaction details
type ViewKeyManager interface {
	// GenerateViewKey generates a view key for an address
	GenerateViewKey(address string) ([]byte, error)

	// StoreViewKey stores a view key
	StoreViewKey(address string, viewKey []byte) error

	// GetViewKey retrieves a view key for an address
	GetViewKey(address string) ([]byte, bool)

	// DecryptWithViewKey decrypts transaction data using a view key
	DecryptWithViewKey(encryptedData []byte, viewKey []byte) ([]byte, error)
}

// NetworkPrivacy defines the interface for network-level privacy features
// Includes features like Tor integration, mixnet routing, etc.
type NetworkPrivacy interface {
	// ObfuscateTransaction obfuscates network-level transaction metadata
	ObfuscateTransaction(txData []byte) ([]byte, error)

	// RoutePrivately routes a transaction through a privacy network
	RoutePrivately(destination string, data []byte) error

	// GetPrivacyMetrics returns network privacy metrics
	GetPrivacyMetrics() map[string]interface{}
}

// MemoEncryptor defines the interface for encrypting transaction memos
type MemoEncryptor interface {
	// EncryptMemo encrypts a transaction memo for a recipient
	EncryptMemo(memo string, recipientPubKey []byte) ([]byte, error)

	// DecryptMemo decrypts a transaction memo
	DecryptMemo(encryptedMemo []byte, privateKey []byte) (string, error)

	// VerifyEncryptedMemo verifies the integrity of an encrypted memo
	VerifyEncryptedMemo(encryptedMemo []byte) bool
}
