// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;
import "contracts/libraries/math/Sigmoid.sol";

contract MockSigmoid is Sigmoid {
    function p(uint256 x) public pure returns (uint256) {
        return Sigmoid._p(x);
    }

    function p_inverse(uint256 x) public pure returns (uint256) {
        return Sigmoid._p_inverse(x);
    }

    function safeAbsSub(uint256 a, uint256 b) public pure returns (uint256) {
        return Sigmoid._safeAbsSub(a, b);
    }

    function min(uint256 a, uint256 b) public pure returns (uint256) {
        return Sigmoid._min(a, b);
    }

    function max(uint256 a, uint256 b) public pure returns (uint256) {
        return Sigmoid._max(a, b);
    }

    function sqrt(uint256 x) public pure returns (uint256) {
        return Sigmoid._sqrt(x);
    }
}
