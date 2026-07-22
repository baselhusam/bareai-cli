package mcp

import (
	"encoding/json"
	"time"

	"github.com/baselhusam/bareai-cli/internal/version"
)

// SchemaVersion is the agent contract version for MCP tool responses.
const SchemaVersion = "1.0"

// AgentEnvelope wraps tool payloads with version metadata.
type AgentEnvelope struct {
	SchemaVersion string    `json:"schema_version"`
	BareaiVersion string    `json:"bareai_version"`
	CollectedAt   time.Time `json:"collected_at"`
	Data          any       `json:"data"`
}

// Wrap builds an envelope around data using the current bareai version.
func Wrap(data any, collectedAt time.Time) AgentEnvelope {
	if collectedAt.IsZero() {
		collectedAt = time.Now().UTC()
	}
	return AgentEnvelope{
		SchemaVersion: SchemaVersion,
		BareaiVersion: version.Version,
		CollectedAt:   collectedAt,
		Data:          data,
	}
}

// MarshalJSON returns the envelope as compact JSON text for MCP tool results.
func MarshalJSON(data any, collectedAt time.Time) (string, error) {
	b, err := json.Marshal(Wrap(data, collectedAt))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
