// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IGovernor.sol";
import {GovernanceErrorCodes} from "contracts/libraries/errorCodes/GovernanceErrorCodes.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

/// @custom:salt Governance
/// @custom:deploy-type deployUpgradeable
contract Governance is IGovernor {
    using Strings for uint16;
    address internal immutable _factory;

    constructor() {
        _factory = msg.sender;
    }

    function updateValue(
        uint256 epoch,
        uint256 key,
        bytes32 value
    ) external {
        require(
            msg.sender == _factory,
            GovernanceErrorCodes.GOVERNANCE_ONLY_FACTORY_ALLOWED.toString()
        );
        emit ValueUpdated(epoch, key, value, msg.sender);
    }
}
