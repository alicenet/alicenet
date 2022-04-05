// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library TXInPreImageParserLibraryErrorCodes {
    // TXInPreImageParserLibrary error codes
    bytes32 public constant TXINPREIMAGEPARSERLIB_DATA_OFFSET_OVERFLOW = "1600"; //"TXInPreImageParserLibrary: Overflow on the dataOffset parameter"
    bytes32 public constant TXINPREIMAGEPARSERLIB_INSUFFICIENT_BYTES = "1601"; // "TXInPreImageParserLibrary: Not enough bytes to extract TXInPreImage"
    bytes32 public constant TXINPREIMAGEPARSERLIB_CHAINID_ZERO = "1602"; // "TXInPreImageParserLibrary: Invalid parsing. The chainId should be greater than 0!"
}
