//go:build !linux

package gpu

import (
	"context"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func collectAMD(_ context.Context) ([]snapshot.GPU, error) {
	return nil, nil
}
