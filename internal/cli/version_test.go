package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	var buf bytes.Buffer
	versionCmd.SetOut(&buf)
	versionCmd.SetArgs(nil)

	if err := versionCmd.RunE(versionCmd, nil); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if !strings.Contains(buf.String(), "bareai") {
		t.Fatalf("expected version output, got %q", buf.String())
	}
}

func TestVersionJSON(t *testing.T) {
	prev := opts.JSON
	opts.JSON = true
	t.Cleanup(func() { opts.JSON = prev })

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)
	if err := versionCmd.RunE(versionCmd, nil); err != nil {
		t.Fatalf("version json failed: %v", err)
	}
	if !strings.Contains(buf.String(), `"version"`) {
		t.Fatalf("expected json version output, got %q", buf.String())
	}
}

func TestCompletionBash(t *testing.T) {
	var buf bytes.Buffer
	completionCmd.SetOut(&buf)
	if err := completionCmd.RunE(completionCmd, []string{"bash"}); err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}
	if !strings.Contains(buf.String(), "bash") {
		t.Fatalf("expected bash completion script, got %q", buf.String())
	}
}
