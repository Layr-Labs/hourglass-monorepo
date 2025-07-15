# Contract Call Sequencing in Hourglass Signature Flow

This document details the complete sequence of on-chain contract calls (both view and state-changing functions) that occur during the signature and certificate creation flow in the Hourglass system.

## Overview

The Hourglass system makes extensive use of on-chain contract calls to coordinate operators, retrieve stake weights, and submit final certificates. The flow is primarily read-heavy, with most calls being view functions for data retrieval.

## Contract Call Phases

### Phase 1: Operator Setup and Registration (One-time)

**Location**: `ponos/pkg/contractCaller/caller/caller.go`

This phase occurs once per operator when they join the system.

#### 1.1 Operator Registration (`CreateOperatorAndRegisterWithAvs`)

```go
// Check if operator already exists (VIEW)
isOperator := delegationManager.IsOperator(operatorAddress)

// Create operator if doesn't exist (STATE CHANGE)
if !isOperator {
    tx := delegationManager.RegisterAsOperator(
        common.Address{},    // earningsReceiver
        allocationDelay,     // allocationDelay
        metadataUri         // metadataURI
    )
}

// Register operator with AVS (STATE CHANGE)
tx := allocationManager.RegisterForOperatorSets(
    operatorAddress,
    IAllocationManager.IAllocationManagerTypesRegisterParams{
        Avs:            avsAddress,
        OperatorSetIds: operatorSetIds,
        Data:           encodedSocket,
    }
)
```

#### 1.2 Key Registration (`RegisterKeyWithKeyRegistrar`)

```go
// Get curve type for operator set (VIEW)
curveType := keyRegistrar.GetOperatorSetCurveType(operatorSet)

// Generate registration message hash (VIEW)
// For BN254:
messageHash := keyRegistrar.GetBN254KeyRegistrationMessageHash(
    operatorAddress,
    operatorSet,
    keyData
)

// For ECDSA:
messageHash := keyRegistrar.GetECDSAKeyRegistrationMessageHash(
    operatorAddress,
    operatorSet,
    signingKeyAddress
)

// Encode key data (VIEW)
// For BN254:
keyData := keyRegistrar.EncodeBN254KeyData(keyRegG1, keyRegG2)

// Register the key (STATE CHANGE)
tx := keyRegistrar.RegisterKey(
    operatorAddress,
    operatorSet,
    keyData,
    sigBytes
)
```

#### 1.3 AVS Operator Set Configuration (`ConfigureAVSOperatorSet`)

```go
// Configure operator set curve type (STATE CHANGE)
tx := keyRegistrar.ConfigureOperatorSet(
    operatorSet,
    solidityCurveType
)
```

#### 1.4 Delegation and Allocation

```go
// Get delegation approver (VIEW)
approver := delegationManager.DelegationApprover(operatorAddress)

// Delegate to operator (STATE CHANGE)
tx := delegationManager.DelegateTo(
    operatorAddress,
    approvalSignature,
    salt
)

// Get allocation delay (VIEW)
allocationDelay := allocationManager.GetAllocationDelay(operatorAddress)

// Modify allocations (STATE CHANGE)
tx := allocationManager.ModifyAllocations(
    operatorAddress,
    allocateParams
)
```

### Phase 2: Task Creation (Per Task)

#### 2.1 Task Creation (`PublishMessageToInbox`)

```go
// Create task in TaskMailbox (STATE CHANGE)
tx := taskMailbox.CreateTask(ITaskMailbox.ITaskMailboxTypesTaskParams{
    RefundCollector: address,
    AvsFee:          new(big.Int).SetUint64(0),
    ExecutorOperatorSet: ITaskMailbox.OperatorSet{
        Avs: common.HexToAddress(avsAddress),
        Id:  operatorSetId,
    },
    Payload: payload,
})
```

### Phase 3: Task Processing - Data Retrieval (Per Task)

This phase involves extensive view function calls to gather operator data before task distribution.

#### 3.1 AVS Configuration Retrieval (`GetAVSConfig`)

```go
// Get AVS registrar address (VIEW)
avsRegistrarAddress := allocationManager.GetAVSRegistrar(avsAddress)

// Get AVS configuration (VIEW)
avsConfig := avsRegistrar.GetAvsConfig()
```

#### 3.2 Operator Table Data Retrieval (`GetOperatorTableDataForOperatorSet`)

**Critical sequence for getting operator weights and metadata:**

```go
// Get operator table calculator address (VIEW)
otcAddr := crossChainRegistry.GetOperatorTableCalculator(operatorSet)

// Get operator weights from table calculator (VIEW)
operatorWeights := operatorTableCalculator.GetOperatorWeights(operatorSet)

// Get supported chains for multichain (VIEW)
chainIds, tableUpdaterAddresses := crossChainRegistry.GetSupportedChains()

// Get table updater reference time and block (VIEW)
latestReferenceTimestamp := tableUpdater.GetLatestReferenceTimestamp()
latestReferenceBlockNumber := tableUpdater.GetReferenceBlockNumberByTimestamp(latestReferenceTimestamp)
```

#### 3.3 Operator Set Members and Keys (`GetOperatorSetMembersWithPeering`)

```go
// Get operator set members (VIEW)
operatorSet := allocationManager.GetMembers(operatorSet)

// For each operator, get their details:
for _, operatorAddress := range operatorSet {
    // Get operator socket information (VIEW)
    socket := avsRegistrar.GetOperatorSocket(operatorAddress)
    
    // Get curve type for operator set (VIEW)
    curveType := keyRegistrar.GetOperatorSetCurveType(operatorSet)
    
    // Get public keys based on curve type (VIEW)
    if curveType == BN254 {
        publicKey := keyRegistrar.GetBN254Key(operatorSet, operatorAddress)
    } else if curveType == ECDSA {
        ecdsaAddress := keyRegistrar.GetECDSAAddress(operatorSet, operatorAddress)
    }
}
```

