package executorConfig

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ExecutorConfig(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		t.Run("Should parse a valid json config with operator and avss", func(t *testing.T) {
			ec, err := NewExecutorConfigFromYamlBytes([]byte(jsonValid))
			assert.Nil(t, err)
			assert.NotNil(t, ec)
			assert.Equal(t, "0xoperator...", ec.Operator.Address)
			assert.Equal(t, "...", ec.Operator.OperatorPrivateKey)
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
			assert.Equal(t, "...", ec.Operator.OperatorPrivateKey)
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

const (
	yamlValid = `
---
operator:
  address: "0xoperator..."
  operatorPrivateKey: "..."
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
  operatorPrivateKey: "..."
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
    "operatorPrivateKey": "...",
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
    "operatorPrivateKey": "...",
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
