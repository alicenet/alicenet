// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

abstract contract Sigmoid {
    // Constants for P function
    uint256 internal constant _P_A = 200;
    uint256 internal constant _P_B = 2500 * 10**18;
    uint256 internal constant _P_C = 5611050234958650739260304 + 125 * 10**39;
    uint256 internal constant _P_D = 4;
    uint256 internal constant _P_S = 2524876234590519489452;

    // Constants for P Inverse function
    uint256 internal constant _P_INV_C_1 = _P_A * ((_P_A + _P_D) * _P_S + _P_A * _P_B);
    uint256 internal constant _P_INV_C_2 = _P_A + _P_D;
    uint256 internal constant _P_INV_C_3 = _P_D * (2 * _P_A + _P_D);
    uint256 internal constant _P_INV_D_0 = ((_P_A + _P_D) * _P_S + _P_A * _P_B)**2;
    uint256 internal constant _P_INV_D_1 = 2 * (_P_A * _P_S + (_P_A + _P_D) * _P_B);

    function _p(uint256 t) internal pure returns (uint256) {
        return
            (_P_A + _P_D) * t + (_P_A * _P_S) - _sqrt(_P_A**2 * ((_safeAbsSub(_P_B, t))**2 + _P_C));
    }

    function _pInverse(uint256 m) internal pure returns (uint256) {
        return
            (_P_INV_C_2 * m + _sqrt(_P_A**2 * (m**2 + _P_INV_D_0 - _P_INV_D_1 * m)) - _P_INV_C_1) /
            _P_INV_C_3;
    }

    function _safeAbsSub(uint256 a, uint256 b) internal pure returns (uint256) {
        return _max(a, b) - _min(a, b);
    }

    function _min(uint256 a_, uint256 b_) internal pure returns (uint256) {
        if (a_ <= b_) {
            return a_;
        }
        return b_;
    }

    function _max(uint256 a_, uint256 b_) internal pure returns (uint256) {
        if (a_ >= b_) {
            return a_;
        }
        return b_;
    }

    // _sqrt implements an algorithm for computing integer square roots
    // based on Newton's method for computing square roots.
    // See https://github.com/alicenet/.github/wiki/technical-articles
    // for an article discussing computing integer square roots in Solidity
    // as well as an explanation about these steps.
    function _sqrt(uint256 x) internal pure returns (uint256 result) {
        unchecked {
            // Take care of integer square root edge cases.
            // Naturally, Isqrt(0) == 0, Isqrt(1) == 1,
            // and Isqrt(x) == (1 << 128) - 1
            // for ((1<<128)-1)**2 <= x < 2**256
            if (x <= 1) {
                return x;
            }
            if (x >= ((1 << 128) - 1)**2) {
                return (1 << 128) - 1;
            }
            // We now compute an initial approximation of Isqrt(x)
            // (stored in result) by computing the largest power of 2
            // less than or equal to Isqrt(x).
            uint256 xAux = x;
            result = 1;
            if (xAux >= (1 << 128)) {
                xAux >>= 128;
                result <<= 64;
            }
            if (xAux >= (1 << 64)) {
                xAux >>= 64;
                result <<= 32;
            }
            if (xAux >= (1 << 32)) {
                xAux >>= 32;
                result <<= 16;
            }
            if (xAux >= (1 << 16)) {
                xAux >>= 16;
                result <<= 8;
            }
            if (xAux >= (1 << 8)) {
                xAux >>= 8;
                result <<= 4;
            }
            if (xAux >= (1 << 4)) {
                xAux >>= 4;
                result <<= 2;
            }
            if (xAux >= (1 << 2)) {
                result <<= 1;
            }
            // At this point, result is the largest power of 2
            // less than or equal to Isqrt(x).
            // We now compute a better initial approximation of Isqrt(x).
            result += (result >> 1);
            // At this point, we perform Newton iterations to compute
            // more accurate approximations of Isqrt(x).
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            // We perform one potential correction to ensure a correct value.
            // It is not possible to overflow at this point because
            // x < ((1 << 128) - 1)**2, so sqrt(x) <= (1 << 128) - 2.
            return result * result <= x ? result : (result - 1);
        }
    }
}
