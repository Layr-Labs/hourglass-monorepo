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
		t.Run("Should parse a valid json config with operator and avss", func(t *testing.T) {
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
