// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import {MutexErrorCodes} from "contracts/libraries/errorCodes/MutexErrorCodes.sol";

abstract contract Mutex {
    uint256 internal constant _LOCKED = 1;
    uint256 internal constant _UNLOCKED = 2;
    uint256 internal _mutex;

    modifier withLock() {
        require(_mutex != _LOCKED, string(abi.encodePacked(MutexErrorCodes.MUTEX_LOCKED)));
        _mutex = _LOCKED;
        _;
        _mutex = _UNLOCKED;
    }

    constructor() {
        _mutex = _UNLOCKED;
    }
}
