package testUtils

import (
	"context"
	"encoding/json"
	"fmt"
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

		match := regexp.MustCompile(`\/hourglass-monorepo\/ponos$`)

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

func StartAnvil(projectRoot string, ctx context.Context) (*exec.Cmd, error) {
	// exec anvil command to start the anvil node
	fullPath, err := filepath.Abs(fmt.Sprintf("%s/internal/testData/anvil-state.json", projectRoot))
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

	args := []string{
		"--fork-url", "https://tame-fabled-liquid.quiknode.pro/f27d4be93b4d7de3679f5c5ae881233f857407a0/",
		"--fork-block-number", "22396947",
		"--load-state", fullPath,
		"--block-time", "2",
		"-vvv",
	}
	cmd := exec.CommandContext(ctx, "anvil", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
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
