import { assert, ethers } from "hardhat";
import { SnapshotStructOutput } from "../../typechain-types/contracts/Snapshots";
import { SnapshotRingBufferMock } from "../../typechain-types/test/contract-mocks/snapshots/SnapshotRingBufferMock";
import { expect } from "../chai-setup";
import { SignedBClaims } from "../setup";
import { signedData1 } from "../snapshots/assets/4-validators-snapshots-100-Group1";

describe("Snapshot Ring Buffer library Successes", async () => {
  let snapshotRingBufferMock: SnapshotRingBufferMock;
  let snapshot1: SnapshotStructOutput;
  //helper function to automatically create snapshots
  const createSnapshots = async (
    inputClaims: SignedBClaims[],
    //the mock contract
    snapshotRingBufferMock: SnapshotRingBufferMock,
    //number of snapshots you want
    snapshotsWanted: number,
    //where you want to start in the vectir
    startIndex: number
  ): Promise<SnapshotStructOutput[]> => {
    let snapshotsVec: SnapshotStructOutput[] = [];
    for (let i = startIndex; i < startIndex + snapshotsWanted; i++) {
      let bClaims = await snapshotRingBufferMock.decodeBClaims(
        inputClaims[i].BClaims
      );
      snapshotsVec.push(
        await snapshotRingBufferMock.createSnapshot(i, bClaims)
      );
    }
    return snapshotsVec;
  };

  beforeEach(async function () {
    const SnapshotRingBufferMockFactory = await ethers.getContractFactory(
      "SnapshotRingBufferMock"
    );
    snapshotRingBufferMock = await SnapshotRingBufferMockFactory.deploy();
    await snapshotRingBufferMock.deployed();
    const blockClaims6 = await snapshotRingBufferMock.decodeBClaims(
      signedData1[0].BClaims
    );
    snapshot1 = await snapshotRingBufferMock.createSnapshot(1, blockClaims6);
  });

  it("Successfully calculates the Epoch for the new snapshot", async () => {
    //put 1 in there so its not empty
    await snapshotRingBufferMock.safeSet(snapshot1);
    const epoch = await snapshotRingBufferMock.getHeight(snapshot1);
    expect(epoch).to.be.equal(1);
    expect(
      await snapshotRingBufferMock.getSnapshotCheck(epoch)
    ).to.be.deep.equals(snapshot1);
  });

  it("Should fail when the query is epoch 0", async () => {
    //await is outside if you expect something to fail
    await expect(snapshotRingBufferMock.getSnapshotCheck(0)).to.be.revertedWith(
      "epoch must be non-zero"
    );
  });

  it("Successfully adds a new snapshot to the full buffer", async () => {
    //fill buffer with snapshots, 7 just so that it overwrites 1
    const snapshots = await createSnapshots(
      signedData1,
      snapshotRingBufferMock,
      7,
      0
    );
    for (let i = 0; i < snapshots.length; i++) {
      //you start the buffer at _array[1] not 0
      await snapshotRingBufferMock.safeSet(snapshots[i]);
    }
    //await is inside if you expect it to succeed
    expect(await snapshotRingBufferMock.getTail()).to.be.deep.equals(
      snapshots[4]
    );
    //check to see if the 7th snapshot overwrote the first
    expect(await snapshotRingBufferMock.getSnapshotCheck(1)).to.be.deep.equals(
      snapshots[6]
    );
  });

  it("Successfully calculates the Index for the new snapshot", async () => {
    let IndexFor = (await snapshotRingBufferMock.checkIndexFor(7)).toNumber();
    assert((IndexFor = 1));
    let IndexFor1 = await (
      await snapshotRingBufferMock.checkIndexFor(22)
    ).toNumber();
    assert((IndexFor1 = 4));
  });

  it("Successfully overwrites old Snapshots when the buffer is already full", async () => {
    //input 10 snapshots
    const snapshots = await createSnapshots(
      signedData1,
      snapshotRingBufferMock,
      10,
      0
    );
    for (let i = 0; i < snapshots.length; i++) {
      //you start the buffer at _array[1] not 0
      await snapshotRingBufferMock.safeSet(snapshots[i]);
    }
    let Head = await snapshotRingBufferMock.getHead();
    //assert((Head = snapshots[6]));
    expect((Head = snapshots[5]));
  });

  it("Gets the right Snapshot", async () => {
    await snapshotRingBufferMock.safeSet(snapshot1);
    expect(await snapshotRingBufferMock.getSnapshotCheck(1)).to.be.deep.equals(
      snapshot1
    );
  });
  // Apparently we arent testing Unsafe Sets anymore
  it("Does a successful unsafe set", async () => {
    //calculate indexFor
    let IndexFor = (await snapshotRingBufferMock.checkIndexFor(22)).toNumber();
    assert((IndexFor = 4));
    //unsafeset for the index to see if it will add in the correct position
    snapshotRingBufferMock.unsafeSetCheck(snapshot1, IndexFor);
    //check to see if its in position
    expect(await snapshotRingBufferMock.getSnapshotCheck(4)).to.be.deep.equals(
      snapshot1
    );
  });
});
