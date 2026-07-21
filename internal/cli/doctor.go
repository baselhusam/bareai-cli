package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/config"
	"github.com/baselhusam/bareai-cli/internal/doctor"
	"github.com/baselhusam/bareai-cli/internal/inspect"
	"github.com/baselhusam/bareai-cli/internal/render"
)

var doctorOpts struct {
	Severity string
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run ranked diagnostics with read-only suggestions",
	Long:  "Analyze host, GPU, Docker, and LLM state and report ranked findings with what/why/try hints.",
	Example: `  bareai doctor
  bareai doctor --severity warn
  bareai doctor --json | jq '.findings[] | select(.severity=="warn")'`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
