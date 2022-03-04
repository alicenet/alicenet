// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

abstract contract Mutex {
    uint256 internal constant _LOCKED = 1;
    uint256 internal constant _UNLOCKED = 2;
    uint256 internal _mutex;

    modifier withLock() {
        require(_mutex != _LOCKED, "Mutex: Couldn't acquire the lock!");
        _mutex = _LOCKED;
        _;
        _mutex = _UNLOCKED;
    }

    constructor() {
        _mutex = _UNLOCKED;
    }
}
