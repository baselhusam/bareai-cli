//go:build linux

package gpu

import "testing"

func TestIsAMDVendor(t *testing.T) {
	tests := []struct {
		vendor string
		want   bool
	}{
		{"0x1002", true},
		{"4098", true},
		{"0x10de", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := isAMDVendor(tt.vendor); got != tt.want {
			t.Errorf("isAMDVendor(%q) = %v, want %v", tt.vendor, got, tt.want)
		}
	}
}
