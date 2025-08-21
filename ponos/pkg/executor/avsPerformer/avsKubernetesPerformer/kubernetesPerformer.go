package avsKubernetesPerformer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/kubernetesManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	defaultDeploymentTimeout  = 5 * time.Minute // Kubernetes deployments can take longer
	defaultRunningWaitTimeout = 5 * time.Minute // K8s pods need more time to start
	defaultGRPCPort           = 8080
)

// PerformerLifecycleState represents the lifecycle state of a performer
type PerformerLifecycleState int

const (
	PerformerStateActive PerformerLifecycleState = iota
	PerformerStateDraining
	PerformerStateShutdown
)

// PerformerTaskState holds task tracking state for a performer
type PerformerTaskState struct {
	waitGroup *sync.WaitGroup
	state     PerformerLifecycleState
}

// PerformerResource holds information about a Kubernetes performer
type PerformerResource struct {
	performerID string
	avsAddress  string
	image       avsPerformer.PerformerImage
	status      avsPerformer.PerformerResourceStatus
	client      performerV1.PerformerServiceClient
	grpcConn    *grpc.ClientConn // Single gRPC connection
	endpoint    string
	statusChan  chan avsPerformer.PerformerStatusEvent
	createdAt   time.Time
}

// AvsKubernetesPerformer implements IAvsPerformer using Kubernetes CRDs
type AvsKubernetesPerformer struct {
	config           *avsPerformer.AvsPerformerConfig
	kubernetesConfig *kubernetesManager.Config
	logger           *zap.Logger
	peeringFetcher   peering.IPeeringDataFetcher
	l1ContractCaller contractCaller.IContractCaller
	aggregatorPeers  []*peering.OperatorPeerInfo

	// Kubernetes operations
	kubernetesManager *kubernetesManager.CRDOperations
	clientWrapper     *kubernetesManager.ClientWrapper

	// Performer tracking
	currentPerformer     atomic.Value // *PerformerResource
	nextPerformer        *PerformerResource
	performerResourcesMu sync.Mutex

	// Task tracking - single mutex protects both maps and lifecycle state
	performerTaskStates map[string]*PerformerTaskState
	performerStatesMu   sync.Mutex

	// Deployment tracking
	activeDeploymentMu sync.Mutex
}

// NewAvsKubernetesPerformer creates a new Kubernetes-based AVS performer
func NewAvsKubernetesPerformer(
	config *avsPerformer.AvsPerformerConfig,
	kubernetesConfig *kubernetesManager.Config,
	peeringFetcher peering.IPeeringDataFetcher,
	l1ContractCaller contractCaller.IContractCaller,
	logger *zap.Logger,
) (*AvsKubernetesPerformer, error) {

	// Initialize Kubernetes client
	clientWrapper, err := kubernetesManager.NewClientWrapper(kubernetesConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Test connection to Kubernetes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := clientWrapper.TestConnection(ctx); err != nil {
		logger.Warn("Failed to test Kubernetes connection, continuing anyway", zap.Error(err))
	}

	// Create CRD operations manager
	crdOps := kubernetesManager.NewCRDOperations(clientWrapper.CRDClient, kubernetesConfig, logger)

	return &AvsKubernetesPerformer{
		config:              config,
		kubernetesConfig:    kubernetesConfig,
		logger:              logger,
		peeringFetcher:      peeringFetcher,
		l1ContractCaller:    l1ContractCaller,
		kubernetesManager:   crdOps,
		clientWrapper:       clientWrapper,
		performerTaskStates: make(map[string]*PerformerTaskState),
	}, nil
}

// Initialize initializes the Kubernetes performer
func (akp *AvsKubernetesPerformer) Initialize(ctx context.Context) error {
	akp.performerResourcesMu.Lock()
	defer akp.performerResourcesMu.Unlock()

	// Fetch aggregator peer information
	aggregatorPeers, err := akp.fetchAggregatorPeerInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch aggregator peers: %w", err)
	}
	akp.aggregatorPeers = aggregatorPeers

	akp.logger.Info("Kubernetes performer initialized",
		zap.String("avsAddress", akp.config.AvsAddress),
		zap.String("namespace", akp.kubernetesConfig.Namespace),
		zap.Int("aggregatorPeers", len(akp.aggregatorPeers)),
	)

	// Skip initial container creation if image info is empty (for deployment-based initialization)
	if akp.config.Image.Repository == "" || akp.config.Image.Tag == "" {
		akp.logger.Info("Starting Kubernetes performer without initial container",
			zap.String("avsAddress", akp.config.AvsAddress),
		)
		return nil
	}

	// Create and start initial performer
	performerResource, err := akp.createPerformerResource(ctx, akp.config.Image)
	if err != nil {
		return fmt.Errorf("failed to create initial performer: %w", err)
	}

	performerResource.status = avsPerformer.PerformerResourceStatusInService
	akp.currentPerformer.Store(performerResource)

	return nil
}

