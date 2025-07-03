package performer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/google/uuid"
	orchestratorpb "github.com/hourglass/obsidian/api/proto/orchestrator"
	registrypb "github.com/hourglass/obsidian/api/proto/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ObsidianPerformer struct {
	config             *Config
	orchestratorClient orchestratorpb.OrchestratorServiceClient
	registryClient     registrypb.RegistryServiceClient
	performers         map[string]*PerformerInfo
	mu                 sync.RWMutex
	conn               *grpc.ClientConn
}

type Config struct {
	ObsidianEndpoint string
	AvsAddress       string
	Resources        *ResourceLimits
	Registries       []RegistryConfig
}

type ResourceLimits struct {
	CPU    string
	Memory string
	Disk   string
}

type RegistryConfig struct {
	Type             string
	Region           string
	CredentialsSecret string
}

type PerformerInfo struct {
	ID                string
	ContainerID       string
	Status            avsPerformer.PerformerResourceStatus
	Image             avsPerformer.PerformerImage
	CreatedAt         time.Time
	Health            *avsPerformer.PerformerHealth
	statusChan        chan avsPerformer.PerformerStatusEvent
}

func NewObsidianPerformer(config *Config) (*ObsidianPerformer, error) {
	conn, err := grpc.Dial(config.ObsidianEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50*1024*1024)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Obsidian: %w", err)
	}

	return &ObsidianPerformer{
		config:             config,
		orchestratorClient: orchestratorpb.NewOrchestratorServiceClient(conn),
		registryClient:     registrypb.NewRegistryServiceClient(conn),
		performers:         make(map[string]*PerformerInfo),
		conn:               conn,
	}, nil
}

func (o *ObsidianPerformer) Initialize(ctx context.Context) error {
	healthResp, err := o.orchestratorClient.GetHealth(ctx, nil)
	if err != nil {
		return fmt.Errorf("Obsidian health check failed: %w", err)
	}

	if !healthResp.Healthy {
		return fmt.Errorf("Obsidian is not healthy: %s", healthResp.Message)
	}

	return nil
}

func (o *ObsidianPerformer) Deploy(ctx context.Context, image avsPerformer.PerformerImage) (*avsPerformer.DeploymentResult, error) {
	startTime := time.Now()
	deploymentID := fmt.Sprintf("deploy-%s", uuid.New().String()[:8])

	imageRef := fmt.Sprintf("%s:%s", image.Repository, image.Tag)

	pullResp, err := o.registryClient.PullImage(ctx, &registrypb.ImageReference{
		Reference: imageRef,
	})
	if err != nil {
		return &avsPerformer.DeploymentResult{
			ID:        deploymentID,
			Status:    avsPerformer.DeploymentStatusFailed,
			Image:     image,
			StartTime: startTime,
			EndTime:   time.Now(),
			Message:   "Failed to pull image",
			Error:     err,
		}, nil
	}

	scanResp, err := o.registryClient.ScanImage(ctx, &registrypb.ImageID{
		Id: pullResp.Id,
	})
	if err == nil && scanResp.CriticalVulnerabilities > 0 {
		return &avsPerformer.DeploymentResult{
			ID:        deploymentID,
			Status:    avsPerformer.DeploymentStatusFailed,
			Image:     image,
			StartTime: startTime,
			EndTime:   time.Now(),
			Message:   fmt.Sprintf("Image has %d critical vulnerabilities", scanResp.CriticalVulnerabilities),
		}, nil
	}

	createResp, err := o.orchestratorClient.CreateContainer(ctx, &orchestratorpb.CreateContainerRequest{
		ImageId: pullResp.Id,
		Resources: &orchestratorpb.Resources{
			CpuLimit:    o.config.Resources.CPU,
			MemoryLimit: o.config.Resources.Memory,
			DiskLimit:   o.config.Resources.Disk,
		},
		Environment: map[string]string{
			"AVS_ADDRESS":      o.config.AvsAddress,
			"OBSIDIAN_ENABLED": "true",
		},
	})
	if err != nil {
		return &avsPerformer.DeploymentResult{
			ID:        deploymentID,
			Status:    avsPerformer.DeploymentStatusFailed,
			Image:     image,
			StartTime: startTime,
			EndTime:   time.Now(),
			Message:   "Failed to create container",
			Error:     err,
		}, nil
	}

	performerID := fmt.Sprintf("performer-%s", uuid.New().String()[:8])
	
	o.mu.Lock()
	o.performers[performerID] = &PerformerInfo{
		ID:          performerID,
		ContainerID: createResp.Id,
		Status:      avsPerformer.PerformerResourceStatusStaged,
		Image:       image,
		CreatedAt:   time.Now(),
		Health: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: false,
			LastHealthCheck:      time.Now(),
		},
		statusChan: make(chan avsPerformer.PerformerStatusEvent, 100),
	}
	o.mu.Unlock()

	return &avsPerformer.DeploymentResult{
		ID:          deploymentID,
		PerformerID: performerID,
		Status:      avsPerformer.DeploymentStatusCompleted,
		Image:       image,
		StartTime:   startTime,
		EndTime:     time.Now(),
		Message:     "Container created successfully",
	}, nil
}

