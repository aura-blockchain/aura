// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// GetQueryCmd returns the query commands for the cryptography module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "cryptography",
		Short:                      "Querying commands for the cryptography module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryKeyRotationSchedule(),
		CmdQueryThresholdScheme(),
		CmdQueryVerifyZKProof(),
		CmdQuerySecureEnclave(),
		CmdQueryQuantumResistantKey(),
		CmdQueryRandomSourceStatus(),
		CmdQueryCertificatePin(),
	)

	return cmd
}

// CmdQueryParams queries the module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the cryptography module parameters",
		Long: `Query the current cryptography module parameters including:
  - Key rotation settings
  - Threshold signature settings
  - ZK proof verification settings
  - Secure enclave settings
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &v1beta1.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryKeyRotationSchedule queries a key rotation schedule
func CmdQueryKeyRotationSchedule() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key-rotation-schedule [id]",
		Short: "Query a key rotation schedule by ID",
		Long: `Query details of a key rotation schedule including:
  - Key ID being rotated
  - Rotation interval
  - Last rotation time
  - Next rotation time
  - Rotation policy

Examples:
  aurad query cryptography key-rotation-schedule schedule-123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.KeyRotationSchedule(context.Background(), &v1beta1.QueryKeyRotationScheduleRequest{
				Id: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryThresholdScheme queries a threshold signature scheme
func CmdQueryThresholdScheme() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "threshold-scheme [scheme-id]",
		Short: "Query a threshold signature scheme by ID",
		Long: `Query details of a threshold signature scheme including:
  - Threshold value
  - Total participants
  - Participant IDs
  - Scheme type (BLS, ECDSA, Ed25519)
  - Combined public key
  - Current signature shares

Examples:
  aurad query cryptography threshold-scheme scheme-123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.ThresholdScheme(context.Background(), &v1beta1.QueryThresholdSchemeRequest{
				SchemeId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVerifyZKProof verifies a zero-knowledge proof
func CmdQueryVerifyZKProof() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-zk-proof [proof-id] [proof-data-hex] [public-inputs-hex]",
		Short: "Verify a zero-knowledge proof",
		Long: `Verify a zero-knowledge proof against a registered circuit.

Returns:
  - Verification result (valid/invalid)
  - Error message if invalid

Examples:
  aurad query cryptography verify-zk-proof proof-123 0xproofdata... 0xinputs...
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
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

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.VerifyZKProof(context.Background(), &v1beta1.QueryVerifyZKProofRequest{
				ProofId:      proofID,
				ProofData:    proofData,
				PublicInputs: publicInputs,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQuerySecureEnclave queries secure enclave information
func CmdQuerySecureEnclave() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secure-enclave [enclave-id]",
		Short: "Query secure enclave configuration",
		Long: `Query details of a registered secure enclave including:
  - Enclave type (HSM, SGX, TPM, TrustZone)
  - Attestation data
  - Registration time
  - Owner
  - Metadata

Examples:
  aurad query cryptography secure-enclave enclave-123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.SecureEnclave(context.Background(), &v1beta1.QuerySecureEnclaveRequest{
				EnclaveId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryQuantumResistantKey queries a quantum-resistant key
func CmdQueryQuantumResistantKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quantum-resistant-key [key-id]",
		Short: "Query a quantum-resistant key",
		Long: `Query details of a quantum-resistant key including:
  - Algorithm (Dilithium, Kyber, Falcon, SPHINCS+)
  - Public key
  - Creation time
  - Expiration time
  - Owner

Examples:
  aurad query cryptography quantum-resistant-key key-123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.QuantumResistantKey(context.Background(), &v1beta1.QueryQuantumResistantKeyRequest{
				KeyId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryRandomSourceStatus queries the status of random sources
func CmdQueryRandomSourceStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "random-source-status",
		Short: "Query the status of cryptographic random sources",
		Long: `Query the status of all cryptographic random number generators including:
  - Source types (drand, VRF, hardware)
  - Health status
  - Last update times
  - Entropy quality metrics

Examples:
  aurad query cryptography random-source-status
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.RandomSourceStatus(context.Background(), &v1beta1.QueryRandomSourceStatusRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryCertificatePin queries certificate pinning configuration
func CmdQueryCertificatePin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificate-pin [hostname]",
		Short: "Query certificate pinning configuration for a hostname",
		Long: `Query certificate pinning details for a hostname including:
  - Pinned certificate hashes
  - Pin type (public key, certificate, SPKI)
  - Expiration time
  - Creator

Examples:
  aurad query cryptography certificate-pin api.aura.network
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.CertificatePin(context.Background(), &v1beta1.QueryCertificatePinRequest{
				Hostname: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
