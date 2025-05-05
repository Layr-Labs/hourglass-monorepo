// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

import {IBN254CertificateVerifier} from "./IBN254CertificateVerifier.sol";
import {ITaskMailboxTypes} from "../../core/ITaskMailbox.sol";

interface IAVSTaskHook {
    // TODO: Should this contract be ERC165 compliant?
    function validatePreTaskCreation(
        address caller,
        OperatorSet memory operatorSet,
        bytes memory payload
    ) external view;

    function validatePostTaskCreation(
        bytes32 taskHash
    ) external;

    function validateTaskResultSubmission(
        bytes32 taskHash,
        IBN254CertificateVerifier.BN254Certificate memory cert
    ) external;
}
