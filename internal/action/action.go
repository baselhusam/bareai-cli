package action

import (
	"time"

	"github.com/baselhusam/bareai-cli/internal/version"
)

// SchemaVersion is the action result JSON contract version.
const SchemaVersion = "1.0"

// Verb names for bareai do.
const (
	VerbLogs     = "logs"
	VerbReprobe  = "reprobe"
	VerbRestart  = "restart"
	VerbStop     = "stop"
	VerbFreeGPU  = "free-gpu"
)

// TargetKind values.
const (
	TargetContainer = "container"
	TargetEndpoint  = "endpoint"
)

// Request describes an action invocation.
type Request struct {
	Verb        string
	FindingID   string
	Container   string
	Endpoint    string
	Runtime     string
	Tail        int
	DryRun      bool
	PlanOnly    bool
	Confirmed   bool
	AutoReprobe bool
	LogMaxBytes int
}

// Target is a resolved action subject from the snapshot.
type Target struct {
	Kind          string `json:"kind"`
	ID            string `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	Endpoint      string `json:"endpoint,omitempty"`
	Runtime       string `json:"runtime,omitempty"`
	Image         string `json:"image,omitempty"`
	State         string `json:"state,omitempty"`
	Status        string `json:"status,omitempty"`
	GPURequested  bool   `json:"gpu_requested,omitempty"`
	GPUIndex      *int   `json:"gpu_index,omitempty"`
}

// ContainerState captures container fields for before/after audit.
type ContainerState struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Image  string `json:"image,omitempty"`
	State  string `json:"state,omitempty"`
	Status string `json:"status,omitempty"`
}

// Step records a sub-step in compound actions (e.g. free-gpu).
type Step struct {
	Verb    string `json:"verb"`
	OK      bool   `json:"ok"`
	Summary string `json:"summary,omitempty"`
	Error   string `json:"error,omitempty"`
	ProbeOK *bool  `json:"probe_ok,omitempty"`
}

// Result is the audit-friendly action outcome.
type Result struct {
	SchemaVersion string            `json:"schema_version"`
	BareaiVersion string            `json:"bareai_version"`
	ExecutedAt    time.Time         `json:"executed_at"`
	Verb          string            `json:"verb"`
	FindingID     string            `json:"finding_id,omitempty"`
	Target        *Target           `json:"target,omitempty"`
	DryRun        bool              `json:"dry_run"`
	Confirmed     bool              `json:"confirmed"`
	Before        *ContainerState   `json:"before,omitempty"`
	After         *ContainerState   `json:"after,omitempty"`
	Would         string            `json:"would,omitempty"`
	Output        string            `json:"output,omitempty"`
	Steps         []Step            `json:"steps,omitempty"`
	OK            bool              `json:"ok"`
	Error         string            `json:"error,omitempty"`
}

// ListEntry describes an available action from current findings.
type ListEntry struct {
	Verb       string `json:"verb"`
	FindingID  string `json:"finding_id"`
	TargetKind string `json:"target_kind"`
	TargetRef  string `json:"target_ref"`
	Summary    string `json:"summary,omitempty"`
	Command    string `json:"command"`
}

// NewResult builds a result shell with version metadata.
func NewResult(req Request) Result {
	return Result{
		SchemaVersion: SchemaVersion,
		BareaiVersion: version.Version,
		ExecutedAt:    time.Now().UTC(),
		Verb:          req.Verb,
		FindingID:     req.FindingID,
		DryRun:        req.DryRun || req.PlanOnly,
		Confirmed:     req.Confirmed,
	}
}
