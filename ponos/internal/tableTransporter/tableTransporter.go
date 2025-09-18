package tableTransporter

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ICrossChainRegistry"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IOperatorTableUpdater"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/multichain-go/pkg/blsSigner"
	"github.com/Layr-Labs/multichain-go/pkg/chainManager"
	"github.com/Layr-Labs/multichain-go/pkg/operatorTableCalculator"
	"github.com/Layr-Labs/multichain-go/pkg/transport"
	"github.com/Layr-Labs/multichain-go/pkg/txSigner"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"
)

type G1Point struct{ X, Y *big.Int }
type BN254OperatorInfo struct {
	Pubkey  G1Point
	Weights []*big.Int
}

type Receipt struct {
	Status          hexutil.Uint64 `json:"status"`
	TransactionHash common.Hash    `json:"transactionHash"`
}

func TransportTable(
	transporterPrivateKey string,
	l1RpcUrl string,
	l1ChainId uint64,
	l2RpcUrl string,
	l2ChainId uint64,
	crossChainRegistryAddress string,
	blsPrivateKey string,
	chainIdsToIgnore []*big.Int,
	l *zap.Logger,
) {
	ctx := context.Background()

	cm := chainManager.NewChainManager()

	l1AnvilConfig := &chainManager.ChainConfig{
		ChainID: l1ChainId,
		RPCUrl:  l1RpcUrl,
	}
	if err := cm.AddChain(l1AnvilConfig); err != nil {
		l.Sugar().Fatalf("Failed to add chain: %v", err)
	}
	holeskyClient, err := cm.GetChainForId(l1AnvilConfig.ChainID)
	if err != nil {
		l.Sugar().Fatalf("Failed to get chain for ID %d: %v", l1AnvilConfig.ChainID, err)
	}

	if l2RpcUrl != "" && l2ChainId != 0 {
		l2ChainConfig := &chainManager.ChainConfig{
			ChainID: l2ChainId,
			RPCUrl:  l2RpcUrl,
		}
		if err := cm.AddChain(l2ChainConfig); err != nil {
			l.Sugar().Fatalf("Failed to add L2 chain: %v", err)
		}
		l.Sugar().Infow("Added L2 chain",
			zap.Any("chainConfig", l2ChainConfig),
		)
	}

	txSign, err := txSigner.NewPrivateKeySigner(transporterPrivateKey)
	if err != nil {
		l.Sugar().Fatalf("Failed to create private key signer: %v", err)
	}

	l.Sugar().Infow("Using CrossChainRegistryAddress",
		zap.String("crossChainRegistryAddress", crossChainRegistryAddress),
	)

	tableCalc, err := operatorTableCalculator.NewStakeTableRootCalculator(&operatorTableCalculator.Config{
		CrossChainRegistryAddress: common.HexToAddress(crossChainRegistryAddress),
	}, holeskyClient.RPCClient, l)
	if err != nil {
		l.Sugar().Fatalf("Failed to create StakeTableRootCalculator: %v", err)
	}

	blockNumber, err := holeskyClient.RPCClient.BlockNumber(ctx)
	if err != nil {
		l.Sugar().Fatalf("Failed to get block number: %v", err)
	}
	// blockNumber = blockNumber - 2
	block, err := holeskyClient.RPCClient.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		l.Sugar().Fatalf("Failed to get block by number: %v", err)
	}

	root, tree, dist, err := tableCalc.CalculateStakeTableRoot(ctx, block.NumberU64())
	if err != nil {
		l.Sugar().Fatalf("Failed to calculate stake table root: %v", err)
	}

	pk, err := bn254.NewPrivateKeyFromHexString(blsPrivateKey)
	if err != nil {
		l.Sugar().Fatalf("Failed to convert BLS private key: %v", err)
	}

	inMemSigner, err := blsSigner.NewInMemoryBLSSigner(pk)
	if err != nil {
		l.Sugar().Fatalf("Failed to create in-memory BLS signer: %v", err)
	}

	stakeTransport, err := transport.NewTransport(
		&transport.TransportConfig{
			L1CrossChainRegistryAddress: common.HexToAddress(crossChainRegistryAddress),
		},
		holeskyClient.RPCClient,
		inMemSigner,
		txSign,
		cm,
		l,
	)
	if err != nil {
		l.Sugar().Fatalf("Failed to create transport: %v", err)
	}

	referenceTimestamp := uint32(block.Time())

	// 1. Transfer ownership of CrossChainRegistry
	transferOwnership(l, l1RpcUrl,
		common.HexToAddress(crossChainRegistryAddress),
		transporterPrivateKey)

	// 2. Get supported chains and configure each one
	ccRegistryCaller, _ := ICrossChainRegistry.NewICrossChainRegistryCaller(
		common.HexToAddress(crossChainRegistryAddress),
		holeskyClient.RPCClient)

	chainIds, addresses, _ := ccRegistryCaller.GetSupportedChains(&bind.CallOpts{})

	for i, chainId := range chainIds {
		addr := addresses[i]
		if chainId.Uint64() != l1ChainId && chainId.Uint64() != l2ChainId {
			continue
		}

		// Determine the correct RPC URL for this chain
		rpcURL := l1RpcUrl
		if chainId.Uint64() == l2ChainId {
			rpcURL = l2RpcUrl
		}

		// Transfer ownership of OperatorTableUpdater
		transferOwnership(l, rpcURL, addr, transporterPrivateKey)

		// Read current generator
		gen, err := getGenerator(ctx, l, cm, chainId, addr)
		if err != nil {
			l.Sugar().Fatalf("Failed to get generator: %v", err)
		}

		// Switch to new operatorSet ID
		if gen.Id == 1 {
			gen.Id = 2
		} else {
			gen.Id = 1
		}

		// Connect to an ethClient to construct contractCaller
		client, err := ethclient.Dial(l1RpcUrl)
		if err != nil {
			l.Error("failed to connect to L1 RPC: %w", zap.Error(err))
		}

		signer, err := transactionSigner.NewPrivateKeySigner(transporterPrivateKey, client, l)
		if err != nil {
			log.Error("Failed to create private key signer")
		}

		// Construct contractCaller with KeyRegistrar
		contractCaller, err := caller.NewContractCaller(client, signer, l)
		if err != nil {
			l.Error("Failed to create contract caller")
		}

		// Derive BN254 keys from the hex string (no keystore files needed) ---
		blsHex := strings.TrimPrefix(blsPrivateKey, "0x")

		// Extract key details
		scheme := bn254.NewScheme()
		skGeneric, err := scheme.NewPrivateKeyFromHexString(blsHex)
		if err != nil {
			l.Error("parse BLS hex: %w", zap.Error(err))
		}
		blsPriv, err := bn254.NewPrivateKeyFromBytes(skGeneric.Bytes())
		if err != nil {
			l.Error("convert BLS key: %w", zap.Error(err))
		}
		blsPub := blsPriv.Public() // <- this is what EncodeBN254KeyData expects

		// Encode keyData for KeyRegistrar from the PUBLIC key ---
		keyData, err := contractCaller.EncodeBN254KeyData(blsPub)
		if err != nil {
			l.Error("encode key data: %w", zap.Error(err))
		}

		// Configure curve type (you need to add CURVE_TYPE_KEY_REGISTRAR_BN254 constant)
		const CURVE_TYPE_KEY_REGISTRAR_BN254 = 2 // or whatever the correct value is

		// You need the KeyRegistrar address - this should come from your config
		keyRegistrarAddress := common.HexToAddress("0xA4dB30D08d8bbcA00D40600bee9F029984dB162a")

		if err := configureCurveTypeAsAVS(
			ctx,
			l,
			l1RpcUrl, // KeyRegistrar is on L1
			keyRegistrarAddress,
			gen.Avs,
			gen.Id,
			CURVE_TYPE_KEY_REGISTRAR_BN254,
		); err != nil {
			l.Sugar().Fatalf("Failed to configure curve type: %v", err)
		}

		// Now you need to register the BLS key in the KeyRegistrar
		// This requires creating a contractCaller and registering the key
		// (This is the part you're completely missing)
		opEOA := mustKey(l, transporterPrivateKey)
		operatorAddress := crypto.PubkeyToAddress(opEOA.PublicKey)

		// Build the message hash per registrar rules and sign with BLS private key
		msgHash, err := contractCaller.GetOperatorRegistrationMessageHash(
			ctx,
			operatorAddress,
			gen.Avs,
			gen.Id,
			keyData,
		)
		if err != nil {
			l.Error("failed to get operator registration message", zap.Error(err))
		}
		sig, err := blsPriv.SignSolidityCompatible(msgHash)
		if err != nil {
			log.Error("BLS sign: %w", err)
		}

		// Register in KeyRegistrar
		if _, err := contractCaller.RegisterKeyWithKeyRegistrar(
			ctx,
			operatorAddress,
			gen.Avs,
			gen.Id,
			keyData,
			sig.Bytes(),
		); err != nil {
			log.Error("register key in key registrar: %w", err)
		}

		// Finally, update the generator with your BLS key
		if err := updateGeneratorFromContext(ctx, l, cm, chainId, addr, txSign, blsPrivateKey, gen); err != nil {
			l.Sugar().Fatalf("Failed to update generator: %v", err)
		}
	}

	err = stakeTransport.SignAndTransportGlobalTableRoot(
		ctx,
		root,
		referenceTimestamp,
		block.NumberU64(),
		chainIdsToIgnore,
	)
	if err != nil {
		l.Sugar().Fatalf("Failed to sign and transport global table root: %v", err)
	}

	opsets := dist.GetOperatorSets()
	if len(opsets) == 0 {
		l.Sugar().Infow("No operator sets found, skipping AVS stake table transport")
		return
	}
	fmt.Printf("Operatorsets to transport: %+v\n", opsets)
	for _, opset := range opsets {
		err = stakeTransport.SignAndTransportAvsStakeTable(
			ctx,
			referenceTimestamp,
			block.NumberU64(),
			opset,
			root,
			tree,
			dist,
			chainIdsToIgnore,
		)
		if err != nil {
			l.Sugar().Fatalf("Failed to sign and transport AVS stake table for opset %v: %v", opset, err)
		} else {
			l.Sugar().Infof("Successfully signed and transported AVS stake table for opset %v", opset)
		}
	}
}

