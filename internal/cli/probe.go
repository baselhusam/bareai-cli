package cli

import "github.com/spf13/cobra"

var probeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Run one-hit smoke tests against discovered LLMs",
	Long:  "Send a lightweight health or completion request to discovered inference endpoints.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stubRun(4, "probe")
	},
}