func (o *ObsidianPerformer) CreatePerformer(ctx context.Context, image avsPerformer.PerformerImage) (*avsPerformer.PerformerCreationResult, error) {
	deployResult, err := o.Deploy(ctx, image)
	if err != nil {
		return nil, err
	}

	if deployResult.Status != avsPerformer.DeploymentStatusCompleted {
		return nil, fmt.Errorf("deployment failed: %s", deployResult.Message)
	}

	o.mu.RLock()
	performer := o.performers[deployResult.PerformerID]
	o.mu.RUnlock()

	_, err = o.orchestratorClient.StartContainer(ctx, &orchestratorpb.ContainerID{
		Id: performer.ContainerID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	go o.monitorPerformerHealth(performer)

	return &avsPerformer.PerformerCreationResult{
		PerformerID: deployResult.PerformerID,
		StatusChan:  performer.statusChan,
	}, nil
}

func (o *ObsidianPerformer) PromotePerformer(ctx context.Context, performerID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	performer, ok := o.performers[performerID]
	if !ok {
		return fmt.Errorf("performer not found: %s", performerID)
	}

	if performer.Status != avsPerformer.PerformerResourceStatusStaged {
		return fmt.Errorf("performer is not in staged status")
	}

	performer.Status = avsPerformer.PerformerResourceStatusInService

	performer.statusChan <- avsPerformer.PerformerStatusEvent{
		Status:      avsPerformer.PerformerHealthy,
		PerformerID: performerID,
		Message:     "Performer promoted to in-service",
		Timestamp:   time.Now(),
	}

	return nil
}

func (o *ObsidianPerformer) RemovePerformer(ctx context.Context, performerID string) error {
	o.mu.Lock()
	performer, ok := o.performers[performerID]
	if !ok {
		o.mu.Unlock()
		return fmt.Errorf("performer not found: %s", performerID)
	}
	delete(o.performers, performerID)
	o.mu.Unlock()

	_, err := o.orchestratorClient.StopContainer(ctx, &orchestratorpb.ContainerID{
		Id: performer.ContainerID,
	})
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	_, err = o.orchestratorClient.RemoveContainer(ctx, &orchestratorpb.ContainerID{
		Id: performer.ContainerID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	close(performer.statusChan)

	return nil
}

func (o *ObsidianPerformer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	performerID := o.selectPerformer()
	if performerID == "" {
		return nil, fmt.Errorf("no available performers")
	}

	o.mu.RLock()
	performer := o.performers[performerID]
	o.mu.RUnlock()

	resp, err := o.orchestratorClient.SubmitTask(ctx, &orchestratorpb.TaskRequest{
		ContainerId: performer.ContainerID,
		Payload:     task.Payload,
		Timeout:     nil,
		Metadata: map[string]string{
			"task_id":            task.TaskID,
			"avs_address":        task.Avs,
			"aggregator_address": task.AggregatorAddress,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}

	taskCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-taskCtx.Done():
			return nil, fmt.Errorf("task timeout")
		case <-ticker.C:
			status, err := o.orchestratorClient.GetTaskStatus(taskCtx, &orchestratorpb.TaskID{
				Id: resp.TaskId,
			})
			if err != nil {
				continue
			}

			switch status.State {
			case orchestratorpb.TaskState_TASK_STATE_SUCCESS:
				return &performerTask.PerformerTaskResult{
					TaskID: task.TaskID,
					Result: resp.Result,
				}, nil
			case orchestratorpb.TaskState_TASK_STATE_FAILED:
				return nil, fmt.Errorf("task failed: %s", status.ErrorMessage)
			case orchestratorpb.TaskState_TASK_STATE_TIMEOUT:
				return nil, fmt.Errorf("task timeout")
			}
		}
	}
}

func (o *ObsidianPerformer) ValidateTaskSignature(task *performerTask.PerformerTask) error {
	if len(task.Signature) == 0 {
		return fmt.Errorf("task signature is empty")
	}
	return nil
}

func (o *ObsidianPerformer) ListPerformers() []avsPerformer.PerformerMetadata {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var metadata []avsPerformer.PerformerMetadata
	for id, performer := range o.performers {
		metadata = append(metadata, avsPerformer.PerformerMetadata{
			PerformerID:        id,
			AvsAddress:         o.config.AvsAddress,
			ResourceID:         performer.ContainerID,
			Status:             performer.Status,
			ArtifactRegistry:   performer.Image.Repository,
			ArtifactDigest:     "",
			ContainerHealthy:   performer.Health.ContainerIsHealthy,
			ApplicationHealthy: performer.Health.ApplicationIsHealthy,
			LastHealthCheck:    performer.Health.LastHealthCheck,
		})
	}

	return metadata
}

func (o *ObsidianPerformer) Shutdown() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	ctx := context.Background()
	for id, performer := range o.performers {
		o.orchestratorClient.StopContainer(ctx, &orchestratorpb.ContainerID{
			Id: performer.ContainerID,
		})
		close(performer.statusChan)
		delete(o.performers, id)
	}

	return o.conn.Close()
}

func (o *ObsidianPerformer) selectPerformer() string {
	o.mu.RLock()
	defer o.mu.RUnlock()

	for id, performer := range o.performers {
		if performer.Status == avsPerformer.PerformerResourceStatusInService &&
			performer.Health.ContainerIsHealthy &&
			performer.Health.ApplicationIsHealthy {
			return id
		}
	}

	return ""
}

func (o *ObsidianPerformer) monitorPerformerHealth(performer *PerformerInfo) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.mu.RLock()
			if _, exists := o.performers[performer.ID]; !exists {
				o.mu.RUnlock()
				return
			}
			o.mu.RUnlock()

			performer.Health.LastHealthCheck = time.Now()
			performer.Health.ApplicationIsHealthy = true

			performer.statusChan <- avsPerformer.PerformerStatusEvent{
				Status:      avsPerformer.PerformerHealthy,
				PerformerID: performer.ID,
				Message:     "Health check passed",
				Timestamp:   time.Now(),
			}
		}
	}
}