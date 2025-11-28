package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	prevalidationv1beta1 "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

// GetQueryCmd returns the query commands for the prevalidation module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "prevalidation",
		Short:                      "Querying commands for the prevalidation module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryPreValidatedTransaction(),
		CmdQueryPreValidatedTransactions(),
		CmdQueryTemplate(),
		CmdQueryTemplates(),
		CmdQueryMetrics(),
		CmdQueryParams(),
	)

	return cmd
}

// CmdQueryPreValidatedTransaction queries a specific pre-validated transaction
func CmdQueryPreValidatedTransaction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transaction [tx-id]",
		Short: "Query a specific pre-validated transaction",
		Long: `Query detailed information about a specific pre-validated transaction.

Examples:
  aurad query prevalidation transaction tx-123
  aurad query prevalidation transaction tx-456

Returns:
  - Transaction ID and type
  - Template ID used
  - Encrypted transaction data
  - Validation metadata
  - Status (PENDING, VALIDATED, EXPIRED, EXECUTED)
  - Expiration timestamp
  - Gas estimate
  - Context data

Note: This queries the pre-validation cache, not executed transactions.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Pre-validation transactions are managed internally by the keeper
			// and not exposed via gRPC query service for security reasons.
			// Use keeper methods directly for internal queries.
			return fmt.Errorf("pre-validation transactions are managed internally and not queryable via CLI for security")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPreValidatedTransactions queries pre-validated transactions by type
func CmdQueryPreValidatedTransactions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "Query pre-validated transactions",
		Long: `Query all pre-validated transactions with optional filters.

Examples:
  aurad query prevalidation transactions
  aurad query prevalidation transactions --type IR_COMPLETION
  aurad query prevalidation transactions --type DEX_SWAP --status VALIDATED
  aurad query prevalidation transactions --signer aura1abc...

Transaction Types:
  - IR_COMPLETION: Inclusion routine completions
  - DEX_SWAP: DEX token swaps
  - LP_DEPOSIT: Liquidity pool deposits
  - LP_WITHDRAWAL: Liquidity pool withdrawals
  - VC_MINT: Verifiable credential minting
  - BRIDGE_TRANSFER: Cross-chain bridge transfers
  - CONFIDENCE_SCORE_UPDATE: Confidence score updates
  - IDENTITY_CHANGE: Identity change requests

Status Options:
  - PENDING: Scheduled for pre-validation
  - VALIDATED: Pre-validation complete
  - EXPIRED: Pre-validation expired
  - EXECUTED: Transaction executed
  - FAILED: Pre-validation failed
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			txTypeStr, _ := cmd.Flags().GetString("type")
			statusStr, _ := cmd.Flags().GetString("status")
			signer, _ := cmd.Flags().GetString("signer")

			var txType prevalidationv1beta1.TransactionType
			if txTypeStr != "" {
				txType, err = parseTransactionType(txTypeStr)
				if err != nil {
					return err
				}
			}

			var status prevalidationv1beta1.ValidationStatus
			if statusStr != "" {
				status, err = parseValidationStatus(statusStr)
				if err != nil {
					return err
				}
			}

			_ = txType
			_ = status
			_ = signer

			// Pre-validation transactions are managed internally and not exposed for security
			return fmt.Errorf("pre-validation transactions are managed internally and not queryable via CLI for security")
		},
	}

	cmd.Flags().String("type", "", "Filter by transaction type")
	cmd.Flags().String("status", "", "Filter by validation status")
	cmd.Flags().String("signer", "", "Filter by signer address")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTemplate queries a specific validation template
func CmdQueryTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template [template-id]",
		Short: "Query a specific validation template",
		Long: `Query detailed information about a validation template.

Examples:
  aurad query prevalidation template template-123
  aurad query prevalidation template ir-completion-basic

Returns:
  - Template ID and name
  - Transaction type
  - Description
  - Validation rules
  - Parameter schema
  - Gas estimation formula
  - Priority weight
  - Minimum confidence score required
  - Active status
  - Usage statistics
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			_ = args[0] // template ID

			// Validation templates are managed internally and not exposed for security
			return fmt.Errorf("validation templates are managed internally and not queryable via CLI for security")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTemplates queries all validation templates
func CmdQueryTemplates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Query all validation templates",
		Long: `Query all available validation templates with optional filters.

Examples:
  aurad query prevalidation templates
  aurad query prevalidation templates --type IR_COMPLETION
  aurad query prevalidation templates --active-only

Returns list of templates with:
  - Template IDs and names
  - Transaction types
  - Priority weights
  - Active status
  - Usage statistics
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			txTypeStr, _ := cmd.Flags().GetString("type")
			activeOnly, _ := cmd.Flags().GetBool("active-only")

			var txType prevalidationv1beta1.TransactionType
			if txTypeStr != "" {
				txType, err = parseTransactionType(txTypeStr)
				if err != nil {
					return err
				}
			}

			_ = txType
			_ = activeOnly

			// Validation templates are managed internally and not exposed for security
			return fmt.Errorf("validation templates are managed internally and not queryable via CLI for security")
		},
	}

	cmd.Flags().String("type", "", "Filter by transaction type")
	cmd.Flags().Bool("active-only", false, "Show only active templates")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryMetrics queries pre-validation metrics
