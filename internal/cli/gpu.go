package cli

import (
	"context"

	"github.com/spf13/cobra"

	gpucollect "github.com/baselhusam/bareai-cli/internal/collect/gpu"
	"github.com/baselhusam/bareai-cli/internal/render"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

var gpuCmd = &cobra.Command{
	Use:   "gpu",
	Short: "Show GPU and accelerator details",
	Long:  "Display NVIDIA, AMD, and Apple Silicon accelerator inventory and metrics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
		defer cancel()

		snap := snapshot.New()
		gpus, skips := gpucollect.SnapshotGPU(ctx)
		snap.GPUs = gpus
		snap.Skipped = skips

		if opts.JSON {
			return render.WriteJSON(cmd.OutOrStdout(), snap)
		}
		return render.WriteGPU(cmd.OutOrStdout(), snap, opts.NoColor)
	},
}
