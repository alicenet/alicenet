// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;


abstract contract MagicValue {
    // _magicValue is a constant that may be used to prevent
    // a user from calling a dangerous method without significant
    // effort or ( hopefully ) reading the code to understand the risk
    uint8 constant _magicValue = 42;
    
    // _getMagic returns the magic constant
    function _getMagic() internal pure returns(uint8) {
        return _magicValue;   
    }
    
    modifier checkMagic(uint8 magic_) {
        require(magic_ == _getMagic(), "BAD MAGIC");
        _;
    }
}