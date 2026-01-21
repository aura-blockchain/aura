// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// StatusCmd returns the status command for the Aura daemon
func StatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query the remote node for status",
		Long: `Query the remote node for status information.

This command will connect to the RPC endpoint and retrieve the current
status of the node, including sync status, latest block, and other metrics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return queryStatus(cmd)
		},
	}

	cmd.Flags().String("node", DefaultRPCAddress, "address of the node to query")
	viper.BindPFlag("node", cmd.Flags().Lookup("node"))

	return cmd
}

// NodeStatus represents the node status response
type NodeStatus struct {
	Chain     string    `json:"chain"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	GRPCAddr  string    `json:"grpc_address,omitempty"`
	APIAddr   string    `json:"api_address,omitempty"`
	RPCAddr   string    `json:"rpc_address,omitempty"`
}

// queryStatus queries the node status
func queryStatus(cmd *cobra.Command) error {
	nodeAddr := viper.GetString("node")

	// Try to get status from API server
	apiAddr := viper.GetString(FlagAPIAddress)
	if apiAddr == "" {
		apiAddr = DefaultAPIAddress
	}

	fmt.Printf("Querying node status...\n")
	fmt.Printf("API Address: %s\n\n", apiAddr)

	// Make HTTP request to API server
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("http://%s/", apiAddr))
	if err != nil {
		return fmt.Errorf("failed to connect to node: %w\nIs the node running? Try: aurad start", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("node returned non-OK status: %d", resp.StatusCode)
	}

	var status NodeStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return fmt.Errorf("failed to decode status response: %w", err)
	}

	// Print status
	fmt.Printf("Node Status:\n")
	fmt.Printf("  Chain:  %s\n", status.Chain)
	fmt.Printf("  Status: %s\n", status.Status)
	fmt.Printf("\n")
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  gRPC: %s\n", viper.GetString(FlagGRPCAddress))
	fmt.Printf("  API:  %s\n", apiAddr)
	fmt.Printf("  RPC:  %s\n", nodeAddr)

	return nil
}
