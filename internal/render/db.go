package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// WriteDB renders discovered database instances.
func WriteDB(w io.Writer, snap *snapshot.Snapshot, noColor bool) error {
	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}

	label := func(s string) string {
		if noColor {
			return s
		}
		return s
	}

	if _, err := fmt.Fprintf(w, "%s\n", label("bareai db")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Collected: %s\n\n", snap.CollectedAt.Format(time.RFC3339)); err != nil {
		return err
	}

	if len(snap.Databases) == 0 {
		if _, err := fmt.Fprintf(w, "%s\n", EmptyHint("db")); err != nil {
			return err
		}
		return writeSkipped(w, snap.Skipped)
	}

	for i, db := range snap.Databases {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := writeDBDetail(w, db); err != nil {
			return err
		}
	}

	return writeSkipped(w, snap.Skipped)
}

func writeDBDetail(w io.Writer, db snapshot.Database) error {
	source := dbSourceLabel(db)
	line := fmt.Sprintf("%s  %s  (%s)", db.Name, db.Address, source)
	if db.Version != "" {
		line += "  v" + db.Version
	}
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}
	if db.Health != nil {
		status := "fail"
		if db.Health.OK {
			status = "ok"
		}
		healthLine := fmt.Sprintf("  Health: %s  %dms", status, db.Health.LatencyMS)
		if db.Health.Message != "" {
			healthLine += "  " + db.Health.Message
		}
		if !db.Health.OK && db.Health.Error != "" {
			healthLine += "  " + db.Health.Error
		}
		if _, err := fmt.Fprintln(w, healthLine); err != nil {
			return err
		}
	}
	if db.ContainerName != "" {
		if _, err := fmt.Fprintf(w, "  Container: %s\n", db.ContainerName); err != nil {
			return err
		}
	}
	if db.PID > 0 {
		if _, err := fmt.Fprintf(w, "  PID: %d\n", db.PID); err != nil {
			return err
		}
	}
	return nil
}

func dbSourceLabel(db snapshot.Database) string {
	switch db.Source {
	case "docker":
		if db.ContainerName != "" {
			return "docker: " + db.ContainerName
		}
		return "docker"
	case "process":
		if db.PID > 0 {
			return fmt.Sprintf("process pid %d", db.PID)
		}
		return "process"
	case "port":
		return "port scan"
	default:
		return db.Source
	}
}

func writeDBSummary(w io.Writer, dbs []snapshot.Database) error {
	if len(dbs) == 0 {
		if _, err := fmt.Fprintf(w, "DB:          %s\n", EmptyHint("db")); err != nil {
			return err
		}
		return nil
	}

	healthy := 0
	engines := make(map[string]bool)
	for _, db := range dbs {
		if db.Engine != "" {
			engines[strings.ToLower(db.Engine)] = true
		}
		if db.Health != nil && db.Health.OK {
			healthy++
		}
	}
	names := make([]string, 0, len(engines))
	for name := range engines {
		names = append(names, name)
	}
	sortStrings(names)

	line := fmt.Sprintf("DB:          %d instances (%s)", len(dbs), strings.Join(names, ", "))
	if healthy > 0 {
		line += fmt.Sprintf(" — %d reachable", healthy)
	}
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}
	return nil
}
