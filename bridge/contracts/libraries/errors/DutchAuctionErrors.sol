// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

library DutchAuctionErrors {
    error StartPriceLowerThanFinalPrice(uint256 startPrice, uint256 finalPrice);
    error ActiveAuctionFound(uint auctionId);
    error NoActiveAuctionFound();
}
