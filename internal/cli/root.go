package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/baselhusam/bareai-cli/internal/config"
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

var appConfig config.Config

// rootCmd is the base command for bareai.
var rootCmd = &cobra.Command{
	Use:   "bareai",
	Short: "Inspect bare-metal AI infrastructure",
	Long: `bareai is a CLI and TUI for solo AI engineers doing AIOps on a single bare-metal box.

Inspect host resources, GPUs, Docker, local databases, and LLM runtimes (Ollama, vLLM, SGLang, Triton, etc.).
Use bareai do for confirm-gated fixes tied to doctor findings; all other commands are read-only.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			return cmd.Help()
		}
		return tui.Run(cmd.Context(), tuiOptionsFromCLI(appConfig.Defaults.Refresh))
	},
}

// Execute runs the root command.
func Execute() error {
	if err := config.Init(); err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	appConfig = config.Global()
	return rootCmd.Execute()
}

// RootCommand exposes the cobra root for man page generation.
func RootCommand() *cobra.Command {
	return rootCmd
}

func init() {
	def := config.Default()
	rootCmd.PersistentFlags().BoolVarP(&opts.JSON, "json", "j", false, "output in JSON format")
	rootCmd.PersistentFlags().DurationVar(&opts.Timeout, "timeout", def.Defaults.Timeout, "timeout for probes and API calls")
	rootCmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", def.Defaults.NoColor, "disable colored output")

	rootCmd.AddCommand(
		statusCmd,
		gpuCmd,
		llmCmd,
		dbCmd,
		dockerCmd,
		inspectCmd,
		probeCmd,
		watchCmd,
		doctorCmd,
		doCmd,
		mcpCmd,
		configCmd,
		versionCmd,
		completionCmd,
	)
}
