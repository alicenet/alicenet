// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "contracts/utils/Admin.sol";

abstract contract LegacyTokenBase is ERC20Upgradeable, Admin {
    function __legacyTokenBaseInit() internal onlyInitializing {
        __ERC20_init("LegacyToken", "LT");
    }
}

contract LegacyToken is LegacyTokenBase {
    constructor() Admin(msg.sender) {}

    function initialize() public onlyAdmin initializer {
        __legacyTokenBaseInit();
        _mint(msg.sender, 320000000 * 10**decimals());
    }
}
