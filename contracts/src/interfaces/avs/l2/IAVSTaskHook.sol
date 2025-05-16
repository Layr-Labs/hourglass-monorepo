// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

import {IBN254CertificateVerifier} from "./IBN254CertificateVerifier.sol";
import {ITaskMailboxTypes} from "../../core/ITaskMailbox.sol";

/**
 * @title IAVSTaskHook
 * @notice Interface for AVS-specific task validation and verification hooks
 * @dev This interface allows AVSs to implement custom validation logic for tasks
 */
interface IAVSTaskHook {
    // TODO: Should this contract be ERC165 compliant?

    /**
     * @notice Validates a task before it is created
     * @param caller Address that is creating the task
     * @param operatorSet The operator set that will execute the task
     * @param payload Task-specific data
     * @dev This function should revert if the task should not be created
     */
    function validatePreTaskCreation(
        address caller,
        OperatorSet memory operatorSet,
        bytes memory payload
    ) external view;

    /**
     * @notice Validates a task after it is created
     * @param taskHash Unique identifier of the task
     * @dev This function can be used to perform additional validation or update AVS-specific state
     */
    function validatePostTaskCreation(
        bytes32 taskHash
    ) external;

    /**
     * @notice Validates a task result submission
     * @param taskHash Unique identifier of the task
     * @param cert Certificate proving the validity of the result
     * @dev This function should revert if the result submission is invalid
     */
    function validateTaskResultSubmission(
        bytes32 taskHash,
        IBN254CertificateVerifier.BN254Certificate memory cert
    ) external;
}
