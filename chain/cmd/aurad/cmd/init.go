package cmd

import (
	"bufio"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aequitas/aura/chain/app"
	"github.com/aequitas/aura/chain/cmd/aurad/cmd/security"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const (
	// Bech32 address prefixes for AURA chain
	Bech32MainPrefix      = "aura"
	Bech32ValidatorPrefix = "auravaloper"
	Bech32ConsensusPrefix = "auravalcons"

	// BIP44 derivation path for Cosmos-based chains
	// m/44'/118'/0'/0/0 is the standard Cosmos derivation path
	DefaultHDPath = "m/44'/118'/0'/0/0"

	// Mnemonic entropy bits (256 = 24 words, 128 = 12 words)
	MnemonicEntropyBits = 256

	// Default token amounts for genesis
	DefaultValidatorTokens = "100000000000" // 100k stake
	DefaultBondedTokens    = "90000000000"  // 90k stake bonded
	DefaultValidatorPower  = "900000"
)

// ValidatorKeyInfo contains all derived addresses from validator key
type ValidatorKeyInfo struct {
	// Hex-encoded consensus address (first 20 bytes of SHA256 of pubkey)
	ConsensusAddressHex string
	// Bech32-encoded account address with "aura" prefix
	AccountAddress string
	// Bech32-encoded validator operator address with "auravaloper" prefix
	OperatorAddress string
	// Bech32-encoded consensus address with "auravalcons" prefix
	ConsensusAddressBech32 string
	// Base64-encoded public key
	PublicKeyBase64 string
	// Raw public key bytes
	PublicKeyBytes []byte
}

// InitCmd returns the init command for the Aura daemon
func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize the Aura node configuration",
		Long: `Initialize the Aura node by creating the necessary configuration files
and directories in the home directory.

This command will:
  1. Generate a new validator key with a 24-word BIP39 mnemonic
  2. Create config/config.toml: Node configuration
  3. Create config/app.toml: Application configuration
  4. Create config/genesis.json: Genesis state with validator
  5. Create data/: Data directory for blockchain state

IMPORTANT: Save the mnemonic phrase displayed during initialization.
It is the ONLY way to recover your validator key.

The moniker is a human-readable name for the node.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return initNode(cmd, args)
		},
	}

	// Add flags
	cmd.Flags().String(FlagChainID, DefaultChainID, "chain ID for the network")
	cmd.Flags().String(FlagMoniker, DefaultMoniker, "moniker (name) for this node")
	cmd.Flags().Bool("recover", false, "recover validator key from existing mnemonic")
	cmd.Flags().BoolP("yes", "y", false, "skip mnemonic confirmation prompt (non-interactive mode)")

	// Bind flags to viper with error checking
	if flag := cmd.Flags().Lookup(FlagChainID); flag != nil {
		if err := viper.BindPFlag(FlagChainID, flag); err != nil {
			// Log but don't fail - flag will use default
			fmt.Fprintf(os.Stderr, "Warning: failed to bind %s flag: %v\n", FlagChainID, err)
		}
	}
	if flag := cmd.Flags().Lookup(FlagMoniker); flag != nil {
		if err := viper.BindPFlag(FlagMoniker, flag); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to bind %s flag: %v\n", FlagMoniker, err)
		}
	}

	return cmd
}

// initNode initializes the node configuration
func initNode(cmd *cobra.Command, args []string) error {
	// Get security logger
	logger := GetSecurityLogger()

	// Get moniker from args or flag
	moniker := viper.GetString(FlagMoniker)
	if len(args) > 0 {
		moniker = args[0]
	}

	chainID := viper.GetString(FlagChainID)
	homeDir := GetHomeDir()
	recover, _ := cmd.Flags().GetBool("recover")
	skipConfirmation, _ := cmd.Flags().GetBool("yes")

	// Validate inputs
	inputValidator := security.NewInputValidator(logger)

	if err := inputValidator.ValidateChainID(chainID); err != nil {
		return fmt.Errorf("invalid chain ID: %w", err)
	}

	if err := inputValidator.ValidateMoniker(moniker); err != nil {
		return fmt.Errorf("invalid moniker: %w", err)
	}

	fmt.Printf("Initializing Aura node...\n")
	fmt.Printf("Chain ID: %s\n", chainID)
	fmt.Printf("Moniker: %s\n", moniker)
	fmt.Printf("Home: %s\n", homeDir)

	// Create directory structure
	if err := createDirectories(homeDir, logger); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate or recover private validator key with BIP39 mnemonic
	keyInfo, err := createPrivateValidatorWithMnemonic(homeDir, recover, skipConfirmation)
	if err != nil {
		return fmt.Errorf("failed to create private validator: %w", err)
	}

	// Create config files (genesis will include validator with proper addresses)
	if err := createConfigFilesWithValidator(homeDir, chainID, moniker, keyInfo, logger); err != nil {
		return fmt.Errorf("failed to create config files: %w", err)
	}

	fmt.Printf("\nNode initialized successfully!\n")
	fmt.Printf("Configuration files created in: %s\n", homeDir)
	fmt.Printf("\nValidator Addresses:\n")
	fmt.Printf("  Account:   %s\n", keyInfo.AccountAddress)
	fmt.Printf("  Operator:  %s\n", keyInfo.OperatorAddress)
	fmt.Printf("  Consensus: %s\n", keyInfo.ConsensusAddressBech32)
	fmt.Printf("\nTo start the node, run:\n  aurad start\n")

	return nil
}

// createPrivateValidatorWithMnemonic generates a private validator key using BIP39 mnemonic
// or recovers from an existing mnemonic if recover is true.
func createPrivateValidatorWithMnemonic(homeDir string, recover bool, skipConfirmation bool) (*ValidatorKeyInfo, error) {
	keyFile := filepath.Join(homeDir, "config", "priv_validator_key.json")
	stateFile := filepath.Join(homeDir, "data", "priv_validator_state.json")

	// Check if key already exists
	if _, err := os.Stat(keyFile); err == nil {
		fmt.Printf("Private validator key already exists: %s\n", keyFile)
		// Read and parse existing key to return key info
		return loadValidatorKeyInfo(keyFile)
	}

	var mnemonic string
	var err error

	if recover {
		// Read mnemonic from stdin using bufio for proper line handling
		mnemonic, err = readMnemonicFromStdin()
		if err != nil {
			return nil, fmt.Errorf("failed to read mnemonic: %w", err)
		}

		// Validate the mnemonic before proceeding
		if !bip39.IsMnemonicValid(mnemonic) {
			return nil, fmt.Errorf("invalid mnemonic phrase: must be a valid BIP39 mnemonic (12, 15, 18, 21, or 24 words)")
		}

		fmt.Println("\nMnemonic validated successfully. Recovering validator key...")
	} else {
		// Generate new mnemonic
		mnemonic, err = generateMnemonic()
		if err != nil {
			return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
		}

		// Display mnemonic with security warnings
		displayMnemonicSecurely(mnemonic, skipConfirmation)
	}

	// Derive ed25519 key from mnemonic (same for both new and recovered)
	privKey, err := deriveEd25519KeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key from mnemonic: %w", err)
	}

	pubKey := privKey.PubKey()

	// Derive all address formats
	keyInfo, err := deriveValidatorAddresses(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive addresses: %w", err)
	}

	// Create key file content in CometBFT format
	keyContent := struct {
		Address string `json:"address"`
		PubKey  struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"pub_key"`
		PrivKey struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"priv_key"`
	}{
		Address: keyInfo.ConsensusAddressHex,
		PubKey: struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		}{
			Type:  "tendermint/PubKeyEd25519",
			Value: keyInfo.PublicKeyBase64,
		},
		PrivKey: struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		}{
			Type:  "tendermint/PrivKeyEd25519",
			Value: base64.StdEncoding.EncodeToString(privKey.Bytes()),
		},
	}

	keyData, err := json.MarshalIndent(keyContent, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	// Write key file with secure permissions (0600 = owner read/write only)
	if err := os.WriteFile(keyFile, keyData, 0o600); err != nil {
		return nil, fmt.Errorf("failed to write key file: %w", err)
	}
	fmt.Printf("Created private validator key: %s\n", keyFile)

	// Create state file
	stateContent := `{
  "height": "0",
  "round": 0,
  "step": 0
}`
	if err := os.WriteFile(stateFile, []byte(stateContent), 0o600); err != nil {
		return nil, fmt.Errorf("failed to write state file: %w", err)
	}
	fmt.Printf("Created private validator state: %s\n", stateFile)

	return keyInfo, nil
}

// generateMnemonic generates a new BIP39 mnemonic phrase (24 words)
func generateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(MnemonicEntropyBits)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// readMnemonicFromStdin reads a BIP39 mnemonic phrase from standard input.
// It prompts the user and reads a full line of input, then normalizes the words.
// Supports 12, 15, 18, 21, or 24 word mnemonics per BIP39 specification.
func readMnemonicFromStdin() (string, error) {
	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      MNEMONIC RECOVERY MODE                          ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║ Enter your BIP39 mnemonic phrase below.                              ║")
	fmt.Println("║ Words should be separated by spaces.                                 ║")
	fmt.Println("║ Valid lengths: 12, 15, 18, 21, or 24 words.                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Print("\nMnemonic: ")

	// Use bufio.Reader for robust stdin handling
	reader := bufio.NewReader(os.Stdin)

	// Read until newline - this properly handles multi-word input
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Normalize the mnemonic:
	// 1. Trim leading/trailing whitespace including the newline
	// 2. Convert to lowercase (BIP39 words are lowercase)
	// 3. Collapse multiple spaces into single spaces
	// 4. Trim again to remove any edge spaces
	mnemonic := strings.TrimSpace(line)
	mnemonic = strings.ToLower(mnemonic)

	// Collapse multiple spaces into single spaces
	words := strings.Fields(mnemonic)
	if len(words) == 0 {
		return "", fmt.Errorf("no mnemonic words provided")
	}

	// Validate word count per BIP39 specification
	// Valid word counts: 12 (128-bit), 15 (160-bit), 18 (192-bit), 21 (224-bit), 24 (256-bit)
	validWordCounts := map[int]bool{12: true, 15: true, 18: true, 21: true, 24: true}
	if !validWordCounts[len(words)] {
		return "", fmt.Errorf("invalid mnemonic length: got %d words, expected 12, 15, 18, 21, or 24", len(words))
	}

	// Rejoin with single spaces for consistent format
	normalizedMnemonic := strings.Join(words, " ")

	// Verify each word is in the BIP39 wordlist by checking if mnemonic is valid
	// The bip39 library will check this, but we provide a more helpful error message
	for i, word := range words {
		// Check if word contains only valid characters (a-z)
		for _, char := range word {
			if char < 'a' || char > 'z' {
				return "", fmt.Errorf("word %d (%q) contains invalid characters: BIP39 words must be lowercase letters only", i+1, word)
			}
		}
	}

	return normalizedMnemonic, nil
}

// deriveEd25519KeyFromMnemonic derives an ed25519 private key from a BIP39 mnemonic
// using SLIP-0010 derivation for ed25519 keys with Cosmos derivation path m/44'/118'/0'/0/0
func deriveEd25519KeyFromMnemonic(mnemonic string) (ed25519PrivKey, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic phrase")
	}

	// Generate seed from mnemonic (no passphrase for validator keys)
	seed := bip39.NewSeed(mnemonic, "")

	// SLIP-0010 ed25519 key derivation
	// Step 1: Generate master key using HMAC-SHA512 with key "ed25519 seed"
	masterKey, chainCode, err := slip0010MasterKeyGeneration(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	// Step 2: Derive through BIP44 path: m/44'/118'/0'/0/0
	// Each segment is hardened (indicated by ')
	path := []uint32{
		0x8000002C, // 44' (purpose)
		0x80000076, // 118' (coin type for Cosmos)
		0x80000000, // 0' (account)
		0x80000000, // 0' (change - hardened for ed25519)
		0x80000000, // 0' (address index - hardened for ed25519)
	}

	derivedKey := masterKey
	derivedChainCode := chainCode

	for _, segment := range path {
		derivedKey, derivedChainCode, err = slip0010DeriveChild(derivedKey, derivedChainCode, segment)
		if err != nil {
			return nil, fmt.Errorf("failed to derive child key: %w", err)
		}
	}

	return ed25519PrivKey(ed25519.NewKeyFromSeed(derivedKey)), nil
}

// slip0010MasterKeyGeneration generates the master key and chain code per SLIP-0010
func slip0010MasterKeyGeneration(seed []byte) ([]byte, []byte, error) {
	// HMAC-SHA512 with key "ed25519 seed"
	hmacKey := []byte("ed25519 seed")
	h := hmacSHA512(hmacKey, seed)

	// Split into key (first 32 bytes) and chain code (last 32 bytes)
	if len(h) != 64 {
		return nil, nil, fmt.Errorf("invalid HMAC output length")
	}

	return h[:32], h[32:], nil
}

// slip0010DeriveChild derives a child key per SLIP-0010 for ed25519
// For ed25519, only hardened derivation is supported (index >= 0x80000000)
func slip0010DeriveChild(parentKey, parentChainCode []byte, index uint32) ([]byte, []byte, error) {
	// Verify hardened derivation (required for ed25519)
	if index < 0x80000000 {
		return nil, nil, fmt.Errorf("ed25519 requires hardened derivation (index >= 0x80000000)")
	}

	// Data = 0x00 || parentKey || index (big-endian)
	data := make([]byte, 1+32+4)
	data[0] = 0x00
	copy(data[1:33], parentKey)
	data[33] = byte(index >> 24)
	data[34] = byte(index >> 16)
	data[35] = byte(index >> 8)
	data[36] = byte(index)

	// HMAC-SHA512 with parent chain code as key
	h := hmacSHA512(parentChainCode, data)

	if len(h) != 64 {
		return nil, nil, fmt.Errorf("invalid HMAC output length")
	}

	return h[:32], h[32:], nil
}

// hmacSHA512 computes HMAC-SHA512
func hmacSHA512(key, data []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// displayMnemonicSecurely displays the mnemonic with security warnings
func displayMnemonicSecurely(mnemonic string, skipConfirmation bool) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    IMPORTANT - SAVE YOUR MNEMONIC                    ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║ Write down these 24 words in order and store them securely.          ║")
	fmt.Println("║ This is the ONLY way to recover your validator key.                  ║")
	fmt.Println("║                                                                      ║")
	fmt.Println("║ DO NOT:                                                              ║")
	fmt.Println("║   - Store digitally (screenshots, files, emails)                     ║")
	fmt.Println("║   - Share with anyone                                                ║")
	fmt.Println("║   - Enter on websites                                                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	words := strings.Split(mnemonic, " ")
	for i, word := range words {
		fmt.Printf("  %2d. %s\n", i+1, word)
	}

	fmt.Println()
	if !skipConfirmation {
		fmt.Println("Press Enter after you have safely recorded your mnemonic...")
		fmt.Scanln()
	} else {
		fmt.Println("(Non-interactive mode: skipping mnemonic confirmation)")
	}
}

// deriveValidatorAddresses derives all address formats from an ed25519 public key
func deriveValidatorAddresses(pubKey ed25519PubKey) (*ValidatorKeyInfo, error) {
	// Get raw public key bytes (32 bytes for ed25519)
	pubKeyBytes := pubKey.Bytes()

	// Compute address bytes: first 20 bytes of SHA256(pubkey)
	hash := sha256.Sum256(pubKeyBytes)
	addrBytes := hash[:20]

	// Encode to various formats
	consensusHex := strings.ToUpper(hex.EncodeToString(addrBytes))
	pubKeyBase64 := base64.StdEncoding.EncodeToString(pubKeyBytes)

	// Encode to bech32 addresses
	accountAddr, err := bech32Encode(Bech32MainPrefix, addrBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encode account address: %w", err)
	}

	operatorAddr, err := bech32Encode(Bech32ValidatorPrefix, addrBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encode operator address: %w", err)
	}

	consensusAddr, err := bech32Encode(Bech32ConsensusPrefix, addrBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encode consensus address: %w", err)
	}

	return &ValidatorKeyInfo{
		ConsensusAddressHex:    consensusHex,
		AccountAddress:         accountAddr,
		OperatorAddress:        operatorAddr,
		ConsensusAddressBech32: consensusAddr,
		PublicKeyBase64:        pubKeyBase64,
		PublicKeyBytes:         pubKeyBytes,
	}, nil
}

// bech32Encode encodes bytes to a bech32 address with the given human-readable prefix
func bech32Encode(hrp string, data []byte) (string, error) {
	// Convert 8-bit bytes to 5-bit groups for bech32
	converted, err := convertBits(data, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("failed to convert bits: %w", err)
	}

	// Compute bech32 checksum
	checksum := bech32Checksum(hrp, converted)

	// Build result string
	result := hrp + "1"
	for _, b := range converted {
		result += string(bech32Charset[b])
	}
	for _, b := range checksum {
		result += string(bech32Charset[b])
	}

	return result, nil
}

// bech32Charset is the character set for bech32 encoding
const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// convertBits converts a byte slice from one bit-per-element to another
func convertBits(data []byte, fromBits, toBits uint, pad bool) ([]byte, error) {
	acc := uint(0)
	bits := uint(0)
	result := make([]byte, 0, len(data)*int(fromBits)/int(toBits)+1)
	maxv := uint((1 << toBits) - 1)

	for _, value := range data {
		acc = (acc << fromBits) | uint(value)
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			result = append(result, byte((acc>>bits)&maxv))
		}
	}

	if pad {
		if bits > 0 {
			result = append(result, byte((acc<<(toBits-bits))&maxv))
		}
	} else if bits >= fromBits || ((acc<<(toBits-bits))&maxv) != 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	return result, nil
}

// bech32Checksum computes the bech32 checksum
func bech32Checksum(hrp string, data []byte) []byte {
	values := bech32HrpExpand(hrp)
	values = append(values, data...)
	values = append(values, []byte{0, 0, 0, 0, 0, 0}...)
	polymod := bech32Polymod(values) ^ 1

	checksum := make([]byte, 6)
	for i := 0; i < 6; i++ {
		checksum[i] = byte((polymod >> uint(5*(5-i))) & 31)
	}
	return checksum
}

// bech32HrpExpand expands the HRP for checksum computation
func bech32HrpExpand(hrp string) []byte {
	result := make([]byte, len(hrp)*2+1)
	for i, c := range hrp {
		result[i] = byte(c >> 5)
		result[i+len(hrp)+1] = byte(c & 31)
	}
	result[len(hrp)] = 0
	return result
}

// bech32Polymod computes the bech32 polymod checksum
func bech32Polymod(values []byte) uint32 {
	generator := []uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := uint32(1)
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ uint32(v)
		for i := 0; i < 5; i++ {
			if (top>>uint(i))&1 == 1 {
				chk ^= generator[i]
			}
		}
	}
	return chk
}

// loadValidatorKeyInfo loads and parses an existing validator key file
func loadValidatorKeyInfo(keyFile string) (*ValidatorKeyInfo, error) {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	var keyContent struct {
		Address string `json:"address"`
		PubKey  struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"pub_key"`
	}

	if err := json.Unmarshal(data, &keyContent); err != nil {
		return nil, fmt.Errorf("failed to parse key file: %w", err)
	}

	// Decode public key from base64
	pubKeyBytes, err := base64.StdEncoding.DecodeString(keyContent.PubKey.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Derive addresses from public key
	return deriveValidatorAddresses(ed25519PubKey(pubKeyBytes))
}

