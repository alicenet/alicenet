// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

interface IGovernor {

    event ValueUpdated(uint256 indexed epoch, uint256 indexed key, bytes32 indexed value, address who);

    function updateValue(uint256 epoch, uint256 key, bytes32 value) external;

}