package action

// AllowedVerbs returns verbs permitted for a finding ID.
func AllowedVerbs(findingID string) []string {
	switch findingID {
	case "llm.unreachable":
		return []string{VerbLogs, VerbReprobe, VerbRestart}
	case "llm.no_models":
		return []string{VerbReprobe}
	case "db.unreachable":
		return []string{VerbLogs, VerbRestart}
	case "gpu.vram_high", "gpu.idle_while_llm":
		return []string{VerbLogs, VerbStop, VerbFreeGPU}
	case "docker.unavailable":
		return []string{VerbLogs}
	default:
		return nil
	}
}

// AllowsVerb reports whether verb is permitted for findingID.
func AllowsVerb(findingID, verb string) bool {
	for _, v := range AllowedVerbs(findingID) {
		if v == verb {
			return true
		}
	}
	return false
}

// Mutates reports whether the verb changes system state.
func Mutates(verb string) bool {
	switch verb {
	case VerbRestart, VerbStop, VerbFreeGPU:
		return true
	default:
		return false
	}
}
