// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/libraries/snapshots/SnapshotRingBuffer.sol";

contract SnapshotsRingBufferMock is SnapshotRingBuffer {
    using EpochLib for Epoch;
    using RingBuffer for SnapshotBuffer;
    uint256 internal constant _epochLength = 1024;
    //epoch counter wrapped in a struct
    Epoch internal _epoch;
    //new snapshot ring buffer
    SnapshotBuffer internal _snapshots;

    function unsafeSet(Snapshot memory new_, uint32 epoch_) public {
        _snapshots.unsafeSet(new_, epoch_);
    }

    function setSnapshot(Snapshot memory snapshot_) public returns (uint32) {
        return _setSnapshot(snapshot_);
    }

    function setEpoch(uint32 epoch_) public {
        _epoch.set(epoch_);
    }

    function getEpoch() public view returns (uint32) {
        return _epoch.get();
    }

    function getSnapshot(uint32 epoch_) public view returns (Snapshot memory) {
        return _getSnapshot(epoch_);
    }

    function getLatestSnapshot() public view returns (Snapshot memory) {
        return _getLatestSnapshot();
    }

    function getSnapshots() public view returns (SnapshotBuffer memory) {
        return _getSnapshots();
    }

    function getEpochRegister() public view returns (Epoch memory) {
        return _epochRegister();
    }

    function _getEpochFromHeight(uint32 height_) internal pure override returns (uint32) {
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

    function _epochRegister() internal view override returns (Epoch storage) {
        return _epoch;
    }
}
