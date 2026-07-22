package action

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// DockerAPI is the scoped Docker Engine surface for actions.
type DockerAPI interface {
	Ping(ctx context.Context) error
	ContainerInspect(ctx context.Context, id string) (containerState, error)
	ContainerRestart(ctx context.Context, id string, timeout time.Duration) error
	ContainerStop(ctx context.Context, id string, timeout time.Duration) error
	ContainerLogs(ctx context.Context, id string, tail int) (string, error)
	Close() error
}

type containerState struct {
	ID     string
	Name   string
	Image  string
	State  string
	Status string
}

type engineDocker struct {
	*client.Client
}

// NewDockerAPI connects to the local Docker Engine.
func NewDockerAPI() (DockerAPI, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &engineDocker{Client: cli}, nil
}

func (e *engineDocker) Ping(ctx context.Context) error {
	_, err := e.Client.Ping(ctx)
	return err
}

func (e *engineDocker) ContainerInspect(ctx context.Context, id string) (containerState, error) {
	inspect, err := e.Client.ContainerInspect(ctx, id)
	if err != nil {
		return containerState{}, err
	}
	name := inspect.Name
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	image := inspect.Config.Image
	if inspect.Image != "" {
		image = inspect.Image
	}
	return containerState{
		ID:     inspect.ID,
		Name:   name,
		Image:  image,
		State:  inspect.State.Status,
		Status: inspect.State.Status,
	}, nil
}

func (e *engineDocker) ContainerRestart(ctx context.Context, id string, timeout time.Duration) error {
	seconds := int(timeout.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return e.Client.ContainerRestart(ctx, id, container.StopOptions{Timeout: &seconds})
}

func (e *engineDocker) ContainerStop(ctx context.Context, id string, timeout time.Duration) error {
	seconds := int(timeout.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return e.Client.ContainerStop(ctx, id, container.StopOptions{Timeout: &seconds})
}

func (e *engineDocker) ContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	if tail <= 0 {
		tail = 100
	}
	rc, err := e.Client.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tail),
		Timestamps: true,
	})
	if err != nil {
		return "", err
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (e *engineDocker) Close() error {
	return e.Client.Close()
}

func stateFromContainer(s containerState) *ContainerState {
	if s.ID == "" && s.Name == "" {
		return nil
	}
	return &ContainerState{
		ID:     s.ID,
		Name:   s.Name,
		Image:  s.Image,
		State:  s.State,
		Status: s.Status,
	}
}

func validateContainerTarget(snapTarget *Target) error {
	if snapTarget == nil || snapTarget.Kind != TargetContainer {
		return fmt.Errorf("action requires a docker container target")
	}
	if snapTarget.ID == "" {
		return fmt.Errorf("resolved container has no ID")
	}
	return nil
}

func readContainerState(ctx context.Context, api DockerAPI, id string) (*ContainerState, error) {
	if api == nil {
		return nil, fmt.Errorf("docker client is nil")
	}
	s, err := api.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}
	return stateFromContainer(s), nil
}

func truncateOutput(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "\n… (truncated)"
}
