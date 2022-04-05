// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import {
    CircuitBreakerErrorCodes
} from "contracts/libraries/errorCodes/CircuitBreakerErrorCodes.sol";

abstract contract CircuitBreaker {
    bool internal constant _OPEN = true;
    bool internal constant _CLOSED = false;

    // cb is the circuit breaker
    // cb is a set only object
    bool internal _cb = _CLOSED;

    // withCB is a modifier to enforce the CB must
    // be set for a call to succeed
    modifier withCB() {
        require(
            _cb == _CLOSED,
            string(abi.encodePacked(CircuitBreakerErrorCodes.CIRCUIT_BREAKER_OPENED))
        );
        _;
    }

    function cbState() public view returns (bool) {
        return _cb;
    }

    function _tripCB() internal {
        require(
            _cb == _CLOSED,
            string(abi.encodePacked(CircuitBreakerErrorCodes.CIRCUIT_BREAKER_OPENED))
        );
        _cb = _OPEN;
    }

    function _resetCB() internal {
        require(
            _cb == _OPEN,
            string(abi.encodePacked(CircuitBreakerErrorCodes.CIRCUIT_BREAKER_CLOSED))
        );
        _cb = _CLOSED;
    }
}
