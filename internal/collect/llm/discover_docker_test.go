package llm

import (
	"testing"
	"time"

	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestDiscoverDockerOllama(t *testing.T) {
	in := Input{
		Docker: &snapshot.Docker{
			Available: true,
			Containers: []snapshot.DockerContainer{{
				Name:  "ollama",
				Image: "ollama/ollama:latest",
				State: "running",
				PID:   100,
				Ports: []snapshot.DockerPort{{
					PublicPort:  11434,
					PrivatePort: 11434,
					Type:        "tcp",
				}},
			}},
		},
	}
	candidates := discoverDocker(in)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].Runtime != probe.RuntimeOllama {
		t.Fatalf("runtime = %q", candidates[0].Runtime)
	}
	if candidates[0].Endpoint != "http://127.0.0.1:11434" {
		t.Fatalf("endpoint = %q", candidates[0].Endpoint)
	}
}

func TestMergeCandidatesPrefersDocker(t *testing.T) {
	merged := mergeCandidates([]candidate{
		{priority: 1, LLM: snapshot.LLM{Endpoint: "http://127.0.0.1:11434", Source: sourcePort}},
		{priority: 3, LLM: snapshot.LLM{Endpoint: "http://127.0.0.1:11434", Source: sourceDocker, ContainerName: "ollama"}},
	})
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged candidate")
	}
	if merged[0].Source != sourceDocker {
		t.Fatalf("source = %q", merged[0].Source)
	}
}

func TestCorrelateGPU(t *testing.T) {
	idx := 0
	llm := &snapshot.LLM{PID: 42}
	matchGPU(llm, []snapshot.GPU{{
		Index: 0,
		Processes: []snapshot.GPUProcess{{
			PID:  42,
			Name: "python",
		}},
	}}, nil)
	if llm.GPUIndex == nil || *llm.GPUIndex != idx {
		t.Fatalf("gpu index = %v", llm.GPUIndex)
	}
}

func TestNormalizeEndpoint(t *testing.T) {
	if got := normalizeEndpoint("127.0.0.1:11434"); got != "http://127.0.0.1:11434" {
		t.Fatalf("got %q", got)
	}
}

func TestDiscoverDockerIgnoresStopped(t *testing.T) {
	in := Input{Docker: &snapshot.Docker{Available: true, Containers: []snapshot.DockerContainer{{
		Name: "ollama", Image: "ollama/ollama", State: "exited", Created: time.Now(),
	}}}}
	if len(discoverDocker(in)) != 0 {
		t.Fatal("expected no candidates for stopped container")
	}
}
