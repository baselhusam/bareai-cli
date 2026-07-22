package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/config"
	bareaimcp "github.com/baselhusam/bareai-cli/internal/mcp"
	"github.com/baselhusam/bareai-cli/internal/render"
)

var probeOpts struct {
	Endpoint string
	Runtime  string
	Model    string
	Prompt   string
}

var probeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Run one-hit smoke tests against discovered LLMs",
	Long:  "Send a lightweight health or completion request to discovered inference endpoints.",
	Example: `  bareai probe
  bareai probe --endpoint http://127.0.0.1:11434 --runtime ollama`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
		defer cancel()

		snap := bareaimcp.RunProbeSnapshot(ctx, bareaimcp.ProbeOptions{
			Endpoint: probeOpts.Endpoint,
			Runtime:  probeOpts.Runtime,
			Model:    probeOpts.Model,
			Prompt:   probeOpts.Prompt,
		})

		if opts.JSON {
			return render.WriteJSON(cmd.OutOrStdout(), snap)
		}
		return render.WriteProbe(cmd.OutOrStdout(), snap, opts.NoColor)
	},
}

func init() {
	def := config.Default()
	probeCmd.Flags().StringVar(&probeOpts.Endpoint, "endpoint", "", "probe a specific endpoint URL")
	probeCmd.Flags().StringVar(&probeOpts.Runtime, "runtime", "", "runtime adapter when using --endpoint (ollama|vllm|sglang|triton)")
	probeCmd.Flags().StringVar(&probeOpts.Model, "model", def.Probe.Model, "model name for smoke request")
	probeCmd.Flags().StringVar(&probeOpts.Prompt, "prompt", def.Probe.Prompt, "prompt text for smoke request")
}
