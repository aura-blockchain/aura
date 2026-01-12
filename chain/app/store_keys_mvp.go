//go:build mvp

// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	compliancetypes "github.com/aequitas/aura/chain/x/compliance/types"
	drtypes "github.com/aequitas/aura/chain/x/dataregistry/types"
	governancetypes "github.com/aequitas/aura/chain/x/governance/types"
	identitytypes "github.com/aequitas/aura/chain/x/identity/types"
	prevalidationtypes "github.com/aequitas/aura/chain/x/prevalidation/types"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	wasmSecurityTypes "github.com/aequitas/aura/chain/x/wasm/types"
)

// storeKeysMVP holds KV store keys for MVP modules only.
// This reduces memory footprint and state complexity for MVP release.
type storeKeysMVP struct {
	// Cosmos SDK standard keys
	account      *storetypes.KVStoreKey
	bank         *storetypes.KVStoreKey
	staking      *storetypes.KVStoreKey
	slashing     *storetypes.KVStoreKey
	distribution *storetypes.KVStoreKey
	params       *storetypes.KVStoreKey
	consensus    *storetypes.KVStoreKey
	upgrade      *storetypes.KVStoreKey

	// MVP AURA modules
	identity      *storetypes.KVStoreKey
	vc            *storetypes.KVStoreKey
	dataRegistry  *storetypes.KVStoreKey
	compliance    *storetypes.KVStoreKey
	prevalidation *storetypes.KVStoreKey
	governance    *storetypes.KVStoreKey
	wasm          *storetypes.KVStoreKey
	wasmSecurity  *storetypes.KVStoreKey
}

// NamesMVP returns all store key names for MVP build.
// This is the single source of truth for MVP store key names.
func (s *storeKeysMVP) Names() []string {
	return []string{
		// Cosmos SDK standard keys
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		slashingtypes.StoreKey,
		distrtypes.StoreKey,
		paramstypes.StoreKey,
		consensustypes.StoreKey,
		upgradetypes.StoreKey,

		// MVP AURA modules
		identitytypes.StoreKey,
		vctypes.StoreKey,
		drtypes.StoreKey,
		compliancetypes.StoreKey,
		prevalidationtypes.StoreKey,
		governancetypes.StoreKey,
		wasmtypes.StoreKey,
		wasmSecurityTypes.StoreKey,
	}
}

// AsMapMVP returns store keys as a map for MountKVStores.
// This is the single source of truth for MVP store key mounting.
func (s *storeKeysMVP) AsMap() map[string]*storetypes.KVStoreKey {
	return map[string]*storetypes.KVStoreKey{
		// Cosmos SDK standard keys
		authtypes.StoreKey:      s.account,
		banktypes.StoreKey:      s.bank,
		stakingtypes.StoreKey:   s.staking,
		slashingtypes.StoreKey:  s.slashing,
		distrtypes.StoreKey:     s.distribution,
		paramstypes.StoreKey:    s.params,
		consensustypes.StoreKey: s.consensus,
		upgradetypes.StoreKey:   s.upgrade,

		// MVP AURA modules
		identitytypes.StoreKey:      s.identity,
		vctypes.StoreKey:            s.vc,
		drtypes.StoreKey:            s.dataRegistry,
		compliancetypes.StoreKey:    s.compliance,
		prevalidationtypes.StoreKey: s.prevalidation,
		governancetypes.StoreKey:    s.governance,
		wasmtypes.StoreKey:          s.wasm,
		wasmSecurityTypes.StoreKey:  s.wasmSecurity,
	}
}

// initStoreKeysMVP creates all store keys for MVP build.
// Adding a new MVP module only requires updating this function.
func initStoreKeysMVP() *storeKeysMVP {
	return &storeKeysMVP{
		// Cosmos SDK standard keys
		account:      storetypes.NewKVStoreKey(authtypes.StoreKey),
		bank:         storetypes.NewKVStoreKey(banktypes.StoreKey),
		staking:      storetypes.NewKVStoreKey(stakingtypes.StoreKey),
		slashing:     storetypes.NewKVStoreKey(slashingtypes.StoreKey),
		distribution: storetypes.NewKVStoreKey(distrtypes.StoreKey),
		params:       storetypes.NewKVStoreKey(paramstypes.StoreKey),
		consensus:    storetypes.NewKVStoreKey(consensustypes.StoreKey),
		upgrade:      storetypes.NewKVStoreKey(upgradetypes.StoreKey),

		// MVP AURA modules
		identity:      storetypes.NewKVStoreKey(identitytypes.StoreKey),
		vc:            storetypes.NewKVStoreKey(vctypes.StoreKey),
		dataRegistry:  storetypes.NewKVStoreKey(drtypes.StoreKey),
		compliance:    storetypes.NewKVStoreKey(compliancetypes.StoreKey),
		prevalidation: storetypes.NewKVStoreKey(prevalidationtypes.StoreKey),
		governance:    storetypes.NewKVStoreKey(governancetypes.StoreKey),
		wasm:          storetypes.NewKVStoreKey(wasmtypes.StoreKey),
		wasmSecurity:  storetypes.NewKVStoreKey(wasmSecurityTypes.StoreKey),
	}
}

// StoreKeyNamesMVP lists all KV store names mounted by the MVP app.
// Uses the centralized storeKeysMVP.Names() as the single source of truth.
func StoreKeyNamesMVP() []string {
	keys := &storeKeysMVP{}
	return keys.Names()
}
