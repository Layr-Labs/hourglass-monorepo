package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/context"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/deploy"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/describe"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/get"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/remove"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/translate"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/web3signer"
)

// GetCommand returns the get command
func GetCommand() *cli.Command {
	return get.Command()
}

// DescribeCommand returns the describe command
func DescribeCommand() *cli.Command {
	return describe.Command()
}

// DeployCommand returns the deploy command
func DeployCommand() *cli.Command {
	return deploy.Command()
}

// TranslateCommand returns the translate command
func TranslateCommand() *cli.Command {
	return translate.Command()
}

// RemoveCommand returns the remove command
func RemoveCommand() *cli.Command {
	return remove.Command()
}

// ContextCommand returns the context command
func ContextCommand() *cli.Command {
	return context.Command()
}

// KeystoreCommand returns the keystore command
func KeystoreCommand() *cli.Command {
	return keystore.Command()
}

// Web3SignerCommand returns the web3signer command
func Web3SignerCommand() *cli.Command {
	return web3signer.Command()
}