// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {Test, console} from "forge-std/Test.sol";
import {TaskMailbox} from "src/core/TaskMailbox.sol";

contract TaskMailboxTest is Test {
    TaskMailbox public taskMailbox;

    function setUp() public {
        // Deploy the TaskMailbox contract
        taskMailbox = new TaskMailbox();
    }

    function testDummy() public pure returns (bool) {
        return true;
    }
}
