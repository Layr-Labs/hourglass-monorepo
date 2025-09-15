package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/common"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Executor struct {
	logger              *zap.Logger
	config              *executorConfig.ExecutorConfig
	avsPerformers       *sync.Map // map[string]avsPerformer.IAvsPerformer
	taskRpcServer       *rpcServer.RpcServer
	managementRpcServer *rpcServer.RpcServer
	ecdsaSigner         signer.ISigner
	bn254Signer         signer.ISigner
	inflightTasks       *sync.Map

	l1ContractCaller contractCaller.IContractCaller

	peeringFetcher peering.IPeeringDataFetcher

	store storage.ExecutorStore

	authVerifier *auth.Verifier
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

	var verifier *auth.Verifier
	if config.AuthConfig != nil && config.AuthConfig.IsEnabled {
		var authSigner signer.ISigner

		if signers.ECDSASigner != nil {
			authSigner = signers.ECDSASigner
		} else if signers.BLSSigner != nil {
			authSigner = signers.BLSSigner
		}

		tokenManager := auth.NewChallengeTokenManager(config.Operator.Address, 5*time.Minute)
		verifier = auth.NewVerifier(tokenManager, authSigner)
	}

	return &Executor{
		logger:              logger,
		config:              config,
		avsPerformers:       &sync.Map{},
		taskRpcServer:       taskRpcServer,
		managementRpcServer: managementRpcServer,
		ecdsaSigner:         signers.ECDSASigner,
		bn254Signer:         signers.BLSSigner,
		inflightTasks:       &sync.Map{},
		peeringFetcher:      peeringFetcher,
		l1ContractCaller:    l1ContractCaller,
		store:               store,
		authVerifier:        verifier,
	}
}

