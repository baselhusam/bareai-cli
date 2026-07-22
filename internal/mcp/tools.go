package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/baselhusam/bareai-cli/internal/config"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

type timeoutArgs struct {
	TimeoutSeconds int `json:"timeout_seconds,omitempty" jsonschema:"optional timeout in seconds (max 120; default from config)"`
}

type snapshotArgs struct {
	Light          bool `json:"light,omitempty" jsonschema:"use lighter refresh (skip model listing, docker detail, db probes)"`
	TimeoutSeconds int  `json:"timeout_seconds,omitempty" jsonschema:"optional timeout in seconds (max 120)"`
}

type llmsArgs struct {
	ListModels     *bool `json:"list_models,omitempty" jsonschema:"fetch model lists (default true)"`
	TimeoutSeconds int   `json:"timeout_seconds,omitempty" jsonschema:"optional timeout in seconds (max 120)"`
}

type doctorArgs struct {
	MinSeverity    string `json:"min_severity,omitempty" jsonschema:"minimum severity: info, warn, or critical (default info)"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"optional timeout in seconds (max 120)"`
}

type probeArgs struct {
	Endpoint       string `json:"endpoint,omitempty" jsonschema:"probe a specific endpoint URL; omit to probe all discovered LLMs"`
	Runtime        string `json:"runtime,omitempty" jsonschema:"runtime when using endpoint: ollama, vllm, sglang, triton"`
	Model          string `json:"model,omitempty" jsonschema:"model for smoke request (default from config)"`
	Prompt         string `json:"prompt,omitempty" jsonschema:"prompt for smoke request (default from config)"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"optional timeout in seconds (max 120)"`
}

type llmsData struct {
	LLMs    []snapshot.LLM  `json:"llms"`
	Skipped []snapshot.Skip `json:"skipped,omitempty"`
}

type databasesData struct {
	Databases []snapshot.Database `json:"databases"`
	Skipped   []snapshot.Skip     `json:"skipped,omitempty"`
}

type correlationsData struct {
	Correlations []snapshot.Correlation `json:"correlations"`
	Skipped      []snapshot.Skip        `json:"skipped,omitempty"`
}

type doctorData struct {
	Findings []snapshot.Finding `json:"findings"`
	Counts   map[string]int     `json:"counts"`
	Skipped  []snapshot.Skip    `json:"skipped,omitempty"`
}

type probeData struct {
	LLMs    []snapshot.LLM  `json:"llms"`
	Skipped []snapshot.Skip `json:"skipped,omitempty"`
}

// RegisterTools adds bareai MCP tools to the server.
func RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bareai_snapshot",
		Description: "Full enriched infrastructure snapshot (host, GPU, Docker, LLM, DB, correlations). Default first call to learn what is on this box.",
	}, handleSnapshot)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bareai_correlations",
		Description: "Lightweight correlation join: model/engine → container → GPU/VRAM or DB address → health.",
	}, handleCorrelations)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bareai_llms",
		Description: "Discovered LLM runtimes with health, models, and metrics.",
	}, handleLLMs)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bareai_databases",
		Description: "Discovered local databases (Postgres, Redis, MongoDB, MySQL, Qdrant, Elasticsearch) with health probes.",
	}, handleDatabases)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bareai_doctor",
		Description: "Ranked diagnostics with what/why/try hints for this box.",
	}, handleDoctor)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bareai_probe",
		Description: "One-hit smoke test against discovered or explicit LLM endpoints.",
	}, handleProbe)
}

func handleSnapshot(ctx context.Context, _ *mcp.CallToolRequest, in snapshotArgs) (*mcp.CallToolResult, any, error) {
	ctx, cancel := WithToolTimeout(ctx, in.TimeoutSeconds)
	defer cancel()

	snap := CollectEnriched(ctx, in.Light)
	return textResult(snap, snap)
}

func handleCorrelations(ctx context.Context, _ *mcp.CallToolRequest, in timeoutArgs) (*mcp.CallToolResult, any, error) {
	ctx, cancel := WithToolTimeout(ctx, in.TimeoutSeconds)
	defer cancel()

	snap := CollectEnriched(ctx, true)
	data := correlationsData{
		Correlations: snap.Correlations,
		Skipped:      snap.Skipped,
	}
	return textResult(snap, data)
}

func handleLLMs(ctx context.Context, _ *mcp.CallToolRequest, in llmsArgs) (*mcp.CallToolResult, any, error) {
	ctx, cancel := WithToolTimeout(ctx, in.TimeoutSeconds)
	defer cancel()

	listModels := true
	if in.ListModels != nil {
		listModels = *in.ListModels
	}

	snap := CollectLLMs(ctx, listModels)
	data := llmsData{LLMs: snap.LLMs, Skipped: snap.Skipped}
	return textResult(snap, data)
}

func handleDatabases(ctx context.Context, _ *mcp.CallToolRequest, in timeoutArgs) (*mcp.CallToolResult, any, error) {
	ctx, cancel := WithToolTimeout(ctx, in.TimeoutSeconds)
	defer cancel()

	snap := CollectDatabases(ctx)
	data := databasesData{Databases: snap.Databases, Skipped: snap.Skipped}
	return textResult(snap, data)
}

func handleDoctor(ctx context.Context, _ *mcp.CallToolRequest, in doctorArgs) (*mcp.CallToolResult, any, error) {
	ctx, cancel := WithToolTimeout(ctx, in.TimeoutSeconds)
	defer cancel()

	sev := in.MinSeverity
	if sev == "" {
		sev = config.Global().Doctor.MinSeverity
	}
	if sev == "" {
		sev = config.Default().Doctor.MinSeverity
	}

	snap := CollectDoctor(ctx, sev)
	data := doctorData{
		Findings: snap.Findings,
		Counts:   countSeverities(snap.Findings),
		Skipped:  snap.Skipped,
	}
	return textResult(snap, data)
}

func handleProbe(ctx context.Context, _ *mcp.CallToolRequest, in probeArgs) (*mcp.CallToolResult, any, error) {
	ctx, cancel := WithToolTimeout(ctx, in.TimeoutSeconds)
	defer cancel()

	cfg := config.Global()
	model := in.Model
	if model == "" {
		model = cfg.Probe.Model
	}
	prompt := in.Prompt
	if prompt == "" {
		prompt = cfg.Probe.Prompt
	}

	snap := RunProbeSnapshot(ctx, ProbeOptions{
		Endpoint: in.Endpoint,
		Runtime:  in.Runtime,
		Model:    model,
		Prompt:   prompt,
	})
	data := probeData{LLMs: snap.LLMs, Skipped: snap.Skipped}
	return textResult(snap, data)
}

func textResult(snap *snapshot.Snapshot, data any) (*mcp.CallToolResult, any, error) {
	collectedAt := snap.CollectedAt
	text, err := MarshalJSON(data, collectedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}

func countSeverities(findings []snapshot.Finding) map[string]int {
	counts := map[string]int{}
	for _, f := range findings {
		sev := f.Severity
		if sev == "" {
			sev = "info"
		}
		counts[sev]++
	}
	return counts
}
