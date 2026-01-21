// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

//go:build !testnet

package config

// Bech32 prefixes for mainnet
const (
	Bech32MainPrefix      = "aura"
	Bech32ValidatorPrefix = "auravaloper"
	Bech32ConsensusPrefix = "auravalcons"
)
