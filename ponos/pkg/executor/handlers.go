package executor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/avsContainerPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (e *Executor) SubmitTask(ctx context.Context, req *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	res, err := e.handleReceivedTask(ctx, req)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to handle received task",
			"taskId", req.TaskId,
			"avsAddress", req.AvsAddress,
			"error", err,
		)
		return nil, fmt.Errorf("Failed to handle received task: %w", err)
	}
	return res, nil
}

// validateDeployArtifactRequest validates the DeployArtifactRequest and returns an error message if invalid
func validateDeployArtifactRequest(req *executorV1.DeployArtifactRequest) string {
	if req.GetAvsAddress() == "" {
		return "AVS address is required"
	}
	if req.GetDigest() == "" {
		return "Artifact digest is required"
	}
	if req.GetRegistryUrl() == "" {
		return "Registry URL is required"
	}
	return ""
}

func (e *Executor) DeployArtifact(ctx context.Context, req *executorV1.DeployArtifactRequest) (*executorV1.DeployArtifactResponse, error) {
	e.logger.Info("Received deploy artifact request",
		zap.String("avsAddress", req.AvsAddress),
		zap.String("digest", req.Digest),
		zap.String("registryUrl", req.RegistryUrl),
	)

	// Verify authentication
	if err := auth.HandleAuthError(e.verifyAuth(req.Auth)); err != nil {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: "Authentication failed",
		}, err
	}

	// Validate request
	if errMsg := validateDeployArtifactRequest(req); errMsg != "" {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: errMsg,
		}, status.Error(codes.InvalidArgument, errMsg)
	}

	avsAddress := strings.ToLower(req.AvsAddress)

	// Find or create the AVS performer
	performer, err := e.getOrCreateAvsPerformer(ctx, avsAddress)
	if err != nil {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get or create AVS performer: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Create deployment record in storage
	deploymentId := fmt.Sprintf("deployment-%s-%d", avsAddress, time.Now().UnixNano())
	deploymentInfo := &storage.DeploymentInfo{
		DeploymentId:     deploymentId,
		AvsAddress:       avsAddress,
		ArtifactRegistry: req.GetRegistryUrl(),
		ArtifactDigest:   req.GetDigest(),
		Status:           storage.DeploymentStatusPending,
		StartedAt:        time.Now(),
		CompletedAt:      nil,
		Error:            "",
	}
	if err := e.store.SaveDeployment(ctx, deploymentId, deploymentInfo); err != nil {
		e.logger.Sugar().Warnw("Failed to save deployment to storage",
			"error", err,
			"deploymentId", deploymentId,
		)
	}

	// Deploy using the performer's Deploy method
	image := avsPerformer.PerformerImage{
		Repository: req.GetRegistryUrl(),
		Digest:     req.GetDigest(),
		Envs: util.Map(req.GetEnv(), func(env *executorV1.PerformerEnv, i uint64) config.AVSPerformerEnv {
			performerEnv := config.AVSPerformerEnv{
				Name:         env.GetName(),
				Value:        env.GetValue(),
				ValueFromEnv: env.GetValueFromEnv(),
			}
			// Map Kubernetes environment variables if present
			if env.GetKubernetesEnv() != nil && env.GetKubernetesEnv().GetValueFrom() != nil {
				performerEnv.KubernetesEnv = &config.KubernetesEnv{
					ValueFrom: struct {
						SecretKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"secretKeyRef" yaml:"secretKeyRef"`
						ConfigMapKeyRef struct {
							Name string `json:"name" yaml:"name"`
							Key  string `json:"key" yaml:"key"`
						} `json:"configMapKeyRef" yaml:"configMapKeyRef"`
					}{},
				}
				if env.GetKubernetesEnv().GetValueFrom().GetSecretKeyRef() != nil {
					performerEnv.KubernetesEnv.ValueFrom.SecretKeyRef.Name = env.GetKubernetesEnv().GetValueFrom().GetSecretKeyRef().GetName()
					performerEnv.KubernetesEnv.ValueFrom.SecretKeyRef.Key = env.GetKubernetesEnv().GetValueFrom().GetSecretKeyRef().GetKey()
				}
				if env.GetKubernetesEnv().GetValueFrom().GetConfigMapKeyRef() != nil {
					performerEnv.KubernetesEnv.ValueFrom.ConfigMapKeyRef.Name = env.GetKubernetesEnv().GetValueFrom().GetConfigMapKeyRef().GetName()
					performerEnv.KubernetesEnv.ValueFrom.ConfigMapKeyRef.Key = env.GetKubernetesEnv().GetValueFrom().GetConfigMapKeyRef().GetKey()
				}
			}
			return performerEnv
		}),
	}

	// Set Kubernetes service account if provided
	if req.GetKubernetes() != nil && req.GetKubernetes().GetServiceAccountName() != "" {
		image.ServiceAccountName = req.GetKubernetes().GetServiceAccountName()
	}

	// Update deployment status to deploying
	if err := e.store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusDeploying); err != nil {
		e.logger.Sugar().Warnw("Failed to update deployment status",
			"error", err,
			"deploymentId", deploymentId,
		)
	}

	result, err := performer.Deploy(ctx, image)
	if err != nil {
		// Update deployment status to failed
		if updateErr := e.store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusFailed); updateErr != nil {
			e.logger.Sugar().Warnw("Failed to update deployment status to failed",
				"error", updateErr,
				"deploymentId", deploymentId,
			)
		}

		// Check for specific error types to return appropriate gRPC status codes
		if strings.Contains(err.Error(), "deployment already in progress") {
			return &executorV1.DeployArtifactResponse{
				Success: false,
				Message: err.Error(),
			}, status.Error(codes.AlreadyExists, "deployment already in progress")
		}

		if strings.Contains(err.Error(), "deployment timeout") {
			return &executorV1.DeployArtifactResponse{
				Success: false,
				Message: err.Error(),
			}, status.Error(codes.DeadlineExceeded, "deployment timeout")
		}

		// Log the error with full context
		e.logger.Error("Deployment failed",
			zap.String("avsAddress", avsAddress),
			zap.String("registryUrl", req.GetRegistryUrl()),
			zap.String("digest", req.GetDigest()),
			zap.Error(err),
		)

		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	e.logger.Info("Deployment completed successfully",
		zap.String("avsAddress", avsAddress),
		zap.String("deploymentId", result.ID),
		zap.String("performerId", result.PerformerID),
		zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
	)

	// Update deployment status to running and save performer state
	if err := e.store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusRunning); err != nil {
		e.logger.Sugar().Warnw("Failed to update deployment status to running",
			"error", err,
			"deploymentId", deploymentId,
		)
	}

	// Save performer state
	performerState := &storage.PerformerState{
		PerformerId:        result.PerformerID,
		AvsAddress:         avsAddress,
		ContainerId:        result.ID,
		Status:             "running",
		ArtifactRegistry:   req.GetRegistryUrl(),
		ArtifactTag:        "", // Not available from request
		ArtifactDigest:     req.GetDigest(),
		DeploymentMode:     "docker", // Default for now
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

	return &executorV1.DeployArtifactResponse{
		Success:      true,
		Message:      result.Message,
		DeploymentId: result.ID,
	}, nil
}

func (e *Executor) createAvsPerformer(avsAddress string) (avsPerformer.IAvsPerformer, error) {

	e.logger.Info("Creating new AVS performer for address", zap.String("avsAddress", avsAddress))

	// Create config without image info - will be deployed via Deploy method
	c := &avsPerformer.AvsPerformerConfig{
		AvsAddress:           avsAddress,
		ProcessType:          avsPerformer.AvsProcessTypeServer,
		Image:                avsPerformer.PerformerImage{},
		PerformerNetworkName: e.config.PerformerNetworkName,
	}

	newPerformer, err := avsContainerPerformer.NewAvsContainerPerformer(
		c,
		e.logger,
	)
	if err != nil {
		return nil, err
	}

	return newPerformer, nil

}

// getOrCreateAvsPerformer gets an existing AVS performer or creates a new one if it doesn't exist
func (e *Executor) getOrCreateAvsPerformer(ctx context.Context, avsAddress string) (avsPerformer.IAvsPerformer, error) {
	// Fast path: try to load existing performer first
	if performer, ok := e.avsPerformers.Load(avsAddress); ok {
		return performer.(avsPerformer.IAvsPerformer), nil
	}

	// Slow path: need to create a new performer
	// Create the performer
	newPerformer, err := e.createAvsPerformer(avsAddress)
	if err != nil {
		return nil, err
	}

	if err := newPerformer.Initialize(ctx); err != nil {
		return nil, err
	}

	// Use LoadOrStore to atomically check-and-store, preventing race conditions
	actual, loaded := e.avsPerformers.LoadOrStore(avsAddress, newPerformer)
	if loaded {
		// Another goroutine already created and stored a performer for this address
		// Shutdown our newly created performer to prevent resource leak
		e.logger.Info("AVS performer already exists, using existing one", zap.String("avsAddress", avsAddress))
		if err := newPerformer.Shutdown(); err != nil {
			e.logger.Sugar().Warnw("Failed to shutdown duplicate performer",
				"avsAddress", avsAddress,
				"error", err,
			)
		}
		return actual.(avsPerformer.IAvsPerformer), nil
	}

	// We successfully stored our performer
	return newPerformer, nil
}

func (e *Executor) handleReceivedTask(ctx context.Context, task *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	e.logger.Sugar().Infow("Received task from AVS",
		"taskId", task.TaskId,
		"avsAddress", task.AvsAddress,
		"executorAddress", task.ExecutorAddress,
		"aggregatorAddress", task.AggregatorAddress,
	)

	// Check if task has already been processed
	processed, err := e.store.IsTaskProcessed(ctx, task.TaskId)
	if err != nil {
		e.logger.Sugar().Warnw("Failed to check if task is processed",
			"taskId", task.TaskId,
			"error", err,
		)
		// Continue processing even if check fails
	} else if processed {
		e.logger.Sugar().Warnw("Task already processed, skipping",
			"taskId", task.TaskId,
			"avsAddress", task.AvsAddress,
		)
		return nil, fmt.Errorf("task %s already processed", task.TaskId)
	}

	avsAddress := strings.ToLower(task.GetAvsAddress())
	if avsAddress == "" {
		return nil, fmt.Errorf("AVS address is empty")
	}

	if task.ExecutorAddress == "" {
		return nil, fmt.Errorf("executor address is empty")
	}

	if !strings.EqualFold(task.ExecutorAddress, e.config.Operator.Address) {
		e.logger.Sugar().Errorw("Task executor address mismatch",
			"expected", e.config.Operator.Address,
			"received", task.ExecutorAddress,
		)
		return nil, fmt.Errorf("task executor address mismatch: expected %s, got %s",
			e.config.Operator.Address, task.ExecutorAddress)
	}

	// Validate operator is in the operator set
	if err := e.validateOperatorInSet(task); err != nil {
		return nil, err
	}

	// Validate task signature
	if err := e.validateTaskSignature(task); err != nil {
		e.logger.Sugar().Errorw("Task signature validation failed",
			"taskId", task.TaskId,
			"error", err,
		)
		return nil, fmt.Errorf("signature validation failed: %w", err)
	}

	value, ok := e.avsPerformers.Load(avsAddress)
	if !ok {
		return nil, fmt.Errorf("AVS performer not found for address %s", avsAddress)
	}
	avsPerf := value.(avsPerformer.IAvsPerformer)

	pt := performerTask.NewPerformerTaskFromTaskSubmissionProto(task)
	e.inflightTasks.Store(task.TaskId, task)

	// Save inflight task to storage
	taskInfo := &storage.TaskInfo{
		TaskId:            task.TaskId,
		AvsAddress:        avsAddress,
		OperatorAddress:   e.config.Operator.Address,
		ReceivedAt:        time.Now(),
		Status:            "processing",
		AggregatorAddress: task.GetAggregatorAddress(),
		OperatorSetId:     task.OperatorSetId,
	}
	if err := e.store.SaveInflightTask(ctx, task.TaskId, taskInfo); err != nil {
		e.logger.Sugar().Warnw("Failed to save inflight task to storage",
			"error", err,
			"taskId", task.TaskId,
		)
	}

	// Cleanup inflight task and delete from storage irrespective of the result of the task.
	defer func() {
		e.inflightTasks.Delete(task.TaskId)
		if err := e.store.DeleteInflightTask(context.Background(), task.TaskId); err != nil {
			e.logger.Sugar().Warnw("Failed to delete inflight task from storage",
				"error", err,
				"taskId", task.TaskId,
			)
		}
	}()

	response, err := avsPerf.RunTask(ctx, pt)

	if err != nil {
		e.logger.Sugar().Errorw("Failed to run task",
			"taskId", task.TaskId,
			"avsAddress", avsAddress,
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "Failed to run task %s", err.Error())
	}

	resultSig, authSig, err := e.signResult(pt, response)

	if err != nil {
		e.logger.Sugar().Errorw("Failed to sign result",
			zap.String("taskId", task.TaskId),
			zap.String("avsAddress", avsAddress),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "Failed to sign result %s", err.Error())
	}

	e.logger.Sugar().Infow("returning task result to aggregator",
		zap.String("taskId", task.TaskId),
		zap.String("avsAddress", avsAddress),
		zap.String("operatorAddress", e.config.Operator.Address),
	)

	result := &executorV1.TaskResult{
		TaskId:          response.TaskID,
		OperatorAddress: e.config.Operator.Address,
		OperatorSetId:   task.OperatorSetId,
		Output:          response.Result,
		ResultSignature: resultSig,
		AuthSignature:   authSig,
		AvsAddress:      avsAddress,
	}

	// Mark task as processed
	if err := e.store.MarkTaskProcessed(ctx, task.TaskId); err != nil {
		e.logger.Sugar().Warnw("Failed to mark task as processed",
			"taskId", task.TaskId,
			"error", err,
		)
		// Don't fail the task, just log the error
	}

	return result, nil
}

// signResult creates both result signature (for aggregation) and auth signature (for identity)
func (e *Executor) signResult(task *performerTask.PerformerTask, result *performerTask.PerformerTaskResult) ([]byte, []byte, error) {
	// Get the curve type for the operator set using the task's block number for historical accuracy
	curveType, err := e.l1ContractCaller.GetOperatorSetCurveType(task.Avs, task.OperatorSetId, task.TaskBlockNumber)
	if err != nil {
		e.logger.Error("Failed to get operator set curve type",
			zap.String("avsAddress", task.Avs),
			zap.Uint32("operatorSetId", task.OperatorSetId),
			zap.Uint64("blockNumber", task.TaskBlockNumber),
			zap.Error(err),
		)
		return nil, nil, fmt.Errorf("failed to get operator set curve type: %w", err)
	}

	// Calculate the bytes that need to be signed for the result signature
	// This uses the contract's certificate digest calculation
	var signedOverBytes []byte
	var signerToUse signer.ISigner

	if curveType == config.CurveTypeBN254 {
		if e.bn254Signer == nil {
			return nil, nil, fmt.Errorf("BN254 signer is not initialized")
		}
		signerToUse = e.bn254Signer

		// Use the contract's BN254 certificate digest calculation
		signedOverBytes, err = e.l1ContractCaller.CalculateBN254CertificateDigestBytes(
			context.Background(),
			task.ReferenceTimestamp,
			util.GetKeccak256Digest(result.Result),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to calculate BN254 certificate digest: %w", err)
		}

	} else if curveType == config.CurveTypeECDSA {
		if e.ecdsaSigner == nil {
			return nil, nil, fmt.Errorf("ECDSA signer is not initialized")
		}
		signerToUse = e.ecdsaSigner

		// Use the contract's ECDSA certificate digest calculation
		signedOverBytes, err = e.l1ContractCaller.CalculateECDSACertificateDigestBytes(
			context.Background(),
			task.ReferenceTimestamp,
			util.GetKeccak256Digest(result.Result),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to calculate ECDSA certificate digest: %w", err)
		}

	} else {
		return nil, nil, fmt.Errorf("unsupported curve type: %s", curveType)
	}

	// Step 1: Create the result signature by signing the certificate digest
	resultSig, err := signerToUse.SignMessageForSolidity(signedOverBytes)
	if err != nil {
		e.logger.Error("Failed to sign result",
			zap.String("taskId", task.TaskID),
			zap.String("avsAddress", task.Avs),
			zap.Error(err),
		)
		return nil, nil, fmt.Errorf("failed to sign result: %w", err)
	}

	// Step 2: Create and sign the auth signature (unique per operator)
	// This binds the operator's identity to the result signature
	resultSigDigest := util.GetKeccak256Digest(resultSig)
	authData := &types.AuthSignatureData{
		TaskId:          task.TaskID,
		AvsAddress:      task.Avs,
		OperatorAddress: e.config.Operator.Address,
		OperatorSetId:   task.OperatorSetId,
		ResultSigDigest: resultSigDigest,
	}
	authBytes := authData.ToSigningBytes()
	authSig, err := signerToUse.SignMessageForSolidity(authBytes)
	if err != nil {
		e.logger.Error("Failed to sign auth data",
			zap.String("taskId", task.TaskID),
			zap.String("avsAddress", task.Avs),
			zap.Error(err),
		)
		return nil, nil, fmt.Errorf("failed to sign auth data: %w", err)
	}

	// Return both signatures
	return resultSig, authSig, nil
}

// ListPerformers returns a list of all performers and their status
func (e *Executor) ListPerformers(ctx context.Context, req *executorV1.ListPerformersRequest) (*executorV1.ListPerformersResponse, error) {
	e.logger.Info("Received list performers request",
		zap.String("avsAddressFilter", req.GetAvsAddress()),
	)

	// Verify authentication
	if err := auth.HandleAuthError(e.verifyAuth(req.Auth)); err != nil {
		return nil, err
	}

	var allPerformers []*executorV1.Performer
	filterAddress := strings.ToLower(req.GetAvsAddress())

	// Iterate through all AVS performers
	e.avsPerformers.Range(func(key, value interface{}) bool {
		avsAddress := key.(string)
		avsServerPerformer := value.(avsPerformer.IAvsPerformer)

		// Apply filter if provided
		if filterAddress != "" && !strings.EqualFold(filterAddress, avsAddress) {
			return true // continue iteration
		}

		// Get performer info from the server performer
		performerInfos := avsServerPerformer.ListPerformers()

		// Convert each performer info to proto format
		for _, info := range performerInfos {
			allPerformers = append(allPerformers, e.performerInfoToProto(info))
		}
		return true // continue iteration
	})

	// Also include persisted performer states from storage
	persistedStates, err := e.store.ListPerformerStates(ctx)
	if err != nil {
		e.logger.Sugar().Warnw("Failed to list performer states from storage",
			"error", err,
		)
	} else {
		// Add persisted states that are not already in the live list
		for _, state := range persistedStates {
			// Apply filter if provided
			if filterAddress != "" && !strings.EqualFold(filterAddress, state.AvsAddress) {
				continue
			}

			// Check if this performer is already in the list
			found := false
			for _, perf := range allPerformers {
				if perf.PerformerId == state.PerformerId {
					found = true
					break
				}
			}

			if !found {
				// Convert persisted state to proto format
				allPerformers = append(allPerformers, &executorV1.Performer{
					PerformerId:        state.PerformerId,
					AvsAddress:         state.AvsAddress,
					Status:             state.Status,
					ArtifactRegistry:   state.ArtifactRegistry,
					ArtifactTag:        state.ArtifactTag,
					ArtifactDigest:     state.ArtifactDigest,
					ResourceHealthy:    state.ContainerHealthy,
					ApplicationHealthy: state.ApplicationHealthy,
					LastHealthCheck:    state.LastHealthCheck.Format(time.RFC3339),
					ContainerId:        state.ContainerId,
				})
			}
		}
	}

	e.logger.Info("Returning performer list",
		zap.Int("count", len(allPerformers)),
		zap.String("avsAddressFilter", req.GetAvsAddress()),
	)

	return &executorV1.ListPerformersResponse{
		Performers: allPerformers,
	}, nil
}

// RemovePerformer removes a performer from the executor
func (e *Executor) RemovePerformer(ctx context.Context, req *executorV1.RemovePerformerRequest) (*executorV1.RemovePerformerResponse, error) {
	e.logger.Info("Received remove performer request",
		zap.String("performerId", req.GetPerformerId()),
	)

	// Verify authentication
	if err := auth.HandleAuthError(e.verifyAuth(req.Auth)); err != nil {
		return &executorV1.RemovePerformerResponse{
			Success: false,
			Message: "Authentication failed",
		}, err
	}

	// Validate request
	if err := e.validateRemovePerformerRequest(req); err != nil {
		return &executorV1.RemovePerformerResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(codes.InvalidArgument, err.Error())
	}

	// Find the performerServer
	avsAddress, performerServer, err := e.findPerformerByID(req.GetPerformerId())
	if err != nil {
		e.logger.Warn("Performer not found for removal",
			zap.String("performerId", req.GetPerformerId()),
		)
		return &executorV1.RemovePerformerResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(codes.NotFound, err.Error())
	}

	// Remove the performerServer
	if err = performerServer.RemovePerformer(ctx, req.GetPerformerId()); err != nil {
		e.logger.Error("Failed to remove performerServer",
			zap.String("performerId", req.GetPerformerId()),
			zap.String("avsAddress", avsAddress),
			zap.Error(err),
		)
		return &executorV1.RemovePerformerResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to remove performerServer: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	e.logger.Info("Successfully removed performerServer",
		zap.String("performerId", req.GetPerformerId()),
		zap.String("avsAddress", avsAddress),
	)

	// Remove performer state from storage
	if err := e.store.DeletePerformerState(ctx, req.GetPerformerId()); err != nil {
		e.logger.Sugar().Warnw("Failed to delete performer state from storage",
			"error", err,
			"performerId", req.GetPerformerId(),
		)
	}

	return &executorV1.RemovePerformerResponse{
		Success: true,
		Message: fmt.Sprintf("Performer %s removed successfully", req.GetPerformerId()),
	}, nil
}

// validateRemovePerformerRequest validates the RemovePerformerRequest
func (e *Executor) validateRemovePerformerRequest(req *executorV1.RemovePerformerRequest) error {
	if req.GetPerformerId() == "" {
		return errors.New("performer ID is required")
	}
	return nil
}

// findPerformerByID finds a performer by ID across all AVS performers
func (e *Executor) findPerformerByID(performerID string) (string, avsPerformer.IAvsPerformer, error) {
	var foundAvsAddress string
	var foundPerformer avsPerformer.IAvsPerformer
	var found bool

	e.avsPerformers.Range(func(key, value interface{}) bool {
		avsAddress := key.(string)
		avsServerPerformer := value.(avsPerformer.IAvsPerformer)
		performerInfos := avsServerPerformer.ListPerformers()
		for _, info := range performerInfos {
			if info.PerformerID == performerID {
				foundAvsAddress = avsAddress
				foundPerformer = avsServerPerformer
				found = true
				return false // stop iteration
			}
		}
		return true // continue iteration
	})

	if found {
		return foundAvsAddress, foundPerformer, nil
	}
	return "", nil, fmt.Errorf("performer with ID %s not found", performerID)
}

// performerInfoToProto converts a PerformerMetadata to the protobuf Performer format
func (e *Executor) performerInfoToProto(info avsPerformer.PerformerMetadata) *executorV1.Performer {
	return &executorV1.Performer{
		PerformerId:        info.PerformerID,
		AvsAddress:         info.AvsAddress,
		Status:             string(info.Status),
		ArtifactRegistry:   info.ArtifactRegistry,
		ArtifactTag:        info.ArtifactTag,
		ArtifactDigest:     info.ArtifactDigest,
		ResourceHealthy:    info.ContainerHealthy,
		ApplicationHealthy: info.ApplicationHealthy,
		LastHealthCheck:    info.LastHealthCheck.Format(time.RFC3339),
		ContainerId:        info.ResourceID,
	}
}

// GetChallengeToken generates a new challenge token for authentication
func (e *Executor) GetChallengeToken(ctx context.Context, req *executorV1.GetChallengeTokenRequest) (*executorV1.GetChallengeTokenResponse, error) {
	e.logger.Info("Received get challenge token request",
		zap.String("operatorAddress", req.GetOperatorAddress()),
	)

	// Validate operator address
	if req.GetOperatorAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "operator address is required")
	}

	// Verify that the requested operator address matches our configured operator
	if !strings.EqualFold(req.GetOperatorAddress(), e.config.Operator.Address) {
		return nil, status.Error(codes.PermissionDenied, "operator address mismatch")
	}

	// Check if auth is enabled
	if e.authVerifier == nil {
		return nil, status.Error(codes.Unimplemented, "authentication is not enabled")
	}

	// Generate a new challenge token
	entry, err := e.authVerifier.GenerateChallengeToken(req.GetOperatorAddress())
	if err != nil {
		e.logger.Error("Failed to generate challenge token",
			zap.String("operatorAddress", req.GetOperatorAddress()),
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, "failed to generate challenge token")
	}

	return &executorV1.GetChallengeTokenResponse{
		ChallengeToken: entry.Token,
		ExpiresAt:      entry.ExpiresAt.Unix(),
	}, nil
}

// validateOperatorInSet checks if the operator is in the specified operator set
func (e *Executor) validateOperatorInSet(task *executorV1.TaskSubmission) error {
	opSet, err := e.l1ContractCaller.GetOperatorSetDetailsForOperator(
		common.HexToAddress(e.config.Operator.Address),
		task.GetAvsAddress(),
		task.OperatorSetId,
		task.TaskBlockNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to get operator set details: %w", err)
	}

	if opSet == nil {
		return fmt.Errorf("invalid task operator set")
	}

	return nil
}

// constructTaskSubmissionMessage creates the message to be signed for a task submission
func (e *Executor) constructTaskSubmissionMessage(task *executorV1.TaskSubmission) []byte {
	// IMPORTANT: We always use the executor configured address, not what's in the task
	// This binds the signature to this specific executor
	return util.EncodeTaskSubmissionMessage(
		task.TaskId,
		task.AvsAddress,
		e.config.Operator.Address,
		task.OperatorSetId,
		task.TaskBlockNumber,
		task.Payload,
	)
}

// validateTaskSignature validates the signature of a task submission
func (e *Executor) validateTaskSignature(task *executorV1.TaskSubmission) error {
	// Get AVS config to find aggregator's operator set using historical block number
	aggConfig, err := e.l1ContractCaller.GetAVSConfig(task.AvsAddress, task.TaskBlockNumber)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to get AVS config",
			zap.String("avsAddress", task.AvsAddress),
			zap.Uint64("blockNumber", task.TaskBlockNumber),
			zap.Error(err),
		)
		return fmt.Errorf("invalid AVS config: %w", err)
	}
	if aggConfig == nil {
		return fmt.Errorf("avs config not found for avs")
	}

	// Get aggregator's operator set details using the task block number
	aggOpSet, err := e.l1ContractCaller.GetOperatorSetDetailsForOperator(
		common.HexToAddress(task.AggregatorAddress),
		task.AvsAddress,
		aggConfig.AggregatorOperatorSetId,
		task.TaskBlockNumber,
	)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to get aggregator operator set",
			zap.String("aggregatorAddress", task.AggregatorAddress),
			zap.String("avsAddress", task.AvsAddress),
			zap.Uint32("operatorSetId", aggConfig.AggregatorOperatorSetId),
			zap.Error(err),
		)
		return fmt.Errorf("invalid aggregator operator set: %w", err)
	}
	if aggOpSet == nil {
		return fmt.Errorf("aggregator operator set not found for aggregator")
	}

	// Create signing scheme based on curve type
	var scheme signing.SigningScheme
	switch aggOpSet.CurveType {
	case config.CurveTypeBN254:
		scheme = bn254.NewScheme()
	case config.CurveTypeECDSA:
		scheme = ecdsa.NewScheme()
	default:
		e.logger.Sugar().Errorw("Unsupported curve type",
			zap.String("curveType", aggOpSet.CurveType.String()),
		)
		return fmt.Errorf("unsupported curve type: %s", aggOpSet.CurveType)
	}

	// Parse signature
	sig, err := scheme.NewSignatureFromBytes(task.Signature)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to parse signature",
			zap.Error(err),
		)
		return fmt.Errorf("invalid signature format: %w", err)
	}

	// Create the message that should have been signed
	messageToVerify := e.constructTaskSubmissionMessage(task)

	// Verify signature based on curve type
	var verified bool
	switch aggOpSet.CurveType {
	case config.CurveTypeBN254:
		verified, err = sig.Verify(aggOpSet.WrappedPublicKey.PublicKey, messageToVerify)
		if err != nil {
			e.logger.Sugar().Errorw("Error verifying BN254 signature",
				zap.String("aggregatorAddress", task.AggregatorAddress),
				zap.Error(err),
			)
			return fmt.Errorf("signature verification failed: %w", err)
		}
	case config.CurveTypeECDSA:
		typedSig, err := ecdsa.NewSignatureFromBytes(sig.Bytes())
		if err != nil {
			e.logger.Sugar().Errorw("Failed to create ECDSA signature",
				zap.Error(err),
			)
			return fmt.Errorf("failed to create ECDSA signature: %w", err)
		}
		// ECDSA verification needs the hash of the message
		messageHash := util.GetKeccak256Digest(messageToVerify)
		verified, err = typedSig.VerifyWithAddress(messageHash[:], aggOpSet.WrappedPublicKey.ECDSAAddress)
		if err != nil {
			e.logger.Sugar().Errorw("Error verifying ECDSA signature",
				zap.String("aggregatorAddress", task.AggregatorAddress),
				zap.Error(err),
			)
			return fmt.Errorf("signature verification failed: %w", err)
		}
	}

	if !verified {
		e.logger.Sugar().Errorw("Signature verification failed",
			zap.String("taskId", task.TaskId),
			zap.String("aggregatorAddress", task.AggregatorAddress),
			zap.String("executorAddress", e.config.Operator.Address),
		)
		return fmt.Errorf("signature verification failed")
	}

	e.logger.Sugar().Infow("Task signature verified successfully",
		zap.String("taskId", task.TaskId),
		zap.String("aggregatorAddress", task.AggregatorAddress),
		zap.Uint32("operatorSetId", aggOpSet.OperatorSetID),
	)

	return nil
}
