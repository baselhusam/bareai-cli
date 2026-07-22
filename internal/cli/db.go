package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/render"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "List discovered database instances",
	Long:  "Discover and display local databases (PostgreSQL, Redis, MongoDB, MySQL, Qdrant, Elasticsearch).",
	Example: `  bareai db
  bareai db --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
		defer cancel()

		snap := collect.SnapshotWithOptions(ctx, collect.FullOptions())

		if opts.JSON {
			return render.WriteJSON(cmd.OutOrStdout(), snap)
		}
		return render.WriteDB(cmd.OutOrStdout(), snap, opts.NoColor)
	},
}
