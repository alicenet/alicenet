// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "contracts/utils/Admin.sol";

abstract contract MadTokenBase is ERC20Upgradeable, Admin {
    function __madTokenBaseInit() internal onlyInitializing {
        __ERC20_init("MadToken", "MT");
    }
}

contract MadToken is MadTokenBase {
    constructor() Admin(msg.sender) {}

    function initialize() public onlyAdmin initializer {
        __madTokenBaseInit();
        _mint(msg.sender, 220000000 * 10**decimals());
    }
}
