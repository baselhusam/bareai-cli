package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// WriteGPU renders a detailed GPU report.
func WriteGPU(w io.Writer, snap *snapshot.Snapshot, noColor bool) error {
	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}

	label := func(s string) string {
		if noColor {
			return s
		}
		return s
	}

	if _, err := fmt.Fprintf(w, "%s\n", label("bareai gpu")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Collected: %s\n\n", snap.CollectedAt.Format(time.RFC3339)); err != nil {
		return err
	}

	if len(snap.GPUs) == 0 {
		if _, err := fmt.Fprintln(w, "No accelerators detected."); err != nil {
			return err
		}
	} else {
		for i, gpu := range snap.GPUs {
			if i > 0 {
				if _, err := fmt.Fprintln(w); err != nil {
					return err
				}
			}
			if err := writeGPUDetail(w, gpu); err != nil {
				return err
			}
		}
	}

	if len(snap.Skipped) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "Skipped:"); err != nil {
			return err
		}
		for _, skip := range snap.Skipped {
			if _, err := fmt.Fprintf(w, "  - %s: %s\n", skip.Component, skip.Reason); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeGPUDetail(w io.Writer, gpu snapshot.GPU) error {
	if _, err := fmt.Fprintf(w, "GPU %d (%s)\n", gpu.Index, gpu.Vendor); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Name:      %s\n", gpu.Name); err != nil {
		return err
	}
	if gpu.UUID != "" {
		if _, err := fmt.Fprintf(w, "  UUID:      %s\n", gpu.UUID); err != nil {
			return err
		}
	}
	if gpu.Driver != "" {
		if _, err := fmt.Fprintf(w, "  Driver:    %s\n", gpu.Driver); err != nil {
			return err
		}
	}

	if gpu.MemoryTotal > 0 || gpu.MemoryUsed > 0 {
		if _, err := fmt.Fprintf(w, "  Memory:    %s / %s\n",
			formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal)); err != nil {
			return err
		}
	} else if gpu.Vendor == "apple" {
		if _, err := fmt.Fprintln(w, "  Memory:    unified (no discrete VRAM reported)"); err != nil {
			return err
		}
	}

	if gpu.Utilization != nil {
		if _, err := fmt.Fprintf(w, "  Util:      %.0f%%\n", *gpu.Utilization); err != nil {
			return err
		}
	}
	if gpu.Temperature != nil {
		if _, err := fmt.Fprintf(w, "  Temp:      %.0f C\n", *gpu.Temperature); err != nil {
			return err
		}
	}
	if gpu.PowerDrawW != nil || gpu.PowerLimitW != nil {
		draw := formatOptionalFloat(gpu.PowerDrawW, "n/a")
		limit := formatOptionalFloat(gpu.PowerLimitW, "n/a")
		if _, err := fmt.Fprintf(w, "  Power:     %s W / %s W\n", draw, limit); err != nil {
			return err
		}
	}

	if len(gpu.Processes) == 0 {
		return nil
	}

	if _, err := fmt.Fprintln(w, "  Processes:"); err != nil {
		return err
	}
	for _, proc := range gpu.Processes {
		name := proc.Name
		if name == "" {
			name = "unknown"
		}
		if _, err := fmt.Fprintf(w, "    pid %d  %s  %s\n",
			proc.PID, name, formatBytes(proc.MemoryUsed)); err != nil {
			return err
		}
	}

	return nil
}

func writeGPUSummary(w io.Writer, gpus []snapshot.GPU) error {
	if len(gpus) == 0 {
		if _, err := fmt.Fprintln(w, "GPUs:        none detected"); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintln(w, "GPUs:"); err != nil {
		return err
	}
	for _, gpu := range gpus {
		util := "n/a"
		if gpu.Utilization != nil {
			util = fmt.Sprintf("%.0f%%", *gpu.Utilization)
		}
		temp := "n/a"
		if gpu.Temperature != nil {
			temp = fmt.Sprintf("%.0fC", *gpu.Temperature)
		}
		mem := "n/a"
		if gpu.MemoryTotal > 0 {
			mem = fmt.Sprintf("%s / %s", formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal))
		} else if gpu.Vendor == "apple" {
			mem = "unified"
		}

		if _, err := fmt.Fprintf(w, "  [%d] %s (%s)  util=%s  mem=%s  temp=%s\n",
			gpu.Index, gpu.Name, gpu.Vendor, util, mem, temp); err != nil {
			return err
		}
	}
	return nil
}

func formatOptionalFloat(v *float64, fallback string) string {
	if v == nil {
		return fallback
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", *v), "0"), ".")
}