// ed25519PrivKey wraps ed25519.PrivateKey with methods needed for CometBFT
type ed25519PrivKey ed25519.PrivateKey

func (k ed25519PrivKey) PubKey() ed25519PubKey {
	pubKey := ed25519.PrivateKey(k).Public()
	if pub, ok := pubKey.(ed25519.PublicKey); ok {
		return ed25519PubKey(pub)
	}
	// This should never happen with valid ed25519 keys
	return nil
}

func (k ed25519PrivKey) Bytes() []byte {
	return []byte(k)
}

// ed25519PubKey wraps ed25519.PublicKey with methods needed for CometBFT
type ed25519PubKey ed25519.PublicKey

func (k ed25519PubKey) Address() []byte {
	// CometBFT address is first 20 bytes of SHA256 hash of public key
	hash := sha256.Sum256(k)
	return hash[:20]
}

func (k ed25519PubKey) Bytes() []byte {
	return []byte(k)
}

// createDirectories creates the necessary directory structure with secure permissions
func createDirectories(homeDir string, logger security.Logger) error {
	dirs := []struct {
		path  string
		perms os.FileMode
	}{
		{filepath.Join(homeDir, "config"), security.ConfigDirPerms},
		{filepath.Join(homeDir, "config", "tls"), security.SecureDirPerms},
		{filepath.Join(homeDir, "data"), security.SecureDirPerms},
		{filepath.Join(homeDir, "keys"), security.SecureDirPerms},
		{filepath.Join(homeDir, "logs"), security.SecureDirPerms},
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir.path, dir.perms); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir.path, err)
		}

		// Explicitly set permissions (some systems may apply umask)
		if err := os.Chmod(dir.path, dir.perms); err != nil {
			return fmt.Errorf("failed to set permissions on %s: %w", dir.path, err)
		}

		fmt.Printf("Created directory: %s (permissions: %o)\n", dir.path, dir.perms)
	}

	return nil
}

