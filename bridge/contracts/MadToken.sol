// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "contracts/utils/Admin.sol";

abstract contract MadTokenBase is ERC20Upgradeable, Admin {
    function __MadTokenBase_init() internal  onlyInitializing {
        __ERC20_init("MadToken", "MT");
    }
}

/// @custom:salt MadToken
/// @custom:deploy-type deployStatic
contract MadToken is MadTokenBase {
    //address constant oldMadToken = address(this);

    // address _minter;
    // address _burner;


    constructor() Admin(msg.sender) {}

    function initialize(address owner_) public onlyAdmin initializer {
        __MadTokenBase_init();
        _mint(owner_, 220000000 * 10 ** decimals());
    }

    /*
    function migrate(uint256 amt) public {
        // transferFrom(msg.sender, (addressThis), amt)
        // _mint(msg.sender, amt)
    }

    function setMinter(address minter_) public onlyAdmin {
        _minter = minter_;
    }

    function setBurner(address burner_) public onlyAdmin {
        _burner = burner_;
    }

    function externalMint(address to, uint256 amt) public onlyMinter {
        _mint(to, amt)
    }

    function externalBurn(address frm, uint256 amt) public onlyBurn {
        _burn(to, amt)
    }

    }
    */
}
