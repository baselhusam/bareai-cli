//go:build darwin

package gpu

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
	"github.com/shirou/gopsutil/v4/cpu"
)

const vendorApple = "apple"

func collectApple(ctx context.Context) ([]snapshot.GPU, error) {
	if gpus := collectAppleSystemProfiler(ctx); len(gpus) > 0 {
		return gpus, nil
	}
	return collectAppleFallback(ctx)
}

func collectAppleSystemProfiler(ctx context.Context) []snapshot.GPU {
	cmd := exec.CommandContext(ctx, "system_profiler", "SPDisplaysDataType", "-json")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(out, &payload); err != nil {
		return nil
	}

	raw, ok := payload["SPDisplaysDataType"]
	if !ok {
		return nil
	}

	var displays []map[string]any
	if err := json.Unmarshal(raw, &displays); err != nil {
		return nil
	}

	var gpus []snapshot.GPU
	index := 0
	for _, display := range displays {
		name := firstString(display, "sppci_model", "_name", "chipset_model")
		if name == "" {
			continue
		}
		gpus = append(gpus, snapshot.GPU{
			Index:  index,
			Vendor: vendorApple,
			Name:   name,
		})
		index++
	}

	return gpus
}

func collectAppleFallback(ctx context.Context) ([]snapshot.GPU, error) {
	if runtime.GOARCH != "arm64" {
		return nil, nil
	}

	infos, err := cpu.InfoWithContext(ctx)
	if err != nil || len(infos) == 0 {
		return nil, nil
	}

	name := strings.TrimSpace(infos[0].ModelName)
	if name == "" {
		name = "Apple Silicon GPU"
	}

	return []snapshot.GPU{{
		Index:  0,
		Vendor: vendorApple,
		Name:   name,
	}}, nil
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
