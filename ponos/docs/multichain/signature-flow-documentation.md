# Signature and Certificate Creation Flow for BN254 and ECDSA

This document describes the complete flow of how signatures and certificates are created for both BN254 and ECDSA signing schemes in the Hourglass aggregator and executor system.

## Overview

The Hourglass system uses two different cryptographic schemes for signing:
- **BN254**: Uses BLS (Boneh-Lynn-Shacham) signatures with elliptic curve operations
- **ECDSA**: Uses traditional ECDSA signatures with Ethereum-style addresses

The flow involves four main phases:
1. **Stake Weight Retrieval**: Getting operator stake weights from contracts
2. **Task Distribution**: Broadcasting tasks to executors
3. **Individual Signing**: Executors signing their work and returning results
4. **Signature Aggregation**: Combining signatures into final certificates

## 1. Getting Stake Weights for Operators

### Location: `ponos/pkg/contractCaller/caller/caller.go`

The aggregator retrieves operator stake weights through the `GetOperatorTableDataForOperatorSet` method:

```go
func (cc *ContractCaller) GetOperatorTableDataForOperatorSet(
    ctx context.Context,
    avsAddress common.Address,
    operatorSetId uint32,
    chainId config.ChainId,
    atBlockNumber uint64,
) (*contractCaller.OperatorTableData, error)
```

**Process:**
1. Query the `CrossChainRegistry` to get the `OperatorTableCalculator` address
2. Call `GetOperatorWeights` to retrieve stake weights for all operators in the set
3. Fetch supported chains and table updater addresses
4. Get the latest reference timestamp and block number

**Key Data Retrieved:**
- `OperatorWeights`: Stake weights for each operator
- `Operators`: List of operator addresses
- `LatestReferenceTimestamp`: Reference time for the operator set
- `LatestReferenceBlockNumber`: Block number for the reference

## 2. Distributing Tasks to Executors

### Location: `ponos/pkg/taskSession/taskSession.go`

The task distribution happens in the `Broadcast` method of `TaskSession`:

**Process:**
1. **Task Session Creation**: Create either BN254 or ECDSA task session based on curve type
   - `NewBN254TaskSession`: For BN254 signing (`taskSession.go:38-87`)
   - `NewECDSATaskSession`: For ECDSA signing (`taskSession.go:90-139`)

2. **Operator Set Preparation**: Extract operator info from `operatorPeersWeight`
   ```go
   // BN254 operators
   operators = append(operators, &aggregation.Operator[signing.PublicKey]{
       Address:   peer.OperatorAddress,
       PublicKey: opset.WrappedPublicKey.PublicKey,
   })
   
   // ECDSA operators  
   operators = append(operators, &aggregation.Operator[common.Address]{
       Address:   peer.OperatorAddress,
       PublicKey: opset.WrappedPublicKey.ECDSAAddress,
   })
   ```

3. **Task Broadcasting**: Send tasks to all operators in parallel (`taskSession.go:199-241`)
   ```go
   taskSubmission := &executorV1.TaskSubmission{
       TaskId:             ts.Task.TaskId,
       AvsAddress:         ts.Task.AVSAddress,
       AggregatorAddress:  ts.aggregatorAddress,
       Payload:            ts.Task.Payload,
       Signature:          ts.aggregatorSignature,
       OperatorSetId:      ts.Task.OperatorSetId,
       ReferenceTimestamp: ts.operatorPeersWeight.RootReferenceTimestamp,
   }
   ```

## 3. Executors Signing Their Work

### Location: `ponos/pkg/executor/handlers.go`

The executor processes tasks in the `handleReceivedTask` method:

**Process:**
1. **Task Validation**: Verify task signature using `ValidateTaskSignature`
2. **Task Execution**: Run the AVS-specific workload via `avsPerf.RunTask`
3. **Result Signing**: Sign the task result using `signResult` method

### Signing Process (`signResult` method - `handlers.go:238-277`)

**Key Differences Between BN254 and ECDSA:**

#### Common Steps:
1. Generate keccak256 hash of the result: `digestBytes = util.GetKeccak256Digest(result.Result)`
2. Determine curve type from contract: `curveType, err := e.l1ContractCaller.GetOperatorSetCurveType(task.Avs, task.OperatorSetId)`

