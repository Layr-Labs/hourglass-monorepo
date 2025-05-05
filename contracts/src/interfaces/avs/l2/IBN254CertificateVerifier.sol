// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {BN254} from "@eigenlayer-middleware/src/libraries/BN254.sol";

interface IBN254CertificateVerifier {
    struct BN254OperatorInfo {
        BN254.G1Point pubkey;
        uint96[] weights;
    }

    struct BN254OperatorInfoWitness {
        uint32 operatorIndex;
        // empty implies already cached in storage
        bytes operatorInfoProofs;
        BN254OperatorInfo operatorInfo;
    }

    struct BN254Certificate {
        uint32 referenceTimestamp;
        bytes32 messageHash; // It can be just the taskHash. Unless we retry..
        // signature data
        BN254.G1Point sig;
        BN254.G2Point apk;
        uint32[] nonsignerIndices;
        BN254OperatorInfoWitness[] nonSignerWitnesses;
    }

    /// @return the maximum amount of seconds that a operator table can be in the past
    function maxOperatorTableStaleness() external returns (uint32);

    /**
     * @notice verifies a certificate
     * @param cert a certificate
     * @return signedStakes the amount of stake that signed the certificate for each stake
     * type
     */
    function verifyCertificate(
        BN254Certificate memory cert
    ) external view returns (uint96[] memory signedStakes);

    /**
     * @notice verifies a certificate and makes sure that the signed stakes meet
     * provided portions of the total stake on the AVS
     * @param cert a certificate
     * @param totalStakeProportionThresholds the proportion of total stake that
     * the signed stake of the certificate should meet
     * @return whether or not certificate is valid and meets thresholds
     */
    function verifyCertificateProportion(
        BN254Certificate memory cert,
        uint16[] memory totalStakeProportionThresholds
    ) external view returns (bool);

    /**
     * @notice verifies a certificate and makes sure that the signed stakes meet
     * provided nominal stake thresholds
     * @param cert a certificate
     * @param totalStakeNominalThresholds the nominal amount of stake that
     * the signed stake of the certificate should meet
     * @return whether or not certificate is valid and meets thresholds
     */
    function verifyCertificateNominal(
        BN254Certificate memory cert,
        uint96[] memory totalStakeNominalThresholds
    ) external view returns (bool);
}