func (e *Executor) Initialize(ctx context.Context) error {
	e.logger.Sugar().Infow("Initializing AVS performers")

	if err := e.rehydratePerformersFromStorage(ctx); err != nil {
		e.logger.Sugar().Warnw("Failed to rehydrate performers from storage, will create fresh performers",
			"error", err,
		)
	}

	for _, avs := range e.config.AvsPerformers {
		avsAddress := strings.ToLower(avs.AvsAddress)
		if _, ok := e.avsPerformers.Load(avsAddress); ok {
			e.logger.Sugar().Infow("AVS performer already exists",
				zap.String("avsAddress", avsAddress),
				zap.String("processType", avs.ProcessType),
			)
			continue
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

			var serviceAccountName string
			if avs.Kubernetes != nil && avs.Kubernetes.ServiceAccountName != "" {
				serviceAccountName = avs.Kubernetes.ServiceAccountName
			}

			image := avsPerformer.PerformerImage{
				Repository:         avs.Image.Repository,
				Tag:                avs.Image.Tag,
				Envs:               avs.Envs,
				ServiceAccountName: serviceAccountName,
			}

			result, err := performer.Deploy(ctx, image)
			if err != nil {
				e.logger.Sugar().Errorw("Failed to deploy performer during startup",
					zap.String("avsAddress", avsAddress),
					zap.String("deploymentMode", string(avs.DeploymentMode)),
					zap.Error(err),
				)
				return err
			} else {
				e.logger.Sugar().Infow("AVS performer deployed successfully",
					zap.String("avsAddress", avsAddress),
					zap.String("deploymentMode", string(avs.DeploymentMode)),
					zap.String("deploymentId", result.ID),
					zap.String("performerId", result.PerformerID),
				)
			}

			e.avsPerformers.Store(avsAddress, performer)

			var envRecords []storage.EnvironmentVarRecord
			for _, env := range avs.Envs {
				envRecords = append(envRecords, storage.EnvironmentVarRecord{
					Name:         env.Name,
					Value:        env.Value,
					ValueFromEnv: env.ValueFromEnv,
				})
			}

			performerState := &storage.PerformerState{
				PerformerId:        result.PerformerID,
				AvsAddress:         avsAddress,
				ContainerId:        result.ID,
				Status:             "running",
				ArtifactRegistry:   avs.Image.Repository,
				ArtifactTag:        avs.Image.Tag,
				ArtifactDigest:     "",
				DeploymentMode:     string(avs.DeploymentMode),
				CreatedAt:          result.StartTime,
				LastHealthCheck:    result.EndTime,
				ContainerHealthy:   true,
				ApplicationHealthy: true,
				NetworkName:        e.config.PerformerNetworkName,
				ContainerEndpoint:  result.Endpoint,
				ContainerHostname:  result.Hostname,
				InternalPort:       8080,
				EnvironmentVars:    envRecords,
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
		e.avsPerformers.Range(func(key, value interface{}) bool {
			avsAdd := key.(string)
			perf := value.(avsPerformer.IAvsPerformer)
			if err := perf.Shutdown(); err != nil {
				e.logger.Sugar().Errorw("Failed to shutdown AVS performer",
					zap.String("avsAddress", avsAdd),
					zap.Error(err),
				)
			}
			return true
		})
	}()

	return nil
}

// rehydratePerformersFromStorage attempts to rehydrate performers from persisted state
func (e *Executor) rehydratePerformersFromStorage(ctx context.Context) error {
	e.logger.Sugar().Infow("Attempting to rehydrate performers from storage")

	states, err := e.store.ListPerformerStates(ctx)
	if err != nil {
		return fmt.Errorf("failed to list performer states: %w", err)
	}

	if len(states) == 0 {
		e.logger.Sugar().Infow("No persisted performer states found")
		return nil
	}

	e.logger.Sugar().Infow("Found persisted performer states",
		"count", len(states),
	)

	rehydratedCount := 0
	failedCount := 0

	for _, state := range states {
		avsAddress := strings.ToLower(state.AvsAddress)

		if _, exists := e.avsPerformers.Load(avsAddress); exists {
			e.logger.Sugar().Infow("Performer already exists for AVS, skipping rehydration",
				"avsAddress", avsAddress,
			)
			continue
		}

		if state.DeploymentMode != string(executorConfig.DeploymentModeDocker) {
			e.logger.Sugar().Warnw("Rehydration not supported for deployment mode, skipping rehydration",
				"deploymentMode", state.DeploymentMode,
				"performerId", state.PerformerId,
			)
			continue
		}

		performer, err := avsContainerPerformer.NewAvsContainerPerformer(
			&avsPerformer.AvsPerformerConfig{
				AvsAddress:           avsAddress,
				ProcessType:          avsPerformer.AvsProcessTypeServer,
				PerformerNetworkName: state.NetworkName,
			},
			e.logger,
		)

		if err != nil {
			e.logger.Sugar().Errorw("Failed to create container performer for rehydration",
				"avsAddress", avsAddress,
				"error", err,
			)
			failedCount++
			continue
		}

		if err := performer.Initialize(ctx); err != nil {
			e.logger.Sugar().Errorw("Failed to initialize performer for rehydration",
				"avsAddress", avsAddress,
				"error", err,
			)
			failedCount++
			continue
		}

		e.logger.Sugar().Infow("Attempting to rehydrate performer",
			"performerId", state.PerformerId,
			"containerId", state.ContainerId,
			"avsAddress", avsAddress,
		)

		err = performer.RehydrateFromState(ctx, state)

		if err != nil {
			e.logger.Sugar().Warnw("Failed to rehydrate performer, cleaning up state",
				"performerId", state.PerformerId,
				"containerId", state.ContainerId,
				"error", err,
			)

			if err := e.store.DeletePerformerState(ctx, state.PerformerId); err != nil {
				e.logger.Sugar().Warnw("Failed to delete stale performer state",
					"performerId", state.PerformerId,
					"error", err,
				)
			}
			failedCount++

		} else {
			e.logger.Sugar().Infow("Successfully rehydrated performer",
				"performerId", state.PerformerId,
				"containerId", state.ContainerId,
				"avsAddress", avsAddress,
			)

			e.avsPerformers.Store(avsAddress, performer)
			rehydratedCount++
		}
	}

	e.logger.Sugar().Infow("Performer rehydration complete",
		"totalStates", len(states),
		"rehydrated", rehydratedCount,
		"failed", failedCount,
	)

	return nil
}

// createPerformer creates an AVS performer based on the deployment mode
func (e *Executor) createPerformer(avs *executorConfig.AvsPerformerConfig, avsAddress string) (avsPerformer.IAvsPerformer, error) {

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
		e.logger,
	)
}

// createKubernetesPerformer creates a Kubernetes-based AVS performer
func (e *Executor) createKubernetesPerformer(avs *executorConfig.AvsPerformerConfig, avsAddress string) (avsPerformer.IAvsPerformer, error) {
	if e.config.Kubernetes == nil {
		return nil, fmt.Errorf("kubernetes configuration is required for kubernetes deployment mode")
	}

	kubernetesConfig := &kubernetesManager.Config{
		Namespace:         e.config.Kubernetes.Namespace,
		OperatorNamespace: e.config.Kubernetes.OperatorNamespace,
		CRDGroup:          e.config.Kubernetes.CRDGroup,
		CRDVersion:        e.config.Kubernetes.CRDVersion,
		ConnectionTimeout: e.config.Kubernetes.ConnectionTimeout,
		KubeconfigPath:    e.config.Kubernetes.KubeConfigPath,
	}

	var endpointOverride string
	if avs.Kubernetes != nil && avs.Kubernetes.EndpointOverride != "" {
		endpointOverride = avs.Kubernetes.EndpointOverride
	}

	return avsKubernetesPerformer.NewAvsKubernetesPerformer(
		&avsPerformer.AvsPerformerConfig{
			AvsAddress:       avsAddress,
			ProcessType:      avsPerformer.AvsProcessType(avs.ProcessType),
			EndpointOverride: endpointOverride,
		},
		kubernetesConfig,
		e.logger,
	)
}

func (e *Executor) Run(ctx context.Context) error {
	e.logger.Info("Executor is running",
		zap.String("version", "1.0.0"),
		zap.String("operatorAddress", e.config.Operator.Address),
	)
	if err := e.taskRpcServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task RPC server: %v", err)
	}

	if e.managementRpcServer != nil && e.managementRpcServer != e.taskRpcServer {
		if err := e.managementRpcServer.Start(ctx); err != nil {
			return fmt.Errorf("failed to start management RPC server: %v", err)
		}
	}
	return nil
}

func (e *Executor) registerHandlers() error {
	executorV1.RegisterExecutorServiceServer(e.taskRpcServer.GetGrpcServer(), e)
	executorV1.RegisterExecutorManagementServiceServer(e.managementRpcServer.GetGrpcServer(), e)

	return nil
}

// verifyAuth is a helper method to verify authentication
func (e *Executor) verifyAuth(auth *commonV1.AuthSignature) error {
	if e.authVerifier == nil {
		if auth != nil {
			return status.Error(codes.Unimplemented, "authentication is not enabled")
		}
		return nil
	}

	if err := e.authVerifier.VerifyAuthentication(auth); err != nil {
		return err
	}
	return nil
}
