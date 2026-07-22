package collect

import (
	"context"
	"testing"
)

func BenchmarkSnapshotFull(b *testing.B) {
	ctx := context.Background()
	opts := FullOptions()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = SnapshotWithOptions(ctx, opts)
	}
}

func BenchmarkSnapshotLightRefresh(b *testing.B) {
	ctx := context.Background()
	opts := LightRefreshOptions()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = SnapshotWithOptions(ctx, opts)
	}
}

func TestLightRefreshOptionsSkipsHeavyWork(t *testing.T) {
	light := LightRefreshOptions()
	full := FullOptions()
	if light.ListModels {
		t.Fatal("light refresh should skip ListModels")
	}
	if light.ProbeDB {
		t.Fatal("light refresh should skip ProbeDB")
	}
	if light.DockerDetail {
		t.Fatal("light refresh should skip DockerDetail")
	}
	if !full.ListModels || !full.ProbeDB || !full.DockerDetail {
		t.Fatal("full options should enable all detail collectors")
	}
}
