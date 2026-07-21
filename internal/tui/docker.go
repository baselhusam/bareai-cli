package tui

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func dockerListTitle(c snapshot.DockerContainer) string {
	gpu := "no"
	if c.GPURequested {
		gpu = "yes"
	}
	return fmt.Sprintf("%-16s  %-10s  gpu=%s", truncate(c.Name, 16), truncate(c.State, 10), gpu)
}

func formatDockerPorts(ports []snapshot.DockerPort) string {
	if len(ports) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		if p.PublicPort > 0 {
			ip := p.IP
			if ip == "" || ip == "0.0.0.0" || ip == "::" {
				parts = append(parts, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else {
				parts = append(parts, fmt.Sprintf("%s:%d->%d/%s", ip, p.PublicPort, p.PrivatePort, p.Type))
			}
			continue
		}
		parts = append(parts, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
	}
	return strings.Join(parts, ", ")
}

func dockerDetailText(c snapshot.DockerContainer) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Container: %s\n", c.Name)
	fmt.Fprintf(&b, "  ID:      %s\n", c.ID)
	fmt.Fprintf(&b, "  Image:   %s\n", c.Image)
	fmt.Fprintf(&b, "  State:   %s\n", c.State)
	fmt.Fprintf(&b, "  Status:  %s\n", c.Status)
	if c.PID > 0 {
		fmt.Fprintf(&b, "  PID:     %d\n", c.PID)
	}
	fmt.Fprintf(&b, "  Ports:   %s\n", formatDockerPorts(c.Ports))
	gpu := "no"
	if c.GPURequested {
		gpu = "yes"
	}
	fmt.Fprintf(&b, "  GPU:     %s\n", gpu)
	if len(c.DeviceRequests) > 0 {
		fmt.Fprintln(&b, "  Device requests:")
		for _, dr := range c.DeviceRequests {
			fmt.Fprintf(&b, "    driver=%s count=%d ids=%s\n",
				dr.Driver, dr.Count, strings.Join(dr.DeviceIDs, ","))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func runningContainers(containers []snapshot.DockerContainer) []snapshot.DockerContainer {
	out := make([]snapshot.DockerContainer, 0, len(containers))
	for _, c := range containers {
		if strings.EqualFold(c.State, "running") {
			out = append(out, c)
		}
	}
	return out
}
