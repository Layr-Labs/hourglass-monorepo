// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {OperatorSet} from "@eigenlayer-contracts/src/contracts/libraries/OperatorSetLib.sol";

import {IAVSTaskHook} from "../../src/interfaces/avs/l2/IAVSTaskHook.sol";
import {IECDSACertificateVerifier} from "../../src/interfaces/avs/l2/IECDSACertificateVerifier.sol";

contract MockAVSTaskHook is IAVSTaskHook {
    function validatePreTaskCreation(
        address, /*caller*/
        OperatorSet memory, /*operatorSet*/
        bytes memory /*payload*/
    ) external view {
        //TODO: Implement
    }

    function validatePostTaskCreation(
        bytes32 /*taskHash*/
    ) external {
        //TODO: Implement
    }

    function validateTaskResultSubmission(
        bytes32, /*taskHash*/
        IECDSACertificateVerifier.ECDSACertificate memory /*cert*/
    ) external {
        //TODO: Implement
    }
}