// fetchAggregatorPeerInfo fetches peer information with retries
func (akp *AvsKubernetesPerformer) fetchAggregatorPeerInfo(ctx context.Context) ([]*peering.OperatorPeerInfo, error) {
	retries := []uint64{1, 3, 5, 10, 20}
	for i, retry := range retries {
		aggPeers, err := akp.peeringFetcher.ListAggregatorOperators(ctx, akp.config.AvsAddress)
		if err != nil {
			akp.logger.Error("Failed to fetch aggregator peers",
				zap.String("avsAddress", akp.config.AvsAddress),
				zap.Error(err),
			)
			if i == len(retries)-1 {
				return nil, err
			}
			time.Sleep(time.Duration(retry) * time.Second)
			continue
		}
		return aggPeers, nil
	}
	return nil, fmt.Errorf("failed to fetch aggregator peers after retries")
}

// generatePerformerID generates a unique performer ID
func (akp *AvsKubernetesPerformer) generatePerformerID() string {
	// Use shortened address hash (6 chars) + shortened UUID (8 chars) for uniqueness
	// This keeps the total length under Kubernetes 63-character limit for labels
	shortUUID := strings.Replace(uuid.New().String(), "-", "", -1)[:8]
	return fmt.Sprintf("performer-%s-%s", containerManager.HashAvsAddress(akp.config.AvsAddress), shortUUID)
}

// buildEnvironmentFromImage builds environment variables from the PerformerImage configuration
func (akp *AvsKubernetesPerformer) buildEnvironmentFromImage(image avsPerformer.PerformerImage) (map[string]string, []kubernetesManager.EnvVarSource) {
	envMap := make(map[string]string)
	var envVarSources []kubernetesManager.EnvVarSource

	// Add default environment variables
	envMap["AVS_ADDRESS"] = akp.config.AvsAddress
	envMap["GRPC_PORT"] = fmt.Sprintf("%d", defaultGRPCPort)

	// Process environment variables from image.Envs
	for _, env := range image.Envs {
		// Skip if no value source is specified
		if env.Value == "" && env.ValueFromEnv == "" && env.KubernetesEnv == nil {
			continue
		}

		// Handle KubernetesEnv (references to secrets/configmaps)
		if env.KubernetesEnv != nil && env.KubernetesEnv.ValueFrom.SecretKeyRef.Name != "" {
			envVarSource := kubernetesManager.EnvVarSource{
				Name: env.Name,
				ValueFrom: &kubernetesManager.EnvValueFrom{
					SecretKeyRef: &kubernetesManager.KeySelector{
						Name: env.KubernetesEnv.ValueFrom.SecretKeyRef.Name,
						Key:  env.KubernetesEnv.ValueFrom.SecretKeyRef.Key,
					},
				},
			}
			envVarSources = append(envVarSources, envVarSource)
		} else if env.KubernetesEnv != nil && env.KubernetesEnv.ValueFrom.ConfigMapKeyRef.Name != "" {
			envVarSource := kubernetesManager.EnvVarSource{
				Name: env.Name,
				ValueFrom: &kubernetesManager.EnvValueFrom{
					ConfigMapKeyRef: &kubernetesManager.KeySelector{
						Name: env.KubernetesEnv.ValueFrom.ConfigMapKeyRef.Name,
						Key:  env.KubernetesEnv.ValueFrom.ConfigMapKeyRef.Key,
					},
				},
			}
			envVarSources = append(envVarSources, envVarSource)
		} else if env.Value != "" {
			// Handle direct value
			envMap[env.Name] = env.Value
		}
		// Note: Ignoring ValueFromEnv for now as requested
	}

	return envMap, envVarSources
}

