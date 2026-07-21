package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/baselhusam/bareai-cli/internal/tui"
)

// Options holds global CLI flags shared across commands.
type Options struct {
	JSON    bool
	Timeout time.Duration
	NoColor bool
}

// Global options populated from persistent flags.
var opts Options

// rootCmd is the base command for bareai.
var rootCmd = &cobra.Command{
	Use:   "bareai",
	Short: "Inspect bare-metal AI infrastructure",
	Long: `bareai is a CLI and TUI for solo AI engineers doing AIOps on a single bare-metal box.

Inspect host resources, GPUs, Docker, and local LLM runtimes (Ollama, vLLM, SGLang, Triton, etc.)
without mutating the system.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			return cmd.Help()
		}
		return tui.Run(cmd.Context(), tuiOptionsFromCLI(3*time.Second))
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&opts.JSON, "json", "j", false, "output in JSON format")
	rootCmd.PersistentFlags().DurationVar(&opts.Timeout, "timeout", 10*time.Second, "timeout for probes and API calls")
	rootCmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "disable colored output")

	rootCmd.AddCommand(
		statusCmd,
		gpuCmd,
		llmCmd,
		dockerCmd,
		inspectCmd,
		probeCmd,
		watchCmd,
		versionCmd,
		completionCmd,
	)
}

func stubRun(phase int, name string) error {
	_, err := fmt.Fprintf(os.Stdout, "bareai %s: not implemented yet (Phase %d)\n", name, phase)
	return err
}
