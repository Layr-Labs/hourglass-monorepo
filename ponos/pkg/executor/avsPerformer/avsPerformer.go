package avsPerformer

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performer"
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
	AvsAddress  string
	ProcessType AvsProcessType
	Image       PerformerImage
	WorkerCount int
}

type IAvsPerformer interface {
	Initialize(ctx context.Context) error
	ProcessTasks(ctx context.Context) error
	RunTask(ctx context.Context, task *performer.Task) error
	Shutdown() error
}

type ReceiveTaskResponse func(response *performer.TaskResult, err error)
