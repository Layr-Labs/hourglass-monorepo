package performerPoolManager

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/performerCapacityPlanner"
)

// IPerformerPoolManager defines the interface for managing performer pools
type IPerformerPoolManager interface {
	// Initialize initializes the pool manager and all AVS performer pools
	Initialize() error

	// Start starts all performer pools and begins lifecycle management
	Start(ctx context.Context) error

	// GetPerformer returns a healthy performer for the given AVS address
	GetPerformer(avsAddress string) (avsPerformer.IAvsPerformer, bool)
}

// IPerformerPool defines the interface for a pool of performers for a specific AVS
type IPerformerPool interface {
	// ExecutePlan executes a capacity plan for this AVS
	ExecutePlan(ctx context.Context, plan *performerCapacityPlanner.PerformerCapacityPlan) error

	// CheckHealth checks health of all performers in the pool and returns the count of healthy performers
	CheckHealth(ctx context.Context) (int, error)

	// Shutdown stops and removes all performers in the pool
	Shutdown() error

	// GetHealthyPerformer returns a healthy performer from the pool
	GetHealthyPerformer() (avsPerformer.IAvsPerformer, bool)

	// GetPerformerCount returns the total number of performers in the pool
	GetPerformerCount() int

	// GetAvsAddress returns the AVS address associated with this pool
	GetAvsAddress() string
}
