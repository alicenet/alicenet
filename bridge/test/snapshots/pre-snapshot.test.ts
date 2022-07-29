import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { Signer } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  Fixture,
  getFixture,
  getTokenIdFromTx,
  getValidatorEthAccount,
} from "../setup";
import { validatorsSnapshots } from "./assets/4-validators-snapshots-1";

describe("Snapshots: Tests Snapshots methods", () => {
  let fixture: Fixture;
  let admin: SignerWithAddress;
  let randomUser: SignerWithAddress;
  let adminSigner: Signer;
  let randomSigner: Signer;
  const stakeAmount = 20000;
  const stakeAmountATokenWei = ethers.utils.parseUnits(
    stakeAmount.toString(),
    18
  );
  const lockTime = 1;
  let validators: any[];
  let stakingTokenIds: any[];

  beforeEach(async function () {
    validators = [];
    stakingTokenIds = [];
    fixture = await getFixture(true, false);
    [admin, , , , , randomUser] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    randomSigner = await getValidatorEthAccount(randomUser.address);

    for (const validator of validatorsSnapshots) {
      validators.push(validator.address);
    }

    await fixture.aToken.approve(
      fixture.validatorPool.address,
      stakeAmountATokenWei.mul(validators.length)
    );
    await fixture.aToken.approve(
      fixture.publicStaking.address,
      stakeAmountATokenWei.mul(validators.length)
    );

    for (const validator of validatorsSnapshots) {
      const tx = await fixture.publicStaking
        .connect(adminSigner)
        .mintTo(validator.address, stakeAmountATokenWei, lockTime);
      const tokenId = getTokenIdFromTx(tx);
      stakingTokenIds.push(tokenId);
      await fixture.publicStaking
        .connect(await getValidatorEthAccount(validator))
        .setApprovalForAll(fixture.validatorPool.address, true);
    }

    await fixture.validatorPool
      .connect(adminSigner)
      .registerValidators(validators, stakingTokenIds);
  });

  it("Does not allow snapshot if sender is not validator", async function () {
    const junkData =
      "0x0000000000000000000000000000000000000000000000000000006d6168616d";
    await expect(
      fixture.snapshots.connect(randomSigner).snapshot(junkData, junkData)
    )
      .to.be.revertedWithCustomError(fixture.snapshots, "OnlyValidatorsAllowed")
      .withArgs(randomUser.address);
  });

  it("Does not allow snapshot consensus is not running", async function () {
    const junkData =
      "0x0000000000000000000000000000000000000000000000000000006d6168616d";
    const validValidator = await getValidatorEthAccount(validatorsSnapshots[0]);
    await expect(
      fixture.snapshots.connect(validValidator).snapshot(junkData, junkData)
    ).to.be.revertedWithCustomError(fixture.snapshots, "ConsensusNotRunning");
  });
});
