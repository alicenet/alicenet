import { BigNumber, Signer } from "ethers";
import { ethers, network } from "hardhat";
import { expect } from "../../chai-setup";
import {
  assertEventValidatorSetCompleted,
  completeETHDKGRound,
} from "../../ethdkg/setup";
import {
  factoryCallAny,
  Fixture,
  getFixture,
  getValidatorEthAccount,
} from "../../setup";
import {
  validatorsSnapshots,
  validSnapshot1024,
} from "../../snapshots/assets/4-validators-snapshots-1";
import { validatorsSnapshots as validatorsSnapshots2 } from "../../snapshots/assets/4-validators-snapshots-2";
import { createValidators, stakeValidators } from "../setup";

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
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.true;
  });

  it("Should not allow pausing “consensus” before 1.5 without snapshots", async function () {
    await expect(
      factoryCallAny(
        fixture,
        "validatorPool",
        "pauseConsensusOnArbitraryHeight",
        [1]
      )
    ).to.be.revertedWith("ValidatorPool: Condition not met to stop consensus!");
  });

  it("Pause consensus after 1.5 days without snapshot", async function () {
    // set consensus running
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.true;
    // simulate a period without snapshots
    await network.provider.send("hardhat_mine", [
      ethers.utils.hexValue(
        (
          await fixture.validatorPool.MAX_INTERVAL_WITHOUT_SNAPSHOTS()
        ).toBigInt() + BigInt(1)
      ),
    ]);
    let tx = await factoryCallAny(
      fixture,
      "validatorPool",
      "pauseConsensusOnArbitraryHeight",
      [1]
    );
    let receipt = await ethers.provider.getTransaction(tx.transactionHash);
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
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
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.true;
    // simulate a period without snapshots
    await network.provider.send("hardhat_mine", [
      ethers.utils.hexValue(
        (
          await fixture.validatorPool.MAX_INTERVAL_WITHOUT_SNAPSHOTS()
        ).toBigInt() + BigInt(1)
      ),
    ]);
    let arbitraryMadnetHeight = 42;
    let tx = await factoryCallAny(
      fixture,
      "validatorPool",
      "pauseConsensusOnArbitraryHeight",
      [arbitraryMadnetHeight]
    );
    let transaction = await ethers.provider.getTransaction(tx.transactionHash);
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
    // special event to halt consensus in the side chain
    await assertEventValidatorSetCompleted(
      transaction,
      0,
      1,
      0,
      0,
      arbitraryMadnetHeight,
      [0, 0, 0, 0]
    );
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);

    let newValidators = await createValidators(fixture, validatorsSnapshots2);
    let newStakingTokenIds = await stakeValidators(fixture, validators);

    // set consensus running
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      newValidators,
      newStakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
    await completeETHDKGRound(
      validatorsSnapshots2,
      {
        ethdkg: fixture.ethdkg,
        validatorPool: fixture.validatorPool,
      },
      undefined,
      arbitraryMadnetHeight
    );
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.true;
  });

  it("Complete ETHDKG and check if the necessary state was set properly", async function () {
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    // Complete ETHDKG Round
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
    expect(await fixture.validatorPool.isMaintenanceScheduled()).to.be.false;
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.true;
  });

  it("Should not allow start ETHDKG if consensus is true or and ETHDKG round is running", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    // Complete ETHDKG Round
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    await expect(
      factoryCallAny(fixture, "validatorPool", "initializeETHDKG")
    ).to.be.revertedWith("ValidatorPool: There's an ETHDKG round running!");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await expect(
      factoryCallAny(fixture, "validatorPool", "initializeETHDKG")
    ).to.be.revertedWith(
      "ValidatorPool: Error Madnet Consensus should be halted!"
    );
  });

  it("Register validators, run ethdkg, schedule maintenance, do a snapshot, replace some validators, and rerun ethdkg", async function () {
    let fixture = await getFixture();
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    // Complete ETHDKG Round
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.true;
    expect(await fixture.validatorPool.isMaintenanceScheduled()).to.be.false;
    await factoryCallAny(fixture, "validatorPool", "scheduleMaintenance");
    expect(await fixture.validatorPool.isMaintenanceScheduled()).to.be.true;
    await fixture.snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims);
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.equal(
      false,
      "Consensus should be halted!"
    );

    await factoryCallAny(fixture, "validatorPool", "unregisterAllValidators");
    let newValidators = await createValidators(fixture, validatorsSnapshots2);
    let newStakingTokenIds = await stakeValidators(fixture, newValidators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      newValidators,
      newStakingTokenIds,
    ]);
    // Complete ETHDKG Round
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
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
    ).to.be.revertedWith("Only NFT contracts allowed to send ethereum!");
  });
});
