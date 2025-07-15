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
)

const (
	defaultApplicationHealthCheckInterval = 15 * time.Second
	defaultDeploymentTimeout              = 5 * time.Minute  // Kubernetes deployments can take longer
	defaultRunningWaitTimeout             = 2 * time.Minute  // K8s pods need more time to start
	maxConsecutiveApplicationHealthFailures = 3
	defaultGRPCPort                       = 9090
)

// PerformerResource holds information about a Kubernetes performer
type PerformerResource struct {
	performerID       string
	avsAddress        string
	image             avsPerformer.PerformerImage
	status            avsPerformer.PerformerResourceStatus
	client            performerV1.PerformerServiceClient
	connectionManager *clients.ConnectionManager
	endpoint          string
	performerHealth   *avsPerformer.PerformerHealth
	statusChan        chan avsPerformer.PerformerStatusEvent
	createdAt         time.Time
	lastHealthCheck   time.Time
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
	currentPerformer      atomic.Value // *PerformerResource
	nextPerformer         *PerformerResource
	performerResourcesMu  sync.Mutex

	// Task tracking
	taskWaitGroups   map[string]*sync.WaitGroup
	taskWaitGroupsMu sync.Mutex

	// Draining tracking
	drainingPerformers   map[string]struct{}
	drainingPerformersMu sync.Mutex

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
	// Set default health check interval if not specified
	if config.ApplicationHealthCheckInterval == 0 {
		config.ApplicationHealthCheckInterval = defaultApplicationHealthCheckInterval
	}

	// Initialize Kubernetes client
	clientWrapper, err := kubernetesManager.NewClientWrapper(kubernetesConfig)
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
	crdOps := kubernetesManager.NewCRDOperations(clientWrapper.CRDClient, kubernetesConfig)

	return &AvsKubernetesPerformer{
		config:            config,
		kubernetesConfig:  kubernetesConfig,
		logger:            logger,
		peeringFetcher:    peeringFetcher,
		l1ContractCaller:  l1ContractCaller,
		kubernetesManager: crdOps,
		clientWrapper:     clientWrapper,
		taskWaitGroups:    make(map[string]*sync.WaitGroup),
		drainingPerformers: make(map[string]struct{}),
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

	// Start monitoring for the initial performer
	go akp.monitorPerformerHealth(ctx, performerResource)

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
	return fmt.Sprintf("performer-%s-%s", akp.config.AvsAddress, uuid.New().String())
}

// createPerformerResource creates a new Kubernetes performer resource
func (akp *AvsKubernetesPerformer) createPerformerResource(
	ctx context.Context,
	image avsPerformer.PerformerImage,
) (*PerformerResource, error) {
	performerID := akp.generatePerformerID()
	
	// Create Kubernetes CRD request
	createRequest := &kubernetesManager.CreatePerformerRequest{
		Name:       performerID,
		AVSAddress: akp.config.AvsAddress,
		Image:      fmt.Sprintf("%s:%s", image.Repository, image.Tag),
		ImageTag:   image.Tag,
		ImageDigest: image.Digest,
		GRPCPort:   defaultGRPCPort,
		Environment: map[string]string{
			"AVS_ADDRESS": akp.config.AvsAddress,
			"GRPC_PORT":   fmt.Sprintf("%d", defaultGRPCPort),
		},
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

	// Create connection manager with retry logic
	retryConfig := &clients.RetryConfig{
		MaxRetries:        5,
		InitialDelay:      2 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		ConnectionTimeout: 15 * time.Second,
	}
	
	connectionManager := clients.NewConnectionManager(createResponse.Endpoint, true, retryConfig)
	
	// Get initial connection to test connectivity
	conn, err := connectionManager.GetConnection()
	if err != nil {
		// Clean up on failure
		if cleanupErr := akp.kubernetesManager.DeletePerformer(ctx, performerID); cleanupErr != nil {
			akp.logger.Error("Failed to clean up performer after connection failure",
				zap.String("performerID", performerID),
				zap.Error(cleanupErr),
			)
		}
		return nil, fmt.Errorf("failed to establish gRPC connection: %w", err)
	}
	
	// Create gRPC client using the connection
	client := performerV1.NewPerformerServiceClient(conn)

	performerResource := &PerformerResource{
		performerID:       performerID,
		avsAddress:        akp.config.AvsAddress,
		image:             image,
		status:            avsPerformer.PerformerResourceStatusStaged,
		client:            client,
		connectionManager: connectionManager,
		endpoint:          createResponse.Endpoint,
		statusChan:        make(chan avsPerformer.PerformerStatusEvent, 10),
		createdAt:         time.Now(),
		performerHealth: &avsPerformer.PerformerHealth{
			ContainerIsHealthy:   true,
			ApplicationIsHealthy: true,
			LastHealthCheck:      time.Now(),
		},
	}

	akp.logger.Info("Kubernetes performer resource created successfully",
		zap.String("performerID", performerID),
		zap.String("endpoint", createResponse.Endpoint),
	)

	return performerResource, nil
}

// waitForPerformerReady waits for the performer to be ready
func (akp *AvsKubernetesPerformer) waitForPerformerReady(ctx context.Context, performerID string) error {
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

			akp.logger.Debug("Performer status check",
				zap.String("performerID", performerID),
				zap.String("phase", string(status.Phase)),
				zap.Bool("ready", status.Ready),
			)

			if status.Ready {
				akp.logger.Info("Performer is ready",
					zap.String("performerID", performerID),
					zap.String("phase", string(status.Phase)),
				)
				return nil
			}
		}
	}
}

