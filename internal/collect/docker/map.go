package docker

import (
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func mapInfo(info system.Info, apiVersion string) snapshot.Docker {
	runtimes := runtimeNames(info.Runtimes)
	return snapshot.Docker{
		Available:      true,
		ServerVersion:  info.ServerVersion,
		APIVersion:     apiVersion,
		OSType:         info.OSType,
		Architecture:   info.Architecture,
		DefaultRuntime: info.DefaultRuntime,
		Runtimes:       runtimes,
		NVIDIARuntime:  hasNVIDIARuntime(info.Runtimes, info.DefaultRuntime),
	}
}

func runtimeNames(runtimes map[string]system.RuntimeWithStatus) []string {
	if len(runtimes) == 0 {
		return nil
	}
	names := make([]string, 0, len(runtimes))
	for name := range runtimes {
		names = append(names, name)
	}
	return names
}

func hasNVIDIARuntime(runtimes map[string]system.RuntimeWithStatus, defaultRuntime string) bool {
	if strings.EqualFold(defaultRuntime, "nvidia") {
		return true
	}
	for name := range runtimes {
		if strings.EqualFold(name, "nvidia") {
			return true
		}
	}
	return false
}

func mapContainer(summary types.Container, inspect types.ContainerJSON) snapshot.DockerContainer {
	name := containerName(summary.Names)
	deviceRequests := deviceRequestsFromInspect(inspect)
	return snapshot.DockerContainer{
		ID:             shortID(summary.ID),
		Name:           name,
		Image:          summary.Image,
		State:          summary.State,
		Status:         summary.Status,
		Created:        time.Unix(summary.Created, 0).UTC(),
		Ports:          mapPorts(summary.Ports),
		Labels:         summary.Labels,
		GPURequested:   gpuRequested(deviceRequests),
		DeviceRequests: deviceRequests,
	}
}

func containerName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func mapPorts(ports []types.Port) []snapshot.DockerPort {
	if len(ports) == 0 {
		return nil
	}
	out := make([]snapshot.DockerPort, 0, len(ports))
	for _, p := range ports {
		out = append(out, snapshot.DockerPort{
			PrivatePort: p.PrivatePort,
			PublicPort:  p.PublicPort,
			Type:        p.Type,
			IP:          p.IP,
		})
	}
	return out
}

func deviceRequestsFromInspect(inspect types.ContainerJSON) []snapshot.DeviceRequest {
	if inspect.ContainerJSONBase == nil || inspect.HostConfig == nil {
		return nil
	}
	return mapDeviceRequests(inspect.HostConfig.DeviceRequests)
}

func mapDeviceRequests(reqs []container.DeviceRequest) []snapshot.DeviceRequest {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]snapshot.DeviceRequest, 0, len(reqs))
	for _, req := range reqs {
		out = append(out, snapshot.DeviceRequest{
			Driver:       req.Driver,
			Count:        req.Count,
			DeviceIDs:    append([]string(nil), req.DeviceIDs...),
			Capabilities: flattenCapabilities(req.Capabilities),
		})
	}
	return out
}

func flattenCapabilities(caps [][]string) []string {
	if len(caps) == 0 {
		return nil
	}
	var out []string
	for _, group := range caps {
		out = append(out, group...)
	}
	return out
}

func gpuRequested(reqs []snapshot.DeviceRequest) bool {
	for _, req := range reqs {
		if strings.EqualFold(req.Driver, "nvidia") {
			return true
		}
		for _, cap := range req.Capabilities {
			lower := strings.ToLower(cap)
			if lower == "gpu" || lower == "nvidia" {
				return true
			}
		}
	}
	return false
}

func mapImage(img image.Summary) snapshot.DockerImage {
	return snapshot.DockerImage{
		ID:       shortID(img.ID),
		RepoTags: append([]string(nil), img.RepoTags...),
		Size:     uint64(img.Size),
		Created:  time.Unix(img.Created, 0).UTC(),
	}
}

func mapVolume(vol volume.Volume) snapshot.DockerVolume {
	return snapshot.DockerVolume{
		Name:       vol.Name,
		Driver:     vol.Driver,
		Mountpoint: vol.Mountpoint,
		Scope:      vol.Scope,
	}
}
