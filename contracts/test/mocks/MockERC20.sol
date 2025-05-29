// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.27;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MockERC20
 * @notice Mock ERC20 token for testing
 * @dev Simple ERC20 implementation with mint function for testing
 */
contract MockERC20 is ERC20 {
    constructor() ERC20("Mock Token", "MOCK") {}

    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }
}
