// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library SnapshotsErrors {
    error OnlyValidatorsAllowed(address caller);
    error ConsensusNotRunning();
    error MinimumBlocksIntervalNotPassed(
        uint256 currentBlocksInterval,
        uint256 minimumBlocksInterval
    );
    error InvalidMasterPublicKey();
    error SignatureVerificationFailed();
    error InvalidBlockHeight(uint256 blockHeight);
    error InvalidChainId(uint256 chainId);
    error MigrationNotAllowedAtCurrentEpoch();
    error MigrationInputDataMismatch(uint256 groupSignatureLength, uint256 bClaimsLength);
}
