package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/render"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show host and infrastructure summary",
	Long:  "Display a one-screen summary of host resources and detected infrastructure.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
		defer cancel()

		snap := collect.Snapshot(ctx)

		if opts.JSON {
			return render.WriteJSON(cmd.OutOrStdout(), snap)
		}
		return render.WriteStatus(cmd.OutOrStdout(), snap, opts.NoColor)
	},
}