// Impersonate the current owner and call *.transferOwnership(newOwner).
func transferOwnership(logger *zap.Logger, rpcURL string, proxy common.Address, privateKey string) {
	ctx := context.Background()
	c, err := rpc.DialContext(ctx, rpcURL)
	if err != nil {
		logger.Error("failed to connect to rpc", zap.Error(err))
	}

	// Your private key - used only to derive the new owner address
	priv := mustKey(logger, privateKey)
	newOwner := crypto.PubkeyToAddress(priv.PublicKey)

	// ABI with owner() and transferOwnership(address)
	ownableABI := mustABI(logger, `[
	  {"inputs":[],"name":"owner","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
	  {"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"}
	]`)

	// read current owner
	currOwner := readOwner(ctx, logger, c, ownableABI, proxy)
	logger.Info("Current owner: %s", zap.String("owner", currOwner.Hex()))

	// impersonate the current owner and fund it
	impersonate(ctx, logger, c, currOwner)
	defer stopImpersonate(ctx, c, currOwner)

	// pack transferOwnership(newOwner)
	calldata, err := ownableABI.Pack("transferOwnership", newOwner)
	if err != nil {
		logger.Error("failed to pack callData %w", zap.Error(err))
	}

	// send tx via eth_sendTransaction from the impersonated owner to the proxy
	tx := map[string]any{
		"from":  currOwner.Hex(),
		"to":    proxy.Hex(),
		"data":  hexutil.Encode(calldata),
		"value": "0x0",
	}
	var txHash common.Hash
	if err := c.CallContext(ctx, &txHash, "eth_sendTransaction", tx); err != nil {
		logger.Error("failed to send tx: %w", zap.Error(err))
	}

	// await for tx receipt
	mustWaitReceipt(ctx, logger, c, txHash)
	logger.Info("TransferOwnership tx: %s", zap.String("owner", txHash.Hex()))

	// verify
	newOwnerRead := readOwner(ctx, logger, c, ownableABI, proxy)
	logger.Info("New owner: %s", zap.String("owner", newOwnerRead.Hex()))
}

