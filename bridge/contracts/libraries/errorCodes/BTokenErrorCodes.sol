// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library BTokenErrorCodes {
    // BToken error codes
    bytes32 public constant BTOKEN_INVALID_DEPOSIT_ID = "300"; //"BToken: Invalid deposit ID!"
    bytes32 public constant BTOKEN_INVALID_BALANCE = "301"; //"BToken: Address balance should be always greater than the pool balance!"
    bytes32 public constant BTOKEN_INVALID_BURN_AMOUNT = "302"; //"BToken: The number of BTokens to be burn should be greater than 0!"
    bytes32 public constant BTOKEN_CONTRACTS_DISALLOWED_DEPOSITS = "303"; //"BToken: Contracts cannot make BTokens deposits!"
    bytes32 public constant BTOKEN_DEPOSIT_AMOUNT_ZERO = "304"; //"BToken: The deposit amount must be greater than zero!"
    bytes32 public constant BTOKEN_DEPOSIT_BURN_FAIL = "305"; //"BToken: Burn failed during the deposit!"
    bytes32 public constant BTOKEN_MARKET_SPREAD_TOO_LOW = "306"; //"BToken: requires at least 4 WEI"
    bytes32 public constant BTOKEN_MINT_INSUFFICIENT_ETH = "307"; //"BToken: could not mint deposit with minimum BTokens given the ether sent!"
    bytes32 public constant BTOKEN_MINIMUM_MINT_NOT_MET = "308"; //"BToken: could not mint minimum BTokens"
    bytes32 public constant BTOKEN_MINIMUM_BURN_NOT_MET = "309"; //"BToken: Couldn't burn the minEth amount"
    bytes32 public constant BTOKEN_SPLIT_VALUE_SUM_ERROR = "310"; //"BToken: All the split values must sum to _PERCENTAGE_SCALE!"
    bytes32 public constant BTOKEN_BURN_AMOUNT_EXCEEDS_SUPPLY = "311"; //"BToken: The number of tokens to be burned is greater than the Total Supply!"
}
