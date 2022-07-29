// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library MerkleProofLibraryErrors {
    error InvalidProofHeight(uint256 proofHeight);
    error InclusionZero();
    error ProofDoesNotMatchTrieRoot();
    error DefaultLeafNotFoundInKeyPath();
    error ProvidedLeafNotFoundInKeyPath();
    error InvalidNonInclusionMerkleProof();
}
