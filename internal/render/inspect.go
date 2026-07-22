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
	kind      int
	identity  int
	container int
	resource  int
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
	if err := writeInspectDatabases(w, snap.Databases); err != nil {
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

func writeInspectDatabases(w io.Writer, dbs []snapshot.Database) error {
	if _, err := fmt.Fprintln(w, "Databases"); err != nil {
		return err
	}
	if len(dbs) == 0 {
		if _, err := fmt.Fprintf(w, "  %s\n", EmptyHint("db")); err != nil {
			return err
		}
		return nil
	}
	for i, db := range dbs {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := writeDBDetail(w, db); err != nil {
			return err
		}
	}
	return nil
}

func correlationLayout(width int) correlationColumns {
	switch {
	case width < 80:
		return correlationColumns{kind: 4, identity: 16, container: 10, resource: 14}
	case width < 120:
		return correlationColumns{kind: 4, identity: 22, container: 12, resource: 18}
	default:
		return correlationColumns{kind: 4, identity: 28, container: 14, resource: 24}
	}
}

func writeCorrelationTable(w io.Writer, rows []snapshot.Correlation, width int) error {
	if _, err := fmt.Fprintln(w, "Correlation"); err != nil {
		return err
	}
	if len(rows) == 0 {
		if _, err := fmt.Fprintf(w, "  %s\n", EmptyHint("correlation")); err != nil {
			return err
		}
		return nil
	}

	cols := correlationLayout(width)
	if _, err := fmt.Fprintf(w, "  %-*s %-*s %-*s %-*s %s\n",
		cols.kind, "KIND",
		cols.identity, "IDENTITY",
		cols.container, "CONTAINER",
		cols.resource, "GPU/ADDR",
		"HEALTH",
	); err != nil {
		return err
	}

	for _, row := range rows {
		kind := snapshot.CorrelationKindOf(row)
		container := row.ContainerName
		if container == "" {
			container = "-"
		}
		if _, err := fmt.Fprintf(w, "  %-*s %-*s %-*s %-*s %s\n",
			cols.kind, kind,
			cols.identity, truncate(correlationIdentity(row), cols.identity),
			cols.container, truncate(container, cols.container),
			cols.resource, truncate(correlationResource(row), cols.resource),
			correlationHealthLabel(row),
		); err != nil {
			return err
		}
	}
	return nil
}

func correlationIdentity(row snapshot.Correlation) string {
	if snapshot.CorrelationKindOf(row) == snapshot.CorrelationKindDB {
		return row.Runtime
	}
	if len(row.Models) > 0 {
		return row.Models[0]
	}
	return row.Runtime
}

func correlationResource(row snapshot.Correlation) string {
	if snapshot.CorrelationKindOf(row) == snapshot.CorrelationKindDB {
		return row.Endpoint
	}
	gpu := "-"
	if row.GPUIndex != nil {
		gpu = fmt.Sprintf("gpu %d", *row.GPUIndex)
		if row.GPUName != "" {
			gpu = truncate(row.GPUName, 12)
		}
	}
	vram := ""
	if row.VRAMBytes > 0 {
		vram = " · " + formatBytes(row.VRAMBytes)
	}
	return gpu + vram
}

func correlationHealthLabel(row snapshot.Correlation) string {
	if row.HealthOK == nil {
		return "?"
	}
	if *row.HealthOK {
		return "ok"
	}
	return "fail"
}

func writeInspectLLMs(w io.Writer, llms []snapshot.LLM) error {
	if _, err := fmt.Fprintln(w, "LLM runtimes"); err != nil {
		return err
	}
	if len(llms) == 0 {
		if _, err := fmt.Fprintf(w, "  %s\n", EmptyHint("llm")); err != nil {
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
		if _, err := fmt.Fprintf(w, "  %s\n", EmptyHint("docker")); err != nil {
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
