// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	vcregistryv1beta1 "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// GetQueryCmd returns the cli query commands for the vcregistry module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "vcregistry",
		Short:                      "Querying commands for the VC Registry module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// Core VC queries
		CmdQueryVC(),
		CmdQueryUserVCs(),
		CmdQueryVCStatus(),
		CmdQueryBatchVCStatus(),

		// Policy queries
		CmdQueryVCPolicy(),
		CmdQueryVCPolicies(),

		// Revocation queries
		CmdQueryRevocationList(),
		CmdQueryCheckRevocation(),

		// DID queries
		CmdQueryResolveDID(),
		CmdQueryDIDByAddress(),

		// Eligibility
		CmdQueryValidateMintEligibility(),

		// Stats
		CmdQueryStats(),
		CmdQueryParams(),

		// Presentation queries (these will be added when presentation service is defined in proto)
		// CmdQueryVerifyPresentation(),

		// Selective disclosure / Attributes queries (these will be added when attributes service is defined in proto)
		// CmdQueryDisclosurePolicy(),
		// CmdQueryAttributeVCs(),
		// CmdQueryParseVoiceCommand(),
		// CmdQueryAttributeVC(),
		// CmdQueryDisclosureRequest(),
		// CmdQueryPendingDisclosureRequests(),
	)

	return cmd
}

// CmdQueryVC queries a specific VC by ID
func CmdQueryVC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vc [vc-id]",
		Short: "Query a verifiable credential by ID",
		Long: `Query details of a specific verifiable credential.

Examples:
  aurad query vcregistry vc vc-123456
  aurad query vcregistry vc vc-789012

Returns:
  - VC ID
  - VC type
  - Holder DID and address
  - Status (active, revoked, expired, suspended)
  - Issuance and expiration timestamps
  - Credential hash
  - Metadata
  - Policy version
  - Prerequisites
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			req := &vcregistryv1beta1.QueryGetVCRequest{
				VcId: args[0],
			}

			res, err := queryClient.GetVC(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryUserVCs queries all VCs for a user
func CmdQueryUserVCs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-vcs [holder-address]",
		Short: "Query all VCs for a user",
		Long: `Query all verifiable credentials owned by a specific address.

Examples:
  aurad query vcregistry user-vcs aura1abc...
  aurad query vcregistry user-vcs aura1def... --status VC_STATUS_ACTIVE
  aurad query vcregistry user-vcs aura1ghi... --type VC_TYPE_VERIFIED_HUMAN

Filters:
  --status: Filter by VC status (VC_STATUS_ACTIVE, VC_STATUS_REVOKED, VC_STATUS_EXPIRED, VC_STATUS_SUSPENDED)
  --type: Filter by VC type

Returns a list of all VCs matching the criteria.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			holderAddress := args[0]

			// Parse optional filters
			statusStr, _ := cmd.Flags().GetString("status")
			var statusFilter vcregistryv1beta1.VCStatus
			if statusStr != "" {
				var err error
				statusFilter, err = parseVCStatus(statusStr)
				if err != nil {
					return err
				}
			}

			typeStr, _ := cmd.Flags().GetString("type")
			var typeFilter vcregistryv1beta1.VCType
			if typeStr != "" {
				var err error
				typeFilter, err = parseVCType(typeStr)
				if err != nil {
					return err
				}
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &vcregistryv1beta1.QueryListUserVCsRequest{
				HolderAddress: holderAddress,
				StatusFilter:  statusFilter,
				TypeFilter:    typeFilter,
				Pagination:    pageReq,
			}

			res, err := queryClient.ListUserVCs(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("type", "", "Filter by VC type")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "user-vcs")
	return cmd
}

// CmdQueryVCStatus checks VC status and validity
func CmdQueryVCStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vc-status [vc-id]",
		Short: "Check if a VC is valid and active",
		Long: `Check the status and validity of a verifiable credential.

Examples:
  aurad query vcregistry vc-status vc-123456

Returns:
  - Current status
  - Whether the VC is valid (active and not expired)
  - Expiration timestamp
  - Revocation details (if revoked)
  - Merkle proof for trustless verification
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			req := &vcregistryv1beta1.QueryCheckVCStatusRequest{
				VcId: args[0],
			}

			res, err := queryClient.CheckVCStatus(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryBatchVCStatus checks status for multiple VCs
func CmdQueryBatchVCStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-vc-status [vc-ids]",
		Short: "Check status of multiple VCs at once",
		Long: `Check the status of multiple verifiable credentials in a single query.

