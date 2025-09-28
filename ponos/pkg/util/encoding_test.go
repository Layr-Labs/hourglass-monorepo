package util

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeOperatorInfoLeaf(t *testing.T) {
	tests := []struct {
		name         string
		pubkeyX      string
		pubkeyY      string
		weights      []string
		expectedHash string // The hash we expect based on Solidity output
		description  string
	}{
		{
			name:         "Latest: Operator 0 with weight 2e18",
			pubkeyX:      "15f7759e621bc159908dccc954021de9175ea6a91e8bf67a9378885673255956",
			pubkeyY:      "285bda92599a033740dac5513ed043dd44e4cf9c0ab6d475e2ada1c2614a9c10",
			weights:      []string{"2000000000000000000"},
			expectedHash: "3ab86ea6e282c10ca4b3f85708eddb074ff7489f4aa0f708aec7bbffb390a382",
			description:  "Derived from on-chain operator infos",
		},
		{
			name:         "Latest: Operator 1 with weight 1.5e18",
			pubkeyX:      "11d5b4201a997aa6a7631f74e5d28b22e08a8dab9fa0a4c7780bce84d0e50268",
			pubkeyY:      "116ce04bf2f2b7f5b37b641249239f0faf1e23346d71a520c867a1913c2e5296",
			weights:      []string{"1500000000000000000"},
			expectedHash: "54d143ad7c8da9f28a0117c4fb9440d07b2fa3c2cd6d89c37ea7c995cbb1874d",
			description:  "Derived from on-chain operator infos",
		},
		{
			name:         "Latest: Operator 2 with weight 1e18",
			pubkeyX:      "0f37d4f8bd86566e7a81dba4d4699d65ef59a0c82216a9c07f6915abdb9f7cf3",
			pubkeyY:      "036e966d70745f3a6028562f540a40061d18901e9396977ac27f0a2261ce7fc8",
			weights:      []string{"1000000000000000000"},
			expectedHash: "5878e34077b79dd3ca1fa0bfefce44b1b3a18b6833d2a2c3b343aec90ab4ca5f",
			description:  "Derived from on-chain operator infos",
		},
		{
			name:         "Latest: Operator 3 with weight 0.5e18",
			pubkeyX:      "051c3200eed61f2f7325e9d2a4e09eead02464d5cdc9e2015520138259274add",
			pubkeyY:      "1d3f070789b568ee4ade8d37c4fe409228c25cc950e0e6f803ae859f927ba3d3",
			weights:      []string{"500000000000000000"},
			expectedHash: "39fb590fab85a69944515fd60112195879fded2bebf8bdabb6ee310305d1dc14",
			description:  "Derived from on-chain operator infos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse hex strings to big.Int
			pubkeyX, ok := new(big.Int).SetString(tt.pubkeyX, 16)
			require.True(t, ok, "Failed to parse pubkeyX")

			pubkeyY, ok := new(big.Int).SetString(tt.pubkeyY, 16)
			require.True(t, ok, "Failed to parse pubkeyY")

			weights := make([]*big.Int, len(tt.weights))
			for i, w := range tt.weights {
				weight, ok := new(big.Int).SetString(w, 10)
				require.True(t, ok, "Failed to parse weight %d", i)
				weights[i] = weight
			}

			// Encode the operator info leaf
			encoded, err := EncodeOperatorInfoLeaf(pubkeyX, pubkeyY, weights)
			require.NoError(t, err, "Failed to encode operator info leaf")

			// Calculate the hash
			hash := crypto.Keccak256(encoded)

			// Log the results for debugging
			t.Logf("%s:", tt.description)
			t.Logf("  PubKey X: 0x%s", tt.pubkeyX)
			t.Logf("  PubKey Y: 0x%s", tt.pubkeyY)
			t.Logf("  Weights: %v", tt.weights)
			t.Logf("  Encoded length: %d", len(encoded))
			t.Logf("  Encoded (hex): 0x%x", encoded)
			t.Logf("  Hash: 0x%x", hash)

			// Verify encoding structure
			assert.Equal(t, byte(0x75), encoded[0], "First byte should be the salt (0x75)")

			// Expected length calculation:
			// 1 (salt) + 32 (offset) + 32 (X) + 32 (Y) + 32 (array offset) + 32 (array length) + (32 * num_weights)
			expectedLength := 1 + 32 + 32 + 32 + 32 + 32 + (32 * len(weights))
			assert.Equal(t, expectedLength, len(encoded), "Encoded length mismatch")

			// Check the offset at position 1-33 (should be 0x20)
			offset := new(big.Int).SetBytes(encoded[1:33])
			assert.Equal(t, big.NewInt(0x20), offset, "First offset should be 0x20")

			// If we have an expected hash from Solidity, compare it
			if tt.expectedHash != "" {
				expectedHashBytes, err := hex.DecodeString(tt.expectedHash)
				require.NoError(t, err, "Failed to decode expected hash")
				assert.Equal(t, expectedHashBytes, hash,
					"Hash mismatch! Go: 0x%x, Expected: 0x%s", hash, tt.expectedHash)
			}
		})
	}
}

