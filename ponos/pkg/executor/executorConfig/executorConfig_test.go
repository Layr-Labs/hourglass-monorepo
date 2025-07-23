package executorConfig

import (
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ExecutorConfig(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		t.Run("Should parse a valid json config with operator and avss", func(t *testing.T) {
			ec, err := NewExecutorConfigFromYamlBytes([]byte(jsonValid))
			assert.Nil(t, err)
			assert.NotNil(t, ec)
			assert.Equal(t, "0xoperator...", ec.Operator.Address)
			assert.Equal(t, "...", ec.Operator.OperatorPrivateKey.PrivateKey)
			assert.NotNil(t, ec.Operator.SigningKeys.BLS)

			assert.Equal(t, "v1.0.0", ec.AvsPerformers[0].Image.Tag)
			assert.Equal(t, "eigenlabs/avs", ec.AvsPerformers[0].Image.Repository)
			assert.Equal(t, "server", ec.AvsPerformers[0].ProcessType)
			assert.Equal(t, "0xavs1...", ec.AvsPerformers[0].AvsAddress)

		})
		t.Run("Should fail to parse an invalid yaml config with invalid fields", func(t *testing.T) {
			_, err := NewExecutorConfigFromYamlBytes([]byte(jsonInvalid))
			assert.NotNil(t, err)

		})
	})
	t.Run("YAML", func(t *testing.T) {
		t.Run("Should parse a valid yaml config with operator and avss", func(t *testing.T) {
			ec, err := NewExecutorConfigFromYamlBytes([]byte(yamlValid))
			assert.Nil(t, err)
			assert.NotNil(t, ec)
			assert.Equal(t, "0xoperator...", ec.Operator.Address)
			assert.Equal(t, "...", ec.Operator.OperatorPrivateKey.PrivateKey)
			assert.NotNil(t, ec.Operator.SigningKeys.BLS)

			assert.Equal(t, "v1.0.0", ec.AvsPerformers[0].Image.Tag)
			assert.Equal(t, "eigenlabs/avs", ec.AvsPerformers[0].Image.Repository)
			assert.Equal(t, "server", ec.AvsPerformers[0].ProcessType)
			assert.Equal(t, "0xavs1...", ec.AvsPerformers[0].AvsAddress)

			assert.NotEmpty(t, ec.OverrideContracts.TaskMailbox.Contract)
		})
		t.Run("Should fail to parse an invalid yaml config with invalid fields", func(t *testing.T) {
			_, err := NewExecutorConfigFromYamlBytes([]byte(yamlInvalid))
			assert.NotNil(t, err)

		})
	})
}

// TestDeploymentMode tests the deployment mode functionality
func TestDeploymentMode(t *testing.T) {
	t.Run("Should default to docker mode when not specified", func(t *testing.T) {
		config := &AvsPerformerConfig{
			AvsAddress:          "0x123",
			ProcessType:         "server",
			AVSRegistrarAddress: "0x456",
			Image: &PerformerImage{
				Repository: "test/image",
				Tag:        "v1.0.0",
			},
		}

		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, DeploymentModeDocker, config.DeploymentMode)
	})

	t.Run("Should accept kubernetes mode", func(t *testing.T) {
		config := &AvsPerformerConfig{
			AvsAddress:          "0x123",
			ProcessType:         "server",
			AVSRegistrarAddress: "0x456",
			DeploymentMode:      DeploymentModeKubernetes,
			Image: &PerformerImage{
				Repository: "test/image",
				Tag:        "v1.0.0",
			},
		}

		err := config.Validate()
		require.NoError(t, err)
		assert.Equal(t, DeploymentModeKubernetes, config.DeploymentMode)
	})

	t.Run("Should reject invalid deployment mode", func(t *testing.T) {
		config := &AvsPerformerConfig{
			AvsAddress:          "0x123",
			ProcessType:         "server",
			AVSRegistrarAddress: "0x456",
			DeploymentMode:      "invalid",
			Image: &PerformerImage{
				Repository: "test/image",
				Tag:        "v1.0.0",
			},
		}

		err := config.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deploymentMode must be one of [docker, kubernetes]")
	})
}

