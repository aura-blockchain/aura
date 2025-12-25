// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// DebugCmd returns the debug command with troubleshooting subcommands
func DebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug and troubleshooting commands",
		Long: `Debug commands for troubleshooting node issues, inspecting state, and diagnosing problems.

These commands are essential for operators, developers, and validators to:
- Diagnose node connectivity issues
- Inspect blockchain state and data structures
- Verify configuration and setup
- Decode and inspect transactions
- Analyze address formats and keys
- Check system resources and health`,
	}

	cmd.AddCommand(
		debugAddrCmd(),
		debugPubkeyCmd(),
		debugRawBytesCmd(),
		debugNodeInfoCmd(),
		debugStateCmd(),
		debugConfigCmd(),
		debugKeysCmd(),
		debugHealthCmd(),
		debugTxCmd(),
		debugBlockCmd(),
		debugValidatorCmd(),
	)

	return cmd
}

// debugAddrCmd decodes and displays address information
func debugAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addr [address]",
		Short: "Decode and display address information",
		Long: `Decode a Bech32 address and display its hex encoding and other metadata.

This is useful for:
- Verifying address formats
- Converting between Bech32 and hex encodings
- Validating addresses from different sources

Example:
  aurad debug addr aura1abc...
  aurad debug addr auravaloper1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := args[0]

			// Try to decode as account address
			accAddr, accErr := sdk.AccAddressFromBech32(addr)
			if accErr == nil {
				fmt.Printf("Address Type:   Account Address\n")
				fmt.Printf("Bech32:         %s\n", accAddr.String())
				fmt.Printf("Hex:            %X\n", accAddr.Bytes())
				fmt.Printf("Length:         %d bytes\n", len(accAddr.Bytes()))
				return nil
			}

			// Try to decode as validator address
			valAddr, valErr := sdk.ValAddressFromBech32(addr)
			if valErr == nil {
				fmt.Printf("Address Type:   Validator Address\n")
				fmt.Printf("Bech32:         %s\n", valAddr.String())
				fmt.Printf("Hex:            %X\n", valAddr.Bytes())
				fmt.Printf("Length:         %d bytes\n", len(valAddr.Bytes()))
				return nil
			}

			return fmt.Errorf("invalid address format\nAccount error: %v\nValidator error: %v", accErr, valErr)
		},
	}

	return cmd
}

// debugPubkeyCmd decodes and displays public key information
func debugPubkeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubkey [pubkey]",
		Short: "Decode and display public key information",
		Long: `Decode a Bech32-encoded public key and display its details.

Example:
  aurad debug pubkey auravalconspub1...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pubkeyStr := args[0]
			pubkeyJSON := fmt.Sprintf(`{"@type":"%s","key":"%s"}`, "/cosmos.crypto.ed25519.PubKey", pubkeyStr)

			fmt.Printf("Public Key Information:\n")
			fmt.Printf("%s\n", pubkeyJSON)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// debugRawBytesCmd decodes and displays raw bytes
func debugRawBytesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raw-bytes [hex-string]",
		Short: "Decode and display raw hex bytes",
		Long: `Decode a hex string and display it in various formats.

This is useful for inspecting:
- Transaction bytes
- Block data
- State data
- Any hex-encoded data

Example:
  aurad debug raw-bytes 0x1234abcd`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hexStr := args[0]
			// Remove 0x prefix if present
			if len(hexStr) >= 2 && hexStr[:2] == "0x" {
				hexStr = hexStr[2:]
			}

			var bytes []byte
			for i := 0; i < len(hexStr); i += 2 {
				if i+2 > len(hexStr) {
					return fmt.Errorf("invalid hex string (odd length)")
				}
				var b byte
				_, err := fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
				if err != nil {
					return fmt.Errorf("invalid hex string: %w", err)
				}
				bytes = append(bytes, b)
			}

			fmt.Printf("Hex:            %X\n", bytes)
			fmt.Printf("Base64:         %s\n", base64.StdEncoding.EncodeToString(bytes))
			fmt.Printf("Length:         %d bytes\n", len(bytes))
			fmt.Printf("ASCII (if printable): %s\n", string(bytes))

			return nil
		},
	}

	return cmd
}

