// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library BClaimsParserLibraryErrorCodes {
    // BClaimsParserLibrary error codes
    uint16 public constant BCLAIMSPARSERLIB_SIZE_THRESHOLD_EXCEEDED = 1100; //"BClaimsParserLibrary: The size of the data section should be 1 or 2 words!"
    uint16 public constant BCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW = 1101; //"BClaimsParserLibrary: Invalid parsing. Overflow on the dataOffset parameter"
    uint16 public constant BCLAIMSPARSERLIB_NOT_ENOUGH_BYTES = 1102; //"BClaimsParserLibrary: Invalid parsing. Not enough bytes to extract BClaims"
    uint16 public constant BCLAIMSPARSERLIB_CHAINID_ZERO = 1103; //"BClaimsParserLibrary: Invalid parsing. The chainId should be greater than 0!"
    uint16 public constant BCLAIMSPARSERLIB_HEIGHT_ZERO = 1104; //"BClaimsParserLibrary: Invalid parsing. The height should be greater than 0!"
}
