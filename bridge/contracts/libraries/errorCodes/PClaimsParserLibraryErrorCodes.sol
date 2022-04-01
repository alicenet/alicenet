// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library PClaimsParserLibraryErrorCodes {
    // PClaimsParserLibrary error codes
    uint16 public constant PCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW = 1300; //"PClaimsParserLibrary: Overflow on the dataOffset parameter"
    uint16 public constant PCLAIMSPARSERLIB_INSUFFICIENT_BYTES = 1301; //"PClaimsParserLibrary: Not enough bytes to extract PClaims"
}
