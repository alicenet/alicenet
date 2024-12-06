// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract LegacyToken is ERC20 {
    constructor() ERC20("LegacyToken", "LT") {
        _mint(msg.sender, 320000000 * 10 ** decimals());
    }
}
