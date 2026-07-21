package llm

import (
	"context"

	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

var wellKnownPorts = []struct {
	port    uint16
	runtime string
}{
	{11434, probe.RuntimeOllama},
	{8000, probe.RuntimeVLLM},
	{30000, probe.RuntimeSGLang},
}

func discoverPorts(ctx context.Context) ([]candidate, []snapshot.Skip) {
	client := probe.NewClient(ctx)
	var out []candidate
	seen := make(map[string]bool)

	for _, wp := range wellKnownPorts {
		endpoint := baseURL(wp.port)
		if seen[endpoint] {
			continue
		}
		adapter := probe.DetectAdapter(ctx, client, endpoint)
		if adapter == nil {
			continue
		}
		seen[endpoint] = true
		out = append(out, candidate{
			priority: 1,
			LLM: snapshot.LLM{
				Runtime:  adapter.Runtime(),
				Name:     displayName(adapter.Runtime()),
				Endpoint: endpoint,
				Source:   sourcePort,
			},
		})
	}
	return out, nil
}
