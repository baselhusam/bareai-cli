package action

import (
	"testing"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestResolveLLMContainer(t *testing.T) {
	snap := testSnap()
	target, err := ResolveTarget(snap, "llm.unreachable", VerbRestart, ResolveHints{Container: "ollama"})
	if err != nil {
		t.Fatal(err)
	}
	if target.Kind != TargetContainer || target.Name != "ollama" {
		t.Fatalf("unexpected target: %+v", target)
	}
}

func TestResolveLLMReprobeEndpoint(t *testing.T) {
	snap := testSnap()
	target, err := ResolveTarget(snap, "llm.no_models", VerbReprobe, ResolveHints{})
	if err != nil {
		t.Fatal(err)
	}
	if target.Kind != TargetEndpoint || target.Endpoint == "" {
		t.Fatalf("unexpected target: %+v", target)
	}
}

func TestResolveRejectsUnknownVerb(t *testing.T) {
	snap := testSnap()
	_, err := ResolveTarget(snap, "llm.unreachable", "deploy", ResolveHints{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildPlanRestart(t *testing.T) {
	target := &Target{Kind: TargetContainer, Name: "ollama", Image: "ollama/ollama", State: "running"}
	res := BuildPlan(Request{Verb: VerbRestart, DryRun: true}, target)
	if res.Would == "" || !res.OK {
		t.Fatalf("unexpected plan: %+v", res)
	}
	if res.Before == nil || res.Before.Name != "ollama" {
		t.Fatalf("unexpected before: %+v", res.Before)
	}
}

func testSnap() *snapshot.Snapshot {
	ok := false
	return &snapshot.Snapshot{
		CollectedAt: time.Now().UTC(),
		Docker: &snapshot.Docker{
			Available: true,
			Containers: []snapshot.DockerContainer{{
				ID:    "abc123",
				Name:  "ollama",
				Image: "ollama/ollama",
				State: "running",
			}},
		},
		LLMs: []snapshot.LLM{{
			Name:          "ollama",
			Endpoint:      "http://127.0.0.1:11434",
			Runtime:       "ollama",
			ContainerID:   "abc123",
			ContainerName: "ollama",
			Health:        &snapshot.ProbeResult{OK: ok},
		}, {
			Name:     "vllm",
			Endpoint: "http://127.0.0.1:8000",
			Runtime:  "vllm",
			Health:   &snapshot.ProbeResult{OK: true},
		}},
	}
}
