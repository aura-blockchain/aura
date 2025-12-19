package cli

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/gogoproto/types"

	"github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// GetTxCmd returns the transaction commands for the walletsecurity module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "walletsecurity",
		Short:                      "Wallet security transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdRegisterHardwareWallet(),
		CmdCreateMultiSigWallet(),
		CmdSignMultiSigTransaction(),
		CmdConfigureSocialRecovery(),
		CmdInitiateRecovery(),
		CmdApproveRecovery(),
		CmdExecuteRecovery(),
		CmdSimulateTransaction(),
		CmdVerifyDomain(),
		CmdSetSpendingLimit(),
		CmdConfigureSession(),
		CmdLockSession(),
		CmdUnlockSession(),
		CmdEnrollBiometric(),
		CmdAuthenticateBiometric(),
		CmdStoreInSecureEnclave(),
		CmdCreateEncryptedBackup(),
		CmdConfigureDustFilter(),
		CmdValidateAddressChecksum(),
	)

	return cmd
}

// CmdRegisterHardwareWallet registers a hardware wallet
func CmdRegisterHardwareWallet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-hw-wallet [type] [device-id] [firmware-version] [derivation-path] [signature-hex]",
		Short: "Register a hardware wallet (Ledger, Trezor, etc.)",
		Long: `Register a hardware wallet for enhanced security.

Examples:
  aurad tx walletsecurity register-hw-wallet LEDGER device123 2.1.0 "m/44'/118'/0'/0/0" 0xsig... --from alice
  aurad tx walletsecurity register-hw-wallet TREZOR device456 1.10.0 "m/44'/118'/0'/0/0" 0xsig... --from bob

Hardware wallet types:
  LEDGER: Ledger Nano S/X
  TREZOR: Trezor One/Model T
  KEYSTONE: Keystone Pro
  GRIDPLUS: GridPlus Lattice1
`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var hwType v1beta1.HardwareWalletType
			switch strings.ToUpper(args[0]) {
			case "LEDGER":
				hwType = v1beta1.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER
			case "TREZOR":
				hwType = v1beta1.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR
			case "KEEPKEY":
				hwType = v1beta1.HardwareWalletType_HARDWARE_WALLET_TYPE_KEEPKEY
			case "COLDCARD":
				hwType = v1beta1.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD
			default:
				return fmt.Errorf("invalid hardware wallet type: %s", args[0])
			}

			signature, err := hex.DecodeString(strings.TrimPrefix(args[4], "0x"))
			if err != nil {
				return fmt.Errorf("invalid signature hex: %w", err)
			}

			msg := &v1beta1.MsgRegisterHardwareWallet{
				Address:         clientCtx.GetFromAddress().String(),
				Type:            hwType,
				DeviceId:        args[1],
				FirmwareVersion: args[2],
				DerivationPath:  args[3],
				Signature:       signature,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateMultiSigWallet creates a multi-signature wallet
func CmdCreateMultiSigWallet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-multisig [signers] [threshold]",
		Short: "Create a multi-signature wallet",
		Long: `Create a multi-signature wallet requiring multiple signatures.

Examples:
  aurad tx walletsecurity create-multisig "alice,bob,charlie" 2 --from alice
  aurad tx walletsecurity create-multisig "val1,val2,val3,val4" 3 --from val1 --signer-weights "val1=2,val2=2,val3=1,val4=1" --weight-threshold 4 --time-lock 3600s

Flags:
  --signer-weights: Comma-separated signer=weight pairs
  --weight-threshold: Required weight sum instead of signature count
  --time-lock: Duration to lock transactions before execution
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signers := strings.Split(args[0], ",")
			threshold, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid threshold: %w", err)
			}

			// Parse optional signer weights
			weightsStr, _ := cmd.Flags().GetString("signer-weights")
			signerWeights := make(map[string]int32)
			if weightsStr != "" {
				pairs := strings.Split(weightsStr, ",")
				for _, pair := range pairs {
					kv := strings.SplitN(pair, "=", 2)
					if len(kv) == 2 {
							weight, err := strconv.ParseInt(kv[1], 10, 32)
							if err != nil {
								return fmt.Errorf("invalid weight for %s: %w", kv[0], err)
							}
							signerWeights[kv[0]] = clampInt64ToInt32(weight)
						}
					}
				}

			weightThreshold, _ := cmd.Flags().GetInt32("weight-threshold")
			timeLockStr, _ := cmd.Flags().GetString("time-lock")
			var timeLock *types.Duration
			if timeLockStr != "" {
				d, err := time.ParseDuration(timeLockStr)
				if err != nil {
					return fmt.Errorf("invalid time-lock: %w", err)
				}
					timeLock = &types.Duration{
						Seconds: int64(d.Seconds()),
						Nanos:   clampNanosecondsToInt32(d.Nanoseconds()),
					}
				}

			msg := &v1beta1.MsgCreateMultiSigWallet{
				Creator:         clientCtx.GetFromAddress().String(),
				Signers:         signers,
					Threshold:       clampInt64ToInt32(threshold),
				SignerWeights:   signerWeights,
				WeightThreshold: weightThreshold,
			}
			if timeLock != nil {
				msg.TimeLock = timeLock
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("signer-weights", "", "Signer weights (e.g., alice=2,bob=1)")
	cmd.Flags().Int32("weight-threshold", 0, "Weight threshold instead of signature count")
	cmd.Flags().String("time-lock", "", "Time lock duration (e.g., 3600s)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSignMultiSigTransaction signs a pending multi-sig transaction
func CmdSignMultiSigTransaction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-multisig-tx [tx-id] [signature-hex]",
		Short: "Sign a pending multi-signature transaction",
		Long: `Sign a pending multi-signature transaction.

Examples:
  aurad tx walletsecurity sign-multisig-tx tx123 0xsignature... --from alice
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txID := args[0]
			signature, err := hex.DecodeString(strings.TrimPrefix(args[1], "0x"))
			if err != nil {
				return fmt.Errorf("invalid signature hex: %w", err)
			}

			msg := &v1beta1.MsgSignMultiSigTransaction{
				TxId:      txID,
				Signer:    clientCtx.GetFromAddress().String(),
				Signature: signature,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdConfigureSocialRecovery configures social recovery
func CmdConfigureSocialRecovery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure-social-recovery [wallet-id] [guardians] [threshold] [delay]",
		Short: "Configure social recovery for wallet",
		Long: `Configure social recovery with trusted guardians.

Examples:
  aurad tx walletsecurity configure-social-recovery wallet123 "guardian1,guardian2,guardian3" 2 "7d" --from alice

The delay specifies how long to wait before recovery can be executed after threshold is met.
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			walletID := args[0]
			guardianAddrs := strings.Split(args[1], ",")
				threshold, err := strconv.ParseInt(args[2], 10, 32)
				if err != nil {
					return fmt.Errorf("invalid threshold: %w", err)
				}

			delay, err := time.ParseDuration(args[3])
			if err != nil {
				return fmt.Errorf("invalid delay: %w", err)
			}

			// Build guardians
			guardians := make([]*v1beta1.Guardian, len(guardianAddrs))
			for i, addr := range guardianAddrs {
				guardians[i] = &v1beta1.Guardian{
					Address: addr,
				}
			}

			msg := &v1beta1.MsgConfigureSocialRecovery{
				WalletId:          walletID,
				Guardians:         guardians,
					RecoveryThreshold: clampInt64ToInt32(threshold),
					RecoveryDelay: &types.Duration{
						Seconds: int64(delay.Seconds()),
						Nanos:   clampNanosecondsToInt32(delay.Nanoseconds()),
					},
				}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdInitiateRecovery initiates a recovery process
func CmdInitiateRecovery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initiate-recovery [wallet-id] [new-address]",
		Short: "Initiate social recovery as a guardian",
		Long: `Initiate social recovery process for a wallet.

Examples:
  aurad tx walletsecurity initiate-recovery wallet123 aura1newaddress... --from guardian1
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgInitiateRecovery{
				WalletId:   args[0],
				NewAddress: args[1],
				Initiator:  clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdApproveRecovery approves a recovery request
func CmdApproveRecovery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve-recovery [request-id] [signature-hex]",
		Short: "Approve a recovery request as a guardian",
		Long: `Approve a pending recovery request.

Examples:
  aurad tx walletsecurity approve-recovery req123 0xsignature... --from guardian2
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			requestID := args[0]
			signature, err := hex.DecodeString(strings.TrimPrefix(args[1], "0x"))
			if err != nil {
				return fmt.Errorf("invalid signature hex: %w", err)
			}

			msg := &v1beta1.MsgApproveRecovery{
				RequestId: requestID,
				Guardian:  clientCtx.GetFromAddress().String(),
				Signature: signature,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdExecuteRecovery executes an approved recovery
func CmdExecuteRecovery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-recovery [request-id]",
		Short: "Execute an approved recovery request",
		Long: `Execute a recovery request that has reached threshold and passed the delay period.

Examples:
  aurad tx walletsecurity execute-recovery req123 --from anyone
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgExecuteRecovery{
				RequestId: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSimulateTransaction simulates a transaction
func CmdSimulateTransaction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "simulate-tx [tx-data-hex]",
		Short: "Simulate a transaction before execution",
		Long: `Simulate a transaction to check for errors and estimate gas.

Examples:
  aurad tx walletsecurity simulate-tx 0xtxdata... --from alice
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txData, err := hex.DecodeString(strings.TrimPrefix(args[0], "0x"))
			if err != nil {
				return fmt.Errorf("invalid tx data hex: %w", err)
			}

			msg := &v1beta1.MsgSimulateTransaction{
				TxData: txData,
				Sender: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdVerifyDomain verifies a domain for phishing protection
func CmdVerifyDomain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-domain [domain] [cert-hash]",
		Short: "Verify a domain for phishing protection",
		Long: `Verify and whitelist a domain to prevent phishing attacks.

Examples:
  aurad tx walletsecurity verify-domain app.aura.network 0xcerthash... --from alice
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgVerifyDomain{
				Domain:          args[0],
				CertificateHash: args[1],
				Verifier:        clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSetSpendingLimit sets spending limits
func CmdSetSpendingLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-spending-limit [wallet-id] [denom] [daily-limit] [weekly-limit] [monthly-limit]",
		Short: "Set spending limits for a wallet",
		Long: `Set daily, weekly, and monthly spending limits.

Examples:
  aurad tx walletsecurity set-spending-limit wallet123 uaura 1000000 5000000 20000000 --from alice
`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgSetSpendingLimit{
				WalletId:     args[0],
				Denom:        args[1],
				DailyLimit:   args[2],
				WeeklyLimit:  args[3],
				MonthlyLimit: args[4],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdConfigureSession configures session settings
func CmdConfigureSession() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure-session [wallet-id] [timeout-duration] [auto-lock] [inactivity-threshold]",
		Short: "Configure wallet session settings",
		Long: `Configure session timeout and auto-lock settings.

Examples:
  aurad tx walletsecurity configure-session wallet123 30m true 600 --from alice
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			timeout, err := time.ParseDuration(args[1])
			if err != nil {
				return fmt.Errorf("invalid timeout: %w", err)
			}

			autoLock, err := strconv.ParseBool(args[2])
			if err != nil {
				return fmt.Errorf("invalid auto-lock: %w", err)
			}

				inactivityThreshold, err := strconv.ParseInt(args[3], 10, 32)
				if err != nil {
					return fmt.Errorf("invalid inactivity threshold: %w", err)
				}

			msg := &v1beta1.MsgConfigureSession{
				WalletId: args[0],
					TimeoutDuration: &types.Duration{
						Seconds: int64(timeout.Seconds()),
						Nanos:   clampNanosecondsToInt32(timeout.Nanoseconds()),
					},
					AutoLockEnabled:            autoLock,
					InactivityThresholdSeconds: clampInt64ToInt32(inactivityThreshold),
				}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdLockSession locks a session
func CmdLockSession() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-session [session-id]",
		Short: "Lock a wallet session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgLockSession{
				SessionId: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUnlockSession unlocks a session
func CmdUnlockSession() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock-session [session-id] [auth-proof-hex]",
		Short: "Unlock a wallet session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			authProof, err := hex.DecodeString(strings.TrimPrefix(args[1], "0x"))
			if err != nil {
				return fmt.Errorf("invalid auth proof hex: %w", err)
			}

			msg := &v1beta1.MsgUnlockSession{
				SessionId:           args[0],
				AuthenticationProof: authProof,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdEnrollBiometric enrolls biometric authentication
func CmdEnrollBiometric() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enroll-biometric [wallet-id] [type] [enrollment-data-hex]",
		Short: "Enroll biometric authentication",
		Long: `Enroll biometric authentication (fingerprint, face, iris).

Types: FINGERPRINT, FACE_ID, IRIS, VOICE
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var bioType v1beta1.BiometricType
			switch strings.ToUpper(args[1]) {
			case "FINGERPRINT":
				bioType = v1beta1.BiometricType_BIOMETRIC_TYPE_FINGERPRINT
			case "FACE_ID", "FACE":
				bioType = v1beta1.BiometricType_BIOMETRIC_TYPE_FACE
			case "IRIS":
				bioType = v1beta1.BiometricType_BIOMETRIC_TYPE_IRIS
			case "VOICE":
				bioType = v1beta1.BiometricType_BIOMETRIC_TYPE_VOICE
			default:
				return fmt.Errorf("invalid biometric type: %s", args[1])
			}

			enrollmentData, err := hex.DecodeString(strings.TrimPrefix(args[2], "0x"))
			if err != nil {
				return fmt.Errorf("invalid enrollment data hex: %w", err)
			}

			msg := &v1beta1.MsgEnrollBiometric{
				WalletId:       args[0],
				Type:           bioType,
				EnrollmentData: enrollmentData,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAuthenticateBiometric authenticates using biometric
func CmdAuthenticateBiometric() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authenticate-biometric [wallet-id] [biometric-proof-hex]",
		Short: "Authenticate using biometric",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			bioProof, err := hex.DecodeString(strings.TrimPrefix(args[1], "0x"))
			if err != nil {
				return fmt.Errorf("invalid biometric proof hex: %w", err)
			}

			msg := &v1beta1.MsgAuthenticateBiometric{
				WalletId:       args[0],
				BiometricProof: bioProof,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdStoreInSecureEnclave stores key in secure enclave
func CmdStoreInSecureEnclave() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store-in-enclave [wallet-id] [enclave-type] [encrypted-key-hex] [attestation-cert]",
		Short: "Store key material in secure enclave",
		Long: `Store key material in a secure enclave.

Enclave types: APPLE_SECURE_ENCLAVE, ANDROID_KEYSTORE, WINDOWS_TPM
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var enclaveType v1beta1.EnclaveType
			switch strings.ToUpper(args[1]) {
			case "TEE":
				enclaveType = v1beta1.EnclaveType_ENCLAVE_TYPE_TEE
			case "SGX":
				enclaveType = v1beta1.EnclaveType_ENCLAVE_TYPE_SGX
			case "KEYCHAIN", "APPLE_SECURE_ENCLAVE":
				enclaveType = v1beta1.EnclaveType_ENCLAVE_TYPE_KEYCHAIN
			case "KEYSTORE", "ANDROID_KEYSTORE":
				enclaveType = v1beta1.EnclaveType_ENCLAVE_TYPE_KEYSTORE
			case "TPM", "WINDOWS_TPM":
				enclaveType = v1beta1.EnclaveType_ENCLAVE_TYPE_TPM
			default:
				return fmt.Errorf("invalid enclave type: %s", args[1])
			}

			encryptedKey, err := hex.DecodeString(strings.TrimPrefix(args[2], "0x"))
			if err != nil {
				return fmt.Errorf("invalid encrypted key hex: %w", err)
			}

			msg := &v1beta1.MsgStoreInSecureEnclave{
				WalletId:              args[0],
				EnclaveType:           enclaveType,
				EncryptedKeyMaterial:  encryptedKey,
				AttestationCertificate: args[3],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateEncryptedBackup creates an encrypted backup
func CmdCreateEncryptedBackup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-backup [wallet-id] [encrypted-seed-hex] [algo] [kdf] [salt-hex] [iterations] [location]",
		Short: "Create an encrypted seed backup",
		Long: `Create an encrypted backup of wallet seed phrase.

Locations: CLOUD, HARDWARE, PAPER, OFFLINE
`,
		Args: cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			encryptedSeed, err := hex.DecodeString(strings.TrimPrefix(args[1], "0x"))
			if err != nil {
				return fmt.Errorf("invalid encrypted seed hex: %w", err)
			}

			salt, err := hex.DecodeString(strings.TrimPrefix(args[4], "0x"))
			if err != nil {
				return fmt.Errorf("invalid salt hex: %w", err)
			}

			iterations, err := strconv.ParseInt(args[5], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid iterations: %w", err)
			}

			var location v1beta1.BackupLocation
			switch strings.ToUpper(args[6]) {
			case "LOCAL":
				location = v1beta1.BackupLocation_BACKUP_LOCATION_LOCAL
			case "CLOUD":
				location = v1beta1.BackupLocation_BACKUP_LOCATION_CLOUD
			case "HARDWARE":
				location = v1beta1.BackupLocation_BACKUP_LOCATION_HARDWARE
			case "PAPER":
				location = v1beta1.BackupLocation_BACKUP_LOCATION_PAPER
			default:
				return fmt.Errorf("invalid backup location: %s", args[6])
			}

			msg := &v1beta1.MsgCreateEncryptedBackup{
				WalletId:              args[0],
				EncryptedSeed:         encryptedSeed,
				EncryptionAlgorithm:   args[2],
				KeyDerivationFunction: args[3],
				Salt:                  salt,
				Iterations:            int32(iterations),
				Location:              location,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdConfigureDustFilter configures dust attack filtering
func CmdConfigureDustFilter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure-dust-filter [wallet-id] [enabled] [min-amount] [max-dust-tx-per-block] [threshold]",
		Short: "Configure dust attack filtering",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			enabled, err := strconv.ParseBool(args[1])
			if err != nil {
				return fmt.Errorf("invalid enabled: %w", err)
			}

			maxDustTx, err := strconv.ParseInt(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid max dust tx: %w", err)
			}

			threshold, err := strconv.ParseInt(args[4], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid threshold: %w", err)
			}

			msg := &v1beta1.MsgConfigureDustFilter{
				WalletId:                      args[0],
				Enabled:                       enabled,
				MinimumAmount:                 args[2],
				MaxDustTransactionsPerBlock:   int32(maxDustTx),
				SuspiciousPatternThreshold:    int32(threshold),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdValidateAddressChecksum validates an address checksum
func CmdValidateAddressChecksum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-address [address] [algorithm]",
		Short: "Validate an address checksum",
		Long: `Validate address checksum to prevent typos.

Algorithms: EIP55, BECH32, BASE58CHECK
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var algorithm v1beta1.ChecksumAlgorithm
			switch strings.ToUpper(args[1]) {
			case "EIP55":
				algorithm = v1beta1.ChecksumAlgorithm_CHECKSUM_ALGORITHM_EIP55
			case "BECH32":
				algorithm = v1beta1.ChecksumAlgorithm_CHECKSUM_ALGORITHM_BECH32
			case "BASE58CHECK":
				algorithm = v1beta1.ChecksumAlgorithm_CHECKSUM_ALGORITHM_BASE58CHECK
			case "CRC32":
				algorithm = v1beta1.ChecksumAlgorithm_CHECKSUM_ALGORITHM_CRC32
			default:
				return fmt.Errorf("invalid algorithm: %s", args[1])
			}

			msg := &v1beta1.MsgValidateAddressChecksum{
				Address:   args[0],
				Algorithm: algorithm,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func clampInt64ToInt32(v int64) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

func clampNanosecondsToInt32(ns int64) int32 {
	if ns > math.MaxInt32 {
		return math.MaxInt32
	}
	if ns < math.MinInt32 {
		return math.MinInt32
	}
	return int32(ns)
}
