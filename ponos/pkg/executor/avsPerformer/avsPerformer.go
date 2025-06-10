package avsPerformer

import (
	"context"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
)

type AvsProcessType string

const (
	AvsProcessTypeServer AvsProcessType = "server"
	AvsProcessTypeOneOff AvsProcessType = "one-off"
)

type PerformerImage struct {
	Repository string
	Tag        string
}

type AvsPerformerConfig struct {
	AvsAddress           string
	ProcessType          AvsProcessType
	Image                PerformerImage
	WorkerCount          int
	PerformerNetworkName string
	SigningCurve         string // bn254, bls381, etc
}

// PerformerStatus represents the current status of a performer
type PerformerStatus string

const (
	PerformerStatusUnknown   PerformerStatus = "unknown"
	PerformerStatusHealthy   PerformerStatus = "healthy"
	PerformerStatusUnhealthy PerformerStatus = "unhealthy"
	PerformerStatusStopped   PerformerStatus = "stopped"
)

// PerformerInfo contains runtime information about a performer
type PerformerInfo struct {
	ContainerID  string
	Status       PerformerStatus
	RestartCount uint32
}

type IAvsPerformer interface {
	Initialize(ctx context.Context) error
	RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error)
	ValidateTaskSignature(task *performerTask.PerformerTask) error
	Shutdown() error

	// Status and information methods
	GetPerformerInfo(ctx context.Context) (*PerformerInfo, error)
}
