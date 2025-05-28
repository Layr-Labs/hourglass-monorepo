// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {BN254} from "@eigenlayer-middleware/src/libraries/BN254.sol";

/**
 * @title IBN254CertificateVerifier
 * @author Layr Labs, Inc.
 * @notice Interface for verifying BN254 certificates with operator proofs
 */
interface IBN254CertificateVerifier {
    /**
     * @notice Information about an operator's BLS public key and stake weights
     * @param pubkey The operator's BLS public key in G1 group format
     * @param weights The operator's stake weights for different stake types
     */
    struct BN254OperatorInfo {
        BN254.G1Point pubkey;
        uint96[] weights;
    }

    /**
     * @notice Witness data for an operator when proving non-participation
     * @param operatorIndex The index of the operator in the operator set
     * @param operatorInfoProofs Merkle proofs for the operator's information (empty if already cached)
     * @param operatorInfo The operator's BLS public key and stake weights
     */
    struct BN254OperatorInfoWitness {
        uint32 operatorIndex;
        // empty implies already cached in storage
        bytes operatorInfoProofs;
        BN254OperatorInfo operatorInfo;
    }

    /**
     * @notice A certificate proving signature by a threshold of operators
     * @param referenceTimestamp The timestamp used for the operator table reference
     * @param messageHash The hash of the message that was signed (typically a taskHash)
     * @param sig The aggregated BLS signature
     * @param apk The aggregate public key of the operators that signed
     * @param nonsignerIndices The indices of operators that did not sign
     * @param nonSignerWitnesses Witness data for the operators that did not sign
     */
    struct BN254Certificate {
        uint32 referenceTimestamp;
        bytes32 messageHash; // It can be just the taskHash. Unless we retry..
        // signature data
        BN254.G1Point sig;
        BN254.G2Point apk;
        uint32[] nonsignerIndices;
        BN254OperatorInfoWitness[] nonSignerWitnesses;
    }

    /**
     * @notice Returns the maximum staleness allowed for operator tables
     * @return The maximum amount of seconds that an operator table can be in the past
     */
    function maxOperatorTableStaleness() external returns (uint32);

    /**
     * @notice Verifies a certificate and returns the signed stakes
     * @param cert The certificate to verify
     * @return signedStakes The amount of stake that signed the certificate for each stake type
     * @dev This function verifies the certificate's signature and returns the stake that participated
     */
    function verifyCertificate(
        BN254Certificate memory cert
    ) external view returns (uint96[] memory signedStakes);

    /**
     * @notice Verifies a certificate against proportion thresholds
     * @param cert The certificate to verify
     * @param totalStakeProportionThresholds The proportion thresholds that signed stake must meet (in basis points)
     * @return Whether the certificate is valid and meets the proportion thresholds
     * @dev This function checks if the signature is valid and if enough stake has signed as a proportion of total stake
     */
    function verifyCertificateProportion(
        BN254Certificate memory cert,
        uint16[] memory totalStakeProportionThresholds
    ) external view returns (bool);

    /**
     * @notice Verifies a certificate against nominal thresholds
     * @param cert The certificate to verify
     * @param totalStakeNominalThresholds The nominal stake thresholds that signed stake must meet
     * @return Whether the certificate is valid and meets the nominal thresholds
     * @dev This function checks if the signature is valid and if enough absolute stake amount has signed
     */
    function verifyCertificateNominal(
        BN254Certificate memory cert,
        uint96[] memory totalStakeNominalThresholds
    ) external view returns (bool);
}
