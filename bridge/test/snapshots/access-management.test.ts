import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { Signer } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getValidatorEthAccount,
} from "../setup";

async function deployFixture() {
  const fixture = await getFixture(true, false);
  const [admin, , , , , randomer] = fixture.namedSigners;
  const adminSigner = await getValidatorEthAccount(admin.address);
  const randomerSigner = await getValidatorEthAccount(randomer.address);
  return { fixture, admin, randomer, adminSigner, randomerSigner };
}

describe("Snapshots: Access control methods", () => {
  let fixture: Fixture;

  let randomer: SignerWithAddress;
  let adminSigner: Signer;
  let randomerSigner: Signer;

  beforeEach(async function () {
    ({ fixture, randomer, adminSigner, randomerSigner } = await loadFixture(
      deployFixture
    ));
  });

  it("Should not allow initialize more than once", async () => {
    await expect(
      fixture.factory.callAny(
        fixture.snapshots.address,
        0,
        fixture.snapshots.interface.encodeFunctionData("initialize", [1, 2])
      )
    ).to.revertedWith("Initializable: contract is already initialized");
  });

  it("Only factory should be allowed to call initialize", async () => {
    const snapshots = await (
      await ethers.getContractFactory("Snapshots")
    ).deploy(1, 2);
    const [, user] = await ethers.getSigners();
    await expect(
      snapshots.connect(user).initialize(1, 2)
    ).to.revertedWithCustomError(snapshots, "OnlyFactory");
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
