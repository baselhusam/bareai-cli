//go:build linux

package gpu

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

const vendorAMD = "amd"

func collectAMD(ctx context.Context) ([]snapshot.GPU, error) {
	if _, err := exec.LookPath("rocm-smi"); err == nil {
		return collectAMDROCm(ctx)
	}
	return collectAMDSysfs()
}

func collectAMDROCm(ctx context.Context) ([]snapshot.GPU, error) {
	cmd := exec.CommandContext(ctx, "rocm-smi", "--showproductname", "--showmeminfo", "vram", "--showtemp", "--showuse", "--json")
	out, err := cmd.Output()
	if err != nil {
		return collectAMDSysfs()
	}
	return parseROCmSMIJSON(string(out))
}

func parseROCmSMIJSON(data string) ([]snapshot.GPU, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return collectAMDSysfs()
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &root); err != nil {
		return collectAMDSysfs()
	}

	var gpus []snapshot.GPU
	index := 0
	for cardKey, raw := range root {
		if !strings.HasPrefix(cardKey, "card") {
			continue
		}
		var card map[string]json.RawMessage
		if err := json.Unmarshal(raw, &card); err != nil {
			continue
		}

		name := rocmString(card, "Card series", "Card model", "Card Series", "Card Model")
		if name == "" {
			name = "AMD GPU"
		}

		gpu := snapshot.GPU{
			Index:  index,
			Vendor: vendorAMD,
			Name:   name,
		}

		if total, used, ok := rocmVRAM(card); ok {
			gpu.MemoryTotal = total
			gpu.MemoryUsed = used
		}
		if temp := rocmFloat(card, "Temperature (Sensor edge) (C)", "Temperature (Edge)", "Temperature"); temp != nil {
			gpu.Temperature = temp
		}
		if util := rocmFloat(card, "GPU use (%)", "GPU Use (%)", "GPU use"); util != nil {
			gpu.Utilization = util
		}

		gpus = append(gpus, gpu)
		index++
	}

	if len(gpus) == 0 {
		return collectAMDSysfs()
	}
	return gpus, nil
}

func rocmString(card map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		raw, ok := card[key]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(raw, &s); err == nil && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func rocmFloat(card map[string]json.RawMessage, keys ...string) *float64 {
	for _, key := range keys {
		raw, ok := card[key]
		if !ok {
			continue
		}
		var f float64
		if err := json.Unmarshal(raw, &f); err == nil {
			return &f
		}
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
			if v, err := strconv.ParseFloat(s, 64); err == nil {
				return &v
			}
		}
	}
	return nil
}

func rocmVRAM(card map[string]json.RawMessage) (total, used uint64, ok bool) {
	for key, raw := range card {
		if !strings.Contains(strings.ToLower(key), "vram") {
			continue
		}
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			continue
		}
		parts := strings.Fields(strings.ReplaceAll(s, "/", " "))
		for _, p := range parts {
			if v, err := parseByteSize(p); err == nil {
				if total == 0 {
					total = v
				} else {
					used = v
					return total, used, true
				}
			}
		}
	}
	return 0, 0, false
}

func parseByteSize(s string) (uint64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	multiplier := uint64(1)
	switch {
	case strings.HasSuffix(s, "GIB"):
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GIB")
	case strings.HasSuffix(s, "MIB"):
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MIB")
	case strings.HasSuffix(s, "GB"):
		multiplier = 1000 * 1000 * 1000
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1000 * 1000
		s = strings.TrimSuffix(s, "MB")
	}
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return uint64(v * float64(multiplier)), nil
}

func collectAMDSysfs() ([]snapshot.GPU, error) {
	cards, err := filepath.Glob("/sys/class/drm/card*/device")
	if err != nil {
		return nil, err
	}

	var gpus []snapshot.GPU
	index := 0
	for _, devicePath := range cards {
		vendor, err := readSysfsTrim(filepath.Join(devicePath, "vendor"))
		if err != nil || !isAMDVendor(vendor) {
			continue
		}

		name, _ := readSysfsTrim(filepath.Join(devicePath, "product_name"))
		if name == "" {
			name, _ = readSysfsTrim(filepath.Join(devicePath, "name"))
		}
		if name == "" {
			name = "AMD GPU"
		}

		memTotal, _ := readSysfsUint(filepath.Join(devicePath, "mem_info_vram_total"))
		memUsed, _ := readSysfsUint(filepath.Join(devicePath, "mem_info_vram_used"))
		temp := readAMDHWMonTemp(devicePath)

		gpu := snapshot.GPU{
			Index:       index,
			Vendor:      vendorAMD,
			Name:        name,
			MemoryTotal: memTotal,
			MemoryUsed:  memUsed,
			Temperature: temp,
		}
		gpus = append(gpus, gpu)
		index++
	}

	return gpus, nil
}

func isAMDVendor(vendor string) bool {
	v := strings.ToLower(strings.TrimSpace(vendor))
	return v == "0x1002" || v == "4098"
}

func readSysfsTrim(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func readSysfsUint(path string) (uint64, error) {
	text, err := readSysfsTrim(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(text, 10, 64)
}

func readAMDHWMonTemp(devicePath string) *float64 {
	matches, err := filepath.Glob(filepath.Join(devicePath, "hwmon", "hwmon*", "temp1_input"))
	if err != nil || len(matches) == 0 {
		return nil
	}
	value, err := readSysfsTrim(matches[0])
	if err != nil {
		return nil
	}
	milli, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	celsius := milli / 1000.0
	return &celsius
}
