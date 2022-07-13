// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library CryptoLibraryErrorCodes {
    // CryptoLibrary error codes
    bytes32 public constant CRYPTOLIB_ELLIPTIC_CURVE_ADDITION = "700"; //"elliptic curve addition failed"
    bytes32 public constant CRYPTOLIB_ELLIPTIC_CURVE_MULTIPLICATION = "701"; //"elliptic curve multiplication failed"
    bytes32 public constant CRYPTOLIB_ELLIPTIC_CURVE_PAIRING = "702"; //"elliptic curve pairing failed"
    bytes32 public constant CRYPTOLIB_MODULAR_EXPONENTIATION = "703"; //"modular exponentiation falied"
    bytes32 public constant CRYPTOLIB_HASH_POINT_NOT_ON_CURVE = "704"; //"Invalid hash point: not on elliptic curve"
    bytes32 public constant CRYPTOLIB_HASH_POINT_UNSAFE = "705"; //"Dangerous hash point: not safe for signing"
    bytes32 public constant CRYPTOLIB_POINT_NOT_ON_CURVE = "706"; //"Invalid point: not on elliptic curve"
    bytes32 public constant CRYPTOLIB_SIGNATURES_INDICES_LENGTH_MISMATCH = "707"; //"Mismatch between length of signatures and index array"
    bytes32 public constant CRYPTOLIB_SIGNATURES_LENGTH_THRESHOLD_NOT_MET = "708"; //"Failed to meet required number of signatures for threshold"
    bytes32 public constant CRYPTOLIB_INVERSE_ARRAY_INCORRECT = "709"; //"invArray does not include correct inverses"
    bytes32 public constant CRYPTOLIB_INVERSE_ARRAY_THRESHOLD_EXCEEDED = "710"; // "checkInverses: insufficient inverses for group signature calculation"
    bytes32 public constant CRYPTOLIB_POINTSG1_INDICES_LENGTH_MISMATCH = "711"; // "Mismatch between pointsG1 and indices arrays"
    bytes32 public constant CRYPTOLIB_K_EQUAL_TO_J = "712"; // "Must have k != j when computing Rj partial constants"
}