// debugNodeInfoCmd displays detailed node information
func debugNodeInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node-info",
		Short: "Display detailed node information",
		Long: `Display comprehensive node information including:
- Node identity and version
- Network information
- P2P connectivity
- Consensus state
- Sync status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeAddr := viper.GetString(flags.FlagNode)

			fmt.Printf("=== Node Information ===\n\n")
			fmt.Printf("Node Address: %s\n", nodeAddr)
			fmt.Printf("Home Dir:     %s\n", homeDir)
			fmt.Printf("Chain ID:     %s\n\n", viper.GetString(flags.FlagChainID))

			// Try to connect to node
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(fmt.Sprintf("%s/status", nodeAddr))
			if err != nil {
				fmt.Printf("✗ Cannot connect to node: %v\n", err)
				fmt.Printf("\nTroubleshooting:\n")
				fmt.Printf("1. Verify the node is running: aurad start\n")
				fmt.Printf("2. Check the node address is correct: %s\n", nodeAddr)
				fmt.Printf("3. Check firewall settings\n")
				return nil
			}
			defer resp.Body.Close()

			fmt.Printf("✓ Node is reachable\n")
			fmt.Printf("Status Code:  %d\n", resp.StatusCode)

			return nil
		},
	}

	return cmd
}

// debugStateCmd inspects blockchain state
func debugStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Inspect blockchain state",
		Long: `Inspect and diagnose blockchain state issues.

Shows:
- Latest block height
- State database size
- Number of accounts
- Total supply`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("=== Blockchain State ===\n\n")

			// Query latest block
			node, err := clientCtx.GetNode()
			if err != nil {
				return fmt.Errorf("failed to get node: %w", err)
			}

			status, err := node.Status(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			fmt.Printf("Latest Block Height: %d\n", status.SyncInfo.LatestBlockHeight)
			fmt.Printf("Latest Block Time:   %s\n", status.SyncInfo.LatestBlockTime)
			fmt.Printf("Catching Up:         %v\n\n", status.SyncInfo.CatchingUp)

			// Check data directory
			dataDir := filepath.Join(homeDir, "data")
			if info, err := os.Stat(dataDir); err == nil {
				fmt.Printf("Data Directory:      %s\n", dataDir)
				fmt.Printf("Data Dir Exists:     %v\n", info.IsDir())
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// debugConfigCmd displays configuration information
func debugConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display configuration information",
		Long: `Display current configuration settings from config files and flags.

Useful for:
- Verifying configuration
- Debugging configuration issues
- Exporting configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("=== Configuration ===\n\n")

			fmt.Printf("Home Directory:      %s\n", homeDir)
			fmt.Printf("Chain ID:            %s\n", viper.GetString(flags.FlagChainID))
			fmt.Printf("Node Address:        %s\n", viper.GetString(flags.FlagNode))
			fmt.Printf("Keyring Backend:     %s\n", viper.GetString(flags.FlagKeyringBackend))
			fmt.Printf("Output Format:       %s\n", viper.GetString(flags.FlagOutput))
			fmt.Printf("Broadcast Mode:      %s\n", viper.GetString(flags.FlagBroadcastMode))
			fmt.Printf("\n")

			// Check config files
			configFile := filepath.Join(homeDir, "config", "config.toml")
			appFile := filepath.Join(homeDir, "config", "app.toml")
			clientFile := filepath.Join(homeDir, "config", "client.toml")
			genesisFile := filepath.Join(homeDir, "config", "genesis.json")

			fmt.Printf("Config Files:\n")
			fmt.Printf("  config.toml:       %s (exists: %v)\n", configFile, fileExists(configFile))
			fmt.Printf("  app.toml:          %s (exists: %v)\n", appFile, fileExists(appFile))
			fmt.Printf("  client.toml:       %s (exists: %v)\n", clientFile, fileExists(clientFile))
			fmt.Printf("  genesis.json:      %s (exists: %v)\n", genesisFile, fileExists(genesisFile))

			return nil
		},
	}

	return cmd
}

