// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library StakingNFTErrorCodes {
    // StakingNFT error codes
    bytes32 public constant STAKENFT_CALLER_NOT_TOKEN_OWNER = "600"; //"PublicStaking: Error, token doesn't exist or doesn't belong to the caller!"
    bytes32 public constant STAKENFT_LOCK_DURATION_GREATER_THAN_GOVERNANCE_LOCK = "601"; //"PublicStaking: Lock Duration is greater than the amount allowed!"
    bytes32 public constant STAKENFT_LOCK_DURATION_GREATER_THAN_MINT_LOCK = "602"; // "PublicStaking: The lock duration must be less or equal than the maxMintLock!"
    bytes32 public constant STAKENFT_LOCK_DURATION_WITHDRAW_TIME_NOT_REACHED = "603"; // "PublicStaking: Cannot withdraw at the moment."
    bytes32 public constant STAKENFT_INVALID_TOKEN_ID = "604"; // "PublicStaking: Error, NFT token doesn't exist!"
    bytes32 public constant STAKENFT_MINT_AMOUNT_EXCEEDS_MAX_SUPPLY = "605"; // PublicStaking: The amount exceeds the maximum number of ATokens that will ever exist!"
    bytes32 public constant STAKENFT_FREE_AFTER_TIME_NOT_REACHED = "606"; //  "PublicStaking: The position is not ready to be burned!"
    bytes32 public constant STAKENFT_BALANCE_LESS_THAN_RESERVE = "607"; //  "PublicStaking: The balance of the contract is less then the tracked reserve!"
    bytes32 public constant STAKENFT_SLUSH_TOO_LARGE = "608"; // "PublicStaking: slush too large"
}
