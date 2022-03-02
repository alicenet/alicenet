import { Fixture, getFixture, factoryCallAny } from "../setup";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";

describe("ValidatorPool Access Control: An user with admin role should be able to:", async function () {
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
    let rcpt = await factoryCallAny(
      fixture,
      "validatorPool",
      "setStakeAmount",
      [stakeAmount]
    );
    expect(rcpt.status).to.equal(1);
  });

  it("Set a maximum number of validators", async function () {
    let rcpt = await factoryCallAny(
      fixture,
      "validatorPool",
      "setMaxNumValidators",
      [maxNumValidators]
    );
    expect(rcpt.status).to.equal(1);
  });

  it("Schedule maintenance", async function () {
    let rcpt = await factoryCallAny(
      fixture,
      "validatorPool",
      "scheduleMaintenance"
    );
    expect(rcpt.status).to.equal(1);
  });

  it("Register validators", async function () {
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        ["0x000000000000000000000000000000000000dEaD"],
        [1],
      ])
    ).to.be.revertedWith("ERC721: owner query for nonexistent token");
  });

  it("Initialize ETHDKG", async function () {
    await expect(
      factoryCallAny(fixture, "validatorPool", "initializeETHDKG")
    ).to.be.revertedWith(
      "ETHDKG: Minimum number of validators staked not met!"
    );
  });

  it("Unregister validators", async function () {
    await expect(
      factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
        ["0x000000000000000000000000000000000000dEaD"],
      ])
    ).to.be.revertedWith(
      "ValidatorPool: There are not enough validators to be removed!"
    );
  });

  it("Pause consensus", async function () {
    await expect(
      factoryCallAny(
        fixture,
        "validatorPool",
        "pauseConsensusOnArbitraryHeight",
        [1]
      )
    ).to.be.revertedWith("ValidatorPool: Condition not met to stop consensus!");
  });
});
