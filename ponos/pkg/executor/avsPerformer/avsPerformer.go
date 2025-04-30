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

type IAvsPerformer interface {
	Initialize(ctx context.Context) error
	ProcessTasks(ctx context.Context) error
	RunTask(ctx context.Context, task *performerTask.PerformerTask) error
	ValidateTaskSignature(task *performerTask.PerformerTask) error
	Shutdown() error
}

type ReceiveTaskResponse func(originalTask *performerTask.PerformerTask, response *performerTask.PerformerTaskResult, err error)
