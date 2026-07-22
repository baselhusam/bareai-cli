package db

import (
	"context"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Input holds context for database discovery.
type Input struct {
	Docker *snapshot.Docker
	Probe  bool
}

type candidate struct {
	snapshot.Database
	priority int
}

const (
	sourceDocker  = "docker"
	sourceProcess = "process"
	sourcePort    = "port"
)

// Collect discovers local database instances and optionally probes them.
func Collect(ctx context.Context, in Input) ([]snapshot.Database, []snapshot.Skip, error) {
	var skips []snapshot.Skip
	var candidates []candidate

	candidates = append(candidates, discoverDocker(in)...)
	procCandidates, procSkips := discoverProcesses(ctx)
	candidates = append(candidates, procCandidates...)
	skips = append(skips, procSkips...)

	portCandidates, portSkips := discoverPorts(ctx)
	candidates = append(candidates, portCandidates...)
	skips = append(skips, portSkips...)

	merged := mergeCandidates(candidates)
	if len(merged) == 0 {
		return nil, skips, nil
	}

	out := make([]snapshot.Database, 0, len(merged))
	for _, c := range merged {
		db := c.Database
		if in.Probe {
			probeDB(ctx, &db)
		}
		correlate(&db, in)
		out = append(out, db)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Engine != out[j].Engine {
			return out[i].Engine < out[j].Engine
		}
		return out[i].Address < out[j].Address
	})
	return out, skips, nil
}

func mergeCandidates(candidates []candidate) []candidate {
	byAddr := make(map[string]candidate)
	for _, c := range candidates {
		key := normalizeAddress(c.Address)
		if key == "" {
			continue
		}
		c.Address = key
		if existing, ok := byAddr[key]; !ok || c.priority > existing.priority {
			byAddr[key] = c
		}
	}
	out := make([]candidate, 0, len(byAddr))
	for _, c := range byAddr {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Address < out[j].Address
	})
	return out
}

func normalizeAddress(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	host, portStr, err := net.SplitHostPort(raw)
	if err != nil {
		if !strings.Contains(raw, ":") {
			return ""
		}
		host = "127.0.0.1"
		portStr = strings.TrimPrefix(raw, ":")
	}
	host = strings.Trim(host, "[]")
	if host == "" || host == "0.0.0.0" || host == "*" {
		host = "127.0.0.1"
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil || port == 0 {
		return ""
	}
	return net.JoinHostPort(host, strconv.Itoa(int(port)))
}

func addressPort(address string) (uint16, bool) {
	_, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return 0, false
	}
	n, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, false
	}
	return uint16(n), true
}

func baseAddr(port uint16) string {
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port)))
}
