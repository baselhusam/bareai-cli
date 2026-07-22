package db

import (
	"testing"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestDiscoverDocker(t *testing.T) {
	in := Input{
		Docker: &snapshot.Docker{
			Available: true,
			Containers: []snapshot.DockerContainer{
				{
					ID:    "abc123",
					Name:  "postgres-main",
					Image: "postgres:16",
					State: "running",
					PID:   1234,
					Ports: []snapshot.DockerPort{
						{PrivatePort: 5432, PublicPort: 5432, Type: "tcp"},
					},
				},
				{
					ID:    "def456",
					Name:  "nginx",
					Image: "nginx:latest",
					State: "running",
					Ports: []snapshot.DockerPort{
						{PrivatePort: 80, PublicPort: 8080, Type: "tcp"},
					},
				},
			},
		},
	}

	candidates := discoverDocker(in)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	c := candidates[0]
	if c.Engine != EnginePostgres {
		t.Errorf("engine = %q, want postgres", c.Engine)
	}
	if c.Address != "127.0.0.1:5432" {
		t.Errorf("address = %q, want 127.0.0.1:5432", c.Address)
	}
	if c.Source != sourceDocker {
		t.Errorf("source = %q, want docker", c.Source)
	}
	if c.ContainerName != "postgres-main" {
		t.Errorf("container = %q", c.ContainerName)
	}
}

func TestMergeCandidates(t *testing.T) {
	merged := mergeCandidates([]candidate{
		{priority: 1, Database: snapshot.Database{Engine: EngineRedis, Address: "127.0.0.1:6379", Source: sourcePort}},
		{priority: 3, Database: snapshot.Database{Engine: EngineRedis, Address: "127.0.0.1:6379", Source: sourceDocker, ContainerName: "redis"}},
	})
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged, got %d", len(merged))
	}
	if merged[0].Source != sourceDocker {
		t.Errorf("expected docker source to win, got %q", merged[0].Source)
	}
}
