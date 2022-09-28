// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/libraries/errors/GenericParserLibraryErrors.sol";
import "contracts/libraries/parsers/BaseParserLibrary.sol";

/// @title Library to parse the VSPreImage structure from a blob of capnproto state
library VSPreImageParserLibrary {
    struct VSPreImage {
        uint32 txOutIdx;
        uint32 chainId;
        uint256 value;
        uint8 ValueStoreSVA;
        uint8 CurveSecp256k1;
        address account;
        bytes32 txHash;
    }
    /** @dev size in bytes of a TXInPreImage cap'npro structure without the cap'n
      proto header bytes*/
    uint256 internal constant _TX_IN_PRE_IMAGE_SIZE = 104;
    /** @dev Number of bytes of a capnproto header, the state starts after the
      header */
    uint256 internal constant _CAPNPROTO_HEADER_SIZE = 8;

    /**
    @notice This function is for deserializing state directly from capnproto
            TXInPreImage. It will skip the first 8 bytes (capnproto headers) and
            deserialize the TXInPreImage Data. If TXInPreImage is being extracted from
            inside of other structure use the
            `extractTXInPreImage(bytes, uint)` instead.
    */
    /// @param src Binary state containing a TXInPreImage serialized struct with Capn Proto headers
    /// @dev Execution cost: 1120 gas
    /// @return a TXInPreImage struct
    function extractVSPreImage(bytes memory src) internal view returns (VSPreImage memory) {
        return extractInnerVSPreImage(src, _CAPNPROTO_HEADER_SIZE);
    }

    /**
    @notice This function is for deserializing the TXInPreImage struct from an defined
            location inside a binary blob. E.G Extract TXInPreImage from inside of
            other structure (E.g RCert capnproto) or skipping the capnproto
            headers.
    */
    /// @param src Binary state containing a TXInPreImage serialized struct without CapnProto headers
    /// @param dataOffset offset to start reading the TXInPreImage state from inside src
    /// @dev Execution cost: 1084 gas
    /// @return txInPreImage a TXInPreImage struct
    function extractInnerVSPreImage(bytes memory src, uint256 dataOffset)
        internal
        view
        returns (VSPreImage memory txInPreImage)
    {
        if (dataOffset + _TX_IN_PRE_IMAGE_SIZE <= dataOffset) {
            revert GenericParserLibraryErrors.DataOffsetOverflow();
        }
        if (src.length < dataOffset + _TX_IN_PRE_IMAGE_SIZE) {
            revert GenericParserLibraryErrors.InsufficientBytes(
                src.length,
                dataOffset + _TX_IN_PRE_IMAGE_SIZE
            );
        }
        txInPreImage.chainId = BaseParserLibrary.extractUInt32(src, dataOffset);
        txInPreImage.txOutIdx = BaseParserLibrary.extractUInt32(src, dataOffset + 4);
        txInPreImage.txHash = BaseParserLibrary.extractBytes32(src, dataOffset + 16);
        txInPreImage.value = BaseParserLibrary.extractUInt32(src, dataOffset + 36);
        txInPreImage.account = address(
            bytes20(BaseParserLibrary.extractBytes(src, dataOffset + 50, 20))
        );
        if (txInPreImage.chainId == 0) {
            revert GenericParserLibraryErrors.ChainIdZero();
        }
    }
}
