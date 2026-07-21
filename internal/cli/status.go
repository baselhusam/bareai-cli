package cli

import "github.com/spf13/cobra"

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show host and infrastructure summary",
	Long:  "Display a one-screen summary of host resources and detected infrastructure.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stubRun(1, "status")
	},
}
