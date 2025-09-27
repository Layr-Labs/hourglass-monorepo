package testUtils

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/tableTransporter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

const (
	L1Web3SignerUrl = "http://localhost:9100"
	L2Web3SignerUrl = "http://localhost:9101"
)

func GetProjectRootPath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	startingPath := ""
	iterations := 0
	for {
		if iterations > 10 {
			panic("Could not find project root path")
		}
		iterations++
		p, err := filepath.Abs(fmt.Sprintf("%s/%s", wd, startingPath))
		if err != nil {
			panic(err)
		}

		match := regexp.MustCompile(`\/hourglass-monorepo(.+)?\/ponos$`)

		if match.MatchString(p) {
			return p
		}
		startingPath = startingPath + "/.."
	}
}

type ChainConfig struct {
	AVSAccountAddress             string `json:"avsAccountAddress"`
	AVSAccountPrivateKey          string `json:"avsAccountPk"`
	AVSAccountPublicKey           string `json:"avsAccountPublicKey"`
	AppAccountAddress             string `json:"appAccountAddress"`
	AppAccountPrivateKey          string `json:"appAccountPk"`
	AppAccountPublicKey           string `json:"appAccountPublicKey"`
	OperatorAccountPrivateKey     string `json:"operatorAccountPk"`
	OperatorAccountAddress        string `json:"operatorAccountAddress"`
	OperatorAccountPublicKey      string `json:"operatorAccountPublicKey"`
	ExecOperatorAccountPk         string `json:"execOperatorAccountPk"`
	ExecOperatorAccountAddress    string `json:"execOperatorAccountAddress"`
	ExecOperatorAccountPublicKey  string `json:"execOperatorAccountPublicKey"`
	ExecOperator2AccountPk        string `json:"execOperator2AccountPk"`
	ExecOperator2AccountAddress   string `json:"execOperator2AccountAddress"`
	ExecOperator2AccountPublicKey string `json:"execOperator2AccountPublicKey"`
	ExecOperator3AccountPk        string `json:"execOperator3AccountPk"`
	ExecOperator3AccountAddress   string `json:"execOperator3AccountAddress"`
	ExecOperator3AccountPublicKey string `json:"execOperator3AccountPublicKey"`
	ExecOperator4AccountPk        string `json:"execOperator4AccountPk"`
	ExecOperator4AccountAddress   string `json:"execOperator4AccountAddress"`
	ExecOperator4AccountPublicKey string `json:"execOperator4AccountPublicKey"`
	AVSTaskRegistrarAddress       string `json:"avsTaskRegistrarAddress"`
	AVSTaskHookAddressL1          string `json:"avsTaskHookAddressL1"`
	AVSTaskHookAddressL2          string `json:"avsTaskHookAddressL2"`
	AggStakerAccountPrivateKey    string `json:"aggStakerAccountPk"`
	AggStakerAccountAddress       string `json:"aggStakerAccountAddress"`
	ExecStakerAccountPrivateKey   string `json:"execStakerAccountPk"`
	ExecStakerAccountAddress      string `json:"execStakerAccountAddress"`
	ExecStaker2AccountPrivateKey  string `json:"execStaker2AccountPk"`
	ExecStaker2AccountAddress     string `json:"execStaker2AccountAddress"`
	ExecStaker3AccountPrivateKey  string `json:"execStaker3AccountPk"`
	ExecStaker3AccountAddress     string `json:"execStaker3AccountAddress"`
	ExecStaker4AccountPrivateKey  string `json:"execStaker4AccountPk"`
	ExecStaker4AccountAddress     string `json:"execStaker4AccountAddress"`
}

func ReadChainConfig(projectRoot string) (*ChainConfig, error) {
	filePath := fmt.Sprintf("%s/internal/testData/chain-config.json", projectRoot)

	// read the file into bytes
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cf *ChainConfig
	if err := json.Unmarshal(file, &cf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}
	return cf, nil
}