// createConfigFilesWithValidator creates configuration files including genesis with validator
func createConfigFilesWithValidator(homeDir, chainID, moniker string, keyInfo *ValidatorKeyInfo, logger security.Logger) error {
	// Create config.toml
	configPath := filepath.Join(homeDir, "config", "config.toml")
	configContent := generateConfigTOML(chainID, moniker)
	if err := writeConfigFile(configPath, configContent, logger); err != nil {
		return err
	}

	// Create app.toml
	appConfigPath := filepath.Join(homeDir, "config", "app.toml")
	appConfigContent := generateAppTOML()
	if err := writeConfigFile(appConfigPath, appConfigContent, logger); err != nil {
		return err
	}

	// Create genesis.json with proper validator addresses
	genesisPath := filepath.Join(homeDir, "config", "genesis.json")
	genesisContent := generateCompleteGenesis(chainID, keyInfo)
	if err := writeConfigFile(genesisPath, genesisContent, logger); err != nil {
		return err
	}

	return nil
}

// writeConfigFile writes a configuration file with secure permissions
func writeConfigFile(path, content string, logger security.Logger) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Configuration file already exists, skipping: %s\n", path)
		return nil
	}

	// Write file with secure permissions
	if err := os.WriteFile(path, []byte(content), security.ConfigFilePerms); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}

	// Explicitly set permissions
	if err := os.Chmod(path, security.ConfigFilePerms); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	fmt.Printf("Created config file: %s (permissions: %o)\n", path, security.ConfigFilePerms)
	return nil
}

