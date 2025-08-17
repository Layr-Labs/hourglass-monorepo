package executor

import (
	"context"
	"errors"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"strings"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (e *Executor) SubmitTask(ctx context.Context, req *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	res, err := e.handleReceivedTask(req)
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
	image := req.GetImage()
	if image == nil {
		return "Image is required"
	}
	if image.GetRegistry() == "" {
		return "Image registry is required"
	}
	if image.GetDigest() == "" && image.GetTag() == "" {
		return "Either artifact digest or tag is required"
	}
	return ""
}

func (e *Executor) convertPerformerK8sConfigFromProto(k8sCfg *executorV1.KubernetesConfig) *executorConfig.AvsPerformerKubernetesConfig {
	if k8sCfg == nil {
		return nil
	}

	return &executorConfig.AvsPerformerKubernetesConfig{
		ServiceAccountName: k8sCfg.GetServiceAccountName(),
	}
}

func (e *Executor) convertEnvsFromProto(reqEnvs []*executorV1.PerformerEnv) []config.AVSPerformerEnv {
	if reqEnvs == nil {
		return nil
	}

	envs := make([]config.AVSPerformerEnv, 0, len(reqEnvs))

	for _, reqE := range reqEnvs {
		e := config.AVSPerformerEnv{
			Name:         reqE.GetName(),
			Value:        reqE.GetValue(),
			ValueFromEnv: reqE.GetValueFromEnv(),
		}
		k8sEnv := reqE.GetKubernetesEnv()
		if k8sEnv != nil && k8sEnv.GetValueFrom() != nil {
			valueFrom := k8sEnv.GetValueFrom()
			e.KubernetesEnv = &config.KubernetesEnv{
				ValueFrom: config.ValueFrom{},
			}
			if valueFrom.GetSecretKeyRef() != nil {
				e.KubernetesEnv.ValueFrom.SecretKeyRef = config.SecretKeyRef{
					Name: valueFrom.GetSecretKeyRef().GetName(),
					Key:  valueFrom.GetSecretKeyRef().GetKey(),
				}
			}
			if valueFrom.GetConfigMapKeyRef() != nil {
				e.KubernetesEnv.ValueFrom.ConfigMapKeyRef = config.ConfigMapKeyRef{
					Name: valueFrom.GetConfigMapKeyRef().GetName(),
					Key:  valueFrom.GetConfigMapKeyRef().GetKey(),
				}
			}
		}
		envs = append(envs, e)
	}
	return envs
}

func (e *Executor) DeployArtifact(ctx context.Context, req *executorV1.DeployArtifactRequest) (*executorV1.DeployArtifactResponse, error) {
	e.logger.Info("Received deploy artifact request",
		zap.String("avsAddress", req.AvsAddress),
		zap.String("digest", req.GetImage().GetRegistry()),
		zap.String("registryUrl", req.GetImage().GetDigest()),
	)

	// Verify authentication
	if err := e.verifyAuth(req.Auth); err != nil {
		// Preserve the original status code if it's already a status error
		if s, ok := status.FromError(err); ok {
			return &executorV1.DeployArtifactResponse{
				Success: false,
				Message: "Authentication failed",
			}, status.Error(s.Code(), s.Message())
		}
		// Fallback to Unauthenticated for non-status errors
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: "Authentication failed",
		}, status.Error(codes.Unauthenticated, err.Error())
	}

	// Validate request
	if errMsg := validateDeployArtifactRequest(req); errMsg != "" {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: errMsg,
		}, status.Error(codes.InvalidArgument, errMsg)
	}

	avsAddress := strings.ToLower(req.AvsAddress)

	// Create deployment record in storage
	deploymentId := fmt.Sprintf("deployment-%s-%d", avsAddress, time.Now().UnixNano())
	deploymentInfo := &storage.DeploymentInfo{
		DeploymentId:     deploymentId,
		AvsAddress:       avsAddress,
		ArtifactRegistry: req.GetImage().GetRegistry(),
		ArtifactDigest:   req.GetImage().GetDigest(),
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

	// Find or create the AVS performer
	performerConfig := &executorConfig.AvsPerformerConfig{
		Image: &executorConfig.PerformerImage{
			Repository: req.GetImage().GetRegistry(),
			Tag:        req.GetImage().GetTag(),
			Digest:     req.GetImage().GetDigest(),
		},
		ProcessType: req.GetProcessType(),
		AvsAddress:  req.GetAvsAddress(),
		Envs:        e.convertEnvsFromProto(req.GetEnvs()),
		Kubernetes:  e.convertPerformerK8sConfigFromProto(req.GetKubernetes()),
	}

	// Update deployment status to deploying
	if err := e.store.UpdateDeploymentStatus(ctx, deploymentId, storage.DeploymentStatusDeploying); err != nil {
		e.logger.Sugar().Warnw("Failed to update deployment status",
			"error", err,
			"deploymentId", deploymentId,
		)
	}

	// create and deploy
	_, result, err := e.getOrCreateAvsPerformer(ctx, performerConfig)
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
			zap.String("registryUrl", req.GetImage().GetRegistry()),
			zap.String("digest", req.GetImage().GetDigest()),
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

	return &executorV1.DeployArtifactResponse{
		Success:      true,
		Message:      result.Message,
		DeploymentId: result.ID,
	}, nil
}

