package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/transport"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type OperatorConfig struct {
	Address    string   `json:"address"`
	PrivateKey string   `json:"privateKey"`
	Weights    []string `json:"weights,omitempty"`
}

type TransportConfig struct {
	TransporterKey      string           `json:"transporterKey"`
	L1RpcUrl            string           `json:"l1RpcUrl"`
	L1ChainId           uint64           `json:"l1ChainId"`
	L2RpcUrl            string           `json:"l2RpcUrl,omitempty"`
	L2ChainId           uint64           `json:"l2ChainId,omitempty"`
	CrossChainRegistry  string           `json:"crossChainRegistry"`
	KeyRegistrarAddress string           `json:"keyRegistrarAddress"`
	AVSAddress          string           `json:"avsAddress"`
	OperatorSetId       uint32           `json:"operatorSetId"`
	CurveType           string           `json:"curveType"`
	TransportBLSKey     string           `json:"transportBlsKey"`
	Operators           []OperatorConfig `json:"operators"`
	ChainsToIgnore      []uint64         `json:"chainsToIgnore,omitempty"`
}

func main() {
	var (
		configFile = flag.String("config", "", "Path to JSON config file")
		verbose    = flag.Bool("v", false, "Verbose logging")
	)
	flag.Parse()

	if *configFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --config flag is required")
		flag.Usage()
		os.Exit(1)
	}

	data, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config file: %v\n", err)
		os.Exit(1)
	}

	var cfg TransportConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config JSON: %v\n", err)
		os.Exit(1)
	}

	var logger *zap.Logger
	if *verbose {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", err)
		}
	}(logger)

	var operators []transport.OperatorKeyInfo
	for _, op := range cfg.Operators {
		var weights []*big.Int
		for _, w := range op.Weights {
			weight := new(big.Int)
			weight.SetString(w, 10)
			weights = append(weights, weight)
		}
		if len(weights) == 0 {
			weights = []*big.Int{big.NewInt(1)}
		}

		operators = append(operators, transport.OperatorKeyInfo{
			PrivateKeyHex:   op.PrivateKey,
			Weights:         weights,
			OperatorAddress: common.HexToAddress(op.Address),
		})
	}

	var chainsToIgnore []*big.Int
	for _, chainId := range cfg.ChainsToIgnore {
		chainsToIgnore = append(chainsToIgnore, big.NewInt(int64(chainId)))
	}

	var curveType config.CurveType
	switch cfg.CurveType {
	case "BN254", "bn254":
		curveType = config.CurveTypeBN254
	case "ECDSA", "ecdsa":
		curveType = config.CurveTypeECDSA
	default:
		fmt.Fprintf(os.Stderr, "Invalid curve type: %s (must be BN254 or ECDSA)\n", cfg.CurveType)
		os.Exit(1)
	}

	transportCfg := &transport.MultipleOperatorConfig{
		TransporterPrivateKey:     cfg.TransporterKey,
		L1RpcUrl:                  cfg.L1RpcUrl,
		L1ChainId:                 cfg.L1ChainId,
		L2RpcUrl:                  cfg.L2RpcUrl,
		L2ChainId:                 cfg.L2ChainId,
		CrossChainRegistryAddress: cfg.CrossChainRegistry,
		KeyRegistrarAddress:       cfg.KeyRegistrarAddress,
		AVSAddress:                common.HexToAddress(cfg.AVSAddress),
		OperatorSetId:             cfg.OperatorSetId,
		CurveType:                 curveType,
		TransportBLSPrivateKey:    cfg.TransportBLSKey,
		ChainIdsToIgnore:          chainsToIgnore,
		Logger:                    logger,
		Operators:                 operators,
	}

	logger.Info("Starting operator table transport",
		zap.String("avsAddress", cfg.AVSAddress),
		zap.Uint32("operatorSetId", cfg.OperatorSetId),
		zap.Int("numOperators", len(operators)),
	)

	if err := transport.TransportTableWithMultiOperators(transportCfg); err != nil {
		logger.Fatal("Transport failed", zap.Error(err))
	}

	logger.Info("Transport completed successfully")
}
