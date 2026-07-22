package rules

import (
	"fmt"

	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

const (
	metricQueueHighThreshold = 5
	metricKVCacheFull        = 0.95
)

// LLM returns LLM runtime findings.
func LLM(snap *snapshot.Snapshot) []snapshot.Finding {
	if snap == nil {
		return nil
	}
	var out []snapshot.Finding

	for _, llm := range snap.LLMs {
		if llm.Health != nil && !llm.Health.OK {
			try := fmt.Sprintf("curl -s %s  ·  bareai probe --endpoint %s --runtime %s", llm.Endpoint, llm.Endpoint, llm.Runtime)
			if llm.Runtime == probe.RuntimeOllama {
				try = fmt.Sprintf("curl -s %s/api/tags  ·  systemctl status ollama  ·  docker ps --filter name=ollama", llm.Endpoint)
			}
			do := containerOffers(llm.ContainerID, llm.ContainerName, "logs", "restart")
			do = append(do, endpointOffer(llm.Endpoint, "reprobe")...)
			out = append(out, findingWithDo(
				"llm.unreachable",
				SeverityWarn,
				"llm",
				rankWarn,
				fmt.Sprintf("%s (%s) is unreachable", llm.Name, llm.Endpoint),
				"Health probe failed; endpoint may be down or blocked.",
				try,
				do,
			))
			continue
		}

		if llm.Health != nil && llm.Health.OK && len(llm.Models) == 0 {
			out = append(out, findingWithDo(
				"llm.no_models",
				SeverityWarn,
				"llm",
				rankWarn+25,
				fmt.Sprintf("%s (%s) is healthy but no models are loaded", llm.Name, llm.Endpoint),
				"Server responds but model list is empty.",
				fmt.Sprintf("bareai llm --json  ·  bareai probe --endpoint %s", llm.Endpoint),
				endpointOffer(llm.Endpoint, "reprobe"),
			))
		}

		if cache, ok := llm.Metrics["vllm:gpu_cache_usage_perc"]; ok && cache > metricKVCacheFull {
			out = append(out, finding(
				"llm.kv_cache_full",
				SeverityWarn,
				"llm",
				rankWarn+30,
				fmt.Sprintf("%s KV cache above 95%% (%.0f%%)", llm.Name, cache*100),
				"KV cache pressure can reduce batch size and increase latency.",
				fmt.Sprintf("bareai llm --json  ·  curl -s %s/metrics | grep gpu_cache", llm.Endpoint),
			))
		}

		if waiting, ok := llm.Metrics["vllm:num_requests_waiting"]; ok && waiting >= metricQueueHighThreshold {
			out = append(out, finding(
				"llm.metrics_queue_high",
				SeverityWarn,
				"llm",
				rankWarn+35,
				fmt.Sprintf("%s has %g waiting requests", llm.Name, waiting),
				"Request queue depth indicates overload or slow generation.",
				fmt.Sprintf("bareai llm --json  ·  curl -s %s/metrics | grep num_requests", llm.Endpoint),
			))
		}
		if queue, ok := llm.Metrics["sglang:num_queue_reqs"]; ok && queue >= metricQueueHighThreshold {
			out = append(out, finding(
				"llm.metrics_queue_high",
				SeverityWarn,
				"llm",
				rankWarn+35,
				fmt.Sprintf("%s has %g queued requests", llm.Name, queue),
				"Request queue depth indicates overload or slow generation.",
				fmt.Sprintf("bareai llm --json  ·  curl -s %s/metrics | grep num_queue", llm.Endpoint),
			))
		}
	}

	if len(snap.LLMs) > 2 {
		out = append(out, finding(
			"llm.multiple_runtimes",
			SeverityInfo,
			"llm",
			rankInfo,
			fmt.Sprintf("%d LLM runtimes discovered on this host", len(snap.LLMs)),
			"Multiple inference servers may compete for GPU memory and ports.",
			"bareai inspect  ·  bareai llm --json",
		))
	}

	return out
}
