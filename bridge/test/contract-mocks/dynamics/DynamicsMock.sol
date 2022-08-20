// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/Dynamics.sol";

contract DynamicsMock is Dynamics {
    using DoublyLinkedListLogic for DoublyLinkedList;

    constructor() Dynamics() {}

    function testChangeDynamicValues(uint32 relativeExecutionEpoch, DynamicValues memory newValue)
        public
    {
        _changeDynamicValues(relativeExecutionEpoch, newValue);
    }

    function testUpdateHead(uint32 currentEpoch) public {
        uint32 nextEpoch = _dynamicValues.getNextEpoch(_dynamicValues.getHead());
        if (nextEpoch != 0 && currentEpoch >= nextEpoch) {
            _dynamicValues.setHead(nextEpoch);
        }
    }

    function testUpdateAliceNetNodeVersion(
        uint32 relativeUpdateEpoch,
        uint32 majorVersion,
        uint32 minorVersion,
        uint32 patch,
        bytes32 binaryHash
    ) public {
        _updateAliceNetNodeVersion(
            relativeUpdateEpoch,
            majorVersion,
            minorVersion,
            patch,
            binaryHash
        );
    }

    function testSetConfiguration(Configuration calldata newConfig) public {
        _configuration = newConfig;
    }

    function testEnconding() public pure returns (bytes memory) {
        DynamicValues memory initialValues = DynamicValues(
            Version.V1,
            4000,
            3000,
            3000,
            3000000,
            0,
            0,
            0
        );
        return Dynamics._encodeDynamicValues(initialValues);
    }

    function testCompactedRepresentation(
        uint32 majorVersion,
        uint32 minorVersion,
        uint32 patch
    ) public pure returns (uint256) {
        return _computeCompactedVersion(majorVersion, minorVersion, patch);
    }
}
