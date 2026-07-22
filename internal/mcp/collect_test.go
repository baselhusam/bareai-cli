package mcp

import (
	"testing"

	"github.com/baselhusam/bareai-cli/internal/collect"
)

func TestCollectOptionsLightVsFull(t *testing.T) {
	light := collect.LightRefreshOptions()
	full := collect.FullOptions()
	if light.ListModels || light.ProbeDB || light.DockerDetail {
		t.Fatal("light refresh should skip heavy work")
	}
	if !full.ListModels || !full.ProbeDB || !full.DockerDetail {
		t.Fatal("full options should enable all detail collectors")
	}
}
