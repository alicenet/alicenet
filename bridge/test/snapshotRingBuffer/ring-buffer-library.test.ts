 import { assert } from "console";
import { ethers } from "hardhat";
import { setup } from "mocha";
 import { SnapshotRingBufferMock__factory, Snapshots} from "../../typechain-types";
 import { expect } from "../chai-setup";
 import { completeETHDKGRound } from "../ethdkg/setup";
 import {validSnapshot1024, validSnapshot2048, 
  validSnapshot7168, validSnapshot8192
 } from "../snapshots/assets/4-validators-snapshots-1";
 import {Snapshot} from "../setup";





  beforeEach(async function () {
    const SnapshotRingBufferMockFactory = await ethers.getContractFactory("SnapshotRingBufferMock");
    const SnapshotRingBufferMock = await SnapshotRingBufferMockFactory.deploy();
    await SnapshotRingBufferMock.deployed();
  });




describe("Snapshot Ring Buffer library Successes", async() => {
  const BlockInterval = 1024;
  //fix imports later for dummy mock data
  const testBuffer = SnapshotRingBufferMock.createBuffer();
  const blockClaims1 = BClaimsParserLibrary.BClaims(0,0,0,0x00,0x00,0x00,0x00);
  const blockClaims2 = BClaimsParserLibrary.BClaims(1,1,1,0x01,0x01,0x01,0x01);
  const blockClaims3 = BClaimsParserLibrary.BClaims(2,2,2,0x02,0x02,0x02,0x02);
  const blockClaims4 = BClaimsParserLibrary.BClaims(3,3,3,0x03,0x03,0x03,0x03);
  const blockClaims5 = BClaimsParserLibrary.BClaims(4,4,4,0x04,0x04,0x04,0x04);
  const blockClaims6 = BClaimsParserLibrary.BClaims(5,5,5,0x05,0x05,0x05,0x05);
  const snapshot1 = validSnapshot1024;
  const snapshot2 = validSnapshot2048;
  const snapshot3 = validSnapshot7168;
  const snapshot4 = validSnapshot8192;
  const snapshot5 = new Snapshot(blockClaims5);
  const snapshot6 = new Snapshot(blockClaims6);
  const _snapshotInsert = new Snapshot(7, blockClaims1);



  it("Successfully adds a new snapshot to the full buffer", async () => {
    const Blockinterval = 1024;
    //fill buffer with snapshots
    SnapshotRingBufferMock.SafeSet(snapshot1, testBuffer);
    SnapshotRingBufferMock.SafeSet(snapshot2, testBuffer);
    SnapshotRingBufferMock.SafeSet(snapshot3, testBuffer);
    SnapshotRingBufferMock.SafeSet(snapshot4, testBuffer);
    SnapshotRingBufferMock.SafeSet(snapshot5, testBuffer);
    SnapshotRingBufferMock.SafeSet(snapshot6, testBuffer);
    const RingBufferTail = SnapshotRingBufferMock.getTail(testBuffer);
    //get the oldHead to compare
    const oldHead = SnapshotRingBuffermock.getHead();
        //check to see if the array is null
        assert(testBuffer._array.length != 0);
        //calculate the position for where it goes
        let IndexFor = SnapshotRingBufferMock.checkIndexFor(_snapshotRingBuffer, 7);
        //puts the insert into the buffer
        SnapshotRingBufferMock.SafeSet(_snapshotInsert,testBuffer);
        //compare the temporary variable to the new snapshot to see if it got replaced
        assert(testBuffer._array[0] != oldHead);
        //see if it didnt just overwrite the tail
        assert(testBuffer._array[5] != _snapshotInsert);
        //assert the size of the ringbuffer is 6
        assert(testBuffer._array.length < 6);
        
      })

  it("Successfully calculates the Index for the new snapshot", async() =>{
    //set it with only 3 snapshots next one should be 4
    SnapshotRingBufferMock.SafeSet(testBuffer, snapshot1);
    SnapshotRingBufferMock.SafeSet(testBuffer, snapshot2);
    SnapshotRingBufferMock.SafeSet(testBuffer, snapshot3);
    //set temp variable for what the Index should be
    let IndexFor = SnapshotRingBufferMock.checkIndexFor(testBuffer, 7);
    assert(IndexFor == 1);
    const IndexFor1 = SnapshotRingBufferMock.checkIndexFor(testBuffer, 22);
    assert(IndexFor = 4);
    //unsafeset for the index to see if it will add in the correct position
    SnapshotRingBufferMock.unsafeSetCheck(testBuffer, snapshot1, IndexFor1);
    //it should be the length of 4 if the index is right
    assert(testBuffer._array.length = 4);
  });

  it("Successfully recovers the last Snapshots added"){
    //input 10 snapshots

    //try to get 6 snapshots from the buffer

    //assert that the first 4 are not in the buffer anymore


  }

  it("Gets the right Snapshot", async() => {
    //fill with 3 snapshots
    SnapshotRingBufferMock.SafeSet(testBuffer, snapshot1);
    SnapshotRingBufferMock.SafeSet(testBuffer, snapshot1);
    SnapshotRingBufferMock.SafeSet(testBuffer, snapshot1);
    //get the third
    let check = SnapshotRingBufferMock.getSnapshotCheck(testBuffer, 3);
    //check if its the third
    assert(check == snapshot3);
  })

  it("Does a successful (safe) set", async() =>{

    SnapshotRingBufferMock.SafeSet(testBuffer, snapshot1);
    //set the temp variable to whatever is in the calculated position
    

    //

  });

  it("Does a successful unsafe set", async() => {
    //calculate indexFor
    let IndexFor = SnapshotRingBufferMock.checkIndexFor(testBuffer, 22);
    assert(IndexFor = 4);
    //unsafeset for the index to see if it will add in the correct position
    SnapshotRingBufferMock.unsafeSetCheck(testBuffer, snapshot1, IndexFor1);
    //get the snapshot
    let check = SnapshotRingBufferMock.getSnapshotCheck(testBuffer, 4);
    //check to see if its in position
    assert(testBuffer._array[3] == snapshot1);
  });

});













 
 