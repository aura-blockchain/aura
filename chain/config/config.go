// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

// Package config provides SDK configuration for the AURA blockchain.
// This package is separate from app to avoid import cycles when test
// utilities need to initialize SDK config.
package config

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var sdkConfigOnce sync.Once

// EnsureSDKConfig sets the Bech32 prefixes once for the process.
// Prefixes are defined in prefix.go (mainnet: aura) and prefix_testnet.go (testnet: auratest)
// This function is safe to call multiple times; only the first call has effect.
func EnsureSDKConfig() {
	sdkConfigOnce.Do(func() {
		cfg := sdk.GetConfig()
		cfg.SetBech32PrefixForAccount(Bech32MainPrefix, Bech32MainPrefix+"pub")
		cfg.SetBech32PrefixForValidator(Bech32ValidatorPrefix, Bech32ValidatorPrefix+"pub")
		cfg.SetBech32PrefixForConsensusNode(Bech32ConsensusPrefix, Bech32ConsensusPrefix+"pub")
		cfg.Seal()
	})
}
