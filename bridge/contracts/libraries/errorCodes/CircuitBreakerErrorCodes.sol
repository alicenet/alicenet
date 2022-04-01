// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library CircuitBreakerErrorCodes {
    // CircuitBreaker error codes
    uint16 public constant CIRCUIT_BREAKER_OPENED = 500; //"CircuitBreaker: The Circuit breaker is opened!"
    uint16 public constant CIRCUIT_BREAKER_CLOSED = 501; //"CircuitBreaker: The Circuit breaker is closed!"
}
