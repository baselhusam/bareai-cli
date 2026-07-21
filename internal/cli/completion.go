package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for bareai.

To load completions:

Bash:
  $ source <(bareai completion bash)

  # or persist:
  $ bareai completion bash > /etc/bash_completion.d/bareai

Zsh:
  $ source <(bareai completion zsh)

  # or persist:
  $ bareai completion zsh > "${fpath[1]}/_bareai"

Fish:
  $ bareai completion fish | source

  # or persist:
  $ bareai completion fish > ~/.config/fish/completions/bareai.fish

PowerShell:
  PS> bareai completion powershell | Out-String | Invoke-Expression

  # or persist:
  PS> bareai completion powershell > bareai.ps1
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		default:
			return cmd.Help()
		}
	},
}

func init() {
	completionCmd.SetOut(os.Stdout)
}
