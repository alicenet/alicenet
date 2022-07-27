// SPDX-License-Identifier: MIT
pragma solidity ^0.8.11;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "hardhat/console.sol";

/// @custom:salt DutchAuction
/// @custom:deploy-type deployUpgradeable
contract DutchAuction is Initializable, ImmutableFactory {
    uint256 private constant BLOCKS_DURATION = 10;

    address payable public seller;
    uint256 private _startingPrice;
    uint256 private _discountPerBlockRate;
    uint256 private _startAt;

    constructor() ImmutableFactory(msg.sender) {}
    /// @dev Initializes and starts auction defining starting price and discount per block rate
    /// @param startingPrice_ the start price of the auction
    /// @param discountPerBlockRate_ the rate that the price is decreasing per block
    function initialize(uint256 startingPrice_, uint256 discountPerBlockRate_)
        public
        initializer
        onlyFactory
    {
        _startingPrice = startingPrice_;
        _discountPerBlockRate = discountPerBlockRate_;
        _resetAuction();
    }

    function getPrice() public view returns (uint256) {
        return _getPrice();
    }

    /// @dev Returns the current offered price depending in the blocks mined since auction start block
        function _getPrice() internal view returns (uint256) {
        uint256 blocksElapsed = block.number - _startAt;
        uint256 discount = _discountPerBlockRate * blocksElapsed;
        return _startingPrice - discount;
    }

    function resetAuction() public {
        _resetAuction();
    }

    /// @dev Starts over the auction setting auction start block to the current block
    function _resetAuction() internal {
        _startAt = block.number;
        require(_startingPrice >= _discountPerBlockRate * BLOCKS_DURATION, "starting price < min");
    }
}
