package rules

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// DB returns database-related findings.
func DB(snap *snapshot.Snapshot) []snapshot.Finding {
	if snap == nil {
		return nil
	}
	var out []snapshot.Finding

	dockerAvailable := snap.Docker != nil && snap.Docker.Available

	for _, db := range snap.Databases {
		if db.Health != nil && !db.Health.OK {
			try := fmt.Sprintf("nc -zv %s  ·  bareai db --json  ·  docker ps --filter name=%s",
				db.Address, db.ContainerName)
			if db.Engine == "postgres" {
				try = fmt.Sprintf("pg_isready -h %s  ·  bareai db --json", db.Address)
			}
			out = append(out, findingWithDo(
				"db.unreachable",
				SeverityWarn,
				"db",
				rankWarn+10,
				fmt.Sprintf("%s (%s) is unreachable", db.Name, db.Address),
				"TCP/version probe failed; service may be down or blocked.",
				try,
				containerOffers(db.ContainerID, db.ContainerName, "logs", "restart"),
			))
		}

		if db.Source == "docker" && !dockerAvailable {
			out = append(out, finding(
				"db.docker_source_no_engine",
				SeverityWarn,
				"db",
				rankWarn+15,
				fmt.Sprintf("%s was discovered via Docker but Docker is unavailable", db.Name),
				"Container metadata may be stale; verify the daemon is running.",
				"docker ps  ·  bareai docker --json  ·  bareai db --json",
			))
		}
	}

	if len(snap.Databases) > 3 {
		out = append(out, finding(
			"db.multiple_instances",
			SeverityInfo,
			"db",
			rankInfo+5,
			fmt.Sprintf("%d database instances discovered on this host", len(snap.Databases)),
			"Multiple local databases may compete for memory and ports.",
			"bareai inspect  ·  bareai db --json",
		))
	}

	return out
}

// EmptyBox returns a friendly first-run finding when the box has no AI stack.
func EmptyBox(snap *snapshot.Snapshot) []snapshot.Finding {
	if snap == nil {
		return nil
	}
	if len(snap.LLMs) > 0 || len(snap.Databases) > 0 {
		return nil
	}
	if len(snap.GPUs) > 0 {
		return nil
	}
	if snap.Docker != nil && snap.Docker.Available {
		running := 0
		for _, c := range snap.Docker.Containers {
			if strings.EqualFold(c.State, "running") {
				running++
			}
		}
		if running > 0 {
			return nil
		}
	}
	if snap.Host == nil {
		return nil
	}

	return []snapshot.Finding{finding(
		"host.empty_box",
		SeverityInfo,
		"host",
		rankInfo,
		"No LLMs, databases, or GPUs detected yet — this box looks idle",
		"bareai watches for Ollama, OpenAI-compatible servers, and local databases on default ports.",
		"Start Ollama (:11434) or Docker Desktop  ·  bareai watch  ·  bareai doctor --share",
	)}
}
