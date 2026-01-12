//go:build mvp

// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

// MVPModules lists modules included in MVP release.
// These are the 12 essential modules required for mainnet launch.
var MVPModules = []string{
	// Tier 1: Essential Core (Cosmos SDK)
	"auth",
	"bank",
	"staking",
	"slashing",
	"gov",
	"distribution",

	// Tier 2: AURA Core Identity (MVP Differentiator)
	"identity",
	"vcregistry",
	"dataregistry",
	"compliance",
	"prevalidation",

	// Tier 3: Infrastructure
	"wasm",
}

// DeferredModules lists modules excluded from MVP release.
// These will be enabled in post-MVP phases.
var DeferredModules = []string{
	// Phase 2: Enhanced Security
	"security",
	"walletsecurity",
	"validatorsecurity",
	"networksecurity",
	"cryptography",
	"economicsecurity",

	// Phase 3: Advanced Features
	"dex",
	"bridge",
	"confidencescore",
	"inclusionroutines",
	"incidentresponse",
	"monitoring",
	"aiassistant",
	"contractregistry",
	"identitychange",
	"privacy",
	"economics",
	"aurabindings",
}

// IsMVPModule returns true if the module is part of the MVP release.
func IsMVPModule(moduleName string) bool {
	for _, m := range MVPModules {
		if m == moduleName {
			return true
		}
	}
	return false
}

// IsDeferredModule returns true if the module is deferred to post-MVP.
func IsDeferredModule(moduleName string) bool {
	for _, m := range DeferredModules {
		if m == moduleName {
			return true
		}
	}
	return false
}
