package dao

import (
	"context"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
)

// SpecDAO provides access to runtime specifications
type SpecDAO interface {
	// GetRuntimeSpec retrieves a specific runtime spec by release ID
	GetRuntimeSpec(ctx context.Context, releaseID string) (*runtime.Spec, error)

	// GetLatestRuntimeSpec retrieves the latest runtime spec
	GetLatestRuntimeSpec(ctx context.Context) (*runtime.Spec, error)
}
