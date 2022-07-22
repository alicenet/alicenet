// SPDX-License-Identifier: MIT
pragma solidity ^0.8.11;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "hardhat/console.sol";

/// @custom:salt DutchAuction
/// @custom:deploy-type deployUpgradeable
contract DutchAuction is Initializable, ImmutableFactory {
    uint256 private constant DURATION = 10 seconds;

    IERC721 public nft;
    uint256 public nftId;

    address payable public seller;
    uint256 public startingPrice;
    uint256 public startAt;
    uint256 public expiresAt;
    uint256 public discountRate;

    constructor() ImmutableFactory(msg.sender) {}

    function initialize(
        uint256 _startingPrice,
        uint256 _discountRate,
        address _seller,
        address _nft,
        uint256 _nftId
    ) public initializer onlyFactory {
        seller = payable(_seller);
        startingPrice = _startingPrice;
        startAt = block.timestamp;
        expiresAt = block.timestamp + DURATION;
        discountRate = _discountRate;
        require(_startingPrice >= _discountRate * DURATION, "starting price < min");
        nft = IERC721(_nft);
        nftId = _nftId;
    }

    function getPrice() public view returns (uint256) {
        uint256 timeElapsed = block.timestamp - startAt;
        uint256 discount = discountRate * timeElapsed;
        return startingPrice - discount;
    }

    function getRemainingTime() public view returns (uint256) {
        uint256 timeElapsed = block.timestamp - startAt;
        uint256 timeRemaining = DURATION - timeElapsed;
        if (timeRemaining < 0) timeRemaining = 0;
        return timeRemaining;
    }

    function buy() external payable {
        require(block.timestamp < expiresAt, "auction expired");
        uint256 price = getPrice();
        require(msg.value >= price, "ETH < price");
        nft.transferFrom(seller, msg.sender, nftId);
        uint256 refund = msg.value - price;
        if (refund > 0) {
            payable(msg.sender).transfer(refund);
        }
        selfdestruct(seller);
    }
}
