package aggregatorConfig

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_AggregatorConfig(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		t.Run("chains only", func(t *testing.T) {
			t.Run("Should create a new aggregator config from a json string", func(t *testing.T) {
				c, err := NewAggregatorConfigFromJsonBytes([]byte(validJsonChainsOnly))
				assert.Nil(t, err)
				assert.NotNil(t, c)
			})
			t.Run("Should fail to create a new aggregator config from an invalid json string", func(t *testing.T) {
				c, err := NewAggregatorConfigFromJsonBytes([]byte(invalidJsonChainsOnly))
				assert.NotNil(t, err)
				assert.Nil(t, c)
			})
		})
		t.Run("chains and avss", func(t *testing.T) {
			t.Run("Should create a new aggregator config from a json string", func(t *testing.T) {
				c, err := NewAggregatorConfigFromJsonBytes([]byte(validJsonChainsAndAvss))
				assert.Nil(t, err)
				assert.NotNil(t, c)
				assert.Equal(t, 1, len(c.Chains))
				assert.Equal(t, 1, len(c.Avss))
				assert.Equal(t, "0x1234avs", c.Avss[0].Address)
				assert.Equal(t, "some private key", c.Avss[0].PrivateKey)
				assert.Equal(t, 3000, c.Avss[0].ResponseTimeout)
				assert.Equal(t, 1, len(c.Avss[0].ChainIds))
			})
			t.Run("Should fail to create a new aggregator config from an invalid json string", func(t *testing.T) {
				c, err := NewAggregatorConfigFromJsonBytes([]byte(invalidJsonChainsAndAvss))
				assert.NotNil(t, err)
				assert.Nil(t, c)
			})
		})
	})
	t.Run("YAML", func(t *testing.T) {
		t.Run("chains only", func(t *testing.T) {
			t.Run("Should create a new aggregator config from a yaml string", func(t *testing.T) {
				c, err := NewAggregatorConfigFromYamlBytes([]byte(validYamlChainsOnly))
				assert.Nil(t, err)
				assert.NotNil(t, c)
			})
			t.Run("Should fail to create a new aggregator config from an invalid yaml string", func(t *testing.T) {
				c, err := NewAggregatorConfigFromYamlBytes([]byte(invalidYamlChainsOnly))
				assert.NotNil(t, err)
				assert.Nil(t, c)
			})
		})
		t.Run("chains and avss", func(t *testing.T) {
			t.Run("Should create a new aggregator config from a yaml string", func(t *testing.T) {
				c, err := NewAggregatorConfigFromYamlBytes([]byte(validYamlChainsAndAvss))
				assert.Nil(t, err)
				assert.NotNil(t, c)
				assert.Equal(t, 1, len(c.Chains))
				assert.Equal(t, 1, len(c.Avss))
				assert.Equal(t, "0x1234avs", c.Avss[0].Address)
				assert.Equal(t, "some private key", c.Avss[0].PrivateKey)
				assert.Equal(t, 3000, c.Avss[0].ResponseTimeout)
				assert.Equal(t, 1, len(c.Avss[0].ChainIds))
			})
			t.Run("Should fail to create a new aggregator config from an invalid yaml string", func(t *testing.T) {
				c, err := NewAggregatorConfigFromYamlBytes([]byte(invalidYamlChainsAndAvss))
				assert.NotNil(t, err)
				assert.Nil(t, c)
			})
		})
	})
}

const (
	validJsonChainsOnly = `
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
	validJsonChainsAndAvss = `
{
	"chains": [
		{
			"name": "ethereum",
			"network": "mainnet",
			"chainId": 1,
			"rpcUrl": "https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID"
		}
	],
	"avss": [
		{
			"address": "0x1234avs",
			"privateKey": "some private key",
			"responseTimeout": 3000,
			"chainIds": [1],
			"signingKeys": {
				"bls": {
					"keystore": "",
					"password": ""
				}
			}
		}
	]
}`
	invalidJsonChainsOnly = `
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
	invalidJsonChainsAndAvss = `
{
	"chains": [
		{
			"name": "ethereum",
			"network": "mainnet",
			"chainId": 1,
			"rpcUrl": "https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID"
		}
	],
	"avss": [
		{
			"address": 1234,
			"privateKey": "some private key",
			"responseTimeout": 3000,
			"chainIds": [1],
			"signingKeys": {
				"bls": {
					"keystore": "",
					"password": ""
				}
			}
		}
	]
}`

	validYamlChainsOnly = `
---
chains:
  - name: ethereum
    network: mainnet
    chainId: 1
    rpcUrl: https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
`
	validYamlChainsAndAvss = `
---
chains:
  - name: ethereum
    network: mainnet
    chainId: 1
    rpcUrl: https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
avss:
  - address: "0x1234avs"
    privateKey: "some private key"
    responseTimeout: 3000
    chainIds: [1]
    signingKeys:
        bls:
            keystore: ""
            password: ""
`
	invalidYamlChainsOnly = `
---
chains:
  - name: ethereum
    network: mainnet
    chainId: True
    rpcUrl: https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
`

	invalidYamlChainsAndAvss = `
---
chains:
  - name: ethereum
    network: mainnet
    chainId: True
    rpcUrl: https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
avss:
  - address: 12345
    privateKey: "some private key"
    responseTimeout: 3000
    chainIds: [1]
    signingKeys:
        bls:
            keystore: ""
            password: ""
`
)
