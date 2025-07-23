package testUtils

import (
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IAllocationManager"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IBN254CertificateVerifier"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ICrossChainRegistry"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IECDSACertificateVerifier"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IKeyRegistrar"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/middleware-bindings/IBN254TableCalculator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/middleware-bindings/IECDSATableCalculator"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"testing"
	"time"
)

func DebugOpsetData(
	t *testing.T,
	chainConfig *ChainConfig,
	eigenlayerContractAddrs *config.CoreContractAddresses,
	l1EthClient *ethclient.Client,
	currentBlock uint64,
	operatorSets []uint32,
) {

	am, err := IAllocationManager.NewIAllocationManager(common.HexToAddress(eigenlayerContractAddrs.AllocationManager), l1EthClient)
	if err != nil {
		t.Fatalf("Failed to create allocation manager: %v", err)
	}
	ccr, err := ICrossChainRegistry.NewICrossChainRegistry(common.HexToAddress(eigenlayerContractAddrs.CrossChainRegistry), l1EthClient)
	if err != nil {
		t.Fatalf("Failed to create cross chain registry: %v", err)
	}

	kr, err := IKeyRegistrar.NewIKeyRegistrar(common.HexToAddress(eigenlayerContractAddrs.KeyRegistrar), l1EthClient)
	if err != nil {
		t.Fatalf("Failed to create key registrar: %v", err)
	}

	bn254CertVerifier, err := IBN254CertificateVerifier.NewIBN254CertificateVerifier(common.HexToAddress(eigenlayerContractAddrs.BN254CertificateVerifier), l1EthClient)
	if err != nil {
		t.Fatalf("Failed to create BN254 certificate verifier: %v", err)
	}

	ecdsaCertVerifier, err := IECDSACertificateVerifier.NewIECDSACertificateVerifier(common.HexToAddress(eigenlayerContractAddrs.ECDSACertificateVerifier), l1EthClient)
	if err != nil {
		t.Fatalf("Failed to create ECDSA certificate verifier: %v", err)
	}

	for _, opsetId := range operatorSets {
		t.Logf("============================ Debugging operator set %d ============================", opsetId)
		strategies, err := am.GetStrategiesInOperatorSet(&bind.CallOpts{}, IAllocationManager.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get strategies in operator set %d: %v", opsetId, err)
		}
		t.Logf("Strategies in operator set %d: %+v", opsetId, strategies)

		members, err := am.GetMembers(&bind.CallOpts{}, IAllocationManager.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get members in operator set %d: %v", opsetId, err)
		}
		t.Logf("Members in operator set %d: %+v", opsetId, members)

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
		t.Logf("Minimum slashable stake for operator set %d: %+v", opsetId, minSlashableStake)

		tableCalcAddr, err := ccr.GetOperatorTableCalculator(&bind.CallOpts{}, ICrossChainRegistry.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get operator table calculator for operator set %d: %v", opsetId, err)
		}
		t.Logf("Operator table calculator for operator set %d: %s", opsetId, tableCalcAddr.String())

		cfg, err := ccr.GetOperatorSetConfig(&bind.CallOpts{}, ICrossChainRegistry.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get operator set config for operator set %d: %v", opsetId, err)
		}
		t.Logf("Operator set config for operator set %d: %+v", opsetId, cfg)

		curve, err := kr.GetOperatorSetCurveType(&bind.CallOpts{}, IKeyRegistrar.OperatorSet{
			Id:  opsetId,
			Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
		})
		if err != nil {
			t.Fatalf("Failed to get operator set curve type: %v", err)
		}
		t.Logf("Operator set curve type for operator set %d: %d", opsetId, curve)

		curveType, err := config.ConvertSolidityEnumToCurveType(curve)
		if err != nil {
			t.Fatalf("Failed to convert curve type: %v", err)
		}

		if curveType == config.CurveTypeBN254 {
			tableCalc, err := IBN254TableCalculator.NewIBN254TableCalculator(tableCalcAddr, l1EthClient)
			if err != nil {
				t.Fatalf("Failed to create operator table calculator for operator set %d: %v", opsetId, err)
			}

			weights, err := tableCalc.GetOperatorSetWeights(&bind.CallOpts{}, IBN254TableCalculator.OperatorSet{
				Id:  opsetId,
				Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
			})
			if err != nil {
				t.Fatalf("Failed to get operator weights for operator set %d: %v", opsetId, err)
			}
			t.Logf("[bn254] Operator weights for operator set %d: %+v", opsetId, weights)

			tableBytes, err := tableCalc.CalculateOperatorTableBytes(&bind.CallOpts{}, IBN254TableCalculator.OperatorSet{
				Id:  opsetId,
				Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
			})
			if err != nil {
				t.Fatalf("Failed to calculate operator table bytes for operator set %d: %v", opsetId, err)
			}
			t.Logf("[bn254] Operator table bytes for operator set %d: %x", opsetId, tableBytes)

			latestRefTimestamp, err := bn254CertVerifier.LatestReferenceTimestamp(&bind.CallOpts{}, IBN254CertificateVerifier.OperatorSet{
				Id:  opsetId,
				Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
			})
			if err != nil {
				t.Fatalf("Failed to get latest reference timestamp for operator set %d: %v", opsetId, err)
			}
			t.Logf("[bn254] Latest reference timestamp for operator set %d: %d", opsetId, latestRefTimestamp)
		} else if curveType == config.CurveTypeECDSA {
			tableCalc, err := IECDSATableCalculator.NewIECDSATableCalculator(tableCalcAddr, l1EthClient)
			if err != nil {
				t.Fatalf("Failed to create operator table calculator for operator set %d: %v", opsetId, err)
			}

			weights, err := tableCalc.GetOperatorSetWeights(&bind.CallOpts{}, IECDSATableCalculator.OperatorSet{
				Id:  opsetId,
				Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
			})
			if err != nil {
				t.Fatalf("Failed to get operator weights for operator set %d: %v", opsetId, err)
			}
			t.Logf("[ecdsa] Operator weights for operator set %d: %+v", opsetId, weights)

			tableBytes, err := tableCalc.CalculateOperatorTableBytes(&bind.CallOpts{}, IECDSATableCalculator.OperatorSet{
				Id:  opsetId,
				Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
			})
			if err != nil {
				time.Sleep(time.Second * 300)
				t.Fatalf("Failed to calculate operator table bytes for operator set %d: %v", opsetId, err)
			}
			t.Logf("[ecdsa] Operator table bytes for operator set %d: %x", opsetId, tableBytes)

			latestRefTimestamp, err := ecdsaCertVerifier.LatestReferenceTimestamp(&bind.CallOpts{}, IECDSACertificateVerifier.OperatorSet{
				Id:  opsetId,
				Avs: common.HexToAddress(chainConfig.AVSAccountAddress),
			})
			if err != nil {
				t.Fatalf("Failed to get latest reference timestamp for operator set %d: %v", opsetId, err)
			}
			t.Logf("[ecdsa] Latest reference timestamp for operator set %d: %d", opsetId, latestRefTimestamp)
		} else {
			t.Fatalf("Unsupported curve type: %s", curveType)
		}
	}
}
