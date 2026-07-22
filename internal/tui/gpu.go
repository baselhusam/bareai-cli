package tui

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func gpuListTitle(gpu snapshot.GPU, s styles, barW int) string {
	util := 0.0
	utilLabel := "n/a"
	if gpu.Utilization != nil {
		util = *gpu.Utilization
		utilLabel = fmt.Sprintf("%.0f%%", util)
	}
	memPct := pctUsed(gpu.MemoryUsed, gpu.MemoryTotal)
	memLabel := "unified"
	if gpu.MemoryTotal > 0 {
		memLabel = fmt.Sprintf("%s/%s", formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal))
	} else if gpu.Vendor == "apple" {
		memLabel = "unified"
	}
	bar := renderBar(s, util, barW)
	memBar := ""
	if gpu.MemoryTotal > 0 {
		memBar = " " + renderBar(s, memPct, barW)
	}
	return fmt.Sprintf("[%d] %s  %s %s  %s%s",
		gpu.Index, truncate(gpu.Name, 20), utilLabel, bar, memLabel, memBar)
}

func gpuFilterValue(gpu snapshot.GPU) string {
	parts := []string{
		fmt.Sprintf("%d", gpu.Index),
		gpu.Vendor,
		gpu.Name,
		gpu.UUID,
		gpu.Driver,
	}
	for _, p := range gpu.Processes {
		parts = append(parts, fmt.Sprintf("%d", p.PID), p.Name)
	}
	return strings.Join(parts, " ")
}

func gpuDetailText(gpu snapshot.GPU, s styles) string {
	var b strings.Builder
	fmt.Fprintf(&b, "GPU %d (%s)\n", gpu.Index, gpu.Vendor)
	fmt.Fprintf(&b, "  Name:      %s\n", gpu.Name)
	if gpu.UUID != "" {
		fmt.Fprintf(&b, "  UUID:      %s\n", gpu.UUID)
	}
	if gpu.Driver != "" {
		fmt.Fprintf(&b, "  Driver:    %s\n", gpu.Driver)
	}
	if gpu.Notes != "" {
		fmt.Fprintf(&b, "  Notes:     %s\n", gpu.Notes)
	}
	if gpu.MemoryTotal > 0 || gpu.MemoryUsed > 0 {
		memPct := pctUsed(gpu.MemoryUsed, gpu.MemoryTotal)
		fmt.Fprintf(&b, "  Memory:    %s / %s  %s\n",
			formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal),
			renderBar(s, memPct, defaultBarWidth))
	} else if gpu.Vendor == "apple" {
		fmt.Fprintln(&b, "  Memory:    unified (no discrete VRAM reported)")
	}
	if gpu.Utilization != nil {
		fmt.Fprintf(&b, "  Util:      %.0f%%  %s\n", *gpu.Utilization,
			renderBar(s, *gpu.Utilization, defaultBarWidth))
	}
	if gpu.Temperature != nil {
		fmt.Fprintf(&b, "  Temp:      %s\n",
			tempStyle(s, *gpu.Temperature).Render(fmt.Sprintf("%.0f C", *gpu.Temperature)))
	}
	if gpu.PowerDrawW != nil || gpu.PowerLimitW != nil {
		draw := formatOptionalFloat(gpu.PowerDrawW, "n/a")
		limit := formatOptionalFloat(gpu.PowerLimitW, "n/a")
		fmt.Fprintf(&b, "  Power:     %s W / %s W\n", draw, limit)
	}
	if len(gpu.Processes) > 0 {
		fmt.Fprintln(&b, "  Processes:")
		for _, proc := range gpu.Processes {
			name := proc.Name
			if name == "" {
				name = "unknown"
			}
			fmt.Fprintf(&b, "    %s  pid %d  %s\n",
				s.label.Render("●"), proc.PID, name+"  "+formatBytes(proc.MemoryUsed))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
