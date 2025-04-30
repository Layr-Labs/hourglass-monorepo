// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {IAVSRegistrar} from "@eigenlayer-middleware/lib/eigenlayer-contracts/src/contracts/interfaces/IAVSRegistrar.sol";

interface ITaskAVSRegistrarTypes {}

interface ITaskAVSRegistrarErrors is ITaskAVSRegistrarTypes {}

interface ITaskAVSRegistrarEvents is ITaskAVSRegistrarTypes {}

interface ITaskAVSRegistrar is ITaskAVSRegistrarErrors, ITaskAVSRegistrarEvents, IAVSRegistrar {}
