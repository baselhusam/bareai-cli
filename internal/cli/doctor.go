package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/config"
	"github.com/baselhusam/bareai-cli/internal/doctor"
	"github.com/baselhusam/bareai-cli/internal/inspect"
	"github.com/baselhusam/bareai-cli/internal/render"
	"github.com/baselhusam/bareai-cli/internal/version"
)

var doctorOpts struct {
	Severity string
	Share    bool
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run ranked diagnostics with read-only suggestions",
	Long:  "Analyze host, GPU, Docker, LLM, and database state and report ranked findings with what/why/try hints.",
	Example: `  bareai doctor
  bareai doctor --severity warn
  bareai doctor --share
  bareai doctor --json | jq '.findings[] | select(.severity=="warn")'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if opts.JSON && doctorOpts.Share {
			return fmt.Errorf("--json and --share cannot be used together")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
		defer cancel()

		snap := collect.SnapshotWithOptions(ctx, collect.FullOptions())
		inspect.Enrich(snap)
		snap.Findings = doctor.Analyze(snap, doctor.Options{
			MinSeverity: doctorSeverity(cmd),
		})

		if opts.JSON {
			return render.WriteDoctorJSON(cmd.OutOrStdout(), snap)
		}
		if doctorOpts.Share {
			return render.WriteDoctorShare(cmd.OutOrStdout(), snap, version.Version)
		}
		return render.WriteDoctor(cmd.OutOrStdout(), snap, render.DoctorOptions{
			NoColor: opts.NoColor,
			Width:   render.TerminalWidth(cmd.OutOrStdout()),
		})
	},
}

func init() {
	def := config.Default()
	doctorOpts.Severity = def.Doctor.MinSeverity
	doctorCmd.Flags().StringVar(&doctorOpts.Severity, "severity", def.Doctor.MinSeverity, "minimum severity to show (info|warn|critical)")
	doctorCmd.Flags().BoolVar(&doctorOpts.Share, "share", false, "paste-friendly report for GitHub issues, Discord, or gists")
}

func doctorSeverity(cmd *cobra.Command) string {
	if cmd != nil && cmd.Flags().Changed("severity") {
		return doctorOpts.Severity
	}
	if cfg := config.Global(); cfg.Doctor.MinSeverity != "" {
		return cfg.Doctor.MinSeverity
	}
	return doctorOpts.Severity
}
