# Certificate Creation Process in Hourglass

This document explains the detailed process of how certificates are created for both BN254 and ECDSA signing schemes in the Hourglass system, focusing on the nuances and steps involved in aggregating individual operator signatures into final certificates that can be submitted on-chain.

## Overview

The certificate creation process is the core of how Hourglass achieves consensus among multiple operators. When a task is distributed to operators, each operator independently computes the result and signs it. The aggregator then collects these signatures and combines them into a single certificate that proves the required threshold of operators agreed on the result.

## Common Foundation

Both BN254 and ECDSA certificate creation share several fundamental concepts:

### Task Result Aggregation
Every certificate creation begins with a **Task Result Aggregator** that manages the collection and verification of individual operator signatures. The aggregator maintains:
- A list of all operators authorized to participate in the task
- A threshold percentage that must be met for the task to be considered complete
- A collection of received signatures from operators
- Logic to determine when enough signatures have been collected

### Threshold Logic
The system uses a percentage-based threshold mechanism. For example, with a 75% threshold and 4 operators, at least 3 operators must provide valid signatures. The threshold calculation always rounds up and ensures at least one signature is required, even for very small operator sets.

### Signature Verification
Before any signature is accepted into the aggregation process, it must be verified against the operator's public key. This ensures that:
- The signature was actually created by the claimed operator
- The signature is valid for the specific task result
- The operator is authorized to participate in this task

## BN254 Certificate Creation Process

BN254 uses **Boneh-Lynn-Shacham (BLS) signatures**, which have a unique mathematical property: multiple signatures can be combined into a single aggregate signature that is verifiable against an aggregate public key.

### Step 1: Individual Signature Processing

When a BN254 signature arrives from an operator:

1. **Operator Validation**: The system first checks that the operator is in the authorized set for this task
2. **Signature Parsing**: The raw signature bytes are converted into a BN254 signature object
3. **Signature Verification**: The signature is verified against the operator's BN254 public key and the task result digest
4. **Duplicate Check**: The system ensures this operator hasn't already submitted a signature for this task

### Step 2: Signature Aggregation

BN254's mathematical properties allow for true signature aggregation:

1. **First Signature**: When the first valid signature arrives, it initializes the aggregation structure with:
   - The signature as the starting point for the aggregate signature
   - The operator's public key as the starting point for the aggregate public key
   - A count of signers (starting at 1)

2. **Subsequent Signatures**: Each additional signature is mathematically combined:
   - The new signature is added to the existing aggregate signature using elliptic curve addition
   - The operator's public key is added to the existing aggregate public key
   - The signer count is incremented

3. **Threshold Checking**: After each signature, the system checks if the threshold has been met

### Step 3: Final Certificate Generation

When the threshold is reached:

1. **Signer Identification**: The system identifies which operators signed and which didn't
2. **Non-Signer Collection**: Operators who didn't sign are collected into a non-signers list
3. **Certificate Assembly**: The final certificate contains:
   - The task ID and response payload
   - The aggregate signature (a single BN254 signature representing all signers)
   - The aggregate public key (a single BN254 public key representing all signers)
   - Lists of all operators and non-signers for verification purposes

### BN254 Certificate Verification

The beauty of BN254 is that the entire certificate can be verified with a single signature verification operation against the aggregate public key. This is highly efficient for on-chain verification.

## ECDSA Certificate Creation Process

ECDSA certificates work differently because ECDSA signatures cannot be mathematically aggregated. Instead, the system collects individual signatures and presents them as a bundle.

### Step 1: Individual Signature Processing

The initial processing is similar to BN254:

1. **Operator Validation**: Verify the operator is authorized
2. **Signature Parsing**: Convert signature bytes to ECDSA signature format
3. **Signature Verification**: Verify the signature against the operator's Ethereum address
4. **Duplicate Check**: Ensure no duplicate submissions

### Step 2: Signature Collection

Instead of aggregation, ECDSA uses collection:

1. **First Signature**: Initialize a map storing operator address → signature bytes
2. **Subsequent Signatures**: Add each new signature to the map
3. **Public Key Tracking**: Maintain a separate list of signer addresses
4. **Threshold Checking**: Count signatures to determine when threshold is met

### Step 3: Final Certificate Generation

When threshold is reached:

1. **Signer Sorting**: All signer addresses are sorted deterministically to ensure consistent certificate format
2. **Non-Signer Identification**: Operators who didn't sign are identified
3. **Signature Concatenation**: All individual signatures are concatenated into a single byte array
4. **Certificate Assembly**: The final certificate contains:
   - The task ID and response payload
   - A map of individual signatures keyed by operator address
   - Sorted lists of signer addresses and non-signer addresses
   - The concatenated signature bytes for on-chain submission

### ECDSA Certificate Verification

ECDSA certificates require individual verification of each signature. The on-chain verifier must:
- Split the concatenated signature bytes back into individual signatures
- Verify each signature against its corresponding operator's address
- Ensure enough signatures are present to meet the threshold

## Key Differences and Trade-offs

### Efficiency
- **BN254**: Single signature verification, smaller on-chain footprint
- **ECDSA**: Multiple signature verifications, larger on-chain footprint

### Complexity
- **BN254**: More complex cryptographic operations but simpler verification
- **ECDSA**: Simpler individual operations but more complex aggregation logic

### Compatibility
- **BN254**: Requires specialized cryptographic libraries and knowledge
- **ECDSA**: Uses standard Ethereum-compatible signatures

### Security Model
- **BN254**: Relies on the discrete logarithm problem in elliptic curve groups
- **ECDSA**: Uses the well-established elliptic curve discrete logarithm problem

## Certificate Validation and Consensus

Both schemes include important validation mechanisms:

### Consensus on Results
The system currently uses a "last writer wins" approach where the final certificate uses the task result from the last operator to meet the threshold. This is noted as a potential improvement area, as it doesn't verify that all operators computed the same result.

### Timestamp Tracking
Certificates include timestamp information to track when signatures were collected and aggregated.

### Reference Block Coordination
The system uses reference timestamps and block numbers to ensure all operators are working with the same view of the blockchain state, which is crucial for consistent operator set membership.

## Error Handling and Edge Cases

### Insufficient Signatures
If not enough operators respond before the deadline, the task fails and no certificate is generated.

### Malformed Signatures
Invalid signatures are rejected during the verification step and don't contribute to the threshold.

### Duplicate Submissions
Operators can only submit one signature per task. Additional submissions are rejected.

### Threshold Boundary Cases
The threshold calculation always requires at least one signature, even for very small operator sets or low percentages.

## On-Chain Submission Format

The final certificates are formatted differently for each scheme:

### BN254 On-Chain Format
- Single aggregate signature in G1 point format
- Single aggregate public key in G2 point format
- Reference timestamp for operator set validity
- Task result hash
- Non-signer witness data (currently empty)

### ECDSA On-Chain Format
- Concatenated signature bytes from all signers
- Reference timestamp for operator set validity
- Task result hash
- The certificate verifier handles parsing individual signatures

## Performance Considerations

### BN254 Performance
- Signature aggregation is computationally intensive
- Verification is very fast (single operation)
- Certificate size is constant regardless of number of signers

### ECDSA Performance
- Signature collection is simple
- Verification time scales linearly with number of signers
- Certificate size grows with number of signers

This process ensures that the Hourglass system can reliably collect and verify operator consensus while supporting both modern BLS signatures and traditional ECDSA signatures based on the requirements of different use cases.