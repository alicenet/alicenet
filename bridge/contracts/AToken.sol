// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "./utils/Admin.sol";
import "./utils/ImmutableAuth.sol";
import "hardhat/console.sol";
import "contracts/interfaces/IAToken.sol";

/// @custom:salt AToken
/// @custom:deploy-type deployStatic
contract AToken is
    IAToken,
    ERC20Upgradeable,
    ImmutableFactory,
    ImmutableATokenMinter,
    ImmutableATokenBurner
{
    address internal immutable _legacyToken;

    constructor(address legacyToken_)
        ImmutableFactory(msg.sender)
        ImmutableATokenMinter()
        ImmutableATokenBurner()
    {
        _legacyToken = legacyToken_;
    }

    function initialize() public onlyFactory initializer {
        __ERC20_init("AToken", "ATK");
    }

    function migrate(uint256 amount) public {
        IERC20(_legacyToken).transferFrom(msg.sender, address(this), amount);
        _mint(msg.sender, amount);
    }

    function externalMint(address to, uint256 amount) public onlyATokenMinter {
        _mint(to, amount);
    }

    function externalBurn(address from, uint256 amount) public onlyATokenBurner {
        //add require
        _burn(from, amount);
    }
}
