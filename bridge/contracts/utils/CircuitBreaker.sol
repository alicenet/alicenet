// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "contracts/libraries/errors/CircuitBreakerErrors.sol";

abstract contract CircuitBreaker {
    bool internal constant _OPEN = true;
    bool internal constant _CLOSED = false;

    // cb is the circuit breaker
    // cb is a set only object
    bool internal _cb = _CLOSED;

    // withCB is a modifier to enforce the CB must
    // be set for a call to succeed
    modifier withCB() {
        if (_cb == _OPEN) {
            revert CircuitBreakerErrors.CircuitBreakerOpened();
        }

        _;
    }

    function cbState() public view returns (bool) {
        return _cb;
    }

    function _tripCB() internal {
        if (_cb == _OPEN) {
            revert CircuitBreakerErrors.CircuitBreakerOpened();
        }
        _cb = _OPEN;
    }

    function _resetCB() internal {
        if (_cb == _CLOSED) {
            revert CircuitBreakerErrors.CircuitBreakerClosed();
        }
        _cb = _CLOSED;
    }
}
