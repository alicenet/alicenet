// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library DynamicsErrors {
    error InvalidScheduledDate(uint256 scheduledDate, uint256 minScheduledDate, uint256 maxScheduledDate);
    error InvalidExtCodeSize(address addr, uint256 codeSize);
    error DynamicValueNotFound(uint256 epoch);
}
