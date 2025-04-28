package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func Test_TaskRequestPayload(t *testing.T) {
	t.Run("Should unmarshal a correct json", func(t *testing.T) {
		jsonStr := `{ "numberToBeSquared": 4 }`

		var payload TaskRequestPayload
		err := json.Unmarshal([]byte(jsonStr), &payload)
		assert.Nil(t, err)

		i, err := payload.GetBigInt()
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(4), i)
	})

}