// TestKubernetesConfig tests the Kubernetes configuration
func TestKubernetesConfig(t *testing.T) {
	t.Run("Should create default kubernetes config", func(t *testing.T) {
		config := NewDefaultKubernetesConfig()

		assert.Equal(t, "default", config.Namespace)
		assert.Equal(t, "hourglass-system", config.OperatorNamespace)
		assert.Equal(t, "hourglass.eigenlayer.io", config.CRDGroup)
		assert.Equal(t, "v1alpha1", config.CRDVersion)
		assert.Equal(t, 30*time.Second, config.ConnectionTimeout)
		assert.True(t, config.InCluster)
		assert.Empty(t, config.KubeConfigPath)
	})

	t.Run("Should validate required fields", func(t *testing.T) {
		config := &KubernetesConfig{}

		err := config.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "namespace is required")
		assert.Contains(t, err.Error(), "operatorNamespace is required")
		assert.Contains(t, err.Error(), "crdGroup is required")
		assert.Contains(t, err.Error(), "crdVersion is required")
		assert.Contains(t, err.Error(), "connectionTimeout is required")
	})

	t.Run("Should require kubeconfig path when not in cluster", func(t *testing.T) {
		config := &KubernetesConfig{
			Namespace:         "test",
			OperatorNamespace: "hourglass-system",
			CRDGroup:          "hourglass.eigenlayer.io",
			CRDVersion:        "v1alpha1",
			ConnectionTimeout: 30 * time.Second,
			InCluster:         false,
		}

		err := config.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kubeConfigPath is required when not running in cluster")
	})

	t.Run("Should validate successfully with all fields", func(t *testing.T) {
		config := &KubernetesConfig{
			Namespace:         "test",
			OperatorNamespace: "hourglass-system",
			CRDGroup:          "hourglass.eigenlayer.io",
			CRDVersion:        "v1alpha1",
			ConnectionTimeout: 30 * time.Second,
			InCluster:         true,
		}

		err := config.Validate()
		require.NoError(t, err)
	})
}

// TestExecutorConfigKubernetes tests the executor configuration with Kubernetes support
func TestExecutorConfigKubernetes(t *testing.T) {
	t.Run("Should require kubernetes config when performer uses kubernetes mode", func(t *testing.T) {
		config := &ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: "0x123",
				OperatorPrivateKey: &config.ECDSAKeyConfig{
					PrivateKey: "private_key",
				},
				SigningKeys: config.SigningKeys{
					BLS: &config.SigningKey{
						Keystore: "keystore_content",
						Password: "password",
					},
				},
			},
			AvsPerformers: []*AvsPerformerConfig{
				{
					AvsAddress:          "0x456",
					ProcessType:         "server",
					AVSRegistrarAddress: "0x789",
					DeploymentMode:      DeploymentModeKubernetes,
					Image: &PerformerImage{
						Repository: "test/image",
						Tag:        "v1.0.0",
					},
				},
			},
			L1Chain: &Chain{
				RpcUrl:  "http://localhost:8545",
				ChainId: 1,
			},
		}

		err := config.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kubernetes configuration is required when using kubernetes deployment mode")
	})

	t.Run("Should validate successfully with kubernetes config", func(t *testing.T) {
		config := &ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: "0x123",
				OperatorPrivateKey: &config.ECDSAKeyConfig{
					PrivateKey: "private_key",
				},
				SigningKeys: config.SigningKeys{
					BLS: &config.SigningKey{
						Keystore: "keystore_content",
						Password: "password",
					},
				},
			},
			AvsPerformers: []*AvsPerformerConfig{
				{
					AvsAddress:          "0x456",
					ProcessType:         "server",
					AVSRegistrarAddress: "0x789",
					DeploymentMode:      DeploymentModeKubernetes,
					Image: &PerformerImage{
						Repository: "test/image",
						Tag:        "v1.0.0",
					},
				},
			},
			L1Chain: &Chain{
				RpcUrl:  "http://localhost:8545",
				ChainId: 1,
			},
			Kubernetes: NewDefaultKubernetesConfig(),
		}

		err := config.Validate()
		require.NoError(t, err)
	})

	t.Run("Should not allow mixed deployment modes", func(t *testing.T) {
		config := &ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: "0x123",
				OperatorPrivateKey: &config.ECDSAKeyConfig{
					PrivateKey: "private_key",
				},
				SigningKeys: config.SigningKeys{
					BLS: &config.SigningKey{
						Keystore: "keystore_content",
						Password: "password",
					},
				},
			},
			AvsPerformers: []*AvsPerformerConfig{
				{
					AvsAddress:          "0x456",
					ProcessType:         "server",
					AVSRegistrarAddress: "0x789",
					DeploymentMode:      DeploymentModeDocker,
					Image: &PerformerImage{
						Repository: "test/image",
						Tag:        "v1.0.0",
					},
				},
				{
					AvsAddress:          "0xabc",
					ProcessType:         "server",
					AVSRegistrarAddress: "0xdef",
					DeploymentMode:      DeploymentModeKubernetes,
					Image: &PerformerImage{
						Repository: "test/image2",
						Tag:        "v1.0.0",
					},
				},
			},
			L1Chain: &Chain{
				RpcUrl:  "http://localhost:8545",
				ChainId: 1,
			},
			Kubernetes: NewDefaultKubernetesConfig(),
		}

		err := config.Validate()
		require.Error(t, err)
	})
}

