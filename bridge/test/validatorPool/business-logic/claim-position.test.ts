import { BigNumber, Signer } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import {
  factoryCallAny,
  Fixture,
  getFixture,
  getValidatorEthAccount,
  mineBlocks,
} from "../../setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import {
  burnStakeTo,
  commitSnapshots,
  createValidators,
  getCurrentState,
  showState,
  stakeValidators,
} from "../setup";

describe("ValidatorPool: Claiming logic", async () => {
  let fixture: Fixture;
  let stakeAmount: bigint;
  let validators: string[];
  let stakingTokenIds: BigNumber[];
  let adminSigner: Signer;

  beforeEach(async function () {
    fixture = await getFixture(false, true, true);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
    stakeAmount = (await fixture.validatorPool.getStakeAmount()).toBigInt();
  });

  it("Should successfully claim exiting NFT positions of all validators", async function () {
    // As this is a complete cycle, expect the initial state to be exactly the same as the final state
    const expectedState = await getCurrentState(fixture, validators);
    for (let index = 0; index < expectedState.validators.length; index++) {
      expectedState.Factory.PublicStaking--;
      expectedState.validators[index].NFT++;
    }
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await mineBlocks(1n);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await commitSnapshots(fixture, 4);
    for (const validatorsSnapshot of validatorsSnapshots) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshot))
        .claimExitingNFTPosition();
    }
    const currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Should successfully claim exiting NFT positions of all validators even with excess of ETH and tokens", async function () {
    // Mint a publicStaking and burn it to the ValidatorPool contract. Besides a contract self destructing
    // itself, this is a method to send eth accidentally to the validatorPool contract
    const etherAmount = ethers.utils.parseEther("1");
    const madTokenAmount = ethers.utils.parseEther("2");
    await burnStakeTo(fixture, etherAmount, madTokenAmount, adminSigner);
    // As this is a complete cycle, expect the initial state to be exactly the same as the final state
    const expectedState = await getCurrentState(fixture, validators);
    for (let index = 0; index < expectedState.validators.length; index++) {
      expectedState.Factory.PublicStaking--;
      expectedState.validators[index].NFT++;
    }
    expectedState.PublicStaking.ETH = BigInt(0);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await mineBlocks(1n);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await commitSnapshots(fixture, 4);
    for (const validatorsSnapshot of validatorsSnapshots) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshot))
        .claimExitingNFTPosition();
    }
    const currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("After claiming, register the user again with a new PublicStaking position", async function () {
    // As this is a complete cycle, expect the initial state to be exactly the same as the final
    // state
    const expectedState = await getCurrentState(fixture, validators);
    for (let index = 0; index < expectedState.validators.length; index++) {
      expectedState.Factory.PublicStaking--;
      expectedState.validators[index].NFT++;
      expectedState.validators[index].Acc = true;
      expectedState.validators[index].Reg = true;
      // Validators already start with stakeAmount (see test config)
      expectedState.validators[index].MAD = stakeAmount * BigInt(2);
      // New Staking
      expectedState.ValidatorPool.ValNFT++;
      expectedState.Admin.MAD -= stakeAmount * BigInt(2);
    }
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await mineBlocks(1n);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await commitSnapshots(fixture, 4);
    for (const validatorsSnapshot of validatorsSnapshots) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshot))
        .claimExitingNFTPosition();
    }
    await showState(
      "After claiming:",
      await getCurrentState(fixture, validators)
    );
    // Re-initialize validators
    const newValidators = await createValidators(fixture, validatorsSnapshots);
    const newPublicStakingIDs = await stakeValidators(fixture, newValidators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      newPublicStakingIDs,
    ]);
    const currentState = await getCurrentState(fixture, validators);
    // Expect that validators funds are transferred again to ValidatorStaking
    expectedState.ValidatorStaking.MAD +=
      BigInt(stakeAmount) * BigInt(validators.length);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Should not allow users to claim position before the right time", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await mineBlocks(1n);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    for (const validator of validatorsSnapshots) {
      await expect(
        fixture.validatorPool
          .connect(await getValidatorEthAccount(validator))
          .claimExitingNFTPosition()
      ).to.be.revertedWith(
        "ValidatorPool: The waiting period is not over yet!"
      );
    }
  });

  it("Should not allow a non-owner try to get PublicStaking position in the exitingQueue", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await mineBlocks(1n);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await commitSnapshots(fixture, 4);
    await expect(
      fixture.validatorPool.connect(adminSigner).claimExitingNFTPosition()
    ).to.be.revertedWith("ValidatorPool: Address not in the exitingQueue!");
  });
});
