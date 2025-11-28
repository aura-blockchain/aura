package cmd

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagInteractive = "interactive"
	flagRecover     = "recover"
	flagNoBackup    = "no-backup"
	flagCoinType    = "coin-type"
	flagAccount     = "account"
	flagIndex       = "index"
	flagMultisig    = "multisig"
	flagNoSort      = "nosort"
	flagHDPath      = "hd-path"
	flagPubKeyType  = "pubkey-type"
	flagDryRun      = "dry-run"
	flagShowAddress = "address"
	flagShowPubKey  = "pubkey"
	flagShowBech32  = "bech"
	flagShowDevice  = "device"
)

// KeysCmd returns the keys command for the Aura daemon
func KeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage keyring and keys",
		Long: `Manage your local keyring for signing transactions.

The keys are stored in the keyring backend specified in the config.
Available backends: os, file, test, memory.

The keyring supports multiple algorithms for key generation:
- secp256k1 (default, Bitcoin/Cosmos standard)
- ed25519 (Tendermint validator keys)
- sr25519 (Schnorrkel/Ristretto)`,
	}

	// Add persistent flags for keyring configuration
	cmd.PersistentFlags().String(flags.FlagHome, getDefaultHomeDir(), "The application home directory")
	cmd.PersistentFlags().String(flags.FlagKeyringDir, "", "The client keyring directory; if omitted, the default 'home' directory will be used")
	cmd.PersistentFlags().String(flags.FlagKeyringBackend, keyring.BackendOS, "Select keyring's backend (os|file|test|memory)")
	cmd.PersistentFlags().String(flags.FlagOutput, "text", "Output format (text|json)")

	// Bind to viper for config file support
	viper.BindPFlag(flags.FlagHome, cmd.PersistentFlags().Lookup(flags.FlagHome))
	viper.BindPFlag(flags.FlagKeyringDir, cmd.PersistentFlags().Lookup(flags.FlagKeyringDir))
	viper.BindPFlag(flags.FlagKeyringBackend, cmd.PersistentFlags().Lookup(flags.FlagKeyringBackend))

	cmd.AddCommand(
		keys.MnemonicKeyCommand(),
		keysAddCmd(),
		keys.ExportKeyCommand(),
		keys.ImportKeyCommand(),
		keys.ImportKeyHexCommand(),
		keysListCmd(),
		keysShowCmd(),
		keysDeleteCmd(),
		keys.ParseKeyStringCommand(),
		keys.MigrateCommand(),
		keysRenameCmd(),
	)

	return cmd
}

