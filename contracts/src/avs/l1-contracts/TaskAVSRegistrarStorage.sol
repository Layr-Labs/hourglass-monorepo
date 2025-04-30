// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {ITaskAVSRegistrar} from "src/interfaces/ITaskAVSRegistrar.sol";

abstract contract TaskAVSRegistrarStorage is ITaskAVSRegistrar {
    /// @notice The avs address for this AVS (used for UAM integration in EigenLayer)
    /// @dev NOTE: Updating this value will break existing OperatorSets and UAM integration.
    /// This value should only be set once.
    address public immutable AVS;

    constructor(
        address avs
    ) {
        AVS = avs;
    }
}