// Impersonate the AVS and call KeyRegistrar.configureOperatorSet(opSet, curveType)
func configureCurveTypeAsAVS(
	ctx context.Context,
	logger *zap.Logger,
	rpcURL string,
	keyRegistrar common.Address,
	avs common.Address,
	opSetId uint32,
	curveType uint8,
) error {
	// Connect to provided RPC
	c, err := rpc.DialContext(ctx, rpcURL)
	if err != nil {
		return fmt.Errorf("rpc dial: %w", err)
	}

	// Build minimal ABI
	krABI := mustABI(logger, `[
      {"inputs":[{"components":[{"internalType":"address","name":"avs","type":"address"},{"internalType":"uint32","name":"id","type":"uint32"}],"internalType":"struct OperatorSet","name":"opSet","type":"tuple"}],"name":"getOperatorSetCurveType","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},
      {"inputs":[{"components":[{"internalType":"address","name":"avs","type":"address"},{"internalType":"uint32","name":"id","type":"uint32"}],"internalType":"struct OperatorSet","name":"opSet","type":"tuple"},{"internalType":"uint8","name":"curveType","type":"uint8"}],"name":"configureOperatorSet","outputs":[],"stateMutability":"nonpayable","type":"function"}
    ]`)

	// Tuple type to match (address avs, uint32 id)
	type opSetT struct {
		Avs common.Address
		Id  uint32
	}
	opSet := opSetT{Avs: avs, Id: opSetId}

	// Read current curve type; skip if already set
	calldataGet, _ := krABI.Pack("getOperatorSetCurveType", opSet)
	var out string
	if err := c.CallContext(ctx, &out, "eth_call",
		map[string]any{"to": keyRegistrar.Hex(), "data": hexutil.Encode(calldataGet)},
		"latest",
	); err != nil {
		return fmt.Errorf("getOperatorSetCurveType call: %w", err)
	}
	decoded, err := krABI.Unpack("getOperatorSetCurveType", common.FromHex(out))
	if err != nil {
		return fmt.Errorf("unpack getOperatorSetCurveType: %w", err)
	}
	if ct, ok := decoded[0].(uint8); ok && ct == curveType {
		logger.Info("Operator set already configured with curveType, skipping")
		return nil
	}

	// Impersonate AVS & fund it
	var ok bool
	if err := c.CallContext(ctx, &ok, "anvil_impersonateAccount", avs.Hex()); err != nil {
		return fmt.Errorf("impersonate avs: %w", err)
	}
	defer func() { _ = c.CallContext(ctx, &ok, "anvil_stopImpersonatingAccount", avs.Hex()) }()
	_ = c.CallContext(ctx, &ok, "anvil_setBalance", avs.Hex(), "0x56BC75E2D63100000") // 100 ETH

	// Send configureOperatorSet from the AVS
	calldataCfg, err := krABI.Pack("configureOperatorSet", opSet, curveType)
	if err != nil {
		return fmt.Errorf("pack configureOperatorSet: %w", err)
	}

	// Construct tx to send from the AVS
	tx := map[string]any{
		"from":  avs.Hex(),
		"to":    keyRegistrar.Hex(),
		"data":  hexutil.Encode(calldataCfg),
		"value": "0x0",
	}
	var txHash common.Hash
	if err := c.CallContext(ctx, &txHash, "eth_sendTransaction", tx); err != nil {
		return fmt.Errorf("send configureOperatorSet: %w", err)
	}

	// Await receipt
	mustWaitReceipt(ctx, logger, c, txHash)
	logger.Info("ConfigureOperatorSet tx sent by AVS: %s", zap.String("owner", txHash.Hex()))
	return nil
}