// keysAddCmd returns the keys add command
func keysAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new key to the keyring",
		Long: `Derive a new private key and encrypt to disk.

Optionally specify a BIP39 mnemonic, a BIP39 passphrase to further secure the mnemonic,
and a BIP32 HD path to derive a specific account. The key signing algorithm can also be
selected (secp256k1, ed25519, sr25519).

The default HD path follows the BIP44 standard: m/44'/<coin_type>'/<account>'/<change>/<index>
For Cosmos chains, the default coin type is 118.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			cdc := clientCtx.Codec

			name := args[0]
			if name == "" {
				return fmt.Errorf("key name cannot be empty")
			}

			// Get flags
			interactive, _ := cmd.Flags().GetBool(flagInteractive)
			recover, _ := cmd.Flags().GetBool(flagRecover)
			noBackup, _ := cmd.Flags().GetBool(flagNoBackup)
			dryRun, _ := cmd.Flags().GetBool(flagDryRun)

			// Get keyring backend and initialize keyring
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			homeDir := viper.GetString(flags.FlagHome)
			keyringDir := viper.GetString(flags.FlagKeyringDir)
			if keyringDir == "" {
				keyringDir = homeDir
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())
			kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, keyringDir, inBuf, cdc)
			if err != nil {
				return fmt.Errorf("failed to initialize keyring: %w", err)
			}

			// Check if key already exists
			_, err = kb.Key(name)
			if err == nil {
				return fmt.Errorf("key with name '%s' already exists in keyring", name)
			}

			// Get HD path parameters
			coinType, _ := cmd.Flags().GetUint32(flagCoinType)
			account, _ := cmd.Flags().GetUint32(flagAccount)
			index, _ := cmd.Flags().GetUint32(flagIndex)
			hdPath, _ := cmd.Flags().GetString(flagHDPath)

			// Use default HD path if not specified
			if hdPath == "" {
				hdPath = hd.CreateHDPath(coinType, account, index).String()
			}

			// Validate HD path
			if _, err := hd.NewParamsFromPath(hdPath); err != nil {
				return fmt.Errorf("invalid HD path '%s': %w", hdPath, err)
			}

			var mnemonic string
			var bip39Passphrase string

			// Recovery mode: import from existing mnemonic
			if recover {
				mnemonic, err = input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return fmt.Errorf("failed to read mnemonic: %w", err)
				}

				if !bip39.IsMnemonicValid(mnemonic) {
					return fmt.Errorf("invalid mnemonic")
				}
			}

			// Get BIP39 passphrase if in interactive mode or recovering
			if interactive || recover {
				bip39Passphrase, err = input.GetString("Enter your bip39 passphrase (leave empty if none)", inBuf)
				if err != nil {
					return fmt.Errorf("failed to read passphrase: %w", err)
				}
			}

			// Get key algorithm
			algo, _ := cmd.Flags().GetString(flagPubKeyType)
			keyringAlgos, _ := kb.SupportedAlgorithms()
			algoStr := hd.PubKeyType(algo)

			// Validate algorithm is supported
			algoInterface, err := keyring.NewSigningAlgoFromString(string(algoStr), keyringAlgos)
			if err != nil {
				return fmt.Errorf("unsupported signing algorithm '%s': %w", algo, err)
			}

			// Check for multisig
			multisigKeys, _ := cmd.Flags().GetStringSlice(flagMultisig)
			if len(multisigKeys) > 0 {
				return createMultisigKey(kb, name, multisigKeys)
			}

			// Dry run: just show what would be created
			if dryRun {
				cmd.Println("Dry run mode - no key will be created")
				cmd.Printf("Name: %s\n", name)
				cmd.Printf("Algorithm: %s\n", algoStr)
				cmd.Printf("HD Path: %s\n", hdPath)
				cmd.Printf("Coin Type: %d\n", coinType)
				cmd.Printf("Account: %d\n", account)
				cmd.Printf("Index: %d\n", index)
				return nil
			}

			// Generate or recover key
			var record *keyring.Record
			if recover {
				record, err = kb.NewAccount(name, mnemonic, bip39Passphrase, hdPath, algoInterface)
				if err != nil {
					return fmt.Errorf("failed to recover account: %w", err)
				}
			} else {
				record, mnemonic, err = kb.NewMnemonic(name, keyring.English, hdPath, bip39Passphrase, algoInterface)
				if err != nil {
					return fmt.Errorf("failed to create new mnemonic: %w", err)
				}
			}

			// Get address and public key
			addr, err := record.GetAddress()
			if err != nil {
				return fmt.Errorf("failed to get address: %w", err)
			}

			pubKey, err := record.GetPubKey()
			if err != nil {
				return fmt.Errorf("failed to get public key: %w", err)
			}

			// Output based on format
			outputFormat, _ := cmd.Flags().GetString(flags.FlagOutput)
			if outputFormat == "json" {
				keyOutput, err := keys.NewKeyOutput(record.Name, record.GetType(), addr, pubKey)
				if err != nil {
					return err
				}
				if !noBackup && !recover {
					keyOutput.Mnemonic = mnemonic
				}
				jsonOut, _ := json.MarshalIndent(keyOutput, "", "  ")
				cmd.Println(string(jsonOut))
			} else {
				keyOutput, err := keys.NewKeyOutput(record.Name, record.GetType(), addr, pubKey)
				if err != nil {
					return err
				}

				cmd.Println()
				cmd.Println("- name:", keyOutput.Name)
				cmd.Println("  type:", keyOutput.Type)
				cmd.Println("  address:", keyOutput.Address)
				cmd.Println("  pubkey:", keyOutput.PubKey)

				// Show mnemonic for backup (only for new keys)
				if !noBackup && !recover {
					cmd.Println()
					cmd.Println("**Important** write this mnemonic phrase in a safe place.")
					cmd.Println("It is the only way to recover your account if you ever forget your password.")
					cmd.Println()
					cmd.Println(mnemonic)
				}
			}

			return nil
		},
	}

	cmd.Flags().Bool(flagInteractive, false, "Interactively prompt for BIP39 passphrase and other options")
	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of generating a new one")
	cmd.Flags().Bool(flagNoBackup, false, "Don't print out seed phrase (if others are watching the terminal)")
	cmd.Flags().Uint32(flagCoinType, sdk.CoinType, "coin type number for HD derivation")
	cmd.Flags().Uint32(flagAccount, 0, "Account number for HD derivation")
	cmd.Flags().Uint32(flagIndex, 0, "Address index number for HD derivation")
	cmd.Flags().String(flagHDPath, "", "Manual HD path (BIP32 format, e.g. m/44'/118'/0'/0/0)")
	cmd.Flags().String(flagPubKeyType, string(hd.Secp256k1Type), "Key signing algorithm to generate keys (secp256k1|ed25519|sr25519)")
	cmd.Flags().StringSlice(flagMultisig, nil, "List of key names to create a multisig account")
	cmd.Flags().Bool(flagDryRun, false, "Perform a simulation of the key generation without actually creating it")

	return cmd
}

// keysListCmd returns the keys list command
func keysListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all keys in the keyring",
		Long: `Return a list of all public keys stored by this key manager
