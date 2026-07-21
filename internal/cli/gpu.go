package cli

import "github.com/spf13/cobra"

var gpuCmd = &cobra.Command{
	Use:   "gpu",
	Short: "Show GPU and accelerator details",
	Long:  "Display NVIDIA, AMD, and Apple Silicon accelerator inventory and metrics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stubRun(2, "gpu")
	},
}
