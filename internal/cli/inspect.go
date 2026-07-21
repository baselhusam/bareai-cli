package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/inspect"
	"github.com/baselhusam/bareai-cli/internal/render"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Show full correlated infrastructure report",
	Long:  "Aggregate host, GPU, Docker, and LLM data into one correlated report.",
	Example: `  bareai inspect
  bareai inspect --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
		defer cancel()

		snap := collect.SnapshotWithOptions(ctx, collect.FullOptions())
		inspect.Enrich(snap)

		if opts.JSON {
			return render.WriteJSON(cmd.OutOrStdout(), snap)
		}
		return render.WriteInspect(cmd.OutOrStdout(), snap, render.InspectOptions{
			NoColor: opts.NoColor,
			Width:   render.TerminalWidth(cmd.OutOrStdout()),
		})
	},
}
