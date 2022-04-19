import { BigNumber, Signer } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import { completeETHDKGRound } from "../../ethdkg/setup";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getValidatorEthAccount,
  mineBlocks,
} from "../../setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import {
  burnStakeTo,
  createValidators,
  getCurrentState,
  stakeValidators,
} from "../setup";

describe("ValidatorPool: Unregistration logic", async () => {
  let fixture: Fixture;
  let stakeAmount: bigint;
  let validators: Array<string>;
  let stakingTokenIds: Array<BigNumber>;
  let adminSigner: Signer;

  beforeEach(async function () {
    fixture = await getFixture(false, true, true);
    stakeAmount = (await fixture.validatorPool.getStakeAmount()).toBigInt();
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
  });

  it("Should not allow unregistering of non-validators (even in the middle of array of validators)", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    const newValidators = validators;
    // Set a non validator address in the middle of array for un-registering
    newValidators[1] = "0x000000000000000000000000000000000000dEaD";
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "unregisterValidators", [
        newValidators,
      ])
    ).to.be.revertedWith("817");
  });

  it("Should not allow unregistering if consensus or an ETHDKG round is running", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    // Complete ETHDKG Round
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "unregisterValidators", [
        validators,
      ])
    ).to.be.revertedWith("802");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "unregisterValidators", [
        validators,
      ])
    ).to.be.revertedWith("801");
  });

  it("Should not allow unregistering more addresses that in the pool", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    // Add an extra validator to unregister array
    validators.push("0x000000000000000000000000000000000000dEaD");
    stakingTokenIds.push(BigNumber.from(0));
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "unregisterValidators", [
        validators,
      ])
    ).to.be.revertedWith("808");
  });

  it("Should not allow registering an address that was unregistered and didn’t claim is publicStaking position", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterValidators",
      [validators]
    );
    const newValidators = await createValidators(fixture, validatorsSnapshots);
    const newPublicStakingIds = await stakeValidators(fixture, newValidators);
    await expect(
      factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
        validators,
        newPublicStakingIds,
      ])
    ).to.be.revertedWith("816");
  });

  it("Should successfully unregister validators if all conditions are met", async function () {
    const expectedState = await getCurrentState(fixture, validators);
    // Expect that NFT are transferred from ValidatorPool to Factory
    for (let index = 0; index < validators.length; index++) {
      expectedState.ValidatorPool.PublicStaking++;
      expectedState.Factory.PublicStaking--;
      expectedState.validators[index].Acc = true;
      expectedState.validators[index].ExQ = true;
    }
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterValidators",
      [validators]
    );
    const currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Should successfully unregister validators if all conditions are met and there are excess of Eth and Tokens", async function () {
    // Mint a publicStaking and burn it to the ValidatorPool contract. Besides a contract self destructing
    // itself, this is a method to send eth accidentally to the validatorPool contract
    const etherAmount = ethers.utils.parseEther("1");
    const aTokenAmount = ethers.utils.parseEther("2");
    await burnStakeTo(fixture, etherAmount, aTokenAmount, adminSigner);

    const expectedState = await getCurrentState(fixture, validators);
    expectedState.PublicStaking.ETH = BigInt(0);
    // Expect that NFT are transferred from ValidatorPool to Factory
    for (let index = 0; index < validators.length; index++) {
      expectedState.ValidatorPool.PublicStaking++;
      expectedState.Factory.PublicStaking--;
      expectedState.validators[index].Acc = true;
      expectedState.validators[index].ExQ = true;
    }
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterValidators",
      [validators]
    );
    const currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Do an ether and token deposit for the validatorStaking contract before unregistering, but don’t collect the profits", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    const eths = ethers.utils.parseEther("4").toBigInt();
    const atokens = ethers.utils.parseEther("4").toBigInt();
    await fixture.validatorStaking.connect(adminSigner).depositEth(42, {
      value: eths,
    });
    await fixture.aToken
      .connect(adminSigner)
      .approve(fixture.validatorStaking.address, atokens);
    await fixture.validatorStaking
      .connect(adminSigner)
      .depositToken(42, atokens);
    const expectedState = await getCurrentState(fixture, validators);

    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterValidators",
      [validators]
    );

    for (let index = 0; index < validators.length; index++) {
      expectedState.ValidatorStaking.ETH -= eths / BigInt(validators.length);
      expectedState.ValidatorStaking.ATK -= atokens / BigInt(validators.length);
      expectedState.validators[index].ATK +=
        atokens / BigInt(validators.length);
      expectedState.validators[index].Reg = false;
      expectedState.validators[index].ExQ = true;
    }
    expectedState.ValidatorStaking.ATK -= BigInt(validators.length);
    expectedState.ValidatorVault.ATK -=
      BigInt(validators.length) * stakeAmount - BigInt(validators.length);
    expectedState.PublicStaking.ATK += stakeAmount * BigInt(validators.length);
    expectedState.ValidatorPool.ValNFT -= BigInt(validators.length);
    expectedState.ValidatorPool.PublicStaking += BigInt(validators.length);

    const currentState = await getCurrentState(fixture, validators);

    expect(currentState).to.be.deep.equal(expectedState);
  });
});