along with their associated name, type, address and public key.

Keys are stored in the keyring backend configured with --keyring-backend flag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			cdc := clientCtx.Codec

			// Get keyring backend and initialize keyring
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			homeDir := viper.GetString(flags.FlagHome)
			keyringDir := viper.GetString(flags.FlagKeyringDir)
			if keyringDir == "" {
				keyringDir = homeDir
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())
			kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, keyringDir, inBuf, cdc)
			if err != nil {
				return fmt.Errorf("failed to initialize keyring: %w", err)
			}

			// Get all keys
			records, err := kb.List()
			if err != nil {
				return fmt.Errorf("failed to list keys: %w", err)
			}

			if len(records) == 0 {
				cmd.Println("No keys found in keyring")
				return nil
			}

			// Get output format
			outputFormat, _ := cmd.Flags().GetString(flags.FlagOutput)

			if outputFormat == "json" {
				return outputKeysJSON(cmd.OutOrStdout(), records)
			}

			return outputKeysText(cmd.OutOrStdout(), records)
		},
	}

	return cmd
}

// keysShowCmd returns the keys show command
func keysShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show key information by name or address",
		Long: `Display details for a specific key stored in the keyring.

The key can be referenced by name. The output can be configured to show
only the address, public key, or full details.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			cdc := clientCtx.Codec

			keyNameOrAddress := args[0]

			// Get keyring backend and initialize keyring
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			homeDir := viper.GetString(flags.FlagHome)
			keyringDir := viper.GetString(flags.FlagKeyringDir)
			if keyringDir == "" {
				keyringDir = homeDir
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())
			kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, keyringDir, inBuf, cdc)
			if err != nil {
				return fmt.Errorf("failed to initialize keyring: %w", err)
			}

			// Try to get key by name first
			record, err := kb.Key(keyNameOrAddress)
			if err != nil {
				// Try to parse as address and find by address
				if addr, err2 := sdk.AccAddressFromBech32(keyNameOrAddress); err2 == nil {
					record, err = kb.KeyByAddress(addr)
					if err != nil {
						return fmt.Errorf("key not found for address '%s': %w", keyNameOrAddress, err)
					}
				} else {
					return fmt.Errorf("key '%s' not found in keyring: %w", keyNameOrAddress, err)
				}
			}

			// Get flags for specific output
			showAddress, _ := cmd.Flags().GetBool(flagShowAddress)
			showPubKey, _ := cmd.Flags().GetBool(flagShowPubKey)
			bechPrefix, _ := cmd.Flags().GetString(flagShowBech32)
			showDevice, _ := cmd.Flags().GetBool(flagShowDevice)

			// Get public key
			pubKey, err := record.GetPubKey()
			if err != nil {
				return fmt.Errorf("failed to get public key: %w", err)
			}

			// Get the appropriate key output function based on bech32 prefix
			var bechKeyOutFn func(*keyring.Record) (keys.KeyOutput, error)
			switch bechPrefix {
			case "val":
				bechKeyOutFn = keys.MkValKeyOutput
			case "cons":
				bechKeyOutFn = keys.MkConsKeyOutput
			default:
				bechKeyOutFn = keys.MkAccKeyOutput
			}

			keyOutput, err := bechKeyOutFn(record)
			if err != nil {
				return err
			}

			// Handle specific output requests
			if showAddress {
				cmd.Println(keyOutput.Address)
				return nil
			}

			if showPubKey {
				cmd.Println(hex.EncodeToString(pubKey.Bytes()))
				return nil
			}

			// Get output format
			outputFormat, _ := cmd.Flags().GetString(flags.FlagOutput)

			if outputFormat == "json" {
				jsonOut, _ := json.MarshalIndent(keyOutput, "", "  ")
				cmd.Println(string(jsonOut))
				return nil
			}

			// Text output
			cmd.Println("- name:", keyOutput.Name)
			cmd.Println("  type:", keyOutput.Type)
			cmd.Println("  address:", keyOutput.Address)
			cmd.Println("  pubkey:", keyOutput.PubKey)

			// Note: showDevice functionality would require ledger support
			_ = showDevice

			return nil
		},
	}

	cmd.Flags().BoolP(flagShowAddress, "a", false, "Output only the address")
	cmd.Flags().BoolP(flagShowPubKey, "p", false, "Output only the public key")
	cmd.Flags().String(flagShowBech32, sdk.PrefixAccount, "Output the address with a specific Bech32 prefix (acc|val|cons)")
	cmd.Flags().BoolP(flagShowDevice, "d", false, "Show hardware device information (if applicable)")

	return cmd
}

// keysDeleteCmd returns the keys delete command
func keysDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a key from the keyring",
		Long: `Delete a key from the keyring by name.

