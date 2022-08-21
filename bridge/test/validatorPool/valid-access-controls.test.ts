import { expect } from "hardhat";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getReceiptForFailedTransaction,
} from "../setup";
import { commitSnapshots } from "./setup";

describe("ValidatorPool Access Control: An user with admin role should be able to:", async function () {
  let fixture: Fixture;
  const maxNumValidators = 5;
  const stakeAmount = 20000;

  beforeEach(async function () {
    fixture = await getFixture(false, true);
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
