// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/ISnapshots.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/utils/ImmutableAuth.sol";

contract SnapshotsMock is ImmutableValidatorPool, ISnapshots {
    uint32 internal _epoch;
    uint32 internal _epochLength;

    // after how many eth blocks of not having a snapshot will we start allowing more validators to
    // make it
    uint32 internal _snapshotDesperationDelay;
    // how quickly more validators will be allowed to make a snapshot, once
    // _snapshotDesperationDelay has passed
    uint32 internal _snapshotDesperationFactor;

    mapping(uint256 => Snapshot) internal _snapshots;

    address internal _admin;
    uint256 internal immutable _chainId;

    constructor(uint32 chainID_, uint32 epochLength_)
        ImmutableFactory(msg.sender)
        ImmutableValidatorPool()
    {
        _admin = msg.sender;
        _chainId = chainID_;
        _epochLength = epochLength_;
    }

    modifier onlyAdmin() {
        require(msg.sender == _admin, "Snapshots: Only admin allowed!");
        _;
    }

    function setEpochLength(uint32 epochLength_) external {
        _epochLength = epochLength_;
    }

    function setSnapshotDesperationDelay(uint32 desperationDelay_) public onlyAdmin {
        _snapshotDesperationDelay = desperationDelay_;
    }

    function getSnapshotDesperationDelay() public view returns (uint256) {
        return _snapshotDesperationDelay;
    }

    function setSnapshotDesperationFactor(uint32 desperationFactor_) public onlyAdmin {
        _snapshotDesperationFactor = desperationFactor_;
    }

    function getSnapshotDesperationFactor() public view returns (uint256) {
        return _snapshotDesperationFactor;
    }

    function getChainId() public view returns (uint256) {
        return _chainId;
    }

    function getEpoch() public view returns (uint256) {
        return _epoch;
    }

    function getEpochLength() public view returns (uint256) {
        return _epochLength;
    }

    function getChainIdFromSnapshot(uint256 epoch_) public view returns (uint256) {
        return _snapshots[epoch_].blockClaims.chainId;
    }

    function getChainIdFromLatestSnapshot() public view returns (uint256) {
        return _snapshots[_epoch].blockClaims.chainId;
    }

    function getBlockClaimsFromSnapshot(uint256 epoch_)
        public
        view
        returns (BClaimsParserLibrary.BClaims memory)
    {
        return _snapshots[epoch_].blockClaims;
    }

    function getBlockClaimsFromLatestSnapshot()
        public
        view
        returns (BClaimsParserLibrary.BClaims memory)
    {
        return _snapshots[_epoch].blockClaims;
    }

    function getCommittedHeightFromSnapshot(uint256 epoch_) public view returns (uint256) {
        return _snapshots[epoch_].committedAt;
    }

    function getCommittedHeightFromLatestSnapshot() public view returns (uint256) {
        return _snapshots[_epoch].committedAt;
    }

    function getMadnetHeightFromSnapshot(uint256 epoch_) public view returns (uint256) {
        return _snapshots[epoch_].blockClaims.height;
    }

    function getMadnetHeightFromLatestSnapshot() public view returns (uint256) {
        return _snapshots[_epoch].blockClaims.height;
    }

    function getSnapshot(uint256 epoch_) public view returns (Snapshot memory) {
        return _snapshots[epoch_];
    }

    function getLatestSnapshot() public view returns (Snapshot memory) {
        return _snapshots[_epoch];
    }

    function snapshot(bytes calldata groupSignature_, bytes calldata bClaims_)
        public
        returns (bool)
    {
        bool isSafeToProceedConsensus = true;
        if (IValidatorPool(_ValidatorPoolAddress()).isMaintenanceScheduled()) {
            isSafeToProceedConsensus = false;
            IValidatorPool(_ValidatorPoolAddress()).pauseConsensus();
        }
        // dummy to silence compiling warnings
        groupSignature_;
        bClaims_;
        _epoch++;
        return true;
    }

    function isMock() public pure returns (bool) {
        return true;
    }

    function setCommittedHeightFromLatestSnapshot(uint256 height_) public returns (uint256) {
        _snapshots[_epoch].committedAt = height_;
        return height_;
    }

    function mayValidatorSnapshot(
        uint256 numValidators,
        uint256 myIdx,
        uint256 blocksSinceDesperation,
        bytes32 blsig,
        uint256 desperationFactor
    ) public pure returns (bool) {
        numValidators;
        myIdx;
        blocksSinceDesperation;
        blsig;
        desperationFactor;
        return true;
    }
}
