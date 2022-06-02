// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import {
    RClaimsParserLibraryErrorCodes
} from "contracts/libraries/errorCodes/RClaimsParserLibraryErrorCodes.sol";

import "./BaseParserLibrary.sol";

/// @title Library to parse the RClaims structure from a blob of capnproto state
library RClaimsParserLibrary {
    struct RClaims {
        uint32 chainId;
        uint32 height;
        uint32 round;
        bytes32 prevBlock;
    }

    /** @dev size in bytes of a RCLAIMS cap'npro structure without the cap'n
      proto header bytes*/
    uint256 internal constant _RCLAIMS_SIZE = 56;
    /** @dev Number of bytes of a capnproto header, the state starts after the
      header */
    uint256 internal constant _CAPNPROTO_HEADER_SIZE = 8;

    /**
    @notice This function is for deserializing state directly from capnproto
            RClaims. It will skip the first 8 bytes (capnproto headers) and
            deserialize the RClaims Data. If RClaims is being extracted from
            inside of other structure (E.g RCert capnproto) use the
            `extractInnerRClaims(bytes, uint)` instead.
    */
    /// @param src Binary state containing a RClaims serialized struct with Capn Proto headers
    /// @dev Execution cost: 1506 gas
    function extractRClaims(bytes memory src) internal pure returns (RClaims memory rClaims) {
        return extractInnerRClaims(src, _CAPNPROTO_HEADER_SIZE);
    }

    /**
    @notice This function is for serializing the RClaims struct from an defined
            location inside a binary blob. E.G Extract RClaims from inside of
            other structure (E.g RCert capnproto) or skipping the capnproto
            headers.
    */
    /// @param src Binary state containing a RClaims serialized struct without Capn Proto headers
    /// @param dataOffset offset to start reading the RClaims state from inside src
    /// @dev Execution cost: 1332 gas
    function extractInnerRClaims(bytes memory src, uint256 dataOffset)
        internal
        pure
        returns (RClaims memory rClaims)
    {
        require(
            dataOffset + _RCLAIMS_SIZE > dataOffset,
            string(
                abi.encodePacked(
                    RClaimsParserLibraryErrorCodes.RCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW
                )
            )
        );
        require(
            src.length >= dataOffset + _RCLAIMS_SIZE,
            string(
                abi.encodePacked(RClaimsParserLibraryErrorCodes.RCLAIMSPARSERLIB_INSUFFICIENT_BYTES)
            )
        );
        rClaims.chainId = BaseParserLibrary.extractUInt32(src, dataOffset);
        require(
            rClaims.chainId > 0,
            string(abi.encodePacked(RClaimsParserLibraryErrorCodes.RCLAIMSPARSERLIB_CHAINID_ZERO))
        );
        rClaims.height = BaseParserLibrary.extractUInt32(src, dataOffset + 4);
        require(
            rClaims.height > 0,
            string(abi.encodePacked(RClaimsParserLibraryErrorCodes.RCLAIMSPARSERLIB_HEIGHT_ZERO))
        );
        rClaims.round = BaseParserLibrary.extractUInt32(src, dataOffset + 8);
        require(
            rClaims.round > 0,
            string(abi.encodePacked(RClaimsParserLibraryErrorCodes.RCLAIMSPARSERLIB_ROUND_ZERO))
        );
        rClaims.prevBlock = BaseParserLibrary.extractBytes32(src, dataOffset + 24);
    }
}
