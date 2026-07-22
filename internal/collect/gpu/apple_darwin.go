//go:build darwin

package gpu

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const vendorApple = "apple"

const appleGPUNotes = "unified memory; no util/temp via public APIs"

func collectApple(ctx context.Context) ([]snapshot.GPU, error) {
	unifiedTotal := hostUnifiedMemory(ctx)
	if gpus := collectAppleSystemProfiler(ctx, unifiedTotal); len(gpus) > 0 {
		return gpus, nil
	}
	return collectAppleFallback(ctx, unifiedTotal)
}

func hostUnifiedMemory(ctx context.Context) uint64 {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return 0
	}
	return vm.Total
}

func collectAppleSystemProfiler(ctx context.Context, unifiedTotal uint64) []snapshot.GPU {
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
		name := firstString(display, "sppci_model", "_name", "chipset_model", "spdisplays_mtlgpufamilysupport")
		if name == "" {
			continue
		}
		chipset := firstString(display, "spdisplays_chipset", "sppci_bus")
		if chipset != "" && !strings.Contains(name, chipset) {
			name = name + " (" + chipset + ")"
		}

		gpu := snapshot.GPU{
			Index:  index,
			Vendor: vendorApple,
			Name:   name,
			Notes:  appleGPUNotes,
		}
		if vram := parseAppleVRAM(display); vram > 0 {
			gpu.MemoryTotal = vram
		} else if unifiedTotal > 0 && runtime.GOARCH == "arm64" {
			gpu.MemoryTotal = unifiedTotal
		}
		gpus = append(gpus, gpu)
		index++
	}

	return gpus
}

func parseAppleVRAM(display map[string]any) uint64 {
	for _, key := range []string{"spdisplays_vram", "spdisplays_vram_shared", "vram_shared"} {
		if v, ok := display[key].(string); ok {
			if n := parseAppleSizeString(v); n > 0 {
				return n
			}
		}
	}
	return 0
}

func parseAppleSizeString(s string) uint64 {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0
	}
	mult := uint64(1)
	switch {
	case strings.HasSuffix(s, "gb"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "gb")
	case strings.HasSuffix(s, "mb"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "mb")
	default:
		return 0
	}
	if v, err := parseFloatPrefix(s); err == nil {
		return uint64(v * float64(mult))
	}
	return 0
}

func parseFloatPrefix(s string) (float64, error) {
	i := 0
	for i < len(s) && (s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
		i++
	}
	if i == 0 {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseFloat(s[:i], 64)
}

func collectAppleFallback(ctx context.Context, unifiedTotal uint64) ([]snapshot.GPU, error) {
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

	gpu := snapshot.GPU{
		Index:  0,
		Vendor: vendorApple,
		Name:   name,
		Notes:  appleGPUNotes,
	}
	if unifiedTotal > 0 {
		gpu.MemoryTotal = unifiedTotal
	}

	return []snapshot.GPU{gpu}, nil
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
