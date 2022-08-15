// SPDX-License-Identifier: MIT
pragma solidity ^0.8.11;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IValidatorPool.sol";

/// @custom:salt DutchAuction
/// @custom:deploy-type deployUpgradeable
contract DutchAuction is Initializable, ImmutableFactory, ImmutableValidatorPool {
    uint256 private immutable _startPrice = 1000000 * 10**18;
    uint256 private immutable _ethdkgValidatorCost = 1200000 * 2 * 100 * 10**9; // Exit and enter at an estimated price of 100 gwei
    uint8 private immutable _decay = 16;
    uint16 private immutable _scaleParameter = 100;
    uint256 private _startBlock;
    uint256 private _finalPrice;

    constructor() ImmutableFactory(msg.sender) {}

    function initialize() public {
        resetAuction();
    }

    /// @dev Re-starts auction defining auction's start block
    function resetAuction() public onlyFactory {
        _finalPrice =
            _ethdkgValidatorCost *
            IValidatorPool(_validatorPoolAddress()).getValidatorsCount();
        _startBlock = block.number;
    }

    /// @dev Returns dutch auction price for current block
    function getPrice() public view returns (uint256) {
        return _dutchAuctionPrice(block.number - _startBlock);
    }

    /// @notice Calculates dutch auction price for the specified period (number of blocks since auction initialization)
    /// @dev
    /// @param blocks blocks since the auction started
    function _dutchAuctionPrice(uint256 blocks) internal view returns (uint256 result) {
        uint256 _alfa = _startPrice - _finalPrice;
        uint256 t1 = _alfa * _scaleParameter;
        uint256 t2 = _decay * blocks + _scaleParameter**2;
        uint256 ratio = t1 / t2;
        return _finalPrice + ratio;
    }
}
