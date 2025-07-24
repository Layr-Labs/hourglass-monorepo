package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/avsContainerPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/avsKubernetesPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/kubernetesManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Executor struct {
	logger        *zap.Logger
	config        *executorConfig.ExecutorConfig
	avsPerformers map[string]avsPerformer.IAvsPerformer
	rpcServer     *rpcServer.RpcServer
	ecdsaSigner   signer.ISigner
	bn254Signer   signer.ISigner
	inflightTasks *sync.Map

	l1ContractCaller contractCaller.IContractCaller

	peeringFetcher peering.IPeeringDataFetcher
}

func NewExecutorWithRpcServer(
	port int,
	config *executorConfig.ExecutorConfig,
	logger *zap.Logger,
	signers signer.Signers,
	peeringFetcher peering.IPeeringDataFetcher,
	l1ContractCaller contractCaller.IContractCaller,
) (*Executor, error) {
	rpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: port,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %v", err)
	}

	return NewExecutor(config, rpc, logger, signers, peeringFetcher, l1ContractCaller), nil
}

func NewExecutor(
	config *executorConfig.ExecutorConfig,
	rpcServer *rpcServer.RpcServer,
	logger *zap.Logger,
	signers signer.Signers,
	peeringFetcher peering.IPeeringDataFetcher,
	l1ContractCaller contractCaller.IContractCaller,
) *Executor {
	return &Executor{
		logger:           logger,
		config:           config,
		avsPerformers:    make(map[string]avsPerformer.IAvsPerformer),
		rpcServer:        rpcServer,
		ecdsaSigner:      signers.ECDSASigner,
		bn254Signer:      signers.BLSSigner,
		inflightTasks:    &sync.Map{},
		peeringFetcher:   peeringFetcher,
		l1ContractCaller: l1ContractCaller,
	}
}

func (e *Executor) Initialize(ctx context.Context) error {
	e.logger.Sugar().Infow("Initializing AVS performers")
	for _, avs := range e.config.AvsPerformers {
		avsAddress := strings.ToLower(avs.AvsAddress)
		if _, ok := e.avsPerformers[avsAddress]; ok {
			e.logger.Sugar().Errorw("AVS performer already exists",
				zap.String("avsAddress", avsAddress),
				zap.String("processType", avs.ProcessType),
			)
		}

		switch avs.ProcessType {
		case string(avsPerformer.AvsProcessTypeServer):
			performer, err := e.createPerformer(avs, avsAddress)
			if err != nil {
				return fmt.Errorf("failed to create AVS performer for %s: %v", avs.ProcessType, err)
			}

			err = performer.Initialize(ctx)
			if err != nil {
				return err
			}

			// Deploy performer using the performer's Deploy method
			image := avsPerformer.PerformerImage{
				Repository: avs.Image.Repository,
				Tag:        avs.Image.Tag,
				Envs:       avs.Envs,
			}

			result, err := performer.Deploy(ctx, image)
			if err != nil {
				e.logger.Sugar().Errorw("Failed to deploy performer during startup",
					zap.String("avsAddress", avsAddress),
					zap.String("deploymentMode", string(avs.DeploymentMode)),
					zap.Error(err),
				)
				return fmt.Errorf("failed to deploy performer for AVS %s: %w", avsAddress, err)
			}

			e.logger.Sugar().Infow("AVS performer deployed successfully",
				zap.String("avsAddress", avsAddress),
				zap.String("deploymentMode", string(avs.DeploymentMode)),
				zap.String("deploymentId", result.ID),
				zap.String("performerId", result.PerformerID),
			)

			e.avsPerformers[avsAddress] = performer

		default:
			e.logger.Sugar().Errorw("Unsupported AVS performer process type",
				zap.String("avsAddress", avsAddress),
				zap.String("processType", avs.ProcessType),
			)
			return fmt.Errorf("unsupported AVS performer process type: %s", avs.ProcessType)
		}
	}

	if err := e.registerHandlers(e.rpcServer.GetGrpcServer()); err != nil {
		e.logger.Sugar().Errorw("Failed to register handlers",
			zap.Error(err),
		)
		return fmt.Errorf("failed to register handlers: %v", err)
	}

	go func() {
		<-ctx.Done()
		e.logger.Sugar().Info("Shutting down AVS performers")
		for avsAdd, perf := range e.avsPerformers {
			if err := perf.Shutdown(); err != nil {
				e.logger.Sugar().Errorw("Failed to shutdown AVS performer",
					zap.String("avsAddress", avsAdd),
					zap.Error(err),
				)
			}
		}
	}()

	return nil
}