// createPerformerResource creates a new Kubernetes performer resource
func (akp *AvsKubernetesPerformer) createPerformerResource(
	ctx context.Context,
	image avsPerformer.PerformerImage,
) (*PerformerResource, error) {
	performerID := akp.generatePerformerID()

	// Build environment variables and sources
	envMap, envVarSources := akp.buildEnvironmentFromImage(image)

	// Create Kubernetes CRD request
	createRequest := &kubernetesManager.CreatePerformerRequest{
		Name:               performerID,
		AVSAddress:         akp.config.AvsAddress,
		Image:              fmt.Sprintf("%s:%s", image.Repository, image.Tag),
		ImagePullPolicy:    "Never", // Use local images only for testing
		ImageTag:           image.Tag,
		ImageDigest:        image.Digest,
		GRPCPort:           defaultGRPCPort,
		Environment:        envMap,
		EnvironmentFrom:    envVarSources,
		ServiceAccountName: image.ServiceAccountName,
		// Add resource requirements if needed
		// Resources: &kubernetesManager.ResourceRequirements{...},
	}

	akp.logger.Info("Creating Kubernetes performer resource",
		zap.String("performerID", performerID),
		zap.String("avsAddress", akp.config.AvsAddress),
		zap.String("image", createRequest.Image),
		zap.String("namespace", akp.kubernetesConfig.Namespace),
	)

	// Create the Performer CRD
	createResponse, err := akp.kubernetesManager.CreatePerformer(ctx, createRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create Performer CRD: %w", err)
	}

	// Wait for the performer to be ready
	if err := akp.waitForPerformerReady(ctx, performerID); err != nil {
		// Clean up on failure
		if cleanupErr := akp.kubernetesManager.DeletePerformer(ctx, performerID); cleanupErr != nil {
			akp.logger.Error("Failed to clean up performer after creation failure",
				zap.String("performerID", performerID),
				zap.Error(cleanupErr),
			)
		}
		return nil, fmt.Errorf("performer not ready: %w", err)
	}

	// Create a single gRPC connection with retry logic
	retryConfig := &clients.RetryConfig{
		MaxRetries:        5,
		InitialDelay:      2 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 15 * time.Second,
	}

	// Use endpoint override if provided (for testing when executor is outside cluster)
	endpoint := createResponse.Endpoint
	if akp.config.EndpointOverride != "" {
		akp.logger.Sugar().Info("Using endpoint override for performer",
			zap.String("performerID", performerID),
			zap.String("originalEndpoint", createResponse.Endpoint),
			zap.String("overrideEndpoint", akp.config.EndpointOverride),
		)
		endpoint = akp.config.EndpointOverride
	}

	akp.logger.Info("Creating gRPC client for performer",
		zap.String("performerID", performerID),
		zap.String("endpoint", endpoint),
	)
	grpcConn, err := clients.NewGrpcClientWithRetry(endpoint, true, retryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// Create the client once for this performer
	client := performerV1.NewPerformerServiceClient(grpcConn)

	akp.logger.Info("Created gRPC client for performer",
		zap.String("performerID", performerID),
		zap.String("endpoint", endpoint),
	)

	performerResource := &PerformerResource{
		performerID: performerID,
		avsAddress:  akp.config.AvsAddress,
		image:       image,
		status:      avsPerformer.PerformerResourceStatusStaged,
		client:      client,
		grpcConn:    grpcConn,
		endpoint:    endpoint, // Use the potentially overridden endpoint
		statusChan:  make(chan avsPerformer.PerformerStatusEvent, 10),
		createdAt:   time.Now(),
	}

	akp.logger.Info("Kubernetes performer resource created successfully",
		zap.String("performerID", performerID),
		zap.String("endpoint", endpoint),
	)

	return performerResource, nil
}

// waitForPerformerReady waits for the performer to be ready
func (akp *AvsKubernetesPerformer) waitForPerformerReady(ctx context.Context, performerID string) error {
	akp.logger.Sugar().Infow("Waiting for performer to be ready",
		zap.String("performerID", performerID),
	)
	timeout := defaultRunningWaitTimeout
	pollInterval := 5 * time.Second

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for performer %s to be ready", performerID)
		case <-ticker.C:
			status, err := akp.kubernetesManager.GetPerformerStatus(ctx, performerID)
			if err != nil {
				akp.logger.Warn("Failed to get performer status",
					zap.String("performerID", performerID),
					zap.Error(err),
				)
				continue
			}

			akp.logger.Sugar().Infow("Performer status check",
				zap.String("performerID", performerID),
				zap.String("phase", string(status.Phase)),
				zap.Bool("ready", status.Ready),
			)

			if status.Ready {
				akp.logger.Sugar().Infow("Performer is ready",
					zap.String("performerID", performerID),
					zap.String("phase", string(status.Phase)),
				)
				return nil
			}
		}
	}
}