// generateConfigTOML generates the config.toml content
func generateConfigTOML(chainID, moniker string) string {
	return fmt.Sprintf(`# Aura Node Configuration
# Generated by aurad init

#######################################################################
###                           Base Configuration                    ###
#######################################################################

# The chain ID
chain-id = "%s"

# A custom human readable name for this node
moniker = "%s"

#######################################################################
###                           RPC Configuration                     ###
#######################################################################

[rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "tcp://127.0.0.1:26657"

# Maximum number of simultaneous connections
max_open_connections = 900

# Maximum number of unique clientIDs that can /subscribe
max_subscription_clients = 100

# Maximum number of unique queries a given client can /subscribe to
max_subscriptions_per_client = 5

#######################################################################
###                           P2P Configuration                     ###
#######################################################################

[p2p]

# Address to listen for incoming connections
laddr = "tcp://0.0.0.0:26656"

# Address to advertise to peers for them to dial
external_address = ""

# Comma separated list of seed nodes to connect to
seeds = ""

# Comma separated list of nodes to keep persistent connections to
persistent_peers = ""

# Maximum number of inbound peers
max_num_inbound_peers = 40

# Maximum number of outbound peers
max_num_outbound_peers = 10

# Set true to enable the peer-exchange reactor
pex = true

#######################################################################
###                       Consensus Configuration                   ###
#######################################################################

[consensus]

# How long we wait for a proposal block before prevoting nil
timeout_propose = "3s"

# How long we wait after receiving +2/3 prevotes for "anything"
timeout_prevote = "1s"

# How long we wait after receiving +2/3 precommits for "anything"
timeout_precommit = "1s"

# How long we wait after committing a block, before starting on the new height
timeout_commit = "5s"

# Make progress as soon as we have all precommits (as opposed to waiting for timeouts)
skip_timeout_commit = false

#######################################################################
###                       Mempool Configuration                     ###
#######################################################################

[mempool]

# Maximum number of transactions in the mempool
size = 5000

# Maximum size of a single transaction
max_tx_bytes = 1048576

# Maximum size of the mempool in bytes
max_txs_bytes = 1073741824

#######################################################################
###                    Instrumentation Configuration                ###
#######################################################################

[instrumentation]

# Prometheus metrics
prometheus = true

# Prometheus listening address
prometheus_listen_addr = ":26660"
`, chainID, moniker)
}