// createPerformer creates an AVS performer based on the deployment mode
func (e *Executor) createPerformer(avs *executorConfig.AvsPerformerConfig, avsAddress string) (avsPerformer.IAvsPerformer, error) {
	// Use default deployment mode if not specified
	deploymentMode := avs.DeploymentMode
	if deploymentMode == "" {
		deploymentMode = executorConfig.DeploymentModeDocker
	}

	switch deploymentMode {
	case executorConfig.DeploymentModeDocker:
		return e.createDockerPerformer(avs, avsAddress)
	case executorConfig.DeploymentModeKubernetes:
		return e.createKubernetesPerformer(avs, avsAddress)
	default:
		return nil, fmt.Errorf("unsupported deployment mode: %s", deploymentMode)
	}
}

// createDockerPerformer creates a Docker-based AVS performer
func (e *Executor) createDockerPerformer(avs *executorConfig.AvsPerformerConfig, avsAddress string) (avsPerformer.IAvsPerformer, error) {
	return avsContainerPerformer.NewAvsContainerPerformer(
		&avsPerformer.AvsPerformerConfig{
			AvsAddress:           avsAddress,
			ProcessType:          avsPerformer.AvsProcessType(avs.ProcessType),
			PerformerNetworkName: e.config.PerformerNetworkName,
		},
		e.peeringFetcher,
		e.l1ContractCaller,
		e.logger,
	)
}

// createKubernetesPerformer creates a Kubernetes-based AVS performer
func (e *Executor) createKubernetesPerformer(avs *executorConfig.AvsPerformerConfig, avsAddress string) (avsPerformer.IAvsPerformer, error) {
	if e.config.Kubernetes == nil {
		return nil, fmt.Errorf("kubernetes configuration is required for kubernetes deployment mode")
	}

	// Convert executor config to kubernetes manager config
	kubernetesConfig := &kubernetesManager.Config{
		Namespace:         e.config.Kubernetes.Namespace,
		OperatorNamespace: e.config.Kubernetes.OperatorNamespace,
		CRDGroup:          e.config.Kubernetes.CRDGroup,
		CRDVersion:        e.config.Kubernetes.CRDVersion,
		ConnectionTimeout: e.config.Kubernetes.ConnectionTimeout,
		KubeconfigPath:    e.config.Kubernetes.KubeConfigPath,
	}

	return avsKubernetesPerformer.NewAvsKubernetesPerformer(
		&avsPerformer.AvsPerformerConfig{
			AvsAddress:         avsAddress,
			ProcessType:        avsPerformer.AvsProcessType(avs.ProcessType),
			SkipConnectionTest: avs.SkipConnectionTest,
		},
		kubernetesConfig,
		e.peeringFetcher,
		e.l1ContractCaller,
		e.logger,
	)
}

func (e *Executor) Run(ctx context.Context) error {
	e.logger.Info("Executor is running",
		zap.String("version", "1.0.0"),
		zap.String("operatorAddress", e.config.Operator.Address),
	)
	if err := e.rpcServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start RPC server: %v", err)
	}
	return nil
}

func (e *Executor) registerHandlers(grpcServer *grpc.Server) error {
	executorV1.RegisterExecutorServiceServer(grpcServer, e)

	return nil
}
