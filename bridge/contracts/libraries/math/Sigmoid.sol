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
        unchecked {
            if (x <= 1) {
                return x;
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
            if (xAux >= 0x4) {
                result <<= 1;
            }

        // The operations can never overflow because the result is max 2^127 when it enters this block.
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

    function _sqrtNew(uint256 x) internal pure returns (uint256) {
        unchecked {
            if (x <= 1) {
                return x;
            }
            if (x >= 0xfffffffffffffffffffffffffffffffe00000000000000000000000000000001) {
                return (1 << 128) - 1;
            }
            // Here, e represents the bit length;
            // its value is at most 256, so it could fit in a uint16.
            uint256 e = 1;
            // Here, result is a copy of x
            uint256 result = x;
            if (result >= 0x100000000000000000000000000000000) {
                result >>= 128;
                e += 128;
            }
            if (result >= 0x10000000000000000) {
                result >>= 64;
                e += 64;
            }
            if (result >= 0x100000000) {
                result >>= 32;
                e += 32;
            }
            if (result >= 0x10000) {
                result >>= 16;
                e += 16;
            }
            if (result >= 0x100) {
                result >>= 8;
                e += 8;
            }
            if (result >= 0x10) {
                result >>= 4;
                e += 4;
            }
            if (result >= 0x4) {
                result >>= 2;
                e += 2;
            }
            if (result >= 0x2) {
                e += 1;
            }

            // e is currently bit length; we overwrite it now
            e = (256 - e) >> 1;
            // m now satisfies 2**254 <= m < 2**256
            uint256 m = x << (2*e);
            // result now stores the result
            result = 1 + (m >> 254);
            result = (result << 1)  + (m >> 251) / result;
            result = (result << 3)  + (m >> 245) / result;
            result = (result << 7)  + (m >> 233) / result;
            result = (result << 15) + (m >> 209) / result;
            result = (result << 31) + (m >> 161) / result;
            result = (result << 63) + (m >>  65) / result;
            result >>= e;
            return result*result <= x ? result : (result-1);
        }
    }
}
