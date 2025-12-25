// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// GetTxCmd returns the transaction commands for the cryptography module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "cryptography",
		Aliases:                    []string{"crypto", "crypt"},
		Short:                      "Cryptography transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdCreateKeyRotationSchedule(),
		CmdRotateKey(),
		CmdCreateThresholdScheme(),
		CmdSubmitThresholdSignatureShare(),
		CmdRegisterZKProofCircuit(),
		CmdSubmitZKProof(),
		CmdRegisterSecureEnclave(),
		CmdGenerateQuantumResistantKey(),
		CmdAddCertificatePin(),
	)

	return cmd
}

// CmdCreateKeyRotationSchedule creates a new key rotation schedule
func CmdCreateKeyRotationSchedule() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-rotation-schedule [key-id] [rotation-interval-seconds] [policy]",
		Short: "Create an automated key rotation schedule",
		Long: `Create an automated key rotation schedule for cryptographic keys.

Examples:
  aurad tx cryptography create-rotation-schedule key-123 86400 ROTATE_AND_KEEP --from alice
  aurad tx cryptography create-rotation-schedule validator-key 604800 ROTATE_AND_REVOKE --from alice

Policies:
  ROTATE_AND_KEEP: Keep old key for decryption
  ROTATE_AND_REVOKE: Revoke old key immediately
  ROTATE_WITH_GRACE: Grace period before revocation
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			keyID := args[0]
			rotationInterval, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid rotation interval: %w", err)
			}

			// Parse max age days from policy argument
			maxAgeDays, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid max age days: %w", err)
			}

			// Create KeyRotationPolicy as a message struct
			policy := &v1beta1.KeyRotationPolicy{
				MaxAgeDays:              int32(maxAgeDays),
				WarningDaysBeforeExpiry: 7,  // default
				AutoRotate:              true, // default
				MaxRotationAttempts:     3,    // default
			}

			msg := &v1beta1.MsgCreateKeyRotationSchedule{
				Creator:                 clientCtx.GetFromAddress().String(),
				KeyId:                   keyID,
				RotationIntervalSeconds: rotationInterval,
				Policy:                  policy,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRotateKey manually rotates a key
func CmdRotateKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate-key [key-id] [new-public-key-hex]",
		Short: "Manually rotate a cryptographic key",
		Long: `Manually rotate a cryptographic key with a new public key.

