//go:build !darwin

package gpu

import (
	"context"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func collectApple(_ context.Context) ([]snapshot.GPU, error) {
	return nil, nil
}
