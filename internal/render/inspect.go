package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// InspectOptions controls inspect human output.
type InspectOptions struct {
	NoColor bool
	Width   int
}

type correlationColumns struct {
	endpoint  int
	runtime   int
	container int
	pid       int
	gpu       int
	vram      int
	models    int
	showModels bool
}

// WriteInspect renders a full correlated infrastructure report.
func WriteInspect(w io.Writer, snap *snapshot.Snapshot, opts InspectOptions) error {
	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}
	if opts.Width <= 0 {
		opts.Width = defaultTerminalWidth
	}

	label := func(s string) string {
		if opts.NoColor {
			return s
		}
		return s
	}

	if _, err := fmt.Fprintf(w, "%s\n", label("bareai inspect")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Collected: %s\n\n", snap.CollectedAt.Format(time.RFC3339)); err != nil {
		return err
	}

	if err := writeInspectOverview(w, snap); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := writeCorrelationTable(w, snap.Correlations, opts.Width); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := writeGPUSummary(w, snap.GPUs); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := writeInspectLLMs(w, snap.LLMs); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := writeInspectDocker(w, snap.Docker); err != nil {
		return err
	}
	if len(snap.Findings) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if err := writeFindings(w, snap.Findings); err != nil {
			return err
		}
	}

	return writeSkipped(w, snap.Skipped)
}

func writeInspectOverview(w io.Writer, snap *snapshot.Snapshot) error {
	if _, err := fmt.Fprintln(w, "Overview"); err != nil {
		return err
	}

	host := "unavailable"
	if snap.Host != nil && snap.Host.Hostname != "" {
		host = snap.Host.Hostname
	}

	gpuCount := len(snap.GPUs)
	docker := "not available"
	if snap.Docker != nil && snap.Docker.Available {
		running := 0
		for _, c := range snap.Docker.Containers {
			if strings.EqualFold(c.State, "running") {
				running++
			}
		}
		docker = fmt.Sprintf("%d running", running)
	}

	llmCount := len(snap.LLMs)
	dbCount := len(snap.Databases)
	if _, err := fmt.Fprintf(w, "  Host: %s   GPUs: %d   Docker: %s   LLMs: %d   DBs: %d\n",
		host, gpuCount, docker, llmCount, dbCount); err != nil {
		return err
	}
	return nil
}

func correlationLayout(width int) correlationColumns {
	switch {
	case width < 80:
		return correlationColumns{
			endpoint: 18, runtime: 8, container: 10, pid: 5, gpu: 3, vram: 8, showModels: false,
		}
	case width < 120:
		return correlationColumns{
			endpoint: 24, runtime: 8, container: 12, pid: 5, gpu: 3, vram: 8, models: width - 65, showModels: true,
		}
	default:
		return correlationColumns{
			endpoint: 36, runtime: 10, container: 14, pid: 5, gpu: 3, vram: 10, models: width - 83, showModels: true,
		}
	}
}

func writeCorrelationTable(w io.Writer, rows []snapshot.Correlation, width int) error {
	if _, err := fmt.Fprintln(w, "Correlation"); err != nil {
		return err
	}
	if len(rows) == 0 {
		if _, err := fmt.Fprintln(w, "  none"); err != nil {
			return err
		}
		return nil
	}

	cols := correlationLayout(width)
	if cols.showModels {
		if _, err := fmt.Fprintf(w, "  %-*s %-*s %-*s %-*s %-*s %-*s %s\n",
			cols.endpoint, "ENDPOINT",
			cols.runtime, "RUNTIME",
			cols.container, "CONTAINER",
			cols.pid, "PID",
			cols.gpu, "GPU",
			cols.vram, "VRAM",
			"MODELS",
		); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "  %-*s %-*s %-*s %-*s %-*s %s\n",
			cols.endpoint, "ENDPOINT",
			cols.runtime, "RUNTIME",
			cols.container, "CONTAINER",
			cols.pid, "PID",
			cols.gpu, "GPU",
			"VRAM",
		); err != nil {
			return err
		}
	}

	for _, row := range rows {
		pid := "-"
		if row.PID > 0 {
			pid = fmt.Sprintf("%d", row.PID)
		}
		gpu := "-"
		if row.GPUIndex != nil {
			gpu = fmt.Sprintf("%d", *row.GPUIndex)
		}
		vram := "-"
		if row.VRAMBytes > 0 {
			vram = formatBytes(row.VRAMBytes)
		}
		container := row.ContainerName
		if container == "" {
			container = "-"
		}
		models := strings.Join(row.Models, ", ")
		if models == "" {
			models = "-"
		}

		if cols.showModels {
			if _, err := fmt.Fprintf(w, "  %-*s %-*s %-*s %-*s %-*s %-*s %s\n",
				cols.endpoint, truncate(row.Endpoint, cols.endpoint),
				cols.runtime, truncate(row.Runtime, cols.runtime),
				cols.container, truncate(container, cols.container),
				cols.pid, pid,
				cols.gpu, gpu,
				cols.vram, vram,
				truncate(models, cols.models),
			); err != nil {
				return err
			}
			continue
		}

		if _, err := fmt.Fprintf(w, "  %-*s %-*s %-*s %-*s %-*s %s\n",
			cols.endpoint, truncate(row.Endpoint, cols.endpoint),
			cols.runtime, truncate(row.Runtime, cols.runtime),
			cols.container, truncate(container, cols.container),
			cols.pid, pid,
			cols.gpu, gpu,
			vram,
		); err != nil {
			return err
		}
	}
	return nil
}

func writeInspectLLMs(w io.Writer, llms []snapshot.LLM) error {
	if _, err := fmt.Fprintln(w, "LLM runtimes"); err != nil {
		return err
	}
	if len(llms) == 0 {
		if _, err := fmt.Fprintln(w, "  none"); err != nil {
			return err
		}
		return nil
	}
	for i, llm := range llms {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := writeLLMDetail(w, llm); err != nil {
			return err
		}
	}
	return nil
}

func writeInspectDocker(w io.Writer, docker *snapshot.Docker) error {
	if _, err := fmt.Fprintln(w, "Docker"); err != nil {
		return err
	}
	if docker == nil || !docker.Available {
		if _, err := fmt.Fprintln(w, "  not available"); err != nil {
			return err
		}
		return nil
	}

	running := 0
	gpuBacked := 0
	for _, c := range docker.Containers {
		if !strings.EqualFold(c.State, "running") {
			continue
		}
		running++
		if c.GPURequested {
			gpuBacked++
		}
	}

	line := fmt.Sprintf("  %d running containers", running)
	if docker.NVIDIARuntime {
		line += ", nvidia runtime"
	}
	if gpuBacked > 0 {
		line += fmt.Sprintf(", %d GPU-backed", gpuBacked)
	}
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %d images, %d volumes\n", len(docker.Images), len(docker.Volumes)); err != nil {
		return err
	}
	return nil
}

func writeFindings(w io.Writer, findings []snapshot.Finding) error {
	if _, err := fmt.Fprintln(w, "Findings"); err != nil {
		return err
	}
	for _, f := range findings {
		severity := f.Severity
		if severity == "" {
			severity = "info"
		}
		if _, err := fmt.Fprintf(w, "  [%s] %s: %s\n", severity, f.ID, f.Summary); err != nil {
			return err
		}
	}
	return nil
}
