import {
  Fixture,
  getFixture,
  getTokenIdFromTx,
  getValidatorEthAccount,
} from "../setup";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { Signer } from "ethers";
import { validatorsSnapshots } from "./assets/4-validators-snapshots-1";

describe("Snapshots: Tests Snapshots methods", () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  let notAdmin1Signer: Signer;
  let randomerSigner: Signer;
  let stakeAmount = 20000;
  let stakeAmountMadWei = ethers.utils.parseUnits(stakeAmount.toString(), 18);
  let lockTime = 1;
  let validators = new Array();
  let stakingTokenIds = new Array();

  beforeEach(async function () {
    validators = [];
    stakingTokenIds = [];
    fixture = await getFixture(true, false);
    const [admin, notAdmin1, , , , randomer] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    notAdmin1Signer = await getValidatorEthAccount(notAdmin1.address);
    randomerSigner = await getValidatorEthAccount(randomer.address);

    for (const validator of validatorsSnapshots) {
      validators.push(validator.address);
    }

    await fixture.madToken.approve(
      fixture.validatorPool.address,
      stakeAmountMadWei.mul(validators.length)
    );
    await fixture.madToken.approve(
      fixture.stakeNFT.address,
      stakeAmountMadWei.mul(validators.length)
    );

    for (const validator of validatorsSnapshots) {
      let tx = await fixture.stakeNFT
        .connect(adminSigner)
        .mintTo(validator.address, stakeAmountMadWei, lockTime);
      let tokenId = getTokenIdFromTx(tx);
      stakingTokenIds.push(tokenId);
      await fixture.stakeNFT
        .connect(await getValidatorEthAccount(validator))
        .setApprovalForAll(fixture.validatorPool.address, true);
    }

    await fixture.validatorPool
      .connect(adminSigner)
      .registerValidators(validators, stakingTokenIds);
  });

  it("Does not allow snapshot if sender is not validator", async function () {
    let junkData =
      "0x0000000000000000000000000000000000000000000000000000006d6168616d";
    await expect(
      fixture.snapshots.connect(randomerSigner).snapshot(junkData, junkData)
    ).to.be.revertedWith("Snapshots: Only validators allowed!");
  });

  it("Does not allow snapshot consensus is not running", async function () {
    let junkData =
      "0x0000000000000000000000000000000000000000000000000000006d6168616d";
    let validValidator = await getValidatorEthAccount(validatorsSnapshots[0]);
    await expect(
      fixture.snapshots.connect(validValidator).snapshot(junkData, junkData)
    ).to.be.revertedWith(`Snapshots: Consensus is not running!`);
  });
});
