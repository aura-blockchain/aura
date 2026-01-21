// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/aequitas/aura/chain/cmd/aurad/cmd/security"
)

// ConfigCmd creates a new config management command
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration profiles",
		Long: `Manage CLI configuration profiles for different environments.
Profiles allow you to quickly switch between different node connections,
networks, and settings.`,
	}

	cmd.AddCommand(
		configListCmd(),
		configSetCmd(),
		configGetCmd(),
		configProfileCmd(),
		configAliasCmd(),
	)

	return cmd
}

func configListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			settings := viper.AllSettings()
			data, err := yaml.Marshal(settings)
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
}

func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			// Validate configuration key and value
			logger := GetSecurityLogger()
			validator := security.NewConfigValidator(logger)

			if err := validator.ValidateKey(key); err != nil {
				return fmt.Errorf("invalid configuration key: %w", err)
			}

			if err := validator.ValidateValue(key, value); err != nil {
				return fmt.Errorf("invalid configuration value: %w", err)
			}

			viper.Set(key, value)
			if err := viper.WriteConfig(); err != nil {
				// If config doesn't exist, create it
				configPath := filepath.Join(GetHomeDir(), "config", "config.toml")
				if err := viper.WriteConfigAs(configPath); err != nil {
					return fmt.Errorf("failed to write config: %w", err)
				}
			}

			fmt.Printf("Set %s = %s\n", key, value)

			logger.SecurityEvent("config_set", map[string]interface{}{
				"key": key,
			})

			return nil
		},
	}
}

func configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := viper.Get(key)
			if value == nil {
				return fmt.Errorf("key not found: %s", key)
			}
			fmt.Printf("%s = %v\n", key, value)
			return nil
		},
	}
}

// Profile management
type Profile struct {
	Name     string            `yaml:"name"`
	NodeURL  string            `yaml:"node_url"`
	ChainID  string            `yaml:"chain_id"`
	Settings map[string]string `yaml:"settings"`
}

type ProfileConfig struct {
	ActiveProfile string              `yaml:"active_profile"`
	Profiles      map[string]*Profile `yaml:"profiles"`
}

func configProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage configuration profiles",
	}

	cmd.AddCommand(
		profileListCmd(),
		profileAddCmd(),
		profileSwitchCmd(),
		profileRemoveCmd(),
	)

	return cmd
}

func getProfileConfigPath() string {
	return filepath.Join(GetHomeDir(), "config", "profiles.yaml")
}

func loadProfiles() (*ProfileConfig, error) {
	configPath := getProfileConfigPath()

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config
		return &ProfileConfig{
			ActiveProfile: "default",
			Profiles: map[string]*Profile{
				"default": {
					Name:     "default",
					NodeURL:  "http://localhost:26657",
					ChainID:  "aura-mvp-1",
					Settings: make(map[string]string),
				},
			},
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ProfileConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveProfiles(config *ProfileConfig) error {
	configPath := getProfileConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644) //nolint:gosec // Config files need to be readable by user
}

func profileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadProfiles()
			if err != nil {
				return err
			}

			fmt.Println("Available Profiles:")
			for name, profile := range config.Profiles {
				active := ""
				if name == config.ActiveProfile {
					active = " (active)"
				}
				fmt.Printf("  %s%s\n", name, active)
				fmt.Printf("    Node URL: %s\n", profile.NodeURL)
				fmt.Printf("    Chain ID: %s\n", profile.ChainID)
			}

			return nil
		},
	}
}

func profileAddCmd() *cobra.Command {
	var nodeURL, chainID string

	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			config, err := loadProfiles()
			if err != nil {
				return err
			}

			if config.Profiles == nil {
				config.Profiles = make(map[string]*Profile)
			}

			config.Profiles[name] = &Profile{
				Name:     name,
				NodeURL:  nodeURL,
				ChainID:  chainID,
				Settings: make(map[string]string),
			}

			if err := saveProfiles(config); err != nil {
				return err
			}

			fmt.Printf("Profile '%s' added successfully\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&nodeURL, "node", "http://localhost:26657", "Node RPC URL")
	cmd.Flags().StringVar(&chainID, "chain-id", "aura-mvp-1", "Chain ID")

	return cmd
}

func profileSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch [name]",
		Short: "Switch to a different profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			config, err := loadProfiles()
			if err != nil {
				return err
			}

			if _, exists := config.Profiles[name]; !exists {
				return fmt.Errorf("profile '%s' does not exist", name)
			}

			config.ActiveProfile = name
			if err := saveProfiles(config); err != nil {
				return err
			}

			profile := config.Profiles[name]
			fmt.Printf("Switched to profile '%s'\n", name)
			fmt.Printf("  Node URL: %s\n", profile.NodeURL)
			fmt.Printf("  Chain ID: %s\n", profile.ChainID)

			return nil
		},
	}
}

func profileRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			config, err := loadProfiles()
			if err != nil {
				return err
			}

			if name == "default" {
				return fmt.Errorf("cannot remove default profile")
			}

			if _, exists := config.Profiles[name]; !exists {
				return fmt.Errorf("profile '%s' does not exist", name)
			}

			delete(config.Profiles, name)

			if config.ActiveProfile == name {
				config.ActiveProfile = "default"
			}

			if err := saveProfiles(config); err != nil {
				return err
			}

			fmt.Printf("Profile '%s' removed successfully\n", name)
			return nil
		},
	}
}

// Alias management
type AliasConfig struct {
	Aliases map[string]string `yaml:"aliases"`
}

func configAliasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage command aliases",
	}

	cmd.AddCommand(
		aliasListCmd(),
		aliasAddCmd(),
		aliasRemoveCmd(),
	)

	return cmd
}

func getAliasConfigPath() string {
	return filepath.Join(GetHomeDir(), "config", "aliases.yaml")
}

func loadAliases() (*AliasConfig, error) {
	configPath := getAliasConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &AliasConfig{
			Aliases: map[string]string{
				"b":   "query block",
				"t":   "query tx",
				"s":   "status",
				"bal": "query bank balances",
			},
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config AliasConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveAliases(config *AliasConfig) error {
	configPath := getAliasConfigPath()

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644) //nolint:gosec // Config files need to be readable by user
}

func aliasListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all command aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadAliases()
			if err != nil {
				return err
			}

			fmt.Println("Command Aliases:")
			for alias, command := range config.Aliases {
				fmt.Printf("  %s -> %s\n", alias, command)
			}

			return nil
		},
	}
}

func aliasAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [alias] [command]",
		Short: "Add a command alias",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			command := args[1:]

			config, err := loadAliases()
			if err != nil {
				return err
			}

			if config.Aliases == nil {
				config.Aliases = make(map[string]string)
			}

			config.Aliases[alias] = fmt.Sprintf("%v", command)

			if err := saveAliases(config); err != nil {
				return err
			}

			fmt.Printf("Alias '%s' added successfully\n", alias)
			return nil
		},
	}
}

func aliasRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [alias]",
		Short: "Remove a command alias",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			config, err := loadAliases()
			if err != nil {
				return err
			}

			if _, exists := config.Aliases[alias]; !exists {
				return fmt.Errorf("alias '%s' does not exist", alias)
			}

			delete(config.Aliases, alias)

			if err := saveAliases(config); err != nil {
				return err
			}

			fmt.Printf("Alias '%s' removed successfully\n", alias)
			return nil
		},
	}
}
