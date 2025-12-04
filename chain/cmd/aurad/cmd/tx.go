package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	distrcli "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"

	"github.com/aequitas/aura/chain/app"
	compliancecli "github.com/aequitas/aura/chain/x/compliance/client/cli"
	confidencescorecli "github.com/aequitas/aura/chain/x/confidencescore/client/cli"
	dexcli "github.com/aequitas/aura/chain/x/dex/client/cli"
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
		confidencescorecli.GetTxCmd(),
		compliancecli.GetTxCmd(),
		dexcli.GetTxCmd(),
		wasmcli.GetTxCmd(),
	)

	// Note: Additional module tx commands can be added here as modules are completed:
	// - monitoring module txs (x/monitoring/client/cli)
	// - security module txs (x/networksecurity/client/cli)
	// - identity module txs (x/identity/client/cli)
	// - vcregistry txs (x/vcregistry/client/cli)
	// etc.

	return cmd
}
