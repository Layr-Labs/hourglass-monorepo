// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {
    IBN254CertificateVerifier,
    IBN254CertificateVerifierTypes
} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254CertificateVerifier.sol";
import {IBN254TableCalculatorTypes} from "@eigenlayer-contracts/src/contracts/interfaces/IBN254TableCalculator.sol";
import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";
import {BN254} from "@eigenlayer-contracts/src/contracts/libraries/BN254.sol";

contract MockBN254CertificateVerifier is IBN254CertificateVerifier {
    function updateOperatorTable(
        OperatorSet calldata, /*operatorSet*/
        uint32, /*referenceTimestamp*/
        IBN254TableCalculatorTypes.BN254OperatorSetInfo memory, /*operatorSetInfo*/
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

    function trySignatureVerification(
        bytes32, /*msgHash*/
        BN254.G1Point memory, /*aggPubkey*/
        BN254.G2Point memory, /*apkG2*/
        BN254.G1Point memory /*signature*/
    ) external pure returns (bool pairingSuccessful, bool signatureValid) {
        return (true, true);
    }

    function getNonsignerOperatorInfo(
        OperatorSet memory, /*operatorSet*/
        uint32, /*referenceTimestamp*/
        uint256 /*operatorIndex*/
    ) external pure returns (IBN254TableCalculatorTypes.BN254OperatorInfo memory) {
        uint256[] memory weights = new uint256[](0);
        return IBN254TableCalculatorTypes.BN254OperatorInfo({pubkey: BN254.G1Point(0, 0), weights: weights});
    }

    function isNonsignerCached(
        OperatorSet memory, /*operatorSet*/
        uint32, /*referenceTimestamp*/
        uint256 /*operatorIndex*/
    ) external pure returns (bool) {
        return false;
    }

    function getOperatorSetInfo(
        OperatorSet memory, /*operatorSet*/
        uint32 /*referenceTimestamp*/
    ) external pure returns (IBN254TableCalculatorTypes.BN254OperatorSetInfo memory) {
        uint256[] memory totalWeights = new uint256[](0);
        return IBN254TableCalculatorTypes.BN254OperatorSetInfo({
            operatorInfoTreeRoot: bytes32(0),
            numOperators: 0,
            aggregatePubkey: BN254.G1Point(0, 0),
            totalWeights: totalWeights
        });
    }
}
