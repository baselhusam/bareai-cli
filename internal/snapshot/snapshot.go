package snapshot

import "time"

// Snapshot is the top-level correlated infrastructure view.
type Snapshot struct {
	CollectedAt time.Time `json:"collected_at"`
	Host        *Host     `json:"host,omitempty"`
	GPUs        []GPU     `json:"gpus,omitempty"`
	Docker      *Docker   `json:"docker,omitempty"`
	LLMs        []LLM     `json:"llms,omitempty"`
	Findings    []Finding `json:"findings,omitempty"`
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

// GPU is filled in Phase 2.
type GPU struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

// Docker is filled in Phase 3.
type Docker struct {
	Available bool `json:"available"`
}

// LLM is filled in Phase 4.
type LLM struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

// Finding is filled in Phase 9.
type Finding struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
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
