// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "./utils/Admin.sol";
import "./utils/ImmutableAuth.sol";
import "contracts/interfaces/IAToken.sol";

import "hardhat/console.sol";

/// @custom:salt ATokenNotImmutable
/// @custom:deploy-type deployStatic
contract ATokenNotImmutable is IAToken, ERC20Upgradeable, ImmutableFactory {
    address internal immutable _oldMadToken;

    address internal _minter;
    address internal _burner;

    modifier onlyMinter() {
        require(msg.sender == _minter, "AToken: onlyMinter role allowed");
        _;
    }

    modifier onlyBurner() {
        require(msg.sender == _burner, "AToken: onlyBurner role allowed");
        _;
    }

    constructor(address oldMadToken_) ImmutableFactory(msg.sender) {
        _oldMadToken = oldMadToken_;
    }

    function initialize() public onlyFactory initializer {
        __ERC20_init("ATokenNotImmutable", "ATKNotImmutable");
    }

    function migrate(uint256 amount) public {
        IERC20(_oldMadToken).transferFrom(msg.sender, address(this), amount);
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
}
