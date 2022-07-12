// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library CircuitBreakerErrorCodes {
    // CircuitBreaker error codes
    bytes32 internal constant CIRCUIT_BREAKER_OPENED = "500"; //"CircuitBreaker: The Circuit breaker is opened!"
    bytes32 internal constant CIRCUIT_BREAKER_CLOSED = "501"; //"CircuitBreaker: The Circuit breaker is closed!"
}
