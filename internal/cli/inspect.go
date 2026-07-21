package cli

import "github.com/spf13/cobra"

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Show full correlated infrastructure report",
	Long:  "Aggregate host, GPU, Docker, and LLM data into one correlated report.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stubRun(5, "inspect")
	},
}
