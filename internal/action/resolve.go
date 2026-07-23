package action

import (
	"fmt"
	"strings"

	"github.com/baselhusam/bareai-cli/internal/doctor"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func filterDockerLLMs(llms []snapshot.LLM) []snapshot.LLM {
	var out []snapshot.LLM
	for _, llm := range llms {
		if llm.ContainerID != "" || llm.ContainerName != "" {
			out = append(out, llm)
		}
	}
	return out
}

// ResolveHints carries optional user overrides for target resolution.
type ResolveHints struct {
	Container string
	Endpoint  string
}

// ResolveTarget maps a finding and snapshot to an action target.
func ResolveTarget(snap *snapshot.Snapshot, findingID, verb string, hints ResolveHints) (*Target, error) {
	if snap == nil {
		return nil, fmt.Errorf("snapshot is nil")
	}
	if !AllowsVerb(findingID, verb) {
		return nil, fmt.Errorf("verb %q is not allowed for finding %q", verb, findingID)
	}
	if !findingPresent(snap, findingID) {
		return nil, fmt.Errorf("finding %q is not present in current snapshot", findingID)
	}

	switch {
	case strings.HasPrefix(findingID, "llm."):
		return resolveLLMTarget(snap, findingID, verb, hints)
	case strings.HasPrefix(findingID, "db."):
		return resolveDBTarget(snap, hints)
	case strings.HasPrefix(findingID, "gpu."):
		return resolveGPUTarget(snap, hints)
	case findingID == "docker.unavailable":
		return resolveDockerUnavailableTarget(snap, hints)
	default:
		return nil, fmt.Errorf("finding %q has no resolvable target", findingID)
	}
}

func findingPresent(snap *snapshot.Snapshot, findingID string) bool {
	findings := doctor.Analyze(snap, doctor.Options{MinSeverity: "info"})
	for _, f := range findings {
		if f.ID == findingID {
			return true
		}
	}
	return false
}

func resolveLLMTarget(snap *snapshot.Snapshot, findingID, verb string, hints ResolveHints) (*Target, error) {
	if verb == VerbReprobe {
		if hints.Endpoint != "" {
			return endpointTarget(hints.Endpoint, ""), nil
		}
		llms := matchingLLMs(snap, hints.Container, func(llm snapshot.LLM) bool {
			if findingID == "llm.no_models" {
				return llm.Health != nil && llm.Health.OK
			}
			return llm.Health == nil || !llm.Health.OK
		})
		if len(llms) == 0 {
			return nil, fmt.Errorf("no LLM matches finding %q", findingID)
		}
		if len(llms) > 1 {
			return nil, fmt.Errorf("multiple LLMs match finding %q; pass --endpoint or --container", findingID)
		}
		return llmEndpointTarget(llms[0]), nil
	}

	llms := matchingLLMs(snap, hints.Container, func(llm snapshot.LLM) bool {
		if findingID == "llm.no_models" {
			return llm.Health != nil && llm.Health.OK
		}
		return llm.Health == nil || !llm.Health.OK
	})
	if verb != VerbReprobe {
		llms = filterDockerLLMs(llms)
	}
	if len(llms) == 0 {
		return nil, fmt.Errorf("no docker LLM matches finding %q", findingID)
	}
	if len(llms) > 1 {
		return nil, fmt.Errorf("multiple LLMs match finding %q; pass --container", findingID)
	}
	return llmContainerTarget(snap, llms[0])
}

func resolveDBTarget(snap *snapshot.Snapshot, hints ResolveHints) (*Target, error) {
	var matches []snapshot.Database
	for _, db := range snap.Databases {
		if db.Health != nil && db.Health.OK {
			continue
		}
		if hints.Container != "" && !containerMatches(hints.Container, db.ContainerID, db.ContainerName) {
			continue
		}
		if db.ContainerID == "" && db.ContainerName == "" {
			continue
		}
		matches = append(matches, db)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no docker database matches finding db.unreachable")
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple databases match; pass --container")
	}
	db := matches[0]
	return containerTargetFromSnap(snap, db.ContainerID, db.ContainerName)
}

func resolveGPUTarget(snap *snapshot.Snapshot, hints ResolveHints) (*Target, error) {
	if hints.Container != "" {
		for _, llm := range snap.LLMs {
			if containerMatches(hints.Container, llm.ContainerID, llm.ContainerName) {
				return llmContainerTarget(snap, llm)
			}
		}
		for _, db := range snap.Databases {
			if containerMatches(hints.Container, db.ContainerID, db.ContainerName) {
				return containerTargetFromSnap(snap, db.ContainerID, db.ContainerName)
			}
		}
		return nil, fmt.Errorf("container %q not found in snapshot", hints.Container)
	}

	var candidates []snapshot.LLM
	for _, llm := range snap.LLMs {
		if llm.ContainerID == "" && llm.ContainerName == "" {
			continue
		}
		if llm.GPUIndex != nil {
			candidates = append(candidates, llm)
		}
	}
	if len(candidates) == 1 {
		return llmContainerTarget(snap, candidates[0])
	}
	if len(candidates) > 1 {
		return nil, fmt.Errorf("multiple GPU-backed containers; pass --container")
	}
	return nil, fmt.Errorf("no docker container correlated with GPU finding")
}

func resolveDockerUnavailableTarget(snap *snapshot.Snapshot, hints ResolveHints) (*Target, error) {
	if hints.Container == "" {
		return nil, fmt.Errorf("docker.unavailable requires --container")
	}
	for _, llm := range snap.LLMs {
		if containerMatches(hints.Container, llm.ContainerID, llm.ContainerName) {
			return llmContainerTarget(snap, llm)
		}
	}
	for _, db := range snap.Databases {
		if containerMatches(hints.Container, db.ContainerID, db.ContainerName) {
			return containerTargetFromSnap(snap, db.ContainerID, db.ContainerName)
		}
	}
	return nil, fmt.Errorf("container %q not found in snapshot", hints.Container)
}

func matchingLLMs(snap *snapshot.Snapshot, hint string, match func(snapshot.LLM) bool) []snapshot.LLM {
	var out []snapshot.LLM
	for _, llm := range snap.LLMs {
		if !match(llm) {
			continue
		}
		if hint != "" && !containerMatches(hint, llm.ContainerID, llm.ContainerName) {
			continue
		}
		out = append(out, llm)
	}
	return out
}

func llmEndpointTarget(llm snapshot.LLM) *Target {
	return &Target{
		Kind:     TargetEndpoint,
		Endpoint: llm.Endpoint,
		Runtime:  llm.Runtime,
		Name:     llm.Name,
	}
}

func endpointTarget(endpoint, runtime string) *Target {
	return &Target{
		Kind:     TargetEndpoint,
		Endpoint: endpoint,
		Runtime:  runtime,
		Name:     runtime,
	}
}

func llmContainerTarget(snap *snapshot.Snapshot, llm snapshot.LLM) (*Target, error) {
	if llm.ContainerID == "" && llm.ContainerName == "" {
		return nil, fmt.Errorf("LLM %s has no docker container (host-process runtime)", llm.Endpoint)
	}
	t, err := containerTargetFromSnap(snap, llm.ContainerID, llm.ContainerName)
	if err != nil {
		return nil, err
	}
	t.Endpoint = llm.Endpoint
	t.Runtime = llm.Runtime
	t.GPUIndex = llm.GPUIndex
	return t, nil
}

func containerTargetFromSnap(snap *snapshot.Snapshot, id, name string) (*Target, error) {
	if snap.Docker == nil || !snap.Docker.Available {
		return nil, fmt.Errorf("docker is unavailable")
	}
	var matches []snapshot.DockerContainer
	for _, c := range snap.Docker.Containers {
		if containerMatches(nameOrID(id, name), c.ID, c.Name) {
			matches = append(matches, c)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("container %q not found in snapshot", nameOrID(id, name))
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple containers match %q", nameOrID(id, name))
	}
	c := matches[0]
	return &Target{
		Kind:         TargetContainer,
		ID:           c.ID,
		Name:         c.Name,
		Image:        c.Image,
		State:        c.State,
		Status:       c.Status,
		GPURequested: c.GPURequested,
	}, nil
}

func containerMatches(hint, id, name string) bool {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return true
	}
	if name != "" && (hint == name || strings.Contains(name, hint)) {
		return true
	}
	if id != "" && (hint == id || strings.HasPrefix(id, hint)) {
		return true
	}
	return false
}

func nameOrID(id, name string) string {
	if name != "" {
		return name
	}
	return id
}

// ListAvailable builds actionable entries from doctor findings.
func ListAvailable(snap *snapshot.Snapshot) []ListEntry {
	if snap == nil {
		return nil
	}
	findings := doctor.Analyze(snap, doctor.Options{MinSeverity: "info"})
	var out []ListEntry
	seen := map[string]struct{}{}
	for _, f := range findings {
		for _, offer := range f.Do {
			key := f.ID + "\x00" + offer.Verb + "\x00" + offer.TargetRef
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, ListEntry{
				Verb:       offer.Verb,
				FindingID:  f.ID,
				TargetKind: offer.TargetKind,
				TargetRef:  offer.TargetRef,
				Summary:    offer.Summary,
				Command:    PlanCommand(offer.Verb, f.ID, offer.TargetKind, offer.TargetRef),
			})
		}
	}
	return out
}

// PlanCommand formats a suggested bareai do plan command.
func PlanCommand(verb, findingID, targetKind, targetRef string) string {
	cmd := fmt.Sprintf("bareai do plan %s --finding %s", verb, findingID)
	switch targetKind {
	case TargetContainer:
		cmd += fmt.Sprintf(" --container %q", targetRef)
	case TargetEndpoint:
		cmd += fmt.Sprintf(" --endpoint %q", targetRef)
	}
	return cmd
}
