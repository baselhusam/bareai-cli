package gpu

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

const vendorNVIDIA = "nvidia"

func collectNVIDIA(ctx context.Context) ([]snapshot.GPU, error) {
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		return nil, nil
	}

	check := exec.CommandContext(ctx, "nvidia-smi", "-L")
	if err := check.Run(); err != nil {
		return nil, nil
	}

	gpuOut, err := runNVIDIAQuery(ctx,
		"index,name,uuid,driver_version,memory.total,memory.used,utilization.gpu,temperature.gpu,power.draw,power.limit",
	)
	if err != nil {
		return nil, err
	}

	gpus, err := parseNVIDIAGPUsCSV(gpuOut)
	if err != nil {
		return nil, fmt.Errorf("parse nvidia-smi gpu query: %w", err)
	}

	procOut, err := runNVIDIAQuery(ctx,
		"gpu_uuid,pid,process_name,used_gpu_memory",
		"compute-apps",
	)
	if err == nil && strings.TrimSpace(procOut) != "" {
		procs, procErr := parseNVIDIAProcessesCSV(procOut)
		if procErr != nil {
			return nil, fmt.Errorf("parse nvidia-smi process query: %w", procErr)
		}
		attachNVIDIAProcesses(gpus, procs)
	}

	return gpus, nil
}

func runNVIDIAQuery(ctx context.Context, fields string, queryType ...string) (string, error) {
	queryFlag := "gpu"
	if len(queryType) > 0 {
		queryFlag = queryType[0]
	}
	args := []string{
		fmt.Sprintf("--query-%s=%s", queryFlag, fields),
		"--format=csv,noheader,nounits",
	}
	cmd := exec.CommandContext(ctx, "nvidia-smi", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("nvidia-smi: %w", err)
	}
	return string(out), nil
}

func parseNVIDIAGPUsCSV(data string) ([]snapshot.GPU, error) {
	reader := csv.NewReader(strings.NewReader(data))
	reader.TrimLeadingSpace = true

	var gpus []snapshot.GPU
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) < 10 {
			return nil, fmt.Errorf("expected 10 fields, got %d", len(record))
		}

		index, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			return nil, fmt.Errorf("gpu index: %w", err)
		}

		memTotal, err := parseNVIDIAMiB(record[4])
		if err != nil {
			return nil, err
		}
		memUsed, err := parseNVIDIAMiB(record[5])
		if err != nil {
			return nil, err
		}

		gpus = append(gpus, snapshot.GPU{
			Index:       index,
			Vendor:      vendorNVIDIA,
			Name:        strings.TrimSpace(record[1]),
			UUID:        strings.TrimSpace(record[2]),
			Driver:      strings.TrimSpace(record[3]),
			MemoryTotal: memTotal,
			MemoryUsed:  memUsed,
			Utilization: parseOptionalFloat(record[6]),
			Temperature: parseOptionalFloat(record[7]),
			PowerDrawW:  parseOptionalFloat(record[8]),
			PowerLimitW: parseOptionalFloat(record[9]),
		})
	}

	return gpus, nil
}

type nvidiaProcess struct {
	GPUUUID string
	Proc    snapshot.GPUProcess
}

func parseNVIDIAProcessesCSV(data string) ([]nvidiaProcess, error) {
	reader := csv.NewReader(strings.NewReader(data))
	reader.TrimLeadingSpace = true

	var procs []nvidiaProcess
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) < 4 {
			return nil, fmt.Errorf("expected 4 process fields, got %d", len(record))
		}

		pid, err := strconv.Atoi(strings.TrimSpace(record[1]))
		if err != nil {
			return nil, fmt.Errorf("process pid: %w", err)
		}

		memUsed, err := parseNVIDIAMiB(record[3])
		if err != nil {
			return nil, err
		}

		procs = append(procs, nvidiaProcess{
			GPUUUID: strings.TrimSpace(record[0]),
			Proc: snapshot.GPUProcess{
				PID:        pid,
				Name:       strings.TrimSpace(record[2]),
				MemoryUsed: memUsed,
			},
		})
	}

	return procs, nil
}

func attachNVIDIAProcesses(gpus []snapshot.GPU, procs []nvidiaProcess) {
	byUUID := make(map[string]int, len(gpus))
	for i := range gpus {
		byUUID[gpus[i].UUID] = i
	}

	for _, proc := range procs {
		idx, ok := byUUID[proc.GPUUUID]
		if !ok {
			continue
		}
		gpus[idx].Processes = append(gpus[idx].Processes, proc.Proc)
	}
}

func parseNVIDIAMiB(value string) (uint64, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "[N/A]") {
		return 0, nil
	}
	miB, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("memory value %q: %w", value, err)
	}
	return uint64(miB * 1024 * 1024), nil
}

func parseOptionalFloat(value string) *float64 {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "[N/A]") {
		return nil
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &f
}
