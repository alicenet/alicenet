// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IStakingToken.sol";
import "contracts/libraries/errors/StakingTokenErrors.sol";
import "contracts/utils/CircuitBreaker.sol";


/// @custom:salt AToken
/// @custom:deploy-type deployStatic
contract AToken is
    IStakingToken,
    ERC20Upgradeable,
    CircuitBreaker,
    ImmutableFactory,
    ImmutableATokenMinter,
    ImmutableATokenBurner
{
    uint256 internal constant _CONVERSION_MULTIPLIER = 1555555555555555556;
    uint256 internal constant _CONVERSION_SCALE = 1000000000000000000;
    address internal immutable _legacyToken;
    bool internal _migrationAllowed;
    bool internal _hasEarlyStageEnded;

    constructor(address legacyToken_)
        ImmutableFactory(msg.sender)
        ImmutableATokenMinter()
        ImmutableATokenBurner()
    {
        _legacyToken = legacyToken_;
    }

    function initialize(uint256 initialMintAmount) public onlyFactory initializer {
        __ERC20_init("AliceNet Staking Token", "ALCA");
        if (totalSupply() == 0) {
            _mint(msg.sender, _convert(initialMintAmount));
        }
    }

    function migrate(uint256 amount) public {
        if (!_migrationAllowed) {
            revert StakingTokenErrors.MigrationNotAllowed();
        }
        IERC20(_legacyToken).transferFrom(msg.sender, address(this), amount);
        _mint(msg.sender, _convert(amount));
    }

    function allowMigration() public onlyFactory {
        _migrationAllowed = true;
    }

    function finishEarlyStage() public onlyFactory {
        _finishEarlyStage();
    }

    function toggleMultiplierOn() public onlyFactory {
        _toggleMultiplierOn();
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

    function convert(uint256 amount) public view returns (uint256) {
        return _convert(amount);
    }

    function _finishEarlyStage() internal {
        _hasEarlyStageEnded = true;
    }

    function _toggleMultiplierOn() internal {
        _hasEarlyStageEnded = false;
    }

    function _convert(uint256 amount) internal view returns (uint256) {
        if (_hasEarlyStageEnded) {
            return amount;
        } else {
            return _multiplyTokens(amount);
        }
    }

    function _multiplyTokens(uint256 amount) internal pure returns (uint256) {
        return (amount * _CONVERSION_MULTIPLIER) / _CONVERSION_SCALE;
    }
}
