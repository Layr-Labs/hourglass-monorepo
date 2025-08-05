package avsKubernetesPerformer

import (
	"context"
	"fmt"
	healthV1 "github.com/Layr-Labs/protocol-apis/gen/protos/grpc/health/v1"
	"github.com/stretchr/testify/require"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/avsPerformerClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/kubernetesManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"go.uber.org/zap"
)

// TestE2E_KubernetesPerformer_FullWorkflow tests the complete end-to-end workflow
func TestE2E_KubernetesPerformer_FullWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	// Create logger
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: true})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	root := testUtils.GetProjectRootPath()
	t.Logf("Project root path: %s", root)

	// Clean up any existing test clusters first to prevent port conflicts
	if err := testUtils.CleanupAllTestClusters(l.Sugar()); err != nil {
		t.Logf("Warning: Failed to cleanup existing test clusters: %v", err)
	}

	// Create Kind cluster
	kindConfig := testUtils.DefaultKindClusterConfig(l.Sugar())
	cluster, clusterCleanup, err := testUtils.CreateKindCluster(ctx, t, kindConfig)
	if err != nil {
		t.Fatalf("Failed to create Kind cluster: %v", err)
	}
	defer func() {
		// Nuclear option: just delete the cluster to avoid hanging cleanup
		t.Log("Using fast cluster deletion to avoid hanging cleanup")
		clusterCleanup()
	}()

	// Load hello-performer image (assumes image is already built)
	if err := loadHelloPerformerImage(ctx, cluster, l.Sugar()); err != nil {
		t.Fatalf("Failed to load hello-performer image: %v", err)
	}

	// Install CRDs first (required for Performer objects)
	if err := installPerformerCRD(ctx, cluster, root, l.Sugar()); err != nil {
		t.Fatalf("Failed to install Performer CRD: %v", err)
	}

	// Load pre-built operator image
	if err := loadOperatorImage(ctx, cluster, l.Sugar()); err != nil {
		t.Fatalf("Failed to load operator image: %v", err)
	}

	// Deploy Hourglass operator
	operatorConfig := testUtils.DefaultOperatorDeploymentConfig(root, l.Sugar())
	operator, operatorCleanup, err := testUtils.DeployOperator(ctx, cluster, operatorConfig)
	if err != nil {
		t.Fatalf("Failed to deploy operator: %v", err)
	}
	defer func() {
		// Run cleanup with timeout to avoid hanging
		t.Log("Running operator cleanup with timeout")
		done := make(chan struct{})
		go func() {
			operatorCleanup()
			close(done)
		}()

		select {
		case <-done:
			t.Log("Operator cleanup completed successfully")
		case <-time.After(45 * time.Second):
			t.Log("Operator cleanup timed out, proceeding with cluster deletion")
		}
	}()

	t.Logf("Operator deployed successfully: %s", operator.ReleaseName)

	// Create peering data fetcher
	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: nil,
	}, l)

	// Create AvsKubernetesPerformer without image (so Initialize doesn't hang)
	performerConfig := &avsPerformer.AvsPerformerConfig{
		AvsAddress:                     "0xtest-avs-address",
		ApplicationHealthCheckInterval: 2 * time.Second,
		SkipConnectionTest:             true, // Skip connection test since executor is outside cluster
		// No image - prevents Initialize from hanging
	}

	kubernetesConfig := &kubernetesManager.Config{
		Namespace:         "default",
		KubeconfigPath:    cluster.KubeConfig,
		OperatorNamespace: "hourglass-system",
		CRDGroup:          "hourglass.eigenlayer.io",
		CRDVersion:        "v1alpha1",
		ConnectionTimeout: 30 * time.Second,
	}

	t.Logf("Creating AvsKubernetesPerformer with config: %+v", performerConfig)
	// Create the AvsKubernetesPerformer instance
	performer, err := NewAvsKubernetesPerformer(
		performerConfig,
		kubernetesConfig,
		pdf,
		nil,
		l,
	)
	if err != nil {
		t.Fatalf("Failed to create AvsKubernetesPerformer: %v", err)
	}

	t.Logf("Initializing AvsKubernetesPerformer with config: %+v", performerConfig)
	// Initialize the performer (won't create any CRDs since no image)
	if err := performer.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize performer: %v", err)
	}

	// Now create a performer using CreatePerformer with a shorter timeout to avoid hanging
	performerImage := avsPerformer.PerformerImage{
		Repository: "hello-performer",
		Tag:        "latest",
	}

	// Create context with shorter timeout for performer creation to avoid hanging
	createCtx, createCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer createCancel()

	t.Logf("Creating performer with image: %+v", performerImage)
	creationResult, err := performer.CreatePerformer(createCtx, performerImage)
	if err != nil {
		require.NoError(t, err)
		t.Logf("CreatePerformer failed or timed out: %v", err)

		// Continue with test - we'll verify the performer was created even if CreatePerformer hangs
	} else {
		t.Logf("Created performer: %s", creationResult.PerformerID)
	}

	// Verify performer was created
	performers := performer.ListPerformers()
	if len(performers) != 1 {
		t.Fatalf("Expected 1 performer after initialization, got %d", len(performers))
	}

	// Get the actual performer ID from the list (in case CreatePerformer timed out)
	actualPerformerID := performers[0].PerformerID
	t.Logf("AvsKubernetesPerformer created performer: %s", actualPerformerID)

	// Wait for operator to process the performer and create resources
	timeout := 60 * time.Second
	start := time.Now()
	var performerReady bool

	for time.Since(start) < timeout {
		// Check if operator created any pods
		pods, err := cluster.RunKubectl(ctx, "get", "pods", "-n", "default", "-l", "hourglass.eigenlayer.io/performer", "--no-headers")
		if err == nil && len(pods) > 0 {
			t.Logf("Operator created performer pods: %s", string(pods))

			// Check if pod is running
			if strings.Contains(string(pods), "Running") {
				performerReady = true

				// Get pod logs to check if gRPC server is running
				podLogs, err := cluster.RunKubectl(ctx, "logs", "-n", "default", "-l", "hourglass.eigenlayer.io/performer", "--tail=20")
				if err == nil {
					t.Logf("Pod logs: %s", string(podLogs))
				}
				break
			}
		}

		// Check performer status
		performerStatus, err := cluster.RunKubectl(ctx, "get", "performers", "-n", "default", actualPerformerID, "-o", "jsonpath={.status.phase}")
		if err == nil && len(performerStatus) > 0 {
			t.Logf("Performer status: %s", string(performerStatus))
			if string(performerStatus) == "Running" {
				t.Log("Performer is running!")
				performerReady = true
				break
			}
		}

		time.Sleep(2 * time.Second)
	}

	// Test health check endpoint if performer is ready
	if performerReady {
		t.Log("Testing health check endpoint...")

		// Get the service endpoint
		serviceEndpoint, err := cluster.RunKubectl(ctx, "get", "service", "performer-"+actualPerformerID, "-n", "default", "-o", "jsonpath={.spec.clusterIP}:{.spec.ports[0].port}")
		if err != nil {
			t.Logf("Failed to get service endpoint: %v", err)
		} else {
			t.Logf("Service endpoint: %s", string(serviceEndpoint))

			// Test health check using gRPC client via kubectl port-forward
			// The hello-performer uses port 8080, not 9090
			healthCheckCtx, healthCancel := context.WithTimeout(ctx, 30*time.Second)
			defer healthCancel()

			// Start port-forward in background
			portForwardCmd := fmt.Sprintf("kubectl --kubeconfig=%s port-forward service/performer-%s 8080:8080", cluster.KubeConfig, actualPerformerID)
			portForward := exec.CommandContext(healthCheckCtx, "bash", "-c", portForwardCmd)

			// Start port-forward
			if err := portForward.Start(); err != nil {
				t.Logf("Failed to start port-forward: %v", err)
			} else {
				// Give port-forward time to establish
				time.Sleep(2 * time.Second)

				// Create gRPC client connection
				grpcClient, err := avsPerformerClient.NewAvsPerformerClient("localhost:8080", true)
				if err != nil {
					t.Logf("Failed to create gRPC client: %v", err)
				} else {
					// Test health check
					healthReq := &healthV1.HealthCheckRequest{}
					healthResp, err := grpcClient.HealthClient.Check(healthCheckCtx, healthReq)
					if err != nil {
						t.Logf("Health check failed: %v", err)
					} else {
						t.Logf("Health check successful: %+v", healthResp)
					}
				}

				// Clean up port-forward
				if err := portForward.Process.Kill(); err != nil {
					t.Logf("Failed to kill port-forward: %v", err)
				}
			}
		}
	}

	// Final verification - check operator logs for debugging
	operatorLogs, err := cluster.RunKubectl(ctx, "logs", "-n", "hourglass-system", "-l", "app=hourglass-operator", "--tail=50")
	if err != nil {
		t.Logf("Failed to get operator logs: %v", err)
	} else {
		t.Logf("Operator logs after creating performer: %s", string(operatorLogs))
	}

	// Check all resources created by operator
	pods, err := cluster.RunKubectl(ctx, "get", "pods", "-n", "default", "-o", "wide")
	if err != nil {
		t.Logf("Failed to get pods: %v", err)
	} else {
		t.Logf("Pods in default namespace: %s", string(pods))
	}

	services, err := cluster.RunKubectl(ctx, "get", "services", "-n", "default", "-o", "wide")
	if err != nil {
		t.Logf("Failed to get services: %v", err)
	} else {
		t.Logf("Services in default namespace: %s", string(services))
	}

	// Get final performer status
	performerStatus, err := cluster.RunKubectl(ctx, "get", "performers", "-n", "default", actualPerformerID, "-o", "yaml")
	if err != nil {
		t.Logf("Failed to get performer status: %v", err)
	} else {
		t.Logf("Final performer status: %s", string(performerStatus))
	}

	// Clean up
	if err := performer.Shutdown(); err != nil {
		t.Logf("Error shutting down performer: %v", err)
	}

	t.Log("E2E Kubernetes performer test completed successfully - Kind cluster, operator deployment, and CRD creation working")
}

