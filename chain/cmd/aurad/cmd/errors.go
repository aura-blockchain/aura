// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// CLIError represents an enhanced CLI error with suggestions
type CLIError struct {
	Err         error
	Command     string
	Suggestions []string
	Examples    []string
}

// Error implements the error interface
func (e *CLIError) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Error: %v\n", e.Err))

	if e.Command != "" {
		sb.WriteString(fmt.Sprintf("\nCommand: %s\n", e.Command))
	}

	if len(e.Suggestions) > 0 {
		sb.WriteString("\nSuggestions:\n")
		for _, suggestion := range e.Suggestions {
			sb.WriteString(fmt.Sprintf("  - %s\n", suggestion))
		}
	}

	if len(e.Examples) > 0 {
		sb.WriteString("\nExamples:\n")
		for _, example := range e.Examples {
			sb.WriteString(fmt.Sprintf("  %s\n", example))
		}
	}

	return sb.String()
}

// WrapErrorWithSuggestions wraps an error with helpful suggestions
func WrapErrorWithSuggestions(err error, cmd *cobra.Command, suggestions []string) error {
	if err == nil {
		return nil
	}

	return &CLIError{
		Err:         err,
		Command:     cmd.CommandPath(),
		Suggestions: suggestions,
	}
}

// NewVCRegistryError creates an error with VC Registry specific suggestions
func NewVCRegistryError(err error, context string) error {
	suggestions := []string{
		"Use 'aurad interactive' for a guided wizard",
		"Check your Confidence Score with 'aurad query confidencescore score <address>'",
		"Verify required Inclusion Routines with 'aurad query inclusionroutines user-irs <address>'",
		"View VC policies with 'aurad query vcregistry policies'",
	}

	examples := []string{
		"aurad tx vcregistry mint-vc did:aura:mainnet:user123 VC_TYPE_VERIFIED_HUMAN --from alice",
		"aurad tx vc mint did:aura:mainnet:user123 VC_TYPE_VERIFIED_HUMAN --from alice",
		"aurad query vc vc vc-123456",
	}

	return &CLIError{
		Err:         err,
		Command:     context,
		Suggestions: suggestions,
		Examples:    examples,
	}
}

// NewBridgeError creates an error with Bridge specific suggestions
func NewBridgeError(err error, context string) error {
	suggestions := []string{
		"Use 'aurad interactive' and select 'bridge' for a guided wizard",
		"Check linked addresses with 'aurad query bridge linked-addresses <address>'",
		"Verify you have sufficient balance for locking",
		"Ensure addresses are properly formatted for their respective chains",
	}

	examples := []string{
		"aurad tx bridge link-address aura1abc... paw1def... xai1ghi... --from alice",
		"aurad tx br link aura1abc... paw1def... xai1ghi... --from alice",
		"aurad query br linked-addresses aura1abc...",
	}

	return &CLIError{
		Err:         err,
		Command:     context,
		Suggestions: suggestions,
		Examples:    examples,
	}
}

// NewInclusionRoutinesError creates an error with IR specific suggestions
func NewInclusionRoutinesError(err error, context string) error {
	suggestions := []string{
		"Use 'aurad interactive' and select 'ir' for a guided wizard",
		"List available IRs with 'aurad query inclusionroutines list'",
		"Check your completed IRs with 'aurad query ir user-irs <address>'",
		"Verify prerequisites with 'aurad query ir ir <ir-id>'",
	}

	examples := []string{
		"aurad tx inclusionroutines complete ir-basic-verification --from alice",
		"aurad tx ir complete ir-basic-verification --from alice",
		"aurad query ir list",
	}

	return &CLIError{
		Err:         err,
		Command:     context,
		Suggestions: suggestions,
		Examples:    examples,
	}
}

// NewAuthError creates an error with Auth specific suggestions
func NewAuthError(err error, context string) error {
	suggestions := []string{
		"Check that your key exists with 'aurad keys list'",
		"Verify your account has sufficient balance",
		"Ensure you're using the correct --from flag",
		"Try 'aurad keys add <name>' to create a new key",
	}

	return &CLIError{
		Err:         err,
		Command:     context,
		Suggestions: suggestions,
	}
}

