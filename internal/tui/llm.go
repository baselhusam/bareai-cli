package tui

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func llmListTitle(llm snapshot.LLM) string {
	health := "?"
	if llm.Health != nil {
		if llm.Health.OK {
			health = "ok"
		} else {
			health = "fail"
		}
	}
	name := llm.Name
	if name == "" {
		name = llm.Runtime
	}
	return fmt.Sprintf("%s  %s  [%s]", truncate(name, 16), truncate(llm.Endpoint, 28), health)
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

func llmDetailText(llm snapshot.LLM) string {
	var b strings.Builder
	source := llmSourceLabel(llm)
	fmt.Fprintf(&b, "%s  %s  (%s)\n", llm.Name, llm.Endpoint, source)
	if llm.Health != nil {
		status := "fail"
		if llm.Health.OK {
			status = "ok"
		}
		line := fmt.Sprintf("  Health: %s  %dms", status, llm.Health.LatencyMS)
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
		fmt.Fprintf(&b, "  Models: %s\n", strings.Join(names, ", "))
	}
	if llm.GPUIndex != nil {
		fmt.Fprintf(&b, "  GPU: %d\n", *llm.GPUIndex)
	}
	if llm.ContainerName != "" {
		fmt.Fprintf(&b, "  Container: %s\n", llm.ContainerName)
	}
	if len(llm.Metrics) > 0 {
		parts := make([]string, 0, len(llm.Metrics))
		for k, v := range llm.Metrics {
			parts = append(parts, fmt.Sprintf("%s=%.0f", k, v))
		}
		fmt.Fprintf(&b, "  Metrics: %s\n", strings.Join(parts, ", "))
	}
	if llm.Probe != nil {
		status := "fail"
		if llm.Probe.OK {
			status = "pass"
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
