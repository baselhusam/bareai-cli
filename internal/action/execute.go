package action

import (
	"context"
	"fmt"
	"time"

	"github.com/baselhusam/bareai-cli/internal/collect"
	"github.com/baselhusam/bareai-cli/internal/doctor"
	"github.com/baselhusam/bareai-cli/internal/inspect"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// Executor runs bareai do actions.
type Executor struct {
	Docker DockerAPI
}

// Run executes or plans an action against a fresh snapshot.
func (e *Executor) Run(ctx context.Context, req Request) (Result, error) {
	if req.Verb == "" {
		return Result{}, fmt.Errorf("verb is required")
	}
	if Mutates(req.Verb) && req.FindingID == "" {
		return Result{}, fmt.Errorf("--finding is required for %q", req.Verb)
	}
	if req.Verb == VerbLogs && req.FindingID == "" {
		return Result{}, fmt.Errorf("--finding is required for logs")
	}
	if req.Verb == VerbReprobe && req.FindingID == "" {
		return Result{}, fmt.Errorf("--finding is required for reprobe")
	}

	snap := collect.SnapshotWithOptions(ctx, collect.LightRefreshOptions())
	inspect.Enrich(snap)
	snap.Findings = doctor.Analyze(snap, doctor.Options{MinSeverity: "info"})

	target, err := ResolveTarget(snap, req.FindingID, req.Verb, ResolveHints{
		Container: req.Container,
		Endpoint:  req.Endpoint,
	})
	if err != nil {
		return Result{OK: false, Error: err.Error(), Verb: req.Verb, FindingID: req.FindingID}, err
	}

	if req.PlanOnly {
		req.DryRun = true
	}
	plan := BuildPlan(req, target)
	if req.DryRun || req.PlanOnly {
		return plan, nil
	}

	if Mutates(req.Verb) && !req.Confirmed {
		return plan, fmt.Errorf("confirmation required: re-run with --yes or use bareai do plan")
	}

	switch req.Verb {
	case VerbLogs:
		return e.runLogs(ctx, req, target)
	case VerbReprobe:
		return e.runReprobe(ctx, req, target)
	case VerbRestart:
		return e.runRestart(ctx, req, target)
	case VerbStop:
		return e.runStop(ctx, req, target)
	case VerbFreeGPU:
		return e.runFreeGPUAction(ctx, req, target)
	default:
		return Result{}, fmt.Errorf("unknown verb %q", req.Verb)
	}
}

func (e *Executor) docker(ctx context.Context) (DockerAPI, error) {
	if e.Docker != nil {
		if err := e.Docker.Ping(ctx); err != nil {
			return nil, fmt.Errorf("docker unavailable: %w", err)
		}
		return e.Docker, nil
	}
	api, err := NewDockerAPI()
	if err != nil {
		return nil, err
	}
	if err := api.Ping(ctx); err != nil {
		api.Close()
		return nil, fmt.Errorf("docker unavailable: %w", err)
	}
	return api, nil
}

func (e *Executor) runLogs(ctx context.Context, req Request, target *Target) (Result, error) {
	res := NewResult(req)
	res.Target = target
	if err := validateContainerTarget(target); err != nil {
		res.Error = err.Error()
		return res, nil
	}
	api, err := e.docker(ctx)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	if e.Docker == nil {
		defer api.Close()
	}
	before, err := readContainerState(ctx, api, target.ID)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	res.Before = before
	out, err := api.ContainerLogs(ctx, target.ID, req.Tail)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	res.Output = truncateOutput(out, req.LogMaxBytes)
	res.OK = true
	return res, nil
}

func (e *Executor) runReprobe(ctx context.Context, req Request, target *Target) (Result, error) {
	res := NewResult(req)
	res.Target = target
	step, _ := runReprobe(ctx, req, target)
	res.Steps = []Step{step}
	res.OK = step.OK
	if !step.OK && step.Error != "" {
		res.Error = step.Error
	}
	return res, nil
}

func (e *Executor) runRestart(ctx context.Context, req Request, target *Target) (Result, error) {
	res := NewResult(req)
	res.Target = target
	if err := validateContainerTarget(target); err != nil {
		res.Error = err.Error()
		return res, nil
	}
	api, err := e.docker(ctx)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	if e.Docker == nil {
		defer api.Close()
	}
	timeout := 10 * time.Second
	before, err := readContainerState(ctx, api, target.ID)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	res.Before = before
	if err := api.ContainerRestart(ctx, target.ID, timeout); err != nil {
		res.Error = err.Error()
		return res, nil
	}
	after, err := readContainerState(ctx, api, target.ID)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	res.After = after
	if req.AutoReprobe && target.Endpoint != "" {
		step, _ := runReprobe(ctx, req, target)
		res.Steps = append(res.Steps, step)
	}
	res.OK = true
	return res, nil
}

func (e *Executor) runStop(ctx context.Context, req Request, target *Target) (Result, error) {
	res := NewResult(req)
	res.Target = target
	if err := validateContainerTarget(target); err != nil {
		res.Error = err.Error()
		return res, nil
	}
	api, err := e.docker(ctx)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	if e.Docker == nil {
		defer api.Close()
	}
	timeout := 10 * time.Second
	before, err := readContainerState(ctx, api, target.ID)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	res.Before = before
	if err := api.ContainerStop(ctx, target.ID, timeout); err != nil {
		res.Error = err.Error()
		return res, nil
	}
	after, err := readContainerState(ctx, api, target.ID)
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	res.After = after
	res.OK = true
	return res, nil
}

func (e *Executor) runFreeGPUAction(ctx context.Context, req Request, target *Target) (Result, error) {
	api, err := e.docker(ctx)
	if err != nil {
		return Result{OK: false, Error: err.Error(), Verb: req.Verb, FindingID: req.FindingID}, nil
	}
	if e.Docker == nil {
		defer api.Close()
	}
	return runFreeGPU(ctx, api, req, target, 10*time.Second), nil
}

// CollectSnapshotForList returns an enriched snapshot with doctor findings.
func CollectSnapshotForList(ctx context.Context) *snapshot.Snapshot {
	snap := collect.SnapshotWithOptions(ctx, collect.LightRefreshOptions())
	inspect.Enrich(snap)
	snap.Findings = doctor.Analyze(snap, doctor.Options{MinSeverity: "info"})
	return snap
}
