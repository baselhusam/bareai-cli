//go:build linux

package gpu

import (
	"context"
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
	// rocm-smi --json output varies by version; fall back to sysfs if parsing fails.
	_ = data
	return collectAMDSysfs()
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
