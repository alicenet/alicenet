// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "./utils/Admin.sol";
import "./utils/ImmutableAuth.sol";
import "hardhat/console.sol";

/// @custom:salt AToken
/// @custom:deploy-type deployStatic
contract AToken is ERC20Upgradeable, ImmutableFactory {
    address internal immutable _oldMadToken = address(this);

    address internal _legacyToken;
    address internal _minter;
    address internal _burner;

    constructor() ImmutableFactory(msg.sender) {
    }

    function initialize(address legacyToken_) public onlyFactory initializer {
        __ERC20_init("AToken", "ATK");
        _legacyToken = legacyToken_;
    }

    function migrate(uint256 amount) public {
        ERC20Upgradeable(_legacyToken).transferFrom(msg.sender, address(this), amount);
        _mint(msg.sender, amount);
    }

    function setMinter(address minter_) public onlyFactory {
        _minter = minter_;
    }

    function setBurner(address burner_) public onlyFactory {
        _burner = burner_;
    }

    function externalMint(address to, uint256 amount) public onlyMinter {
        _mint(to, amount);
    }

    function externalBurn(address from, uint256 amount) public onlyBurner {
        _burn(from, amount);
    }

    modifier onlyMinter() {
        require(msg.sender == _minter, "onlyMinter role allowed");
        _;
    }

    modifier onlyBurner() {
        require(msg.sender == _burner, "onlyBurner role allowed");
        _;
    }
}
