package transport

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IOperatorTableUpdater"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/tools"
	"github.com/Layr-Labs/multichain-go/pkg/blsSigner"
	"github.com/Layr-Labs/multichain-go/pkg/chainManager"
	"github.com/Layr-Labs/multichain-go/pkg/txSigner"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type G1Point struct{ X, Y *big.Int }

type BN254OperatorInfo struct {
	Pubkey  G1Point
	Weights []*big.Int
}

// Read OperatorTableUpdater.getGenerator() as a typed struct.
func getGenerator(
	ctx context.Context,
	logger *zap.Logger,
	cm chainManager.IChainManager,
	chainId *big.Int,
	updaterAddr common.Address,
) (IOperatorTableUpdater.OperatorSet, error) {
	chain, err := cm.GetChainForId(chainId.Uint64())
	if err != nil {
		return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("get chain %d: %w", chainId.Uint64(), err)
	}

	abiGet := tools.MustABI(logger, `[
      {"inputs":[],"name":"getGenerator","outputs":[{"components":[{"internalType":"address","name":"avs","type":"address"},{"internalType":"uint32","name":"id","type":"uint32"}],"internalType":"struct OperatorSet","name":"","type":"tuple"}],"stateMutability":"view","type":"function"}
    ]`)

	c := bind.NewBoundContract(updaterAddr, abiGet, chain.RPCClient, nil, nil)

	var outs []any
	if err := c.Call(&bind.CallOpts{Context: ctx}, &outs, "getGenerator"); err != nil {
		return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("getGenerator call: %w", err)
	}
	if len(outs) != 1 {
		return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("unexpected outputs len: %d", len(outs))
	}

	v := reflect.ValueOf(outs[0])
	switch v.Kind() {
	case reflect.Struct:
		fAvs := v.FieldByName("Avs")
		fId := v.FieldByName("Id")
		if !fAvs.IsValid() || !fId.IsValid() {
			return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("tuple missing fields Avs/Id")
		}
		avs, ok := fAvs.Interface().(common.Address)
		if !ok {
			return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("unexpected Avs type: %T", fAvs.Interface())
		}
		var id uint32
		switch fId.Kind() {
		case reflect.Uint32:
			id = uint32(fId.Uint())
		case reflect.Uint64:
			id = uint32(fId.Uint())
		case reflect.Uint:
			id = uint32(fId.Uint())
		case reflect.Int32:
			id = uint32(fId.Int())
		case reflect.Int64:
			id = uint32(fId.Int())
		case reflect.Int:
			id = uint32(fId.Int())
		default:
			return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("unexpected Id kind: %s", fId.Kind())
		}
		return IOperatorTableUpdater.OperatorSet{Avs: avs, Id: id}, nil

	case reflect.Slice:
		// Some decoders may return the tuple as a 2-element slice
		if v.Len() != 2 {
			return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("tuple slice wrong length: %d", v.Len())
		}
		avs, ok := v.Index(0).Interface().(common.Address)
		if !ok {
			return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("unexpected Avs type: %T", v.Index(0).Interface())
		}
		// Coerce Id safely
		idVal := v.Index(1)
		var id uint32
		switch idVal.Kind() {
		case reflect.Uint32:
			id = uint32(idVal.Uint())
		case reflect.Uint64:
			id = uint32(idVal.Uint())
		case reflect.Uint:
			id = uint32(idVal.Uint())
		case reflect.Int32:
			id = uint32(idVal.Int())
		case reflect.Int64:
			id = uint32(idVal.Int())
		case reflect.Int:
			id = uint32(idVal.Int())
		default:
			return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("unexpected Id kind in slice: %s", idVal.Kind())
		}
		return IOperatorTableUpdater.OperatorSet{Avs: avs, Id: id}, nil
	default:
		return IOperatorTableUpdater.OperatorSet{}, fmt.Errorf("unexpected tuple type: %T", outs[0])
	}
}

