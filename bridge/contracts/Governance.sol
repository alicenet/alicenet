// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IGovernor.sol";

/// @custom:salt Governance
/// @custom:deploy-type deployUpgradeable
contract Governance is IGovernor {
    address internal immutable _factory;

    constructor() {
        _factory = msg.sender;
    }

    function updateValue(uint256 epoch, uint256 key, bytes32 value) external {
        require(msg.sender == _factory , "Governance: Only factory allowed!");
        emit ValueUpdated(epoch, key, value, msg.sender);
    }
}