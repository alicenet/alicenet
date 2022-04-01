// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library RCertParserLibraryErrorCodes {
    // RCertParserLibrary error codes
    uint16 public constant RCERTPARSERLIB_DATA_OFFSET_OVERFLOW = 1400; //"RClaimsParserLibrary: Overflow on the dataOffset parameter"
    uint16 public constant RCERTPARSERLIB_INSUFFICIENT_BYTES = 1401; // "RCertParserLibrary: Not enough bytes to extract"
    uint16 public constant RCERTPARSERLIB_INSUFFICIENT_BYTES_TO_EXTRACT = 1402; // "RCertParserLibrary: Not enough bytes to extract RCert"
}
