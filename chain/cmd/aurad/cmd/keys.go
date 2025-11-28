package cmd

import (
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/spf13/cobra"
)

// KeysCmd returns the keys command using the SDK's standard implementation
func KeysCmd() *cobra.Command {
	// Use the SDK's standard keys commands which properly handle keyring migration
	// and are tested to work with SDK 0.53.x
	cmd := keys.Commands()
	cmd.Use = "keys"
	cmd.Short = "Manage keyring and keys"
	cmd.Long = `Manage your local keyring for signing transactions.

The keys are stored in the keyring backend specified in the config.
Available backends: os, file, test, memory.

The keyring supports multiple algorithms for key generation:
- secp256k1 (default, Bitcoin/Cosmos standard)
- ed25519 (Tendermint validator keys)
- sr25519 (Schnorrkel/Ristretto)

Examples:
  # Add a new key
  aurad keys add mykey

  # Add a key with recovery from mnemonic
  aurad keys add mykey --recover

  # List all keys
  aurad keys list

  # Show key details
  aurad keys show mykey

  # Delete a key
  aurad keys delete mykey`

	return cmd
}
