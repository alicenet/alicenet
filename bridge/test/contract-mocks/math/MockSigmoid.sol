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

    function p_a() public pure returns (uint256) {
        return Sigmoid._P_A;
    }

    function p_b() public pure returns (uint256) {
        return Sigmoid._P_B;
    }

    function p_c() public pure returns (uint256) {
        return Sigmoid._P_C;
    }

    function p_d() public pure returns (uint256) {
        return Sigmoid._P_D;
    }

    function p_inv_s() public pure returns (uint256) {
        return Sigmoid._P_INV_S;
    }

    function p_inv_c1() public pure returns (uint256) {
        return Sigmoid._P_INV_C_1;
    }

    function p_inv_c2() public pure returns (uint256) {
        return Sigmoid._P_INV_C_2;
    }

    function p_inv_c3() public pure returns (uint256) {
        return Sigmoid._P_INV_C_3;
    }

    function p_inv_d0() public pure returns (uint256) {
        return Sigmoid._P_INV_D_0;
    }

    function p_inv_d1() public pure returns (uint256) {
        return Sigmoid._P_INV_D_1;
    }
}
