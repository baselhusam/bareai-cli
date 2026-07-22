package rules

import (
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestLLMUnreachableDockerOffers(t *testing.T) {
	ok := false
	findings := LLM(&snapshot.Snapshot{
		LLMs: []snapshot.LLM{{
			Name:          "ollama",
			Endpoint:      "http://127.0.0.1:11434",
			ContainerID:   "abc",
			ContainerName: "ollama",
			Health:        &snapshot.ProbeResult{OK: ok},
		}},
	})
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if len(findings[0].Do) < 2 {
		t.Fatalf("expected docker do offers, got %+v", findings[0].Do)
	}
}

func TestLLMUnreachableHostNoDoOffers(t *testing.T) {
	ok := false
	findings := LLM(&snapshot.Snapshot{
		LLMs: []snapshot.LLM{{
			Name:     "ollama",
			Endpoint: "http://127.0.0.1:11434",
			Health:   &snapshot.ProbeResult{OK: ok},
		}},
	})
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if len(findings[0].Do) != 1 || findings[0].Do[0].Verb != "reprobe" {
		t.Fatalf("expected reprobe-only offer, got %+v", findings[0].Do)
	}
}
