package render

import "runtime"

// EmptyHint returns actionable guidance when a collector pane is empty.
func EmptyHint(component string) string {
	switch component {
	case "gpu":
		if runtime.GOOS == "darwin" {
			return "no accelerators detected — on Mac expect Apple GPU identity; unified memory, no util/temp via public APIs"
		}
		return "no accelerators detected — on Linux install drivers and check nvidia-smi; AMD uses sysfs/ROCm"
	case "llm":
		return "none discovered — start Ollama (:11434) or an OpenAI-compatible server; Docker image name heuristics also apply"
	case "docker":
		return "Docker not available — start Docker Desktop / daemon; set DOCKER_HOST if needed"
	case "db":
		return "none discovered — local Postgres/Redis/Mongo/MySQL/Qdrant/ES on default ports or in Docker"
	case "correlation":
		return "none yet — run an LLM or DB locally and press r to refresh"
	default:
		return "nothing here yet"
	}
}