### Phase 4: Signature Creation (Executor Side)

**Location**: `ponos/pkg/executor/handlers.go`

#### 4.1 Curve Type Determination

```go
// Get curve type for signing (VIEW)
curveType := e.l1ContractCaller.GetOperatorSetCurveType(task.Avs, task.OperatorSetId)
```

#### 4.2 ECDSA-Specific Digest Calculation

```go
// For ECDSA only, calculate certificate digest (VIEW)
if curveType == config.CurveTypeECDSA {
    digest := e.l1ContractCaller.CalculateECDSACertificateDigest(
        context.Background(),
        task.ReferenceTimestamp,
        digestBytes,
    )
}
```

### Phase 5: Certificate Submission (Aggregator Side)

**Location**: `ponos/pkg/contractCaller/caller/caller.go`

#### 5.1 BN254 Certificate Submission (`SubmitBN254TaskResult`)

```go
// Get certificate bytes for validation (VIEW)
certBytes := taskMailbox.GetBN254CertificateBytes(certificate)

// Submit the result with certificate (STATE CHANGE)
tx := taskMailbox.SubmitResult(taskId, certBytes, taskResponse)
```

#### 5.2 ECDSA Certificate Submission (`SubmitECDSATaskResult`)

```go
// Get certificate bytes for validation (VIEW)
certBytes := taskMailbox.GetECDSACertificateBytes(certificate)

// Submit the result with certificate (STATE CHANGE)
tx := taskMailbox.SubmitResult(taskId, certBytes, taskResponse)
```

#### 5.3 Optional Certificate Verification (`VerifyECDSACertificate`)

```go
// Verify ECDSA certificate (STATE CHANGE)
tx := ecdsaCertVerifier.VerifyCertificateProportion(
    operatorSet,
    certificate,
    thresholds
)
```

## Complete Flow Sequence

### Timeline View

```
Phase 1: Operator Setup (One-time per operator)
├── delegationManager.IsOperator() [VIEW]
├── delegationManager.RegisterAsOperator() [STATE CHANGE]
├── allocationManager.RegisterForOperatorSets() [STATE CHANGE]
├── keyRegistrar.ConfigureOperatorSet() [STATE CHANGE]
├── keyRegistrar.Get{BN254|ECDSA}KeyRegistrationMessageHash() [VIEW]
├── keyRegistrar.RegisterKey() [STATE CHANGE]
├── delegationManager.DelegateTo() [STATE CHANGE]
└── allocationManager.ModifyAllocations() [STATE CHANGE]

Phase 2: Task Creation (Per task)
└── taskMailbox.CreateTask() [STATE CHANGE] → Emits TaskCreated event

Phase 3: Task Processing (Per task - triggered by TaskCreated event)
├── allocationManager.GetAVSRegistrar() [VIEW]
├── avsRegistrar.GetAvsConfig() [VIEW]
├── crossChainRegistry.GetOperatorTableCalculator() [VIEW]
├── operatorTableCalculator.GetOperatorWeights() [VIEW]
├── crossChainRegistry.GetSupportedChains() [VIEW]
├── tableUpdater.GetLatestReferenceTimestamp() [VIEW]
├── tableUpdater.GetReferenceBlockNumberByTimestamp() [VIEW]
├── allocationManager.GetMembers() [VIEW]
├── keyRegistrar.GetOperatorSetCurveType() [VIEW]
├── keyRegistrar.Get{BN254Key|ECDSAAddress}() [VIEW] (per operator)
└── avsRegistrar.GetOperatorSocket() [VIEW] (per operator)

Phase 4: Signature Creation (Executor side - per operator)
├── e.l1ContractCaller.GetOperatorSetCurveType() [VIEW]
└── e.l1ContractCaller.CalculateECDSACertificateDigest() [VIEW] (ECDSA only)

Phase 5: Certificate Submission (Aggregator side - after threshold met)
├── taskMailbox.Get{BN254|ECDSA}CertificateBytes() [VIEW]
├── taskMailbox.SubmitResult() [STATE CHANGE]
└── ecdsaCertVerifier.VerifyCertificateProportion() [STATE CHANGE] (optional)
```

## Key Characteristics

### Call Distribution
- **~90% View Functions**: Most calls are for data retrieval
- **~10% State Changes**: Primarily registration and final submission

### Retry Logic
- **Certificate Submission**: Has exponential backoff retry (1, 3, 5, 10, 20 seconds)
- **Data Retrieval**: No retry logic (assumed to be reliable)

### Cross-Chain Coordination
- **L1**: Provides operator data, stake weights, and key registrations
- **L2**: Can create tasks and submit results
- **Reference Timestamps**: Ensure consistent operator set snapshots across chains

### Curve Type Impact
- **BN254**: Requires G1/G2 point operations and aggregation
- **ECDSA**: Requires individual signature concatenation and address-based verification

### Performance Considerations
- **Batch Operations**: Multiple operators processed in parallel
- **Caching**: Operator data fetched once and reused
- **Block Number Pinning**: Ensures consistent state across multiple calls

### Error Handling
- **Contract Reverts**: Wrapped in appropriate error messages
- **Network Failures**: Retry logic for critical operations
- **Invalid Signatures**: Filtered out during aggregation phase

This sequencing ensures that the Hourglass system maintains consistency and security while efficiently coordinating multi-operator signature aggregation across potentially multiple chains.