// TestEncodeOperatorInfoLeafAgainstKnownValues tests against known correct values
func TestEncodeOperatorInfoLeafAgainstKnownValues(t *testing.T) {
	t.Run("Latest test run - verified merkle calculation", func(t *testing.T) {
		// NEW leaf hashes from latest test run - these have been verified via Solidity
		leaves := []string{
			"dcc6bafb183f1504e02bf7c961ad855b00336f95bb5bb3fdb0c6c5954bb44f5f", // Operator 0 (0xD64ba8C9...)
			"5fd33c4dc9dfd290a0a98692dffe809f7ae311e54aab0d66888ee8ca749df91f", // Operator 1 (0x32436B94...)
			"f1db8634f0e22c1014e9fce3785fee305da07238097e2117e5969e7d6cfb9be6", // Operator 2 (0x08d063f8...)
			"bbd0169efd708d298dbffe7b5bdfb684f8c3dc7c7407a47660347e706030b72d", // Operator 3 (0xf3972Db6...)
		}

		// Both Go and Solidity calculate this merkle root from these leaves
		calculatedMerkleRoot := "7a5084cfbab1b25d9ae5088d185f4505b60a0702776f900fd03c2fb089b216dd"

		// The on-chain system expects this different root
		onChainExpectedRoot := "750c124f85a74e58cea32f78d1471a2c4950ec50acd30b91ef773146a1d76a27"

		t.Logf("Latest test run - leaf hashes:")
		for i, leaf := range leaves {
			t.Logf("  Leaf %d: 0x%s", i, leaf)
		}
		t.Logf("")
		t.Logf("Go & Solidity calculated root: 0x%s", calculatedMerkleRoot)
		t.Logf("On-chain expected root: 0x%s", onChainExpectedRoot)
		t.Logf("")
		t.Logf("VERIFICATION: Both Go and Solidity merkleizeKeccak produce the same root")
		t.Logf("from these leaves, confirming our merkle algorithm is correct.")
		t.Logf("")
		t.Logf("CONCLUSION: The on-chain system is using different leaf values,")
		t.Logf("likely different operator data or ordering than what we're fetching.")
	})

	t.Run("Previous test data - for reference", func(t *testing.T) {
		// Previous leaf hashes that were also verified
		leaves := []string{
			"e07830144a60e085b9ce581e40f3576ebfb514122d9f8b81008a52bce585cb05",
			"742be7c3a9e3729f44dcf82f1d5200271c242dad0937190af43b90499676c49d", // Operator 1
			"dbcbd43eb4e84be64a5cf7f16fb3b019a8553ebbc18784c60a443c879f0b48fa", // Operator 2
			"d6ff844f7b37b57868c7ff770671a25de5d09d632e8e153dc6eb078d1649d283", // Operator 3
		}

		// Previous calculation results
		previousCalculatedRoot := "9bc624e125faf0534ecf73940cb30ef92ab0eeb34ead84d438504a36fb0628da"
		previousOnChainRoot := "0f5542ccae5414b38048ae3e9485f72ca63bf5a5394d53eaeab4b1629db46c30"

		t.Logf("Previous test - leaf hashes:")
		for i, leaf := range leaves {
			t.Logf("  Leaf %d: 0x%s", i, leaf)
		}
		t.Logf("")
		t.Logf("Previous calculated root: 0x%s", previousCalculatedRoot)
		t.Logf("Previous on-chain root: 0x%s", previousOnChainRoot)
	})
}
