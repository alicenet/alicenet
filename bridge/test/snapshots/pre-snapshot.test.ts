import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
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
  let randomUser: SignerWithAddress;
  let randomSigner: Signer;
  const stakeAmount = 20000;
  const stakeAmountALCAWei = ethers.utils.parseUnits(
    stakeAmount.toString(),
    18
  );
  const lockTime = 1;

  async function deployFixture() {
    const validators = [];
    const stakingTokenIds = [];
    const fixture = await getFixture(true, false);
    const [admin, , , , , randomUser] = fixture.namedSigners;
    const adminSigner = await getValidatorEthAccount(admin.address);
    const randomSigner = await getValidatorEthAccount(randomUser.address);

    for (const validator of validatorsSnapshots) {
      validators.push(validator.address);
    }

    await fixture.alca.approve(
      fixture.validatorPool.address,
      stakeAmountALCAWei.mul(validators.length)
    );
    await fixture.alca.approve(
      fixture.publicStaking.address,
      stakeAmountALCAWei.mul(validators.length)
    );

    for (const validator of validatorsSnapshots) {
      const tx = await fixture.publicStaking
        .connect(adminSigner)
        .mintTo(validator.address, stakeAmountALCAWei, lockTime);
      const tokenId = getTokenIdFromTx(tx);
      stakingTokenIds.push(tokenId);
      await fixture.publicStaking
        .connect(await getValidatorEthAccount(validator))
        .setApprovalForAll(fixture.validatorPool.address, true);
    }

    await fixture.validatorPool
      .connect(adminSigner)
      .registerValidators(validators, stakingTokenIds);

    return {
      fixture,
      admin,
      randomUser,
      adminSigner,
      randomSigner,
      validators,
      stakingTokenIds,
    };
  }

  beforeEach(async function () {
    ({ fixture, randomUser, randomSigner } = await loadFixture(deployFixture));
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