// generateAppTOML generates the app.toml content
func generateAppTOML() string {
	return `# Aura Application Configuration
# Generated by aurad init

#######################################################################
###                           gRPC Configuration                    ###
#######################################################################

[grpc]

# Enable defines if the gRPC server should be enabled.
enable = true

# Address defines the gRPC server address to bind to.
address = "localhost:9090"

# MaxRecvMsgSize defines the max message size in bytes the server can receive.
max-recv-msg-size = 10485760

# MaxSendMsgSize defines the max message size in bytes the server can send.
max-send-msg-size = 10485760

#######################################################################
###                           API Configuration                     ###
#######################################################################

[api]

# Enable defines if the API server should be enabled.
enable = true

# Address defines the API server to listen on.
address = "tcp://localhost:1317"

# EnableUnsafeCORS defines if CORS should be enabled (unsafe - use it for development only)
enabled-unsafe-cors = false

# MaxOpenConnections defines the number of maximum open connections.
max-open-connections = 1000

#######################################################################
###                       Logging Configuration                     ###
#######################################################################

[logging]

# Level defines the logging level
# Options: trace, debug, info, warn, error, fatal, panic
level = "info"

# Format defines the logging format
# Options: json, plain
format = "plain"

#######################################################################
###                     State Sync Configuration                    ###
#######################################################################

[state-sync]

# Enable state sync
snapshot-interval = 0

# Number of recent snapshots to keep
snapshot-keep-recent = 2

#######################################################################
###                       Module Configuration                      ###
#######################################################################

[modules]

# AURA consolidated modules - enabled by default

# Consolidated security module (network, validator, wallet, incident, crypto, privacy)
[modules.security]
enabled = true

# Consolidated identity module (auth, identitychange)
[modules.identity]
enabled = true

# Consolidated economics module (economicsecurity, governance)
[modules.economics]
enabled = true

# Core AURA modules
[modules.vcregistry]
enabled = true

[modules.inclusionroutines]
enabled = true

[modules.confidencescore]
enabled = true

[modules.dex]
enabled = true

[modules.bridge]
enabled = true

[modules.compliance]
enabled = true

# CosmWasm module (absorbs contractregistry)
[modules.wasm]
enabled = true

#######################################################################
###                    Store Pruning Configuration                 ###
#######################################################################

[store]

# The type of pruning to perform
# Options: "default" | "nothing" | "everything" | "custom"
# Use "nothing" for archive nodes that keep all versions for queries
pruning = "nothing"

# These are only applied if pruning = "custom"
pruning-keep-recent = 0
pruning-interval = 0
`
}

