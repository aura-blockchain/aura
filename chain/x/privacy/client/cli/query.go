package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	privacyv1beta1 "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// GetQueryCmd returns the query commands for the privacy module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "privacy",
		Short:                      "Querying commands for the privacy module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryMixingPool(),
		CmdQueryMixingPools(),
		CmdQueryViewKey(),
		CmdQueryViewKeys(),
		CmdQueryVerifyZKProof(),
		// SECURITY: CmdQueryDecryptWithViewKey removed - decryption must be client-side
	)

	return cmd
}

// CmdQueryParams queries module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query privacy module parameters",
		Long: `Query the current privacy module parameters.

Examples:
  aurad query privacy params

Returns:
  - ZK proofs enabled status
  - Stealth addresses enabled status
  - Ring signatures enabled status
  - Confidential transactions enabled status
  - Network privacy enabled status
  - Mixing enabled status
  - Min/max ring sizes
  - Min mixing participants
  - Mixing fee
  - ZK proof verification cost
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := privacyv1beta1.NewQueryClient(clientCtx)

			req := &privacyv1beta1.QueryParamsRequest{}

			res, err := queryClient.Params(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryMixingPool queries a specific mixing pool
func CmdQueryMixingPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mixing-pool [pool-id]",
		Short: "Query a specific mixing pool",
		Long: `Query detailed information about a specific mixing pool.

Examples:
  aurad query privacy mixing-pool pool-123
  aurad query privacy mixing-pool pool-456

Returns:
  - Pool ID and status
  - Participant addresses
  - Min/max participants
  - Denomination amount
  - Mixing rounds
  - Deadline timestamp
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := privacyv1beta1.NewQueryClient(clientCtx)

			req := &privacyv1beta1.QueryMixingPoolRequest{
				PoolId: args[0],
			}

			res, err := queryClient.MixingPool(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryMixingPools queries all mixing pools
func CmdQueryMixingPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mixing-pools",
		Short: "Query all mixing pools",
		Long: `Query all mixing pools with optional status filter.

Examples:
  aurad query privacy mixing-pools
  aurad query privacy mixing-pools --status OPEN
  aurad query privacy mixing-pools --status MIXING

Status options:
  - OPEN: Accepting participants
  - MIXING: Currently mixing
  - COMPLETED: Mixing completed
  - EXPIRED: Deadline passed
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := privacyv1beta1.NewQueryClient(clientCtx)

			status, _ := cmd.Flags().GetString("status")

			req := &privacyv1beta1.QueryMixingPoolsRequest{
				Status: status,
			}

			res, err := queryClient.MixingPools(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by status (OPEN, MIXING, COMPLETED, EXPIRED)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryViewKey queries a specific view key
func CmdQueryViewKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view-key [public-view-key]",
		Short: "Query a specific view key",
		Long: `Query information about a registered view key.

Examples:
  aurad query privacy view-key abc123def456...
  aurad query privacy view-key 789abcdef012...

Returns:
  - Key type (INCOMING, OUTGOING, AUDIT)
  - Public view key
  - Associated address
  - Permissions
  - Expiration timestamp

Note: Private view key is not returned for security.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := privacyv1beta1.NewQueryClient(clientCtx)

			// Parse public view key
			publicViewKey, err := hex.DecodeString(args[0])
			if err != nil {
				return fmt.Errorf("invalid public-view-key: %w", err)
			}

			req := &privacyv1beta1.QueryViewKeyRequest{
				PublicViewKey: publicViewKey,
			}

			res, err := queryClient.ViewKey(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryViewKeys queries all view keys for an address
func CmdQueryViewKeys() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view-keys [address]",
		Short: "Query all view keys for an address",
		Long: `Query all registered view keys for a specific address.

Examples:
  aurad query privacy view-keys aura1abc...
  aurad query privacy view-keys aura1def...

Returns:
  List of all view keys registered by the address with:
  - Key types
  - Public view keys
  - Permissions
  - Expiration timestamps
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := privacyv1beta1.NewQueryClient(clientCtx)

			req := &privacyv1beta1.QueryViewKeysRequest{
				Address: args[0],
			}

			res, err := queryClient.ViewKeys(cmd.Context(), req)
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
		Use:   "verify-zk-proof [proof-file]",
		Short: "Verify a zero-knowledge proof",
		Long: `Verify the validity of a zero-knowledge proof without execution.

Examples:
  aurad query privacy verify-zk-proof proof.json
  aurad query privacy verify-zk-proof zk_proof_data.json

The proof file should contain a JSON-encoded ZKProof with:
  - Proof type (GROTH16, PLONK, SNARK, STARK)
  - Proof data
  - Public inputs
  - Verification key
  - Circuit ID

Returns:
  - Whether the proof is valid
  - Error message if verification failed
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := privacyv1beta1.NewQueryClient(clientCtx)

			// Note: In actual implementation, would read and parse the proof file
			// For now, this is a placeholder

			req := &privacyv1beta1.QueryVerifyZKProofRequest{
				// ZkProof would be populated from file
			}

			res, err := queryClient.VerifyZKProof(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// SECURITY NOTE: CmdQueryDecryptWithViewKey has been removed.
// Decryption must be performed client-side using private keys that never leave the client.
// Private view keys must NEVER be transmitted to the blockchain or passed as CLI arguments.
//
// To decrypt transaction data:
// 1. Download encrypted data from the blockchain using standard query commands
// 2. Decrypt locally using your private view key (stored securely in your keystore)
// 3. Implement client-side decryption in your application or wallet software
//
// Never expose private keys in CLI commands, API calls, or any blockchain interaction.
