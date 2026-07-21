package cli

import "github.com/spf13/cobra"

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "List discovered LLM runtimes and models",
	Long:  "Discover and display local inference servers (Ollama, vLLM, SGLang, Triton, etc.).",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stubRun(4, "llm")
	},
}
