import { expect } from "chai";
import { contract } from "hardhat";
import { Snapshots } from "../../typechain-types";
import { completeETHDKGRound } from "../ethdkg/setup";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getValidatorEthAccount,
  mineBlocks,
} from "../setup";
import { createValidators, stakeValidators } from "../validatorPool/setup";
import {
  signedData1,
  validatorsSnapshotsG1,
} from "./assets/4-validators-snapshots-100-Group1";
import {
  signedData2,
  validatorsSnapshotsG2,
} from "./assets/4-validators-snapshots-100-Group2";

contract("SnapshotRingBuffer 0state", async () => {
  const epochLength = 1024;
  let fixture: Fixture;
  let snapshots: Snapshots;
  describe("Snapshot upgrade integration", async () => {
    beforeEach(async () => {
      // deploys the new snapshot contract with buffer and zero state
      fixture = await getFixture(true, false);
      await completeETHDKGRound(validatorsSnapshotsG1, {
        ethdkg: fixture.ethdkg,
        validatorPool: fixture.validatorPool,
      });

      snapshots = fixture.snapshots as Snapshots;
    });

    it("adds 6 new snapshots to the snapshot buffer", async () => {
      let epochs = (await fixture.snapshots.getEpoch()).toNumber();
      const signedSnapshots = signedData1;
      const numSnaps = epochs + 6;
      const snapshots = fixture.snapshots.connect(
        await getValidatorEthAccount(validatorsSnapshotsG1[0])
      );
      // take 6 snapshots
      for (let i = epochs + 1; i <= numSnaps; i++) {
        await mineBlocks(
          (await snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
        );
        const contractTx = await snapshots.snapshot(
          signedSnapshots[i - 1].GroupSignature,
          signedSnapshots[i - 1].BClaims,
          { gasLimit: 30000000 }
        );
        await contractTx.wait();
      }
      epochs = (await snapshots.getEpoch()).toNumber();
      const lastSnapshot = await snapshots.getLatestSnapshot();
      expect(lastSnapshot.blockClaims.height).to.equal(epochs * epochLength);
    });

    it("adds 6 new snapshots to the snapshot buffer does a validator change and adds 6 more snapshots", async () => {
      let epochs = (await fixture.snapshots.getEpoch()).toNumber();
      const signedSnapshots = signedData1;
      const validators: Array<string> = [];
      for (const validator of validatorsSnapshotsG1) {
        validators.push(validator.address);
      }

      const snapshotsG1 = fixture.snapshots.connect(
        await getValidatorEthAccount(validatorsSnapshotsG1[0])
      );
      const snapshotsG2 = fixture.snapshots.connect(
        await getValidatorEthAccount(validatorsSnapshotsG2[0])
      );

      // take 6 snapshots
      for (let i = epochs + 1; i <= 6; i++) {
        await mineBlocks(
          (await snapshotsG1.getMinimumIntervalBetweenSnapshots()).toBigInt()
        );
        const contractTx = await snapshotsG1.snapshot(
          signedSnapshots[i - 1].GroupSignature,
          signedSnapshots[i - 1].BClaims,
          { gasLimit: 30000000 }
        );
        await contractTx.wait();
      }
      epochs = (await snapshotsG1.getEpoch()).toNumber();
      const lastSnapshot = await snapshotsG1.getLatestSnapshot();
      expect(lastSnapshot.blockClaims.height).to.equal(epochs * epochLength);
      // schedule maintenace
      await factoryCallAnyFixture(
        fixture,
        "validatorPool",
        "scheduleMaintenance"
      );

      // unregister validators
      await factoryCallAnyFixture(
        fixture,
        "validatorPool",
        "unregisterValidators",
        [validators]
      );
      // register the new validators
      const newValidators = await createValidators(
        fixture,
        validatorsSnapshotsG2
      );
      await stakeValidators(fixture, newValidators);
      await completeETHDKGRound(
        validatorsSnapshotsG2,
        {
          ethdkg: fixture.ethdkg,
          validatorPool: fixture.validatorPool,
        },
        epochs,
        epochs * epochLength,
        (await snapshots.getCommittedHeightFromLatestSnapshot()).toNumber()
      );
      epochs = epochs + 1;
      await mineBlocks(
        (
          await fixture.snapshots.getMinimumIntervalBetweenSnapshots()
        ).toBigInt()
      );
      for (let i = epochs; i <= epochs + 12; i++) {
        await mineBlocks(
          (await snapshotsG2.getMinimumIntervalBetweenSnapshots()).toBigInt()
        );
        const contractTx = await snapshotsG2.snapshot(
          signedData2[i - 1].GroupSignature,
          signedData2[i - 1].BClaims,
          { gasLimit: 30000000 }
        );
        await contractTx.wait();
      }
    });

    it("gets the last snapshot in the buffer", async () => {
      let epochs = (await fixture.snapshots.getEpoch()).toNumber();
      const signedSnapshots = signedData1;
      const validators: Array<string> = [];
      for (const validator of validatorsSnapshotsG1) {
        validators.push(validator.address);
      }
      const snapshotsG1 = fixture.snapshots.connect(
        await getValidatorEthAccount(validatorsSnapshotsG1[0])
      );
      // take 6 snapshots
      for (let i = epochs + 1; i <= 12; i++) {
        await mineBlocks(
          (await snapshotsG1.getMinimumIntervalBetweenSnapshots()).toBigInt()
        );
        const contractTx = await snapshotsG1.snapshot(
          signedSnapshots[i - 1].GroupSignature,
          signedSnapshots[i - 1].BClaims,
          { gasLimit: 30000000 }
        );
        await contractTx.wait();
      }
      epochs = (await snapshotsG1.getEpoch()).toNumber();
      const lastSnapshot = await snapshotsG1.getLatestSnapshot();
      expect(lastSnapshot.blockClaims.height).to.equal(epochs * epochLength);
      const endBuffShot = await snapshots.getSnapshot(epochs - 5);
      expect(endBuffShot.blockClaims.chainId).to.eq(1);
      expect(endBuffShot.blockClaims.height).to.eq((epochs - 5) * epochLength);
    });

    it("getter functions should be always able to work at epoch 0", async () => {
      let epochs = (await fixture.snapshots.getEpoch()).toNumber();
      expect(epochs).to.be.equal(0);
      const lastSnapshot = await fixture.snapshots.getLatestSnapshot();
      expect(lastSnapshot.blockClaims.height).to.equal(0);
      let zeroSnapshot = await snapshots.getSnapshot(0);
      expect(zeroSnapshot.blockClaims.height).to.equal(0);
      expect(zeroSnapshot.blockClaims.chainId).to.eq(0);
      expect(zeroSnapshot.committedAt).to.eq(0);
      expect(
        await fixture.snapshots.getAliceNetHeightFromSnapshot(0)
      ).to.be.equals(0);
      expect(
        await fixture.snapshots.getBlockClaimsFromSnapshot(0)
      ).to.be.deep.equals([
        0,
        0,
        0,
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
      ]);
      expect(await fixture.snapshots.getChainIdFromSnapshot(0)).to.be.equals(0);
      expect(
        await fixture.snapshots.getCommittedHeightFromSnapshot(0)
      ).to.be.equals(0);

      expect(
        await fixture.snapshots.getAliceNetHeightFromLatestSnapshot()
      ).to.be.equals(0);
      expect(
        await fixture.snapshots.getBlockClaimsFromLatestSnapshot()
      ).to.be.deep.equals([
        0,
        0,
        0,
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
      ]);
      expect(
        await fixture.snapshots.getChainIdFromLatestSnapshot()
      ).to.be.equals(0);
      expect(
        await fixture.snapshots.getCommittedHeightFromLatestSnapshot()
      ).to.be.equals(0);
      const signedSnapshots = signedData1;
      const validators: Array<string> = [];
      for (const validator of validatorsSnapshotsG1) {
        validators.push(validator.address);
      }
      const snapshotsG1 = fixture.snapshots.connect(
        await getValidatorEthAccount(validatorsSnapshotsG1[0])
      );
      // take 12 snapshots
      for (let i = epochs + 1; i <= 12; i++) {
        await mineBlocks(
          (await snapshotsG1.getMinimumIntervalBetweenSnapshots()).toBigInt()
        );
        const contractTx = await snapshotsG1.snapshot(
          signedSnapshots[i - 1].GroupSignature,
          signedSnapshots[i - 1].BClaims,
          { gasLimit: 30000000 }
        );
        await contractTx.wait();
      }

      zeroSnapshot = await snapshots.getSnapshot(0);
      expect(zeroSnapshot.blockClaims.height).to.equal(0);
      expect(zeroSnapshot.blockClaims.chainId).to.eq(0);
      expect(zeroSnapshot.committedAt).to.eq(0);
      expect(
        await fixture.snapshots.getAliceNetHeightFromSnapshot(0)
      ).to.be.equals(0);
      expect(
        await fixture.snapshots.getBlockClaimsFromSnapshot(0)
      ).to.be.deep.equals([
        0,
        0,
        0,
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
      ]);
      expect(await fixture.snapshots.getChainIdFromSnapshot(0)).to.be.equals(0);
      expect(
        await fixture.snapshots.getCommittedHeightFromSnapshot(0)
      ).to.be.equals(0);
    });

    it("test latest getter functions", async () => {
      let epochs = (await fixture.snapshots.getEpoch()).toNumber();
      expect(epochs).to.be.equal(0);
      let lastSnapshot = await fixture.snapshots.getLatestSnapshot();
      expect(lastSnapshot.blockClaims.height).to.equal(0);
      expect(
        await fixture.snapshots.getAliceNetHeightFromLatestSnapshot()
      ).to.be.equals(0);
      expect(
        await fixture.snapshots.getBlockClaimsFromLatestSnapshot()
      ).to.be.deep.equals([
        0,
        0,
        0,
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
      ]);
      expect(
        await fixture.snapshots.getChainIdFromLatestSnapshot()
      ).to.be.equals(0);
      expect(
        await fixture.snapshots.getCommittedHeightFromLatestSnapshot()
      ).to.be.equals(0);
      const signedSnapshots = signedData1;
      const validators: Array<string> = [];
      for (const validator of validatorsSnapshotsG1) {
        validators.push(validator.address);
      }
      const snapshotsG1 = fixture.snapshots.connect(
        await getValidatorEthAccount(validatorsSnapshotsG1[0])
      );
      let lastSnapshotBlockNumber
      // take 12 snapshots
      for (let i = epochs + 1; i <= 12; i++) {
        await mineBlocks(
          (await snapshotsG1.getMinimumIntervalBetweenSnapshots()).toBigInt()
        );
        const contractTx = await snapshotsG1.snapshot(
          signedSnapshots[i - 1].GroupSignature,
          signedSnapshots[i - 1].BClaims,
          { gasLimit: 30000000 }
        );
        const rcpt = await contractTx.wait();
        lastSnapshotBlockNumber = rcpt.blockNumber
      }
      epochs = (await snapshotsG1.getEpoch()).toNumber();
      lastSnapshot = await fixture.snapshots.getLatestSnapshot();
      expect(lastSnapshot.blockClaims.height).to.equal(epochs * epochLength);
      expect(
        await fixture.snapshots.getAliceNetHeightFromLatestSnapshot()
      ).to.be.equals(epochs * epochLength);
      expect(
        await fixture.snapshots.getBlockClaimsFromLatestSnapshot()
      ).to.be.deep.equals([
        1,
        12288,
        0,
        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
      ]);
      expect(
        await fixture.snapshots.getChainIdFromLatestSnapshot()
      ).to.be.equals(1);
      expect(
        await fixture.snapshots.getCommittedHeightFromLatestSnapshot()
      ).to.be.equals(lastSnapshotBlockNumber);
    });

    it("attempts to get a parameters from snapshot that is no longer in the buffer", async () => {
      let epochs = (await fixture.snapshots.getEpoch()).toNumber();
      const signedSnapshots = signedData1;
      const validators: Array<string> = [];
      for (const validator of validatorsSnapshotsG1) {
        validators.push(validator.address);
      }
      const snapshotsG1 = fixture.snapshots.connect(
        await getValidatorEthAccount(validatorsSnapshotsG1[0])
      );
      // take 12 snapshots
      for (let i = epochs + 1; i <= 12; i++) {
        await mineBlocks(
          (await snapshotsG1.getMinimumIntervalBetweenSnapshots()).toBigInt()
        );
        const contractTx = await snapshotsG1.snapshot(
          signedSnapshots[i - 1].GroupSignature,
          signedSnapshots[i - 1].BClaims,
          { gasLimit: 30000000 }
        );
        await contractTx.wait();
      }
      epochs = (await snapshotsG1.getEpoch()).toNumber();
      const lastSnapshot = await snapshotsG1.getLatestSnapshot();
      expect(lastSnapshot.blockClaims.height).to.equal(epochs * epochLength);
      await expect(snapshots.getSnapshot(epochs - 6))
        .to.be.revertedWithCustomError(
          fixture.snapshots,
          `SnapshotsNotInBuffer`
        )
        .withArgs(epochs - 6);
        await expect(snapshots.getAliceNetHeightFromSnapshot(epochs - 6))
        .to.be.revertedWithCustomError(
          fixture.snapshots,
          `SnapshotsNotInBuffer`
        )
        .withArgs(epochs - 6);
        await expect(snapshots.getBlockClaimsFromSnapshot(epochs - 6))
        .to.be.revertedWithCustomError(
          fixture.snapshots,
          `SnapshotsNotInBuffer`
        )
        .withArgs(epochs - 6);
        await expect(snapshots.getCommittedHeightFromSnapshot(epochs - 6))
        .to.be.revertedWithCustomError(
          fixture.snapshots,
          `SnapshotsNotInBuffer`
        )
        .withArgs(epochs - 6);
    });
  });
});
