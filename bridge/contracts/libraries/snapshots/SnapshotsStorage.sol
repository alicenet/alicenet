// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;


import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/utils/ImmutableAuth.sol";


abstract contract SnapshotsStorage is ImmutableETHDKG, ImmutableValidatorPool {

    uint256 internal immutable _epochLength;

    uint256 internal immutable _chainId;

    uint32 internal _epoch;

    // after how many eth blocks of not having a snapshot will we start allowing more validators to
    // make it
    uint32 internal _snapshotDesperationDelay;

    // how quickly more validators will be allowed to make a snapshot, once
    // _snapshotDesperationDelay has passed
    uint32 internal _snapshotDesperationFactor;

    mapping(uint256 => Snapshot) internal _snapshots;

    constructor(uint256 chainId_, uint256 epochLength_) ImmutableFactory(msg.sender) ImmutableETHDKG() ImmutableValidatorPool() {
        _chainId = chainId_;
        _epochLength = epochLength_;
    }
}