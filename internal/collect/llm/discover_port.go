package llm

import (
	"context"

	"github.com/baselhusam/bareai-cli/internal/config"
	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

var defaultPortRuntimes = map[uint16]string{
	11434: probe.RuntimeOllama,
	8000:  probe.RuntimeVLLM,
	30000: probe.RuntimeSGLang,
}

func discoverPorts(ctx context.Context) ([]candidate, []snapshot.Skip) {
	client := probe.NewClient(ctx)
	var out []candidate
	seen := make(map[string]bool)

	for _, endpoint := range config.DiscoveryEndpoints() {
		endpoint = normalizeEndpoint(endpoint)
		if endpoint == "" || seen[endpoint] {
			continue
		}
		adapter := probe.DetectAdapter(ctx, client, endpoint)
		if adapter == nil {
			continue
		}
		seen[endpoint] = true
		out = append(out, candidate{
			priority: 2,
			LLM: snapshot.LLM{
				Runtime:  adapter.Runtime(),
				Name:     displayName(adapter.Runtime()),
				Endpoint: endpoint,
				Source:   sourcePort,
			},
		})
	}

	for _, port := range config.DiscoveryPorts() {
		if port <= 0 || port > 65535 {
			continue
		}
		endpoint := baseURL(uint16(port))
		if seen[endpoint] {
			continue
		}
		adapter := probe.DetectAdapter(ctx, client, endpoint)
		if adapter == nil {
			continue
		}
		seen[endpoint] = true
		name := displayName(adapter.Runtime())
		if rt, ok := defaultPortRuntimes[uint16(port)]; ok && adapter.Runtime() == rt {
			name = displayName(rt)
		}
		out = append(out, candidate{
			priority: 1,
			LLM: snapshot.LLM{
				Runtime:  adapter.Runtime(),
				Name:     name,
				Endpoint: endpoint,
				Source:   sourcePort,
			},
		})
	}
	return out, nil
}
