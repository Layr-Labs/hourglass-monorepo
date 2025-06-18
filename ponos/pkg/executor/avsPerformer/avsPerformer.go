package avsPerformer

import (
	"context"
	"time"

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

// PerformerStatus represents the health status of a performer container
type PerformerStatus int

const (
	PerformerHealthy PerformerStatus = iota
	PerformerUnhealthy
)

// PerformerStatusEvent contains status information sent to deployment watchers
type PerformerStatusEvent struct {
	Status      PerformerStatus
	ContainerID string
	Message     string
	Timestamp   time.Time
}

type AvsPerformerConfig struct {
	AvsAddress           string
	ProcessType          AvsProcessType
	Image                PerformerImage
	WorkerCount          int
	PerformerNetworkName string
	SigningCurve         string // bn254, bls381, etc
}

type IAvsPerformer interface {
	Initialize(ctx context.Context) error
	RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error)
	ValidateTaskSignature(task *performerTask.PerformerTask) error
	DeployContainer(
		ctx context.Context,
		avsId string,
		image PerformerImage,
	) (<-chan PerformerStatusEvent, error)
	PromoteContainer(ctx context.Context) error
	CancelDeployment(ctx context.Context) error
	Shutdown() error
}