func ReadTenderlyChainConfig(projectRoot string) (*ChainConfig, error) {
	filePath := fmt.Sprintf("%s/internal/testData/tenderly-chain-config.json", projectRoot)

	// read the file into bytes
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cf *ChainConfig
	if err := json.Unmarshal(file, &cf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}
	return cf, nil
}

func WaitForAnvil(
	anvilWg *sync.WaitGroup,
	ctx context.Context,
	t *testing.T,
	ethereumClient *ethereum.EthereumClient,
	errorsChan chan error,
) {
	defer anvilWg.Done()
	time.Sleep(2 * time.Second) // give anvil some time to start

	for {
		select {
		case <-ctx.Done():
			t.Logf("Failed to start l1Anvil: %v", ctx.Err())
			errorsChan <- fmt.Errorf("failed to start l1Anvil: %w", ctx.Err())
			return
		case <-time.After(2 * time.Second):
			t.Logf("Checking if anvil is up and running...")
			block, err := ethereumClient.GetLatestBlock(ctx)
			if err != nil {
				t.Logf("Failed to get latest block, will retry: %v", err)
				continue
			}
			t.Logf("L1 Anvil is up and running, latest block: %v", block)
			return
		}
	}
}

func KillallAnvils() error {
	cmd := exec.Command("pkill", "-f", "anvil")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to kill all anvils: %w", err)
	}
	fmt.Println("All anvil processes killed successfully")
	return nil
}

func StartL1Anvil(projectRoot string, ctx context.Context) (*exec.Cmd, error) {
	forkUrl := "https://practical-serene-mound.ethereum-sepolia.quiknode.pro/3aaa48bd95f3d6aed60e89a1a466ed1e2a440b61/"
	portNumber := "8545"
	blockTime := "2"
	forkBlockNumber := "9259025"
	chainId := "31337"

	fullPath, err := filepath.Abs(fmt.Sprintf("%s/internal/testData/anvil-l1-state.json", projectRoot))
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	stat, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", fullPath)
	}

	return StartAnvil(projectRoot, ctx, &AnvilConfig{
		ForkUrl:         forkUrl,
		ForkBlockNumber: forkBlockNumber,
		BlockTime:       blockTime,
		PortNumber:      portNumber,
		StateFilePath:   fullPath,
		ChainId:         chainId,
	})
}

func StartL2Anvil(projectRoot string, ctx context.Context) (*exec.Cmd, error) {
	forkUrl := "https://soft-alpha-grass.base-sepolia.quiknode.pro/fd5e4bf346247d9b6e586008a9f13df72ce6f5b2/"
	portNumber := "9545"
	blockTime := "2"
	forkBlockNumber := "31408197"
	chainId := "31338"

	fullPath, err := filepath.Abs(fmt.Sprintf("%s/internal/testData/anvil-l2-state.json", projectRoot))
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	stat, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", fullPath)
	}

	return StartAnvil(projectRoot, ctx, &AnvilConfig{
		ForkUrl:         forkUrl,
		ForkBlockNumber: forkBlockNumber,
		BlockTime:       blockTime,
		PortNumber:      portNumber,
		StateFilePath:   fullPath,
		ChainId:         chainId,
	})
}

type AnvilConfig struct {
	ForkUrl         string `json:"forkUrl"`
	ForkBlockNumber string `json:"forkBlockNumber"`
	BlockTime       string `json:"blockTime"`
	PortNumber      string `json:"portNumber"`
	StateFilePath   string `json:"stateFilePath"`
	ChainId         string `json:"chainId"`
}

