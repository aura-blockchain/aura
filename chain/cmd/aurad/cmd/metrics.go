// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

// StartMetricsServer starts the Prometheus metrics HTTP server
func StartMetricsServer(port int) error {
	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", port)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(fmt.Sprintf("Failed to start metrics server: %v", err))
		}
	}()

	fmt.Printf("✓ Prometheus metrics server started on http://localhost:%d/metrics\n", port)
	return nil
}

// NewMetricsCmd returns the metrics server command
func NewMetricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Start standalone Prometheus metrics server",
		Long:  `Start a standalone HTTP server to expose Prometheus metrics`,
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			fmt.Printf("Starting Prometheus metrics server on port %d...\n", port)

			if err := StartMetricsServer(port); err != nil {
				return err
			}

			// Block forever
			select {}
		},
	}

	cmd.Flags().Int("port", 26660, "Port for metrics HTTP server")

	return cmd
}
