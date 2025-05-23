package avsPerformer

import (
	"context"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"slices"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
)

type AvsProcessType string

const (
	AvsProcessTypeServer AvsProcessType = "server"
	AvsProcessTypeOneOff AvsProcessType = "one-off"
)

type PerformerImage struct {
	Registry string
	Digest   string
	Tag      string
}

type AvsPerformerConfig struct {
	AvsAddress           string
	ProcessType          AvsProcessType
	Image                PerformerImage
	WorkerCount          int
	PerformerNetworkName string
	SigningCurve         string // bn254, bls381, etc
	AVSRegistrarAddress  string
}

type IAvsPerformer interface {
	Initialize(ctx context.Context) error
	RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error)
	ValidateTaskSignature(task *performerTask.PerformerTask) error
	Shutdown() error
	GetContainerId() string
}

func (ap *AvsPerformerConfig) Validate() error {
	var allErrors field.ErrorList
	if ap.Image.Registry == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("image.repository"), "image.repository is required"))
	}
	if ap.Image.Digest == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("image.tag"), "image.tag is required"))
	}
	if ap.SigningCurve == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("signingCurve"), "signingCurve is required"))
	} else if !slices.Contains([]string{"bn254", "bls381"}, ap.SigningCurve) {
		allErrors = append(allErrors, field.Invalid(field.NewPath("signingCurve"), ap.SigningCurve, "signingCurve must be one of [bn254, bls381]"))
	}

	if ap.WorkerCount == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("workerCount"), "workerCount is required"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}