This is a destructive operation that cannot be undone unless you have
backed up your mnemonic phrase. You will be prompted for confirmation
unless the --yes flag is used.

WARNING: If you delete a key without backing up the mnemonic, you will
permanently lose access to any assets controlled by that key.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			cdc := clientCtx.Codec

			keyName := args[0]

			if keyName == "" {
				return fmt.Errorf("key name cannot be empty")
			}

			// Get keyring backend and initialize keyring
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			homeDir := viper.GetString(flags.FlagHome)
			keyringDir := viper.GetString(flags.FlagKeyringDir)
			if keyringDir == "" {
				keyringDir = homeDir
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())
			kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, keyringDir, inBuf, cdc)
			if err != nil {
				return fmt.Errorf("failed to initialize keyring: %w", err)
			}

			// Check if key exists
			record, err := kb.Key(keyName)
			if err != nil {
				return fmt.Errorf("key '%s' not found in keyring: %w", keyName, err)
			}

			// Get address for confirmation
			addr, err := record.GetAddress()
			if err != nil {
				return fmt.Errorf("failed to get address: %w", err)
			}

			// Check for --yes flag to skip confirmation
			skipConfirm, _ := cmd.Flags().GetBool("yes")
			dryRun, _ := cmd.Flags().GetBool(flagDryRun)

			if dryRun {
				cmd.Printf("Dry run mode - would delete key '%s' (address: %s)\n", keyName, addr.String())
				return nil
			}

			if !skipConfirm {
				// Prompt for confirmation
				cmd.Printf("WARNING: This will permanently delete the key '%s' (address: %s)\n", keyName, addr.String())
				cmd.Println("Make sure you have backed up your mnemonic phrase before proceeding.")
				cmd.Println()

				response, err := input.GetString("Type the key name to confirm deletion", inBuf)
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				if strings.TrimSpace(response) != keyName {
					return fmt.Errorf("key name confirmation does not match, aborting deletion")
				}
			}

			// Delete the key
			if err := kb.Delete(keyName); err != nil {
				return fmt.Errorf("failed to delete key '%s': %w", keyName, err)
			}

			cmd.Printf("Key '%s' (address: %s) deleted successfully\n", keyName, addr.String())
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool(flagDryRun, false, "Perform a simulation without actually deleting the key")

	return cmd
}

