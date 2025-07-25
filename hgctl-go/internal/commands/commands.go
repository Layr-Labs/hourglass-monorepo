package commands

import (
    "github.com/urfave/cli/v2"
)

// GetCommand returns the get command
func GetCommand() *cli.Command {
    return &cli.Command{
        Name:  "get",
        Usage: "Get resources",
        Subcommands: []*cli.Command{
            {
                Name:      "release",
                Usage:     "List releases for an AVS",
                ArgsUsage: "<avs-address>",
                Flags: []cli.Flag{
                    &cli.Uint64Flag{
                        Name:  "limit",
                        Usage: "Maximum number of releases to return",
                        Value: 10,
                    },
                },
                Action: getReleaseAction,
            },
            {
                Name:      "performer",
                Usage:     "List deployed performers",
                ArgsUsage: "[avs-address]",
                Action:    getPerformerAction,
            },
        },
    }
}

// DescribeCommand returns the describe command
func DescribeCommand() *cli.Command {
    return &cli.Command{
        Name:  "describe",
        Usage: "Show detailed information about resources",
        Subcommands: []*cli.Command{
            {
                Name:      "release",
                Usage:     "Show detailed release information with runtime spec",
                ArgsUsage: "<avs-address> <release-id>",
                Flags: []cli.Flag{
                    &cli.Uint64Flag{
                        Name:     "operator-set-id",
                        Usage:    "Operator set ID",
                        Required: true,
                    },
                },
                Action: describeReleaseAction,
            },
        },
    }
}

// DeployCommand returns the deploy command
func DeployCommand() *cli.Command {
    return &cli.Command{
        Name:  "deploy",
        Usage: "Deploy resources",
        Subcommands: []*cli.Command{
            {
                Name:      "artifact",
                Usage:     "Deploy an AVS artifact from a release",
                ArgsUsage: "<avs-address>",
                Flags: []cli.Flag{
                    &cli.Uint64Flag{
                        Name:     "operator-set-id",
                        Usage:    "Operator set ID",
                        Required: true,
                    },
                    &cli.StringFlag{
                        Name:  "version",
                        Usage: "Release version to deploy (defaults to latest)",
                    },
                    &cli.StringFlag{
                        Name:  "legacy-digest",
                        Usage: "Legacy mode: deploy using direct digest",
                    },
                    &cli.StringFlag{
                        Name:  "registry-url",
                        Usage: "Legacy mode: registry URL",
                    },
                },
                Action: deployArtifactAction,
            },
        },
    }
}

// TranslateCommand returns the translate command
func TranslateCommand() *cli.Command {
    return &cli.Command{
        Name:  "translate",
        Usage: "Translate runtime specs to different formats",
        Subcommands: []*cli.Command{
            {
                Name:  "compose",
                Usage: "Translate EigenRuntime spec to Docker Compose format",
                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:    "input",
                        Aliases: []string{"i"},
                        Usage:   "Input file (- for stdin)",
                        Value:   "-",
                    },
                    &cli.StringFlag{
                        Name:    "output",
                        Aliases: []string{"o"},
                        Usage:   "Output file (- for stdout)",
                        Value:   "-",
                    },
                },
                Action: translateComposeAction,
            },
            {
                Name:  "container",
                Usage: "Translate EigenRuntime spec to container run commands",
                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:    "input",
                        Aliases: []string{"i"},
                        Usage:   "Input file (- for stdin)",
                        Value:   "-",
                    },
                    &cli.StringFlag{
                        Name:    "output",
                        Aliases: []string{"o"},
                        Usage:   "Output file (- for stdout)",
                        Value:   "-",
                    },
                },
                Action: translateContainerAction,
            },
        },
    }
}

// RemoveCommand returns the remove command
func RemoveCommand() *cli.Command {
    return &cli.Command{
        Name:  "remove",
        Usage: "Remove resources",
        Subcommands: []*cli.Command{
            {
                Name:      "performer",
                Usage:     "Remove a deployed performer",
                ArgsUsage: "<performer-id>",
                Action:    removePerformerAction,
            },
        },
    }
}

// ContextCommand returns the context command
func ContextCommand() *cli.Command {
    return &cli.Command{
        Name:  "context",
        Usage: "Manage contexts",
        Subcommands: []*cli.Command{
            {
                Name:   "list",
                Usage:  "List all contexts",
                Action: contextListAction,
            },
            {
                Name:      "use",
                Usage:     "Switch to a different context",
                ArgsUsage: "<context-name>",
                Action:    contextUseAction,
            },
            {
                Name:  "set",
                Usage: "Set values in the current context",
                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:  "executor-address",
                        Usage: "Executor service address",
                    },
                    &cli.StringFlag{
                        Name:  "avs-address",
                        Usage: "AVS address",
                    },
                    &cli.Uint64Flag{
                        Name:  "operator-set-id",
                        Usage: "Operator set ID",
                    },
                    &cli.StringFlag{
                        Name:  "rpc-url",
                        Usage: "Ethereum RPC URL",
                    },
                    &cli.StringFlag{
                        Name:  "release-manager",
                        Usage: "Release manager contract address",
                    },
                },
                Action: contextSetAction,
            },
            {
                Name:   "show",
                Usage:  "Show current context",
                Action: contextShowAction,
            },
        },
    }
}