// CreatePerformer creates a new performer and returns the creation result
func (akp *AvsKubernetesPerformer) CreatePerformer(
	ctx context.Context,
	image avsPerformer.PerformerImage,
) (*avsPerformer.PerformerCreationResult, error) {
	akp.performerResourcesMu.Lock()
	defer akp.performerResourcesMu.Unlock()

	akp.logger.Info("Creating new Kubernetes performer",
		zap.String("avsAddress", akp.config.AvsAddress),
		zap.String("imageRepository", image.Repository),
		zap.String("imageTag", image.Tag),
	)

	// Check if next performer slot is already occupied
	if akp.nextPerformer != nil {
		return nil, fmt.Errorf("a next performer already exists (ID: %s). Please remove it explicitly before creating a new one", akp.nextPerformer.performerID)
	}

	// Create the new performer resource
	newPerformer, err := akp.createPerformerResource(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("failed to create performer resource: %w", err)
	}

	// Always deploy as next performer
	akp.nextPerformer = newPerformer
	akp.nextPerformer.status = avsPerformer.PerformerResourceStatusStaged

	akp.logger.Info("Kubernetes performer created successfully",
		zap.String("performerID", newPerformer.performerID),
		zap.String("endpoint", newPerformer.endpoint),
	)

	return &avsPerformer.PerformerCreationResult{
		PerformerID: newPerformer.performerID,
		StatusChan:  newPerformer.statusChan,
	}, nil
}

// PromotePerformer promotes a performer to current
func (akp *AvsKubernetesPerformer) PromotePerformer(ctx context.Context, performerID string) error {
	akp.performerResourcesMu.Lock()
	defer akp.performerResourcesMu.Unlock()

	// Check if the performer is already current
	if current := akp.currentPerformer.Load(); current != nil {
		currentPerformer := current.(*PerformerResource)
		if currentPerformer.performerID == performerID {
			akp.logger.Info("Performer is already current",
				zap.String("performerID", performerID),
			)
			return nil
		}
	}

	// Check if the performer is the next performer
	if akp.nextPerformer == nil || akp.nextPerformer.performerID != performerID {
		return fmt.Errorf("performer %s is not in the next deployment slot", performerID)
	}

	akp.logger.Info("Promoting performer to current",
		zap.String("performerID", performerID),
	)

	// Initiate draining of old performer
	if current := akp.currentPerformer.Load(); current != nil {
		oldPerformer := current.(*PerformerResource)
		if oldPerformer != nil {
			akp.logger.Info("Initiating drain of old performer",
				zap.String("oldPerformerID", oldPerformer.performerID),
			)
			akp.startDrainAndRemove(oldPerformer)
		}
	}

	// Promote next to current
	akp.nextPerformer.status = avsPerformer.PerformerResourceStatusInService
	akp.currentPerformer.Store(akp.nextPerformer)
	akp.nextPerformer = nil

	akp.logger.Info("Performer promotion completed",
		zap.String("performerID", performerID),
	)

	return nil
}

