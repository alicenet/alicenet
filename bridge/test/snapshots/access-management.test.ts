import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { Signer } from "ethers";
import { expect } from "../chai-setup";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getValidatorEthAccount,
} from "../setup";

describe("Snapshots: Access control methods", () => {
  let fixture: Fixture;
  let admin: SignerWithAddress;
  let randomer: SignerWithAddress;
  let adminSigner: Signer;
  let randomerSigner: Signer;

  beforeEach(async function () {
    fixture = await getFixture(true, false);
    [admin, , , , , randomer] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    randomerSigner = await getValidatorEthAccount(randomer.address);
  });

  it("GetEpochLength returns 1024", async function () {
    const expectedEpochLength = 1024;

    const epochLength = await fixture.snapshots.getEpochLength();
    await expect(epochLength).to.be.equal(expectedEpochLength);
  });

  it("Does not allow setSnapshotDesperationDelay if sender is not admin", async function () {
    const expectedDelay = 123;
    await expect(
      fixture.snapshots
        .connect(randomerSigner)
        .setSnapshotDesperationDelay(expectedDelay)
    )
      .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
      .withArgs(randomer.address, fixture.factory.address);
  });

  it("Allows setSnapshotDesperationDelay from admin address", async function () {
    const expectedDelay = 123;
    await factoryCallAnyFixture(
      fixture,
      "snapshots",
      "setSnapshotDesperationDelay",
      [expectedDelay]
    );

    const delay = await fixture.snapshots.getSnapshotDesperationDelay();
    await expect(delay).to.be.equal(expectedDelay);
  });

  it("Does not allow setSnapshotDesperationFactor if sender is not admin", async function () {
    const expectedFactor = 123;
    await expect(
      fixture.snapshots
        .connect(randomerSigner)
        .setSnapshotDesperationFactor(expectedFactor)
    )
      .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
      .withArgs(randomer.address, fixture.factory.address);
  });

  it("Allows setSnapshotDesperationFactor from admin address", async function () {
    const expectedFactor = 123;

    await factoryCallAnyFixture(
      fixture,
      "snapshots",
      "setSnapshotDesperationFactor",
      [expectedFactor]
    );

    const delay = await fixture.snapshots
      .connect(adminSigner)
      .getSnapshotDesperationFactor();
    await expect(delay).to.be.equal(expectedFactor);
  });
});
