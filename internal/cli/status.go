package cli

import (
	"context"

	"github.com/spf13/cobra"

	hostcollect "github.com/baselhusam/bareai-cli/internal/collect/host"
	"github.com/baselhusam/bareai-cli/internal/render"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show host and infrastructure summary",
	Long:  "Display a one-screen summary of host resources and detected infrastructure.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
		defer cancel()

		snap := snapshot.New()
		host, err := hostcollect.Collect(ctx)
		if err != nil {
			snap.Skipped = append(snap.Skipped, snapshot.Skip{
				Component: "host",
				Reason:    err.Error(),
			})
		} else {
			snap.Host = &host
		}

		if opts.JSON {
			return render.WriteJSON(cmd.OutOrStdout(), snap)
		}
		return render.WriteStatus(cmd.OutOrStdout(), snap, opts.NoColor)
	},
}
