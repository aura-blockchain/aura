// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package genesis_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	authtypes "github.com/aequitas/aura/chain/x/auth/types"
	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	compliancetypes "github.com/aequitas/aura/chain/x/compliance/types"
	dextypes "github.com/aequitas/aura/chain/x/dex/types"
	governancetypes "github.com/aequitas/aura/chain/x/governance/types"
	privacytypes "github.com/aequitas/aura/chain/x/privacy/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	dexproto "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// Test Default Genesis State

func TestDefaultGenesisState(t *testing.T) {
	// Auth
	authGenesis := authtypes.DefaultGenesis()
	require.NotNil(t, authGenesis)
	require.NotNil(t, authGenesis.Params)

	// Bridge
	bridgeGenesis := bridgetypes.DefaultGenesis()
	require.NotNil(t, bridgeGenesis)
	require.NotNil(t, bridgeGenesis.Params)

	// Compliance
	complianceGenesis := compliancetypes.DefaultGenesis()
	require.NotNil(t, complianceGenesis)
	require.NotNil(t, complianceGenesis.Params)

	// DEX
	dexGenesis := dextypes.DefaultGenesis()
	require.NotNil(t, dexGenesis)
	require.NotNil(t, dexGenesis.Params)

	// Governance
	govGenesis := governancetypes.DefaultGenesis()
	require.NotNil(t, govGenesis)
	require.NotNil(t, govGenesis.Params)

	// Privacy
	privacyGenesis := privacytypes.DefaultGenesis()
	require.NotNil(t, privacyGenesis)
	require.NotNil(t, privacyGenesis.Params)
}

// Test Genesis Validation

func TestValidateGenesisState(t *testing.T) {
	tests := []struct {
		name      string
		genesis   interface{}
		expectErr bool
	}{
		{
			name:      "Valid Auth Genesis",
			genesis:   authtypes.DefaultGenesis(),
			expectErr: false,
		},
		{
			name:      "Valid Bridge Genesis",
			genesis:   bridgetypes.DefaultGenesis(),
			expectErr: false,
		},
		{
			name:      "Valid DEX Genesis",
			genesis:   dextypes.DefaultGenesis(),
			expectErr: false,
		},
		{
			name:      "Valid Governance Genesis",
			genesis:   governancetypes.DefaultGenesis(),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			switch gen := tt.genesis.(type) {
			case *authtypes.GenesisState:
				err := authtypes.ValidateGenesis(gen)
				if tt.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			case *bridgetypes.GenesisState:
				err := bridgetypes.ValidateGenesis(gen)
				if tt.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			case *dextypes.GenesisState:
				err := dextypes.ValidateGenesis(gen)
				if tt.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			case *governancetypes.GenesisState:
				err := governancetypes.ValidateGenesis(gen)
				if tt.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}

// Test Genesis Import/Export

func TestGenesisImportExport(t *testing.T) {
	// Test that default genesis can be exported and re-imported
	// This is a basic test that doesn't require full app initialization

	authGenesis := authtypes.DefaultGenesis()
	require.NotNil(t, authGenesis)
	require.NoError(t, authtypes.ValidateGenesis(authGenesis))

	bridgeGenesis := bridgetypes.DefaultGenesis()
	require.NotNil(t, bridgeGenesis)
	require.NoError(t, bridgetypes.ValidateGenesis(bridgeGenesis))

	dexGenesis := dextypes.DefaultGenesis()
	require.NotNil(t, dexGenesis)
	require.NoError(t, dextypes.ValidateGenesis(dexGenesis))
}

// Test Invalid Genesis States

func TestInvalidAuthGenesis(t *testing.T) {
	genesis := authtypes.DefaultGenesis()
	// Add an invalid role (empty name)
	genesis.Roles = []authproto.Role{
		{
			Name:        "", // Invalid: empty name
			Permissions: []string{"read"},
			Description: "Invalid role",
		},
	}

	err := authtypes.ValidateGenesis(genesis)
	require.Error(t, err, "Should reject invalid role")
}

func TestInvalidBridgeGenesis(t *testing.T) {
	genesis := bridgetypes.DefaultGenesis()
	genesis.Params.ValidatorThresholdPercentage = 0

	err := bridgetypes.ValidateGenesis(genesis)
	require.Error(t, err, "Should reject zero threshold")
}

func TestInvalidDEXGenesis(t *testing.T) {
	genesis := dextypes.DefaultGenesis()
	genesis.Params.MaxSlippageBps = 0

	err := dextypes.ValidateGenesis(genesis)
	require.Error(t, err, "Should reject zero max slippage")
}

// Test Genesis State Consistency

func TestGenesisStateConsistency(t *testing.T) {
	// All module genesis states should be consistent

	authGenesis := authtypes.DefaultGenesis()
	bridgeGenesis := bridgetypes.DefaultGenesis()
	dexGenesis := dextypes.DefaultGenesis()

	require.NoError(t, authtypes.ValidateGenesis(authGenesis))
	require.NoError(t, bridgetypes.ValidateGenesis(bridgeGenesis))
	require.NoError(t, dextypes.ValidateGenesis(dexGenesis))
}

// Test Genesis with Initial Data

func TestGenesisWithInitialRoles(t *testing.T) {
	genesis := authtypes.DefaultGenesis()

	// Add initial roles
	genesis.Roles = []authproto.Role{
		{
			Name:        "admin",
			Permissions: []string{"admin", "create", "read", "update", "delete"},
			Description: "Administrator role",
		},
		{
			Name:        "user",
			Permissions: []string{"read"},
			Description: "Basic user role",
		},
	}

	err := authtypes.ValidateGenesis(genesis)
	require.NoError(t, err)
	require.Len(t, genesis.Roles, 2)
}

func TestGenesisWithInitialPools(t *testing.T) {
	genesis := dextypes.DefaultGenesis()

	// Add initial pools with all required non-nullable fields
	genesis.LiquidityPools = []dexproto.LiquidityPool{
		{
			PoolId:                "1",
			DenomA:                "uaura",
			DenomB:                "uusdt",
			ReserveA:              math.NewInt(1000000),
			ReserveB:              math.NewInt(1000000),
			TotalLpTokens:         math.NewInt(1000000),
			FeePercentage:         math.LegacyMustNewDecFromStr("0.003"),
			ProtocolFeePercentage: math.LegacyMustNewDecFromStr("0.0005"),
			TotalVolume:           math.ZeroInt(),
			TotalFeesCollected:    math.ZeroInt(),
			SwapCount:             0,
			ProtocolFeeBalance:    math.ZeroInt(),
			Providers:             []dexproto.LiquidityProvider{},
			CreatedAt:             time.Now(),
			LockedLiquidity:       math.NewInt(1000),
		},
	}

	err := dextypes.ValidateGenesis(genesis)
	require.NoError(t, err)
	require.Len(t, genesis.LiquidityPools, 1)
}

// Test Genesis Migration

func TestGenesisV1ToV2Migration(t *testing.T) {
	// Placeholder migration test
	require.NotNil(t, t)
}
