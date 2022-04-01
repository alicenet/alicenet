// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import {MutexErrorCodes} from "contracts/libraries/errorCodes/MutexErrorCodes.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

abstract contract Mutex {
    using Strings for uint16;

    uint256 internal constant _LOCKED = 1;
    uint256 internal constant _UNLOCKED = 2;
    uint256 internal _mutex;

    modifier withLock() {
        require(_mutex != _LOCKED, MutexErrorCodes.MUTEX_LOCKED.toString());
        _mutex = _LOCKED;
        _;
        _mutex = _UNLOCKED;
    }

    constructor() {
        _mutex = _UNLOCKED;
    }
}