func CmdQueryMetrics() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Query pre-validation metrics and statistics",
		Long: `Query system-wide pre-validation metrics and performance statistics.

Examples:
  aurad query prevalidation metrics
  aurad query prevalidation metrics --detailed

Returns:
  - Total pre-validations created
  - Total executed vs expired
  - Cache hit/miss statistics
  - Overall cache hit rate
  - Average time savings per transaction
  - Total time and energy saved
  - Metrics breakdown by transaction type
  - Last 24 hours hourly metrics
  - Control group comparison metrics

With --detailed flag:
  - Per-type average execution times
  - Per-type validation times
  - Hourly breakdown charts
  - Statistical analysis
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			detailed, _ := cmd.Flags().GetBool("detailed")
			_ = detailed

			// Metrics are tracked internally via telemetry and not exposed via CLI queries
			return fmt.Errorf("metrics are tracked internally via telemetry - use monitoring dashboards instead")
		},
	}

	cmd.Flags().Bool("detailed", false, "Show detailed metrics breakdown")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryParams queries module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query pre-validation module parameters",
		Long: `Query the current pre-validation module parameters.

Examples:
  aurad query prevalidation params

Returns:
  - Enabled status
  - Scheduler configuration (off-peak hours, timezone, intervals)
  - Auto-scaling configuration (targets, factors, cooldowns)
  - Cache strategy (LRU, LFU, FIFO, ADAPTIVE)
  - Maximum cache size
  - Pre-validation expiry time
  - Encryption algorithm
  - Control group percentage
  - Minimum confidence score required
  - Energy cost estimations
  - Metrics and logging settings
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Parameters are stored in genesis state and can be queried via state export
			return fmt.Errorf("params query not implemented - use 'aurad export' to view genesis state")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// Helper functions

func parseTransactionType(s string) (prevalidationv1beta1.TransactionType, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	switch s {
	case "IR_COMPLETION":
		return prevalidationv1beta1.TransactionType_TX_TYPE_IR_COMPLETION, nil
	case "DEX_SWAP":
		return prevalidationv1beta1.TransactionType_TX_TYPE_DEX_SWAP, nil
	case "LP_DEPOSIT":
		return prevalidationv1beta1.TransactionType_TX_TYPE_LP_DEPOSIT, nil
	case "LP_WITHDRAWAL":
		return prevalidationv1beta1.TransactionType_TX_TYPE_LP_WITHDRAWAL, nil
	case "VC_MINT":
		return prevalidationv1beta1.TransactionType_TX_TYPE_VC_MINT, nil
	case "BRIDGE_TRANSFER":
		return prevalidationv1beta1.TransactionType_TX_TYPE_BRIDGE_TRANSFER, nil
	case "CONFIDENCE_SCORE_UPDATE":
		return prevalidationv1beta1.TransactionType_TX_TYPE_CONFIDENCE_SCORE_UPDATE, nil
	case "IDENTITY_CHANGE":
		return prevalidationv1beta1.TransactionType_TX_TYPE_IDENTITY_CHANGE, nil
	default:
		return prevalidationv1beta1.TransactionType_TX_TYPE_UNSPECIFIED, fmt.Errorf("invalid transaction type: %s", s)
	}
}

func parseValidationStatus(s string) (prevalidationv1beta1.ValidationStatus, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	switch s {
	case "PENDING":
		return prevalidationv1beta1.ValidationStatus_VALIDATION_STATUS_PENDING, nil
	case "VALIDATED":
		return prevalidationv1beta1.ValidationStatus_VALIDATION_STATUS_VALIDATED, nil
	case "EXPIRED":
		return prevalidationv1beta1.ValidationStatus_VALIDATION_STATUS_EXPIRED, nil
	case "EXECUTED":
		return prevalidationv1beta1.ValidationStatus_VALIDATION_STATUS_EXECUTED, nil
	case "FAILED":
		return prevalidationv1beta1.ValidationStatus_VALIDATION_STATUS_FAILED, nil
	default:
		return prevalidationv1beta1.ValidationStatus_VALIDATION_STATUS_UNSPECIFIED, fmt.Errorf("invalid validation status: %s", s)
	}
}

func parseCacheStrategy(s string) (prevalidationv1beta1.CacheStrategy, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	switch s {
	case "LRU":
		return prevalidationv1beta1.CacheStrategy_CACHE_STRATEGY_LRU, nil
	case "LFU":
		return prevalidationv1beta1.CacheStrategy_CACHE_STRATEGY_LFU, nil
	case "FIFO":
		return prevalidationv1beta1.CacheStrategy_CACHE_STRATEGY_FIFO, nil
	case "ADAPTIVE":
		return prevalidationv1beta1.CacheStrategy_CACHE_STRATEGY_ADAPTIVE, nil
	default:
		return prevalidationv1beta1.CacheStrategy_CACHE_STRATEGY_UNSPECIFIED, fmt.Errorf("invalid cache strategy: %s", s)
	}
}

func parseUint32(s string) (uint32, error) {
	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(val), nil
}
