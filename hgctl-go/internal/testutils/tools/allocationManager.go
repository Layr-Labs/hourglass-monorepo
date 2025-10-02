package tools

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	AllocationDelayInfoSlot = 155
)

// InitializeAllocationDelay manipulates the AllocationManager storage to initialize
// the AllocationDelayInfo for a given operator address. This is necessary because
// ModifyAllocations requires the allocation delay to be initialized (isSet = false).
func InitializeAllocationDelay(
	rpcClient *rpc.Client,
	allocationManagerAddr common.Address,
	operatorAddr common.Address,
	currentBlock uint32,
) error {

	slotBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(slotBytes[24:], AllocationDelayInfoSlot)
	keyBytes := common.LeftPadBytes(operatorAddr.Bytes(), 32)

	encoded := append(keyBytes, slotBytes...)
	storageKey := common.BytesToHash(crypto.Keccak256(encoded))

	// AllocationDelayInfo model
	var (
		delay        uint32 = 0            // rightmost 4 bytes
		isSet        byte   = 0x00         // 1 byte before delay (MUST be 0x01 to mark as initialized!)
		pendingDelay uint32 = 0            // 4 bytes before isSet
		effectBlock         = currentBlock // 4 bytes before pendingDelay (leftmost)
	)

	structValue := make([]byte, 32)

	offset := 32
	offset -= 4
	binary.BigEndian.PutUint32(structValue[offset:], delay)

	offset -= 1
	structValue[offset] = isSet

	offset -= 4
	binary.BigEndian.PutUint32(structValue[offset:], pendingDelay)

	offset -= 4
	binary.BigEndian.PutUint32(structValue[offset:], effectBlock)

	var setStorageResult interface{}
	err := rpcClient.Call(&setStorageResult, "anvil_setStorageAt",
		allocationManagerAddr.Hex(),
		storageKey.Hex(),
		"0x"+hex.EncodeToString(structValue))
	if err != nil {
		return fmt.Errorf("failed to set AllocationDelayInfo storage for operator %s: %w", operatorAddr.Hex(), err)
	}

	return nil
}
