// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {TaskAVSRegistrarStorage} from "src/avs/l1-contracts/TaskAVSRegistrarStorage.sol";

contract TaskAVSRegistrar is TaskAVSRegistrarStorage {
    constructor(
        address avs
    ) TaskAVSRegistrarStorage(avs) {}

    function registerOperator(
        address operator,
        address avs,
        uint32[] calldata operatorSetIds,
        bytes calldata data
    ) external {
        // TODO: Implement
    }

    function deregisterOperator(address operator, address avs, uint32[] calldata operatorSetIds) external {
        // TODO: Implement
    }

    function supportsAVS(
        address avs
    ) external view returns (bool) {
        return avs == AVS;
    }
}
