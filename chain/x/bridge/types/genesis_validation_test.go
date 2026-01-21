// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// Valid test address (valid bech32)
const validBech32Addr = "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"

func TestValidateGenesis_NilGenesis(t *testing.T) {
	err := types.ValidateGenesis(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "genesis state cannot be nil")
}

func TestValidateGenesis_ValidEmpty(t *testing.T) {
	genesis := &types.GenesisState{
		Params: types.BridgeParams{
			Enabled:                      true,
			MinConfirmations:             3,
			BridgeFeeBasisPoints:         30,
			MaxTransferAmount:            sdkmath.NewInt(1000000),
			ValidatorThresholdPercentage: 67,
		},
	}

	err := types.ValidateGenesis(genesis)
	require.NoError(t, err)
}

func TestValidateGenesis_InvalidParams(t *testing.T) {
	tests := []struct {
		name   string
		params types.BridgeParams
		errMsg string
	}{
		{
			name: "zero min confirmations",
			params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             0,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			errMsg: "min confirmations must be greater than zero",
		},
		{
			name: "excessive min confirmations",
			params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             1001,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			errMsg: "min confirmations cannot exceed 1000",
		},
		{
			name: "excessive bridge fee",
			params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         10001,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			errMsg: "bridge fee basis points must be 10000 or less",
		},
		{
			name: "zero validator threshold",
			params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 0,
			},
			errMsg: "validator threshold percentage must be greater than zero",
		},
		{
			name: "excessive validator threshold",
			params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 101,
			},
			errMsg: "validator threshold percentage cannot exceed 100",
		},
		{
			name: "negative max transfer amount",
			params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(-1000),
				ValidatorThresholdPercentage: 67,
			},
			errMsg: "max transfer amount cannot be negative",
		},
		{
			name: "zero max transfer amount",
			params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.ZeroInt(),
				ValidatorThresholdPercentage: 67,
			},
			errMsg: "max transfer amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := &types.GenesisState{
				Params: tt.params,
			}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestValidateGenesis_ValidTransfers(t *testing.T) {
	genesis := &types.GenesisState{
		Params: types.BridgeParams{
			Enabled:                      true,
			MinConfirmations:             3,
			BridgeFeeBasisPoints:         30,
			MaxTransferAmount:            sdkmath.NewInt(1000000),
			ValidatorThresholdPercentage: 67,
		},
		Transfers: []types.CrossChainTransfer{
			{
				TransferId:            "transfer-1",
				SourceChain:           "aura",
				TargetChain:           "paw",
				Sender:                validBech32Addr,
				Recipient:             "paw1recipient",
				Amount:                sdkmath.NewInt(1000),
				Denom:                 "uaura",
				Status:                types.TransferStatus_PENDING,
				Confirmations:         1,
				RequiredConfirmations: 3,
			},
		},
	}

	err := types.ValidateGenesis(genesis)
	require.NoError(t, err)
}

