package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/harness"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
)

func main() {
	// Define command-line flags
	var (
		outputDir   = flag.String("output", "./internal/testData", "Output directory for generated states")
		overwrite   = flag.Bool("overwrite", false, "Overwrite existing states")
		testSuite   = flag.String("suite", "", "Specific test suite to generate (empty for all)")
		verbose     = flag.Bool("verbose", false, "Enable verbose logging")
		projectRoot = flag.String("root", "", "Project root directory (auto-detected if not specified)")
	)
	
	flag.Parse()
	
	// Determine project root
	root := *projectRoot
	if root == "" {
		root = testUtils.GetProjectRootPath()
	}
	
	fmt.Printf("Test State Generator\n")
	fmt.Printf("===================\n")
	fmt.Printf("Project root: %s\n", root)
	fmt.Printf("Output directory: %s\n", *outputDir)
	fmt.Printf("Overwrite existing: %v\n", *overwrite)
	
	// Load chain configuration
	chainConfig, err := testUtils.ReadChainConfig(root)
	if err != nil {
		log.Fatalf("Failed to read chain config: %v", err)
	}
	
	// Create generator configuration
	generatorConfig := &harness.GeneratorConfig{
		OutputDir:       *outputDir,
		Overwrite:       *overwrite,
		L1RPCURL:        "http://localhost:8545",
		L2RPCURL:        "http://localhost:9545",
		L1WSURL:         "ws://localhost:8545",
		L2WSURL:         "ws://localhost:9545",
		BaseL1StatePath: filepath.Join(*outputDir, "anvil-l1-state.json"),
		BaseL2StatePath: filepath.Join(*outputDir, "anvil-l2-state.json"),
		ChainConfigPath: filepath.Join(*outputDir, "chain-config.json"),
		Verbose:         *verbose,
	}
	
	// Create test data generator
	generator, err := harness.NewTestDataGenerator(generatorConfig)
	if err != nil {
		log.Fatalf("Failed to create test data generator: %v", err)
	}
	
	// Register aggregator state generators
	registerAggregatorGenerators(generator, chainConfig)
	
	// Generate states
	ctx := context.Background()
	
	if *testSuite != "" {
		// Generate specific test suite
		fmt.Printf("\nGenerating state for test suite: %s\n", *testSuite)
		state, err := generator.GenerateState(ctx, *testSuite)
		if err != nil {
			log.Fatalf("Failed to generate state: %v", err)
		}
		printGeneratedState(state)
	} else {
		// Generate all registered states
		fmt.Printf("\nGenerating states for all test suites...\n")
		if err := generator.GenerateAll(ctx); err != nil {
			log.Fatalf("Failed to generate states: %v", err)
		}
	}
	
	fmt.Printf("\n✅ State generation completed successfully!\n")
}

// registerAggregatorGenerators registers all aggregator test state generators
func registerAggregatorGenerators(generator *harness.TestDataGenerator, chainConfig *testUtils.ChainConfig) {
	// Define the test configurations (matching aggregator_test.go)
	testConfigs := []struct {
		name              string
		aggregatorCurve   config.CurveType
		executorCurve     config.CurveType
		aggregatorOpsetId uint32
		executorOpsetId   uint32
	}{
		{
			name:              "aggregator/bn254_ecdsa",
			aggregatorCurve:   config.CurveTypeBN254,
			executorCurve:     config.CurveTypeECDSA,
			aggregatorOpsetId: 0,
			executorOpsetId:   1,
		},
		{
			name:              "aggregator/ecdsa_bn254",
			aggregatorCurve:   config.CurveTypeECDSA,
			executorCurve:     config.CurveTypeBN254,
			aggregatorOpsetId: 0,
			executorOpsetId:   1,
		},
	}
	
	// Register a generator for each configuration
	for _, cfg := range testConfigs {
		stateGen := aggregator.NewAggregatorStateGenerator(
			cfg.aggregatorCurve,
			cfg.executorCurve,
			cfg.aggregatorOpsetId,
			cfg.executorOpsetId,
			chainConfig,
		)
		
		generator.RegisterGenerator(cfg.name, stateGen)
		
		fmt.Printf("Registered generator: %s\n", cfg.name)
		fmt.Printf("  - %s\n", stateGen.GetDescription())
	}
}

// printGeneratedState prints information about a generated state
func printGeneratedState(state *harness.GeneratedState) {
	fmt.Printf("\nGenerated State:\n")
	fmt.Printf("  ID: %s\n", state.Definition.ID)
	fmt.Printf("  Description: %s\n", state.Definition.Description)
	fmt.Printf("  Test Suite: %s\n", state.Definition.TestSuite)
	fmt.Printf("  Created: %s\n", state.Definition.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Files:\n")
	fmt.Printf("    - L1 State: %s\n", state.L1StatePath)
	fmt.Printf("    - L2 State: %s\n", state.L2StatePath)
	fmt.Printf("    - Metadata: %s\n", state.MetadataPath)
	
	if len(state.Definition.Config) > 0 {
		fmt.Printf("  Configuration:\n")
		for k, v := range state.Definition.Config {
			fmt.Printf("    - %s: %s\n", k, v)
		}
	}
}

// Usage function for help text
func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nGenerates test states for integration tests by pre-executing expensive blockchain operations.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate all test states\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate states for aggregator tests only\n")
		fmt.Fprintf(os.Stderr, "  %s -suite aggregator/bn254_ecdsa\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Regenerate existing states\n")
		fmt.Fprintf(os.Stderr, "  %s -overwrite\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Available test suites:\n")
		fmt.Fprintf(os.Stderr, "  - aggregator/bn254_ecdsa\n")
		fmt.Fprintf(os.Stderr, "  - aggregator/ecdsa_bn254\n")
	}
}

// Helper function to parse test suite argument
func parseTestSuite(suite string) (string, string) {
	parts := strings.Split(suite, "/")
	if len(parts) != 2 {
		return suite, ""
	}
	return parts[0], parts[1]
}