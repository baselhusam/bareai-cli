package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	if rootCmd.Use != "bareai" {
		t.Fatalf("expected Use bareai, got %q", rootCmd.Use)
	}

	expected := map[string]bool{
		"status":  false,
		"gpu":     false,
		"llm":     false,
		"docker":  false,
		"inspect": false,
		"probe":   false,
		"watch":   false,
	}

	for _, cmd := range rootCmd.Commands() {
		if _, ok := expected[cmd.Name()]; ok {
			expected[cmd.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

func TestStubCommands(t *testing.T) {
	tests := []struct {
		cmd  *cobra.Command
		name string
	}{
		{watchCmd, "watch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cmd.RunE(tt.cmd, nil); err != nil {
				t.Fatalf("stub command failed: %v", err)
			}
		})
	}
}
