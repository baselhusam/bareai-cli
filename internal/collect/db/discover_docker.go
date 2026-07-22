package db

import (
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func discoverDocker(in Input) []candidate {
	if in.Docker == nil || !in.Docker.Available {
		return nil
	}
	var out []candidate
	for _, c := range in.Docker.Containers {
		if !strings.EqualFold(c.State, "running") {
			continue
		}
		hint := matchEngineHint(c.Image, c.Name)
		if hint == nil {
			continue
		}
		port := pickPort(c.Ports, hint.ports)
		if port == 0 {
			continue
		}
		out = append(out, candidate{
			priority: 3,
			Database: snapshot.Database{
				Engine:        hint.engine,
				Name:          displayName(hint.engine),
				Address:       baseAddr(port),
				Source:        sourceDocker,
				PID:           c.PID,
				ContainerID:   c.ID,
				ContainerName: c.Name,
			},
		})
	}
	return out
}

func pickPort(ports []snapshot.DockerPort, defaults []uint16) uint16 {
	defaultSet := make(map[uint16]bool, len(defaults))
	for _, p := range defaults {
		defaultSet[p] = true
	}
	for _, p := range ports {
		if p.PublicPort > 0 && (len(defaultSet) == 0 || defaultSet[p.PublicPort]) {
			return p.PublicPort
		}
	}
	for _, p := range ports {
		if p.PublicPort > 0 {
			return p.PublicPort
		}
	}
	for _, p := range ports {
		if p.PrivatePort > 0 && (len(defaultSet) == 0 || defaultSet[p.PrivatePort]) {
			return p.PrivatePort
		}
	}
	for _, p := range ports {
		if p.PrivatePort > 0 {
			return p.PrivatePort
		}
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return 0
}