Examples:
  aurad query vcregistry batch-vc-status vc-123,vc-456,vc-789

VC IDs should be comma-separated.

Returns status information for each VC.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			vcIDsStr := args[0]
			vcIDs := strings.Split(vcIDsStr, ",")
			for i := range vcIDs {
				vcIDs[i] = strings.TrimSpace(vcIDs[i])
			}

			req := &vcregistryv1beta1.QueryBatchVCStatusRequest{
				VcIds: vcIDs,
			}

			res, err := queryClient.BatchVCStatus(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVCPolicy queries a specific VC policy
func CmdQueryVCPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy [vc-type-name]",
		Short: "Query a VC policy by type name",
		Long: `Query the policy for a specific verifiable credential type.

Examples:
  aurad query vcregistry policy "Verified Developer"
  aurad query vcregistry policy "High Assurance Focus"

Returns:
  - Policy details
  - CS threshold requirement
  - Required Inclusion Routines
  - Required arena and score
  - Expiration duration
  - Singleton flag
  - Annual renewal requirement
  - Policy status
  - Policy version
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			req := &vcregistryv1beta1.QueryGetVCPolicyRequest{
				VcTypeName: args[0],
			}

			res, err := queryClient.GetVCPolicy(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVCPolicies queries all VC policies
func CmdQueryVCPolicies() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policies",
		Short: "Query all VC policies",
		Long: `Query all verifiable credential policies.

Examples:
  aurad query vcregistry policies
  aurad query vcregistry policies --status VC_POLICY_STATUS_ACTIVE
  aurad query vcregistry policies --status VC_POLICY_STATUS_DEPRECATED

Status filters:
  VC_POLICY_STATUS_DRAFT
  VC_POLICY_STATUS_ACTIVE
  VC_POLICY_STATUS_DEPRECATED

Returns a list of all policies matching the criteria.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			statusStr, _ := cmd.Flags().GetString("status")
			var statusFilter vcregistryv1beta1.VCPolicyStatus
			if statusStr != "" {
				var err error
				statusFilter, err = parseVCPolicyStatus(statusStr)
				if err != nil {
					return err
				}
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &vcregistryv1beta1.QueryListVCPoliciesRequest{
				StatusFilter: statusFilter,
				Pagination:   pageReq,
			}

			res, err := queryClient.ListVCPolicies(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by policy status")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "policies")
	return cmd
}

// CmdQueryRevocationList queries the revocation Merkle tree
func CmdQueryRevocationList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revocation-list",
		Short: "Query the revocation Merkle tree",
		Long: `Query the global revocation list and Merkle root.

Example:
  aurad query vcregistry revocation-list

Returns:
  - Current Merkle root
  - Total number of revocations
  - Last update height and timestamp

This allows trustless verification of VC revocation status.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			req := &vcregistryv1beta1.QueryGetRevocationListRequest{}

			res, err := queryClient.GetRevocationList(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryCheckRevocation checks if a VC is revoked
func CmdQueryCheckRevocation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-revocation [vc-id]",
		Short: "Check if a VC is revoked",
		Long: `Check if a specific verifiable credential is revoked.

Examples:
  aurad query vcregistry check-revocation vc-123456

Returns:
  - Whether the VC is revoked
  - Revocation record details (if revoked)
  - Merkle proof for trustless verification
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			req := &vcregistryv1beta1.QueryCheckRevocationRequest{
				VcId: args[0],
			}

			res, err := queryClient.CheckRevocation(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryResolveDID resolves a DID to its document
func CmdQueryResolveDID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-did [did]",
		Short: "Resolve a DID to its document",
		Long: `Resolve a Decentralized Identifier to its DID document.

Examples:
  aurad query vcregistry resolve-did did:aura:mainnet:user123
  aurad query vcregistry resolve-did did:aura:testnet:user456

Returns:
  - DID document
  - Controller address
  - Verification methods
  - Associated active VCs
  - Service endpoints
  - Metadata URI
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			req := &vcregistryv1beta1.QueryResolveDIDRequest{
				Did: args[0],
			}

			res, err := queryClient.ResolveDID(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryDIDByAddress gets DIDs by controller address
func CmdQueryDIDByAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "did-by-address [controller-address]",
		Short: "Get DIDs controlled by an address",
		Long: `Get all DIDs controlled by a specific address.

Examples:
  aurad query vcregistry did-by-address aura1abc...

Returns:
  - List of DIDs controlled by the address

Note: Multiple DIDs per address are allowed.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			req := &vcregistryv1beta1.QueryGetDIDByAddressRequest{
				Controller: args[0],
			}

			res, err := queryClient.GetDIDByAddress(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryValidateMintEligibility checks if user can mint a VC
func CmdQueryValidateMintEligibility() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-mint [holder-address] [vc-type]",
		Short: "Check if a user is eligible to mint a VC",
		Long: `Validate whether a user meets all requirements to mint a specific VC type.

Examples:
  aurad query vcregistry validate-mint aura1abc... VC_TYPE_VERIFIED_HUMAN
  aurad query vcregistry validate-mint aura1def... VC_TYPE_CUSTOM --custom-type "MyCustomVC"

Returns:
  - Whether the user is eligible
  - Missing requirements (if any)
  - Current and required Confidence Score
  - Completed and required Inclusion Routines
  - Rate limit status

Useful for checking eligibility before attempting to mint.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			holderAddress := args[0]
			vcTypeStr := args[1]

			vcType, err := parseVCType(vcTypeStr)
			if err != nil {
				return err
			}

			customType, _ := cmd.Flags().GetString("custom-type")

			req := &vcregistryv1beta1.QueryValidateMintEligibilityRequest{
				HolderAddress: holderAddress,
				VcType:        vcType,
				VcTypeCustom:  customType,
			}

			res, err := queryClient.ValidateMintEligibility(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("custom-type", "", "Custom VC type name (for VC_TYPE_CUSTOM)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryStats queries registry statistics
func CmdQueryStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Query VC registry statistics",
		Long: `Query overall statistics for the VC registry.

Example:
  aurad query vcregistry stats

Returns:
  - Total VCs minted
  - Total active VCs
  - Total revoked VCs
  - Total expired VCs
  - Total DIDs
  - Total policies
  - VCs by type breakdown

Useful for monitoring registry health and usage.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			req := &vcregistryv1beta1.QueryStatsRequest{}

			res, err := queryClient.Stats(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryParams queries module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query VC registry module parameters",
		Long: `Query the current parameters for the VC registry module.

Example:
  aurad query vcregistry params

Returns:
  - Max VCs per user
  - Max mints per day/hour
  - Default VC expiry duration
  - Revocation Merkle update frequency
  - DID prefix and network
  - Minting and revocation fees
  - Rate limiting settings
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistryv1beta1.NewQueryClient(clientCtx)

			req := &vcregistryv1beta1.QueryParamsRequest{}

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
// Helper functions

func parseVCStatus(s string) (vcregistryv1beta1.VCStatus, error) {
	switch strings.ToUpper(s) {
	case "VC_STATUS_PENDING":
		return vcregistryv1beta1.VCStatus_VC_STATUS_PENDING, nil
	case "VC_STATUS_ACTIVE":
		return vcregistryv1beta1.VCStatus_VC_STATUS_ACTIVE, nil
	case "VC_STATUS_REVOKED":
		return vcregistryv1beta1.VCStatus_VC_STATUS_REVOKED, nil
	case "VC_STATUS_EXPIRED":
		return vcregistryv1beta1.VCStatus_VC_STATUS_EXPIRED, nil
	case "VC_STATUS_SUSPENDED":
		return vcregistryv1beta1.VCStatus_VC_STATUS_SUSPENDED, nil
	default:
		return vcregistryv1beta1.VCStatus_VC_STATUS_UNSPECIFIED, fmt.Errorf("invalid VC status: %s", s)
	}
}

func parseVCPolicyStatus(s string) (vcregistryv1beta1.VCPolicyStatus, error) {
	switch strings.ToUpper(s) {
	case "VC_POLICY_STATUS_DRAFT":
		return vcregistryv1beta1.VCPolicyStatus_VC_POLICY_STATUS_DRAFT, nil
	case "VC_POLICY_STATUS_ACTIVE":
		return vcregistryv1beta1.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE, nil
	case "VC_POLICY_STATUS_DEPRECATED":
		return vcregistryv1beta1.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED, nil
	default:
		return vcregistryv1beta1.VCPolicyStatus_VC_POLICY_STATUS_UNSPECIFIED, fmt.Errorf("invalid VC policy status: %s", s)
	}
}
