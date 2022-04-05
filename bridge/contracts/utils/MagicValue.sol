// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;
import {MagicValueErrorCodes} from "contracts/libraries/errorCodes/MagicValueErrorCodes.sol";

abstract contract MagicValue {
    // _MAGIC_VALUE is a constant that may be used to prevent
    // a user from calling a dangerous method without significant
    // effort or ( hopefully ) reading the code to understand the risk
    uint8 internal constant _MAGIC_VALUE = 42;

    modifier checkMagic(uint8 magic_) {
        require(
            magic_ == _getMagic(),
            string(abi.encodePacked(MagicValueErrorCodes.MAGICVALUE_BAD_MAGIC))
        );
        _;
    }

    // _getMagic returns the magic constant
    function _getMagic() internal pure returns (uint8) {
        return _MAGIC_VALUE;
    }
}
