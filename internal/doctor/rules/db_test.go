package rules

import (
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestDBUnreachable(t *testing.T) {
	snap := &snapshot.Snapshot{
		Databases: []snapshot.Database{{
			Name:    "Postgres",
			Address: "127.0.0.1:5432",
			Engine:  "postgres",
			Health:  &snapshot.ProbeResult{OK: false},
		}},
	}
	findings := DB(snap)
	if len(findings) == 0 || findings[0].ID != "db.unreachable" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestDBDockerSourceNoEngine(t *testing.T) {
	snap := &snapshot.Snapshot{
		Databases: []snapshot.Database{{
			Name:   "Redis",
			Source: "docker",
			Health: &snapshot.ProbeResult{OK: true},
		}},
	}
	findings := DB(snap)
	found := false
	for _, f := range findings {
		if f.ID == "db.docker_source_no_engine" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected docker source finding, got %+v", findings)
	}
}

func TestEmptyBoxFinding(t *testing.T) {
	snap := &snapshot.Snapshot{
		Host: &snapshot.Host{Hostname: "dev"},
	}
	findings := EmptyBox(snap)
	if len(findings) != 1 || findings[0].ID != "host.empty_box" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestEmptyBoxSkippedWhenLLM(t *testing.T) {
	snap := &snapshot.Snapshot{
		Host: &snapshot.Host{Hostname: "dev"},
		LLMs: []snapshot.LLM{{Endpoint: "http://127.0.0.1:11434"}},
	}
	if findings := EmptyBox(snap); len(findings) != 0 {
		t.Fatalf("expected no empty box finding, got %+v", findings)
	}
}
