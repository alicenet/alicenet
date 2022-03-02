import {
  Fixture,
  getValidatorEthAccount,
  getFixture,
  factoryCallAny,
} from "../../setup";
import { completeETHDKGRound } from "../../ethdkg/setup";
import { expect } from "../../chai-setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import { validatorsSnapshots as validatorsSnapshots2 } from "../../snapshots/assets/4-validators-snapshots-2";
import { BigNumber, Signer } from "ethers";
import { SnapshotsMock } from "../../../typechain-types";
import {
  claimPosition,
  commitSnapshots,
  createValidators,
  getCurrentState,
  showState,
  stakeValidators,
} from "../setup";
import { ethers } from "hardhat";
import { exit } from "process";

describe("ValidatorPool: Consensus dependent logic ", async () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  let validators: string[];
  let stakingTokenIds: BigNumber[];

  beforeEach(async function () {
    fixture = await getFixture(false, true, true);
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
  });

  it("Should not allow pausing “consensus” before 1.5 without snapshots.", async function () {
    await expect(
      factoryCallAny(
        fixture,
        "validatorPool",
        "pauseConsensusOnArbitraryHeight",
        [1]
      )
    ).to.be.revertedWith("ValidatorPool: Condition not met to stop consensus!");
  });

  it("Pause consensus after 1.5 days without snapshot.", async function () {
    // Simulating in mock with _maxIntervalWithoutSnapshot = 0 instead of 8192
    fixture = await getFixture(true, true, true);
    await factoryCallAny(
      fixture,
      "validatorPool",
      "pauseConsensusOnArbitraryHeight",
      [1]
    );
  });

  it("After 1.5 days without snapshots, replace validators and run ETHDKG to completion.", async function () {
    // Simulating in mock with _maxIntervalWithoutSnapshot = 0 instead of 8192
    fixture = await getFixture(true, true, true);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    // Complete ETHDKG Round
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
  });

  it("Complete ETHDKG and check if the necessary state was set properly", async function () {
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
    expect(await fixture.validatorPool.isMaintenanceScheduled()).to.be.false;
    //Consensus Running should be true
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith(
      "ValidatorPool: Error Madnet Consensus should be halted!"
    );
  });

  it("Should not allow start ETHDKG if consensus is true or and ETHDKG round is running.", async function () {
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
    await expect(
      factoryCallAny(fixture, "validatorPool", "initializeETHDKG")
    ).to.be.revertedWith(
      "ValidatorPool: Error Madnet Consensus should be halted!"
    );
  });

  it("Test running ETHDKG with the arbitrary height sent as input, see if the value (the arbitrary height) is correctly added to the ethdkg completion event.", async function () {
    const height = 4;
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await expect(fixture.ethdkg.setCustomMadnetHeight(height))
      .to.emit(fixture.ethdkg, "ValidatorSetCompleted")
      .withArgs(0, 0, 0, 0, height, 0, 0, 0, 0);
  });

  it("Check if consensus is halted after maintenance is scheduled and snapshot is done.", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await showState(
      "after registering",
      await getCurrentState(fixture, validators)
    );
    let expectedState = await getCurrentState(fixture, validators);
    // Complete ETHDKG Round
    await factoryCallAny(fixture, "ethdkg", "initializeETHDKG");
    await showState(
      "after initializing",
      await getCurrentState(fixture, validators)
    );
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await factoryCallAny(fixture, "validatorPool", "scheduleMaintenance");
    await commitSnapshots(fixture, 1);
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
  });

  it("Register validators, run ethdkg, schedule maintenance, do a snapshot, replace some validators, and rerun ethdkg", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await showState(
      "after registering",
      await getCurrentState(fixture, validators)
    );
    let expectedState = await getCurrentState(fixture, validators);
    // // Complete ETHDKG Round
    await factoryCallAny(fixture, "ethdkg", "initializeETHDKG");
    await showState(
      "After initializing",
      await getCurrentState(fixture, validators)
    );
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await factoryCallAny(fixture, "validatorPool", "scheduleMaintenance");
    await commitSnapshots(fixture, 1);
    expect(await fixture.validatorPool.isConsensusRunning()).to.be.false;
    await showState(
      "After maintenance",
      await getCurrentState(fixture, validators)
    );
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await showState(
      "After unregistering",
      await getCurrentState(fixture, validators)
    );
    await commitSnapshots(fixture, 4);
    for (const validator of validatorsSnapshots) {
      await claimPosition(fixture, validator);
    }
    // Re mint positions
    let newValidators = await createValidators(fixture, validatorsSnapshots2);
    let newStakeNFTIDs = await stakeValidators(fixture, newValidators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      newValidators,
      newStakeNFTIDs,
    ]);
    await showState(
      "After registering",
      await getCurrentState(fixture, validators)
    );
    let currentState = await getCurrentState(fixture, validators);
    await factoryCallAny(fixture, "ethdkg", "initializeETHDKG");
    await showState(
      "After initializing",
      await getCurrentState(fixture, validators)
    );
    await completeETHDKGRound(
      validatorsSnapshots2,
      {
        ethdkg: fixture.ethdkg,
        validatorPool: fixture.validatorPool,
      },
      5
    );
  });
});
