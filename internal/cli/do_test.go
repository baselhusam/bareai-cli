package cli

import "testing"

func TestDoCmdRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "do" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("missing do subcommand")
	}
}

func TestDoPlanSubcommand(t *testing.T) {
	if doPlanCmd == nil {
		t.Fatal("missing do plan subcommand")
	}
}
