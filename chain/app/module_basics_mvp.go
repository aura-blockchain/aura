//go:build mvp

// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/types/module"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	consensus "github.com/cosmos/cosmos-sdk/x/consensus"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	params "github.com/cosmos/cosmos-sdk/x/params"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing"
	staking "github.com/cosmos/cosmos-sdk/x/staking"

	// MVP Aura modules
	compliance "github.com/aequitas/aura/chain/x/compliance"
	"github.com/aequitas/aura/chain/x/dataregistry"
	"github.com/aequitas/aura/chain/x/governance"
	"github.com/aequitas/aura/chain/x/identity"
	"github.com/aequitas/aura/chain/x/prevalidation"
	"github.com/aequitas/aura/chain/x/vcregistry"
	aurawasm "github.com/aequitas/aura/chain/x/wasm"
)

// ModuleBasicsMVP exposes only the MVP modules for genesis CLI commands.
// This includes:
// - Cosmos SDK core modules (auth, bank, staking, slashing, distribution, params, consensus, genutil)
// - CosmWasm base module
// - MVP AURA modules (compliance, identity, vcregistry, dataregistry, governance, prevalidation, wasm)
var ModuleBasicsMVP = module.NewBasicManager(
	// Cosmos SDK core modules
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	staking.AppModuleBasic{},
	slashing.AppModuleBasic{},
	distribution.AppModuleBasic{},
	params.AppModuleBasic{},
	consensus.AppModuleBasic{},
	genutil.AppModuleBasic{},

	// CosmWasm base module
	wasm.AppModuleBasic{},

	// MVP AURA modules
	compliance.AppModuleBasic{},
	identity.AppModuleBasic{},
	vcregistry.AppModuleBasic{},
	dataregistry.AppModuleBasic{},
	governance.AppModuleBasic{},
	prevalidation.AppModuleBasic{},
	aurawasm.AppModuleBasic{},
)
