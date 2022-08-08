// SPDX-License-Identifier: MIT
pragma solidity ^0.8.11;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/utils/ImmutableAuth.sol";
import {DutchAuctionErrors} from "contracts/libraries/errors/DutchAuctionErrors.sol";
import "hardhat/console.sol";

/// @custom:salt DutchAuction
/// @custom:deploy-type deployUpgradeable
contract DutchAuction is Initializable, ImmutableFactory {
    uint32 private _bidders = 4; // todo get the current number of validators from ValidatorPool
    uint256 private immutable _startPrice = 1000000 * 10 ** 18; // 100000 ETH
    uint256 private immutable _finalPrice = 1200000 * 100 * _bidders; // 1200000 is the cost in gas units for an ETHDKG complete round and 100 is an estimated gas price in weis
    uint256 private immutable _initialDelta = 20;
    uint256 private _startBlock = 0;

    constructor() ImmutableFactory(msg.sender) {}

    function initialize() public {
    }

    /// @dev Starts auction defining auction's start block
    function startAuction(
    ) public onlyFactory {
        _startBlock = block.number;
    }

    /// @dev Returns dutch auction price for current block
    function getPrice() public view returns (uint256) {
        if (_startBlock == 0 ) revert DutchAuctionErrors.AuctionNotStarted();
        return _dutchAuctionPrice(block.number - _startBlock, _bidders);  
    }

    /// @notice Calculates dutch auction price for the specified period (number of blocks since auction initialization)
    /// and number of bidders
    /// @dev
    /// @param blocks blocks since the auction started
    /// @param n number of bidders (_bidders)
    function _dutchAuctionPrice(uint256 blocks, uint256 n) internal view returns (uint256 result) {
        uint256 alfa = _startPrice - _finalPrice;
        uint256 decay = _computeDecay(n);
        uint256 t1 = alfa * ( n + 1);
        uint256 t2 = decay * blocks + (n+ 1 ) ** 2;
        uint256 ratio = t1 / t2;
        return _finalPrice + ratio;
    }

    /// @notice Computes decay rate for the specified number of bidders
    /// @dev
    /// @param n current number of validators
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
