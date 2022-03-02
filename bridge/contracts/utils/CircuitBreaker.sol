// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;


abstract contract CircuitBreaker {

    bool constant open = true;
    bool constant closed = false;

    // cb is the circuit breaker
    // cb is a set only object
    bool _cb = closed;

    // withCB is a modifier to enforce the CB must
    // be set for a call to succeed
    modifier withCB() {
        require(_cb == closed, "CircuitBreaker: The Circuit breaker is opened!");
        _;
    }

    function cbState() public view returns(bool) {
        return _cb;
    }

    function _tripCB() internal {
        require(_cb == closed, "CircuitBreaker: The Circuit breaker is opened!");
        _cb = open;
    }

    function _resetCB() internal {
        require(_cb == open, "CircuitBreaker: The Circuit breaker is closed!");
        _cb = closed;
    }

}