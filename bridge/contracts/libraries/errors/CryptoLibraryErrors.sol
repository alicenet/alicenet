// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library CryptoLibraryErrors {
    // CryptoLibrary errors
    error EllipticCurveAdditionFailed();
    error HashPointNotOnCurve();
    error HashPointUnsafeForSigning();
    error PointNotOnCurve();
    error SignatureIndicesLengthMismatch(uint256 signaturesLength, uint256 indicesLength);
    error SignaturesLengthThresholdNotMet(uint256 signaturesLength, uint256 threshold);
    error InverseArrayIncorrect();
}