// TestConfigSerialization tests YAML/JSON serialization with new fields
func TestConfigSerialization(t *testing.T) {
	t.Run("Should parse kubernetes config from YAML", func(t *testing.T) {
		yamlWithKubernetes := `
operator:
  address: "0xoperator..."
  operatorPrivateKey:
    privateKey: "..."
  signingKeys:
    bls: 
      keystore: ""
      password: ""
avsPerformers:
- image:
    repository: "eigenlabs/avs"
    tag: "v1.0.0"
  processType: "server"
  avsAddress: "0xavs1..."
  deploymentMode: "kubernetes"
  avsRegistrarAddress: "0x789"
l1Chain:
  rpcUrl: "http://localhost:8545"
  chainId: 1
kubernetes:
  namespace: "test-namespace"
  operatorNamespace: "hourglass-system"
  crdGroup: "hourglass.eigenlayer.io"
  crdVersion: "v1alpha1"
  connectionTimeout: 30000000000
  inCluster: true
`

		config, err := NewExecutorConfigFromYamlBytes([]byte(yamlWithKubernetes))
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Equal(t, DeploymentModeKubernetes, config.AvsPerformers[0].DeploymentMode)
		assert.NotNil(t, config.Kubernetes)
		assert.Equal(t, "test-namespace", config.Kubernetes.Namespace)
		assert.Equal(t, "hourglass-system", config.Kubernetes.OperatorNamespace)
		assert.Equal(t, "hourglass.eigenlayer.io", config.Kubernetes.CRDGroup)
		assert.Equal(t, "v1alpha1", config.Kubernetes.CRDVersion)
		assert.Equal(t, 30*time.Second, config.Kubernetes.ConnectionTimeout)
		assert.True(t, config.Kubernetes.InCluster)
	})

	t.Run("Should parse docker config from YAML (backward compatibility)", func(t *testing.T) {
		yamlWithDocker := `
operator:
  address: "0xoperator..."
  operatorPrivateKey:
    privateKey: "..."
  signingKeys:
    bls: 
      keystore: ""
      password: ""
avsPerformers:
- image:
    repository: "eigenlabs/avs"
    tag: "v1.0.0"
  processType: "server"
  avsAddress: "0xavs1..."
  deploymentMode: "docker"
  avsRegistrarAddress: "0x789"
l1Chain:
  rpcUrl: "http://localhost:8545"
  chainId: 1
`

		config, err := NewExecutorConfigFromYamlBytes([]byte(yamlWithDocker))
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Equal(t, DeploymentModeDocker, config.AvsPerformers[0].DeploymentMode)
		assert.Nil(t, config.Kubernetes)
	})
}

