package aggregator

import (
	"context"
	"fmt"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// AggregatorStateGenerator generates test states for aggregator tests
type AggregatorStateGenerator struct {
	aggregatorCurve   config.CurveType
	executorCurve     config.CurveType
	aggregatorOpsetId uint32
	executorOpsetId   uint32
	stateID           string
	description       string
	chainConfig       *testUtils.ChainConfig
	logger            *zap.SugaredLogger
}

// NewAggregatorStateGenerator creates a new aggregator state generator
func NewAggregatorStateGenerator(
	aggregatorCurve config.CurveType,
	executorCurve config.CurveType,
	aggregatorOpsetId uint32,
	executorOpsetId uint32,
	chainConfig *testUtils.ChainConfig,
) *AggregatorStateGenerator {
	stateID := fmt.Sprintf("%s_agg_%s_exec", 
		getCurveShortName(aggregatorCurve), 
		getCurveShortName(executorCurve))
	
	description := fmt.Sprintf("Aggregator with %s curve (opset %d), Executor with %s curve (opset %d)",
		aggregatorCurve, aggregatorOpsetId, executorCurve, executorOpsetId)
	
	log, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	
	return &AggregatorStateGenerator{
		aggregatorCurve:   aggregatorCurve,
		executorCurve:     executorCurve,
		aggregatorOpsetId: aggregatorOpsetId,
		executorOpsetId:   executorOpsetId,
		stateID:           stateID,
		description:       description,
		chainConfig:       chainConfig,
		logger:            log.Sugar(),
	}
}

// GetStateID returns the unique identifier for this state configuration
func (g *AggregatorStateGenerator) GetStateID() string {
	return g.stateID
}

// GetDescription returns a human-readable description of the state
func (g *AggregatorStateGenerator) GetDescription() string {
	return g.description
}

// GenerateState performs all expensive setup operations that can be pre-generated
func (g *AggregatorStateGenerator) GenerateState(ctx context.Context, l1Caller, l2Caller contractCaller.IContractCaller) error {
	g.logger.Infow("Starting aggregator state generation",
		"aggregatorCurve", g.aggregatorCurve,
		"executorCurve", g.executorCurve,
		"aggregatorOpsetId", g.aggregatorOpsetId,
		"executorOpsetId", g.executorOpsetId,
	)
	
	startTime := time.Now()
	
	// Step 1: Configure operator sets with their curve types
	if err := g.configureOperatorSets(ctx, l1Caller); err != nil {
		return fmt.Errorf("failed to configure operator sets: %w", err)
	}
	g.logger.Infow("Configured operator sets", "duration", time.Since(startTime))
	
	// Step 2: Register operators with their respective keys
	if err := g.registerOperators(ctx, l1Caller); err != nil {
		return fmt.Errorf("failed to register operators: %w", err)
	}
	g.logger.Infow("Registered operators", "duration", time.Since(startTime))
	
	// Step 3: Setup operator peering information
	if err := g.setupOperatorPeering(ctx, l1Caller); err != nil {
		return fmt.Errorf("failed to setup operator peering: %w", err)
	}
	g.logger.Infow("Setup operator peering", "duration", time.Since(startTime))
	
	// Step 4: Create generation reservations for operators
	if err := g.createGenerationReservations(ctx, l1Caller); err != nil {
		return fmt.Errorf("failed to create generation reservations: %w", err)
	}
	g.logger.Infow("Created generation reservations", "duration", time.Since(startTime))
	
	// Step 5: Transport stake tables (the most expensive operation)
	g.logger.Infow("Starting stake table transport")
	testUtils.TransportStakeTables(g.logger.Desugar(), true)
	g.logger.Infow("Completed stake table transport", "duration", time.Since(startTime))
	
	// Wait for tables to be fully synced
	time.Sleep(6 * time.Second)
	
	// Step 6: Setup task mailbox for executor operator sets
	if err := g.setupTaskMailbox(ctx, l1Caller, l2Caller); err != nil {
		return fmt.Errorf("failed to setup task mailbox: %w", err)
	}
	g.logger.Infow("Setup task mailbox", "duration", time.Since(startTime))
	
	g.logger.Infow("Aggregator state generation completed",
		"totalDuration", time.Since(startTime),
	)
	
	return nil
}

// configureOperatorSets configures the operator sets with their respective curve types
func (g *AggregatorStateGenerator) configureOperatorSets(ctx context.Context, l1Caller contractCaller.IContractCaller) error {
	// Need to cast to concrete type to access the underlying client
	concreteCaller := l1Caller.(*caller.ContractCaller)
	ethClient := concreteCaller.GetEthClient()
	
	// Create AVS config signer
	avsConfigSigner, err := transactionSigner.NewPrivateKeySigner(
		g.chainConfig.AVSAccountPrivateKey,
		ethClient,
		&logger.Logger{Logger: g.logger.Desugar()},
	)
	if err != nil {
		return fmt.Errorf("failed to create AVS config signer: %w", err)
	}
	
	// Create AVS config caller
	avsConfigCaller, err := caller.NewContractCaller(ethClient, avsConfigSigner, &logger.Logger{Logger: g.logger.Desugar()})
	if err != nil {
		return fmt.Errorf("failed to create AVS config caller: %w", err)
	}
	
	avsAddress := common.HexToAddress(g.chainConfig.AVSAccountAddress)
	
	// Configure aggregator operator set
	g.logger.Infow("Configuring aggregator operator set",
		"opsetId", g.aggregatorOpsetId,
		"curve", g.aggregatorCurve,
	)
	if _, err := avsConfigCaller.ConfigureAVSOperatorSet(ctx, avsAddress, g.aggregatorOpsetId, g.aggregatorCurve); err != nil {
		return fmt.Errorf("failed to configure aggregator operator set: %w", err)
	}
	
	// Configure executor operator set if different
	if g.executorOpsetId != g.aggregatorOpsetId {
		g.logger.Infow("Configuring executor operator set",
			"opsetId", g.executorOpsetId,
			"curve", g.executorCurve,
		)
		if _, err := avsConfigCaller.ConfigureAVSOperatorSet(ctx, avsAddress, g.executorOpsetId, g.executorCurve); err != nil {
			return fmt.Errorf("failed to configure executor operator set: %w", err)
		}
	}
	
	return nil
}

// registerOperators registers both aggregator and executor operators
func (g *AggregatorStateGenerator) registerOperators(ctx context.Context, l1Caller contractCaller.IContractCaller) error {
	// Cast to concrete type
	concreteCaller := l1Caller.(*caller.ContractCaller)
	ethClient := concreteCaller.GetEthClient()
	
	avsAddress := common.HexToAddress(g.chainConfig.AVSAccountAddress)
	
	// Register aggregator operator
	aggOperatorAddress := common.HexToAddress(g.chainConfig.OperatorAccountAddress)
	aggSigner, err := transactionSigner.NewPrivateKeySigner(
		g.chainConfig.OperatorAccountPrivateKey,
		ethClient,
		&logger.Logger{Logger: g.logger.Desugar()},
	)
	if err != nil {
		return fmt.Errorf("failed to create aggregator signer: %w", err)
	}
	
	aggCaller, err := caller.NewContractCaller(ethClient, aggSigner, &logger.Logger{Logger: g.logger.Desugar()})
	if err != nil {
		return fmt.Errorf("failed to create aggregator caller: %w", err)
	}
	
	// Create and register aggregator operator
	g.logger.Infow("Registering aggregator operator",
		"address", aggOperatorAddress.Hex(),
		"opsetId", g.aggregatorOpsetId,
	)
	
	_, err = aggCaller.CreateOperatorAndRegisterWithAvs(
		ctx,
		avsAddress,
		aggOperatorAddress,
		[]uint32{g.aggregatorOpsetId},
		"aggregator:9090",
		0, // allocation delay
		"aggregator-metadata",
	)
	if err != nil {
		return fmt.Errorf("failed to register aggregator operator: %w", err)
	}
	
	// Register aggregator keys based on curve type
	if err := g.registerOperatorKeys(ctx, aggCaller, aggOperatorAddress, g.aggregatorOpsetId, g.aggregatorCurve); err != nil {
		return fmt.Errorf("failed to register aggregator keys: %w", err)
	}
	
	// Register executor operator
	execOperatorAddress := common.HexToAddress(g.chainConfig.ExecOperatorAccountAddress)
	execSigner, err := transactionSigner.NewPrivateKeySigner(
		g.chainConfig.ExecOperatorAccountPk,
		ethClient,
		&logger.Logger{Logger: g.logger.Desugar()},
	)
	if err != nil {
		return fmt.Errorf("failed to create executor signer: %w", err)
	}
	
	execCaller, err := caller.NewContractCaller(ethClient, execSigner, &logger.Logger{Logger: g.logger.Desugar()})
	if err != nil {
		return fmt.Errorf("failed to create executor caller: %w", err)
	}
	
	g.logger.Infow("Registering executor operator",
		"address", execOperatorAddress.Hex(),
		"opsetId", g.executorOpsetId,
	)
	
	_, err = execCaller.CreateOperatorAndRegisterWithAvs(
		ctx,
		avsAddress,
		execOperatorAddress,
		[]uint32{g.executorOpsetId},
		"executor:9091",
		0, // allocation delay
		"executor-metadata",
	)
	if err != nil {
		return fmt.Errorf("failed to register executor operator: %w", err)
	}
	
	// Register executor keys based on curve type
	if err := g.registerOperatorKeys(ctx, execCaller, execOperatorAddress, g.executorOpsetId, g.executorCurve); err != nil {
		return fmt.Errorf("failed to register executor keys: %w", err)
	}
	
	return nil
}

// registerOperatorKeys registers the appropriate keys for an operator based on curve type
func (g *AggregatorStateGenerator) registerOperatorKeys(
	ctx context.Context,
	opCaller *caller.ContractCaller,
	operatorAddress common.Address,
	opsetId uint32,
	curveType config.CurveType,
) error {
	avsAddress := common.HexToAddress(g.chainConfig.AVSAccountAddress)
	
	switch curveType {
	case config.CurveTypeBN254:
		// For BN254, we need to register BLS keys
		// This would typically involve generating or loading BLS keys
		// For now, we'll use placeholder logic
		g.logger.Infow("Registering BN254 keys for operator",
			"operator", operatorAddress.Hex(),
			"opsetId", opsetId,
		)
		// TODO: Implement actual BN254 key registration
		
	case config.CurveTypeECDSA:
		// For ECDSA, register the operator's address as the signing key
		g.logger.Infow("Registering ECDSA keys for operator",
			"operator", operatorAddress.Hex(),
			"opsetId", opsetId,
		)
		
		msgHash, err := opCaller.GetOperatorECDSAKeyRegistrationMessageHash(
			ctx,
			operatorAddress,
			avsAddress,
			opsetId,
			operatorAddress, // Using operator address as signing key
		)
		if err != nil {
			return fmt.Errorf("failed to get ECDSA registration message hash: %w", err)
		}
		
		// Sign the message hash
		// TODO: Implement actual signing logic
		_ = msgHash
		
	default:
		return fmt.Errorf("unsupported curve type: %v", curveType)
	}
	
	return nil
}

// setupOperatorPeering sets up peering information for operators
func (g *AggregatorStateGenerator) setupOperatorPeering(ctx context.Context, l1Caller contractCaller.IContractCaller) error {
	// This is typically handled by the operator registration process
	// Additional peering setup can be added here if needed
	g.logger.Info("Operator peering setup completed via registration")
	return nil
}

// createGenerationReservations creates generation reservations for operators
func (g *AggregatorStateGenerator) createGenerationReservations(ctx context.Context, l1Caller contractCaller.IContractCaller) error {
	avsAddress := common.HexToAddress(g.chainConfig.AVSAccountAddress)
	
	// Get table calculator addresses for each curve type
	aggTableCalculator := l1Caller.GetTableCalculatorAddress(g.aggregatorCurve)
	execTableCalculator := l1Caller.GetTableCalculatorAddress(g.executorCurve)
	
	// Create reservation for aggregator operator set
	g.logger.Infow("Creating generation reservation for aggregator",
		"opsetId", g.aggregatorOpsetId,
		"calculator", aggTableCalculator.Hex(),
	)
	
	_, err := l1Caller.CreateGenerationReservation(
		ctx,
		avsAddress,
		g.aggregatorOpsetId,
		aggTableCalculator,
		avsAddress, // owner
		300,        // max staleness period
	)
	if err != nil {
		return fmt.Errorf("failed to create aggregator generation reservation: %w", err)
	}
	
	// Create reservation for executor operator set if different
	if g.executorOpsetId != g.aggregatorOpsetId {
		g.logger.Infow("Creating generation reservation for executor",
			"opsetId", g.executorOpsetId,
			"calculator", execTableCalculator.Hex(),
		)
		
		_, err = l1Caller.CreateGenerationReservation(
			ctx,
			avsAddress,
			g.executorOpsetId,
			execTableCalculator,
			avsAddress, // owner
			300,        // max staleness period
		)
		if err != nil {
			return fmt.Errorf("failed to create executor generation reservation: %w", err)
		}
	}
	
	// Set operator table calculators
	_, err = l1Caller.SetOperatorTableCalculator(ctx, avsAddress, g.aggregatorOpsetId, aggTableCalculator)
	if err != nil {
		return fmt.Errorf("failed to set aggregator table calculator: %w", err)
	}
	
	if g.executorOpsetId != g.aggregatorOpsetId {
		_, err = l1Caller.SetOperatorTableCalculator(ctx, avsAddress, g.executorOpsetId, execTableCalculator)
		if err != nil {
			return fmt.Errorf("failed to set executor table calculator: %w", err)
		}
	}
	
	return nil
}

// setupTaskMailbox sets up the task mailbox for executor operator sets
func (g *AggregatorStateGenerator) setupTaskMailbox(ctx context.Context, l1Caller, l2Caller contractCaller.IContractCaller) error {
	avsAddress := common.HexToAddress(g.chainConfig.AVSAccountAddress)
	taskHookAddress := common.HexToAddress(g.chainConfig.AVSTaskHookAddressL1)
	
	// Setup task mailbox on L1
	g.logger.Infow("Setting up task mailbox on L1",
		"avs", avsAddress.Hex(),
		"taskHook", taskHookAddress.Hex(),
	)
	
	err := l1Caller.SetupTaskMailboxForAvs(
		ctx,
		avsAddress,
		taskHookAddress,
		[]uint32{g.executorOpsetId},
		[]config.CurveType{g.executorCurve},
	)
	if err != nil {
		return fmt.Errorf("failed to setup L1 task mailbox: %w", err)
	}
	
	// Setup task mailbox on L2 if needed
	if g.chainConfig.AVSTaskHookAddressL2 != "" {
		taskHookAddressL2 := common.HexToAddress(g.chainConfig.AVSTaskHookAddressL2)
		g.logger.Infow("Setting up task mailbox on L2",
			"avs", avsAddress.Hex(),
			"taskHook", taskHookAddressL2.Hex(),
		)
		
		err = l2Caller.SetupTaskMailboxForAvs(
			ctx,
			avsAddress,
			taskHookAddressL2,
			[]uint32{g.executorOpsetId},
			[]config.CurveType{g.executorCurve},
		)
		if err != nil {
			return fmt.Errorf("failed to setup L2 task mailbox: %w", err)
		}
	}
	
	return nil
}

// getCurveShortName returns a short name for the curve type
func getCurveShortName(curve config.CurveType) string {
	switch curve {
	case config.CurveTypeBN254:
		return "bn254"
	case config.CurveTypeECDSA:
		return "ecdsa"
	default:
		return "unknown"
	}
}