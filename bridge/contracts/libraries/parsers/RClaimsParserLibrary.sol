// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "./BaseParserLibrary.sol";

/// @title Library to parse the RClaims structure from a blob of capnproto data
library RClaimsParserLibrary {
    /** @dev size in bytes of a RCLAIMS cap'npro structure without the cap'n
      proto header bytes*/
    uint256 internal constant RCLAIMS_SIZE = 56;
    /** @dev Number of bytes of a capnproto header, the data starts after the
      header */
    uint256 internal constant CAPNPROTO_HEADER_SIZE = 8;

    struct RClaims {
        uint32 chainId;
        uint32 height;
        uint32 round;
        bytes32 prevBlock;
    }

    /**
    @notice This function is for deserializing data directly from capnproto
            RClaims. It will skip the first 8 bytes (capnproto headers) and
            deserialize the RClaims Data. If RClaims is being extracted from
            inside of other structure (E.g RCert capnproto) use the
            `extractInnerRClaims(bytes, uint)` instead.
    */
    /// @param src Binary data containing a RClaims serialized struct with Capn Proto headers
    /// @dev Execution cost: 1506 gas
    function extractRClaims(bytes memory src)
        internal
        pure
        returns (RClaims memory rClaims)
    {
        return extractInnerRClaims(src, CAPNPROTO_HEADER_SIZE);
    }

    /**
    @notice This function is for serializing the RClaims struct from an defined
            location inside a binary blob. E.G Extract RClaims from inside of
            other structure (E.g RCert capnproto) or skipping the capnproto
            headers.
    */
    /// @param src Binary data containing a RClaims serialized struct without Capn Proto headers
    /// @param dataOffset offset to start reading the RClaims data from inside src
    /// @dev Execution cost: 1332 gas
    function extractInnerRClaims(bytes memory src, uint256 dataOffset)
        internal
        pure
        returns (RClaims memory rClaims)
    {
        require(
            dataOffset + RCLAIMS_SIZE > dataOffset,
            "RClaimsParserLibrary: Overflow on the dataOffset parameter"
        );
        require(
            src.length >= dataOffset + RCLAIMS_SIZE,
            "RClaimsParserLibrary: Not enough bytes to extract RClaims"
        );
        rClaims.chainId = BaseParserLibrary.extractUInt32(src, dataOffset);
        require(rClaims.chainId > 0, "RClaimsParserLibrary: Invalid parsing. The chainId should be greater than 0!");
        rClaims.height = BaseParserLibrary.extractUInt32(src, dataOffset + 4);
        require(rClaims.height > 0, "RClaimsParserLibrary: Invalid parsing. The height should be greater than 0!");
        rClaims.round = BaseParserLibrary.extractUInt32(src, dataOffset + 8);
        require(rClaims.round > 0, "RClaimsParserLibrary: Invalid parsing. The round should be greater than 0!");
        rClaims.prevBlock = BaseParserLibrary.extractBytes32(
            src,
            dataOffset + 24
        );
    }
}
