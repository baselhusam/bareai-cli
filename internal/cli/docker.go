package cli

import "github.com/spf13/cobra"

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Show Docker containers, images, and volumes",
	Long:  "Inspect Docker Engine state relevant to AI workloads (read-only).",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stubRun(3, "docker")
	},
}
