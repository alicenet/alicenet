// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "contracts/utils/auth/ImmutableFactory.sol";
import "contracts/utils/auth/ImmutableValidatorPool.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/libraries/errors/DutchAuctionErrors.sol";

contract DutchAuction is ImmutableFactory, ImmutableValidatorPool {
    uint256 private _startPrice;
    uint8 private immutable _decay;
    uint16 private immutable _scaleParameter;
    uint256 private _auctionId;
    uint256 private _startBlock;
    uint256 private _finalPrice;
    bool private _auctionActive;

    event AuctionStarted(
        uint256 _auctionId,
        uint256 _startBlock,
        uint256 _startPrice,
        uint256 _finalPrice
    );
    event AuctionStopped(uint256 _auctionId);
    event BidPlaced(uint256 _auctionId, address winner, uint256 _winPrice);

    constructor(uint8 decay_, uint16 scaleParameter_) ImmutableFactory(msg.sender) {
        _decay = decay_;
        _scaleParameter = scaleParameter_;
    }

    /// @dev Starts auction defining auction's start block, this auction continues to run until a new start
    /// @param startPrice_ the start price of the auction
    function startAuction(uint256 startPrice_) public onlyFactory {
        if (_auctionActive == true) {
            revert DutchAuctionErrors.ActiveAuctionFound(_auctionId);
        }
        _startPrice = startPrice_;
        uint256 gasPrice;
        assembly ("memory-safe") {
            gasPrice := gasprice()
        }
        uint256 ethdkgValidatorCost = 1200000 * 2 * gasPrice; // ETHDKG ceremony is approx 1200000 gas units x2 (one to exit and one to re-enter) at current network gas price in weis
        _finalPrice =
            ethdkgValidatorCost *
            IValidatorPool(_validatorPoolAddress()).getValidatorsCount();
        if (_startPrice <= _finalPrice) {
            revert DutchAuctionErrors.StartPriceLowerThanFinalPrice(_startPrice, _finalPrice);
        }
        _startBlock = block.number;
        _auctionId++;
        _auctionActive = true;
        emit AuctionStarted(_auctionId, _startBlock, _startPrice, _finalPrice);
    }

    function stopAuction() public onlyFactory {
        _auctionActive = false;
        emit AuctionStopped(_auctionId);
    }

    /// @dev Put a bid on current price and finish auction
    function bid() public {
        if (_auctionActive != true) {
            revert DutchAuctionErrors.NoActiveAuctionFound();
        }
        emit BidPlaced(_auctionId, msg.sender, _dutchAuctionPrice(block.number - _startBlock));
        _auctionActive = false;
    }

    /// @dev Returns dutch auction price for current block
    function getPrice() public view returns (uint256) {
        if (_auctionActive != true) {
            revert DutchAuctionErrors.NoActiveAuctionFound();
        }
        return _dutchAuctionPrice(block.number - _startBlock);
    }

    /// @dev Calculates dutch auction price for the specified period (number of blocks since auction initialization)
    /// @param blocks blocks since the auction started
    function _dutchAuctionPrice(uint256 blocks) internal view returns (uint256 result) {
        uint256 _alfa = _startPrice - _finalPrice;
        uint256 t1 = _alfa * _scaleParameter;
        uint256 t2 = _decay * blocks + _scaleParameter ** 2;
        uint256 ratio = t1 / t2;
        return _finalPrice + ratio;
    }
}
