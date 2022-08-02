// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";

library MerkleProofLibrary {
    /// @notice Check if the bit at the given `index` in `self` is set. Function
    /// used to decode the bitmap, i.e, knowing when to use  a leaf node or a
    // default leaf node hash when reconstructing the proof.
    /// @param self the input bitmap as bytes
    /// @param index the index to check if it's set
    /// @return `true` if the value of the bit is `1`, `false` if the value of the bit is `0`
    function bitSet(bytes memory self, uint16 index) internal pure returns (bool) {
        uint256 val;
        assembly {
            val := shr(sub(255, index), and(mload(add(self, 0x20)), shl(sub(255, index), 1)))
        }
        return val == 1;
    }

    /// @notice Check if the bit at the given `index` in `self` is set. Similar
    //  to `bitSet(bytes)` but used to decide which side of the binary tree to
    //  follow using the key when reconstructing the merkle proof.
    /// @param self the input bitmap as bytes32 / @param index the index to
    ///check if it's set
    /// @return `true` if the value of the bit is `1`, `false` if the value of
    /// the bit is `0`
    function bitSetBytes32(bytes32 self, uint16 index) internal pure returns (bool) {
        uint256 val;
        assembly {
            val := shr(sub(255, index), and(self, shl(sub(255, index), 1)))
        }
        return val == 1;
    }

    /// @notice Computes the leaf hash.
    /// @param key the key/UTXOID
    /// @param value the value
    /// @param proofHeight the proof height (number of elements in the uncompressed merkle proof)
    /// @return the leaf hash
    function computeLeafHash(
        bytes32 key,
        bytes32 value,
        uint16 proofHeight
    ) internal pure returns (bytes32) {
        require(
            proofHeight <= 256 && proofHeight >= 0,
            "MerkleProofLibrary: Invalid proofHeight, should be [0, 256]"
        );
        return keccak256(abi.encodePacked(key, value, uint8(256 - proofHeight)));
    }

    /// @notice Checks if `proof` is a valid inclusion proof.
    /// @param _proof the merkle proof (audit path)
    /// @param root the root of the tree
    function verifyInclusion(MerkleProofParserLibrary.MerkleProof memory _proof, bytes32 root)
        internal
        pure
    {
        require(_proof.proofValue != 0, "MerkleProofLibrary: Invalid Inclusion Merkle proof!");
        bytes32 _keyHash = computeLeafHash(_proof.key, _proof.proofValue, _proof.keyHeight);
        bool result = checkProof(
            _proof.auditPath,
            root,
            _keyHash,
            _proof.key,
            _proof.bitmap,
            _proof.keyHeight
        );
        require(result, "MerkleProofLibrary: The proof doesn't match the root of the trie!");
    }

    /// @notice Checks if `proof` is a valid non-inclusion proof.
    /// @param _proof the merkle proof (audit path)
    /// @param root the root of the tree
    function verifyNonInclusion(MerkleProofParserLibrary.MerkleProof memory _proof, bytes32 root)
        internal
        pure
    {
        if (_proof.proofKey == 0 && _proof.proofValue == 0) {
            // Non-inclusion default value
            bytes32 _keyHash = bytes32(
                hex"bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"
            );
            bool result = checkProof(
                _proof.auditPath,
                root,
                _keyHash,
                _proof.key,
                _proof.bitmap,
                _proof.keyHeight
            );
            require(result, "MerkleProofLibrary: Default leaf not found in the key's path!");
        } else if (_proof.proofKey != 0 && _proof.proofValue != 0) {
            // Non-inclusion leaf node
            bytes32 _keyHash = computeLeafHash(
                _proof.proofKey,
                _proof.proofValue,
                _proof.keyHeight
            );
            bool result = checkProof(
                _proof.auditPath,
                root,
                _keyHash,
                _proof.key,
                _proof.bitmap,
                _proof.keyHeight
            );
            require(
                result,
                "MerkleProofLibrary: The Leaf node provided was not found in the key's path!"
            );
        } else {
            // _proof.proofKey != 0 && _proof.proofValue == 0 or _proof.proofKey == 0 && _proof.proofValue != 0
            revert("MerkleProofLibrary: Invalid Non Inclusion Merkle proof!");
        }
    }

    /// @notice Checks if `proof` is a valid inclusion proof.
    /// @param auditPath the audit path to reconstruct the proof
    /// @param root the root of the tree
    /// @param keyHash the leaf hash used to reconstruct the proof
    /// @param key the key of the transaction
    /// @param bitmap the bitmap of the compact merkle proof
    /// @param proofHeight the height of the proof
    /// @return `true` if the proof is valid, `false` otherwise
    function checkProof(
        bytes memory auditPath,
        bytes32 root,
        bytes32 keyHash,
        bytes32 key,
        bytes memory bitmap,
        uint16 proofHeight
    ) internal pure returns (bool) {
        bytes32 el;
        bytes32 h = keyHash;

        bytes32 defaultLeaf = bytes32(
            hex"bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"
        );

        uint16 proofIdx = 0;
        require(
            proofHeight >= 0 && proofHeight <= 256,
            "MerkleProofLibrary: proofHeight should be in the range [0, 256]"
        );
        for (uint256 i = 0; i < proofHeight; i++) {
            if (bitSet(bitmap, uint16(i))) {
                proofIdx += 32;
                assembly {
                    el := mload(add(auditPath, proofIdx))
                }
            } else {
                el = defaultLeaf;
            }

            if (bitSetBytes32(key, proofHeight - 1 - uint16(i))) {
                h = keccak256(abi.encodePacked(el, h));
            } else {
                h = keccak256(abi.encodePacked(h, el));
            }
        }
        return h == root;
    }
}
