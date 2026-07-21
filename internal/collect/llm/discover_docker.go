package llm

import (
	"strings"

	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

type runtimeHint struct {
	runtime       string
	imagePatterns []string
	namePatterns  []string
	ports         []uint16
}

var runtimeHints = []runtimeHint{
	{
		runtime:       probe.RuntimeOllama,
		imagePatterns: []string{"ollama"},
		namePatterns:  []string{"ollama"},
		ports:         []uint16{11434},
	},
	{
		runtime:       probe.RuntimeVLLM,
		imagePatterns: []string{"vllm"},
		namePatterns:  []string{"vllm"},
		ports:         []uint16{8000},
	},
	{
		runtime:       probe.RuntimeSGLang,
		imagePatterns: []string{"sglang"},
		namePatterns:  []string{"sglang"},
		ports:         []uint16{30000, 8000},
	},
	{
		runtime:       probe.RuntimeTriton,
		imagePatterns: []string{"triton"},
		namePatterns:  []string{"triton"},
		ports:         []uint16{8000, 8001, 8002},
	},
}

func discoverDocker(in Input) []candidate {
	if in.Docker == nil || !in.Docker.Available {
		return nil
	}
	var out []candidate
	for _, c := range in.Docker.Containers {
		if !strings.EqualFold(c.State, "running") {
			continue
		}
		hint := matchRuntimeHint(c.Image, c.Name)
		if hint == nil {
			continue
		}
		port := pickPort(c.Ports, hint.ports)
		if port == 0 {
			continue
		}
		out = append(out, candidate{
			priority: 3,
			LLM: snapshot.LLM{
				Runtime:       hint.runtime,
				Name:          displayName(hint.runtime),
				Endpoint:      baseURL(port),
				Source:        sourceDocker,
				PID:           c.PID,
				ContainerID:   c.ID,
				ContainerName: c.Name,
			},
		})
	}
	return out
}

func matchRuntimeHint(image, name string) *runtimeHint {
	image = strings.ToLower(image)
	name = strings.ToLower(name)
	for i := range runtimeHints {
		h := &runtimeHints[i]
		for _, p := range h.imagePatterns {
			if strings.Contains(image, p) {
				return h
			}
		}
		for _, p := range h.namePatterns {
			if strings.Contains(name, p) {
				return h
			}
		}
	}
	return nil
}

func pickPort(ports []snapshot.DockerPort, defaults []uint16) uint16 {
	for _, p := range ports {
		if p.PublicPort > 0 {
			return p.PublicPort
		}
	}
	for _, p := range ports {
		if p.PrivatePort > 0 {
			return p.PrivatePort
		}
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return 0
}

func displayName(runtime string) string {
	switch runtime {
	case probe.RuntimeOllama:
		return "Ollama"
	case probe.RuntimeVLLM:
		return "vLLM"
	case probe.RuntimeSGLang:
		return "SGLang"
	case probe.RuntimeTriton:
		return "Triton"
	default:
		return runtime
	}
}
