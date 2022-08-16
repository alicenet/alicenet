// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "./utils/Admin.sol";
import "./utils/ImmutableAuth.sol";
import "contracts/interfaces/IAToken.sol";
import "contracts/utils/CircuitBreaker.sol";

/// @custom:salt AToken
/// @custom:deploy-type deployStatic
contract AToken is
    IAToken,
    ERC20Upgradeable,
    CircuitBreaker,
    ImmutableFactory,
    ImmutableATokenMinter,
    ImmutableATokenBurner
{
    uint256 internal constant CONVERSION_MULTIPLIER = 15;
    uint256 internal constant CONVERSION_SCALE = 10;
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
        __ERC20_init("AToken", "ALC");
    }

    function migrate(uint256 amount) public {
        require(_migrationAllowed, "MadTokens migration not allowed");
        IERC20(_legacyToken).transferFrom(msg.sender, address(this), amount);
        uint256 scaledAmount = (amount * CONVERSION_MULTIPLIER) / CONVERSION_SCALE;
        _mint(msg.sender, scaledAmount);
    }

    function allowMigration() public onlyFactory {
        _migrationAllowed = true;
    }

    /// tripCB opens the circuit breaker, may only be called by the factory
    function tripCB() public onlyFactory {
        _tripCB();
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
