package tui

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func llmListTitle(llm snapshot.LLM, s styles) string {
	health := s.muted.Render("?")
	if llm.Health != nil {
		label := "fail"
		if llm.Health.OK {
			label = "ok"
		}
		health = s.healthStyle(llm.Health.OK).Render(label)
	}
	name := llm.Name
	if name == "" {
		name = llm.Runtime
	}
	pid := ""
	if llm.PID > 0 {
		pid = fmt.Sprintf(" pid=%d", llm.PID)
	}
	return fmt.Sprintf("%s  %s%s  [%s]",
		truncate(name, 14), truncate(llm.Endpoint, 24), pid, health)
}

func llmFilterValue(llm snapshot.LLM) string {
	parts := []string{
		llm.Runtime,
		llm.Name,
		llm.Endpoint,
		llm.Source,
		llm.ContainerName,
		llm.ContainerID,
		fmt.Sprintf("%d", llm.PID),
	}
	for _, m := range llm.Models {
		parts = append(parts, m.ID, m.Name)
	}
	return strings.Join(parts, " ")
}

func llmSourceLabel(llm snapshot.LLM) string {
	switch llm.Source {
	case "docker":
		if llm.ContainerName != "" {
			return "docker: " + llm.ContainerName
		}
		return "docker"
	case "process":
		if llm.PID > 0 {
			return fmt.Sprintf("process pid %d", llm.PID)
		}
		return "process"
	case "port":
		return "port scan"
	default:
		return llm.Source
	}
}

func llmDetailText(llm snapshot.LLM, s styles) string {
	var b strings.Builder
	source := llmSourceLabel(llm)
	fmt.Fprintf(&b, "%s  %s\n", s.value.Render(llm.Runtime+" / "+llm.Name), llm.Endpoint)
	fmt.Fprintf(&b, "  Provider:  %s  (%s)\n", llm.Runtime, source)
	if llm.PID > 0 {
		fmt.Fprintf(&b, "  PID:       %d\n", llm.PID)
	}
	if llm.Health != nil {
		status := s.fail.Render("fail")
		if llm.Health.OK {
			status = s.ok.Render("ok")
		}
		line := fmt.Sprintf("  Health:    %s  %dms", status, llm.Health.LatencyMS)
		if llm.Health.Message != "" {
			line += "  " + llm.Health.Message
		}
		if !llm.Health.OK && llm.Health.Error != "" {
			line += "  " + llm.Health.Error
		}
		fmt.Fprintln(&b, line)
	}
	if len(llm.Models) > 0 {
		names := make([]string, 0, len(llm.Models))
		for _, m := range llm.Models {
			if m.ID != "" {
				names = append(names, m.ID)
			}
		}
		fmt.Fprintf(&b, "  Models:    %s\n", strings.Join(names, ", "))
	}
	if llm.GPUIndex != nil {
		fmt.Fprintf(&b, "  GPU:       %d\n", *llm.GPUIndex)
	}
	if llm.ContainerName != "" {
		fmt.Fprintf(&b, "  Container: %s\n", llm.ContainerName)
	}
	if len(llm.Metrics) > 0 {
		if line := probe.MetricsLine(llm.Runtime, llm.Metrics); line != "" {
			fmt.Fprintf(&b, "  Metrics:   %s\n", line)
		}
	}
	if llm.Probe != nil {
		status := s.fail.Render("fail")
		if llm.Probe.OK {
			status = s.ok.Render("pass")
		}
		fmt.Fprintf(&b, "  Last probe: %s  %dms", status, llm.Probe.LatencyMS)
		if llm.Probe.Message != "" {
			fmt.Fprintf(&b, "  %s", llm.Probe.Message)
		}
		if !llm.Probe.OK && llm.Probe.Error != "" {
			fmt.Fprintf(&b, "  %s", llm.Probe.Error)
		}
		fmt.Fprintln(&b)
	}
	return strings.TrimRight(b.String(), "\n")
}