func TestValidateGenesis_InvalidTransfers(t *testing.T) {
	baseParams := types.BridgeParams{
		Enabled:                      true,
		MinConfirmations:             3,
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            sdkmath.NewInt(1000000),
		ValidatorThresholdPercentage: 67,
	}

	tests := []struct {
		name     string
		transfer types.CrossChainTransfer
		errMsg   string
	}{
		{
			name: "empty source chain",
			transfer: types.CrossChainTransfer{
				TransferId:  "transfer-1",
				SourceChain: "",
				TargetChain: "paw",
				Sender:      validBech32Addr,
				Recipient:   "paw1recipient",
				Amount:      sdkmath.NewInt(1000),
				Denom:       "uaura",
			},
			errMsg: "source chain cannot be empty",
		},
		{
			name: "empty target chain",
			transfer: types.CrossChainTransfer{
				TransferId:  "transfer-1",
				SourceChain: "aura",
				TargetChain: "",
				Sender:      validBech32Addr,
				Recipient:   "paw1recipient",
				Amount:      sdkmath.NewInt(1000),
				Denom:       "uaura",
			},
			errMsg: "target chain cannot be empty",
		},
		{
			name: "same source and target chain",
			transfer: types.CrossChainTransfer{
				TransferId:  "transfer-1",
				SourceChain: "aura",
				TargetChain: "aura",
				Sender:      validBech32Addr,
				Recipient:   "aura1recipient",
				Amount:      sdkmath.NewInt(1000),
				Denom:       "uaura",
			},
			errMsg: "source chain and target chain must be different",
		},
		{
			name: "empty sender",
			transfer: types.CrossChainTransfer{
				TransferId:  "transfer-1",
				SourceChain: "aura",
				TargetChain: "paw",
				Sender:      "",
				Recipient:   "paw1recipient",
				Amount:      sdkmath.NewInt(1000),
				Denom:       "uaura",
			},
			errMsg: "sender address cannot be empty",
		},
		{
			name: "empty recipient",
			transfer: types.CrossChainTransfer{
				TransferId:  "transfer-1",
				SourceChain: "aura",
				TargetChain: "paw",
				Sender:      validBech32Addr,
				Recipient:   "",
				Amount:      sdkmath.NewInt(1000),
				Denom:       "uaura",
			},
			errMsg: "recipient address cannot be empty",
		},
		{
			name: "zero amount",
			transfer: types.CrossChainTransfer{
				TransferId:  "transfer-1",
				SourceChain: "aura",
				TargetChain: "paw",
				Sender:      validBech32Addr,
				Recipient:   "paw1recipient",
				Amount:      sdkmath.ZeroInt(),
				Denom:       "uaura",
			},
			errMsg: "transfer amount must be positive",
		},
		{
			name: "empty denom",
			transfer: types.CrossChainTransfer{
				TransferId:  "transfer-1",
				SourceChain: "aura",
				TargetChain: "paw",
				Sender:      validBech32Addr,
				Recipient:   "paw1recipient",
				Amount:      sdkmath.NewInt(1000),
				Denom:       "",
			},
			errMsg: "denom cannot be empty",
		},
		{
			name: "confirmations exceed required",
			transfer: types.CrossChainTransfer{
				TransferId:            "transfer-1",
				SourceChain:           "aura",
				TargetChain:           "paw",
				Sender:                validBech32Addr,
				Recipient:             "paw1recipient",
				Amount:                sdkmath.NewInt(1000),
				Denom:                 "uaura",
				Confirmations:         5,
				RequiredConfirmations: 3,
			},
			errMsg: "confirmations (5) cannot exceed required confirmations (3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := &types.GenesisState{
				Params:    baseParams,
				Transfers: []types.CrossChainTransfer{tt.transfer},
			}
			err := types.ValidateGenesis(genesis)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestValidateGenesis_DuplicateTransferIDs(t *testing.T) {
	genesis := &types.GenesisState{
		Params: types.BridgeParams{
			Enabled:                      true,
			MinConfirmations:             3,
			BridgeFeeBasisPoints:         30,
			MaxTransferAmount:            sdkmath.NewInt(1000000),
			ValidatorThresholdPercentage: 67,
		},
		Transfers: []types.CrossChainTransfer{
			{
				TransferId:  "transfer-1",
				SourceChain: "aura",
				TargetChain: "paw",
				Sender:      validBech32Addr,
				Recipient:   "paw1recipient",
				Amount:      sdkmath.NewInt(1000),
				Denom:       "uaura",
			},
			{
				TransferId:  "transfer-1", // Duplicate
				SourceChain: "aura",
				TargetChain: "xai",
				Sender:      validBech32Addr,
				Recipient:   "xai1recipient",
				Amount:      sdkmath.NewInt(2000),
				Denom:       "uaura",
			},
		},
	}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate transfer ID")
}

func TestValidateGenesis_EmptyTransferID(t *testing.T) {
	// Empty transfer IDs should be skipped gracefully
	genesis := &types.GenesisState{
		Params: types.BridgeParams{
			Enabled:                      true,
			MinConfirmations:             3,
			BridgeFeeBasisPoints:         30,
			MaxTransferAmount:            sdkmath.NewInt(1000000),
			ValidatorThresholdPercentage: 67,
		},
		Transfers: []types.CrossChainTransfer{
			{
				TransferId:  "", // Empty, should be skipped
				SourceChain: "aura",
				TargetChain: "paw",
			},
			{
				TransferId:  "transfer-1",
				SourceChain: "aura",
				TargetChain: "paw",
				Sender:      validBech32Addr,
				Recipient:   "paw1recipient",
				Amount:      sdkmath.NewInt(1000),
				Denom:       "uaura",
			},
		},
	}

	err := types.ValidateGenesis(genesis)
	require.NoError(t, err)
}

func TestValidateGenesis_ValidatorSignatures(t *testing.T) {
	t.Run("valid signatures", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "aura",
					TargetChain: "paw",
					Sender:      validBech32Addr,
					Recipient:   "paw1recipient",
					Amount:      sdkmath.NewInt(1000),
					Denom:       "uaura",
					ValidatorSignatures: []types.ValidatorSignature{
						{ValidatorAddress: validBech32Addr, Signature: []byte("signature1")},
					},
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})

	t.Run("empty validator address in signature", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "aura",
					TargetChain: "paw",
					Sender:      validBech32Addr,
					Recipient:   "paw1recipient",
					Amount:      sdkmath.NewInt(1000),
					Denom:       "uaura",
					ValidatorSignatures: []types.ValidatorSignature{
						{ValidatorAddress: "", Signature: []byte("signature1")},
					},
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "validator address cannot be empty")
	})

	t.Run("empty signature", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "aura",
					TargetChain: "paw",
					Sender:      validBech32Addr,
					Recipient:   "paw1recipient",
					Amount:      sdkmath.NewInt(1000),
					Denom:       "uaura",
					ValidatorSignatures: []types.ValidatorSignature{
						{ValidatorAddress: validBech32Addr, Signature: []byte{}},
					},
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "signature cannot be empty")
	})

	t.Run("duplicate validator signatures", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: types.BridgeParams{
				Enabled:                      true,
				MinConfirmations:             3,
				BridgeFeeBasisPoints:         30,
				MaxTransferAmount:            sdkmath.NewInt(1000000),
				ValidatorThresholdPercentage: 67,
			},
			Transfers: []types.CrossChainTransfer{
				{
					TransferId:  "transfer-1",
					SourceChain: "aura",
					TargetChain: "paw",
					Sender:      validBech32Addr,
					Recipient:   "paw1recipient",
					Amount:      sdkmath.NewInt(1000),
					Denom:       "uaura",
					ValidatorSignatures: []types.ValidatorSignature{
						{ValidatorAddress: validBech32Addr, Signature: []byte("sig1")},
						{ValidatorAddress: validBech32Addr, Signature: []byte("sig2")}, // Duplicate
					},
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate validator signature")
	})
}

