package testUtils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractStore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contracts"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
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
	MailboxContractAddress     string `json:"mailboxContractAddress"`
	AVSTaskRegistrarAddress    string `json:"avsTaskRegistrarAddress"`
	OperatorAccountPrivateKey  string `json:"operatorAccountPk"`
	OperatorAccountAddress     string `json:"operatorAccountAddress"`
	ExecOperatorAccountPk      string `json:"execOperatorAccountPk"`
	ExecOperatorAccountAddress string `json:"execOperatorAccountAddress"`
}

type MultiChainConfig struct {
	L1 *ChainConfig `json:"l1"`
}

func ReadChainConfig(projectRoot string) (*MultiChainConfig, error) {
	filePath := fmt.Sprintf("%s/internal/testData/chain-config.json", projectRoot)

	// read the file into bytes
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cf *MultiChainConfig
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
	forkUrl := "https://tame-fabled-liquid.quiknode.pro/f27d4be93b4d7de3679f5c5ae881233f857407a0/"
	portNumber := "8545"
	blockTime := "2"
	forkBlockNumber := "22396947"

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
	})
}

type AnvilConfig struct {
	ForkUrl         string `json:"forkUrl"`
	ForkBlockNumber string `json:"forkBlockNumber"`
	BlockTime       string `json:"blockTime"`
	PortNumber      string `json:"portNumber"`
	StateFilePath   string `json:"stateFilePath"`
}

func StartAnvil(projectRoot string, ctx context.Context, cfg *AnvilConfig) (*exec.Cmd, error) {
	// exec anvil command to start the anvil node
	args := []string{
		"--fork-url", cfg.ForkUrl,
		"--fork-block-number", cfg.ForkBlockNumber,
		"--load-state", cfg.StateFilePath,
		"--block-time", cfg.BlockTime,
		"-vvv",
	}
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

	for i := 1; i < 10; i++ {
		res, err := http.Post("http://localhost:8545", "application/json", nil)
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
	existingMailboxContract := util.Find(allContracts, func(c *contracts.Contract) bool {
		return c.Name == config.ContractName_TaskMailbox && c.ChainId == config.ChainId_EthereumAnvil
	})
	if existingMailboxContract == nil {
		return fmt.Errorf("existing mailbox contract not found for chain ID %d", config.ChainId_EthereumAnvil)
	}

	if err := cs.OverrideContract(config.ContractName_TaskMailbox, []config.ChainId{31337}, &contracts.Contract{
		Address:     chainConfig.MailboxContractAddress,
		AbiVersions: existingMailboxContract.AbiVersions,
	}); err != nil {
		return fmt.Errorf("failed to override mailbox contract: %w", err)
	}
	return nil
}
