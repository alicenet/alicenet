//SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/libraries/snapshots/SnapshotsStorage.sol";
import "contracts/libraries/snapshots/SnapshotRingBuffer.sol";


//epochfor is the name of the function signature not the function you pass in
contract SnapshotRingBufferMock is SnapshotRingBuffer, SnapshotsStorage {
    //the actual
    SnapshotBuffer internal _snapshotBuffer;
    Snapshot internal RingBufferHead;
    Snapshot internal RingBufferTail;
    Snapshot internal RingBufferInsert;
    using RingBuffer for SnapshotBuffer;

    constructor() SnapshotsStorage(1337,1024) {
    }

    function getSnapshotBuffer()public view returns(SnapshotBuffer memory){
        return _snapshotBuffer;
    }

    //get test for the library
    function getSnapshotCheck(SnapshotBuffer storage _snapshotBuffer, uint32 epoch_ ) internal view returns(Snapshot memory) {
        return RingBuffer.get(_snapshotBuffer, epoch_);
    }
    //_indexFor test
    function checkIndexFor(SnapshotBuffer storage _snapshotBuffer, uint32 epoch_) internal view returns(uint256){
        return RingBuffer._indexFor(_snapshotBuffer, epoch_);
    }
    //check for the unsafeSet
    function unsafeSetCheck(SnapshotBuffer storage _snapshotBuffer, Snapshot memory _insert, uint32 epoch) internal {
        RingBuffer.unsafeSet(_snapshotBuffer, _insert, epoch);
    }
    //set test with a check inside of it
   /* function SuccessfulInsert(Snapshot storage _insert, SnapshotBuffer storage _snapshotBuffer) internal returns(bool ok) {
        RingBufferInsert = _insert;
        RingBuffer.set(_snapshotBuffer, super._getEpochFromHeight, RingBufferInsert);
        assert(_snapshotBuffer._array.length < 7);
        for (uint i = 0; i < _snapshotBuffer._array.length; i++){
            if (_snapshotBuffer._array[i] == _insert){
                ok == true;
                return ok;
            } 
        }
    }*/
    //regular set test
    function SafeSet(Snapshot storage _insert, SnapshotBuffer storage _snapshotBuffer) internal{
        RingBuffer.set(_snapshotBuffer, super._getEpochFromHeight, _insert);
        //check for buffer overflow if function 
        assert(_snapshotBuffer._array.length < 7);
    }

    /*function EmptyInsertSetCheck(SnapshotBuffer storage _snapshotBuffer) internal {
        require(_snapshotBuffer._array.length == 0, "Not an empty buffer");
        //add dummy data later
        RingBufferInsert = new Snapshot();
        //use the contract's functions to set
        snapshotRingBuffer.set(_snapshotBuffer, super._getEpochFromHeight, RingBufferInsert);
        assert (_snapshotBuffer._array.length == 1);
    }

    function getSnapshot()public returns (bool ok, Snapshot memory _snapshot){
        //put in mock data later
        RingBufferInsert = new Snapshot();
        //get the epoch return from the set function
        uint epochInsert = snapshotRingBuffer._setSnapshot(RingBufferInsert);
        //checks to see if it returns from the insert
        Snapshot memory temp = snapshotRingBuffer._getSnapshot(epochInsert);
        return temp;
    }
    
    //set function is doing dirty writes its not knocking out the first element
    //indexFor is circular buffer thats why it has the modulus
    //make a test for the epoch 200
    //try to fill it with 10 then recover the first 6
    //7 snapshot is going to be position 1, it doesnt shift down
    function RingBufferKnockout(SnapshotBuffer storage _snapshotBuffer, Snapshot memory _snapshot) internal {
        //null check
        require(_snapshotBuffer._array.length != 0, "Its an empty snapshot buffer");
        //checks to see if its full for someone to knockout
        require(_snapshotBuffer._array.length == 6, "The Snapshot Buffer isnt full to play Knockout");
        //insert mockdata for snapshot later set it to be the latest
        //making a snapshot go to snapshot mock line 87, blockclaims is defined right above
        and block.number is passed in any uint256, or it already knows do later, dont need to pass in 
        block number can pass in any integer
        
        RingBufferInsert = new Snapshot();
        //assign the head to the first one in the buffer
        RingBufferHead = _snapshotBuffer[0];
        RingBufferTail = _snapshotBuffer[5];
        snapshotRingBuffer.set(_snapshotBuffer, super._getEpochFromHeight, RingBufferInsert);
        assert(_snapshotBuffer[0] != RingBufferHead, "Head has not been replaced");
        assert(_snapshotBuffer[5] != RingBufferTail, "Ring Buffer Tail has not shifted");
        assert(_snapshotBuffer[4] == RingBufferTail, "Ring Buffer tail is in the wrong place");
        assert(_snapshotBuffer[5] == RingBufferInsert, "Insert is not at the end");
    }*/
/*
    function getLatestCheck(SnapshotBuffer storage _snapshotBuffer) internal returns (Snapshot memory){
        require(_snapshotBuffer._array.length != 0, "The buffer is empty");
        Snapshot memory check = _snapshotBuffer._getLatestSnapshot();
        return check;
    }*/

    function EpochIntervalCheck(SnapshotBuffer storage _buffer)internal view returns(bool ok){
        //check to see if the minimum snapshot interval is met
        uint256 interval = 1024;
        //run the check through the array see if their block 
        for (uint i = 0; i < _buffer._array.length; i++){
            //fix later, i+1 null check
            //while(_buffer[i+1] != null)
            assert(_buffer._array[i+1].blockClaims.height - _buffer._array[i].blockClaims.height > interval);
            return true;
        }
    }

    //get tail function
    function getTail(SnapshotBuffer storage _buffer)internal view returns (Snapshot memory){
        return _buffer._array[5];
    }
        //get head function
    function getHead(SnapshotBuffer storage _buffer)internal view returns (Snapshot memory) {
        return _buffer._array[0];
    }

    //get height function
    function getHeight(Snapshot memory _snapshot)public view returns (uint32){
        //check later to see what's in here to get the u32
        return super._getEpochFromHeight(_snapshot.blockClaims.height);
    }

}