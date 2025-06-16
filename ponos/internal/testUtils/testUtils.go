package testUtils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/tableTransporter"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
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
	AVSAccountAddress          string `json:"avsAccountAddress"`
	AVSAccountPrivateKey       string `json:"avsAccountPk"`
	AppAccountAddress          string `json:"appAccountAddress"`
	AppAccountPrivateKey       string `json:"appAccountPk"`
	OperatorAccountPrivateKey  string `json:"operatorAccountPk"`
	OperatorAccountAddress     string `json:"operatorAccountAddress"`
	ExecOperatorAccountPk      string `json:"execOperatorAccountPk"`
	ExecOperatorAccountAddress string `json:"execOperatorAccountAddress"`
	MailboxContractAddressL1   string `json:"mailboxContractAddressL1"`
	MailboxContractAddressL2   string `json:"mailboxContractAddressL2"`
	AVSTaskRegistrarAddress    string `json:"avsTaskRegistrarAddress"`
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

func StartL1Anvil(projectRoot string, ctx context.Context) (*exec.Cmd, error) {
	forkUrl := "https://special-yolo-river.ethereum-holesky.quiknode.pro/2d21099a19e7c896a22b9fcc23dc8ce80f2214a5/"
	portNumber := "8545"
	blockTime := "2"
	forkBlockNumber := "3994152"
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
	forkUrl := "https://few-sly-dew.base-mainnet.quiknode.pro/eaecd36554bb2845570742c4e7aeda6f7dd0d5c1/"
	portNumber := "9545"
	blockTime := "2"
	forkBlockNumber := "30611001"
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

func ReadMailboxAbiJson(projectRoot string) ([]byte, error) {
	// read the mailbox ABI json file
	path, err := filepath.Abs(fmt.Sprintf("%s/../contracts/out/ITaskMailbox.sol/ITaskMailbox.json", projectRoot))
	if err != nil {
		return nil, err
	}

	abiJson, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("failed to read mailbox ABI json: %w", err))
	}

	type abiFile struct {
		Abi json.RawMessage `json:"abi"`
	}
	var abiFileData abiFile
	if err := json.Unmarshal(abiJson, &abiFileData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mailbox ABI json: %w", err)
	}

	return abiFileData.Abi, nil
}

func ReplaceMailboxAddressWithTestAddress(cs contractStore.IContractStore, chainConfig *ChainConfig) error {
	allContracts := cs.ListContracts()
	existingL1MailboxContract := util.Find(allContracts, func(c *contracts.Contract) bool {
		return c.Name == config.ContractName_TaskMailbox && c.ChainId == config.ChainId_EthereumAnvil
	})
	if existingL1MailboxContract == nil {
		return fmt.Errorf("existing mailbox contract not found for chain ID %d", config.ChainId_EthereumAnvil)
	}
	if err := cs.OverrideContract(config.ContractName_TaskMailbox, []config.ChainId{config.ChainId_EthereumAnvil}, &contracts.Contract{
		Address:     chainConfig.MailboxContractAddressL1,
		AbiVersions: existingL1MailboxContract.AbiVersions,
	}); err != nil {
		return fmt.Errorf("failed to override mailbox contract: %w", err)
	}

	existingL2MailboxContract := util.Find(allContracts, func(c *contracts.Contract) bool {
		return c.Name == config.ContractName_TaskMailbox && c.ChainId == config.ChainId_BaseAnvil
	})
	if existingL2MailboxContract == nil {
		return fmt.Errorf("existing mailbox contract not found for chain ID %d", config.ChainId_EthereumAnvil)
	}
	if err := cs.OverrideContract(config.ContractName_TaskMailbox, []config.ChainId{config.ChainId_BaseAnvil}, &contracts.Contract{
		Address:     chainConfig.MailboxContractAddressL2,
		AbiVersions: existingL1MailboxContract.AbiVersions,
	}); err != nil {
		return fmt.Errorf("failed to override mailbox contract: %w", err)
	}
	return nil
}

const (
	transportEcdsaPrivateKey = "0x2ba58f64c57faa1073d63add89799f2a0101855a8b289b1330cb500758d5d1ee"
	transportBlsPrivateKey   = "0x2ba58f64c57faa1073d63add89799f2a0101855a8b289b1330cb500758d5d1ee"
)

func TransportL1Tables(l *zap.Logger) {
	tableTransporter.TransportTable(
		transportEcdsaPrivateKey,
		"http://localhost:8545",
		31337,
		"0x0022d2014901F2AFBF5610dDFcd26afe2a65Ca6F",
		transportBlsPrivateKey,
		l,
	)
}
