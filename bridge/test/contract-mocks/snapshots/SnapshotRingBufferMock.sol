//SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/libraries/snapshots/SnapshotsStorage.sol";
import "contracts/libraries/snapshots/SnapshotRingBuffer.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";

//epochfor is the name of the function signature not the function you pass in
contract SnapshotRingBufferMock is SnapshotRingBuffer, SnapshotsStorage {
    //the actual
    SnapshotBuffer internal _snapshotBuffer;
    Snapshot internal RingBufferHead;
    Snapshot internal RingBufferTail;
    Snapshot internal RingBufferInsert;
    using RingBuffer for SnapshotBuffer;
    using EpochLib for Epoch;

    constructor() SnapshotsStorage(1337, 1024) {}

    function decodeBClaims(bytes calldata data_)
        public
        pure
        returns (BClaimsParserLibrary.BClaims memory)
    {
        BClaimsParserLibrary.BClaims memory blockClaims = BClaimsParserLibrary.extractBClaims(
            data_
        );
        return blockClaims;
    }

    function createBClaims(
        uint32 chainId,
        uint32 height,
        uint32 txCount,
        bytes32 prevBlock,
        bytes32 txRoot,
        bytes32 stateRoot,
        bytes32 headerRoot
    ) public pure returns (BClaimsParserLibrary.BClaims memory) {
        return
            BClaimsParserLibrary.BClaims(
                chainId,
                height,
                txCount,
                prevBlock,
                txRoot,
                stateRoot,
                headerRoot
            );
    }

    function createSnapshot(uint256 committedAt, BClaimsParserLibrary.BClaims memory bClaims_)
        public
        pure
        returns (Snapshot memory)
    {
        return Snapshot(committedAt, bClaims_);
    }

    function getSnapshotBuffer() public view returns (SnapshotBuffer memory) {
        return _snapshotBuffer;
    }

    //get test for the library
    function getSnapshotCheck(uint32 epoch_) public view returns (Snapshot memory) {
        return _snapshotBuffer.get(epoch_);
    }

    //_indexFor test
    function checkIndexFor(uint32 epoch_) public view returns (uint256) {
        return _snapshotBuffer.indexFor(epoch_);
    }

    //check for the unsafeSet, we arent testing for this anymore
    function unsafeSetCheck(Snapshot memory _insert, uint32 epoch) public {
        _snapshotBuffer.unsafeSet(_insert, epoch);
    }

    //regular set test
    function safeSet(Snapshot memory insert) public {
        _snapshotBuffer.set(super._getEpochFromHeight, insert);
        //check for buffer overflow if function
        assert(_snapshotBuffer._array.length < 7);
    }

    //get tail function
    function getTail() public view returns (Snapshot memory) {
        return _snapshotBuffer._array[5];
    }

    //get head function
    function getHead() public view returns (Snapshot memory) {
        return _snapshotBuffer._array[0];
    }

    //get height function
    function getHeight(Snapshot memory snapshot) public view returns (uint32) {
        //check later to see what's in here to get the u32
        return super._getEpochFromHeight(snapshot.blockClaims.height);
    }
}
