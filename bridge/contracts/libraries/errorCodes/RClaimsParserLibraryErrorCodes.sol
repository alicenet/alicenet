// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library RClaimsParserLibraryErrorCodes {
    // RClaimsParserLibrary error codes
    bytes32 internal constant RCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW = "1500"; //"RClaimsParserLibrary: Overflow on the dataOffset parameter"
    bytes32 internal constant RCLAIMSPARSERLIB_INSUFFICIENT_BYTES = "1501"; // "RClaimsParserLibrary: Not enough bytes to extract RClaims"
    bytes32 internal constant RCLAIMSPARSERLIB_CHAINID_ZERO = "1502"; // "RClaimsParserLibrary: Invalid parsing. The chainId should be greater than 0!"
    bytes32 internal constant RCLAIMSPARSERLIB_HEIGHT_ZERO = "1503"; // "RClaimsParserLibrary: Invalid parsing. The height should be greater than 0!"
    bytes32 internal constant RCLAIMSPARSERLIB_ROUND_ZERO = "1504"; // "RClaimsParserLibrary: Invalid parsing. The round should be greater than 0!"
}
