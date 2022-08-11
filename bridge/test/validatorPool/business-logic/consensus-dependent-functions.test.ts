import { BigNumber, Signer } from "ethers";
import { ethers, network } from "hardhat";
import { expect } from "../../chai-setup";
import {
  assertEventValidatorSetCompleted,
  completeETHDKGRound,
} from "../../ethdkg/setup";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getReceiptForFailedTransaction,
  getValidatorEthAccount,
} from "../../setup";
import {
  validatorsSnapshots,
  validSnapshot1024,
} from "../../snapshots/assets/4-validators-snapshots-1";
import { validatorsSnapshots as validatorsSnapshots2 } from "../../snapshots/assets/4-validators-snapshots-2";
import { commitSnapshots, createValidators, stakeValidators } from "../setup";

describe("ValidatorPool: Consensus dependent logic ", async () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  let validators: string[];
  let stakingTokenIds: BigNumber[];

  beforeEach(async function () {
    fixture = await getFixture(false, true, false);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
  });

  it("Initialize ETHDKG", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(
      false
    );
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(true);
  });

  it("Should not allow pausing “consensus” before 1.5 without snapshots", async function () {
    await commitSnapshots(fixture, 1);
    const latestSnapshotHeight =
      await fixture.snapshots.getCommittedHeightFromLatestSnapshot();
    const maxInterval =
      await fixture.validatorPool.MAX_INTERVAL_WITHOUT_SNAPSHOTS();
    const txPromise = factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "pauseConsensusOnArbitraryHeight",
      [1]
    );
    const expectedBlockNumber = (
      await getReceiptForFailedTransaction(txPromise)
    ).blockNumber;
    await expect(txPromise)
      .to.be.revertedWithCustomError(
        fixture.validatorPool,
        "MinimumBlockIntervalNotMet"
      )
      .withArgs(expectedBlockNumber, latestSnapshotHeight.add(maxInterval));
  });

  it("Pause consensus after 1.5 days without snapshot", async function () {
    // set consensus running
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(
      false
    );
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(true);
    // simulate a period without snapshots
    await network.provider.send("hardhat_mine", [
      ethers.utils.hexValue(
        (
          await fixture.validatorPool.MAX_INTERVAL_WITHOUT_SNAPSHOTS()
        ).toBigInt() + BigInt(1)
      ),
    ]);
    const tx = await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "pauseConsensusOnArbitraryHeight",
      [1]
    );
    const receipt = await ethers.provider.getTransaction(tx.transactionHash);
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(
      false
    );
    // special event to halt consensus in the side chain
    await assertEventValidatorSetCompleted(
      receipt,
      0,
      1,
      0,
      0,
      1,
      [0, 0, 0, 0]
    );
  });

  it("After 1.5 days without snapshots, replace validators and run ETHDKG to completion", async function () {
    // set consensus running
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(
      false
    );
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(true);
    // simulate a period without snapshots
    await network.provider.send("hardhat_mine", [
      ethers.utils.hexValue(
        (
          await fixture.validatorPool.MAX_INTERVAL_WITHOUT_SNAPSHOTS()
        ).toBigInt() + BigInt(1)
      ),
    ]);
    const arbitraryAliceNetHeight = 42;
    const tx = await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "pauseConsensusOnArbitraryHeight",
      [arbitraryAliceNetHeight]
    );
    const transaction = await ethers.provider.getTransaction(
      tx.transactionHash
    );
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(
      false
    );
    // special event to halt consensus in the side chain
    await assertEventValidatorSetCompleted(
      transaction,
      0,
      1,
      0,
      0,
      arbitraryAliceNetHeight,
      [0, 0, 0, 0]
    );
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterValidators",
      [validators]
    );

    const newValidators = await createValidators(fixture, validatorsSnapshots2);
    const newStakingTokenIds = await stakeValidators(fixture, validators);

    // set consensus running
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [newValidators, newStakingTokenIds]
    );
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(
      false
    );
    await completeETHDKGRound(
      validatorsSnapshots2,
      {
        ethdkg: fixture.ethdkg,
        validatorPool: fixture.validatorPool,
      },
      undefined,
      arbitraryAliceNetHeight
    );
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(true);
  });

  it("Complete ETHDKG and check if the necessary state was set properly", async function () {
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(
      false
    );
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    // Complete ETHDKG Round
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(
      false
    );
    expect(await fixture.validatorPool.isMaintenanceScheduled()).to.be.equals(
      false
    );
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(true);
  });

  it("Should not allow start ETHDKG if consensus is true or and ETHDKG round is running", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    // Complete ETHDKG Round
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG")
    ).to.be.revertedWithCustomError(
      fixture.validatorPool,
      "ETHDKGRoundRunning"
    );
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG")
    ).to.be.revertedWithCustomError(fixture.validatorPool, "ConsensusRunning");
  });

  it("Register validators, run ethdkg, schedule maintenance, do a snapshot, replace some validators, and rerun ethdkg", async function () {
    const fixture = await getFixture();
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    // Complete ETHDKG Round
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equals(true);
    expect(await fixture.validatorPool.isMaintenanceScheduled()).to.be.equals(
      false
    );
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "scheduleMaintenance"
    );
    expect(await fixture.validatorPool.isMaintenanceScheduled()).to.be.equals(
      true
    );
    await fixture.snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims);
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equal(
      false,
      "Consensus should be halted!"
    );

    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterAllValidators"
    );
    const newValidators = await createValidators(fixture, validatorsSnapshots2);
    const newStakingTokenIds = await stakeValidators(fixture, newValidators);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [newValidators, newStakingTokenIds]
    );
    // Complete ETHDKG Round
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(
      validatorsSnapshots2,
      {
        ethdkg: fixture.ethdkg,
        validatorPool: fixture.validatorPool,
      },
      1,
      1024,
      (
        await fixture.snapshots.getCommittedHeightFromLatestSnapshot()
      ).toNumber()
    );
    expect(await fixture.validatorPool.isMaintenanceScheduled()).to.be.equal(
      false,
      "No maintenance should be scheduled!"
    );
  });

  it("Normal users should not be able to send ethereum to the contract", async function () {
    await expect(
      adminSigner.sendTransaction({
        to: fixture.validatorPool.address,
        value: 1000,
      })
    ).to.be.revertedWithCustomError(
      fixture.validatorPool,
      "OnlyStakingContractsAllowed"
    );
  });
});
