// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// CompletionCmd creates a new completion command
func CompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for aurad.

To load completions:

Bash:
  $ source <(aurad completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ aurad completion bash > /etc/bash_completion.d/aurad
  # macOS:
  $ aurad completion bash > /usr/local/etc/bash_completion.d/aurad

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ aurad completion zsh > "${fpath[1]}/_aurad"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ aurad completion fish | source

  # To load completions for each session, execute once:
  $ aurad completion fish > ~/.config/fish/completions/aurad.fish

PowerShell:
  PS> aurad completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> aurad completion powershell > aurad.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}
