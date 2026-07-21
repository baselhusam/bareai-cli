package host

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCollect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	host, err := Collect(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skip("host info unavailable in restricted environment")
		}
		t.Fatalf("Collect failed: %v", err)
	}

	if host.Hostname == "" {
		t.Error("expected non-empty hostname")
	}
	if host.OS == "" {
		t.Error("expected non-empty OS")
	}
	if host.Arch == "" {
		t.Error("expected non-empty arch")
	}
	if host.MemTotal == 0 {
		t.Error("expected MemTotal > 0")
	}
	if host.CPUCores <= 0 {
		t.Error("expected CPUCores > 0")
	}
}

func TestSkipMount(t *testing.T) {
	tests := []struct {
		mount string
		fstype string
		want   bool
	}{
		{"/proc", "proc", true},
		{"/sys", "sysfs", true},
		{"/", "apfs", false},
		{"/Volumes/Data", "apfs", false},
	}

	for _, tt := range tests {
		got := skipMount(tt.mount, tt.fstype)
		if got != tt.want {
			t.Errorf("skipMount(%q, %q) = %v, want %v", tt.mount, tt.fstype, got, tt.want)
		}
	}
}
