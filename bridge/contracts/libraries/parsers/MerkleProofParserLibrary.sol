// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "./BaseParserLibrary.sol";

/// @title Library to parse the MerkleProof structure from a blob of binary data
library MerkleProofParserLibrary {
    /** @dev minimum size in bytes of a MerkleProof binary structure
      (without proofs and bitmap) */
    uint256 internal constant MERKLE_PROOF_SIZE = 103;

    struct MerkleProof {
        bool included;
        uint16 keyHeight;
        bytes32 key;
        bytes32 proofKey;
        bytes32 proofValue;
        bytes bitmap;
        bytes auditPath;
    }

    /**
    @notice This function is for deserializing the MerkleProof struct from a
            binary blob.
    */
    /// @param src Binary data containing a MerkleProof serialized struct
    /// @return mProof a MerkleProof struct
    /// @dev Execution cost: ~4000-51000 gas for a 10-256 height proof respectively
    function extractMerkleProof(bytes memory src)
        internal
        pure
        returns (MerkleProof memory mProof)
    {
        require(
            src.length >= MERKLE_PROOF_SIZE,
            "MerkleProofParserLibrary: Not enough bytes to extract a minimum MerkleProof"
        );
        uint16 bitmapLength = BaseParserLibrary.extractUInt16FromBigEndian(src, 99);
        uint16 auditPathLength = BaseParserLibrary.extractUInt16FromBigEndian(src, 101);
        require(
            src.length >= MERKLE_PROOF_SIZE + bitmapLength + auditPathLength * 32,
            "MerkleProofParserLibrary: Not enough bytes to extract MerkleProof"
        );
        mProof.included = BaseParserLibrary.extractBool(src, 0);
        mProof.keyHeight = BaseParserLibrary.extractUInt16FromBigEndian(src, 1);
        require(mProof.keyHeight >= 0 && mProof.keyHeight <= 256, "MerkleProofParserLibrary: keyHeight should be in the range [0, 256]");
        mProof.key = BaseParserLibrary.extractBytes32(src, 3);
        mProof.proofKey = BaseParserLibrary.extractBytes32(src, 35);
        mProof.proofValue = BaseParserLibrary.extractBytes32(src, 67);
        mProof.bitmap = BaseParserLibrary.extractBytes(src, 103, bitmapLength);
        mProof.auditPath = BaseParserLibrary.extractBytes(src, 103 + bitmapLength, auditPathLength*32);
    }
}