func TestValidateGenesis_ChainConfigs(t *testing.T) {
	baseParams := types.BridgeParams{
		Enabled:                      true,
		MinConfirmations:             3,
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            sdkmath.NewInt(1000000),
		ValidatorThresholdPercentage: 67,
	}

	t.Run("valid chain configs", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			ChainConfigs: []types.ChainConfig{
				{
					ChainId:          "paw",
					ChainName:        "PAW Chain",
					AddressPrefix:    "paw",
					MinConfirmations: 3,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})

	t.Run("empty chain id", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			ChainConfigs: []types.ChainConfig{
				{
					ChainId:          "",
					ChainName:        "PAW Chain",
					AddressPrefix:    "paw",
					MinConfirmations: 3,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "chain ID cannot be empty")
	})

	t.Run("empty chain name", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			ChainConfigs: []types.ChainConfig{
				{
					ChainId:          "paw",
					ChainName:        "",
					AddressPrefix:    "paw",
					MinConfirmations: 3,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "chain name cannot be empty")
	})

	t.Run("empty address prefix", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			ChainConfigs: []types.ChainConfig{
				{
					ChainId:          "paw",
					ChainName:        "PAW Chain",
					AddressPrefix:    "",
					MinConfirmations: 3,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address prefix cannot be empty")
	})

	t.Run("zero min confirmations", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			ChainConfigs: []types.ChainConfig{
				{
					ChainId:          "paw",
					ChainName:        "PAW Chain",
					AddressPrefix:    "paw",
					MinConfirmations: 0,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "min confirmations must be greater than zero")
	})

	t.Run("duplicate chain ids", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			ChainConfigs: []types.ChainConfig{
				{
					ChainId:          "paw",
					ChainName:        "PAW Chain",
					AddressPrefix:    "paw",
					MinConfirmations: 3,
				},
				{
					ChainId:          "paw", // Duplicate
					ChainName:        "PAW Chain 2",
					AddressPrefix:    "paw2",
					MinConfirmations: 5,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate chain ID")
	})
}

func TestValidateGenesis_Validators(t *testing.T) {
	baseParams := types.BridgeParams{
		Enabled:                      true,
		MinConfirmations:             3,
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            sdkmath.NewInt(1000000),
		ValidatorThresholdPercentage: 67,
	}

	t.Run("empty validators allowed", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params:     baseParams,
			Validators: []types.BridgeValidator{},
		}

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})

	t.Run("valid validators", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			Validators: []types.BridgeValidator{
				{Address: validBech32Addr, Power: 100},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})

	t.Run("empty validator address", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			Validators: []types.BridgeValidator{
				{Address: "", Power: 100},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "validator address cannot be empty")
	})

	t.Run("invalid validator address", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			Validators: []types.BridgeValidator{
				{Address: "invalid", Power: 100},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid validator address")
	})

	t.Run("zero voting power", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			Validators: []types.BridgeValidator{
				{Address: validBech32Addr, Power: 0},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "validator voting power must be greater than zero")
	})

	t.Run("duplicate validators", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			Validators: []types.BridgeValidator{
				{Address: validBech32Addr, Power: 100},
				{Address: validBech32Addr, Power: 200}, // Duplicate
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate validator address")
	})
}