// generateCompleteGenesis generates a complete genesis.json with consolidated module structure
func generateCompleteGenesis(chainID string, keyInfo *ValidatorKeyInfo) string {
	encoding := app.MakeEncodingConfig()
	genesisState := defaultGenesisAppState(encoding.Codec)

	genesisState[auth.ModuleName] = updateAuthGenesisState(encoding.Codec, genesisState[auth.ModuleName], keyInfo.AccountAddress)
	genesisState[bank.ModuleName] = updateBankGenesisState(encoding.Codec, genesisState[bank.ModuleName], keyInfo.AccountAddress)
	genesisState[staking.ModuleName] = updateStakingGenesisState(encoding.Codec, genesisState[staking.ModuleName], keyInfo)

	genesisDoc := map[string]any{
		"genesis_time":   "2025-01-01T00:00:00.000000Z",
		"chain_id":       chainID,
		"initial_height": "1",
		"consensus_params": map[string]any{
			"block": map[string]string{
				"max_bytes": "22020096",
				"max_gas":   "-1",
			},
			"evidence": map[string]string{
				"max_age_num_blocks": "100000",
				"max_age_duration":   "172800000000000",
				"max_bytes":          "1048576",
			},
			"validator": map[string][]string{
				"pub_key_types": {"ed25519"},
			},
			"version": map[string]string{"app": "0"},
			"abci": map[string]string{
				"vote_extensions_enable_height": "0",
			},
		},
		"app_hash": "",
		"validators": []map[string]any{
			{
				"address": keyInfo.ConsensusAddressHex,
				"pub_key": map[string]string{
					"type":  "tendermint/PubKeyEd25519",
					"value": keyInfo.PublicKeyBase64,
				},
				"power": DefaultValidatorPower,
				"name":  "genesis-validator",
			},
		},
		"app_state": genesisState,
	}

	bytes, err := json.MarshalIndent(genesisDoc, "", "  ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal genesis doc: %w", err))
	}

	return string(bytes)
}

