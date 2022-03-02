// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;


abstract contract GovernanceMaxLock {

    // _maxGovernanceLock describes the maximum interval
    // a position may remained locked due to a
    // governance action
    // this value is approx 30 days worth of blocks
    // prevents double spend of voting weight
    uint256 constant _maxGovernanceLock = 172800;

}