// RemovePerformer removes a performer by ID
func (akp *AvsKubernetesPerformer) RemovePerformer(ctx context.Context, performerID string) error {
	akp.performerResourcesMu.Lock()
	defer akp.performerResourcesMu.Unlock()

	var targetPerformer *PerformerResource

	// Find the performer to remove
	if current := akp.currentPerformer.Load(); current != nil {
		currentPerformer, ok := current.(*PerformerResource)
		if !ok {
			akp.logger.Error("Invalid type in currentPerformer atomic.Value during removal")
			return fmt.Errorf("invalid performer type stored in currentPerformer")
		}
		if currentPerformer.performerID == performerID {
			targetPerformer = currentPerformer
			akp.currentPerformer.Store((*PerformerResource)(nil))
		}
	}

	if targetPerformer == nil && akp.nextPerformer != nil && akp.nextPerformer.performerID == performerID {
		targetPerformer = akp.nextPerformer
		akp.nextPerformer = nil
	}

	if targetPerformer == nil {
		return fmt.Errorf("performer with ID %s not found", performerID)
	}

	// Close status channel
	if targetPerformer.statusChan != nil {
		close(targetPerformer.statusChan)
	}

	// Close gRPC connection
	if targetPerformer.grpcConn != nil {
		if err := targetPerformer.grpcConn.Close(); err != nil {
			akp.logger.Warn("Failed to close gRPC connection",
				zap.String("performerID", performerID),
				zap.Error(err),
			)
		}
	}

	// Delete the Kubernetes performer resource
	if err := akp.kubernetesManager.DeletePerformer(ctx, performerID); err != nil {
		akp.logger.Error("Failed to delete Kubernetes performer",
			zap.String("performerID", performerID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete performer: %w", err)
	}

	akp.logger.Info("Performer removed successfully",
		zap.String("performerID", performerID),
	)

	return nil
}

// RunTask executes a task on the current performer
func (akp *AvsKubernetesPerformer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	akp.logger.Debug("Processing task", zap.String("taskID", task.TaskID))

	// Load current performer
	current := akp.currentPerformer.Load()
	if current == nil {
		return nil, fmt.Errorf("no current performer available to execute task")
	}

	currentPerformer, ok := current.(*PerformerResource)
	if !ok || currentPerformer == nil || currentPerformer.client == nil {
		return nil, fmt.Errorf("no current performer client available to execute task")
	}

	// Track this task with the performer's WaitGroup
	if !akp.tryAddTask(currentPerformer.performerID) {
		return nil, fmt.Errorf("performer %s is no longer accepting tasks (draining or shutdown)", currentPerformer.performerID)
	}
	defer akp.taskCompleted(currentPerformer.performerID)

	// Execute the task using the pre-created client
	res, err := currentPerformer.client.ExecuteTask(ctx, &performerV1.TaskRequest{
		TaskId:  []byte(task.TaskID),
		Payload: task.Payload,
	})
	if err != nil {
		akp.logger.Error("Performer failed to handle task",
			zap.String("performerID", currentPerformer.performerID),
			zap.String("taskID", task.TaskID),
			zap.Error(err),
		)
		return nil, err
	}

	akp.logger.Debug("Performer handled task successfully",
		zap.String("performerID", currentPerformer.performerID),
		zap.String("taskID", task.TaskID),
	)

	return performerTask.NewTaskResultFromResultProto(res), nil
}

// ValidateTaskSignature validates task signatures (same as container implementation)
func (akp *AvsKubernetesPerformer) ValidateTaskSignature(t *performerTask.PerformerTask) error {
	peer := util.Find(akp.aggregatorPeers, func(p *peering.OperatorPeerInfo) bool {
		return strings.EqualFold(p.OperatorAddress, t.AggregatorAddress)
	})
	if peer == nil {
		return fmt.Errorf("failed to find peer for task")
	}

	isVerified := false

	for _, opset := range peer.OperatorSets {
		var scheme signing.SigningScheme
		switch opset.CurveType {
		case config.CurveTypeBN254:
			scheme = bn254.NewScheme()
		case config.CurveTypeECDSA:
			scheme = ecdsa.NewScheme()
		default:
			return fmt.Errorf("unsupported curve type for signature verification: %s", opset.CurveType)
		}

		sig, err := scheme.NewSignatureFromBytes(t.Signature)
		if err != nil {
			return err
		}

		var verified bool
		payloadHash := crypto.Keccak256Hash(t.Payload)
		switch opset.CurveType {
		case config.CurveTypeBN254:
			verified, err = sig.Verify(opset.WrappedPublicKey.PublicKey, payloadHash[:])
		case config.CurveTypeECDSA:
			typedSig, err := ecdsa.NewSignatureFromBytes(sig.Bytes())
			if err != nil {
				continue
			}
			verified, err = typedSig.VerifyWithAddress(payloadHash[:], opset.WrappedPublicKey.ECDSAAddress)
			if err != nil {
				continue
			}
		}

		if err != nil {
			continue
		}

		if verified {
			isVerified = true
			break
		}
	}

	if !isVerified {
		return fmt.Errorf("failed to verify signature with any operator set")
	}

	return nil
}

// Deploy performs a synchronous deployment
func (akp *AvsKubernetesPerformer) Deploy(ctx context.Context, image avsPerformer.PerformerImage) (*avsPerformer.DeploymentResult, error) {
	// Use deployment mutex to prevent concurrent deployments
	if !akp.activeDeploymentMu.TryLock() {
		return nil, fmt.Errorf("deployment in progress for avs %s", akp.config.AvsAddress)
	}
	defer akp.activeDeploymentMu.Unlock()

	deploymentID := fmt.Sprintf("k8s-deployment-%s-%s", akp.config.AvsAddress, uuid.New().String())
	timeout := defaultDeploymentTimeout

	deploymentCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := &avsPerformer.DeploymentResult{
		ID:        deploymentID,
		Status:    avsPerformer.DeploymentStatusPending,
		Image:     image,
		StartTime: time.Now(),
	}

	akp.logger.Info("Starting Kubernetes deployment",
		zap.String("deploymentID", deploymentID),
		zap.String("image", fmt.Sprintf("%s:%s", image.Repository, image.Tag)),
	)

	// Create the performer
	creationResult, err := akp.CreatePerformer(deploymentCtx, image)
	if err != nil {
		result.Status = avsPerformer.DeploymentStatusFailed
		result.EndTime = time.Now()
		result.Error = err
		result.Message = fmt.Sprintf("Failed to create performer: %v", err)
		return result, err
	}

	result.PerformerID = creationResult.PerformerID
	result.Status = avsPerformer.DeploymentStatusInProgress

	// Monitor deployment until pod is ready (Kubernetes handles health checks)
	if err := akp.waitForPerformerReady(deploymentCtx, creationResult.PerformerID); err != nil {
		// Deployment failed, clean up
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()

		if removeErr := akp.RemovePerformer(cleanupCtx, creationResult.PerformerID); removeErr != nil {
			akp.logger.Error("Failed to clean up failed deployment",
				zap.String("deploymentID", deploymentID),
				zap.Error(removeErr),
			)
		}

		result.Status = avsPerformer.DeploymentStatusFailed
		result.EndTime = time.Now()
		result.Error = err
		result.Message = fmt.Sprintf("Deployment failed: %v", err)
		return result, err
	}

	// Promote the performer
	if err := akp.PromotePerformer(deploymentCtx, creationResult.PerformerID); err != nil {
		// Clean up failed deployment
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()

		if removeErr := akp.RemovePerformer(cleanupCtx, creationResult.PerformerID); removeErr != nil {
			akp.logger.Error("Failed to clean up after promotion failure",
				zap.String("deploymentID", deploymentID),
				zap.Error(removeErr),
			)
		}

		result.Status = avsPerformer.DeploymentStatusFailed
		result.EndTime = time.Now()
		result.Error = err
		result.Message = fmt.Sprintf("Failed to promote performer: %v", err)
		return result, err
	}

	// Deployment successful
	result.Status = avsPerformer.DeploymentStatusCompleted
	result.EndTime = time.Now()
	result.Message = "Kubernetes deployment completed successfully"

	akp.logger.Info("Kubernetes deployment completed successfully",
		zap.String("deploymentID", deploymentID),
		zap.String("performerID", creationResult.PerformerID),
		zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
	)

	return result, nil
}

// ListPerformers returns information about current and next performers
func (akp *AvsKubernetesPerformer) ListPerformers() []avsPerformer.PerformerMetadata {
	akp.performerResourcesMu.Lock()
	defer akp.performerResourcesMu.Unlock()

	var performers []avsPerformer.PerformerMetadata

	// Add current performer
	if current := akp.currentPerformer.Load(); current != nil {
		currentPerformer, ok := current.(*PerformerResource)
		if !ok {
			akp.logger.Error("Invalid type in currentPerformer atomic.Value during listing")
			return performers
		}
		performers = append(performers, akp.convertPerformerResource(currentPerformer))
	}

	// Add next performer
	if akp.nextPerformer != nil {
		performers = append(performers, akp.convertPerformerResource(akp.nextPerformer))
	}

	return performers
}

// convertPerformerResource converts a PerformerResource to PerformerMetadata
func (akp *AvsKubernetesPerformer) convertPerformerResource(performer *PerformerResource) avsPerformer.PerformerMetadata {
	return avsPerformer.PerformerMetadata{
		PerformerID:        performer.performerID,
		AvsAddress:         performer.avsAddress,
		Status:             performer.status,
		ArtifactRegistry:   performer.image.Repository,
		ArtifactTag:        performer.image.Tag,
		ArtifactDigest:     performer.image.Digest,
		ContainerHealthy:   true, // Kubernetes handles health checks
		ApplicationHealthy: true, // Kubernetes handles health checks
		LastHealthCheck:    time.Now(),
		ResourceID:         performer.performerID, // In K8s, we use performer ID as resource ID
	}
}

// tryAddTask attempts to add a task for the performer if it's in active state
// LOCKING: this will acquire lock on the state to modify.
func (akp *AvsKubernetesPerformer) tryAddTask(performerID string) bool {
	akp.performerStatesMu.Lock()
	defer akp.performerStatesMu.Unlock()

	state, exists := akp.performerTaskStates[performerID]
	if !exists {
		// Create new active state for this performer
		state = &PerformerTaskState{
			waitGroup: &sync.WaitGroup{},
			state:     PerformerStateActive,
		}
		akp.performerTaskStates[performerID] = state
	} else if state.state == PerformerStateShutdown {
		// Don't allow tasks on shutdown performers
		return false
	}

	// Only allow task addition if performer is active
	if state.state != PerformerStateActive {
		return false
	}

	state.waitGroup.Add(1)
	return true
}

// taskCompleted marks a task as completed
// LOCKING: this will acquire lock on the state to modify.
func (akp *AvsKubernetesPerformer) taskCompleted(performerID string) {
	akp.performerStatesMu.Lock()
	defer akp.performerStatesMu.Unlock()

	if state, exists := akp.performerTaskStates[performerID]; exists {
		state.waitGroup.Done()
	}
}

// waitForTaskCompletion waits for all tasks on a performer to complete
// LOCKING: this will acquire lock on the state to modify.
func (akp *AvsKubernetesPerformer) waitForTaskCompletion(performerID string) {
	akp.performerStatesMu.Lock()
	state, exists := akp.performerTaskStates[performerID]
	if exists {
		// Mark as draining to prevent new tasks (unless already shutdown)
		if state.state != PerformerStateShutdown {
			state.state = PerformerStateDraining
		}
		wg := state.waitGroup
		akp.performerStatesMu.Unlock()

		// Wait outside the critical section
		wg.Wait()
	} else {
		akp.performerStatesMu.Unlock()
	}
}

// cleanupTaskWaitGroup marks the performer state as shutdown but keeps the state
// to prevent race conditions with waitForTaskCompletion
// LOCKING: this will acquire lock on the state to modify.
func (akp *AvsKubernetesPerformer) cleanupTaskWaitGroup(performerID string) {
	akp.performerStatesMu.Lock()
	defer akp.performerStatesMu.Unlock()

	if state, exists := akp.performerTaskStates[performerID]; exists {
		state.state = PerformerStateShutdown
		// Do not delete the state to prevent waitGroup race conditions
		// The state will be cleaned up when the performer map itself is cleaned up
	}
}

// startDrainAndRemove initiates draining of a performer in a separate goroutine
func (akp *AvsKubernetesPerformer) startDrainAndRemove(performer *PerformerResource) {
	performerID := performer.performerID

	// Check if already draining
	akp.performerStatesMu.Lock()
	state, exists := akp.performerTaskStates[performerID]
	if !exists {
		// Create draining state if performer doesn't exist (but not if it was shutdown)
		state = &PerformerTaskState{
			waitGroup: &sync.WaitGroup{},
			state:     PerformerStateDraining,
		}
		akp.performerTaskStates[performerID] = state
	} else if state.state == PerformerStateDraining || state.state == PerformerStateShutdown {
		// Already draining or shutdown - don't restart drain process
		akp.performerStatesMu.Unlock()
		return
	} else {
		// Mark as draining
		state.state = PerformerStateDraining
	}
	akp.performerStatesMu.Unlock()

	go func() {
		akp.logger.Info("Starting performer drain",
			zap.String("performerID", performerID),
		)

		// Wait for all tasks to complete
		akp.waitForTaskCompletion(performerID)

		akp.logger.Info("Performer drained, removing",
			zap.String("performerID", performerID),
		)

		// Remove the performer
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Close gRPC connection
		if performer.grpcConn != nil {
			if err := performer.grpcConn.Close(); err != nil {
				akp.logger.Warn("Failed to close gRPC connection during drain",
					zap.String("performerID", performerID),
					zap.Error(err),
				)
			}
		}

		if err := akp.kubernetesManager.DeletePerformer(ctx, performerID); err != nil {
			akp.logger.Error("Failed to remove drained performer",
				zap.String("performerID", performerID),
				zap.Error(err),
			)
		}

		akp.cleanupTaskWaitGroup(performerID)

		akp.logger.Info("Performer drain completed",
			zap.String("performerID", performerID),
		)
	}()
}

// Shutdown shuts down the Kubernetes performer
func (akp *AvsKubernetesPerformer) Shutdown() error {
	akp.performerResourcesMu.Lock()
	defer akp.performerResourcesMu.Unlock()

	akp.logger.Info("Shutting down Kubernetes performer",
		zap.String("avsAddress", akp.config.AvsAddress),
	)

	// Mark all performers as shutdown
	akp.performerStatesMu.Lock()
	for performerID, state := range akp.performerTaskStates {
		state.state = PerformerStateShutdown
		delete(akp.performerTaskStates, performerID)
	}
	akp.performerStatesMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var errs []error

	// Shutdown current performer
	if current := akp.currentPerformer.Load(); current != nil {
		currentPerformer, ok := current.(*PerformerResource)
		if !ok {
			akp.logger.Error("Invalid type in currentPerformer atomic.Value during shutdown")
			errs = append(errs, fmt.Errorf("invalid performer type stored in currentPerformer"))
		}
		// Close gRPC connection
		if currentPerformer.grpcConn != nil {
			if err := currentPerformer.grpcConn.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close current performer gRPC connection: %w", err))
			}
		}

		akp.waitForTaskCompletion(currentPerformer.performerID)
		if err := akp.kubernetesManager.DeletePerformer(ctx, currentPerformer.performerID); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown current performer: %w", err))
		}
		akp.cleanupTaskWaitGroup(currentPerformer.performerID)
		akp.currentPerformer.Store((*PerformerResource)(nil))
	}

	// Shutdown next performer
	if akp.nextPerformer != nil {
		// Close gRPC connection
		if akp.nextPerformer.grpcConn != nil {
			if err := akp.nextPerformer.grpcConn.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close next performer gRPC connection: %w", err))
			}
		}

		akp.waitForTaskCompletion(akp.nextPerformer.performerID)
		if err := akp.kubernetesManager.DeletePerformer(ctx, akp.nextPerformer.performerID); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown next performer: %w", err))
		}
		akp.cleanupTaskWaitGroup(akp.nextPerformer.performerID)
		akp.nextPerformer = nil
	}

	// Close client wrapper
	if akp.clientWrapper != nil {
		if err := akp.clientWrapper.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Kubernetes client: %w", err))
		}
	}

	if len(errs) > 0 {
		return errors.New(fmt.Sprintf("shutdown errors: %v", errs))
	}

	akp.logger.Info("Kubernetes performer shutdown completed")
	return nil
}
