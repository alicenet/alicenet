// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library BTokenErrors {
    error InvalidDepositId(uint256 depositID);
    error InvalidBalance(uint256 contractBalance, uint256 poolBalance);
    error InvalidBurnAmount(uint256 amount);
    error ContractsDisallowedDeposits(address toAddress);
    error DepositAmountZero();
    error DepositBurnFail(uint256 amount);
    error MarketSpreadTooLow(uint256 amount);
    error InsufficientEth(uint256 amount, uint256 minimum);
    error MinimumMintNotMet(uint256 amount, uint256 minimum);
    error MinimumBurnNotMet(uint256 amount, uint256 minimum);
    error SplitValueSumError();
    error BurnAmountExceedsSupply(uint256 amount, uint256 supply);
}