// keysRenameCmd returns the keys rename command
func keysRenameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename a key in the keyring",
		Long: `Rename an existing key in the keyring.

The key's address and public key remain unchanged; only the local name is modified.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			cdc := clientCtx.Codec

			oldName := args[0]
			newName := args[1]

			if oldName == "" || newName == "" {
				return fmt.Errorf("both old and new key names must be provided")
			}

			// Get keyring backend and initialize keyring
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			homeDir := viper.GetString(flags.FlagHome)
			keyringDir := viper.GetString(flags.FlagKeyringDir)
			if keyringDir == "" {
				keyringDir = homeDir
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())
			kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, keyringDir, inBuf, cdc)
			if err != nil {
				return fmt.Errorf("failed to initialize keyring: %w", err)
			}

			// Check if old key exists
			_, err = kb.Key(oldName)
			if err != nil {
				return fmt.Errorf("key '%s' not found in keyring: %w", oldName, err)
			}

			// Check if new name is already taken
			_, err = kb.Key(newName)
			if err == nil {
				return fmt.Errorf("key with name '%s' already exists", newName)
			}

			// Rename the key
			if err := kb.Rename(oldName, newName); err != nil {
				return fmt.Errorf("failed to rename key: %w", err)
			}

			cmd.Printf("Key renamed from '%s' to '%s' successfully\n", oldName, newName)
			return nil
		},
	}

	return cmd
}

// Helper functions

// createMultisigKey creates a multisig key from multiple existing keys
func createMultisigKey(kb keyring.Keyring, name string, keyNames []string) error {
	if len(keyNames) < 2 {
		return fmt.Errorf("multisig requires at least 2 keys")
	}

	var pks []cryptotypes.PubKey
	for _, keyName := range keyNames {
		record, err := kb.Key(keyName)
		if err != nil {
			return fmt.Errorf("key '%s' not found: %w", keyName, err)
		}
		pk, err := record.GetPubKey()
		if err != nil {
			return fmt.Errorf("failed to get public key for '%s': %w", keyName, err)
		}
		pks = append(pks, pk)
	}

	// Create multisig public key
	// Note: This requires additional implementation based on your multisig threshold
	return fmt.Errorf("multisig creation not yet fully implemented - requires threshold specification")
}

// outputKeysJSON outputs keys in JSON format
func outputKeysJSON(w io.Writer, records []*keyring.Record) error {
	keyOutputs, err := keys.MkAccKeysOutput(records)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(keyOutputs)
}

// outputKeysText outputs keys in text format
func outputKeysText(w io.Writer, records []*keyring.Record) error {
	for _, record := range records {
		keyOutput, err := keys.MkAccKeyOutput(record)
		if err != nil {
			continue
		}

		fmt.Fprintf(w, "- name: %s\n", keyOutput.Name)
		fmt.Fprintf(w, "  type: %s\n", keyOutput.Type)
		fmt.Fprintf(w, "  address: %s\n", keyOutput.Address)
		fmt.Fprintf(w, "  pubkey: %s\n", keyOutput.PubKey)
		fmt.Fprintln(w)
	}
	return nil
}

// formatPubKey formats a public key for display
func formatPubKey(cdc codec.Codec, pk cryptotypes.PubKey) (string, error) {
	apk, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return "", err
	}
	bz, err := codec.ProtoMarshalJSON(apk, nil)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
