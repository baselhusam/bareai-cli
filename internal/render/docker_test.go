package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestWriteDockerUnavailable(t *testing.T) {
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		Docker:      &snapshot.Docker{Available: false},
		Skipped: []snapshot.Skip{{
			Component: "docker",
			Reason:    "Cannot connect to the Docker daemon",
		}},
	}

	var buf bytes.Buffer
	if err := WriteDocker(&buf, snap, true, DockerOptions{}); err != nil {
		t.Fatalf("WriteDocker failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"bareai docker",
		"Docker not available",
		"Cannot connect to the Docker daemon",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestWriteDockerAvailable(t *testing.T) {
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		Docker: &snapshot.Docker{
			Available:      true,
			ServerVersion:  "27.5.1",
			APIVersion:     "1.47",
			OSType:         "linux",
			Architecture:   "amd64",
			DefaultRuntime: "runc",
			NVIDIARuntime:  true,
			Containers: []snapshot.DockerContainer{{
				ID:           "abc123",
				Name:         "ollama",
				Image:        "ollama/ollama:latest",
				State:        "running",
				Status:       "Up 2 hours",
				GPURequested: true,
				Ports: []snapshot.DockerPort{{
					PrivatePort: 11434,
					PublicPort:  11434,
					Type:        "tcp",
					IP:          "0.0.0.0",
				}},
			}},
			Images: []snapshot.DockerImage{{
				ID:       "img123",
				RepoTags: []string{"ollama/ollama:latest"},
				Size:     2 * giB,
			}},
			Volumes: []snapshot.DockerVolume{{
				Name:   "ollama_data",
				Driver: "local",
			}},
		},
	}

	var buf bytes.Buffer
	if err := WriteDocker(&buf, snap, true, DockerOptions{}); err != nil {
		t.Fatalf("WriteDocker failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"Engine: Docker 27.5.1",
		"nvidia: yes",
		"Containers (1 running / 1 total)",
		"ollama",
		"GPU",
		"yes",
		"Images: 1",
		"Volumes: 1",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestWriteDockerWithDetailFlags(t *testing.T) {
	snap := &snapshot.Snapshot{
		CollectedAt: time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC),
		Docker: &snapshot.Docker{
			Available: true,
			Images: []snapshot.DockerImage{{
				ID:       "img123",
				RepoTags: []string{"ollama/ollama:latest"},
				Size:     2 * giB,
			}},
			Volumes: []snapshot.DockerVolume{{
				Name:       "ollama_data",
				Driver:     "local",
				Mountpoint: "/var/lib/docker/volumes/ollama_data/_data",
			}},
		},
	}

	var buf bytes.Buffer
	if err := WriteDocker(&buf, snap, true, DockerOptions{Images: true, Volumes: true}); err != nil {
		t.Fatalf("WriteDocker failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"Images (1)",
		"ollama/ollama:latest",
		"2.0 GiB",
		"Volumes (1)",
		"ollama_data",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestFormatDockerPorts(t *testing.T) {
	got := formatDockerPorts([]snapshot.DockerPort{{
		PrivatePort: 11434,
		PublicPort:  11434,
		Type:        "tcp",
		IP:          "0.0.0.0",
	}})
	if got != "11434->11434/tcp" {
		t.Fatalf("formatDockerPorts() = %q", got)
	}
}

func TestWriteDockerSummary(t *testing.T) {
	var buf bytes.Buffer
	if err := writeDockerSummary(&buf, &snapshot.Docker{
		Available:     true,
		NVIDIARuntime: true,
		Containers: []snapshot.DockerContainer{{
			State: "running",
		}},
		Images: make([]snapshot.DockerImage, 3),
	}); err != nil {
		t.Fatalf("writeDockerSummary failed: %v", err)
	}
	if !strings.Contains(buf.String(), "available — 1 running, 3 images, nvidia runtime") {
		t.Fatalf("unexpected summary: %s", buf.String())
	}
}
