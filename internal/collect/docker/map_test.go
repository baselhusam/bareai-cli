package docker

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func TestHasNVIDIARuntime(t *testing.T) {
	tests := []struct {
		name           string
		runtimes       map[string]system.RuntimeWithStatus
		defaultRuntime string
		want           bool
	}{
		{
			name:           "nvidia runtime registered",
			runtimes:       map[string]system.RuntimeWithStatus{"nvidia": {}},
			defaultRuntime: "runc",
			want:           true,
		},
		{
			name:           "default runtime nvidia",
			runtimes:       map[string]system.RuntimeWithStatus{"runc": {}},
			defaultRuntime: "nvidia",
			want:           true,
		},
		{
			name:           "no nvidia runtime",
			runtimes:       map[string]system.RuntimeWithStatus{"runc": {}},
			defaultRuntime: "runc",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasNVIDIARuntime(tt.runtimes, tt.defaultRuntime); got != tt.want {
				t.Fatalf("hasNVIDIARuntime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGPURequested(t *testing.T) {
	tests := []struct {
		name string
		reqs []snapshot.DeviceRequest
		want bool
	}{
		{
			name: "nvidia driver",
			reqs: []snapshot.DeviceRequest{{Driver: "nvidia"}},
			want: true,
		},
		{
			name: "gpu capability",
			reqs: []snapshot.DeviceRequest{{Capabilities: []string{"gpu"}}},
			want: true,
		},
		{
			name: "empty",
			reqs: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gpuRequested(tt.reqs); got != tt.want {
				t.Fatalf("gpuRequested() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapDeviceRequests(t *testing.T) {
	reqs := []container.DeviceRequest{{
		Driver:       "nvidia",
		Count:        1,
		DeviceIDs:    []string{"0"},
		Capabilities: [][]string{{"gpu"}},
	}}

	got := mapDeviceRequests(reqs)
	if len(got) != 1 {
		t.Fatalf("expected 1 request, got %d", len(got))
	}
	if got[0].Driver != "nvidia" {
		t.Fatalf("driver = %q", got[0].Driver)
	}
	if !gpuRequested(got) {
		t.Fatal("expected gpu requested")
	}
}

func TestMapContainer(t *testing.T) {
	summary := types.Container{
		ID:      "abc123def456",
		Names:   []string{"/ollama"},
		Image:   "ollama/ollama:latest",
		State:   "running",
		Status:  "Up 2 hours",
		Created: 1_700_000_000,
		Ports: []types.Port{{
			IP:          "0.0.0.0",
			PrivatePort: 11434,
			PublicPort:  11434,
			Type:        "tcp",
		}},
	}
	inspect := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				Resources: container.Resources{
					DeviceRequests: []container.DeviceRequest{{
						Driver:       "nvidia",
						Capabilities: [][]string{{"gpu"}},
					}},
				},
			},
		},
	}

	got := mapContainer(summary, inspect)
	if got.Name != "ollama" {
		t.Fatalf("name = %q", got.Name)
	}
	if got.ID != "abc123def456" {
		t.Fatalf("id = %q", got.ID)
	}
	if !got.GPURequested {
		t.Fatal("expected gpu requested")
	}
	if len(got.Ports) != 1 || got.Ports[0].PublicPort != 11434 {
		t.Fatalf("ports = %+v", got.Ports)
	}
}

func TestMapInfo(t *testing.T) {
	info := system.Info{
		ServerVersion:  "27.5.1",
		OSType:         "linux",
		Architecture:   "amd64",
		DefaultRuntime: "runc",
		Runtimes: map[string]system.RuntimeWithStatus{
			"runc":   {},
			"nvidia": {},
		},
	}

	got := mapInfo(info, "1.47")
	if !got.Available {
		t.Fatal("expected available")
	}
	if got.ServerVersion != "27.5.1" {
		t.Fatalf("server version = %q", got.ServerVersion)
	}
	if !got.NVIDIARuntime {
		t.Fatal("expected nvidia runtime")
	}
}
