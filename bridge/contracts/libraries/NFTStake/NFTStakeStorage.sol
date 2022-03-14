// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

abstract contract NFTStakeStorage {
    // Position describes a staked position
    struct Position {
        // number of madToken
        uint224 shares;
        // block number after which the position may be burned.
        // prevents double spend of voting weight
        uint32 freeAfter;
        // block number after which the position may be collected or burned.
        uint256 withdrawFreeAfter;
        // the last value of the ethState accumulator this account performed a
        // withdraw at
        uint256 accumulatorEth;
        // the last value of the tokenState accumulator this account performed a
        // withdraw at
        uint256 accumulatorToken;
    }

    // Accumulator is a struct that allows values to be collected such that the
    // remainders of floor division may be cleaned up
    struct Accumulator {
        // accumulator is a sum of all changes always increasing
        uint256 accumulator;
        // slush stores division remainders until they may be distributed evenly
        uint256 slush;
    }

    // _MAX_MINT_LOCK describes the maximum interval a Position may be locked
    // during a call to mintTo
    uint256 internal constant _MAX_MINT_LOCK = 1051200;
    // 10**18
    uint256 internal constant _ACCUMULATOR_SCALE_FACTOR = 1000000000000000000;
    // constants for the cb state
    bool internal constant _CIRCUIT_BREAKER_OPENED = true;
    bool internal constant _CIRCUIT_BREAKER_CLOSED = false;

    // monotonically increasing counter
    uint256 internal _counter;

    // cb is the circuit breaker
    // cb is a set only object
    bool internal _circuitBreaker;

    // _shares stores total amount of MadToken staked in contract
    uint256 internal _shares;

    // _tokenState tracks distribution of MadToken that originate from slashing
    // events
    Accumulator internal _tokenState;

    // _ethState tracks the distribution of Eth that originate from the sale of
    // MadBytes
    Accumulator internal _ethState;

    // _positions tracks all staked positions based on tokenID
    mapping(uint256 => Position) internal _positions;

    // state to keep track of the amount of Eth deposited and collected from the
    // contract
    uint256 internal _reserveEth;

    // state to keep track of the amount of MadTokens deposited and collected
    // from the contract
    uint256 internal _reserveToken;
}
