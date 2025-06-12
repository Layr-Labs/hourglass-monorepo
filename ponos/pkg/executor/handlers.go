package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/serverPerformer"
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

	// Create container manager for the deployment
	containerMgr, err := containerManager.NewDefaultDockerContainerManager(e.logger)
	if err != nil {
		e.logger.Error("Failed to create container manager for deployment",
			zap.String("avsAddress", avsAddress),
			zap.Error(err),
		)
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create container manager: %v", err),
		}, status.Error(codes.Internal, fmt.Sprintf("Failed to create container manager: %v", err))
	}

	// Find or create the AVS performer
	performer, ok := e.avsPerformers[avsAddress]
	if !ok {
		// Create new AVS performer for this address
		e.logger.Info("Creating new AVS performer for address", zap.String("avsAddress", avsAddress))

		config := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           avsAddress,
			ProcessType:          avsPerformer.AvsProcessTypeServer,
			Image:                avsPerformer.PerformerImage{Repository: req.GetRegistryUrl(), Tag: req.GetDigest()},
			WorkerCount:          1,
			PerformerNetworkName: e.config.PerformerNetworkName,
			SigningCurve:         "bn254",
		}

		newPerformer := serverPerformer.NewAvsPerformerServer(
			config,
			e.peeringFetcher,
			e.logger,
			containerMgr,
		)

		// Add to performers map
		e.avsPerformers[avsAddress] = newPerformer
		performer = newPerformer
	}

	// Cast to server performer to access deployment functionality
	serverPerformer, ok := performer.(*serverPerformer.AvsPerformerServer)
	if !ok {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: "Deployment is only supported for server-type performers",
		}, status.Error(codes.Unimplemented, "Deployment is only supported for server-type performers")
	}

	// Parse the registry URL and digest to create image reference
	// For now, assume the digest format is sha256:xxxx and the registry URL contains the repository
	imageRef := fmt.Sprintf("%s@%s", req.RegistryUrl, req.Digest)

	// Split the image reference to get repository and tag/digest
	parts := strings.Split(imageRef, "@")
	if len(parts) != 2 {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: "Invalid image reference format",
		}, status.Error(codes.InvalidArgument, "Invalid image reference format")
	}

	repository := parts[0]
	digest := parts[1]

	// Create AVS performer config for deployment
	deployConfig := &avsPerformer.AvsPerformerConfig{
		AvsAddress:  avsAddress,
		ProcessType: avsPerformer.AvsProcessTypeServer,
		Image: avsPerformer.PerformerImage{
			Repository: repository,
			Tag:        digest, // TODO: revisit this naming
		},
		WorkerCount:          1,
		PerformerNetworkName: e.config.PerformerNetworkName,
		SigningCurve:         "bn254",
	}

	// Deploy the container
	err = serverPerformer.DeployContainer(ctx, avsAddress, deployConfig, containerMgr)
	if err != nil {
		e.logger.Error("Failed to deploy container",
			zap.String("avsAddress", avsAddress),
			zap.String("imageRef", imageRef),
			zap.Error(err),
		)
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to deploy container: %v", err),
		}, status.Error(codes.Internal, fmt.Sprintf("Failed to deploy container: %v", err))
	}

	deploymentId := fmt.Sprintf("%s-%d", avsAddress, time.Now().Unix())

	e.logger.Info("Artifact deployment started successfully",
		zap.String("avsAddress", avsAddress),
		zap.String("deploymentId", deploymentId),
		zap.String("imageRef", imageRef),
	)

	return &executorV1.DeployArtifactResponse{
		Success:      true,
		Message:      "Artifact deployment started successfully",
		DeploymentId: deploymentId,
	}, nil
}

func (e *Executor) handleReceivedTask(task *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	e.logger.Sugar().Infow("Received task from AVS avsPerformer",
		"taskId", task.TaskId,
		"avsAddress", task.AvsAddress,
	)
	avsAddress := strings.ToLower(task.GetAvsAddress())
	if avsAddress == "" {
		return nil, fmt.Errorf("AVS address is empty")
	}

	avsPerformer, ok := e.avsPerformers[task.AvsAddress]
	if !ok {
		return nil, fmt.Errorf("AVS avsPerformer not found for address %s", task.AvsAddress)
	}

	pt := performerTask.NewPerformerTaskFromTaskSubmissionProto(task)

	if err := avsPerformer.ValidateTaskSignature(pt); err != nil {
		return nil, fmt.Errorf("failed to validate task signature: %w", err)
	}
	e.inflightTasks.Store(task.TaskId, task)

	response, err := avsPerformer.RunTask(context.Background(), pt)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to run task",
			"taskId", task.TaskId,
			"avsAddress", task.AvsAddress,
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "Failed to run task %s", err.Error())
	}

	sig, err := e.signResult(response)
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
	}, nil
}

func (e *Executor) signResult(result *performerTask.PerformerTaskResult) ([]byte, error) {
	// Generate a keccak256 hash of the result so that our signature is fixed in size.
	// This is for compatibility with the certificate verifier.
	digestBytes := util.GetKeccak256Digest(result.Result)

	return e.signer.SignMessage(digestBytes[:])
}