// monitorPerformerHealth monitors the health of a performer
func (akp *AvsKubernetesPerformer) monitorPerformerHealth(ctx context.Context, performer *PerformerResource) {
	ticker := time.NewTicker(akp.config.ApplicationHealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			akp.performApplicationHealthCheck(ctx, performer)
		}
	}
}

// performApplicationHealthCheck checks the health of a performer
func (akp *AvsKubernetesPerformer) performApplicationHealthCheck(ctx context.Context, performer *PerformerResource) {
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Check if circuit breaker is open
	if performer.connectionManager.IsCircuitOpen() {
		akp.logger.Warn("Circuit breaker is open for performer, skipping health check",
			zap.String("performerID", performer.performerID),
			zap.Any("connectionStats", performer.connectionManager.GetConnectionStats()),
		)
		performer.performerHealth.ApplicationIsHealthy = false
		performer.performerHealth.ConsecutiveApplicationHealthFailures++
		return
	}

	// Get a healthy connection
	conn, err := performer.connectionManager.GetConnection()
	if err != nil {
		akp.logger.Warn("Failed to get healthy connection for health check",
			zap.String("performerID", performer.performerID),
			zap.Error(err),
		)
		performer.performerHealth.ApplicationIsHealthy = false
		performer.performerHealth.ConsecutiveApplicationHealthFailures++
		return
	}

	// Create a new client with the healthy connection
	client := performerV1.NewPerformerServiceClient(conn)

	_, err = client.HealthCheck(healthCtx, &performerV1.HealthCheckRequest{})
	performer.lastHealthCheck = time.Now()
	performer.performerHealth.LastHealthCheck = time.Now()

	if err != nil {
		// Health check failed
		performer.performerHealth.ApplicationIsHealthy = false
		performer.performerHealth.ConsecutiveApplicationHealthFailures++

		akp.logger.Warn("Application health check failed",
			zap.String("performerID", performer.performerID),
			zap.Error(err),
			zap.Int("consecutiveFailures", performer.performerHealth.ConsecutiveApplicationHealthFailures),
			zap.Any("connectionStats", performer.connectionManager.GetConnectionStats()),
		)

		// Send unhealthy status event
		if performer.statusChan != nil {
			select {
			case performer.statusChan <- avsPerformer.PerformerStatusEvent{
				Status:      avsPerformer.PerformerUnhealthy,
				PerformerID: performer.performerID,
				Message:     fmt.Sprintf("Health check failed: %v", err),
				Timestamp:   time.Now(),
			}:
			default:
			}
		}

		// Handle consecutive failures
		if performer.performerHealth.ConsecutiveApplicationHealthFailures >= maxConsecutiveApplicationHealthFailures {
			akp.logger.Error("Performer failed consecutive health checks",
				zap.String("performerID", performer.performerID),
				zap.Int("consecutiveFailures", performer.performerHealth.ConsecutiveApplicationHealthFailures),
			)
			// In Kubernetes, we could potentially restart the pod or recreate the performer
			// For now, just log the failure - the operator will handle restarts
		}
	} else {
		// Health check succeeded
		performer.performerHealth.ApplicationIsHealthy = true
		performer.performerHealth.ConsecutiveApplicationHealthFailures = 0

		// Send healthy status event
		if performer.statusChan != nil {
			select {
			case performer.statusChan <- avsPerformer.PerformerStatusEvent{
				Status:      avsPerformer.PerformerHealthy,
				PerformerID: performer.performerID,
				Message:     "Performer is healthy",
				Timestamp:   time.Now(),
			}:
			default:
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

	// Start monitoring for the new performer
	go akp.monitorPerformerHealth(ctx, newPerformer)

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

	// Verify next performer is healthy
	if !akp.nextPerformer.performerHealth.ApplicationIsHealthy {
		return fmt.Errorf("cannot promote unhealthy performer %s", performerID)
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
		currentPerformer := current.(*PerformerResource)
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

	// Close connection manager
	if targetPerformer.connectionManager != nil {
		if err := targetPerformer.connectionManager.Close(); err != nil {
			akp.logger.Warn("Failed to close connection manager",
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

	currentPerformer := current.(*PerformerResource)
	if currentPerformer == nil || currentPerformer.connectionManager == nil {
		return nil, fmt.Errorf("no current performer connection manager available to execute task")
	}

	// Check if circuit breaker is open
	if currentPerformer.connectionManager.IsCircuitOpen() {
		return nil, fmt.Errorf("circuit breaker is open for performer %s, task execution unavailable", currentPerformer.performerID)
	}

	// Track this task with the performer's WaitGroup
	wg := akp.getOrCreateTaskWaitGroup(currentPerformer.performerID)
	wg.Add(1)
	defer wg.Done()

	// Get a healthy connection for task execution
	conn, err := currentPerformer.connectionManager.GetConnection()
	if err != nil {
		akp.logger.Error("Failed to get healthy connection for task execution",
			zap.String("performerID", currentPerformer.performerID),
			zap.String("taskID", task.TaskID),
			zap.Error(err),
			zap.Any("connectionStats", currentPerformer.connectionManager.GetConnectionStats()),
		)
		return nil, fmt.Errorf("failed to get healthy connection: %w", err)
	}

	// Create a new client with the healthy connection
	client := performerV1.NewPerformerServiceClient(conn)

	// Execute the task
	res, err := client.ExecuteTask(ctx, &performerV1.TaskRequest{
		TaskId:  []byte(task.TaskID),
		Payload: task.Payload,
	})
	if err != nil {
		akp.logger.Error("Performer failed to handle task",
			zap.String("performerID", currentPerformer.performerID),
			zap.String("taskID", task.TaskID),
			zap.Error(err),
			zap.Any("connectionStats", currentPerformer.connectionManager.GetConnectionStats()),
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

	// Monitor deployment until healthy
	if err := akp.waitForHealthy(deploymentCtx, creationResult.StatusChan); err != nil {
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

// waitForHealthy waits for the performer to become healthy
func (akp *AvsKubernetesPerformer) waitForHealthy(ctx context.Context, statusChan <-chan avsPerformer.PerformerStatusEvent) error {
	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("deployment timeout")
			}
			return ctx.Err()
		case status, ok := <-statusChan:
			if !ok {
				return fmt.Errorf("status channel closed unexpectedly")
			}

			switch status.Status {
			case avsPerformer.PerformerHealthy:
				return nil
			case avsPerformer.PerformerUnhealthy:
				akp.logger.Warn("Performer is unhealthy, continuing to monitor",
					zap.String("performerID", status.PerformerID),
					zap.String("message", status.Message),
				)
			}
		}
	}
}

// ListPerformers returns information about current and next performers
func (akp *AvsKubernetesPerformer) ListPerformers() []avsPerformer.PerformerMetadata {
	akp.performerResourcesMu.Lock()
	defer akp.performerResourcesMu.Unlock()

	var performers []avsPerformer.PerformerMetadata

	// Add current performer
	if current := akp.currentPerformer.Load(); current != nil {
		currentPerformer := current.(*PerformerResource)
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
		ContainerHealthy:   performer.performerHealth.ContainerIsHealthy,
		ApplicationHealthy: performer.performerHealth.ApplicationIsHealthy,
		LastHealthCheck:    performer.performerHealth.LastHealthCheck,
		ResourceID:         performer.performerID, // In K8s, we use performer ID as resource ID
	}
}

// getOrCreateTaskWaitGroup returns the WaitGroup for a performer
func (akp *AvsKubernetesPerformer) getOrCreateTaskWaitGroup(performerID string) *sync.WaitGroup {
	akp.taskWaitGroupsMu.Lock()
	defer akp.taskWaitGroupsMu.Unlock()

	wg, exists := akp.taskWaitGroups[performerID]
	if !exists {
		wg = &sync.WaitGroup{}
		akp.taskWaitGroups[performerID] = wg
	}
	return wg
}

// waitForTaskCompletion waits for all tasks on a performer to complete
func (akp *AvsKubernetesPerformer) waitForTaskCompletion(performerID string) {
	akp.taskWaitGroupsMu.Lock()
	wg, exists := akp.taskWaitGroups[performerID]
	akp.taskWaitGroupsMu.Unlock()

	if exists && wg != nil {
		wg.Wait()
	}
}

// cleanupTaskWaitGroup removes the WaitGroup for a performer
func (akp *AvsKubernetesPerformer) cleanupTaskWaitGroup(performerID string) {
	akp.taskWaitGroupsMu.Lock()
	defer akp.taskWaitGroupsMu.Unlock()
	delete(akp.taskWaitGroups, performerID)
}

// startDrainAndRemove initiates draining of a performer in a separate goroutine
func (akp *AvsKubernetesPerformer) startDrainAndRemove(performer *PerformerResource) {
	performerID := performer.performerID

	// Check if already draining
	akp.drainingPerformersMu.Lock()
	if _, exists := akp.drainingPerformers[performerID]; exists {
		akp.drainingPerformersMu.Unlock()
		return
	}
	akp.drainingPerformers[performerID] = struct{}{}
	akp.drainingPerformersMu.Unlock()

	go func() {
		defer func() {
			akp.drainingPerformersMu.Lock()
			delete(akp.drainingPerformers, performerID)
			akp.drainingPerformersMu.Unlock()
		}()

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

		// Close connection manager
		if performer.connectionManager != nil {
			if err := performer.connectionManager.Close(); err != nil {
				akp.logger.Warn("Failed to close connection manager during drain",
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

	// Clear draining performers
	akp.drainingPerformersMu.Lock()
	for performerID := range akp.drainingPerformers {
		delete(akp.drainingPerformers, performerID)
	}
	akp.drainingPerformersMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var errs []error

	// Shutdown current performer
	if current := akp.currentPerformer.Load(); current != nil {
		currentPerformer := current.(*PerformerResource)
		
		// Close connection manager
		if currentPerformer.connectionManager != nil {
			if err := currentPerformer.connectionManager.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close current performer connection manager: %w", err))
			}
		}
		
		if err := akp.kubernetesManager.DeletePerformer(ctx, currentPerformer.performerID); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown current performer: %w", err))
		}
		akp.cleanupTaskWaitGroup(currentPerformer.performerID)
		akp.currentPerformer.Store((*PerformerResource)(nil))
	}

	// Shutdown next performer
	if akp.nextPerformer != nil {
		// Close connection manager
		if akp.nextPerformer.connectionManager != nil {
			if err := akp.nextPerformer.connectionManager.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close next performer connection manager: %w", err))
			}
		}
		
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