#### ECDSA-Specific Steps:
```go
if curveType == config.CurveTypeECDSA {
    digest, err := e.l1ContractCaller.CalculateECDSACertificateDigest(
        context.Background(),
        task.ReferenceTimestamp,
        digestBytes,
    )
    if err != nil {
        return nil, digestBytes, fmt.Errorf("failed to calculate ECDSA certificate digest: %w", err)
    }
    digestBytes = digest
}
```

#### Final Signing:
```go
sig, err := e.signer.SignMessageForSolidity(digestBytes)
```

**Return Format:**
```go
return &executorV1.TaskResult{
    TaskId:          response.TaskID,
    OperatorAddress: e.config.Operator.Address,
    Output:          response.Result,
    Signature:       sig,
    AvsAddress:      task.AvsAddress,
    OutputDigest:    digest[:],
}, nil
```

## 4. Aggregator Combining Signatures

### Location: `ponos/pkg/signing/aggregation/`

The aggregator combines signatures using different strategies for BN254 and ECDSA.

### BN254 Aggregation (`bn254.go`)

**Key Components:**
- `BN254TaskResultAggregator`: Main aggregation logic
- `AggregatedBN254Certificate`: Final certificate structure

**Process:**
1. **Signature Processing** (`ProcessNewSignature` - `bn254.go:145-228`):
   - Validate operator is in allowed set
   - Verify signature using `VerifyResponseSignature`
   - Add to aggregated data structure

2. **Signature Verification** (`VerifyResponseSignature` - `bn254.go:232-247`):
   ```go
   if verified, err := sig.VerifySolidityCompatible(operator.PublicKey.(*bn254.PublicKey), digest); err != nil {
       return nil, fmt.Errorf("signature verification failed: %w", err)
   }
   ```

3. **Aggregation Logic** (`bn254.go:201-225`):
   ```go
   if tra.aggregatedOperators == nil {
       // First signature
       tra.aggregatedOperators = &aggregatedBN254Operators{
           signersG2: bn254.NewZeroG2Point().AddPublicKey(bn254PubKey),
           signersAggSig: sig,
           signersOperatorSet: map[string]bool{taskResponse.OperatorAddress: true},
           totalSigners: 1,
       }
   } else {
       // Subsequent signatures
       tra.aggregatedOperators.signersG2.AddPublicKey(bn254PubKey)
       tra.aggregatedOperators.signersAggSig.Add(sig)
       tra.aggregatedOperators.totalSigners++
   }
   ```

4. **Final Certificate Generation** (`GenerateFinalCertificate` - `bn254.go:250-294`):
   ```go
   return &AggregatedBN254Certificate{
       TaskId:              taskIdBytes,
       TaskResponse:        tra.aggregatedOperators.lastReceivedResponse.TaskResult.Output,
       TaskResponseDigest:  tra.aggregatedOperators.lastReceivedResponse.Digest,
       NonSignersPubKeys:   nonSignerPublicKeys,
       AllOperatorsPubKeys: allPublicKeys,
       SignersPublicKey:    tra.aggregatedOperators.signersG2,    // Aggregated G2 point
       SignersSignature:    tra.aggregatedOperators.signersAggSig, // Aggregated signature
       SignedAt:            new(time.Time),
   }
   ```

### ECDSA Aggregation (`ecdsa.go`)

**Key Components:**
- `ECDSATaskResultAggregator`: Main aggregation logic
- `AggregatedECDSACertificate`: Final certificate structure

**Process:**
1. **Signature Processing** (`ProcessNewSignature` - `ecdsa.go:152-233`):
   - Similar validation as BN254
   - Store individual signatures in map instead of aggregating

2. **Signature Verification** (`VerifyResponseSignature` - `ecdsa.go:247-259`):
   ```go
   if verified, err := sig.VerifyWithAddress(taskResponse.OutputDigest, operator.PublicKey); err != nil {
       return nil, fmt.Errorf("signature verification failed: %w", err)
   }
   ```

