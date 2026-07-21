package probe

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

var metricAllowlist = []string{
	"vllm:num_requests_running",
	"vllm:gpu_cache_usage_perc",
	"sglang:num_running_reqs",
}

// FetchMetrics scrapes known Prometheus metrics from the adapter metrics path.
func FetchMetrics(ctx context.Context, client *http.Client, baseURL string, adapter Adapter) map[string]float64 {
	if adapter == nil {
		return nil
	}
	path := adapter.MetricsPath()
	if path == "" {
		return nil
	}
	resp, body, _, err := get(ctx, client, joinURL(baseURL, path))
	if err != nil || resp == nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil
	}
	return parsePrometheusMetrics(string(body), metricAllowlist)
}

func parsePrometheusMetrics(text string, allowlist []string) map[string]float64 {
	allowed := make(map[string]bool, len(allowlist))
	for _, name := range allowlist {
		allowed[name] = true
	}
	out := make(map[string]float64)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		if idx := strings.Index(name, "{"); idx >= 0 {
			name = name[:idx]
		}
		if !allowed[name] {
			continue
		}
		val, err := strconv.ParseFloat(parts[len(parts)-1], 64)
		if err != nil {
			continue
		}
		out[name] = val
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
