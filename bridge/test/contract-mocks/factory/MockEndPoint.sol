
// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

interface IMockEndPoint {
    function i() external view returns (uint256);

    function addOne() external;

    function addTwo() external;

    function factory() external returns (address);
}

/// @custom:salt MockEndPoint
contract MockEndPoint {
    address public immutable factory_;
    address public owner;
    uint256 public i;
    event addedOne(uint256 indexed i);
    event addedTwo(uint256 indexed i);
    event UpgradeLock(bool indexed lock);

    constructor(address f) {
        factory_ = f;
    }

    function addOne() public {
        i++;
        emit addedOne(i);
    }

    function addTwo() public {
        i = i + 2;
        emit addedTwo(i);
    }
}
