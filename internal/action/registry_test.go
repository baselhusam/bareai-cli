package action

import "testing"

func TestAllowedVerbs(t *testing.T) {
	tests := map[string][]string{
		"llm.unreachable":    {VerbLogs, VerbReprobe, VerbRestart},
		"llm.no_models":      {VerbReprobe},
		"db.unreachable":     {VerbLogs, VerbRestart},
		"gpu.vram_high":      {VerbLogs, VerbStop, VerbFreeGPU},
		"gpu.idle_while_llm": {VerbLogs, VerbStop, VerbFreeGPU},
		"host.mem_high":      nil,
	}
	for id, want := range tests {
		got := AllowedVerbs(id)
		if len(got) != len(want) {
			t.Fatalf("%s: got %v want %v", id, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: got %v want %v", id, got, want)
			}
		}
	}
}

func TestMutates(t *testing.T) {
	if !Mutates(VerbRestart) || !Mutates(VerbStop) || !Mutates(VerbFreeGPU) {
		t.Fatal("expected mutate verbs")
	}
	if Mutates(VerbLogs) || Mutates(VerbReprobe) {
		t.Fatal("expected read-only verbs")
	}
}
