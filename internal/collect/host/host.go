package host

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// Collect gathers host inventory using gopsutil.
func Collect(ctx context.Context) (snapshot.Host, error) {
	if err := ctx.Err(); err != nil {
		return snapshot.Host{}, err
	}

	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return snapshot.Host{}, fmt.Errorf("host info: %w", err)
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return snapshot.Host{}, fmt.Errorf("memory: %w", err)
	}

	cores, err := cpu.CountsWithContext(ctx, false)
	if err != nil {
		return snapshot.Host{}, fmt.Errorf("cpu cores: %w", err)
	}

	logical, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return snapshot.Host{}, fmt.Errorf("cpu logical: %w", err)
	}

	h := snapshot.Host{
		Hostname:     info.Hostname,
		OS:           info.OS,
		Platform:     info.Platform,
		PlatformVer:  info.PlatformVersion,
		Arch:         runtime.GOARCH,
		Uptime:       timeFromSeconds(info.Uptime),
		CPUCores:     cores,
		CPULogical:   logical,
		MemTotal:     vm.Total,
		MemUsed:      vm.Used,
		MemAvailable: vm.Available,
	}

	h.CPUModel = cpuModel(ctx)
	h.Load1, h.Load5, h.Load15 = loadAverages(ctx)
	h.Disks = collectDisks(ctx)

	return h, nil
}

func cpuModel(ctx context.Context) string {
	infos, err := cpu.InfoWithContext(ctx)
	if err != nil || len(infos) == 0 {
		return ""
	}
	return strings.TrimSpace(infos[0].ModelName)
}

func loadAverages(ctx context.Context) (float64, float64, float64) {
	avg, err := load.AvgWithContext(ctx)
	if err != nil {
		return 0, 0, 0
	}
	return avg.Load1, avg.Load5, avg.Load15
}

func collectDisks(ctx context.Context) []snapshot.Disk {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil
	}

	disks := make([]snapshot.Disk, 0, len(partitions))
	for _, part := range partitions {
		if skipMount(part.Mountpoint, part.Fstype) {
			continue
		}

		usage, err := disk.UsageWithContext(ctx, part.Mountpoint)
		if err != nil {
			continue
		}

		disks = append(disks, snapshot.Disk{
			Mount:  part.Mountpoint,
			FSType: part.Fstype,
			Total:  usage.Total,
			Used:   usage.Used,
			Free:   usage.Free,
		})
	}

	return disks
}

func skipMount(mountpoint, fstype string) bool {
	switch mountpoint {
	case "/proc", "/sys", "/dev", "/run", "/dev/shm":
		return true
	}
	if strings.HasPrefix(mountpoint, "/proc/") ||
		strings.HasPrefix(mountpoint, "/sys/") ||
		strings.HasPrefix(mountpoint, "/dev/") {
		return true
	}
	if fstype == "tmpfs" && (mountpoint == "/run" || strings.HasPrefix(mountpoint, "/run/")) {
		return true
	}
	return false
}

func timeFromSeconds(seconds uint64) time.Duration {
	return time.Duration(seconds) * time.Second
}
