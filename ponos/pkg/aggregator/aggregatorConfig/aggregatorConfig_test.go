package aggregatorConfig

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	validJson = `
{
	"chains": [
		{
			"name": "ethereum",
			"network": "mainnet",
			"chainId": 1,
			"rpcUrl": "https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID"
		}
	]
}`
	invalidJson = `
{
	"chains": [
		{
			"name": 5679,
			"network": "mainnet",
			"chainId": 1,
			"rpcUrl": "https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID"
		}
	]
}`

	validYaml = `
---
chains:
  - name: ethereum
    network: mainnet
    chainId: 1
    rpcUrl: https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
`
	invalidYaml = `
---
chains:
  - name: ethereum
    network: mainnet
    chainId: True
    rpcUrl: https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
`
)

func Test_AggregatorConfig(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		t.Run("Should create a new aggregator config from a json string", func(t *testing.T) {
			c, err := NewAggregatorConfigFromJsonBytes([]byte(validJson))
			assert.Nil(t, err)
			assert.NotNil(t, c)
		})
		t.Run("Should fail to create a new aggregator config from an invalid json string", func(t *testing.T) {
			c, err := NewAggregatorConfigFromJsonBytes([]byte(invalidJson))
			assert.NotNil(t, err)
			assert.Nil(t, c)
		})
	})
	t.Run("YAML", func(t *testing.T) {
		t.Run("Should create a new aggregator config from a yaml string", func(t *testing.T) {
			c, err := NewAggregatorConfigFromYamlBytes([]byte(validYaml))
			assert.Nil(t, err)
			assert.NotNil(t, c)
		})
		t.Run("Should fail to create a new aggregator config from an invalid yaml string", func(t *testing.T) {
			c, err := NewAggregatorConfigFromYamlBytes([]byte(invalidYaml))
			assert.NotNil(t, err)
			assert.Nil(t, c)
		})
	})

}
