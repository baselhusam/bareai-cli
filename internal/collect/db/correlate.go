package db

import "github.com/baselhusam/bareai-cli/internal/snapshot"

func correlate(db *snapshot.Database, in Input) {
	if db == nil {
		return
	}
	if db.Source != sourceDocker {
		matchContainerByPortOrPID(db, in.Docker)
	}
}

func matchContainerByPortOrPID(db *snapshot.Database, docker *snapshot.Docker) {
	if docker == nil || !docker.Available {
		return
	}
	port, ok := addressPort(db.Address)
	if !ok {
		return
	}
	for _, c := range docker.Containers {
		if db.PID > 0 && c.PID == db.PID {
			db.ContainerID = c.ID
			db.ContainerName = c.Name
			return
		}
		for _, p := range c.Ports {
			if p.PublicPort == port || p.PrivatePort == port {
				db.ContainerID = c.ID
				db.ContainerName = c.Name
				if db.PID == 0 {
					db.PID = c.PID
				}
				return
			}
		}
	}
}
