package action

import (
	"context"
	"fmt"
	"time"
)

func runFreeGPU(ctx context.Context, api DockerAPI, req Request, target *Target, timeout time.Duration) Result {
	res := NewResult(req)
	res.Target = target
	if err := validateContainerTarget(target); err != nil {
		res.Error = err.Error()
		return res
	}

	before, err := readContainerState(ctx, api, target.ID)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.Before = before

	res.Would = fmt.Sprintf("Stop container %q, wait 2s, restart, reprobe %s", target.Name, target.Endpoint)
	if req.DryRun || req.PlanOnly {
		res.OK = true
		return res
	}

	stopStep := Step{Verb: VerbStop, Summary: fmt.Sprintf("stop %s", target.Name)}
	if err := api.ContainerStop(ctx, target.ID, timeout); err != nil {
		stopStep.OK = false
		stopStep.Error = err.Error()
		res.Steps = append(res.Steps, stopStep)
		res.Error = err.Error()
		return res
	}
	stopStep.OK = true
	res.Steps = append(res.Steps, stopStep)

	time.Sleep(2 * time.Second)

	restartStep := Step{Verb: VerbRestart, Summary: fmt.Sprintf("restart %s", target.Name)}
	if err := api.ContainerRestart(ctx, target.ID, timeout); err != nil {
		restartStep.OK = false
		restartStep.Error = err.Error()
		res.Steps = append(res.Steps, restartStep)
		res.Error = err.Error()
		return res
	}
	restartStep.OK = true
	res.Steps = append(res.Steps, restartStep)

	if req.AutoReprobe {
		step, _ := runReprobe(ctx, req, target)
		res.Steps = append(res.Steps, step)
	}

	after, err := readContainerState(ctx, api, target.ID)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.After = after
	res.OK = true
	return res
}
