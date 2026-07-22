package db

import (
	"context"
	"regexp"
	"strconv"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

var portFlagRE = regexp.MustCompile(`(?i)(?:--port|-p)\s*(?:=?\s*)?(\d{2,5})`)

func discoverProcesses(ctx context.Context) ([]candidate, []snapshot.Skip) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, []snapshot.Skip{{Component: "db.process", Reason: err.Error()}}
	}

	var out []candidate
	seen := make(map[string]bool)
	for _, proc := range procs {
		name, err := proc.NameWithContext(ctx)
		if err != nil {
			continue
		}
		cmdline, _ := proc.CmdlineWithContext(ctx)
		hint := matchProcessEngine(name, cmdline)
		if hint == nil {
			continue
		}
		pid := int(proc.Pid)
		port := extractPortFromCmdline(cmdline)
		if port == 0 {
			port = listenPortForPID(ctx, pid, hint.ports)
		}
		if port == 0 {
			continue
		}
		addr := baseAddr(port)
		if seen[addr] {
			continue
		}
		seen[addr] = true
		out = append(out, candidate{
			priority: 2,
			Database: snapshot.Database{
				Engine:  hint.engine,
				Name:    displayName(hint.engine),
				Address: addr,
				Source:  sourceProcess,
				PID:     pid,
			},
		})
	}
	return out, nil
}

func extractPortFromCmdline(cmdline string) uint16 {
	if m := portFlagRE.FindStringSubmatch(cmdline); len(m) == 2 {
		if p, err := strconv.ParseUint(m[1], 10, 16); err == nil {
			return uint16(p)
		}
	}
	return 0
}

func listenPortForPID(ctx context.Context, pid int, ports []uint16) uint16 {
	conns, err := net.ConnectionsPidWithContext(ctx, "inet", int32(pid))
	if err != nil {
		return 0
	}
	portSet := make(map[uint16]bool, len(ports))
	for _, p := range ports {
		portSet[p] = true
	}
	for _, c := range conns {
		if c.Status != "LISTEN" {
			continue
		}
		p := uint16(c.Laddr.Port)
		if len(portSet) == 0 || portSet[p] {
			return p
		}
	}
	return 0
}
