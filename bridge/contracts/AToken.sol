// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "contracts/utils/Admin.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IStakingToken.sol";

/// @custom:salt AToken
/// @custom:deploy-type deployStatic
contract AToken is
    IStakingToken,
    ERC20Upgradeable,
    ImmutableFactory,
    ImmutableATokenMinter,
    ImmutableATokenBurner
{
    address internal immutable _legacyToken;
    bool internal _migrationAllowed;

    constructor(address legacyToken_)
        ImmutableFactory(msg.sender)
        ImmutableATokenMinter()
        ImmutableATokenBurner()
    {
        _legacyToken = legacyToken_;
    }

    function initialize() public onlyFactory initializer {
        __ERC20_init("AliceNet Staking Token", "ALCA");
    }

    function migrate(uint256 amount) public {
        require(_migrationAllowed, "MadTokens migration not allowed");
        IERC20(_legacyToken).transferFrom(msg.sender, address(this), amount);
        _mint(msg.sender, amount);
    }

    function allowMigration() public onlyFactory {
        _migrationAllowed = true;
    }

    function externalMint(address to, uint256 amount) public onlyATokenMinter {
        _mint(to, amount);
    }

    function externalBurn(address from, uint256 amount) public onlyATokenBurner {
        _burn(from, amount);
    }

    function getLegacyTokenAddress() public view returns (address) {
        return _legacyToken;
    }
}
