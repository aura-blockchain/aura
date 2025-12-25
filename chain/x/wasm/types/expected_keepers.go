// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// WasmKeeper defines the expected interface for the CosmWasm keeper
type WasmKeeper interface {
	// Execute executes a smart contract
	Execute(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) ([]byte, error)

	// Query queries a smart contract
	Query(ctx sdk.Context, contractAddress sdk.AccAddress, req []byte) ([]byte, error)

	// Instantiate instantiates a smart contract
	Instantiate(ctx sdk.Context, codeID uint64, creator sdk.AccAddress, admin sdk.AccAddress, initMsg []byte, label string, deposit sdk.Coins) (sdk.AccAddress, []byte, error)

	// GetContractInfo retrieves contract information
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
}

// ContractRegistryKeeper defines the expected interface for the contract registry keeper
type ContractRegistryKeeper interface {
	// IsContractRegistered checks if a contract is registered
	IsContractRegistered(ctx sdk.Context, contractAddress string) bool

	// GetContractPolicy retrieves the policy for a contract
	GetContractPolicy(ctx sdk.Context, contractAddress string) (interface{}, error)

	// ValidateContract validates a contract against policies
	ValidateContract(ctx sdk.Context, contractAddress string) error
}
