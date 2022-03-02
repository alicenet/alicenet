import {
  Fixture,
  getFixture,
  getValidatorEthAccount,
  factoryCallAny,
} from "../setup";

import { expect } from "../chai-setup";

import { Signer } from "ethers";

describe("Snapshots: Access control methods", () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  let randomerSigner: Signer;

  beforeEach(async function () {
    fixture = await getFixture(true, false);
    const [admin, , , , , randomer] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    randomerSigner = await getValidatorEthAccount(randomer.address);
  });

  it("GetEpochLength returns 1024", async function () {
    let expectedEpochLength = 1024;

    const epochLength = await fixture.snapshots.getEpochLength();
    await expect(epochLength).to.be.equal(expectedEpochLength);
  });

  it("Does not allow setSnapshotDesperationDelay if sender is not admin", async function () {
    let expectedDelay = 123;
    await expect(
      fixture.snapshots
        .connect(randomerSigner)
        .setSnapshotDesperationDelay(expectedDelay)
    ).to.be.revertedWith("onlyFactory");
  });

  it("Allows setSnapshotDesperationDelay from admin address", async function () {
    let expectedDelay = 123;
    await factoryCallAny(fixture, "snapshots", "setSnapshotDesperationDelay", [
      expectedDelay,
    ]);

    const delay = await fixture.snapshots.getSnapshotDesperationDelay();
    await expect(delay).to.be.equal(expectedDelay);
  });

  it("Does not allow setSnapshotDesperationFactor if sender is not admin", async function () {
    let expectedFactor = 123;
    await expect(
      fixture.snapshots
        .connect(randomerSigner)
        .setSnapshotDesperationFactor(expectedFactor)
    ).to.be.revertedWith("onlyFactory");
  });

  it("Allows setSnapshotDesperationFactor from admin address", async function () {
    let expectedFactor = 123;

    await factoryCallAny(fixture, "snapshots", "setSnapshotDesperationFactor", [
      expectedFactor,
    ]);

    const delay = await fixture.snapshots
      .connect(adminSigner)
      .getSnapshotDesperationFactor();
    await expect(delay).to.be.equal(expectedFactor);
  });
});
