package cli

import "testing"

func TestMCPCmdRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "mcp" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("missing mcp subcommand")
	}
}

func TestMCPCmdHelp(t *testing.T) {
	if mcpCmd.Short == "" {
		t.Fatal("expected mcp short help")
	}
	if mcpCmd.Long == "" {
		t.Fatal("expected mcp long help")
	}
}