const (
	yamlValid = `
---
operator:
  address: "0xoperator..."
  operatorPrivateKey:
    privateKey: "..."
  signingKeys:
    bls: 
        keystore: ""
        password: ""
avsPerformers:
- image:
    repository: "eigenlabs/avs"
    tag: "v1.0.0"
  processType: "server"
  avsAddress: "0xavs1..."
overrideContracts:
  taskMailbox:
    chainIds: [31337]
    contract: |
      {
          "name": "TaskMailbox",
          "address": "0x7306a649b451ae08781108445425bd4e8acf1e00",
          "chainId": 31337,
          "abiVersions": [
              "[{\"type\":\"function\",\"name\":\"cancelTask\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createTask\",\"inputs\":[{\"name\":\"taskParams\",\"type\":\"tuple\",\"internalType\":\"struct ITaskMailboxTypes.TaskParams\",\"components\":[{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"executorOperatorSet\",\"type\":\"tuple\",\"internalType\":\"struct OperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAvsConfig\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"struct ITaskMailboxTypes.AvsConfig\",\"components\":[{\"name\":\"resultSubmitter\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"executorOperatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getExecutorOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"struct OperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"struct ITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contract IAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contract IERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskInfo\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"struct ITaskMailboxTypes.Task\",\"components\":[{\"name\":\"creator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"creationTime\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enum ITaskMailboxTypes.TaskStatus\"},{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"resultSubmitter\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"refundCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"feeSplit\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"executorOperatorSetTaskConfig\",\"type\":\"tuple\",\"internalType\":\"struct ITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contract IAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contract IERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"payload\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskResult\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskStatus\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enum ITaskMailboxTypes.TaskStatus\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isAvsRegistered\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isExecutorOperatorSetRegistered\",\"inputs\":[{\"name\":\"operatorSetKey\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerAvs\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"isRegistered\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAvsConfig\",\"inputs\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"struct ITaskMailboxTypes.AvsConfig\",\"components\":[{\"name\":\"resultSubmitter\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"executorOperatorSetIds\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setExecutorOperatorSetTaskConfig\",\"inputs\":[{\"name\":\"operatorSet\",\"type\":\"tuple\",\"internalType\":\"struct OperatorSet\",\"components\":[{\"name\":\"avs\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"id\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]},{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"struct ITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contract IAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contract IERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitResult\",\"inputs\":[{\"name\":\"taskHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"cert\",\"type\":\"tuple\",\"internalType\":\"struct IBN254CertificateVerifier.BN254Certificate\",\"components\":[{\"name\":\"referenceTimestamp\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"sig\",\"type\":\"tuple\",\"internalType\":\"struct BN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"apk\",\"type\":\"tuple\",\"internalType\":\"struct BN254.G2Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"Y\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"nonsignerIndices\",\"type\":\"uint32[]\",\"internalType\":\"uint32[]\"},{\"name\":\"nonSignerWitnesses\",\"type\":\"tuple[]\",\"internalType\":\"struct IBN254CertificateVerifier.BN254OperatorInfoWitness[]\",\"components\":[{\"name\":\"operatorIndex\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"operatorInfoProofs\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"operatorInfo\",\"type\":\"tuple\",\"internalType\":\"struct IBN254CertificateVerifier.BN254OperatorInfo\",\"components\":[{\"name\":\"pubkey\",\"type\":\"tuple\",\"internalType\":\"struct BN254.G1Point\",\"components\":[{\"name\":\"X\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"Y\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"weights\",\"type\":\"uint96[]\",\"internalType\":\"uint96[]\"}]}]}]},{\"name\":\"result\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AvsConfigSet\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"resultSubmitter\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"aggregatorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"executorOperatorSetIds\",\"type\":\"uint32[]\",\"indexed\":false,\"internalType\":\"uint32[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AvsRegistered\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"isRegistered\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ExecutorOperatorSetTaskConfigSet\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"config\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"struct ITaskMailboxTypes.ExecutorOperatorSetTaskConfig\",\"components\":[{\"name\":\"certificateVerifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskHook\",\"type\":\"address\",\"internalType\":\"contract IAVSTaskHook\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"contract IERC20\"},{\"name\":\"feeCollector\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskSLA\",\"type\":\"uint96\",\"internalType\":\"uint96\"},{\"name\":\"stakeProportionThreshold\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"taskMetadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskCanceled\",\"inputs\":[{\"name\":\"creator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskCreated\",\"inputs\":[{\"name\":\"creator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"refundCollector\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"avsFee\",\"type\":\"uint96\",\"indexed\":false,\"internalType\":\"uint96\"},{\"name\":\"taskDeadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"payload\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskVerified\",\"inputs\":[{\"name\":\"aggregator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"taskHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executorOperatorSetId\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"result\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AvsNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"CertificateVerificationFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DuplicateExecutorOperatorSetId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExecutorOperatorSetNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExecutorOperatorSetTaskConfigNotSet\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddressZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAggregatorOperatorSetId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskCreator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskResultSubmitter\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskStatus\",\"inputs\":[{\"name\":\"expected\",\"type\":\"uint8\",\"internalType\":\"enum ITaskMailboxTypes.TaskStatus\"},{\"name\":\"actual\",\"type\":\"uint8\",\"internalType\":\"enum ITaskMailboxTypes.TaskStatus\"}]},{\"type\":\"error\",\"name\":\"PayloadIsEmpty\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TaskSLAIsZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TimestampAtCreation\",\"inputs\":[]}]"
          ]
      }
`

	yamlInvalid = `
---
operator:
  address: "0xoperator..."
  operatorPrivateKey:
    privateKey: "..."
  signingKeys:
    bls:
        keystore: ""
        password: ""
avsPerformers:
   image:
    repository: "eigenlabs/avs"
    tag: "v1.0.0"
  processType: "server"
  avsAddress: "0xavs1..."
`

	jsonValid = `{
  "operator": {
    "address": "0xoperator...",
    "operatorPrivateKey": {
        "privateKey": "..."
    },
    "signingKeys": {
      "bls": {
        "keystore": "",
        "password": ""
      }
    }
  },
  "avsPerformers": [
    {
      "image": {
        "repository": "eigenlabs/avs",
        "tag": "v1.0.0"
      },
      "processType": "server",
      "avsAddress": "0xavs1..."
    }
  ]
}`

	jsonInvalid = `{
  "operator": {
    "address": "0xoperator...",
    "operatorPrivateKey": {
        "privateKey": "..."
    },
    "signingKeys": {
      "bls": {
        "keystore": "",
        "password": ""
      }
    }
  },
  "avsPerformers": {
    "image": {
      "repository": "eigenlabs/avs",
      "tag": "v1.0.0"
    },
    "processType": "server",
    "avsAddress": "0xavs1..."
  }
}`
)

