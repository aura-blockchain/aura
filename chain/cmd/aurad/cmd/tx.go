// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	distrcli "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"

	"github.com/aequitas/aura/chain/app"
	aiassistantcli "github.com/aequitas/aura/chain/x/aiassistant/client/cli"
	bridgecli "github.com/aequitas/aura/chain/x/bridge/client/cli"
	compliancecli "github.com/aequitas/aura/chain/x/compliance/client/cli"
	confidencescorecli "github.com/aequitas/aura/chain/x/confidencescore/client/cli"
	cryptographycli "github.com/aequitas/aura/chain/x/cryptography/client/cli"
	dataregistrycli "github.com/aequitas/aura/chain/x/dataregistry/client/cli"
	dexcli "github.com/aequitas/aura/chain/x/dex/client/cli"
	economicsecuritycli "github.com/aequitas/aura/chain/x/economicsecurity/client/cli"
	governancecli "github.com/aequitas/aura/chain/x/governance/client/cli"
	identitychangecli "github.com/aequitas/aura/chain/x/identitychange/client/cli"
	monitoringcli "github.com/aequitas/aura/chain/x/monitoring/client/cli"
	networksecuritycli "github.com/aequitas/aura/chain/x/networksecurity/client/cli"
	prevalidationcli "github.com/aequitas/aura/chain/x/prevalidation/client/cli"
	privacycli "github.com/aequitas/aura/chain/x/privacy/client/cli"
	validatorsecuritycli "github.com/aequitas/aura/chain/x/validatorsecurity/client/cli"
	vcregistrycli "github.com/aequitas/aura/chain/x/vcregistry/client/cli"
	walletsecuritycli "github.com/aequitas/aura/chain/x/walletsecurity/client/cli"
	wasmcli "github.com/aequitas/aura/chain/x/wasm/client/cli"
)

// TxCmd returns the production transaction command tree (bank/send, staking, wasm, custom modules)
// powered by the Cosmos SDK client context.
func TxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Create, sign, and broadcast transactions",
		Long:  "Transaction subcommands for all modules, including signing/broadcast helpers and offline flows.",
		RunE:  client.ValidateCmd,
	}

	// Core signing helpers and tx utilities
	cmd.AddCommand(
		GetAuraSignCommand(),
		authcli.GetSignBatchCommand(),
		authcli.GetMultiSignCommand(),
		authcli.GetValidateSignaturesCommand(),
		authcli.GetBroadcastCommand(),
		authcli.GetEncodeCommand(),
		authcli.GetDecodeCommand(),
		authcli.GetSimulateCmd(),
	)

	// Core Cosmos SDK tx commands with properly configured address codecs.
	accAddrCodec := app.AccountAddressCodec()
	valAddrCodec := app.ValidatorAddressCodec()

	cmd.AddCommand(
		bankcli.NewTxCmd(accAddrCodec),
		stakingcli.NewTxCmd(valAddrCodec, accAddrCodec),
		distrcli.NewTxCmd(valAddrCodec, accAddrCodec),
	)

	// Aura module tx commands (retain full coverage for custom modules).
	cmd.AddCommand(
		bridgecli.GetTxCmd(),
		compliancecli.GetTxCmd(),
		confidencescorecli.GetTxCmd(),
		cryptographycli.GetTxCmd(),
		dataregistrycli.GetTxCmd(),
		dexcli.GetTxCmd(),
		economicsecuritycli.GetTxCmd(),
		governancecli.GetTxCmd(),
		aiassistantcli.NewTxCmd(),
		identitychangecli.GetTxCmd(),
		monitoringcli.GetTxCmd(),
		networksecuritycli.GetTxCmd(),
		prevalidationcli.GetTxCmd(),
		privacycli.GetTxCmd(),
		validatorsecuritycli.GetTxCmd(),
		vcregistrycli.GetTxCmd(),
		walletsecuritycli.GetTxCmd(),
		wasmcli.GetTxCmd(),
	)

	return cmd
}
