package cli

import (
	"context"

	"github.com/spf13/cobra"

	dockercollect "github.com/baselhusam/bareai-cli/internal/collect/docker"
	"github.com/baselhusam/bareai-cli/internal/render"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

var dockerOpts struct {
	All     bool
	Images  bool
	Volumes bool
}

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Show Docker containers, images, and volumes",
	Long:  "Inspect Docker Engine state relevant to AI workloads (read-only).",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
		defer cancel()

		snap := snapshot.New()
		docker, skips, err := dockercollect.Collect(ctx)
		if err != nil {
			return err
		}
		snap.Docker = &docker
		snap.Skipped = skips

		if opts.JSON {
			return render.WriteJSON(cmd.OutOrStdout(), snap)
		}
		return render.WriteDocker(cmd.OutOrStdout(), snap, opts.NoColor, render.DockerOptions{
			All:     dockerOpts.All,
			Images:  dockerOpts.Images,
			Volumes: dockerOpts.Volumes,
		})
	},
}

func init() {
	dockerCmd.Flags().BoolVarP(&dockerOpts.All, "all", "a", false, "include non-running containers in human output")
	dockerCmd.Flags().BoolVar(&dockerOpts.Images, "images", false, "show image list in human output")
	dockerCmd.Flags().BoolVar(&dockerOpts.Volumes, "volumes", false, "show volume list in human output")
}
