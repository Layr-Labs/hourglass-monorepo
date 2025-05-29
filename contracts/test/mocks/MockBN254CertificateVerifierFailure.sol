// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IBN254CertificateVerifier} from "../../src/interfaces/avs/l2/IBN254CertificateVerifier.sol";

/**
 * @title MockBN254CertificateVerifierFailure
 * @notice Mock BN254 certificate verifier that always fails verification
 * @dev Used for testing certificate verification failure scenarios
 */
contract MockBN254CertificateVerifierFailure is IBN254CertificateVerifier {
    function maxOperatorTableStaleness() external pure returns (uint32) {
        return 86_400;
    }

    function verifyCertificate(
        BN254Certificate memory /*cert*/
    ) external pure returns (uint96[] memory signedStakes) {
        return new uint96[](0);
    }

    function verifyCertificateProportion(
        BN254Certificate memory, /*cert*/
        uint16[] memory /*totalStakeProportionThresholds*/
    ) external pure returns (bool) {
        return false; // Always fail
    }

    function verifyCertificateNominal(
        BN254Certificate memory, /*cert*/
        uint96[] memory /*totalStakeNominalThresholds*/
    ) external pure returns (bool) {
        return false; // Always fail
    }
}
