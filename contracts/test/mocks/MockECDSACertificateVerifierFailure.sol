// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IECDSACertificateVerifier} from "../../src/interfaces/avs/l2/IECDSACertificateVerifier.sol";

/**
 * @title MockECDSACertificateVerifierFailure
 * @notice Mock ECDSA certificate verifier that always fails verification
 * @dev Used for testing certificate verification failure scenarios
 */
contract MockECDSACertificateVerifierFailure is IECDSACertificateVerifier {
    function maxOperatorTableStaleness() external pure returns (uint32) {
        return 86_400;
    }

    function verifyCertificate(
        IECDSACertificateVerifier.ECDSACertificate memory /*cert*/
    ) external pure returns (uint96[] memory signedStakes) {
        return new uint96[](0);
    }

    function verifyCertificateProportion(
        IECDSACertificateVerifier.ECDSACertificate memory, /*cert*/
        uint16[] memory /*totalStakeProportionThresholds*/
    ) external pure returns (bool) {
        return false; // Always fail
    }

    function verifyCertificateNominal(
        IECDSACertificateVerifier.ECDSACertificate memory, /*cert*/
        uint96[] memory /*totalStakeNominalThresholds*/
    ) external pure returns (bool) {
        return false; // Always fail
    }
}
