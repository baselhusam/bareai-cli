package cli

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/config"
	"github.com/baselhusam/bareai-cli/internal/tui"
)

var watchRefresh time.Duration

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Launch live TUI monitoring dashboard",
	Long:  "Open an interactive Bubble Tea dashboard for browsing and monitoring infrastructure.",
	Example: `  bareai watch
  bareai watch --refresh 5s`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run(cmd.Context(), tuiOptionsFromCLI(watchRefresh))
	},
}

func init() {
	def := config.Default()
	watchCmd.Flags().DurationVar(&watchRefresh, "refresh", def.Defaults.Refresh, "interval between snapshot refreshes")
}

func tuiOptionsFromCLI(refresh time.Duration) tui.Options {
	return tui.Options{
		Timeout: opts.Timeout,
		Refresh: refresh,
		NoColor: opts.NoColor,
	}
}