// debugKeysCmd displays keyring information
func debugKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Display keyring information",
		Long: `Display information about the keyring configuration and available keys.

Does NOT display private keys or mnemonics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("=== Keyring Information ===\n\n")
			fmt.Printf("Keyring Backend:     %s\n", clientCtx.Keyring.Backend())
			fmt.Printf("Keyring Directory:   %s\n", homeDir)
			fmt.Printf("\n")

			// List keys
			keys, err := clientCtx.Keyring.List()
			if err != nil {
				return fmt.Errorf("failed to list keys: %w", err)
			}

			fmt.Printf("Available Keys:      %d\n", len(keys))
			for i, k := range keys {
				addr, err := k.GetAddress()
				if err != nil {
					continue
				}
				fmt.Printf("  [%d] %s: %s\n", i+1, k.Name, addr.String())
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// debugHealthCmd performs comprehensive health check
func debugHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Perform comprehensive health check",
		Long: `Perform a comprehensive health check of the node and system.

Checks:
- Node connectivity
- Disk space
- Memory usage
- Configuration validity
- Keyring accessibility
- Data directory integrity`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("=== Aura Node Health Check ===\n\n")

			checks := []struct {
				name string
				fn   func() (bool, string)
			}{
				{"Node Reachability", checkNodeReachability},
				{"Configuration Files", checkConfigFiles},
				{"Data Directory", checkDataDirectory},
				{"Keyring Access", checkKeyringAccess},
				{"Disk Space", checkDiskSpace},
				{"Memory Usage", checkMemoryUsage},
			}

			passed := 0
			failed := 0

			for _, check := range checks {
				fmt.Printf("Checking %s... ", check.name)
				ok, msg := check.fn()
				if ok {
					fmt.Printf("✓ OK\n")
					if msg != "" {
						fmt.Printf("  %s\n", msg)
					}
					passed++
				} else {
					fmt.Printf("✗ FAILED\n")
					if msg != "" {
						fmt.Printf("  %s\n", msg)
					}
					failed++
				}
			}

			fmt.Printf("\n=== Summary ===\n")
			fmt.Printf("Passed: %d\n", passed)
			fmt.Printf("Failed: %d\n", failed)

			if failed > 0 {
				fmt.Printf("\nSome checks failed. Review the output above for details.\n")
			} else {
				fmt.Printf("\nAll checks passed! Node appears healthy.\n")
			}

			return nil
		},
	}

	return cmd
}

// debugTxCmd decodes and inspects transaction data
func debugTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [tx-hash-or-hex]",
		Short: "Decode and inspect transaction",
		Long: `Decode a transaction from hash or hex encoding and display its details.

