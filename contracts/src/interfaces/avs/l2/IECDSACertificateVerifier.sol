// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {OperatorSet} from "eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

/**
 * @title IECDSACertificateVerifier
 * @notice Interface for the ECDSA certificate verifier
 */
interface IECDSACertificateVerifier {
    /**
     * @notice Structure for an ECDSA certificate
     * @param referenceTimestamp The timestamp identifying the operator table to verify against
     * @param messageHash The hash of the message which has been signed by operators
     * @param sig The concatenated ECDSA signatures of the signing operators (each 65 bytes)
     */
    struct ECDSACertificate {
        uint32 referenceTimestamp;
        bytes32 messageHash;
        // The concatenated signature of each signing operator
        bytes sig;
    }

    /**
     * @notice Returns the maximum staleness allowed for an operator table
     * @return The maximum staleness in seconds
     */
    function maxOperatorTableStaleness() external view returns (uint32);

    /**
     * @notice Verifies a certificate
     * @param cert A certificate
     * @return signedStakes The amount of stake that signed the certificate for each stake type
     */
    function verifyCertificate(
        ECDSACertificate memory cert
    ) external view returns (uint96[] memory signedStakes);

    /**
     * @notice Verifies a certificate and makes sure that the signed stakes meet provided portions of the total stake
     * @param cert A certificate
     * @param totalStakeProportionThresholds The proportion of total stake that the signed stake should meet (in basis points)
     * @return Whether or not certificate is valid and meets thresholds
     */
    function verifyCertificateProportion(
        ECDSACertificate memory cert,
        uint16[] memory totalStakeProportionThresholds
    ) external view returns (bool);

    /**
     * @notice Verifies a certificate and makes sure that the signed stakes meet provided nominal stake thresholds
     * @param cert A certificate
     * @param totalStakeNominalThresholds The nominal amount of stake that the signed stake should meet
     * @return Whether or not certificate is valid and meets thresholds
     */
    function verifyCertificateNominal(
        ECDSACertificate memory cert,
        uint96[] memory totalStakeNominalThresholds
    ) external view returns (bool);
}