Examples:
  aurad tx cryptography rotate-key key-123 0x1234abcd... --from alice
  aurad tx cryptography rotate-key validator-key 0xabcdef12... --from alice
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			keyID := args[0]
			newPublicKey, err := hex.DecodeString(strings.TrimPrefix(args[1], "0x"))
			if err != nil {
				return fmt.Errorf("invalid public key hex: %w", err)
			}

			msg := &v1beta1.MsgRotateKey{
				Creator:      clientCtx.GetFromAddress().String(),
				KeyId:        keyID,
				NewPublicKey: newPublicKey,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateThresholdScheme creates a new threshold signature scheme
func CmdCreateThresholdScheme() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-threshold-scheme [threshold] [total-participants] [participant-ids] [scheme-type]",
		Short: "Create a threshold signature scheme",
		Long: `Create a threshold signature scheme for multi-party signing.

Examples:
  aurad tx cryptography create-threshold-scheme 2 3 "alice,bob,charlie" BLS --from alice
  aurad tx cryptography create-threshold-scheme 3 5 "val1,val2,val3,val4,val5" ECDSA --from alice

Scheme types:
  BLS: BLS threshold signatures
  ECDSA: ECDSA threshold signatures
  ED25519: Ed25519 threshold signatures
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			threshold, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid threshold: %w", err)
			}

			totalParticipants, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid total participants: %w", err)
			}

			participantIDs := strings.Split(args[2], ",")

			var schemeType v1beta1.ThresholdSchemeType
			switch strings.ToUpper(args[3]) {
			case "BLS":
				schemeType = v1beta1.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_BLS
			case "ECDSA":
				schemeType = v1beta1.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA
			case "EDDSA", "ED25519":
				schemeType = v1beta1.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_EDDSA
			case "SCHNORR":
				schemeType = v1beta1.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_SCHNORR
			default:
				return fmt.Errorf("invalid scheme type: %s", args[3])
			}

			msg := &v1beta1.MsgCreateThresholdScheme{
				Creator:           clientCtx.GetFromAddress().String(),
				Threshold:         int32(threshold),
				TotalParticipants: int32(totalParticipants),
				ParticipantIds:    participantIDs,
				SchemeType:        schemeType,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSubmitThresholdSignatureShare submits a signature share
func CmdSubmitThresholdSignatureShare() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-threshold-share [scheme-id] [signature-share-hex] [message-hash-hex]",
		Short: "Submit a threshold signature share",
		Long: `Submit your signature share for a threshold signing operation.

Examples:
  aurad tx cryptography submit-threshold-share scheme-123 0xabcd1234... 0xhash5678... --from alice
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			schemeID := args[0]

			signatureShare, err := hex.DecodeString(strings.TrimPrefix(args[1], "0x"))
			if err != nil {
				return fmt.Errorf("invalid signature share hex: %w", err)
			}

			messageHash, err := hex.DecodeString(strings.TrimPrefix(args[2], "0x"))
			if err != nil {
				return fmt.Errorf("invalid message hash hex: %w", err)
			}

			msg := &v1beta1.MsgSubmitThresholdSignatureShare{
				Submitter:      clientCtx.GetFromAddress().String(),
				SchemeId:       schemeID,
				SignatureShare: signatureShare,
				MessageHash:    messageHash,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRegisterZKProofCircuit registers a new ZK proof circuit
func CmdRegisterZKProofCircuit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-zk-circuit [circuit-id] [proof-type] [public-params-hex] [verification-key-hex]",
		Short: "Register a zero-knowledge proof circuit",
		Long: `Register a new zero-knowledge proof circuit for verification.

Examples:
  aurad tx cryptography register-zk-circuit circuit-1 GROTH16 0xparams... 0xkey... --from alice
  aurad tx cryptography register-zk-circuit circuit-2 PLONK 0xparams... 0xkey... --from alice

Proof types:
  GROTH16: Groth16 ZK-SNARK
  PLONK: PLONK proof system
  STARK: STARK proof system
  BULLETPROOFS: Bulletproofs range proofs
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			circuitID := args[0]

			var proofType v1beta1.ZKProofType
			switch strings.ToUpper(args[1]) {
			case "GROTH16":
				proofType = v1beta1.ZKProofType_ZK_PROOF_TYPE_GROTH16
			case "PLONK":
				proofType = v1beta1.ZKProofType_ZK_PROOF_TYPE_PLONK
			case "STARK":
				proofType = v1beta1.ZKProofType_ZK_PROOF_TYPE_STARK
			case "BULLETPROOFS":
				proofType = v1beta1.ZKProofType_ZK_PROOF_TYPE_BULLETPROOFS
			default:
				return fmt.Errorf("invalid proof type: %s", args[1])
			}

			publicParams, err := hex.DecodeString(strings.TrimPrefix(args[2], "0x"))
			if err != nil {
				return fmt.Errorf("invalid public params hex: %w", err)
			}

			verificationKey, err := hex.DecodeString(strings.TrimPrefix(args[3], "0x"))
			if err != nil {
				return fmt.Errorf("invalid verification key hex: %w", err)
			}

			msg := &v1beta1.MsgRegisterZKProofCircuit{
				Creator:          clientCtx.GetFromAddress().String(),
				ProofType:        proofType,
				PublicParameters: publicParams,
				VerificationKey:  verificationKey,
				CircuitId:        circuitID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSubmitZKProof submits a zero-knowledge proof for verification
func CmdSubmitZKProof() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-zk-proof [proof-id] [proof-data-hex] [public-inputs-hex]",
		Short: "Submit a zero-knowledge proof for verification",
		Long: `Submit a zero-knowledge proof to be verified against a registered circuit.

Examples:
  aurad tx cryptography submit-zk-proof proof-123 0xproofdata... 0xinputs... --from alice
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proofID := args[0]

			proofData, err := hex.DecodeString(strings.TrimPrefix(args[1], "0x"))
			if err != nil {
				return fmt.Errorf("invalid proof data hex: %w", err)
			}

			publicInputs, err := hex.DecodeString(strings.TrimPrefix(args[2], "0x"))
			if err != nil {
				return fmt.Errorf("invalid public inputs hex: %w", err)
			}

			msg := &v1beta1.MsgSubmitZKProof{
				Submitter:    clientCtx.GetFromAddress().String(),
				ProofId:      proofID,
				ProofData:    proofData,
				PublicInputs: publicInputs,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRegisterSecureEnclave registers a secure enclave
func CmdRegisterSecureEnclave() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-enclave [enclave-type] [attestation-hex]",
		Short: "Register a secure enclave for key storage",
		Long: `Register a secure enclave (HSM, SGX, or TPM) for cryptographic key storage.

Examples:
  aurad tx cryptography register-enclave SGX 0xattestation... --from alice --metadata "version=2.0,model=SGX2"
  aurad tx cryptography register-enclave TPM 0xattestation... --from alice

Enclave types:
  HSM: Hardware Security Module
  SGX: Intel SGX
  TPM: Trusted Platform Module
  TRUSTZONE: ARM TrustZone
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var enclaveType v1beta1.SecureEnclaveType
			switch strings.ToUpper(args[0]) {
			case "HSM":
				enclaveType = v1beta1.SecureEnclaveType_SECURE_ENCLAVE_TYPE_HSM
			case "SGX":
				enclaveType = v1beta1.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX
			case "SEV":
				enclaveType = v1beta1.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SEV
			case "TPM":
				enclaveType = v1beta1.SecureEnclaveType_SECURE_ENCLAVE_TYPE_TPM
			case "KEYCHAIN":
				enclaveType = v1beta1.SecureEnclaveType_SECURE_ENCLAVE_TYPE_KEYCHAIN
			default:
				return fmt.Errorf("invalid enclave type: %s", args[0])
			}

			attestationData, err := hex.DecodeString(strings.TrimPrefix(args[1], "0x"))
			if err != nil {
				return fmt.Errorf("invalid attestation hex: %w", err)
			}

			// Parse optional metadata
			metadataStr, _ := cmd.Flags().GetString("metadata")
			metadata := make(map[string]string)
			if metadataStr != "" {
				pairs := strings.Split(metadataStr, ",")
				for _, pair := range pairs {
					kv := strings.SplitN(pair, "=", 2)
					if len(kv) == 2 {
						metadata[kv[0]] = kv[1]
					}
				}
			}

			msg := &v1beta1.MsgRegisterSecureEnclave{
				Creator:         clientCtx.GetFromAddress().String(),
				EnclaveType:     enclaveType,
				AttestationData: attestationData,
				EnclaveMetadata: metadata,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("metadata", "", "Enclave metadata (comma-separated key=value pairs)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdGenerateQuantumResistantKey generates a quantum-resistant key
func CmdGenerateQuantumResistantKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-qr-key [algorithm]",
		Short: "Generate a quantum-resistant key pair",
		Long: `Generate a quantum-resistant cryptographic key pair.

Examples:
  aurad tx cryptography generate-qr-key DILITHIUM --from alice --expires-in 365d
  aurad tx cryptography generate-qr-key KYBER --from alice --expires-in 730d

Algorithms:
  DILITHIUM: CRYSTALS-Dilithium (signature)
  KYBER: CRYSTALS-Kyber (KEM)
  FALCON: FALCON (signature)
  SPHINCS: SPHINCS+ (signature)
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var algorithm v1beta1.QuantumResistantAlgorithm
			switch strings.ToUpper(args[0]) {
			case "DILITHIUM":
				algorithm = v1beta1.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM
			case "KYBER":
				algorithm = v1beta1.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER
			case "FALCON":
				algorithm = v1beta1.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON
			case "SPHINCS":
				algorithm = v1beta1.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS
			case "NTRU":
				algorithm = v1beta1.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_NTRU
			default:
				return fmt.Errorf("invalid algorithm: %s", args[0])
			}

			// Parse optional expiration
			expiresIn, _ := cmd.Flags().GetString("expires-in")
			var expiresAt *time.Time
			if expiresIn != "" {
				duration, err := time.ParseDuration(expiresIn)
				if err != nil {
					return fmt.Errorf("invalid expires-in duration: %w", err)
				}
				t := time.Now().Add(duration)
				expiresAt = &t
			}

			msg := &v1beta1.MsgGenerateQuantumResistantKey{
				Creator:   clientCtx.GetFromAddress().String(),
				Algorithm: algorithm,
			}
			if expiresAt != nil {
				msg.ExpiresAt = expiresAt
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("expires-in", "", "Key expiration duration (e.g., 365d, 730d)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAddCertificatePin adds certificate pinning
func CmdAddCertificatePin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-cert-pin [hostname] [cert-hashes-hex] [pin-type]",
		Short: "Add certificate pinning for a hostname",
		Long: `Add certificate pinning to prevent MITM attacks.

Examples:
  aurad tx cryptography add-cert-pin api.aura.network 0xhash1,0xhash2 PUBLIC_KEY --from alice --expires-in 365d
  aurad tx cryptography add-cert-pin bridge.aura.network 0xhash1 CERTIFICATE --from alice

Pin types:
  PUBLIC_KEY: Pin public key
  CERTIFICATE: Pin full certificate
  SPKI: Pin Subject Public Key Info
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hostname := args[0]

			hashStrs := strings.Split(args[1], ",")
			certHashes := make([][]byte, len(hashStrs))
			for i, hashStr := range hashStrs {
				hash, err := hex.DecodeString(strings.TrimPrefix(hashStr, "0x"))
				if err != nil {
					return fmt.Errorf("invalid cert hash %d: %w", i, err)
				}
				certHashes[i] = hash
			}

			var pinType v1beta1.CertificatePinType
			switch strings.ToUpper(args[2]) {
			case "SPKI":
				pinType = v1beta1.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI
			case "FULL_CERT", "CERTIFICATE":
				pinType = v1beta1.CertificatePinType_CERTIFICATE_PIN_TYPE_FULL_CERT
			case "INTERMEDIATE":
				pinType = v1beta1.CertificatePinType_CERTIFICATE_PIN_TYPE_INTERMEDIATE
			default:
				return fmt.Errorf("invalid pin type: %s", args[2])
			}

			// Parse optional expiration
			expiresIn, _ := cmd.Flags().GetString("expires-in")
			var expiresAt *time.Time
			if expiresIn != "" {
				duration, err := time.ParseDuration(expiresIn)
				if err != nil {
					return fmt.Errorf("invalid expires-in duration: %w", err)
				}
				t := time.Now().Add(duration)
				expiresAt = &t
			}

			msg := &v1beta1.MsgAddCertificatePin{
				Creator:          clientCtx.GetFromAddress().String(),
				Hostname:         hostname,
				CertificateHashes: certHashes,
				PinType:          pinType,
			}
			if expiresAt != nil {
				msg.ExpiresAt = expiresAt
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("expires-in", "", "Pin expiration duration (e.g., 365d)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