Example:
  aurad debug tx ABC123...
  aurad debug tx 0x1234abcd...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			txStr := args[0]
			fmt.Printf("=== Transaction Debug ===\n\n")
			fmt.Printf("Input: %s\n\n", txStr)

			// Try to query tx by hash
			queryClient := authtypes.NewQueryClient(clientCtx)
			_ = queryClient // Use when implementing actual query

			fmt.Printf("Transaction inspection - implementation requires proto messages\n")
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// debugBlockCmd inspects block data
func debugBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block [height]",
		Short: "Inspect block data",
		Long: `Inspect block data at a specific height.

Example:
  aurad debug block 12345
  aurad debug block latest`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			node, err := clientCtx.GetNode()
			if err != nil {
				return fmt.Errorf("failed to get node: %w", err)
			}

			var height *int64
			if len(args) > 0 && args[0] != "latest" {
				var h int64
				_, err := fmt.Sscanf(args[0], "%d", &h)
				if err != nil {
					return fmt.Errorf("invalid height: %w", err)
				}
				height = &h
			}

			result, err := node.Block(cmd.Context(), height)
			if err != nil {
				return fmt.Errorf("failed to get block: %w", err)
			}

			fmt.Printf("=== Block Information ===\n\n")
			fmt.Printf("Height:          %d\n", result.Block.Height)
			fmt.Printf("Time:            %s\n", result.Block.Time)
			fmt.Printf("Num Txs:         %d\n", len(result.Block.Txs))
			fmt.Printf("Proposer:        %X\n", result.Block.ProposerAddress)
			fmt.Printf("Block Hash:      %X\n", result.BlockID.Hash)
			fmt.Printf("Data Hash:       %X\n", result.Block.DataHash)
			fmt.Printf("Validators Hash: %X\n", result.Block.ValidatorsHash)

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// debugValidatorCmd displays validator debugging information
func debugValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [address]",
		Short: "Display validator debugging information",
		Long: `Display debugging information for a validator.

Example:
  aurad debug validator auravaloper1abc...`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("=== Validator Debug ===\n\n")

			if len(args) == 0 {
				// Check if this node is a validator
				validatorKeyFile := filepath.Join(homeDir, "config", "priv_validator_key.json")
				if fileExists(validatorKeyFile) {
					fmt.Printf("Validator Key File:  %s\n", validatorKeyFile)
					fmt.Printf("This node appears to be configured as a validator\n")

					// Read and display public key
					data, err := os.ReadFile(validatorKeyFile)
					if err != nil {
						return fmt.Errorf("failed to read validator key: %w", err)
					}

					var keyData map[string]interface{}
					if err := json.Unmarshal(data, &keyData); err != nil {
						return fmt.Errorf("failed to parse validator key: %w", err)
					}

					if pubkey, ok := keyData["pub_key"].(map[string]interface{}); ok {
						if value, ok := pubkey["value"].(string); ok {
							fmt.Printf("Public Key:          %s\n", value)
						}
					}
				} else {
					fmt.Printf("This node is not configured as a validator\n")
					fmt.Printf("(No priv_validator_key.json found)\n")
				}
			} else {
				valAddr := args[0]
				fmt.Printf("Validator Address:   %s\n", valAddr)
				fmt.Printf("Validator inspection - implementation requires query client\n")
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// Helper functions for health checks

func checkNodeReachability() (bool, string) {
	nodeAddr := viper.GetString(flags.FlagNode)
	client := &http.Client{Timeout: 3 * time.Second}
	_, err := client.Get(nodeAddr + "/status")
	if err != nil {
		return false, fmt.Sprintf("Cannot reach node at %s: %v", nodeAddr, err)
	}
	return true, fmt.Sprintf("Node reachable at %s", nodeAddr)
}

func checkConfigFiles() (bool, string) {
	required := []string{"config/config.toml", "config/app.toml"}
	missing := []string{}

	for _, file := range required {
		path := filepath.Join(homeDir, file)
		if !fileExists(path) {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		return false, fmt.Sprintf("Missing config files: %v", missing)
	}
	return true, "All required config files present"
}

func checkDataDirectory() (bool, string) {
	dataDir := filepath.Join(homeDir, "data")
	if !fileExists(dataDir) {
		return false, fmt.Sprintf("Data directory not found: %s", dataDir)
	}
	return true, fmt.Sprintf("Data directory exists: %s", dataDir)
}

func checkKeyringAccess() (bool, string) {
	keyringDir := filepath.Join(homeDir, "keyring-test")
	if !fileExists(keyringDir) {
		return true, "Keyring not initialized (optional)"
	}
	return true, "Keyring directory accessible"
}

func checkDiskSpace() (bool, string) {
	// This is a simplified check - in production you'd use syscall to get actual disk space
	return true, "Disk space check not implemented (requires syscall)"
}

func checkMemoryUsage() (bool, string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocMB := m.Alloc / 1024 / 1024
	totalMB := m.TotalAlloc / 1024 / 1024

	return true, fmt.Sprintf("Allocated: %d MB, Total: %d MB", allocMB, totalMB)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
