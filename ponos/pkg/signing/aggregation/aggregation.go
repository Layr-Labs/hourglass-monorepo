package aggregation

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

type ITaskResultAggregator[SigT, CertT, PubKeyT any] interface {
	SigningThresholdMet() bool

	ProcessNewSignature(
		ctx context.Context,
		taskResponse *types.TaskResult,
	) error

	VerifyResponseSignature(taskResponse *types.TaskResult, operator *Operator[PubKeyT], outputDigest [32]byte) (*SigT, error)

	GenerateFinalCertificate() (*CertT, error)
}

type Operator[PubKeyT any] struct {
	Address       string
	PublicKey     PubKeyT
	OperatorIndex uint32
}

func (o *Operator[PubKeyT]) GetAddress() common.Address {
	return common.HexToAddress(o.Address)
}

// Error variables for input validation
var (
	ErrInvalidTaskId             = fmt.Errorf("taskId must not be empty")
	ErrNoOperatorAddresses       = fmt.Errorf("operatorAddresses must not be empty")
	ErrInvalidThreshold          = fmt.Errorf("thresholdPercentage must be between 1 and 100")
	ErrInvalidReferenceTimestamp = fmt.Errorf("referenceTimestamp must be positive")
)
