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
        uint8 valueStoreSVA;
        uint8 curveSecp256k1;
        address account;
        bytes32 txHash;
    }
    /** @dev size in bytes of a VSPreImage capnproto structure without the cap'n
      proto header bytes*/
    uint256 internal constant _VS_PRE_IMAGE_SIZE = 104;
    /** @dev Number of bytes of a capnproto header, the state starts after the
      header */
    uint256 internal constant _CAPNPROTO_HEADER_SIZE = 8;

    /// @notice This function is for deserializing the VSPreImage struct from an defined
    ///         location inside a binary blob.
    /// @param src Binary state containing a TXInPreImage serialized struct without CapnProto headers
    function extractVSPreImage(bytes memory src) internal view returns (VSPreImage memory) {
        return extractInnerVSPreImage(src, _CAPNPROTO_HEADER_SIZE);
    }

    /// @notice This function is for deserializing the VSPreImage struct from an defined
    ///         location inside a binary blob.
    /// @param src Binary state containing a TXInPreImage serialized struct without CapnProto headers
    /// @param dataOffset offset to start reading the VSPreImage state from inside src
    /// @return vsPreImage a VSPreImage struct
    function extractInnerVSPreImage(bytes memory src, uint256 dataOffset)
        internal
        view
        returns (VSPreImage memory vsPreImage)
    {
        if (dataOffset + _VS_PRE_IMAGE_SIZE <= dataOffset) {
            revert GenericParserLibraryErrors.DataOffsetOverflow();
        }
        if (src.length < dataOffset + _VS_PRE_IMAGE_SIZE) {
            revert GenericParserLibraryErrors.InsufficientBytes(
                src.length,
                dataOffset + _VS_PRE_IMAGE_SIZE
            );
        }
        vsPreImage.chainId = BaseParserLibrary.extractUInt32(src, dataOffset);
        vsPreImage.txOutIdx = BaseParserLibrary.extractUInt32(src, dataOffset + 4);
        vsPreImage.txHash = BaseParserLibrary.extractBytes32(src, dataOffset + 16);
        vsPreImage.value = BaseParserLibrary.extractUInt32(src, dataOffset + 36); // todo verify this
        vsPreImage.account = address(
            bytes20(BaseParserLibrary.extractBytes(src, dataOffset + 50, 20))
        );
        if (vsPreImage.chainId == 0) {
            revert GenericParserLibraryErrors.ChainIdZero();
        }
    }
}
