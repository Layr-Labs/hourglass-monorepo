// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {
    IBN254CertificateVerifier,
    IBN254CertificateVerifierTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

contract MockBN254CertificateVerifier is IBN254CertificateVerifier {
    function updateOperatorTable(
        OperatorSet calldata, /*operatorSet*/
        uint32, /*referenceTimestamp*/
        BN254OperatorSetInfo memory, /*operatorSetInfo*/
        OperatorSetConfig calldata /*operatorSetConfig*/
    ) external pure {}

    function verifyCertificate(
        OperatorSet memory, /*operatorSet*/
        BN254Certificate memory /*cert*/
    ) external pure returns (uint256[] memory signedStakes) {
        return new uint256[](0);
    }

    function verifyCertificateProportion(
        OperatorSet memory, /*operatorSet*/
        BN254Certificate memory, /*cert*/
        uint16[] memory /*totalStakeProportionThresholds*/
    ) external pure returns (bool) {
        return true;
    }

    function verifyCertificateNominal(
        OperatorSet memory, /*operatorSet*/
        BN254Certificate memory, /*cert*/
        uint256[] memory /*totalStakeNominalThresholds*/
    ) external pure returns (bool) {
        return true;
    }

    // Implement IBaseCertificateVerifier required functions
    function operatorTableUpdater(
        OperatorSet memory /*operatorSet*/
    ) external pure returns (address) {
        return address(0);
    }

    function getLatestReferenceTimestamp(
        OperatorSet memory /*operatorSet*/
    ) external pure returns (uint32) {
        return 0;
    }

    function getOperatorSetOwner(
        OperatorSet memory /*operatorSet*/
    ) external pure returns (address) {
        return address(0);
    }

    function latestReferenceTimestamp(
        OperatorSet memory /*operatorSet*/
    ) external pure returns (uint32) {
        return 0;
    }

    function maxOperatorTableStaleness(
        OperatorSet memory /*operatorSet*/
    ) external pure returns (uint32) {
        return 86_400;
    }
}
