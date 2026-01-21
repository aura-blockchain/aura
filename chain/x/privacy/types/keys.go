// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// ModuleName is the name of the privacy module
	ModuleName = "privacy"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the privacy module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the privacy module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	ParamsKey              = []byte{0x01}
	CommitmentPrefix       = []byte{0x02}
	NullifierPrefix        = []byte{0x03}
	MerkleTreePrefix       = []byte{0x04}
	ZKProofPrefix          = []byte{0x05}
	ShieldedTxPrefix       = []byte{0x06}
	MixingPoolPrefix       = []byte{0x07}
	MixingPoolKeyPrefix    = []byte{0x07} // Alias for compatibility
	ViewKeyPrefix          = []byte{0x08}
	EncryptionKeyKeyPrefix = []byte{0x09}
	RingSignatureKeyPrefix = []byte{0x0a}
	KeyImagePrefix         = []byte{0x0b}
	RingMemberPrefix       = []byte{0x0c}
	SpentCommitmentPrefix  = []byte{0x0d}
	ConfidentialTxPrefix   = []byte{0x0e}
)
