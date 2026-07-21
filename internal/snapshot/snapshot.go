package snapshot

import "time"

// Snapshot is the top-level correlated infrastructure view.
type Snapshot struct {
	CollectedAt time.Time `json:"collected_at"`
	Host        *Host     `json:"host,omitempty"`
	GPUs        []GPU     `json:"gpus,omitempty"`
	Docker      *Docker   `json:"docker,omitempty"`
	LLMs          []LLM         `json:"llms,omitempty"`
	Correlations  []Correlation `json:"correlations,omitempty"`
	Findings      []Finding     `json:"findings,omitempty"`
	Skipped     []Skip    `json:"skipped,omitempty"`
}

// Host holds bare-metal host inventory.
type Host struct {
	Hostname     string        `json:"hostname"`
	OS           string        `json:"os"`
	Platform     string        `json:"platform"`
	PlatformVer  string        `json:"platform_version"`
	Arch         string        `json:"arch"`
	Uptime       time.Duration `json:"uptime"`
	CPUModel     string        `json:"cpu_model"`
	CPUCores     int           `json:"cpu_cores"`
	CPULogical   int           `json:"cpu_logical"`
	Load1        float64       `json:"load1"`
	Load5        float64       `json:"load5"`
	Load15       float64       `json:"load15"`
	MemTotal     uint64        `json:"mem_total_bytes"`
	MemUsed      uint64        `json:"mem_used_bytes"`
	MemAvailable uint64        `json:"mem_available_bytes"`
	Disks        []Disk        `json:"disks"`
}

// Disk describes a mounted filesystem.
type Disk struct {
	Mount  string `json:"mount"`
	FSType string `json:"fstype"`
	Total  uint64 `json:"total_bytes"`
	Used   uint64 `json:"used_bytes"`
	Free   uint64 `json:"free_bytes"`
}

// GPU holds accelerator inventory and metrics.
type GPU struct {
	Index       int          `json:"index"`
	Vendor      string       `json:"vendor"`
	Name        string       `json:"name"`
	UUID        string       `json:"uuid,omitempty"`
	Driver      string       `json:"driver,omitempty"`
	MemoryTotal uint64       `json:"memory_total_bytes"`
	MemoryUsed  uint64       `json:"memory_used_bytes"`
	Utilization *float64     `json:"utilization_pct,omitempty"`
	Temperature *float64     `json:"temperature_c,omitempty"`
	PowerDrawW  *float64     `json:"power_draw_w,omitempty"`
	PowerLimitW *float64     `json:"power_limit_w,omitempty"`
	Processes   []GPUProcess `json:"processes,omitempty"`
}

// GPUProcess describes a process using a GPU.
type GPUProcess struct {
	PID        int    `json:"pid"`
	Name       string `json:"name,omitempty"`
	MemoryUsed uint64 `json:"memory_used_bytes"`
}

// Docker holds read-only Docker Engine inventory.
type Docker struct {
	Available      bool              `json:"available"`
	ServerVersion  string            `json:"server_version,omitempty"`
	APIVersion     string            `json:"api_version,omitempty"`
	OSType         string            `json:"os_type,omitempty"`
	Architecture   string            `json:"architecture,omitempty"`
	DefaultRuntime string            `json:"default_runtime,omitempty"`
	Runtimes       []string          `json:"runtimes,omitempty"`
	NVIDIARuntime  bool              `json:"nvidia_runtime"`
	Containers     []DockerContainer `json:"containers,omitempty"`
	Images         []DockerImage     `json:"images,omitempty"`
	Volumes        []DockerVolume    `json:"volumes,omitempty"`
}

// DockerContainer describes a container relevant to AI workloads.
type DockerContainer struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Image          string            `json:"image"`
	State          string            `json:"state"`
	Status         string            `json:"status"`
	Created        time.Time         `json:"created"`
	PID            int               `json:"pid,omitempty"`
	Ports          []DockerPort      `json:"ports,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	GPURequested   bool              `json:"gpu_requested"`
	DeviceRequests []DeviceRequest   `json:"device_requests,omitempty"`
}

// DockerPort describes a published or exposed port.
type DockerPort struct {
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port,omitempty"`
	Type        string `json:"type"`
	IP          string `json:"ip,omitempty"`
}

// DeviceRequest describes a container device request (e.g. NVIDIA GPU).
type DeviceRequest struct {
	Driver       string   `json:"driver,omitempty"`
	Count        int      `json:"count,omitempty"`
	DeviceIDs    []string `json:"device_ids,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// DockerImage describes a local image.
type DockerImage struct {
	ID       string    `json:"id"`
	RepoTags []string  `json:"repo_tags,omitempty"`
	Size     uint64    `json:"size_bytes"`
	Created  time.Time `json:"created"`
}

// DockerVolume describes a named volume.
type DockerVolume struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Mountpoint string `json:"mountpoint,omitempty"`
	Scope      string `json:"scope,omitempty"`
}

// LLM describes a discovered local inference server.
type LLM struct {
	Runtime       string             `json:"runtime"`
	Name          string             `json:"name"`
	Endpoint      string             `json:"endpoint"`
	Source        string             `json:"source"`
	PID           int                `json:"pid,omitempty"`
	ContainerID   string             `json:"container_id,omitempty"`
	ContainerName string             `json:"container_name,omitempty"`
	GPUIndex      *int               `json:"gpu_index,omitempty"`
	Models        []LLMModel         `json:"models,omitempty"`
	Health        *ProbeResult       `json:"health,omitempty"`
	Probe         *ProbeResult       `json:"probe,omitempty"`
	Metrics       map[string]float64 `json:"metrics,omitempty"`
}

// LLMModel describes a model served by an inference runtime.
type LLMModel struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Size uint64 `json:"size_bytes,omitempty"`
}

// ProbeResult holds the outcome of an HTTP health or smoke probe.
type ProbeResult struct {
	OK        bool   `json:"ok"`
	LatencyMS int64  `json:"latency_ms"`
	Status    int    `json:"status_code,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Correlation links an LLM endpoint to container, process, and GPU resources.
type Correlation struct {
	Endpoint      string   `json:"endpoint"`
	Runtime       string   `json:"runtime"`
	ContainerName string   `json:"container_name,omitempty"`
	ContainerID   string   `json:"container_id,omitempty"`
	PID           int      `json:"pid,omitempty"`
	GPUIndex      *int     `json:"gpu_index,omitempty"`
	VRAMBytes     uint64   `json:"vram_bytes,omitempty"`
	Models        []string `json:"models,omitempty"`
	HealthOK      *bool    `json:"health_ok,omitempty"`
}

// Finding holds a diagnostic finding from inspect or doctor.
type Finding struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	Severity string `json:"severity,omitempty"`
	Why      string `json:"why,omitempty"`
	Try      string `json:"try,omitempty"`
	Category string `json:"category,omitempty"`
	Rank     int    `json:"rank,omitempty"`
}

// Skip records a collector that could not run.
type Skip struct {
	Component string `json:"component"`
	Reason    string `json:"reason"`
}

// New returns a snapshot with CollectedAt set to now.
func New() *Snapshot {
	return &Snapshot{
		CollectedAt: time.Now().UTC(),
	}
}
