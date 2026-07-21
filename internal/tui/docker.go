package tui

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func dockerListTitle(c snapshot.DockerContainer, s styles) string {
	gpu := s.muted.Render("no-gpu")
	if c.GPURequested {
		gpu = s.ok.Render("gpu")
	}
	pid := ""
	if c.PID > 0 {
		pid = fmt.Sprintf(" pid=%d", c.PID)
	}
	stateStyle := s.value
	switch strings.ToLower(c.State) {
	case "running":
		stateStyle = s.ok
	case "exited", "dead":
		stateStyle = s.fail
	case "paused":
		stateStyle = s.warn
	}
	return fmt.Sprintf("%-16s  %s%s  %s",
		truncate(c.Name, 16),
		stateStyle.Render(truncate(c.State, 10)),
		pid, gpu)
}

func dockerFilterValue(c snapshot.DockerContainer) string {
	parts := []string{
		c.ID,
		c.Name,
		c.Image,
		c.State,
		c.Status,
		fmt.Sprintf("%d", c.PID),
	}
	for _, p := range c.Ports {
		parts = append(parts, fmt.Sprintf("%d", p.PublicPort), fmt.Sprintf("%d", p.PrivatePort))
	}
	return strings.Join(parts, " ")
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

func dockerDetailText(c snapshot.DockerContainer, s styles) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Container: %s\n", c.Name)
	fmt.Fprintf(&b, "  ID:      %s\n", c.ID)
	fmt.Fprintf(&b, "  Image:   %s\n", c.Image)
	state := s.value.Render(c.State)
	switch strings.ToLower(c.State) {
	case "running":
		state = s.ok.Render(c.State)
	case "exited", "dead":
		state = s.fail.Render(c.State)
	case "paused":
		state = s.warn.Render(c.State)
	}
	fmt.Fprintf(&b, "  State:   %s\n", state)
	fmt.Fprintf(&b, "  Status:  %s\n", c.Status)
	if c.PID > 0 {
		fmt.Fprintf(&b, "  PID:     %d\n", c.PID)
	}
	fmt.Fprintf(&b, "  Ports:   %s\n", formatDockerPorts(c.Ports))
	gpu := s.muted.Render("no")
	if c.GPURequested {
		gpu = s.ok.Render("yes")
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
	return dockerContainersForList(containers, false)
}

func dockerContainersForList(containers []snapshot.DockerContainer, showAll bool) []snapshot.DockerContainer {
	if showAll {
		out := make([]snapshot.DockerContainer, len(containers))
		copy(out, containers)
		return out
	}
	out := make([]snapshot.DockerContainer, 0, len(containers))
	for _, c := range containers {
		if strings.EqualFold(c.State, "running") {
			out = append(out, c)
		}
	}
	return out
}
