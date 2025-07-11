package executor

import (
	"context"
	"errors"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"strings"
	"time"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/avsContainerPerformer"
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

	// Deploy using the performer's Deploy method
	image := avsPerformer.PerformerImage{
		Repository: req.GetRegistryUrl(),
		Digest:     req.GetDigest(),
		Envs: util.Map(req.GetEnv(), func(env *executorV1.PerformerEnv, i uint64) config.AVSPerformerEnv {
			return config.AVSPerformerEnv{
				Name:         env.GetName(),
				Value:        env.GetValue(),
				ValueFromEnv: env.GetValueFromEnv(),
			}
		}),
	}

	result, err := performer.Deploy(ctx, image)
	if err != nil {
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

	return &executorV1.DeployArtifactResponse{
		Success:      true,
		Message:      result.Message,
		DeploymentId: result.ID,
	}, nil
}

// getOrCreateAvsPerformer gets an existing AVS performer or creates a new one if it doesn't exist
func (e *Executor) getOrCreateAvsPerformer(ctx context.Context, avsAddress string) (avsPerformer.IAvsPerformer, error) {
	// Check if performer already exists
	performer, ok := e.avsPerformers[avsAddress]
	if ok {
		return performer, nil
	}

	// Create new AVS performer for this address
	e.logger.Info("Creating new AVS performer for address", zap.String("avsAddress", avsAddress))

	// Create config without image info - will be deployed via Deploy method
	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress:           avsAddress,
		ProcessType:          avsPerformer.AvsProcessTypeServer,
		Image:                avsPerformer.PerformerImage{},
		WorkerCount:          1,
		PerformerNetworkName: e.config.PerformerNetworkName,
		SigningCurve:         "bn254",
	}

	newPerformer, err := avsContainerPerformer.NewAvsContainerPerformer(
		config,
		e.peeringFetcher,
		e.l1ContractCaller,
		e.logger,
	)
	if err != nil {
		return nil, err
	}

	// Initialize the performer (fetches aggregator peers, but won't create container)
	if err := newPerformer.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize AVS performer: %w", err)
	}

	// Add to performers map
	e.avsPerformers[avsAddress] = newPerformer

	return newPerformer, nil
}

func (e *Executor) handleReceivedTask(task *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	e.logger.Sugar().Infow("Received task from AVS avsPerf",
		"taskId", task.TaskId,
		"avsAddress", task.AvsAddress,
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
	return &executorV1.TaskResult{
		TaskId:          response.TaskID,
		OperatorAddress: e.config.Operator.Address,
		Output:          response.Result,
		Signature:       sig,
		AvsAddress:      task.AvsAddress,
		OutputDigest:    digest[:],
	}, nil
}

// TODO(seanmcgary): since curves are oriented around the operator set ID. If we're dealing with an ECDSA key, we need
// to call `calculateCertificateDigest(cert.referenceTimestamp, cert.messageHash)` to generate the digest

// signResult signs the result of a task and returns the signature and the digest.
func (e *Executor) signResult(task *performerTask.PerformerTask, result *performerTask.PerformerTaskResult) ([]byte, [32]byte, error) {
	// Generate a keccak256 hash of the result so that our signature is fixed in size.
	// This is for compatibility with the certificate verifier.
	var digestBytes [32]byte
	digestBytes = util.GetKeccak256Digest(result.Result)

	curveType, err := e.l1ContractCaller.GetOperatorSetCurveType(task.Avs, task.OperatorSetId)
	if err != nil {
		e.logger.Error("Failed to get operator set curve type",
			zap.String("avsAddress", task.Avs),
			zap.Uint32("operatorSetId", task.OperatorSetId),
			zap.Error(err),
		)
		return nil, digestBytes, fmt.Errorf("failed to get operator set curve type: %w", err)
	}

	// ecdsa is a special snowflake and requires a different digest calculation
	if curveType == config.CurveTypeECDSA {
		digest, err := e.l1ContractCaller.CalculateECDSACertificateDigest(
			context.Background(),
			task.ReferenceTimestamp,
			digestBytes,
		)
		if err != nil {
			return nil, digestBytes, fmt.Errorf("failed to calculate ECDSA certificate digest: %w", err)
		}
		digestBytes = digest
	}

	sig, err := e.signer.SignMessageForSolidity(digestBytes)
	if err != nil {
		e.logger.Error("Failed to sign result",
			zap.String("taskId", task.TaskID),
			zap.String("avsAddress", task.Avs),
			zap.Error(err),
		)
		return nil, digestBytes, fmt.Errorf("failed to sign result: %w", err)
	}
	return sig, digestBytes, nil
}

// ListPerformers returns a list of all performers and their status
func (e *Executor) ListPerformers(_ context.Context, req *executorV1.ListPerformersRequest) (*executorV1.ListPerformersResponse, error) {
	e.logger.Info("Received list performers request",
		zap.String("avsAddressFilter", req.GetAvsAddress()),
	)

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
