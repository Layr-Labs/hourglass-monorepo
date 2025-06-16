package mailbox

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IAllocationManager"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IBN254TableCalculator"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ICrossChainRegistry"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IKeyRegistrar"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	chainPoller2 "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller/EVMChainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore/inMemoryContractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/eigenlayer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"math/big"
	"os"
	"testing"
	"time"
)

func Test_L1Mailbox(t *testing.T) {
	const (
		L1RpcUrl = "http://127.0.0.1:8545"
	)

	// t.Skip("Flaky, skipping for now")
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	root := testUtils.GetProjectRootPath()
	t.Logf("Project root path: %s", root)

	aggPrivateKey, _, err := bn254.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	execPrivateKey, execPublicKey, err := bn254.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	chainConfig, err := testUtils.ReadChainConfig(root)
	if err != nil {
		t.Fatalf("Failed to read chain config: %v", err)
	}

	coreContracts, err := eigenlayer.LoadContracts()
	if err != nil {
		t.Fatalf("Failed to load core contracts: %v", err)
	}

	imContractStore := inMemoryContractStore.NewInMemoryContractStore(coreContracts, l)

	if err = testUtils.ReplaceMailboxAddressWithTestAddress(imContractStore, chainConfig); err != nil {
		t.Fatalf("Failed to replace mailbox address with test address: %v", err)
	}

	tlp := transactionLogParser.NewTransactionLogParser(imContractStore, l)

	l1EthereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   L1RpcUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	logsChan := make(chan *chainPoller2.LogWithBlock)

	l1Poller := EVMChainPoller.NewEVMChainPoller(l1EthereumClient, logsChan, tlp, &EVMChainPoller.EVMChainPollerConfig{
		ChainId:              config.ChainId_EthereumAnvil,
		PollingInterval:      time.Duration(10) * time.Second,
		InterestingContracts: imContractStore.ListContractAddressesForChain(config.ChainId_EthereumAnvil),
	}, l)

	l1EthClient, err := l1EthereumClient.GetEthereumContractCaller()
	if err != nil {
		t.Fatalf("Failed to get Ethereum contract caller: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l1Anvil, err := testUtils.StartL1Anvil(root, ctx)
	if err != nil {
		t.Fatalf("Failed to start L1 Anvil: %v", err)
	}

	if os.Getenv("CI") == "" {
		fmt.Printf("Sleeping for 10 seconds\n\n")
		time.Sleep(10 * time.Second)
	} else {
		fmt.Printf("Sleeping for 30 seconds\n\n")
		time.Sleep(30 * time.Second)
	}
	fmt.Println("Checking if l1Anvil is up and running...")

	l1ChainId, err := l1EthClient.ChainID(ctx)
	if err != nil {
		t.Fatalf("Failed to get L1 chain ID: %v", err)
	}
	t.Logf("L1 Chain ID: %s", l1ChainId.String())

	eigenlayerContractAddrs, err := config.GetCoreContractsForChainId(config.ChainId(l1ChainId.Uint64()))
	if err != nil {
		t.Fatalf("Failed to get core contracts for chain ID: %v", err)
	}

	l1CC, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:                chainConfig.AppAccountPrivateKey,
		AVSRegistrarAddress:       chainConfig.AVSTaskRegistrarAddress, // technically not used...
		TaskMailboxAddress:        chainConfig.MailboxContractAddressL2,
		CrossChainRegistryAddress: eigenlayerContractAddrs.CrossChainRegistry,
		KeyRegistrarAddress:       eigenlayerContractAddrs.KeyRegistrar,
	}, l1EthClient, l)
	if err != nil {
		t.Fatalf("Failed to create L2 contract caller: %v", err)
	}

	reservations, err := l1CC.GetActiveGenerationReservations()
	if err != nil {
		t.Fatalf("Failed to get active generation reservations: %v", err)
	}
	for _, reservation := range reservations {
		fmt.Printf("Active generation reservation: %+v\n", reservation)
	}

	kr, err := IKeyRegistrar.NewIKeyRegistrar(common.HexToAddress(eigenlayerContractAddrs.KeyRegistrar), l1EthClient)
	if err != nil {
		t.Fatalf("Failed to create key registrar: %v", err)
	}

	l.Sugar().Infow("Setting up operator peering",
		zap.String("AVSAccountAddress", chainConfig.AVSAccountAddress),
	)
	err = testUtils.SetupOperatorPeering(
		ctx,
		chainConfig,
		config.ChainId(l1ChainId.Uint64()),
		l1EthClient,
		aggPrivateKey,
		execPrivateKey,
		"localhost:9000",
		l,
	)
	if err != nil {
		t.Fatalf("Failed to set up operator peering: %v", err)
	}

	am, err := IAllocationManager.NewIAllocationManager(common.HexToAddress("0xfdd5749e11977d60850e06bf5b13221ad95eb6b4"), l1EthClient)
	if err != nil {
		t.Fatalf("Failed to create allocation manager: %v", err)
	}
	ccr, err := ICrossChainRegistry.NewICrossChainRegistry(common.HexToAddress(eigenlayerContractAddrs.CrossChainRegistry), l1EthClient)
	if err != nil {
		t.Fatalf("Failed to create cross chain registry: %v", err)
	}

	currentBlock, err := l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}

	for _, opsetId := range []uint32{0, 1} {
		strategies, err := am.GetStrategiesInOperatorSet(&bind.CallOpts{}, IAllocationManager.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get strategies in operator set %d: %v", opsetId, err)
		}
		fmt.Printf("Strategies in operator set %d: %+v\n", opsetId, strategies)

		members, err := am.GetMembers(&bind.CallOpts{}, IAllocationManager.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get members in operator set %d: %v", opsetId, err)
		}
		fmt.Printf("Members in operator set %d: %+v\n", opsetId, members)

		minSlashableStake, err := am.GetMinimumSlashableStake(
			&bind.CallOpts{},
			IAllocationManager.OperatorSet{
				Id:  opsetId,
				Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
			},
			members,
			strategies,
			uint32(currentBlock+100),
		)
		if err != nil {
			t.Fatalf("Failed to get minimum slashable stake for operator set %d: %v", opsetId, err)
		}
		fmt.Printf("minimum slashable stake for operator set %d: %+v\n", opsetId, minSlashableStake)

		tableCalcAddr, err := ccr.GetOperatorTableCalculator(&bind.CallOpts{}, ICrossChainRegistry.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get operator table calculator for operator set %d: %v", opsetId, err)
		}
		fmt.Printf("Operator table calculator for operator set %d: %s\n", opsetId, tableCalcAddr.String())

		cfg, err := ccr.GetOperatorSetConfig(&bind.CallOpts{}, ICrossChainRegistry.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get operator set config for operator set %d: %v", opsetId, err)
		}
		fmt.Printf("Operator set config for operator set %d: %+v\n", opsetId, cfg)

		curve, err := kr.GetOperatorSetCurveType(&bind.CallOpts{}, IKeyRegistrar.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get operator set curve type: %v", err)
		}
		fmt.Printf("Operator set curve type: %v\n", curve)

		tableCalc, err := IBN254TableCalculator.NewIBN254TableCalculator(tableCalcAddr, l1EthClient)
		if err != nil {
			t.Fatalf("Failed to create operator table calculator for operator set %d: %v", opsetId, err)
		}

		weights, err := tableCalc.GetOperatorWeights(&bind.CallOpts{}, IBN254TableCalculator.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get operator weights for operator set %d: %v", opsetId, err)
		}
		fmt.Printf("Operator weights for operator set %d: %+v\n", opsetId, weights)

		tableBytes, err := tableCalc.CalculateOperatorTableBytes(&bind.CallOpts{}, IBN254TableCalculator.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to calculate operator table bytes for operator set %d: %v", opsetId, err)
		}
		fmt.Printf("Operator table bytes for operator set %d: %x\n", opsetId, tableBytes)

	}

	l.Sugar().Infow("------------------------ Transporting L1 tables ------------------------")
	// transport the tables for good measure
	testUtils.TransportL1Tables(l)
	l.Sugar().Infow("Sleeping for 6 seconds to allow table transport to complete")
	time.Sleep(time.Second * 6)

	// update current block to account for transport
	currentBlock, err = l1EthClient.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block number: %v", err)
	}

	if err := l1Poller.Start(ctx); err != nil {
		cancel()
		t.Fatalf("Failed to start EVM L1Chain Poller: %v", err)
	}

	tableData, err := l1CC.GetOperatorTableDataForOperatorSet(
		ctx,
		common.HexToAddress(chainConfig.AVSAccountAddress),
		1,
		config.ChainId(l1ChainId.Uint64()),
		currentBlock,
	)
	if err != nil {
		t.Fatalf("Failed to get operator table data: %v", err)
	}
	fmt.Printf("Operator table data: %+v\n", tableData)

	// TODO(seanmcgary): need to actually set up operators and their keys in order to pass the certificate verifier

	hasErrors := false
	go func() {
		for logWithBlock := range logsChan {
			fmt.Printf("Received logWithBlock: %+v\n", logWithBlock.Log)
			if logWithBlock.Log.EventName != "TaskCreated" {
				continue
			}
			t.Logf("Found created task log: %+v", logWithBlock.Log)
			assert.Equal(t, "TaskCreated", logWithBlock.Log.EventName)

			task, err := types.NewTaskFromLog(logWithBlock.Log, logWithBlock.Block, chainConfig.MailboxContractAddressL1)
			assert.Nil(t, err)

			assert.Equal(t, common.HexToAddress(chainConfig.AVSAccountAddress), common.HexToAddress(task.AVSAddress))
			assert.True(t, len(task.TaskId) > 0)
			assert.True(t, len(task.Payload) > 0)

			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to create task session: %v", err)
				cancel()
				return
			}

			operators := []*aggregation.Operator{
				{
					Address:   chainConfig.ExecOperatorAccountAddress,
					PublicKey: execPublicKey,
				},
			}

			resultAgg, err := aggregation.NewTaskResultAggregator(
				ctx,
				task.TaskId,
				task.BlockNumber,
				task.OperatorSetId,
				100,
				task.Payload,
				task.DeadlineUnixSeconds,
				operators,
			)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to create task result aggregator: %v", err)
				cancel()
				return
			}

			outputResult := util.BigIntToHex(new(big.Int).SetUint64(16))
			signer := inMemorySigner.NewInMemorySigner(execPrivateKey)
			digest := util.GetKeccak256Digest(outputResult)

			sig, err := signer.SignMessageForSolidity(digest)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to sign message: %v", err)
				cancel()
				return
			}

			taskResult := &types.TaskResult{
				TaskId:          task.TaskId,
				AvsAddress:      chainConfig.AVSAccountAddress,
				CallbackAddr:    chainConfig.AVSAccountAddress,
				OperatorSetId:   1,
				Output:          outputResult,
				ChainId:         config.ChainId(l1ChainId.Uint64()),
				BlockNumber:     task.BlockNumber,
				BlockHash:       task.BlockHash,
				OperatorAddress: chainConfig.ExecOperatorAccountAddress,
				Signature:       sig,
			}
			err = resultAgg.ProcessNewSignature(ctx, task.TaskId, taskResult)
			assert.Nil(t, err)

			assert.True(t, resultAgg.SigningThresholdMet())

			cert, err := resultAgg.GenerateFinalCertificate()
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to generate final certificate: %v", err)
				cancel()
				return
			}
			signedAt := time.Unix(int64(logWithBlock.Block.Timestamp.Value()), 0).Add(10 * time.Second)
			cert.SignedAt = &signedAt
			fmt.Printf("cert: %+v\n", cert)

			time.Sleep(10 * time.Second)

			avsCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
				PrivateKey:          chainConfig.AVSAccountPrivateKey,
				AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
				TaskMailboxAddress:  chainConfig.MailboxContractAddressL1,
			}, l1EthClient, l)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to create contract caller: %v", err)
				cancel()
				return
			}

			fmt.Printf("Submitting task result to AVS\n\n\n")
			receipt, err := avsCc.SubmitTaskResult(ctx, cert, tableData.LatestReferenceTimestamp)
			if err != nil {
				hasErrors = true
				l.Sugar().Errorf("Failed to submit task result: %v", err)
				cancel()
				return
			}
			assert.Nil(t, err)
			fmt.Printf("Receipt: %+v\n", receipt)

			cancel()
		}
	}()

	// submit a task
	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	task, err := l1CC.PublishMessageToInbox(ctx, chainConfig.AVSAccountAddress, 1, payloadJsonBytes)
	if err != nil {
		t.Fatalf("Failed to publish message to inbox: %v", err)
	}
	t.Logf("Task published: %+v", task)

	select {
	case <-time.After(90 * time.Second):
		cancel()
		t.Fatalf("Test timed out after 10 seconds")
	case <-ctx.Done():
		t.Logf("Test completed")
	}

	_ = l1Anvil.Process.Kill()
	assert.False(t, hasErrors)
}
