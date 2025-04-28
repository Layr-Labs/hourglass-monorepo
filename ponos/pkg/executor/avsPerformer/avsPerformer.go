package avsPerformer

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/tasks"
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
}

type IAvsPerformer interface {
	Initialize(ctx context.Context) error
	ProcessTasks(ctx context.Context) error
	RunTask(ctx context.Context, task *tasks.Task) error
	Shutdown() error
}

type ReceiveTaskResponse func(originalTask *tasks.Task, response *tasks.TaskResult, err error)