func updateGeneratorFromContext(
	ctx context.Context,
	logger *zap.Logger,
	cm chainManager.IChainManager,
	chainId *big.Int,
	updaterAddr common.Address,
	txSign txSigner.ITransactionSigner,
	blsHex string,
	gen IOperatorTableUpdater.OperatorSet,
) error {
	chain, err := cm.GetChainForId(chainId.Uint64())
	if err != nil {
		return fmt.Errorf("get chain %d: %w", chainId.Uint64(), err)
	}

	updaterTx, err := IOperatorTableUpdater.NewIOperatorTableUpdater(updaterAddr, chain.RPCClient)
	if err != nil {
		return fmt.Errorf("bind updater tx: %w", err)
	}

	// Derive BLS pubkey
	scheme := bn254.NewScheme()
	skGeneric, err := scheme.NewPrivateKeyFromHexString(strings.TrimPrefix(blsHex, "0x"))
	if err != nil {
		return fmt.Errorf("parse bls: %w", err)
	}
	sk, err := bn254.NewPrivateKeyFromBytes(skGeneric.Bytes())
	if err != nil {
		return fmt.Errorf("bls convert: %w", err)
	}
	signer, err := blsSigner.NewInMemoryBLSSigner(sk)
	if err != nil {
		return fmt.Errorf("bls signer: %w", err)
	}
	pub, err := signer.GetPublicKey()
	if err != nil {
		return fmt.Errorf("pubkey: %w", err)
	}
	g1 := bn254.NewZeroG1Point().AddPublicKey(pub)
	g1b, err := g1.ToPrecompileFormat()
	if err != nil {
		return fmt.Errorf("g1 bytes: %w", err)
	}
	pkG1 := G1Point{
		X: new(big.Int).SetBytes(g1b[0:32]),
		Y: new(big.Int).SetBytes(g1b[32:64]),
	}

	// One-operator info
	info := BN254OperatorInfo{Pubkey: pkG1, Weights: []*big.Int{big.NewInt(1)}}

	// Calculate the root leaf
	root, err := calcOperatorInfoLeaf(ctx, logger, chain.RPCClient, common.HexToAddress("0xff58a373c18268f483c1f5ca03cf885c0c43373a"), info)
	if err != nil {
		return fmt.Errorf("calc operatorInfo leaf: %w", err)
	}

	genInfo := IOperatorTableUpdater.IOperatorTableCalculatorTypesBN254OperatorSetInfo{
		OperatorInfoTreeRoot: root,
		NumOperators:         new(big.Int).SetUint64(1),
		AggregatePubkey:      IOperatorTableUpdater.BN254G1Point{X: pkG1.X, Y: pkG1.Y},
		TotalWeights:         []*big.Int{big.NewInt(1)},
	}

	auth, err := txSign.GetTransactOpts(ctx, chainId)
	if err != nil {
		return fmt.Errorf("opts: %w", err)
	}

	tx, err := updaterTx.UpdateGenerator(auth, gen, genInfo)
	if err != nil {
		return fmt.Errorf("updateGenerator tx: %w", err)
	}

	rcpt, err := bind.WaitMined(ctx, chain.RPCClient, tx)
	if err != nil {
		return fmt.Errorf("wait mined: %w", err)
	}
	if rcpt.Status != 1 {
		return fmt.Errorf("updateGenerator reverted: %s", tx.Hash().Hex())
	}
	return nil
}

func calcOperatorInfoLeaf(
	ctx context.Context,
	logger *zap.Logger,
	backend bind.ContractCaller,
	addr common.Address,
	info BN254OperatorInfo,
) ([32]byte, error) {
	abiCalc := tools.MustABI(logger, `[
		{
			"inputs": [
				{
					"components": [
						{
							"components": [
								{
									"internalType": "uint256",
									"name": "X",
									"type": "uint256"
								},
								{
									"internalType": "uint256",
									"name": "Y",
									"type": "uint256"
								}
							],
							"internalType": "struct BN254.G1Point",
							"name": "pubkey",
							"type": "tuple"
						},
						{
							"internalType": "uint256[]",
							"name": "weights",
							"type": "uint256[]"
						}
					],
					"internalType": "struct IOperatorTableCalculatorTypes.BN254OperatorInfo",
					"name": "operatorInfo",
					"type": "tuple"
				}
			],
			"name": "calculateOperatorInfoLeaf",
			"outputs": [
				{
					"internalType": "bytes32",
					"name": "",
					"type": "bytes32"
				}
			],
			"stateMutability": "pure",
			"type": "function"
		}
	]`)

	c := bind.NewBoundContract(addr, abiCalc, backend, nil, nil)

	var outs []any
	if err := c.Call(&bind.CallOpts{Context: ctx}, &outs, "calculateOperatorInfoLeaf", info); err != nil {
		return [32]byte{}, fmt.Errorf("calcOperatorInfoLeaf call: %w", err)
	}
	if len(outs) != 1 {
		return [32]byte{}, fmt.Errorf("unexpected outputs len: %d", len(outs))
	}

	var leaf [32]byte
	switch v := outs[0].(type) {
	case [32]uint8:
		for i := 0; i < 32; i++ {
			leaf[i] = byte(v[i])
		}
	case []byte:
		if len(v) != 32 {
			return [32]byte{}, fmt.Errorf("bytes32 wrong length: %d", len(v))
		}
		copy(leaf[:], v)
	case common.Hash:
		leaf = v
	default:
		return [32]byte{}, fmt.Errorf("unexpected return type %T", v)
	}

	return leaf, nil
}
