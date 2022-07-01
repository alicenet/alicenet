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
import {
  signedData1,
  validatorsSnapshotsG1,
} from "../sharedConstants/4-validators-snapshots-100-Group1";
import {
  signedData2,
  validatorsSnapshotsG2,
} from "../sharedConstants/4-validators-snapshots-100-Group2";
import { createValidators, stakeValidators } from "../validatorPool/setup";

contract("SnapshotRingBuffer 0state", async () => {
  const epochLength = 1024;
  let fixture: Fixture;
  let snapshots: Snapshots;
  describe("Snapshot upgrade integration", async () => {
    beforeEach(async () => {
      //deploys the new snapshot contract with buffer and zero state
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
      //schedule maintenace
      await factoryCallAnyFixture(
        fixture,
        "validatorPool",
        "scheduleMaintenance"
      );

      //unregister validators
      await factoryCallAnyFixture(
        fixture,
        "validatorPool",
        "unregisterValidators",
        [validators]
      );
      //register the new validators
      const newValidators = await createValidators(
        fixture,
        validatorsSnapshotsG2
      );
      const newStakingTokenIds = await stakeValidators(fixture, newValidators);
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
        // console.log(`epoch: ${i}, gas: ${receipt.gasUsed}`);
      }
      epochs = (await snapshotsG1.getEpoch()).toNumber();
      const lastSnapshot = await snapshotsG1.getLatestSnapshot();
      expect(lastSnapshot.blockClaims.height).to.equal(epochs * epochLength);
      const endBuffShot = await snapshots.getSnapshot(epochs - 5);
      expect(endBuffShot.blockClaims.chainId).to.eq(1);
      expect(endBuffShot.blockClaims.height).to.eq((epochs - 5) * epochLength);
    });

    it("attempts to get a snapshot that is no longer in the buffer", async () => {
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
        // console.log(`epoch: ${i}, gas: ${receipt.gasUsed}`);
      }
      epochs = (await snapshotsG1.getEpoch()).toNumber();
      const lastSnapshot = await snapshotsG1.getLatestSnapshot();
      expect(lastSnapshot.blockClaims.height).to.equal(epochs * epochLength);
      expect(snapshots.getSnapshot(epochs - 6)).to.be.revertedWith("410");
    });
  });
});
