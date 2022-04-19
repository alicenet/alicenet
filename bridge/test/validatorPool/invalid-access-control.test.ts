import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "../setup";

describe("ValidatorPool Access Control: An user without admin role should not be able to:", async function () {
  let fixture: Fixture;
  let notAdmin1Signer: SignerWithAddress;
  const maxNumValidators = 5;
  const validators: any[] = [];
  const stakingTokenIds: any[] = [];

  beforeEach(async function () {
    fixture = await getFixture();
    const [, notAdmin1, , ,] = fixture.namedSigners;
    notAdmin1Signer = await ethers.getSigner(notAdmin1.address);
  });

  it("Set a maximum number of validators", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .setMaxNumValidators(maxNumValidators)
    ).to.be.revertedWith("2000");
  });

  it("Set disputer reward", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).setDisputerReward(1)
    ).to.be.revertedWith("2000");
  });

  it("Set stake Amount", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).setStakeAmount(1)
    ).to.be.revertedWith("823");
  });

  it("Initialize ETHDKG", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).initializeETHDKG()
    ).to.be.revertedWith("2000");
  });

  it("Complete ETHDKG", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).completeETHDKG()
    ).to.be.revertedWith("2016");
  });

  it("Schedule maintenance", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).scheduleMaintenance()
    ).to.be.revertedWith("2000");
  });

  it("Pause consensus on arbitrary height", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .pauseConsensusOnArbitraryHeight(1)
    ).to.be.revertedWith("2000");
  });

  it("Pause consensus", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).pauseConsensus()
    ).to.be.revertedWith("2008");
  });

  it("Register validators", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .registerValidators(validators, stakingTokenIds)
    ).to.be.revertedWith("2000");
  });

  it("Unregister validators", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .unregisterValidators(validators)
    ).to.be.revertedWith("2000");
  });

  it("Unregister all validators", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).unregisterAllValidators()
    ).to.be.revertedWith("2000");
  });

  it("Set location", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).setLocation("0.0.0.1")
    ).to.be.revertedWith("800");
  });

  it("Collect profit", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).collectProfits()
    ).to.be.revertedWith("800");
  });

  it("Major slash a validator", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .majorSlash(notAdmin1Signer.address, notAdmin1Signer.address)
    ).to.be.revertedWith("2016");
  });

  it("Minor slash a validator", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .minorSlash(notAdmin1Signer.address, notAdmin1Signer.address)
    ).to.be.revertedWith("2016");
  });
});
