package collect

import (
	"context"
	"testing"
)

func BenchmarkSnapshotWithOptions(b *testing.B) {
	ctx := context.Background()
	opts := FullOptions()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = SnapshotWithOptions(ctx, opts)
	}
}
