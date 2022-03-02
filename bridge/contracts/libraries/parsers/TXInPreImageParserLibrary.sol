// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "./BaseParserLibrary.sol";

/// @title Library to parse the TXInPreImage structure from a blob of capnproto data
library TXInPreImageParserLibrary {
    /** @dev size in bytes of a TXInPreImage cap'npro structure without the cap'n
      proto header bytes*/
    uint256 internal constant TX_IN_PRE_IMAGE_SIZE = 48;
    /** @dev Number of bytes of a capnproto header, the data starts after the
      header */
    uint256 internal constant CAPNPROTO_HEADER_SIZE = 8;

    struct TXInPreImage {
        uint32 chainId;
        uint32 consumedTxIdx;
        bytes32 consumedTxHash; //todo: is always 32 bytes?
    }

    /**
    @notice This function is for deserializing data directly from capnproto
            TXInPreImage. It will skip the first 8 bytes (capnproto headers) and
            deserialize the TXInPreImage Data. If TXInPreImage is being extracted from
            inside of other structure use the
            `extractTXInPreImage(bytes, uint)` instead.
    */
    /// @param src Binary data containing a TXInPreImage serialized struct with Capn Proto headers
    /// @dev Execution cost: 1120 gas
    /// @return a TXInPreImage struct
    function extractTXInPreImage(bytes memory src)
        internal
        pure
        returns (TXInPreImage memory)
    {
        return extractInnerTXInPreImage(src, CAPNPROTO_HEADER_SIZE);
    }

    /**
    @notice This function is for deserializing the TXInPreImage struct from an defined
            location inside a binary blob. E.G Extract TXInPreImage from inside of
            other structure (E.g RCert capnproto) or skipping the capnproto
            headers.
    */
    /// @param src Binary data containing a TXInPreImage serialized struct without CapnProto headers
    /// @param dataOffset offset to start reading the TXInPreImage data from inside src
    /// @dev Execution cost: 1084 gas
    /// @return txInPreImage a TXInPreImage struct
    function extractInnerTXInPreImage(bytes memory src, uint256 dataOffset)
        internal
        pure
        returns (TXInPreImage memory txInPreImage)
    {
        require(
            dataOffset + TX_IN_PRE_IMAGE_SIZE > dataOffset,
            "TXInPreImageParserLibrary: Overflow on the dataOffset parameter"
        );
        require(
            src.length >= dataOffset + TX_IN_PRE_IMAGE_SIZE,
            "TXInPreImageParserLibrary: Not enough bytes to extract TXInPreImage"
        );
        txInPreImage.chainId = BaseParserLibrary.extractUInt32(src, dataOffset);
        require(txInPreImage.chainId > 0, "TXInPreImageParserLibrary: Invalid parsing. The chainId should be greater than 0!");
        txInPreImage.consumedTxIdx = BaseParserLibrary.extractUInt32(src, dataOffset + 4);
        txInPreImage.consumedTxHash = BaseParserLibrary.extractBytes32(src, dataOffset + 16);
    }
}
