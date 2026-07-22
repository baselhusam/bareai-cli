package tui

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func dbListTitle(db snapshot.Database, s styles) string {
	health := s.muted.Render("?")
	if db.Health != nil {
		label := "fail"
		if db.Health.OK {
			label = "ok"
		}
		health = s.healthStyle(db.Health.OK).Render(label)
	}
	name := db.Name
	if name == "" {
		name = db.Engine
	}
	pid := ""
	if db.PID > 0 {
		pid = fmt.Sprintf(" pid=%d", db.PID)
	}
	return fmt.Sprintf("%s  %s%s  [%s]",
		truncate(name, 14), truncate(db.Address, 24), pid, health)
}

func dbFilterValue(db snapshot.Database) string {
	parts := []string{
		db.Engine,
		db.Name,
		db.Address,
		db.Source,
		db.ContainerName,
		db.ContainerID,
		db.Version,
		fmt.Sprintf("%d", db.PID),
	}
	return strings.Join(parts, " ")
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

func dbDetailText(db snapshot.Database, s styles) string {
	var b strings.Builder
	source := dbSourceLabel(db)
	fmt.Fprintf(&b, "%s  %s\n", s.value.Render(db.Engine+" / "+db.Name), db.Address)
	fmt.Fprintf(&b, "  Engine:    %s  (%s)\n", db.Engine, source)
	if db.Version != "" {
		fmt.Fprintf(&b, "  Version:   %s\n", db.Version)
	}
	if db.PID > 0 {
		fmt.Fprintf(&b, "  PID:       %d\n", db.PID)
	}
	if db.Health != nil {
		status := s.fail.Render("fail")
		if db.Health.OK {
			status = s.ok.Render("ok")
		}
		line := fmt.Sprintf("  Health:    %s  %dms", status, db.Health.LatencyMS)
		if db.Health.Message != "" {
			line += "  " + db.Health.Message
		}
		if !db.Health.OK && db.Health.Error != "" {
			line += "  " + db.Health.Error
		}
		fmt.Fprintln(&b, line)
	}
	if db.ContainerName != "" {
		fmt.Fprintf(&b, "  Container: %s\n", db.ContainerName)
	}
	return strings.TrimRight(b.String(), "\n")
}
