package llm

import (
	"context"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Input holds context for LLM discovery.
type Input struct {
	Docker       *snapshot.Docker
	GPUs         []snapshot.GPU
	Probe        bool
	ListModels   bool
	FetchMetrics bool
}

// candidate is an internal discovery record before verification.
type candidate struct {
	snapshot.LLM
	priority int
}

const (
	sourceDocker  = "docker"
	sourceProcess = "process"
	sourcePort    = "port"
)

// Collect discovers local LLM runtimes and optionally probes them.
func Collect(ctx context.Context, in Input) ([]snapshot.LLM, []snapshot.Skip, error) {
	var skips []snapshot.Skip
	var candidates []candidate

	candidates = append(candidates, discoverDocker(in)...)
	procCandidates, procSkips := discoverProcesses(ctx)
	candidates = append(candidates, procCandidates...)
	skips = append(skips, procSkips...)

	portCandidates, portSkips := discoverPorts(ctx)
	candidates = append(candidates, portCandidates...)
	skips = append(skips, portSkips...)

	merged := mergeCandidates(candidates)
	if len(merged) == 0 {
		return nil, skips, nil
	}

	client := probe.NewClient(ctx)
	out := make([]snapshot.LLM, 0, len(merged))
	for _, c := range merged {
		llm := c.LLM
		adapter := probe.DetectAdapter(ctx, client, llm.Endpoint)
		if adapter == nil {
			if in.Probe {
				skips = append(skips, snapshot.Skip{
					Component: "llm." + llm.Endpoint,
					Reason:    "endpoint not recognized as known runtime",
				})
			}
			continue
		}
		llm.Runtime = adapter.Runtime()
		if llm.Name == "" {
			llm.Name = displayName(llm.Runtime)
		}
		if in.Probe {
			probe.Enrich(ctx, client, &llm, adapter, probe.EnrichOptions{
				ListModels:   in.ListModels,
				FetchMetrics: in.FetchMetrics,
			})
		}
		correlate(&llm, in)
		out = append(out, llm)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Endpoint < out[j].Endpoint
	})
	return out, skips, nil
}

func mergeCandidates(candidates []candidate) []candidate {
	byEndpoint := make(map[string]candidate)
	for _, c := range candidates {
		key := normalizeEndpoint(c.Endpoint)
		if key == "" {
			continue
		}
		c.Endpoint = key
		if existing, ok := byEndpoint[key]; !ok || c.priority > existing.priority {
			byEndpoint[key] = c
		}
	}
	out := make([]candidate, 0, len(byEndpoint))
	for _, c := range byEndpoint {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Endpoint < out[j].Endpoint
	})
	return out
}

func normalizeEndpoint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Hostname() == "" {
		return ""
	}
	port := u.Port()
	if port == "" {
		port = "80"
		if u.Scheme == "https" {
			port = "443"
		}
	}
	return u.Scheme + "://" + u.Hostname() + ":" + port
}

func endpointPort(endpoint string) (uint16, bool) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return 0, false
	}
	port := u.Port()
	if port == "" {
		return 0, false
	}
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, false
	}
	return uint16(n), true
}

func baseURL(port uint16) string {
	return "http://127.0.0.1:" + strconv.Itoa(int(port))
}
