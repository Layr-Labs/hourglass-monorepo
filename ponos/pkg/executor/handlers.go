package executor

import (
	"context"
	"fmt"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/serverPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"time"
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

// DeployArtifact deploys a new artifact to an AVS performer
func (e *Executor) DeployArtifact(ctx context.Context, req *executorV1.DeployArtifactRequest) (*executorV1.DeployArtifactResponse, error) {
	e.logger.Info("Received deploy artifact request",
		zap.String("avsAddress", req.AvsAddress),
		zap.String("registryUrl", req.RegistryUrl),
		zap.String("digest", req.Digest),
	)

	// Find the performer for this AVS
	performer := e.getPerformerByAVS(req.AvsAddress)
	if performer == nil {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: "performer not found for AVS",
		}, nil
	}

	// Cast to serverPerformer to access deployment methods
	serverPerf, ok := performer.(*serverPerformer.AvsPerformerServer)
	if !ok {
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: "performer does not support deployments",
		}, nil
	}

	// Deploy with 30 second delay for validation
	activationTime := time.Now().Unix() + 30

	deploymentID, err := serverPerf.DeployNewPerformerVersion(ctx, req.RegistryUrl, req.Digest, activationTime)
	if err != nil {
		e.logger.Error("Failed to deploy artifact",
			zap.String("avsAddress", req.AvsAddress),
			zap.Error(err),
		)
		return &executorV1.DeployArtifactResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	e.logger.Info("Successfully scheduled artifact deployment",
		zap.String("avsAddress", req.AvsAddress),
		zap.String("deploymentId", deploymentID),
		zap.Int64("activationTime", activationTime),
	)

	return &executorV1.DeployArtifactResponse{
		Success:      true,
		Message:      "deployment scheduled successfully",
		DeploymentId: deploymentID,
	}, nil
}

// ListPerformers returns a list of all performers and their status
func (e *Executor) ListPerformers(ctx context.Context, req *executorV1.ListPerformersRequest) (*executorV1.ListPerformersResponse, error) {
	e.logger.Debug("Received list performers request",
		zap.String("avsAddressFilter", req.AvsAddress),
	)

	var performers []*executorV1.Performer

	for avsAddress, performer := range e.avsPerformers {
		// Skip if filtering by AVS address
		if req.AvsAddress != "" && req.AvsAddress != avsAddress {
			continue
		}

		// Cast to serverPerformer to access health methods
		serverPerf, ok := performer.(*serverPerformer.AvsPerformerServer)
		if !ok {
			// Create a basic performer entry for non-server performers
			performers = append(performers, &executorV1.Performer{
				PerformerId:        fmt.Sprintf("%s-legacy", avsAddress),
				AvsAddress:         avsAddress,
				Status:             "Active",
				ResourceHealthy:    true,
				ApplicationHealthy: true,
				LastHealthCheck:    time.Now().Format(time.RFC3339),
			})
			continue
		}

		// Get health status of all containers for this performer
		healthStatuses, err := serverPerf.GetAllContainerHealth(ctx)
		if err != nil {
			e.logger.Error("Failed to get container health",
				zap.String("avsAddress", avsAddress),
				zap.Error(err),
			)
			continue
		}

		// Create performer entries for each container
		for _, status := range healthStatuses {
			performers = append(performers, &executorV1.Performer{
				PerformerId:        fmt.Sprintf("%s-%s", avsAddress, status.ContainerID[:8]),
				AvsAddress:         avsAddress,
				Status:             status.Status.String(),
				ArtifactRegistry:   extractRegistry(status.Image),
				ArtifactDigest:     extractDigest(status.Image),
				ResourceHealthy:    status.ContainerHealth == serverPerformer.ContainerHealthHealthy,
				ApplicationHealthy: status.ApplicationHealth == serverPerformer.AppHealthHealthy,
				LastHealthCheck:    status.LastHealthCheck.Format(time.RFC3339),
				ContainerId:        status.ContainerID,
			})
		}
	}

	e.logger.Debug("Returning performers list",
		zap.Int("count", len(performers)),
	)

	return &executorV1.ListPerformersResponse{
		Performers: performers,
	}, nil
}

// RemovePerformer removes a performer from the executor
func (e *Executor) RemovePerformer(ctx context.Context, req *executorV1.RemovePerformerRequest) (*executorV1.RemovePerformerResponse, error) {
	e.logger.Info("Received remove performer request",
		zap.String("performerId", req.PerformerId),
	)

	// Parse performer ID to get AVS address and container ID
	avsAddress, containerID := parsePerformerID(req.PerformerId)
	if avsAddress == "" {
		return &executorV1.RemovePerformerResponse{
			Success: false,
			Message: "invalid performer ID format",
		}, nil
	}

	performer := e.getPerformerByAVS(avsAddress)
	if performer == nil {
		return &executorV1.RemovePerformerResponse{
			Success: false,
			Message: "performer not found",
		}, nil
	}

	// Cast to serverPerformer to access removal methods
	serverPerf, ok := performer.(*serverPerformer.AvsPerformerServer)
	if !ok {
		return &executorV1.RemovePerformerResponse{
			Success: false,
			Message: "performer does not support container removal",
		}, nil
	}

	// Remove specific container (if specified) or entire performer
	if containerID != "" {
		err := serverPerf.RemoveContainer(containerID)
		if err != nil {
			e.logger.Error("Failed to remove container",
				zap.String("avsAddress", avsAddress),
				zap.String("containerID", containerID),
				zap.Error(err),
			)
			return &executorV1.RemovePerformerResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}
		return &executorV1.RemovePerformerResponse{
			Success: true,
			Message: "container removed successfully",
		}, nil
	} else {
		// Remove entire performer (not implemented yet)
		return &executorV1.RemovePerformerResponse{
			Success: false,
			Message: "removing entire performers not yet supported",
		}, nil
	}
}

// Helper functions

func (e *Executor) getPerformerByAVS(avsAddress string) interface{} {
	return e.avsPerformers[avsAddress]
}

func parsePerformerID(performerID string) (avsAddress, containerID string) {
	parts := strings.Split(performerID, "-")
	if len(parts) < 2 {
		return "", ""
	}

	// Format: avsAddress-containerID or avsAddress-legacy
	avsAddress = parts[0]
	if len(parts) > 1 && parts[1] != "legacy" {
		containerID = parts[1]
	}

	return avsAddress, containerID
}

func extractRegistry(image string) string {
	parts := strings.Split(image, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return image
}

func extractDigest(image string) string {
	parts := strings.Split(image, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}
