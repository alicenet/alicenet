import { expect } from "chai";
import { ethers } from "hardhat";
import { SnapshotsRingBufferMock } from "../../typechain-types";
import { signedData1 } from "../snapshots/assets/4-validators-snapshots-100-Group1";
describe("Ring Buffer Library", async () => {
  let ringBuffer: SnapshotsRingBufferMock;
  beforeEach(async () => {
    const ringBufferBase = await ethers.getContractFactory(
      "SnapshotsRingBufferMock"
    );
    ringBuffer = await ringBufferBase.deploy();
  });

  it("sets the epoch", async () => {
    let epoch = await ringBuffer.getEpoch();
    expect(epoch).to.equal(0);
    const txResponse = await ringBuffer.setEpoch(1);
    await txResponse.wait();
    epoch = await ringBuffer.getEpoch();
    expect(epoch).to.equal(1);
  });

  it("stores a snapshot on the ring buffer", async () => {
    let txResponse = await ringBuffer.setSnapshot(signedData1[0].BClaims);
    await txResponse.wait();
    txResponse = await ringBuffer.setEpoch(1);
    await txResponse.wait();
    const snapshot = await ringBuffer.getLatestSnapshot();
    expect(snapshot.blockClaims.height).to.eq(1024);
  });

  it("stores 7 snapshot on the ring buffer", async () => {
    for (let i = 1; i <= 7; i++) {
      await ringBuffer.setSnapshot(signedData1[i - 1].BClaims);
      await ringBuffer.setEpoch(i);
    }
    const epoch = await ringBuffer.getEpoch();
    expect(epoch).to.equal(7);
    const snapshot = await ringBuffer.getSnapshot(7);
    expect(snapshot.blockClaims.height).to.eq(1024 * 7);
  });

  it("attempts to get a snapshot that is no longer in the ring buffer", async () => {
    for (let i = 1; i <= 7; i++) {
      await ringBuffer.setSnapshot(signedData1[i - 1].BClaims);
      await ringBuffer.setEpoch(i);
    }
    const epoch = await ringBuffer.getEpoch();
    expect(epoch).to.equal(7);
    const snapshot = ringBuffer.getSnapshot(1);
    await expect(snapshot)
      .to.be.revertedWithCustomError(ringBuffer, "SnapshotsNotInBuffer")
      .withArgs(1);
  });
  it("get the 0 snapshot", async () => {
    for (let i = 1; i <= 7; i++) {
      await ringBuffer.setSnapshot(signedData1[i - 1].BClaims);
      await ringBuffer.setEpoch(i);
    }
    const epoch = await ringBuffer.getEpoch();
    expect(epoch).to.equal(7);
    const snapshot = await ringBuffer.getSnapshot(0);
    expect(snapshot.committedAt).to.eq(0);
    expect(snapshot.blockClaims.height).to.eq(0);
    expect(snapshot.blockClaims.chainId).to.eq(0);
  });
});
