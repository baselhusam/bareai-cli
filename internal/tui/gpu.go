package tui

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func gpuListTitle(gpu snapshot.GPU) string {
	util := "n/a"
	if gpu.Utilization != nil {
		util = fmt.Sprintf("%.0f%%", *gpu.Utilization)
	}
	mem := "n/a"
	if gpu.MemoryTotal > 0 {
		mem = fmt.Sprintf("%s/%s", formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal))
	} else if gpu.Vendor == "apple" {
		mem = "unified"
	}
	return fmt.Sprintf("[%d] %s  util=%s  mem=%s", gpu.Index, truncate(gpu.Name, 24), util, mem)
}

func gpuDetailText(gpu snapshot.GPU) string {
	var b strings.Builder
	fmt.Fprintf(&b, "GPU %d (%s)\n", gpu.Index, gpu.Vendor)
	fmt.Fprintf(&b, "  Name:      %s\n", gpu.Name)
	if gpu.UUID != "" {
		fmt.Fprintf(&b, "  UUID:      %s\n", gpu.UUID)
	}
	if gpu.Driver != "" {
		fmt.Fprintf(&b, "  Driver:    %s\n", gpu.Driver)
	}
	if gpu.MemoryTotal > 0 || gpu.MemoryUsed > 0 {
		fmt.Fprintf(&b, "  Memory:    %s / %s\n", formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal))
	} else if gpu.Vendor == "apple" {
		fmt.Fprintln(&b, "  Memory:    unified (no discrete VRAM reported)")
	}
	if gpu.Utilization != nil {
		fmt.Fprintf(&b, "  Util:      %.0f%%\n", *gpu.Utilization)
	}
	if gpu.Temperature != nil {
		fmt.Fprintf(&b, "  Temp:      %.0f C\n", *gpu.Temperature)
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
			fmt.Fprintf(&b, "    pid %d  %s  %s\n", proc.PID, name, formatBytes(proc.MemoryUsed))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
