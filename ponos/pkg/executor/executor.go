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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/kubernetesManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/rpcServer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"go.uber.org/zap"
)

type Executor struct {
	logger              *zap.Logger
	config              *executorConfig.ExecutorConfig
	avsPerformers       map[string]avsPerformer.IAvsPerformer
	taskRpcServer       *rpcServer.RpcServer
	managementRpcServer *rpcServer.RpcServer
	ecdsaSigner         signer.ISigner
	bn254Signer         signer.ISigner
	inflightTasks       *sync.Map

	l1ContractCaller contractCaller.IContractCaller

	peeringFetcher peering.IPeeringDataFetcher

	// store is the persistence layer
	store storage.ExecutorStore
}

func NewExecutorWithRpcServers(
	taskServerPort int,
	managementServerPort int,
	config *executorConfig.ExecutorConfig,
	logger *zap.Logger,
	signers signer.Signers,
	peeringFetcher peering.IPeeringDataFetcher,
	l1ContractCaller contractCaller.IContractCaller,
	store storage.ExecutorStore,
) (*Executor, error) {
	taskRpc, err := rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
		GrpcPort: taskServerPort,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %v", err)
	}

	var managementRpc *rpcServer.RpcServer
	if managementServerPort == 0 || managementServerPort == taskServerPort {
		managementRpc = taskRpc
	} else {
		managementRpc, err = rpcServer.NewRpcServer(&rpcServer.RpcServerConfig{
			GrpcPort: managementServerPort,
		}, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create management RPC server: %v", err)
		}
	}

	return NewExecutor(config, taskRpc, managementRpc, logger, signers, peeringFetcher, l1ContractCaller, store), nil
}

func NewExecutor(
	config *executorConfig.ExecutorConfig,
	taskRpcServer *rpcServer.RpcServer,
	managementRpcServer *rpcServer.RpcServer,
	logger *zap.Logger,
	signers signer.Signers,
	peeringFetcher peering.IPeeringDataFetcher,
	l1ContractCaller contractCaller.IContractCaller,
	store storage.ExecutorStore,
) *Executor {
	if store == nil {
		panic("store is required")
	}
	return &Executor{
		logger:              logger,
		config:              config,
		avsPerformers:       make(map[string]avsPerformer.IAvsPerformer),
		taskRpcServer:       taskRpcServer,
		managementRpcServer: managementRpcServer,
		ecdsaSigner:         signers.ECDSASigner,
		bn254Signer:         signers.BLSSigner,
		inflightTasks:       &sync.Map{},
		peeringFetcher:      peeringFetcher,
		l1ContractCaller:    l1ContractCaller,
		store:               store,
	}
}

func (e *Executor) Initialize(ctx context.Context) error {
	e.logger.Sugar().Infow("Initializing AVS performers")

	// Perform recovery from storage
	if err := e.recoverFromStorage(ctx); err != nil {
		e.logger.Sugar().Warnw("Failed to recover from storage", "error", err)
		// Continue anyway - this is not a fatal error
	}

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
			} else {
				e.logger.Sugar().Infow("AVS performer deployed successfully",
					zap.String("avsAddress", avsAddress),
					zap.String("deploymentMode", string(avs.DeploymentMode)),
					zap.String("deploymentId", result.ID),
					zap.String("performerId", result.PerformerID),
				)
			}

			e.avsPerformers[avsAddress] = performer

			// Save performer state to storage
			performerState := &storage.PerformerState{
				PerformerId:        result.PerformerID,
				AvsAddress:         avsAddress,
				ContainerId:        result.ID,
				Status:             "running",
				ArtifactRegistry:   avs.Image.Repository,
				ArtifactTag:        avs.Image.Tag,
				ArtifactDigest:     "", // Not available during initialization
				DeploymentMode:     string(avs.DeploymentMode),
				CreatedAt:          result.StartTime,
				LastHealthCheck:    result.EndTime,
				ContainerHealthy:   true,
				ApplicationHealthy: true,
			}
			if err := e.store.SavePerformerState(ctx, result.PerformerID, performerState); err != nil {
				e.logger.Sugar().Warnw("Failed to save performer state to storage",
					"error", err,
					"performerId", result.PerformerID,
				)
			}

		default:
			e.logger.Sugar().Errorw("Unsupported AVS performer process type",
				zap.String("avsAddress", avsAddress),
				zap.String("processType", avs.ProcessType),
			)
			return fmt.Errorf("unsupported AVS performer process type: %s", avs.ProcessType)
		}
	}

	if err := e.registerHandlers(); err != nil {
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

// recoverFromStorage loads performer states from storage and verifies they're still running
func (e *Executor) recoverFromStorage(ctx context.Context) error {
	performerStates, err := e.store.ListPerformerStates(ctx)
	if err != nil {
		return fmt.Errorf("failed to list performer states: %w", err)
	}

	e.logger.Sugar().Infow("Recovering performer states from storage",
		"count", len(performerStates),
	)

	// TODO: In a future milestone, we will verify if containers/pods still exist
	// and re-create missing performers. For now, just log the recovery.
	for _, state := range performerStates {
		e.logger.Sugar().Infow("Found performer state in storage",
			"performerId", state.PerformerId,
			"avsAddress", state.AvsAddress,
			"status", state.Status,
			"containerId", state.ContainerId,
		)
	}

	// Load inflight tasks
	inflightTasks, err := e.store.ListInflightTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to list inflight tasks: %w", err)
	}

	e.logger.Sugar().Infow("Recovering inflight tasks from storage",
		"count", len(inflightTasks),
	)

	for _, task := range inflightTasks {
		e.inflightTasks.Store(task.TaskId, task)
		e.logger.Sugar().Infow("Recovered inflight task",
			"taskId", task.TaskId,
			"avsAddress", task.AvsAddress,
		)
	}

	return nil
}

func (e *Executor) Run(ctx context.Context) error {
	e.logger.Info("Executor is running",
		zap.String("version", "1.0.0"),
		zap.String("operatorAddress", e.config.Operator.Address),
	)
	if err := e.taskRpcServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start RPC server: %v", err)
	}
	return nil
}

func (e *Executor) registerHandlers() error {
	executorV1.RegisterExecutorServiceServer(e.taskRpcServer.GetGrpcServer(), e)
	executorV1.RegisterExecutorManagementServiceServer(e.managementRpcServer.GetGrpcServer(), e)

	return nil
}
