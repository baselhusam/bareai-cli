package probe

import "testing"

func TestFormatMetricsVLLM(t *testing.T) {
	metrics := map[string]float64{
		"vllm:num_requests_running":     2,
		"vllm:gpu_cache_usage_perc":       0.87,
		"vllm:num_requests_waiting":       5,
		"vllm:avg_generation_throughput_toks_per_s": 120,
	}
	got := FormatMetrics(RuntimeVLLM, metrics)
	if len(got) < 3 {
		t.Fatalf("expected formatted metrics, got %v", got)
	}
	if got[0] == "" {
		t.Fatal("empty metric line")
	}
}

func TestMetricsLine(t *testing.T) {
	line := MetricsLine(RuntimeSGLang, map[string]float64{
		"sglang:num_running_reqs": 1,
		"sglang:num_queue_reqs":   3,
	})
	if line == "" {
		t.Fatal("expected metrics line")
	}
}
