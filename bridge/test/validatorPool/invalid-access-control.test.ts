import { Fixture, getFixture } from "../setup";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";

describe("ValidatorPool Access Control: An user without admin role should not be able to:", async function () {
  let fixture: Fixture;
  let adminSigner: SignerWithAddress;
  let notAdmin1Signer: SignerWithAddress;
  let maxNumValidators = 5;
  let stakeAmount = 20000;
  let validators = new Array();
  let stakingTokenIds = new Array();

  beforeEach(async function () {
    validators = [];
    stakingTokenIds = [];
    fixture = await getFixture();
    const [admin, notAdmin1, , ,] = fixture.namedSigners;
    adminSigner = await ethers.getSigner(admin.address);
    notAdmin1Signer = await ethers.getSigner(notAdmin1.address);
  });

  it("Set a minimum stake", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).setStakeAmount(stakeAmount)
    ).to.be.revertedWith("onlyFactory");
  });

  it("Set a maximum number of validators", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .setMaxNumValidators(maxNumValidators)
    ).to.be.revertedWith("onlyFactory");
  });

  it("Schedule maintenance", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).scheduleMaintenance()
    ).to.be.revertedWith("onlyFactory");
  });

  it("Register validators", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .registerValidators(validators, stakingTokenIds)
    ).to.be.revertedWith("onlyFactory");
  });

  it("Initialize ETHDKG", async function () {
    await expect(
      fixture.validatorPool.connect(notAdmin1Signer).initializeETHDKG()
    ).to.be.revertedWith("onlyFactory");
  });

  it("Unregister validators", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .unregisterValidators(validators)
    ).to.be.revertedWith("onlyFactory");
  });

  it("Pause consensus", async function () {
    await expect(
      fixture.validatorPool
        .connect(notAdmin1Signer)
        .pauseConsensusOnArbitraryHeight(1)
    ).to.be.revertedWith("onlyFactory");
  });
});
