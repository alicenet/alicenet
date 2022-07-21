// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

abstract contract Sigmoid {
    // Constants for P function
    uint256 internal constant _P_A = 200;
    uint256 internal constant _P_B = 2500 * 10**18;
    uint256 internal constant _P_C = 5611050234958650739260304 + 125 * 10**39;
    uint256 internal constant _P_D = 1;
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

    /// @notice Calculates the square root of x, rounding down.
    /// @dev Uses the Babylonian method https://en.wikipedia.org/wiki/Methods_of_computing_square_roots#Babylonian_method.
    ///
    /// Caveats:
    /// - This function does not work with fixed-point numbers.
    ///
    /// @param x The uint256 number for which to calculate the square root.
    /// @return result The result as an uint256.
    function _sqrt(uint256 x) internal pure returns (uint256 result) {
        if (x == 0) {
            return 0;
        }

        // Set the initial guess to the closest power of two that is higher than x.
        uint256 xAux = uint256(x);
        result = 1;
        if (xAux >= 0x100000000000000000000000000000000) {
            xAux >>= 128;
            result <<= 64;
        }
        if (xAux >= 0x10000000000000000) {
            xAux >>= 64;
            result <<= 32;
        }
        if (xAux >= 0x100000000) {
            xAux >>= 32;
            result <<= 16;
        }
        if (xAux >= 0x10000) {
            xAux >>= 16;
            result <<= 8;
        }
        if (xAux >= 0x100) {
            xAux >>= 8;
            result <<= 4;
        }
        if (xAux >= 0x10) {
            xAux >>= 4;
            result <<= 2;
        }
        if (xAux >= 0x8) {
            result <<= 1;
        }

        // The operations can never overflow because the result is max 2^127 when it enters this block.
        unchecked {
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1;
            result = (result + x / result) >> 1; // Seven iterations should be enough
            uint256 roundedDownResult = x / result;
            return result >= roundedDownResult ? roundedDownResult : result;
        }
    }
}
