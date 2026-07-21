package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// DoctorOptions configures doctor human output.
type DoctorOptions struct {
	NoColor bool
	Width   int
}

type doctorReport struct {
	CollectedAt time.Time        `json:"collected_at"`
	Findings    []snapshot.Finding `json:"findings"`
	Counts      map[string]int   `json:"counts"`
}

// WriteDoctor renders ranked doctor findings.
func WriteDoctor(w io.Writer, snap *snapshot.Snapshot, opts DoctorOptions) error {
	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}
	width := opts.Width
	if width <= 0 {
		width = 80
	}

	counts := countSeverities(snap.Findings)
	if _, err := fmt.Fprintln(w, "bareai doctor"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Collected: %s\n", snap.CollectedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s\n\n", formatFindingCounts(len(snap.Findings), counts)); err != nil {
		return err
	}

	if len(snap.Findings) == 0 {
		if _, err := fmt.Fprintln(w, "No findings."); err != nil {
			return err
		}
		return writeSkipped(w, snap.Skipped)
	}

	for _, f := range snap.Findings {
		if err := writeDoctorFinding(w, f, width); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return writeSkipped(w, snap.Skipped)
}

// WriteDoctorJSON writes doctor output as JSON.
func WriteDoctorJSON(w io.Writer, snap *snapshot.Snapshot) error {
	report := doctorReport{
		CollectedAt: snap.CollectedAt,
		Findings:    snap.Findings,
		Counts:      countSeverities(snap.Findings),
	}
	return WriteJSON(w, report)
}

func writeDoctorFinding(w io.Writer, f snapshot.Finding, width int) error {
	severity := f.Severity
	if severity == "" {
		severity = "info"
	}
	if _, err := fmt.Fprintf(w, "[%s] %s — %s\n", severity, f.ID, f.Summary); err != nil {
		return err
	}
	if f.Why != "" {
		for _, line := range wrapText("Why: "+f.Why, width, "  ") {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	if f.Try != "" {
		for _, line := range wrapText("Try: "+f.Try, width, "  ") {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	return nil
}

func countSeverities(findings []snapshot.Finding) map[string]int {
	counts := map[string]int{}
	for _, f := range findings {
		sev := f.Severity
		if sev == "" {
			sev = "info"
		}
		counts[sev]++
	}
	return counts
}

func formatFindingCounts(total int, counts map[string]int) string {
	if total == 0 {
		return "0 findings"
	}
	label := "findings"
	if total == 1 {
		label = "finding"
	}
	order := []string{"critical", "warn", "info"}
	var sevParts []string
	for _, sev := range order {
		if n := counts[sev]; n > 0 {
			sevParts = append(sevParts, fmt.Sprintf("%d %s", n, sev))
		}
	}
	if len(sevParts) == 0 {
		return fmt.Sprintf("%d %s", total, label)
	}
	return fmt.Sprintf("%d %s (%s)", total, label, strings.Join(sevParts, ", "))
}

func wrapText(text string, width int, indent string) []string {
	if width <= len(indent)+10 {
		width = 80
	}
	max := width - len(indent)
	if max < 20 {
		max = 60
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	line := indent + words[0]
	for _, word := range words[1:] {
		candidate := line + " " + word
		if len(candidate) > max {
			lines = append(lines, line)
			line = indent + word
		} else {
			line = candidate
		}
	}
	lines = append(lines, line)
	return lines
}
