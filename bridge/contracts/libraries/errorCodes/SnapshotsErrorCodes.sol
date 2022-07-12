// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library SnapshotsErrorCodes {
    // Snapshot error codes
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
}