// NewTransactionError creates an error with transaction-specific suggestions
func NewTransactionError(err error, context string) error {
	suggestions := []string{
		"Check gas estimation with --gas=auto",
		"Verify account sequence number",
		"Ensure sufficient balance for fees",
		"Try increasing gas with --gas-adjustment=1.5",
		"Check chain-id matches your network",
	}

	examples := []string{
		"aurad tx ... --gas=auto --gas-adjustment=1.5 --from alice",
		"aurad tx ... --fees=1000uaura --from alice",
		"aurad tx ... --gas-prices=0.025uaura --from alice",
	}

	return &CLIError{
		Err:         err,
		Command:     context,
		Suggestions: suggestions,
		Examples:    examples,
	}
}

// NewQueryError creates an error with query-specific suggestions
func NewQueryError(err error, context string) error {
	suggestions := []string{
		"Verify the resource ID or address is correct",
		"Check that you're connected to the correct node with --node flag",
		"Ensure the resource exists on-chain",
		"Try using the interactive query wizard: 'aurad interactive'",
	}

	return &CLIError{
		Err:         err,
		Command:     context,
		Suggestions: suggestions,
	}
}

// ValidateAddress validates an AURA address and provides helpful error messages
func ValidateAddress(address string) error {
	if address == "" {
		return &CLIError{
			Err: fmt.Errorf("address cannot be empty"),
			Suggestions: []string{
				"Provide a valid AURA address (e.g., aura1...)",
				"Use 'aurad keys list' to see your addresses",
			},
		}
	}

	if !strings.HasPrefix(address, "aura1") {
		return &CLIError{
			Err: fmt.Errorf("invalid address format: %s", address),
			Suggestions: []string{
				"AURA addresses must start with 'aura1'",
				"Example: aura1abc123def456...",
			},
		}
	}

	return nil
}

// ValidateDID validates a DID format and provides helpful error messages
func ValidateDID(did string) error {
	if did == "" {
		return &CLIError{
			Err: fmt.Errorf("DID cannot be empty"),
			Suggestions: []string{
				"Provide a valid DID in format: did:aura:network:identifier",
				"Example: did:aura:mainnet:user123",
			},
		}
	}

	parts := strings.Split(did, ":")
	if len(parts) != 4 || parts[0] != "did" || parts[1] != "aura" {
		return &CLIError{
			Err: fmt.Errorf("invalid DID format: %s", did),
			Suggestions: []string{
				"DIDs must follow format: did:aura:network:identifier",
				"Valid networks: mainnet, testnet, devnet",
				"Example: did:aura:mainnet:user123",
			},
			Examples: []string{
				"did:aura:mainnet:alice",
				"did:aura:testnet:bob123",
				"did:aura:devnet:test456",
			},
		}
	}

	validNetworks := map[string]bool{
		"mainnet": true,
		"testnet": true,
		"devnet":  true,
	}

	if !validNetworks[parts[2]] {
		return &CLIError{
			Err: fmt.Errorf("invalid network in DID: %s", parts[2]),
			Suggestions: []string{
				"Valid networks are: mainnet, testnet, devnet",
			},
		}
	}

	return nil
}

// ValidateVCID validates a VC ID format
func ValidateVCID(vcID string) error {
	if vcID == "" {
		return &CLIError{
			Err: fmt.Errorf("VC ID cannot be empty"),
			Suggestions: []string{
				"Provide a valid VC ID",
				"Query your VCs with 'aurad query vcregistry vcs-by-holder <address>'",
			},
		}
	}

	if !strings.HasPrefix(vcID, "vc-") {
		return &CLIError{
			Err: fmt.Errorf("invalid VC ID format: %s", vcID),
			Suggestions: []string{
				"VC IDs typically start with 'vc-'",
				"Example: vc-123456",
			},
		}
	}

	return nil
}
