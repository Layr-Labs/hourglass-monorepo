package avsPerformer

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
)

type AvsProcessType string

const (
	AvsProcessTypeServer AvsProcessType = "server"
)

type PerformerImage struct {
	RegistryUrl string
	// TODO: remove repository once unused by config, devkit, templates, executor, etc.
	Repository string
	Tag        string
	Digest     string
}

type AvsPerformerConfig struct {
	AvsAddress           string
	ProcessType          AvsProcessType
	Image                PerformerImage
	WorkerCount          int
	PerformerNetworkName string
	SigningCurve         string // bn254, bls381, etc
}

// IAvsPerformer defines the interface for AVS performers
type IAvsPerformer interface {
	// Initialize sets up the performer (creates containers, establishes connections, etc.)
	Initialize(ctx context.Context) error

	// Start begins any background processes (like artifact monitoring)
	Start(ctx context.Context)

	// ValidateTaskSignature validates the signature of a task
	ValidateTaskSignature(t *performerTask.PerformerTask) error

	// RunTask executes a task and returns the result
	RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error)

	// Shutdown gracefully stops the performer and cleans up resources
	Shutdown() error
}