func TestValidateGenesis_WrappedTokens(t *testing.T) {
	baseParams := types.BridgeParams{
		Enabled:                      true,
		MinConfirmations:             3,
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            sdkmath.NewInt(1000000),
		ValidatorThresholdPercentage: 67,
	}

	t.Run("valid wrapped tokens", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			WrappedTokens: []types.WrappedToken{
				{
					WrappedDenom:  "wpaw",
					SourceChain:   "paw",
					OriginalDenom: "upaw",
					TotalSupply:   sdkmath.NewInt(1000000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})

	t.Run("empty wrapped denom", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			WrappedTokens: []types.WrappedToken{
				{
					WrappedDenom:  "",
					SourceChain:   "paw",
					OriginalDenom: "upaw",
					TotalSupply:   sdkmath.NewInt(1000000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "wrapped denom cannot be empty")
	})

	t.Run("empty source chain", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			WrappedTokens: []types.WrappedToken{
				{
					WrappedDenom:  "wpaw",
					SourceChain:   "",
					OriginalDenom: "upaw",
					TotalSupply:   sdkmath.NewInt(1000000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "source chain cannot be empty")
	})

	t.Run("empty original denom", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			WrappedTokens: []types.WrappedToken{
				{
					WrappedDenom:  "wpaw",
					SourceChain:   "paw",
					OriginalDenom: "",
					TotalSupply:   sdkmath.NewInt(1000000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "original denom cannot be empty")
	})

	t.Run("negative total supply", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			WrappedTokens: []types.WrappedToken{
				{
					WrappedDenom:  "wpaw",
					SourceChain:   "paw",
					OriginalDenom: "upaw",
					TotalSupply:   sdkmath.NewInt(-1000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "total supply cannot be nil or negative")
	})

	t.Run("duplicate wrapped tokens", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			WrappedTokens: []types.WrappedToken{
				{
					WrappedDenom:  "wpaw",
					SourceChain:   "paw",
					OriginalDenom: "upaw",
					TotalSupply:   sdkmath.NewInt(1000000),
				},
				{
					WrappedDenom:  "wpaw",  // Same denom
					SourceChain:   "paw",   // Same chain
					OriginalDenom: "upaw2", // Different original
					TotalSupply:   sdkmath.NewInt(2000000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate wrapped token")
	})
}

func TestValidateGenesis_SharedIdentities(t *testing.T) {
	baseParams := types.BridgeParams{
		Enabled:                      true,
		MinConfirmations:             3,
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            sdkmath.NewInt(1000000),
		ValidatorThresholdPercentage: 67,
	}

	t.Run("valid shared identities", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			SharedIdentities: []types.SharedIdentity{
				{
					Address:         validBech32Addr,
					ReputationScore: 500,
					AuraIrScore:     75,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})

	t.Run("empty address", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			SharedIdentities: []types.SharedIdentity{
				{
					Address:         "",
					ReputationScore: 500,
					AuraIrScore:     75,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address cannot be empty")
	})

	t.Run("excessive reputation score", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			SharedIdentities: []types.SharedIdentity{
				{
					Address:         validBech32Addr,
					ReputationScore: 1001,
					AuraIrScore:     75,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "reputation score cannot exceed 1000")
	})

	t.Run("excessive IR score", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			SharedIdentities: []types.SharedIdentity{
				{
					Address:         validBech32Addr,
					ReputationScore: 500,
					AuraIrScore:     101,
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "AURA IR score cannot exceed 100")
	})

	t.Run("duplicate addresses", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			SharedIdentities: []types.SharedIdentity{
				{Address: validBech32Addr, ReputationScore: 500, AuraIrScore: 75},
				{Address: validBech32Addr, ReputationScore: 600, AuraIrScore: 80}, // Duplicate
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate shared identity address")
	})
}

func TestValidateGenesis_CrossChainSwaps(t *testing.T) {
	baseParams := types.BridgeParams{
		Enabled:                      true,
		MinConfirmations:             3,
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            sdkmath.NewInt(1000000),
		ValidatorThresholdPercentage: 67,
	}

	t.Run("valid swap", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			CrossChainSwaps: []types.CrossChainSwap{
				{
					SwapId:          "swap-1",
					Sender:          validBech32Addr,
					SourceChain:     "aura",
					TargetChain:     "paw",
					TargetDenom:     "upaw",
					MinTargetAmount: sdkmath.NewInt(900),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})

	t.Run("empty swap id", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			CrossChainSwaps: []types.CrossChainSwap{
				{
					SwapId:          "",
					Sender:          validBech32Addr,
					SourceChain:     "aura",
					TargetChain:     "paw",
					TargetDenom:     "upaw",
					MinTargetAmount: sdkmath.NewInt(900),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "swap ID cannot be empty")
	})

	t.Run("empty sender", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			CrossChainSwaps: []types.CrossChainSwap{
				{
					SwapId:          "swap-1",
					Sender:          "",
					SourceChain:     "aura",
					TargetChain:     "paw",
					TargetDenom:     "upaw",
					MinTargetAmount: sdkmath.NewInt(900),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "sender address cannot be empty")
	})

	t.Run("same source and target chain", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			CrossChainSwaps: []types.CrossChainSwap{
				{
					SwapId:          "swap-1",
					Sender:          validBech32Addr,
					SourceChain:     "aura",
					TargetChain:     "aura",
					TargetDenom:     "uaura",
					MinTargetAmount: sdkmath.NewInt(900),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "source chain and target chain must be different")
	})

	t.Run("empty target denom", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			CrossChainSwaps: []types.CrossChainSwap{
				{
					SwapId:          "swap-1",
					Sender:          validBech32Addr,
					SourceChain:     "aura",
					TargetChain:     "paw",
					TargetDenom:     "",
					MinTargetAmount: sdkmath.NewInt(900),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "target denom cannot be empty")
	})

	t.Run("negative min target amount", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			CrossChainSwaps: []types.CrossChainSwap{
				{
					SwapId:          "swap-1",
					Sender:          validBech32Addr,
					SourceChain:     "aura",
					TargetChain:     "paw",
					TargetDenom:     "upaw",
					MinTargetAmount: sdkmath.NewInt(-100),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "min target amount cannot be nil or negative")
	})

	t.Run("duplicate swap ids", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			CrossChainSwaps: []types.CrossChainSwap{
				{
					SwapId:          "swap-1",
					Sender:          validBech32Addr,
					SourceChain:     "aura",
					TargetChain:     "paw",
					TargetDenom:     "upaw",
					MinTargetAmount: sdkmath.NewInt(900),
				},
				{
					SwapId:          "swap-1", // Duplicate
					Sender:          validBech32Addr,
					SourceChain:     "paw",
					TargetChain:     "aura",
					TargetDenom:     "uaura",
					MinTargetAmount: sdkmath.NewInt(800),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate swap ID")
	})
}

func TestValidateGenesis_RelayerStats(t *testing.T) {
	baseParams := types.BridgeParams{
		Enabled:                      true,
		MinConfirmations:             3,
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            sdkmath.NewInt(1000000),
		ValidatorThresholdPercentage: 67,
	}

	t.Run("valid relayer stats", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			RelayerStats: []types.RelayerStats{
				{
					RelayerAddress:        validBech32Addr,
					TotalTransfersRelayed: 100,
					SuccessfulTransfers:   95,
					FailedTransfers:       5,
					TotalVolume:           sdkmath.NewInt(1000000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})

	t.Run("empty relayer address", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			RelayerStats: []types.RelayerStats{
				{
					RelayerAddress:        "",
					TotalTransfersRelayed: 100,
					SuccessfulTransfers:   95,
					FailedTransfers:       5,
					TotalVolume:           sdkmath.NewInt(1000000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "relayer address cannot be empty")
	})

	t.Run("invalid relayer address", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			RelayerStats: []types.RelayerStats{
				{
					RelayerAddress:        "invalid",
					TotalTransfersRelayed: 100,
					SuccessfulTransfers:   95,
					FailedTransfers:       5,
					TotalVolume:           sdkmath.NewInt(1000000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid relayer address")
	})

	t.Run("inconsistent transfer counts", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			RelayerStats: []types.RelayerStats{
				{
					RelayerAddress:        validBech32Addr,
					TotalTransfersRelayed: 100,
					SuccessfulTransfers:   90,
					FailedTransfers:       5, // 90 + 5 = 95 != 100
					TotalVolume:           sdkmath.NewInt(1000000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "sum of successful")
	})

	t.Run("negative total volume", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			RelayerStats: []types.RelayerStats{
				{
					RelayerAddress:        validBech32Addr,
					TotalTransfersRelayed: 100,
					SuccessfulTransfers:   95,
					FailedTransfers:       5,
					TotalVolume:           sdkmath.NewInt(-1000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "total volume cannot be nil or negative")
	})

	t.Run("duplicate relayer addresses", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			RelayerStats: []types.RelayerStats{
				{
					RelayerAddress:        validBech32Addr,
					TotalTransfersRelayed: 100,
					SuccessfulTransfers:   95,
					FailedTransfers:       5,
					TotalVolume:           sdkmath.NewInt(1000000),
				},
				{
					RelayerAddress:        validBech32Addr, // Duplicate
					TotalTransfersRelayed: 50,
					SuccessfulTransfers:   45,
					FailedTransfers:       5,
					TotalVolume:           sdkmath.NewInt(500000),
				},
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate relayer stats")
	})
}

func TestValidateGenesis_ProcessedSourceHashes(t *testing.T) {
	baseParams := types.BridgeParams{
		Enabled:                      true,
		MinConfirmations:             3,
		BridgeFeeBasisPoints:         30,
		MaxTransferAmount:            sdkmath.NewInt(1000000),
		ValidatorThresholdPercentage: 67,
	}

	t.Run("valid hashes", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			ProcessedSourceHashes: []string{
				"hash1", "hash2", "hash3",
			},
		}

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)
	})

	t.Run("empty hash", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			ProcessedSourceHashes: []string{
				"hash1", "", "hash3",
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "processed source hash cannot be empty")
	})

	t.Run("duplicate hashes", func(t *testing.T) {
		genesis := &types.GenesisState{
			Params: baseParams,
			ProcessedSourceHashes: []string{
				"hash1", "hash2", "hash1",
			},
		}

		err := types.ValidateGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate processed source hash")
	})
}
