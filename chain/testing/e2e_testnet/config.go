// Package e2e_testnet provides end-to-end testing against live AURA testnet infrastructure.
// This package connects to actual running validators and tests real network behavior.
package e2e_testnet

import (
	"fmt"
	"os"
	"time"
)

// ValidatorConfig represents a single validator endpoint
type ValidatorConfig struct {
	Name     string
	Host     string // Hostname or IP (use localhost when running on that server)
	RPCPort  int
	RESTPort int
	GRPCPort int
	P2PPort  int
	Home     string // Validator home directory
}

// TestnetConfig holds the complete testnet configuration
type TestnetConfig struct {
	ChainID    string
	Validators []ValidatorConfig
	Faucet     FaucetConfig
	Explorer   ExplorerConfig
	Timeout    time.Duration
}

// FaucetConfig holds faucet service configuration
type FaucetConfig struct {
	Endpoint string
	Port     int
}

// ExplorerConfig holds explorer service configuration
type ExplorerConfig struct {
	Endpoint string
	Port     int
}

// DefaultTestnetConfig returns the configuration for aura-mvp-1
func DefaultTestnetConfig() *TestnetConfig {
	return &TestnetConfig{
		ChainID: "aura-mvp-1",
		Validators: []ValidatorConfig{
			{
				Name:     "val1",
				Host:     "127.0.0.1", // Local when running on aura-testnet
				RPCPort:  10657,
				RESTPort: 10317,
				GRPCPort: 10090,
				P2PPort:  10656,
				Home:     "~/.aura-val1",
			},
			{
				Name:     "val2",
				Host:     "127.0.0.1",
				RPCPort:  10757,
				RESTPort: 10417,
				GRPCPort: 10190,
				P2PPort:  10756,
				Home:     "~/.aura-val2",
			},
			{
				Name:     "val3",
				Host:     "services-testnet", // SSH alias for remote validator
				RPCPort:  10857,
				RESTPort: 10517,
				GRPCPort: 10290,
				P2PPort:  10856,
				Home:     "~/.aura-val3",
			},
			{
				Name:     "val4",
				Host:     "services-testnet",
				RPCPort:  10957,
				RESTPort: 10617,
				GRPCPort: 10390,
				P2PPort:  10956,
				Home:     "~/.aura-val4",
			},
		},
		Faucet: FaucetConfig{
			Endpoint: "http://127.0.0.1",
			Port:     8081,
		},
		Explorer: ExplorerConfig{
			Endpoint: "http://127.0.0.1",
			Port:     10080,
		},
		Timeout: 60 * time.Second,
	}
}

// GetRPCEndpoint returns the RPC endpoint for a validator
func (v *ValidatorConfig) GetRPCEndpoint() string {
	return fmt.Sprintf("http://%s:%d", v.Host, v.RPCPort)
}

// GetRESTEndpoint returns the REST API endpoint for a validator
func (v *ValidatorConfig) GetRESTEndpoint() string {
	return fmt.Sprintf("http://%s:%d", v.Host, v.RESTPort)
}

// GetGRPCEndpoint returns the gRPC endpoint for a validator
func (v *ValidatorConfig) GetGRPCEndpoint() string {
	return fmt.Sprintf("%s:%d", v.Host, v.GRPCPort)
}

// LoadConfigFromEnv loads configuration overrides from environment variables
func LoadConfigFromEnv(cfg *TestnetConfig) {
	if chainID := os.Getenv("AURA_CHAIN_ID"); chainID != "" {
		cfg.ChainID = chainID
	}
	if timeout := os.Getenv("AURA_TEST_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.Timeout = d
		}
	}
}

// LocalOnlyConfig returns config with only local validators (no SSH required)
// Use this when running tests from the testnet server itself
func LocalOnlyConfig() *TestnetConfig {
	cfg := DefaultTestnetConfig()
	// Keep only validators that are local (127.0.0.1)
	localVals := make([]ValidatorConfig, 0)
	for _, v := range cfg.Validators {
		if v.Host == "127.0.0.1" || v.Host == "localhost" {
			localVals = append(localVals, v)
		}
	}
	cfg.Validators = localVals
	return cfg
}

// DefaultOutputDir returns the default output directory for results
func DefaultOutputDir() string {
	// Try to use ~/testnets/aura-mvp-1/results if it exists
	home := os.Getenv("HOME")
	if home != "" {
		resultsDir := home + "/testnets/aura-mvp-1/results"
		if _, err := os.Stat(resultsDir); err == nil {
			return resultsDir
		}
		// Try to create it
		if err := os.MkdirAll(resultsDir, 0755); err == nil {
			return resultsDir
		}
	}
	return "."
}

// PrimaryValidator returns the first validator (val1) for primary operations
func (c *TestnetConfig) PrimaryValidator() *ValidatorConfig {
	if len(c.Validators) > 0 {
		return &c.Validators[0]
	}
	return nil
}

// GetValidatorByName returns a validator by name
func (c *TestnetConfig) GetValidatorByName(name string) *ValidatorConfig {
	for i := range c.Validators {
		if c.Validators[i].Name == name {
			return &c.Validators[i]
		}
	}
	return nil
}
