package probe

import (
	"fmt"
	"sort"
	"strings"
)

type metricSpec struct {
	Key    string
	Label  string
	Format string // number, percent
}

var metricsByRuntime = map[string][]metricSpec{
	RuntimeVLLM: {
		{Key: "vllm:num_requests_running", Label: "Running requests", Format: "number"},
		{Key: "vllm:num_requests_waiting", Label: "Queue depth", Format: "number"},
		{Key: "vllm:gpu_cache_usage_perc", Label: "KV cache usage", Format: "percent"},
		{Key: "vllm:avg_generation_throughput_toks_per_s", Label: "Generation tok/s", Format: "number"},
		{Key: "vllm:avg_prompt_throughput_toks_per_s", Label: "Prompt tok/s", Format: "number"},
	},
	RuntimeSGLang: {
		{Key: "sglang:num_running_reqs", Label: "Running requests", Format: "number"},
		{Key: "sglang:num_queue_reqs", Label: "Queue depth", Format: "number"},
		{Key: "sglang:token_usage", Label: "Token usage", Format: "number"},
	},
	RuntimeTriton: {
		{Key: "nv_inference_request_success", Label: "Inference success", Format: "number"},
		{Key: "nv_inference_request_failure", Label: "Inference failure", Format: "number"},
		{Key: "nv_inference_pending_request_count", Label: "Pending requests", Format: "number"},
	},
}

var metricAllowlist = func() []string {
	seen := make(map[string]bool)
	var keys []string
	for _, specs := range metricsByRuntime {
		for _, s := range specs {
			if !seen[s.Key] {
				seen[s.Key] = true
				keys = append(keys, s.Key)
			}
		}
	}
	sort.Strings(keys)
	return keys
}()

// FormatMetrics returns labeled metric lines for display.
func FormatMetrics(runtime string, metrics map[string]float64) []string {
	if len(metrics) == 0 {
		return nil
	}
	specs := metricsByRuntime[runtime]
	if len(specs) == 0 {
		keys := make([]string, 0, len(metrics))
		for k := range metrics {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make([]string, 0, len(keys))
		for _, k := range keys {
			out = append(out, fmt.Sprintf("%s=%.2f", k, metrics[k]))
		}
		return out
	}
	var out []string
	for _, spec := range specs {
		v, ok := metrics[spec.Key]
		if !ok {
			continue
		}
		out = append(out, fmt.Sprintf("%s: %s", spec.Label, formatMetricValue(v, spec.Format)))
	}
	return out
}

func formatMetricValue(v float64, format string) string {
	switch format {
	case "percent":
		if v <= 1 {
			return fmt.Sprintf("%.0f%%", v*100)
		}
		return fmt.Sprintf("%.0f%%", v)
	default:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.2f", v)
	}
}

// MetricsLine joins formatted metrics for inline display.
func MetricsLine(runtime string, metrics map[string]float64) string {
	parts := FormatMetrics(runtime, metrics)
	return strings.Join(parts, ", ")
}