// getOrCreateAvsPerformer gets an existing AVS performer or creates a new one if it doesn't exist
func (e *Executor) getOrCreateAvsPerformer(ctx context.Context, avsConfig *executorConfig.AvsPerformerConfig) (avsPerformer.IAvsPerformer, *avsPerformer.DeploymentResult, error) {
	avsAddress := strings.ToLower(avsConfig.AvsAddress)
	// Check if performer already exists
	performer, ok := e.avsPerformers[avsAddress]
	if ok {
		return performer, nil, nil
	}

	// Create new AVS performer for this address
	e.logger.Info("Creating new AVS performer for address", zap.String("avsAddress", avsAddress))

	avsConfig.PerformerNetworkName = e.config.PerformerNetworkName

	newPerformer, result, err := e.createAndInitializePerformer(ctx, avsConfig)

	if err != nil {
		e.logger.Error("Failed to create and initialize AVS performer",
			zap.String("avsAddress", avsAddress),
			zap.Error(err),
		)
		return nil, nil, fmt.Errorf("failed to create AVS performer: %w", err)
	}

	return newPerformer, result, nil
}

func (e *Executor) handleReceivedTask(task *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	e.logger.Sugar().Infow("Received task from AVS avsPerf",
		zap.String("taskId", task.TaskId),
		zap.String("avsAddress", task.AvsAddress),
		zap.String("operatorAddress", e.config.Operator.Address),
		zap.String("aggregatorAddress", task.GetAggregatorAddress()),
	)
	avsAddress := strings.ToLower(task.GetAvsAddress())
	if avsAddress == "" {
		return nil, fmt.Errorf("AVS address is empty")
	}

	avsPerf, ok := e.avsPerformers[task.AvsAddress]
	if !ok {
		return nil, fmt.Errorf("AVS avsPerf not found for address %s", task.AvsAddress)
	}

	pt := performerTask.NewPerformerTaskFromTaskSubmissionProto(task)

	if err := avsPerf.ValidateTaskSignature(pt); err != nil {
		return nil, fmt.Errorf("failed to validate task signature: %w", err)
	}
	e.inflightTasks.Store(task.TaskId, task)

	// Save inflight task to storage
	taskInfo := &storage.TaskInfo{
		TaskId:            task.TaskId,
		AvsAddress:        task.AvsAddress,
		OperatorAddress:   e.config.Operator.Address,
		ReceivedAt:        time.Now(),
		Status:            "processing",
		AggregatorAddress: task.GetAggregatorAddress(),
		OperatorSetId:     task.OperatorSetId,
	}
	if err := e.store.SaveInflightTask(context.Background(), task.TaskId, taskInfo); err != nil {
		e.logger.Sugar().Warnw("Failed to save inflight task to storage",
			"error", err,
			"taskId", task.TaskId,
		)
	}

	response, err := avsPerf.RunTask(context.Background(), pt)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to run task",
			"taskId", task.TaskId,
			"avsAddress", task.AvsAddress,
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "Failed to run task %s", err.Error())
	}

	sig, digest, err := e.signResult(pt, response)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to sign result",
			zap.String("taskId", task.TaskId),
			zap.String("avsAddress", task.AvsAddress),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "Failed to sign result %s", err.Error())
	}

	e.logger.Sugar().Infow("returning task result to aggregator",
		zap.String("taskId", task.TaskId),
		zap.String("avsAddress", task.AvsAddress),
		zap.String("operatorAddress", e.config.Operator.Address),
		zap.String("signature", string(sig)),
	)

	e.inflightTasks.Delete(task.TaskId)

	// Remove inflight task from storage
	if err := e.store.DeleteInflightTask(context.Background(), task.TaskId); err != nil {
		e.logger.Sugar().Warnw("Failed to delete inflight task from storage",
			"error", err,
			"taskId", task.TaskId,
		)
	}

	return &executorV1.TaskResult{
		TaskId:          response.TaskID,
		OperatorAddress: e.config.Operator.Address,
		Output:          response.Result,
		Signature:       sig,
		AvsAddress:      task.AvsAddress,
		OutputDigest:    digest[:],
	}, nil
}

