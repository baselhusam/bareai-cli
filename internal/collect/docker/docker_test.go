package docker

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
)

type mockAPIClient struct {
	containerListErr error
}

func (m *mockAPIClient) Ping(ctx context.Context) error { return nil }

func (m *mockAPIClient) Info(ctx context.Context) (system.Info, error) {
	return system.Info{
		ServerVersion:  "27.5.1",
		OSType:         "linux",
		Architecture:   "amd64",
		DefaultRuntime: "runc",
	}, nil
}

func (m *mockAPIClient) ClientVersion() string { return "1.47" }

func (m *mockAPIClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	if m.containerListErr != nil {
		return nil, m.containerListErr
	}
	return nil, nil
}

func (m *mockAPIClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return types.ContainerJSON{}, nil
}

func (m *mockAPIClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	return nil, nil
}

func (m *mockAPIClient) VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error) {
	return volume.ListResponse{}, nil
}

func (m *mockAPIClient) Close() error { return nil }

func TestCollectWithClientContainerListError(t *testing.T) {
	wantErr := errors.New("permission denied")
	d, skips, err := collectWithClient(context.Background(), &mockAPIClient{containerListErr: wantErr}, Options{Detail: true})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !d.Available {
		t.Fatal("expected docker available when ping/info succeed")
	}
	if len(skips) != 1 || skips[0].Component != "docker.containers" {
		t.Fatalf("unexpected skips: %+v", skips)
	}
}
