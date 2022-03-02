// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;


abstract contract Sigmoid {
    
    function _fx(uint256 x) internal pure returns(uint256) {
      return 201*x + 504975246918103897890400 - _sqrt(200**2*((_safeAbsSub(2500000000000000000000, x))**2 + 125000000000000005611050234958650739260304));
    }

    function _fp(uint256 p) internal pure returns(uint256) {
      return (201*p + _sqrt(200**2*(p**2 + 1015056498152694417606544040542564448516855541904 - 2014950493836207795780800*p)) - 201500024630538883475970400)/401;
    }

    function _safeAbsSub(uint256 a, uint256 b) internal pure returns(uint256) {
      return _max(a,b)- _min(a,b);
    }

    function _min(uint256 a_, uint256 b_) internal pure returns(uint256) {
      if (a_ <= b_) {
        return a_;
      }
      return b_;
    }

    function _max(uint256 a_, uint256 b_) internal pure returns(uint256) {
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
