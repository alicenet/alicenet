import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "hardhat";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getReceiptForFailedTransaction,
} from "../setup";
import { validatorsSnapshots } from "../snapshots/assets/4-validators-snapshots-1";
import { commitSnapshots, createValidators, stakeValidators } from "./setup";

function deployFixture() {
  return getFixture(false, true);
}

describe("ValidatorPool Access Control: An user with admin role should be able to:", async function () {
  let fixture: Fixture;
  const maxNumValidators = 5;
  const stakeAmount = 20000;

  beforeEach(async function () {
    fixture = await loadFixture(deployFixture);
  });

  it("Set a minimum stake", async function () {
    const rcpt = await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "setStakeAmount",
      [stakeAmount]
    );
    expect(rcpt.status).to.equal(1);
  });

  it("Set a maximum number of validators", async function () {
    const rcpt = await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "setMaxNumValidators",
      [maxNumValidators]
    );
    expect(rcpt.status).to.equal(1);
  });

  it("Set Max Interval Without Snapshots", async function () {
    const rcpt = await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "setMaxIntervalWithoutSnapshots",
      [1]
    );
    expect(rcpt.status).to.equal(1);
    expect(
      await fixture.validatorPool.getMaxIntervalWithoutSnapshots()
    ).to.be.equals(1);
  });

  it("Schedule maintenance", async function () {
    const rcpt = await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "scheduleMaintenance"
    );
    expect(rcpt.status).to.equal(1);
  });

  it("Register validators", async function () {
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        ["0x000000000000000000000000000000000000dEaD"],
        [1],
      ])
    ).to.be.revertedWith("ERC721: invalid token ID");
  });

  it("Initialize ETHDKG", async function () {
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG")
    )
      .to.be.revertedWithCustomError(fixture.ethdkg, `MinimumValidatorsNotMet`)
      .withArgs(0);
  });

  it("Unregister validators", async function () {
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "unregisterValidators", [
        ["0x000000000000000000000000000000000000dEaD"],
      ])
    )
      .to.be.revertedWithCustomError(
        fixture.validatorPool,
        "LengthGreaterThanAvailableValidators"
      )
      .withArgs(1, 0);
  });

  it("Pause consensus", async function () {
    await commitSnapshots(fixture, 1);
    const latestSnapshotHeight =
      await fixture.snapshots.getCommittedHeightFromLatestSnapshot();
    const maxInterval =
      await fixture.validatorPool.getMaxIntervalWithoutSnapshots();

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
});

describe("ValidatorPool Access Control: An user with admin role should not be able to:", async function () {
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await loadFixture(deployFixture);
  });

  it("Set Max Interval Without Snapshots to 0", async function () {
    await expect(
      fixture.factory.callAny(
        fixture.validatorPool.address,
        0,
        fixture.validatorPool.interface.encodeFunctionData(
          "setMaxIntervalWithoutSnapshots",
          [0]
        )
      )
    ).to.be.revertedWithCustomError(
      fixture.validatorPool,
      "MaxIntervalWithoutSnapshotsMustBeNonZero"
    );
  });

  it("Set Max Num Validators to a number smaller then validatorPool length", async function () {
    const validators = await createValidators(fixture, validatorsSnapshots);
    const stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await expect(
      fixture.factory.callAny(
        fixture.validatorPool.address,
        0,
        fixture.validatorPool.interface.encodeFunctionData(
          "setMaxNumValidators",
          [3]
        )
      )
    )
      .to.be.revertedWithCustomError(
        fixture.validatorPool,
        "MaxNumValidatorsIsTooLow"
      )
      .withArgs(3, 4);
  });
});
