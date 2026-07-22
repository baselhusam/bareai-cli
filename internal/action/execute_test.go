package action

import (
	"context"
	"testing"
	"time"
)

type mockDocker struct {
	logs string
}

func (m *mockDocker) Ping(ctx context.Context) error { return nil }

func (m *mockDocker) ContainerInspect(_ context.Context, id string) (containerState, error) {
	return containerState{ID: id, Name: "ollama", Image: "ollama/ollama", State: "running"}, nil
}

func (m *mockDocker) ContainerRestart(context.Context, string, time.Duration) error { return nil }

func (m *mockDocker) ContainerStop(context.Context, string, time.Duration) error { return nil }

func (m *mockDocker) ContainerLogs(context.Context, string, int) (string, error) {
	return m.logs, nil
}

func (m *mockDocker) Close() error { return nil }

func TestExecuteLogsDryRunPlan(t *testing.T) {
	snap := testSnap()
	target, err := ResolveTarget(snap, "llm.unreachable", VerbLogs, ResolveHints{Container: "ollama"})
	if err != nil {
		t.Fatal(err)
	}
	res := BuildPlan(Request{Verb: VerbLogs, FindingID: "llm.unreachable", DryRun: true, Tail: 50}, target)
	if !res.OK || res.Would == "" {
		t.Fatalf("unexpected plan: %+v", res)
	}
}

func TestExecuteLogsWithMock(t *testing.T) {
	ctx := context.Background()
	ex := &Executor{Docker: &mockDocker{logs: "hello\n"}}
	snap := testSnap()
	target, err := ResolveTarget(snap, "llm.unreachable", VerbLogs, ResolveHints{Container: "ollama"})
	if err != nil {
		t.Fatal(err)
	}
	_ = target
	res, err := ex.runLogs(ctx, Request{
		Verb:        VerbLogs,
		FindingID:   "llm.unreachable",
		Tail:        10,
		LogMaxBytes: 1024,
	}, target)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Output != "hello\n" {
		t.Fatalf("unexpected result: %+v", res)
	}
}