func updateAuthGenesisState(cdc codec.Codec, raw json.RawMessage, addrBech32 string) json.RawMessage {
	var gs auth.GenesisState
	if len(raw) == 0 {
		gs = *auth.DefaultGenesisState()
	} else {
		if err := cdc.UnmarshalJSON(raw, &gs); err != nil {
			panic(fmt.Errorf("failed to unmarshal auth genesis: %w", err))
		}
	}

	addr, err := sdk.AccAddressFromBech32(addrBech32)
	if err != nil {
		panic(fmt.Errorf("invalid account address %s: %w", addrBech32, err))
	}

	baseAcc := auth.NewBaseAccountWithAddress(addr)
	accAny, err := codectypes.NewAnyWithValue(baseAcc)
	if err != nil {
		panic(fmt.Errorf("failed to encode auth account to Any: %w", err))
	}
	gs.Accounts = append(gs.Accounts, accAny)

	return cdc.MustMarshalJSON(&gs)
}

func updateBankGenesisState(cdc codec.Codec, raw json.RawMessage, addrBech32 string) json.RawMessage {
	var gs bank.GenesisState
	if len(raw) == 0 {
		gs = *bank.DefaultGenesisState()
	} else {
		if err := cdc.UnmarshalJSON(raw, &gs); err != nil {
			panic(fmt.Errorf("failed to unmarshal bank genesis: %w", err))
		}
	}

	addr, err := sdk.AccAddressFromBech32(addrBech32)
	if err != nil {
		panic(fmt.Errorf("invalid bank address %s: %w", addrBech32, err))
	}

	stakeAmt, ok := sdkmath.NewIntFromString(DefaultValidatorTokens)
	if !ok {
		panic("invalid DefaultValidatorTokens")
	}
	uauraAmt, ok := sdkmath.NewIntFromString("1000000000000")
	if !ok {
		panic("invalid uaura supply")
	}

	gs.Balances = []bank.Balance{
		{
			Address: addr.String(),
			Coins: sdk.NewCoins(
				sdk.NewCoin("stake", stakeAmt),
				sdk.NewCoin("uaura", uauraAmt),
			),
		},
	}

	gs.Supply = sdk.NewCoins(
		sdk.NewCoin("stake", stakeAmt),
		sdk.NewCoin("uaura", uauraAmt),
	)

	return cdc.MustMarshalJSON(&gs)
}