func StartAnvil(projectRoot string, ctx context.Context, cfg *AnvilConfig) (*exec.Cmd, error) {
	// exec anvil command to start the anvil node
	args := []string{
		"--fork-url", cfg.ForkUrl,
		"--load-state", cfg.StateFilePath,
		"--chain-id", cfg.ChainId,
		"--port", cfg.PortNumber,
		"--block-time", cfg.BlockTime,
		"--fork-block-number", cfg.ForkBlockNumber,
		"-vvv",
	}
	fmt.Printf("Starting anvil with args: %v\n", args)
	cmd := exec.CommandContext(ctx, "anvil", args...)
	cmd.Stderr = os.Stderr

	joinOutput := os.Getenv("JOIN_ANVIL_OUTPUT")
	if joinOutput == "true" {
		cmd.Stdout = os.Stdout
	}

	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start anvil: %w", err)
	}

	rpcUrl := fmt.Sprintf("http://localhost:%s", cfg.PortNumber)

	for i := 1; i < 10; i++ {
		res, err := http.Post(rpcUrl, "application/json", nil)
		if err == nil && res.StatusCode == 200 {
			fmt.Println("Anvil is up and running")
			return cmd, nil
		}
		fmt.Printf("Anvil not ready yet, retrying... %d\n", i)
		time.Sleep(time.Second * time.Duration(i))
	}

	return nil, fmt.Errorf("failed to start anvil")
}

func KillAnvil(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return fmt.Errorf("anvil command is not running")
	}

	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill anvil process: %w", err)
	}
	_ = cmd.Wait()

	fmt.Println("Anvil process killed successfully")
	return nil
}

func TransportStakeTables(l *zap.Logger, includeL2 bool) {
	transportBlsPrivateKey := os.Getenv("HOURGLASS_TRANSPORT_BLS_KEY")
	if transportBlsPrivateKey == "" {
		transportBlsPrivateKey = "0x5f8e6420b9cb0c940e3d3f8b99177980785906d16fb3571f70d7a05ecf5f2172"
	}
	transportEcdsaPrivateKey := transportBlsPrivateKey
	chainIdsToIgnore := []*big.Int{
		new(big.Int).SetUint64(11155111), // eth sepolia
		new(big.Int).SetUint64(17000),    // holesky
		new(big.Int).SetUint64(84532),    // base sepolia
	}

	var l2RpcUrl string
	var l2ChainId uint64
	if includeL2 {
		l2RpcUrl = "http://localhost:9545"
		l2ChainId = 31338
	} else {
		chainIdsToIgnore = append(chainIdsToIgnore, new(big.Int).SetUint64(31338))
	}

	contractAddresses := config.CoreContracts[config.ChainId_EthereumAnvil]

	tableTransporter.TransportTable(
		transportEcdsaPrivateKey,
		"http://localhost:8545",
		31337,
		l2RpcUrl,
		l2ChainId,
		contractAddresses.CrossChainRegistry,
		transportBlsPrivateKey,
		chainIdsToIgnore,
		l,
	)
}

// TransportStakeTablesWithMultipleOperatorsConfig transports stake tables with configurable L2 support
func TransportStakeTablesWithMultipleOperatorsConfig(
	l *zap.Logger,
	operators []tableTransporter.OperatorBLSInfo,
	transporterPrivateKey string,
	operatorSetId uint32,
	avsAddress string,
	l2RpcUrl string,
	l2ChainId uint64,
	chainIdsToIgnore []*big.Int,
) error {
	contractAddresses := config.CoreContracts[config.ChainId_EthereumAnvil]

	transportBLSKey := "0x5f8e6420b9cb0c940e3d3f8b99177980785906d16fb3571f70d7a05ecf5f2172"

	cfg := &tableTransporter.MultipleOperatorConfig{
		TransporterPrivateKey:     transporterPrivateKey,
		L1RpcUrl:                  "http://localhost:8545",
		L1ChainId:                 31337,
		L2RpcUrl:                  l2RpcUrl,
		L2ChainId:                 l2ChainId,
		CrossChainRegistryAddress: contractAddresses.CrossChainRegistry,
		ChainIdsToIgnore:          chainIdsToIgnore,
		Logger:                    l,
		Operators:                 operators,
		AVSAddress:                common.HexToAddress(avsAddress),
		OperatorSetId:             operatorSetId,
		TransportBLSPrivateKey:    transportBLSKey,
	}

	return tableTransporter.TransportTableWithSimpleMultiOperators(cfg)
}
