package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/serverPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const deploymentTimeout = 1 * time.Minute

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

	// Check if a deployment is already in progress for this AVS
	if _, exists := e.activeDeployments.LoadOrStore(avsAddress, true); exists {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: "Deployment already in progress for this AVS",
		}, status.Error(codes.AlreadyExists, "deployment already in progress")
	}

	// Ensure we clean up the active deployment marker on exit
	defer e.activeDeployments.Delete(avsAddress)

	// Find or create the AVS performer
	performer, err := e.getOrCreateAvsPerformer(ctx, avsAddress)
	if err != nil {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get or create AVS performer: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	performerImage := avsPerformer.PerformerImage{
		Repository: req.GetRegistryUrl(),
		Tag:        req.GetDigest(),
	}

	// Deploy the container and get status channel
	statusChan, err := performer.DeployContainer(ctx, avsAddress, performerImage)
	if err != nil {
		e.logger.Error("Failed to deploy container",
			zap.String("avsAddress", avsAddress),
			zap.String("registryUrl", performerImage.Repository),
			zap.String("digest", performerImage.Tag),
			zap.Error(err),
		)
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to deploy container: %v", err),
		}, status.Error(codes.Internal, fmt.Sprintf("Failed to deploy container: %v", err))
	}

	deploymentId := fmt.Sprintf("%s-%d", avsAddress, time.Now().Unix())

	e.logger.Info("Container deployment started, monitoring for events",
		zap.String("avsAddress", avsAddress),
		zap.String("deploymentId", deploymentId),
		zap.String("registryUrl", performerImage.Repository),
		zap.String("digest", performerImage.Tag),
	)

	return e.processDeploymentEvents(ctx, performer, statusChan, avsAddress, deploymentId)
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

	// Create config without image info - will be deployed via DeployContainer
	config := &avsPerformer.AvsPerformerConfig{
		AvsAddress:           avsAddress,
		ProcessType:          avsPerformer.AvsProcessTypeServer,
		Image:                avsPerformer.PerformerImage{},
		WorkerCount:          1,
		PerformerNetworkName: e.config.PerformerNetworkName,
		SigningCurve:         "bn254",
	}

	newPerformer := serverPerformer.NewAvsPerformerServer(
		config,
		e.peeringFetcher,
		e.logger,
		e.containerMgr,
	)

	// Initialize the performer (fetches aggregator peers, but won't create container)
	if err := newPerformer.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize AVS performer: %w", err)
	}

	// Add to performers map
	e.avsPerformers[avsAddress] = newPerformer

	return newPerformer, nil
}

// processDeploymentEvents monitors the deployment status channel and handles container health events
func (e *Executor) processDeploymentEvents(
	ctx context.Context,
	performer avsPerformer.IAvsPerformer,
	statusChan <-chan avsPerformer.PerformerStatusEvent,
	avsAddress string,
	deploymentId string,
) (*executorV1.DeployArtifactResponse, error) {
	// Wait for deployment status with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, deploymentTimeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			e.logger.Error("Deployment timeout waiting for container to become healthy",
				zap.String("avsAddress", avsAddress),
				zap.String("deploymentId", deploymentId),
				zap.Duration("timeout", deploymentTimeout),
			)

			// Cancel the deployment to clean up resources
			if cancelErr := performer.CancelDeployment(ctx); cancelErr != nil {
				e.logger.Error("Failed to cancel deployment after timeout",
					zap.String("avsAddress", avsAddress),
					zap.String("deploymentId", deploymentId),
					zap.Error(cancelErr),
				)
			}

			return &executorV1.DeployArtifactResponse{
				Success: false,
				Message: fmt.Sprintf("Deployment timeout: container failed to become healthy within %v", deploymentTimeout),
			}, status.Error(codes.DeadlineExceeded, "deployment timeout")

		case statusEvent, ok := <-statusChan:
			if !ok {
				// Channel closed unexpectedly
				e.logger.Error("Status channel closed unexpectedly",
					zap.String("avsAddress", avsAddress),
					zap.String("deploymentId", deploymentId),
				)

				// Cancel the deployment to clean up
				if cancelErr := performer.CancelDeployment(ctx); cancelErr != nil {
					e.logger.Error("Failed to cancel deployment after promotion failure",
						zap.String("avsAddress", avsAddress),
						zap.String("deploymentId", deploymentId),
						zap.Error(cancelErr),
					)
				}

				return &executorV1.DeployArtifactResponse{
					Success: false,
					Message: "Deployment monitoring failed",
				}, status.Error(codes.Internal, "deployment monitoring failed")
			}

			e.logger.Info("Received deployment status update",
				zap.String("avsAddress", avsAddress),
				zap.String("deploymentId", deploymentId),
				zap.String("status", fmt.Sprintf("%v", statusEvent.Status)),
				zap.String("message", statusEvent.Message),
			)

			switch statusEvent.Status {
			case avsPerformer.PerformerHealthy:
				e.logger.Info("Container deployment completed successfully and is healthy",
					zap.String("avsAddress", avsAddress),
					zap.String("deploymentId", deploymentId),
					zap.String("containerID", statusEvent.ContainerID),
				)

				// Promote the healthy container to current
				if promoteErr := performer.PromoteContainer(ctx); promoteErr != nil {
					e.logger.Error("Failed to promote healthy container",
						zap.String("avsAddress", avsAddress),
						zap.String("deploymentId", deploymentId),
						zap.Error(promoteErr),
					)

					// Cancel the deployment to clean up
					if cancelErr := performer.CancelDeployment(ctx); cancelErr != nil {
						e.logger.Error("Failed to cancel deployment after promotion failure",
							zap.String("avsAddress", avsAddress),
							zap.String("deploymentId", deploymentId),
							zap.Error(cancelErr),
						)
					}

					return &executorV1.DeployArtifactResponse{
						Success: false,
						Message: fmt.Sprintf("Failed to promote container: %v", promoteErr),
					}, status.Error(codes.Internal, "container promotion failed")
				}

				e.logger.Info("Container promoted successfully",
					zap.String("avsAddress", avsAddress),
					zap.String("deploymentId", deploymentId),
					zap.String("containerID", statusEvent.ContainerID),
				)

				return &executorV1.DeployArtifactResponse{
					Success:      true,
					Message:      "Container deployment completed successfully and is healthy",
					DeploymentId: deploymentId,
				}, nil

			case avsPerformer.PerformerUnhealthy:
				e.logger.Error("Container deployment container is unhealthy",
					zap.String("avsAddress", avsAddress),
					zap.String("deploymentId", deploymentId),
					zap.String("containerID", statusEvent.ContainerID),
					zap.String("reason", statusEvent.Message),
				)
			}
		}
	}
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

	return e.signer.SignMessageForSolidity(digestBytes)
}
