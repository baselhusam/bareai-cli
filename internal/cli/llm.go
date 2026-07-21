package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/render"
)

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "List discovered LLM runtimes and models",
	Long:  "Discover and display local inference servers (Ollama, vLLM, SGLang, Triton, etc.).",
	Example: `  bareai llm
  bareai llm --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
		defer cancel()

		snap := collect.SnapshotWithOptions(ctx, collect.FullOptions())

		if opts.JSON {
			return render.WriteJSON(cmd.OutOrStdout(), snap)
		}
		return render.WriteLLM(cmd.OutOrStdout(), snap, opts.NoColor)
	},
}
