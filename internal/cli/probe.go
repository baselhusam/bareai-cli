package cli

import (
	"context"

	"github.com/spf13/cobra"

	dockercollect "github.com/baselhusam/bareai-cli/internal/collect/docker"
	gpucollect "github.com/baselhusam/bareai-cli/internal/collect/gpu"
	llmcollect "github.com/baselhusam/bareai-cli/internal/collect/llm"
	"github.com/baselhusam/bareai-cli/internal/config"
	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/render"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
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

		snap := snapshot.New()
		client := probe.NewClient(ctx)

		if probeOpts.Endpoint != "" {
			llm := snapshot.LLM{
				Endpoint: probeOpts.Endpoint,
				Runtime:  probeOpts.Runtime,
				Name:     probeOpts.Runtime,
			}
			adapter := probe.AdapterForRuntime(probeOpts.Runtime)
			if adapter == nil {
				adapter = probe.DetectAdapter(ctx, client, probeOpts.Endpoint)
			}
			if adapter == nil {
				snap.Skipped = append(snap.Skipped, snapshot.Skip{
					Component: "probe",
					Reason:    "unknown runtime for endpoint",
				})
			} else {
				llm.Runtime = adapter.Runtime()
				llm.Name = llm.Runtime
				if probeOpts.Model == "" {
					if models, err := adapter.ListModels(ctx, client, llm.Endpoint); err == nil {
						llm.Models = models
					}
				}
				result := probe.Smoke(ctx, client, llm, adapter, probeOpts.Model, probeOpts.Prompt)
				llm.Probe = &result
				snap.LLMs = []snapshot.LLM{llm}
			}
		} else {
			docker, dockerSkips, err := dockercollect.Collect(ctx)
			snap.Docker = &docker
			snap.Skipped = append(snap.Skipped, dockerSkips...)
			if err != nil {
				snap.Skipped = append(snap.Skipped, snapshot.Skip{
					Component: "docker",
					Reason:    err.Error(),
				})
			}

			gpus, gpuSkips := gpucollect.SnapshotGPU(ctx)
			snap.GPUs = gpus
			snap.Skipped = append(snap.Skipped, gpuSkips...)

			llms, llmSkips, err := llmcollect.Collect(ctx, llmcollect.Input{
				Docker:       snap.Docker,
				GPUs:         snap.GPUs,
				Probe:        true,
				ListModels:   true,
				FetchMetrics: true,
			})
			if err != nil {
				return err
			}
			snap.Skipped = append(snap.Skipped, llmSkips...)
			snap.LLMs = probe.SmokeAll(ctx, client, llms, probeOpts.Model, probeOpts.Prompt)
		}

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
