// SPDX-License-Identifier: MIT-open-group
pragma solidity >=0.7.0 <0.9.0;

/// @custom:salt CallAny
/// @custom:deploy-type deployUpgradeable
contract CallAny {
    function callAny(
        address target_,
        uint256 value_,
        bytes memory cdata_
    ) public {
        assembly {
            let size := mload(cdata_)
            let ptr := add(0x20, cdata_)
            if iszero(call(gas(), target_, value_, ptr, size, 0x00, 0x00)) {
                returndatacopy(0x00, 0x00, returndatasize())
                revert(0x00, returndatasize())
            }
            returndatacopy(0x00, 0x00, returndatasize())
            return(0x00, returndatasize())
        }
    }
}