// signResult signs the result of a task and returns the signature and the digest.
func (e *Executor) signResult(task *performerTask.PerformerTask, result *performerTask.PerformerTaskResult) ([]byte, []byte, error) {
	// the bytes of the result that we need to sign over.
	// the ecdsaSigner will end up hashing this to get the digest, so
	// this value is the raw []bytes that get hashed
	var signedOverBytes []byte

	curveType, err := e.l1ContractCaller.GetOperatorSetCurveType(task.Avs, task.OperatorSetId)
	if err != nil {
		e.logger.Error("Failed to get operator set curve type",
			zap.String("avsAddress", task.Avs),
			zap.Uint32("operatorSetId", task.OperatorSetId),
			zap.Error(err),
		)
		return nil, signedOverBytes, fmt.Errorf("failed to get operator set curve type: %w", err)
	}

	var signerToUse signer.ISigner
	if curveType == config.CurveTypeBN254 {
		if e.bn254Signer == nil {
			return nil, signedOverBytes, fmt.Errorf("BN254 signer is not initialized")
		}
		signerToUse = e.bn254Signer

		signedOverBytes = result.Result
	} else if curveType == config.CurveTypeECDSA {
		if e.ecdsaSigner == nil {
			return nil, signedOverBytes, fmt.Errorf("ECDSA signer is not initialized")
		}
		signerToUse = e.ecdsaSigner

		digestBytes := util.GetKeccak256Digest(result.Result)
		// ecdsa is a special snowflake and requires an EIP-712 digest calculation
		digest, err := e.l1ContractCaller.CalculateECDSACertificateDigestBytes(
			context.Background(),
			task.ReferenceTimestamp,
			digestBytes,
		)
		if err != nil {
			return nil, signedOverBytes, fmt.Errorf("failed to calculate ECDSA certificate digest: %w", err)
		}
		signedOverBytes = digest
	} else {
		return nil, signedOverBytes, fmt.Errorf("unsupported curve type: %s", curveType)
	}

	sig, err := signerToUse.SignMessageForSolidity(signedOverBytes)
	if err != nil {
		e.logger.Error("Failed to sign result",
			zap.String("taskId", task.TaskID),
			zap.String("avsAddress", task.Avs),
			zap.Error(err),
		)
		return nil, signedOverBytes, fmt.Errorf("failed to sign result: %w", err)
	}
	// signResult() is expected to return the digest of what was signed over.
	// We do this as the very last thing since some signing backends hash the raw bytes themselves
	// but dont return the digest.
	signedOverDigest := util.GetKeccak256Digest(signedOverBytes)
	return sig, signedOverDigest[:], nil
}

// ListPerformers returns a list of all performers and their status
func (e *Executor) ListPerformers(ctx context.Context, req *executorV1.ListPerformersRequest) (*executorV1.ListPerformersResponse, error) {
	e.logger.Info("Received list performers request",
		zap.String("avsAddressFilter", req.GetAvsAddress()),
	)

	// Verify authentication
	if err := e.verifyAuth(req.Auth); err != nil {
		// Preserve the original status code if it's already a status error
		if s, ok := status.FromError(err); ok {
			return nil, status.Error(s.Code(), s.Message())
		}
		// Fallback to Unauthenticated for non-status errors
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	var allPerformers []*executorV1.Performer
	filterAddress := strings.ToLower(req.GetAvsAddress())

	// Iterate through all AVS performers
	for avsAddress, avsServerPerformer := range e.avsPerformers {
		// Apply filter if provided
		if filterAddress != "" && !strings.EqualFold(filterAddress, avsAddress) {
			continue
		}

		// Get performer info from the server performer
		performerInfos := avsServerPerformer.ListPerformers()

		// Convert each performer info to proto format
		for _, info := range performerInfos {
			allPerformers = append(allPerformers, e.performerInfoToProto(info))
		}
	}

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
	if err := e.verifyAuth(req.Auth); err != nil {
		// Preserve the original status code if it's already a status error
		if s, ok := status.FromError(err); ok {
			return &executorV1.RemovePerformerResponse{
				Success: false,
				Message: "Authentication failed",
			}, status.Error(s.Code(), s.Message())
		}
		// Fallback to Unauthenticated for non-status errors
		return &executorV1.RemovePerformerResponse{
			Success: false,
			Message: "Authentication failed",
		}, status.Error(codes.Unauthenticated, err.Error())
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
	for avsAddress, avsServerPerformer := range e.avsPerformers {
		performerInfos := avsServerPerformer.ListPerformers()
		for _, info := range performerInfos {
			if info.PerformerID == performerID {
				return avsAddress, avsServerPerformer, nil
			}
		}
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
