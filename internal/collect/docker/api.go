package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// apiClient abstracts the Docker Engine API for collection and testing.
type apiClient interface {
	Ping(ctx context.Context) error
	Info(ctx context.Context) (system.Info, error)
	ClientVersion() string
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
	VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error)
	Close() error
}

type engineClient struct {
	*client.Client
}

func newEngineClient() (apiClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &engineClient{Client: cli}, nil
}

func (c *engineClient) Ping(ctx context.Context) error {
	_, err := c.Client.Ping(ctx)
	return err
}
