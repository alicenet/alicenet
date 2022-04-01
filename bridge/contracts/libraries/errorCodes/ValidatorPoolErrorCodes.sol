// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library ValidatorPoolErrorCodes {
    // ValidatorPool error codes
    uint16 public constant VALIDATORPOOL_CALLER_NOT_VALIDATOR = 800; //"ValidatorPool: Only validators allowed!"
    uint16 public constant VALIDATORPOOL_CONSENSUS_RUNNING = 801; //"ValidatorPool: Error AliceNet Consensus should be halted!"
    uint16 public constant VALIDATORPOOL_ETHDKG_ROUND_RUNNING = 802; //"ValidatorPool: There's an ETHDKG round running!"
    uint16 public constant VALIDATORPOOL_ONLY_CONTRACTS_ALLOWED = 803; //"Only NFT contracts allowed to send ethereum!"
    uint16 public constant VALIDATORPOOL_MIN_BLOCK_INTERVAL_NOT_MET = 804; //"ValidatorPool: Condition not met to stop consensus!"
    uint16 public constant VALIDATORPOOL_MAX_VALIDATORS_MET = 805; //"ValidatorPool: There are not enough free spots for all new validators!"
    uint16 public constant VALIDATORPOOL_REGISTRATION_PARAMETER_LENGTH_MISMATCH = 806; //"ValidatorPool: Both input array should have same length!"
    uint16 public constant VALIDATORPOOL_FACTORY_SHOULD_OWN_POSITION = 807; //"ValidatorPool: The factory should be the owner of the StakeNFT position!"
    uint16 public constant VALIDATORPOOL_VALIDATORS_GREATER_THAN_AVAILABLE = 808; //"ValidatorPool: There are not enough validators to be removed!"
    uint16 public constant VALIDATORPOOL_PROFITS_ONLY_CLAIMABLE_DURING_CONSENSUS = 809; //"ValidatorPool: Profits can only be claimable when consensus is running!"
    uint16 public constant VALIDATORPOOL_TOKEN_BALANCE_CHANGED = 810; //"ValidatorPool: Invalid transaction, token balance of the contract changed!"
    uint16 public constant VALIDATORPOOL_ETH_BALANCE_CHANGED = 811; //"ValidatorPool: Invalid transaction, eth balance of the contract changed!"
    uint16 public constant VALIDATORPOOL_SENDER_NOT_IN_EXITING_QUEUE = 812; //"ValidatorPool: Address not in the exitingQueue!"
    uint16 public constant VALIDATORPOOL_WAITING_PERIOD_NOT_MET = 813; //"ValidatorPool: The waiting period is not over yet!"
    uint16 public constant VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE = 814; //"ValidatorPool: DishonestValidator should be a validator or be in the exiting line!"
    uint16 public constant VALIDATORPOOL_INVALID_INDEX = 815; //"Index out boundaries!"
    uint16 public constant VALIDATORPOOL_ADDRESS_ALREADY_VALIDATOR = 816; // "ValidatorPool: Address is already a validator or it is in the exiting line!"
    uint16 public constant VALIDATORPOOL_ADDRESS_NOT_VALIDATOR = 817; // "ValidatorPool: Address is not a validator_!"
    uint16 public constant VALIDATORPOOL_MINIMUM_STAKE_NOT_MET = 818; // "ValidatorStakeNFT: Error, the Stake position doesn't have enough funds!"
    uint16 public constant VALIDATORPOOL_PAYOUT_TOO_LOW = 819; // "ValidatorPool: Miner shares greater then the total payout in tokens!"
    uint16 public constant VALIDATORPOOL_ADDRESS_NOT_ACCUSABLE = 820; // "ValidatorPool: Address is not accusable!"
    uint16 public constant VALIDATORPOOL_INSUFFICIENT_FUNDS_IN_STAKE_POSITION = 821; // "ValidatorPool: Error, the Stake position doesn't have enough funds!"
}
