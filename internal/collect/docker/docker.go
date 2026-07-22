package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Options controls Docker collection depth.
type Options struct {
	Detail bool // when false, skip images and volumes
}

// Collect gathers read-only Docker Engine inventory.
func Collect(ctx context.Context, opts Options) (snapshot.Docker, []snapshot.Skip, error) {
	cli, err := newEngineClient()
	if err != nil {
		return unavailable(err), []snapshot.Skip{{
			Component: "docker",
			Reason:    err.Error(),
		}}, nil
	}
	defer cli.Close()

	return collectWithClient(ctx, cli, opts)
}

func collectWithClient(ctx context.Context, cli apiClient, opts Options) (snapshot.Docker, []snapshot.Skip, error) {
	if err := cli.Ping(ctx); err != nil {
		return unavailable(err), []snapshot.Skip{{
			Component: "docker",
			Reason:    daemonUnavailableReason(err),
		}}, nil
	}

	info, err := cli.Info(ctx)
	if err != nil {
		return unavailable(err), []snapshot.Skip{{
			Component: "docker",
			Reason:    err.Error(),
		}}, nil
	}

	d := mapInfo(info, cli.ClientVersion())
	var skips []snapshot.Skip

	containers, containerSkips, _ := collectContainers(ctx, cli)
	d.Containers = containers
	skips = append(skips, containerSkips...)

	if !opts.Detail {
		return d, skips, nil
	}

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		skips = append(skips, snapshot.Skip{
			Component: "docker.images",
			Reason:    err.Error(),
		})
	} else {
		d.Images = make([]snapshot.DockerImage, 0, len(images))
		for _, img := range images {
			d.Images = append(d.Images, mapImage(img))
		}
	}

	volumes, err := cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		skips = append(skips, snapshot.Skip{
			Component: "docker.volumes",
			Reason:    err.Error(),
		})
	} else {
		d.Volumes = make([]snapshot.DockerVolume, 0, len(volumes.Volumes))
		for _, vol := range volumes.Volumes {
			if vol == nil {
				continue
			}
			d.Volumes = append(d.Volumes, mapVolume(*vol))
		}
	}

	return d, skips, nil
}

func collectContainers(ctx context.Context, cli apiClient) ([]snapshot.DockerContainer, []snapshot.Skip, error) {
	list, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, []snapshot.Skip{{
			Component: "docker.containers",
			Reason:    err.Error(),
		}}, nil
	}

	containers := make([]snapshot.DockerContainer, 0, len(list))
	var skips []snapshot.Skip

	for _, summary := range list {
		inspect, err := cli.ContainerInspect(ctx, summary.ID)
		if err != nil {
			skips = append(skips, snapshot.Skip{
				Component: "docker.container." + shortID(summary.ID),
				Reason:    err.Error(),
			})
			containers = append(containers, mapContainer(summary, types.ContainerJSON{}))
			continue
		}
		containers = append(containers, mapContainer(summary, inspect))
	}

	return containers, skips, nil
}

func unavailable(err error) snapshot.Docker {
	return snapshot.Docker{Available: false}
}

func daemonUnavailableReason(err error) string {
	if err == nil {
		return "docker daemon not available"
	}
	return err.Error()
}
