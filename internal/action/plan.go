package action

import "fmt"

// BuildPlan fills preview fields on a result before execution.
func BuildPlan(req Request, target *Target) Result {
	res := NewResult(req)
	res.Target = target
	if target == nil {
		return res
	}

	switch req.Verb {
	case VerbLogs:
		res.Would = fmt.Sprintf("Fetch last %d log lines from container %q (%s)", maxTail(req.Tail), target.Name, target.Image)
	case VerbReprobe:
		res.Would = fmt.Sprintf("Run smoke probe against %s (%s)", target.Endpoint, target.Runtime)
	case VerbRestart:
		res.Would = fmt.Sprintf("Restart container %q (%s) currently %s", target.Name, target.Image, target.State)
	case VerbStop:
		res.Would = fmt.Sprintf("Stop container %q (%s) currently %s", target.Name, target.Image, target.State)
	case VerbFreeGPU:
		res.Would = fmt.Sprintf("Stop container %q (%s), wait, restart, and reprobe %s to free GPU memory",
			target.Name, target.Image, target.Endpoint)
	}
	if target.Kind == TargetContainer {
		res.Before = &ContainerState{
			ID:     target.ID,
			Name:   target.Name,
			Image:  target.Image,
			State:  target.State,
			Status: target.Status,
		}
	}
	res.OK = true
	return res
}

func maxTail(tail int) int {
	if tail <= 0 {
		return 100
	}
	return tail
}