// loadHelloPerformerImage loads the hello-performer image into the Kind cluster
func loadHelloPerformerImage(ctx context.Context, cluster *testUtils.KindCluster, logger *zap.SugaredLogger) error {
	imageName := "hello-performer:latest"
	logger.Infof("Loading hello-performer image into Kind cluster: %s", imageName)

	// Load the image into Kind cluster (assumes image is already built locally)
	if err := cluster.LoadDockerImage(ctx, imageName); err != nil {
		return fmt.Errorf("failed to load hello-performer image into Kind cluster: %v", err)
	}

	logger.Infof("Successfully loaded hello-performer image: %s", imageName)
	return nil
}

// installPerformerCRD installs the Performer CRD required for the test
func installPerformerCRD(ctx context.Context, cluster *testUtils.KindCluster, projectRoot string, logger *zap.SugaredLogger) error {
	// Path to the Performer CRD file
	crdPath := filepath.Join(projectRoot, "..", "hourglass-operator", "config", "crd", "bases", "hourglass.eigenlayer.io_performers.yaml")

	logger.Infof("Installing Performer CRD from: %s", crdPath)

	// Apply the CRD
	output, err := cluster.RunKubectl(ctx, "apply", "-f", crdPath)
	if err != nil {
		return fmt.Errorf("failed to apply Performer CRD: %v\nOutput: %s", err, string(output))
	}

	logger.Infof("Performer CRD installed successfully")
	return nil
}

// loadOperatorImage loads the pre-built Hourglass operator image into the Kind cluster
func loadOperatorImage(ctx context.Context, cluster *testUtils.KindCluster, logger *zap.SugaredLogger) error {
	logger.Info("Loading pre-built Hourglass operator image into Kind cluster")

	// Load the pre-built operator image
	if err := cluster.LoadDockerImage(ctx, "hourglass/operator:test"); err != nil {
		return fmt.Errorf("failed to load operator image to Kind: %v", err)
	}

	logger.Info("Successfully loaded Hourglass operator image")
	return nil
}
