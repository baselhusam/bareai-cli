package llm

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

var (
	portFlagRE    = regexp.MustCompile(`(?i)(?:--port|-p)\s*(?:=?\s*)?(\d{2,5})`)
	hostPortRE    = regexp.MustCompile(`(?i)(?:OLLAMA_HOST|HOST)\s*[:=]\s*[^:]*:(\d{2,5})`)
	processNames  = map[string]string{
		"ollama":       probe.RuntimeOllama,
		"tritonserver": probe.RuntimeTriton,
		"vllm":         probe.RuntimeVLLM,
		"sglang":       probe.RuntimeSGLang,
	}
)

func discoverProcesses(ctx context.Context) ([]candidate, []snapshot.Skip) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, []snapshot.Skip{{Component: "llm.process", Reason: err.Error()}}
	}

	var out []candidate
	seen := make(map[string]bool)
	for _, proc := range procs {
		name, err := proc.NameWithContext(ctx)
		if err != nil {
			continue
		}
		cmdline, _ := proc.CmdlineWithContext(ctx)
		runtime := matchProcessRuntime(name, cmdline)
		if runtime == "" {
			continue
		}
		pid := int(proc.Pid)
		port := extractPortFromCmdline(cmdline)
		if port == 0 {
			port = listenPortForPID(ctx, pid, defaultPortsForRuntime(runtime))
		}
		if port == 0 {
			continue
		}
		endpoint := baseURL(port)
		if seen[endpoint] {
			continue
		}
		seen[endpoint] = true
		out = append(out, candidate{
			priority: 2,
			LLM: snapshot.LLM{
				Runtime:  runtime,
				Name:     displayName(runtime),
				Endpoint: endpoint,
				Source:   sourceProcess,
				PID:      pid,
			},
		})
	}
	return out, nil
}

func matchProcessRuntime(name, cmdline string) string {
	name = strings.ToLower(name)
	name = strings.TrimSuffix(name, ".exe")
	cmdline = strings.ToLower(cmdline)
	if runtime, ok := processNames[name]; ok {
		return runtime
	}
	for key, runtime := range processNames {
		if strings.Contains(cmdline, key) {
			return runtime
		}
	}
	if strings.Contains(name, "python") || strings.Contains(name, "python3") {
		if strings.Contains(cmdline, "vllm") {
			return probe.RuntimeVLLM
		}
		if strings.Contains(cmdline, "sglang") {
			return probe.RuntimeSGLang
		}
	}
	return ""
}

func extractPortFromCmdline(cmdline string) uint16 {
	if m := portFlagRE.FindStringSubmatch(cmdline); len(m) == 2 {
		if p, err := strconv.ParseUint(m[1], 10, 16); err == nil {
			return uint16(p)
		}
	}
	if m := hostPortRE.FindStringSubmatch(cmdline); len(m) == 2 {
		if p, err := strconv.ParseUint(m[1], 10, 16); err == nil {
			return uint16(p)
		}
	}
	return 0
}

func defaultPortsForRuntime(runtime string) []uint16 {
	for _, h := range runtimeHints {
		if h.runtime == runtime {
			return h.ports
		}
	}
	return nil
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
