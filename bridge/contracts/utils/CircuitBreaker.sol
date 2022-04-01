// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import {
    CircuitBreakerErrorCodes
} from "contracts/libraries/errorCodes/CircuitBreakerErrorCodes.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

abstract contract CircuitBreaker {
    using Strings for uint16;
    bool internal constant _OPEN = true;
    bool internal constant _CLOSED = false;

    // cb is the circuit breaker
    // cb is a set only object
    bool internal _cb = _CLOSED;

    // withCB is a modifier to enforce the CB must
    // be set for a call to succeed
    modifier withCB() {
        require(_cb == _CLOSED, CircuitBreakerErrorCodes.CIRCUIT_BREAKER_OPENED.toString());
        _;
    }

    function cbState() public view returns (bool) {
        return _cb;
    }

    function _tripCB() internal {
        require(_cb == _CLOSED, CircuitBreakerErrorCodes.CIRCUIT_BREAKER_OPENED.toString());
        _cb = _OPEN;
    }

    function _resetCB() internal {
        require(_cb == _OPEN, CircuitBreakerErrorCodes.CIRCUIT_BREAKER_CLOSED.toString());
        _cb = _CLOSED;
    }
}
