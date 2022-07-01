// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/snapshots/SnapshotRingBuffer.sol";

abstract contract SnapshotsStorage is ImmutableETHDKG, ImmutableValidatorPool, SnapshotRingBuffer {
    bytes32 internal constant SNAPSHOT_ONLY_VALIDATORS_ALLOWED = "400"; //"Snapshots: Only validators allowed!"
    bytes32 internal constant SNAPSHOT_CONSENSUS_RUNNING = "401"; //"Snapshots: Consensus is not running!"
    bytes32 internal constant SNAPSHOT_MIN_BLOCKS_INTERVAL_NOT_PASSED = "402"; //"Snapshots: Necessary amount of ethereum blocks has not passed since last snapshot!"
    bytes32 internal constant SNAPSHOT_CALLER_NOT_ETHDKG_PARTICIPANT = "403"; //"Snapshots: Caller didn't participate in the last ethdkg round!"
    bytes32 internal constant SNAPSHOT_WRONG_MASTER_PUBLIC_KEY = "404"; //"Snapshots: Wrong master public key!"
    bytes32 internal constant SNAPSHOT_SIGNATURE_VERIFICATION_FAILED = "405"; //"Snapshots: Signature verification failed!"
    bytes32 internal constant SNAPSHOT_INCORRECT_BLOCK_HEIGHT = "406"; //"Snapshots: Incorrect AliceNet height for snapshot!"
    bytes32 internal constant SNAPSHOT_INCORRECT_CHAIN_ID = "407"; //"Snapshots: Incorrect chainID for snapshot!"
    bytes32 internal constant SNAPSHOT_MIGRATION_NOT_ALLOWED = "408"; //Snapshots: Migration only allowed at epoch 0!
    bytes32 internal constant SNAPSHOT_MIGRATION_INPUT_DATA_MISMATCH = "409"; //Snapshots: Mismatch calldata length!
    bytes32 internal constant SNAPSHOT_NOT_IN_BUFFER = "410"; //Snapshots: Snapshot no longer in buffer
    uint256 internal immutable _epochLength;

    uint256 internal immutable _chainId;

    // uint32 internal _epoch;

    // Number of ethereum blocks that we should wait between snapshots. Mainly used to prevent the
    // submission of snapshots in short amount of time by validators that could be potentially being
    // malicious
    uint32 internal _minimumIntervalBetweenSnapshots;

    // after how many eth blocks of not having a snapshot will we start allowing more validators to
    // make it
    uint32 internal _snapshotDesperationDelay;

    // how quickly more validators will be allowed to make a snapshot, once
    // _snapshotDesperationDelay has passed
    uint32 internal _snapshotDesperationFactor;

    //epoch counter wrapped in a struct
    Epoch internal _epoch;
    //new snapshot ring buffer
    SnapshotBuffer internal _snapshots;

    constructor(uint256 chainId_, uint256 epochLength_)
        ImmutableFactory(msg.sender)
        ImmutableETHDKG()
        ImmutableValidatorPool()
    {
        _chainId = chainId_;
        _epochLength = epochLength_;
    }

    function _getEpochFromHeight(uint32 height_) internal view override returns (uint32) {
        if (height_ <= _epochLength) {
            return 1;
        }
        if (height_ % _epochLength == 0) {
            return uint32(height_ / _epochLength);
        }
        return uint32((height_ / _epochLength) + 1);
    }

    function _getSnapshots() internal view override returns (SnapshotBuffer storage) {
        return _snapshots;
    }

    function _epochReg() internal view override returns (Epoch storage) {
        return _epoch;
    }
}