func updateStakingGenesisState(cdc codec.Codec, raw json.RawMessage, keyInfo *ValidatorKeyInfo) json.RawMessage {
	var gs staking.GenesisState
	if len(raw) == 0 {
		gs = *staking.DefaultGenesisState()
	} else {
		if err := cdc.UnmarshalJSON(raw, &gs); err != nil {
			panic(fmt.Errorf("failed to unmarshal staking genesis: %w", err))
		}
	}

	tokenAmt, ok := sdkmath.NewIntFromString(DefaultBondedTokens)
	if !ok {
		panic("invalid DefaultBondedTokens")
	}
	powerAmt, ok := sdkmath.NewIntFromString(DefaultValidatorPower)
	if !ok {
		panic("invalid DefaultValidatorPower")
	}

	pubKeyAny, err := codectypes.NewAnyWithValue(&cosmosed25519.PubKey{Key: keyInfo.PublicKeyBytes})
	if err != nil {
		panic(fmt.Errorf("failed to encode consensus pubkey: %w", err))
	}

	sharesValue, err := sdkmath.LegacyNewDecFromStr(tokenAmt.String() + ".000000000000000000")
	if err != nil {
		panic(fmt.Errorf("invalid shares decimal: %w", err))
	}

	rate, err := sdkmath.LegacyNewDecFromStr("0.1")
	if err != nil {
		panic(fmt.Errorf("invalid commission rate: %w", err))
	}
	maxRate, err := sdkmath.LegacyNewDecFromStr("0.2")
	if err != nil {
		panic(fmt.Errorf("invalid commission max rate: %w", err))
	}
	maxChange, err := sdkmath.LegacyNewDecFromStr("0.01")
	if err != nil {
		panic(fmt.Errorf("invalid commission max change rate: %w", err))
	}

	validator := staking.Validator{
		OperatorAddress: keyInfo.OperatorAddress,
		ConsensusPubkey: pubKeyAny,
		Jailed:          false,
		Status:          staking.Bonded,
		Tokens:          tokenAmt,
		DelegatorShares: sharesValue,
		Description:     staking.Description{Moniker: "genesis-validator"},
		UnbondingHeight: 0,
		UnbondingTime:   time.Unix(0, 0),
		Commission: staking.Commission{
			CommissionRates: staking.CommissionRates{
				Rate:          rate,
				MaxRate:       maxRate,
				MaxChangeRate: maxChange,
			},
			UpdateTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		MinSelfDelegation: sdkmath.NewInt(1),
	}

	gs.LastTotalPower = powerAmt
	gs.LastValidatorPowers = []staking.LastValidatorPower{
		{
			Address: keyInfo.OperatorAddress,
			Power:   powerAmt.Int64(),
		},
	}
	gs.Validators = []staking.Validator{validator}
	gs.Delegations = []staking.Delegation{
		{
			DelegatorAddress: keyInfo.AccountAddress,
			ValidatorAddress: keyInfo.OperatorAddress,
			Shares:           sharesValue,
		},
	}

	return cdc.MustMarshalJSON(&gs)
}

func defaultGenesisAppState(cdc codec.JSONCodec) map[string]json.RawMessage {
	state := make(map[string]json.RawMessage)
	for name, mod := range app.ModuleBasics {
		state[name] = safeDefaultGenesis(mod, cdc)
	}
	return state
}

func safeDefaultGenesis(mod module.AppModuleBasic, cdc codec.JSONCodec) (raw json.RawMessage) {
	defer func() {
		if r := recover(); r != nil {
			raw = json.RawMessage(`{}`)
		}
	}()
	if mod == nil {
		return json.RawMessage(`{}`)
	}
	if hasGen, ok := mod.(interface {
		DefaultGenesis(codec.JSONCodec) json.RawMessage
	}); ok {
		raw = hasGen.DefaultGenesis(cdc)
	} else {
		return json.RawMessage(`{}`)
	}
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	return
}