// call calculateOperatorInfoLeaf via a bound contract
func calcOperatorInfoLeaf(
	ctx context.Context,
	logger *zap.Logger,
	backend bind.ContractCaller,
	addr common.Address,
	info BN254OperatorInfo,
) ([32]byte, error) {
	abiCalc := mustABI(logger, `[
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

	// Minimal ABI: getGenerator() -> (address avs, uint32 id)
	abiGet := mustABI(logger, `[
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

// Update the generator onchain to verify against context provided BLS key
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
	root, err := calcOperatorInfoLeaf(ctx, logger, chain.RPCClient, common.HexToAddress("0xff58A373c18268F483C1F5cA03Cf885c0C43373a"), info)
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

func readOwner(ctx context.Context, logger *zap.Logger, c *rpc.Client, ab abi.ABI, proxy common.Address) common.Address {
	data, _ := ab.Pack("owner")
	call := map[string]any{"to": proxy.Hex(), "data": hexutil.Encode(data)}
	var out string
	if err := c.CallContext(ctx, &out, "eth_call", call, "latest"); err != nil {
		logger.Error("failed to call contract: %w", zap.Error(err))
	}
	b := common.FromHex(out)
	return common.BytesToAddress(b[len(b)-20:])
}

func impersonate(ctx context.Context, logger *zap.Logger, c *rpc.Client, who common.Address) {
	var ok bool
	if err := c.CallContext(ctx, &ok, "anvil_impersonateAccount", who.Hex()); err != nil {
		logger.Error("failed to impersonate: %w", zap.Error(err))
	}
	// fund so it can pay gas
	_ = c.CallContext(ctx, &ok, "anvil_setBalance", who.Hex(), "0x56BC75E2D63100000") // 100 ETH
}

func stopImpersonate(ctx context.Context, c *rpc.Client, who common.Address) {
	var ok bool
	_ = c.CallContext(ctx, &ok, "anvil_stopImpersonatingAccount", who.Hex())
}

func mustABI(logger *zap.Logger, s string) abi.ABI {
	a, err := abi.JSON(strings.NewReader(s))
	if err != nil {
		logger.Error("invalid abi: %w", zap.Error(err))
	}
	return a
}

func mustKey(logger *zap.Logger, hex string) *ecdsa.PrivateKey {
	if strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X") {
		hex = hex[2:]
	}
	k, err := crypto.HexToECDSA(hex)
	if err != nil {
		logger.Error("invalid key: %w", zap.Error(err))
	}
	return k
}

func mustWaitReceipt(ctx context.Context, logger *zap.Logger, c *rpc.Client, h common.Hash) {
	var r Receipt
	for {
		_ = c.CallContext(ctx, &r, "eth_getTransactionReceipt", h)
		if r.TransactionHash != (common.Hash{}) {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if r.Status != 1 {
		// Get reason
		var trace map[string]any
		_ = c.CallContext(ctx, &trace, "debug_traceTransaction", h.Hex(), map[string]any{"disableStack": true})
		logger.Error("tx reverted. trace")
	}
}