3. **"Aggregation" Logic** (`ecdsa.go:206-230`):
   ```go
   if tra.aggregatedOperators == nil {
       tra.aggregatedOperators = &aggregatedECDSAOperators{
           signersPublicKeys: []common.Address{operator.PublicKey},
           signersSignatures: map[common.Address][]byte{
               operator.GetAddress(): taskResponse.Signature,
           },
           totalSigners: 1,
       }
   } else {
       // Add individual signatures to map
       tra.aggregatedOperators.signersSignatures[operator.GetAddress()] = taskResponse.Signature
       tra.aggregatedOperators.signersPublicKeys = append(tra.aggregatedOperators.signersPublicKeys, operator.PublicKey)
       tra.aggregatedOperators.totalSigners++
   }
   ```

4. **Final Certificate Generation** (`GenerateFinalCertificate` - `ecdsa.go:261-316`):
   ```go
   return &AggregatedECDSACertificate{
       TaskId:              taskIdBytes,
       TaskResponse:        tra.aggregatedOperators.lastReceivedResponse.TaskResult.Output,
       TaskResponseDigest:  tra.aggregatedOperators.lastReceivedResponse.Digest,
       NonSignersPubKeys:   nonSignerPublicKeys,
       AllOperatorsPubKeys: allPublicKeys,
       SignersPublicKeys:   tra.aggregatedOperators.signersPublicKeys,   // Array of addresses
       SignersSignatures:   tra.aggregatedOperators.signersSignatures,  // Map of individual sigs
       SignedAt:            new(time.Time),
   }
   ```

## 5. Submitting Certificates On-Chain

### Location: `ponos/pkg/contractCaller/caller/caller.go`

The final step is submitting the aggregated certificate back to the blockchain.

### BN254 Submission (`SubmitBN254TaskResult` - `caller.go:170-240`):
```go
// Convert signature to G1 point in precompile format
g1Point := &bn254.G1Point{
    G1Affine: aggCert.SignersSignature.GetG1Point(),
}
g1Bytes, err := g1Point.ToPrecompileFormat()

// Convert public key to G2 point in precompile format  
g2Bytes, err := aggCert.SignersPublicKey.ToPrecompileFormat()

cert := ITaskMailbox.IBN254CertificateVerifierTypesBN254Certificate{
    ReferenceTimestamp: globalTableRootReferenceTimestamp,
    MessageHash:        digest,
    Signature:          ITaskMailbox.BN254G1Point{...},
    Apk:                ITaskMailbox.BN254G2Point{...},
    NonSignerWitnesses: []ITaskMailbox.IBN254CertificateVerifierTypesBN254OperatorInfoWitness{},
}
```

### ECDSA Submission (`SubmitECDSATaskResult` - `caller.go:271-316`):
```go
finalSig, err := aggCert.GetFinalSignature() // Concatenates all individual signatures

cert := ITaskMailbox.IECDSACertificateVerifierTypesECDSACertificate{
    ReferenceTimestamp: globalTableRootReferenceTimestamp,
    MessageHash:        aggCert.GetTaskMessageHash(),
    Sig:                finalSig,
}
```

## Key Differences Summary

| Aspect | BN254 | ECDSA |
|--------|-------|-------|
| **Signature Aggregation** | True mathematical aggregation using elliptic curve operations | Individual signatures concatenated together |
| **Public Key Handling** | Aggregated into single G2 point | Array of individual addresses |
| **Digest Calculation** | Direct keccak256 of result | Requires additional `CalculateECDSACertificateDigest` call |
| **Verification** | Single aggregate signature verification | Multiple individual signature verifications |
| **On-chain Storage** | More efficient (single signature + pubkey) | Less efficient (multiple signatures) |
| **Certificate Structure** | Contains aggregated signature and pubkey | Contains map of individual signatures |

## Threshold Logic

Both schemes use the same threshold logic in their respective aggregators:
```go
func (tra *TaskResultAggregator) SigningThresholdMet() bool {
    required := int((float64(tra.ThresholdPercentage) / 100.0) * float64(len(tra.Operators)))
    if required == 0 {
        required = 1 // Always require at least one
    }
    return tra.aggregatedOperators.totalSigners >= required
}
```

The threshold is met when enough operators have submitted valid signatures, triggering the final certificate generation and on-chain submission.