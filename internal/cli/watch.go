package cli

import "github.com/spf13/cobra"

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Launch live TUI monitoring dashboard",
	Long:  "Open an interactive Bubble Tea dashboard for browsing and monitoring infrastructure.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stubRun(6, "watch")
	},
}
