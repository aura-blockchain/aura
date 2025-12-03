package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "compliance",
		Aliases:                    []string{"kyc", "comp"},
		Short:                      "Compliance transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdSubmitKYC(),
		CmdReportSuspiciousActivity(),
		CmdScreenSanctions(),
		CmdRecordGDPRConsent(),
		CmdRequestGDPRData(),
		CmdGenerateTaxReport(),
	)

	return cmd
}

// CmdSubmitKYC submits KYC verification
func CmdSubmitKYC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-kyc [address] [kyc-level] [provider] [pii-commitment-hex] [jurisdiction]",
		Short: "Submit KYC verification for an address",
		Long: `Submit Know Your Customer (KYC) verification for an address.

This command submits a KYC record using GDPR-compliant commitment-based storage.
The PII commitment should be a 64-character hex string (SHA-256 hash of off-chain PII data).
The jurisdiction must be a 2-letter ISO 3166-1 alpha-2 country code.

KYC Levels:
  1 - NONE: No KYC verification
  2 - BASIC: Basic identity verification
  3 - INTERMEDIATE: Government ID verification
  4 - ADVANCED: Enhanced due diligence

OFAC Compliance:
  Jurisdictions from OFAC-sanctioned countries will be rejected (e.g., KP, IR, SY, CU, RU, BY).

Example:
  aurad tx compliance submit-kyc aura1abc... 3 cosmos1provider... a1b2c3d4...f0 US --from provider
`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			kycLevelInt, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid kyc level: %w", err)
			}
			provider := args[2]
			piiCommitmentHex := args[3]
			jurisdiction := args[4]

			// Parse PII commitment from hex (should be 64 hex chars = 32 bytes)
			if len(piiCommitmentHex) != 64 {
				return fmt.Errorf("pii-commitment must be 64 hex characters (32 bytes SHA-256 hash)")
			}
			piiCommitment := make([]byte, 32)
			_, err = fmt.Sscanf(piiCommitmentHex, "%x", &piiCommitment)
			if err != nil {
				return fmt.Errorf("invalid pii-commitment hex string: %w", err)
			}

			// Validate jurisdiction format
			if len(jurisdiction) != 2 {
				return fmt.Errorf("jurisdiction must be 2-letter ISO 3166-1 alpha-2 country code")
			}

			msg := &v1beta1.MsgSubmitKYC{
				Address:       address,
				KycLevel:      v1beta1.KYCLevel(kycLevelInt),
				Provider:      provider,
				PiiCommitment: piiCommitment,
				Jurisdiction:  jurisdiction,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdReportSuspiciousActivity reports suspicious activity
func CmdReportSuspiciousActivity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report-suspicious [address] [tx-hash] [activity-type] [description]",
		Short: "Report suspicious activity",
		Long: `Report suspicious activity for compliance monitoring and potential SAR filing.

Activity Types: structuring, smurfing, unusual_pattern, high_risk_jurisdiction, layering

Example:
  aurad tx compliance report-suspicious aura1abc... "ABC123..." "structuring" "Multiple small transactions" --indicators="velocity,amount_pattern" --from alice
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			txHash := args[1]
			activityType := args[2]
			description := args[3]

			// Get optional indicators
			indicatorsStr, _ := cmd.Flags().GetString("indicators")
			var indicators []string
			if indicatorsStr != "" {
				indicators = strings.Split(indicatorsStr, ",")
			}

			msg := &v1beta1.MsgReportSuspiciousActivity{
				Reporter:        clientCtx.GetFromAddress().String(),
				Address:         address,
				TransactionHash: txHash,
				ActivityType:    activityType,
				Description:     description,
				Indicators:      indicators,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("indicators", "", "Comma-separated list of risk indicators")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdScreenSanctions screens address against sanctions lists
func CmdScreenSanctions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screen-sanctions [address]",
		Short: "Screen address against sanctions lists",
		Long: `Screen an address against global sanctions lists (OFAC SDN, EU, UN, etc).

Example:
  aurad tx compliance screen-sanctions aura1abc... --force-refresh --from alice
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			forceRefresh, _ := cmd.Flags().GetBool("force-refresh")

			msg := &v1beta1.MsgScreenSanctions{
				Address:      address,
				ForceRefresh: forceRefresh,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool("force-refresh", false, "Force new screening instead of using cache")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdRecordGDPRConsent records GDPR consent
func CmdRecordGDPRConsent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record-consent [address] [consent-type] [consented] [version]",
		Short: "Record GDPR consent",
		Long: `Record GDPR consent for data processing.

Consent Types: data_processing, marketing, analytics, third_party_sharing

Example:
  aurad tx compliance record-consent aura1abc... "data_processing" true "v1.2" --from alice
  aurad tx compliance record-consent aura1abc... "marketing" false "v1.2" --from alice
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			consentType := args[1]
			consented, err := strconv.ParseBool(args[2])
			if err != nil {
				return fmt.Errorf("invalid consented value (must be true/false): %w", err)
			}
			version := args[3]

			msg := &v1beta1.MsgRecordGDPRConsent{
				Address:        address,
				ConsentType:    consentType,
				Consented:      consented,
				ConsentVersion: version,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRequestGDPRData requests GDPR data
func CmdRequestGDPRData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-data [address] [request-type]",
		Short: "Request GDPR data",
		Long: `Request GDPR data access, rectification, erasure, or portability.

Request Types: access, rectification, erasure, portability

Example:
  aurad tx compliance request-data aura1abc... "access" --from alice
  aurad tx compliance request-data aura1abc... "erasure" --from alice
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			requestType := args[1]

			msg := &v1beta1.MsgRequestGDPRData{
				Address:     address,
				RequestType: requestType,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdGenerateTaxReport generates a tax report
func CmdGenerateTaxReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-tax-report [address] [tax-year] [jurisdiction] [report-type]",
		Short: "Generate tax report",
		Long: `Generate tax report for specified year and jurisdiction.

Report Types: 1099-MISC, 1099-K, 8949, comprehensive

Example:
  aurad tx compliance generate-tax-report aura1abc... "2024" "US" "1099-MISC" --from alice
  aurad tx compliance generate-tax-report aura1abc... "2024" "UK" "comprehensive" --from alice
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			taxYear := args[1]
			jurisdiction := args[2]
			reportType := args[3]

			msg := &v1beta1.MsgGenerateTaxReport{
				Address:      address,
				TaxYear:      taxYear,
				Jurisdiction: jurisdiction,
				ReportType:   reportType,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
