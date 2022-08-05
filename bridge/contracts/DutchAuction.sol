// SPDX-License-Identifier: MIT
pragma solidity ^0.8.11;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/utils/ImmutableAuth.sol";
import {DutchAuctionErrors} from "contracts/libraries/errors/DutchAuctionErrors.sol";

/// @custom:salt DutchAuction
/// @custom:deploy-type deployUpgradeable
contract DutchAuction is Initializable, ImmutableFactory {
    uint256 private _startPrice;
    uint256 private _finalPrice;
    uint32 private _bidders;
    uint256 private _startBlock;
    uint256 private _stopBlock;
    uint256 private _durationInBlocks;
    uint256 private _initialDelta;

    constructor() ImmutableFactory(msg.sender) {}

    /// @dev Initializes contract with initialDelta
    /// @param initialDelta_ initial delta for decay computing
    function initialize(uint256 initialDelta_) public initializer onlyFactory {
        _initialDelta = initialDelta_;
    }

    /// @dev Starts auction defining starting price, final price, number of bidders and duration in blocks
    /// @param startPrice_ start price for the auction
    /// @param finalPrice_ final price for the auction
    /// @param bidders_ number of bidders
    /// @param durationInBlocks_ auction duration expressed in Ethereum blocks
    function startAuction(
        uint256 startPrice_,
        uint256 finalPrice_,
        uint32 bidders_,
        uint256 durationInBlocks_
    ) public onlyFactory {
        if (startPrice_ <= finalPrice_) revert DutchAuctionErrors.IcorrectInitialPrices();
        _startPrice = startPrice_;
        _finalPrice = finalPrice_;
        _bidders = bidders_;
        _durationInBlocks = durationInBlocks_;
        _startAuction();
    }

    /// @dev Returns dutch auction price for current block
    function getPrice() public view returns (uint256) {
        if (block.number - _startBlock > _durationInBlocks)
            revert DutchAuctionErrors.AuctionClosed();
        return _dutchAuctionPrice(block.number - _startBlock, _bidders);
    }

    /// @dev Starts the auction setting auction start block to the current block
    function _startAuction() internal {
        _startBlock = block.number;
    }

    /// @notice Calculates dutch auction price for the specified period (number of blocks since auction initialization)
    /// and number of bidders
    /// @dev
    /// @param blocks blocks since the auction started
    /// @param n number of bidders (_bidders)
    function _dutchAuctionPrice(uint256 blocks, uint256 n) internal view returns (uint256 result) {
        uint256 alfa = _startPrice - _finalPrice;
        uint256 decay = _computeDecay(n);
        uint256 t1 = alfa * _durationInBlocks;
        uint256 t2 = decay * blocks + _durationInBlocks**2;
        uint256 ratio = t1 / t2;
        return _finalPrice + ratio;
    }

    /// @notice Computes decay rate for the specified number of bidders
    /// @dev
    /// @param n number of bidders (_bidders)
    function _computeDecay(uint256 n) internal view returns (uint256 result) {
        uint256 delta;
        if (n >= 64) {
            delta = 1;
        } else if (n >= 32) {
            delta = 2;
        } else if (n >= 16) {
            delta = 4;
        } else if (n >= 8) {
            delta = 8;
        } else if (n >= 4) {
            delta = 16;
        } else delta = 32;
        return _initialDelta + delta;
    }
}
