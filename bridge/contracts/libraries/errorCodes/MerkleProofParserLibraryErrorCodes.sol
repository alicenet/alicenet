// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library MerkleProofParserLibraryErrorCodes {
    // MerkleProofParserLibrary error codes
    bytes32 public constant MERKLEPROOFPARSERLIB_INVALID_PROOF_MINIMUM_SIZE = "1200"; //"MerkleProofParserLibrary: Not enough bytes to extract a minimum MerkleProof"
    bytes32 public constant MERKLEPROOFPARSERLIB_INVALID_PROOF_SIZE = "1201"; //"MerkleProofParserLibrary: Not enough bytes to extract a minimum MerkleProof"
    bytes32 public constant MERKLEPROOFPARSERLIB_INVALID_KEY_HEIGHT = "1202"; //"MerkleProofParserLibrary: keyHeight should be in the range [0, 256]"
}
