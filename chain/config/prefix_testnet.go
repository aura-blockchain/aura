// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

//go:build testnet

package config

// Bech32 prefixes for testnet - uses distinct prefix to prevent key reuse
const (
	Bech32MainPrefix      = "auratest"
	Bech32ValidatorPrefix = "auratestvaloper"
	Bech32ConsensusPrefix = "auratestvalcons"
)
