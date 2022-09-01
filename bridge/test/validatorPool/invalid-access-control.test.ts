import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { ValidatorPool } from "../../typechain-types";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "../setup";

describe("Initialization", async function () {
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture();
  });

  it("Should not allow initialize more than once", async () => {
    await expect(
      fixture.factory.callAny(
        fixture.validatorPool.address,
        0,
        (fixture.validatorPool as ValidatorPool).interface.encodeFunctionData(
          "initialize",
          [1, 2, 3, 4]
        )
      )
    ).to.revertedWith("Initializable: contract is already initialized");
  });

  it("Only factory should be allowed to call initialize", async () => {
    const validatorPool = await (
      await ethers.getContractFactory("ValidatorPool")
    ).deploy();
    const [, user] = await ethers.getSigners();
    await expect(
      validatorPool.connect(user).initialize(1, 2, 3, 4)
    ).to.revertedWithCustomError(validatorPool, "OnlyFactory");
  });
});

describe("ValidatorPool Access Control: An user without admin role should not be able to:", async function () {
  let fixture: Fixture;
  let notAdmin1: SignerWithAddress;
  let notAdmin1Signer: SignerWithAddress;
  const maxNumValidators = 5;
  const stakeAmount = 20000;
  const validators: any[] = [];
  const stakingTokenIds: any[] = [];

  beforeEach(async function () {
    fixture = await getFixture();
    [, notAdmin1, , ,] = fixture.namedSigners;
    notAdmin1Signer = await ethers.getSigner(notAdmin1.address);
  });

  it("Set a minimum stake", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).setStakeAmount(stakeAmount)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Set a maximum number of validators", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .setMaxNumValidators(maxNumValidators)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Set disputer reward", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).setDisputerReward(1)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Set Max Interval Without Snapshots", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .setMaxIntervalWithoutSnapshots(1)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Set stake Amount", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).setStakeAmount(1)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Initialize ETHDKG", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).initializeETHDKG()
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Complete ETHDKG", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).completeETHDKG()
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyETHDKG`)
      .withArgs(notAdmin1.address, fixture.ethdkg.address);
  });

  it("Schedule maintenance", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).scheduleMaintenance()
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Pause consensus on arbitrary height", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .pauseConsensusOnArbitraryHeight(1)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Pause consensus", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).pauseConsensus()
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlySnapshots`)
      .withArgs(notAdmin1.address, fixture.snapshots.address);
  });

  it("Register validators", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .registerValidators(validators, stakingTokenIds)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Unregister validators", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .unregisterValidators(validators)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Unregister all validators", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).unregisterAllValidators()
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyFactory`)
      .withArgs(notAdmin1.address, fixture.factory.address);
  });

  it("Set location", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).setLocation("0.0.0.1")
    )
      .to.be.revertedWithCustomError(
        fixture.validatorPool,
        "CallerNotValidator"
      )
      .withArgs(notAdmin1.address);
  });

  it("Collect profit", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).collectProfits()
    )
      .to.be.revertedWithCustomError(
        fixture.validatorPool,
        "CallerNotValidator"
      )
      .withArgs(notAdmin1.address);
  });

  it("Major slash a validator", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .majorSlash(notAdmin1Signer.address, notAdmin1Signer.address)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyETHDKG`)
      .withArgs(notAdmin1.address, fixture.ethdkg.address);
  });

  it("Minor slash a validator", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .minorSlash(notAdmin1Signer.address, notAdmin1Signer.address)
    )
      .to.be.revertedWithCustomError(fixture.validatorPool, `OnlyETHDKG`)
      .withArgs(notAdmin1.address, fixture.ethdkg.address);
  });
});