// TestMixedDeploymentModeValidation tests that mixed deployment modes are rejected
func TestMixedDeploymentModeValidation(t *testing.T) {
	t.Run("Should reject mixed deployment modes", func(t *testing.T) {
		config := &ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: "0x123",
				OperatorPrivateKey: &config.ECDSAKeyConfig{
					PrivateKey: "private_key",
				},
				SigningKeys: config.SigningKeys{
					BLS: &config.SigningKey{
						Keystore: "keystore_content",
						Password: "password",
					},
				},
			},
			AvsPerformers: []*AvsPerformerConfig{
				{
					AvsAddress:          "0x456",
					ProcessType:         "server",
					AVSRegistrarAddress: "0x789",
					DeploymentMode:      DeploymentModeDocker, // Docker mode
					Image: &PerformerImage{
						Repository: "test/image",
						Tag:        "v1.0.0",
					},
				},
				{
					AvsAddress:          "0xabc",
					ProcessType:         "server",
					AVSRegistrarAddress: "0xdef",
					DeploymentMode:      DeploymentModeKubernetes, // Kubernetes mode
					Image: &PerformerImage{
						Repository: "test/image2",
						Tag:        "v1.0.0",
					},
				},
			},
			L1Chain: &Chain{
				RpcUrl:  "http://localhost:8545",
				ChainId: 1,
			},
		}

		err := config.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mixed deployment modes not supported")
	})

	t.Run("Should accept all Docker deployment modes", func(t *testing.T) {
		config := &ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: "0x123",
				OperatorPrivateKey: &config.ECDSAKeyConfig{
					PrivateKey: "private_key",
				},
				SigningKeys: config.SigningKeys{
					BLS: &config.SigningKey{
						Keystore: "keystore_content",
						Password: "password",
					},
				},
			},
			AvsPerformers: []*AvsPerformerConfig{
				{
					AvsAddress:          "0x456",
					ProcessType:         "server",
					AVSRegistrarAddress: "0x789",
					DeploymentMode:      DeploymentModeDocker,
					Image: &PerformerImage{
						Repository: "test/image",
						Tag:        "v1.0.0",
					},
				},
				{
					AvsAddress:          "0xabc",
					ProcessType:         "server",
					AVSRegistrarAddress: "0xdef",
					DeploymentMode:      DeploymentModeDocker,
					Image: &PerformerImage{
						Repository: "test/image2",
						Tag:        "v1.0.0",
					},
				},
			},
			L1Chain: &Chain{
				RpcUrl:  "http://localhost:8545",
				ChainId: 1,
			},
		}

		err := config.Validate()
		require.NoError(t, err)
	})

	t.Run("Should accept all Kubernetes deployment modes with proper config", func(t *testing.T) {
		config := &ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: "0x123",
				OperatorPrivateKey: &config.ECDSAKeyConfig{
					PrivateKey: "private_key",
				},
				SigningKeys: config.SigningKeys{
					BLS: &config.SigningKey{
						Keystore: "keystore_content",
						Password: "password",
					},
				},
			},
			AvsPerformers: []*AvsPerformerConfig{
				{
					AvsAddress:          "0x456",
					ProcessType:         "server",
					AVSRegistrarAddress: "0x789",
					DeploymentMode:      DeploymentModeKubernetes,
					Image: &PerformerImage{
						Repository: "test/image",
						Tag:        "v1.0.0",
					},
				},
				{
					AvsAddress:          "0xabc",
					ProcessType:         "server",
					AVSRegistrarAddress: "0xdef",
					DeploymentMode:      DeploymentModeKubernetes,
					Image: &PerformerImage{
						Repository: "test/image2",
						Tag:        "v1.0.0",
					},
				},
			},
			L1Chain: &Chain{
				RpcUrl:  "http://localhost:8545",
				ChainId: 1,
			},
			Kubernetes: NewDefaultKubernetesConfig(),
		}

		err := config.Validate()
		require.NoError(t, err)
	})
}
