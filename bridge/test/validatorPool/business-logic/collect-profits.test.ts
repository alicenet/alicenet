import { BigNumber, Signer } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import { completeETHDKGRound } from "../../ethdkg/setup";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getValidatorEthAccount,
} from "../../setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import {
  burnStakeTo,
  createValidators,
  getCurrentState,
  showState,
  stakeValidators,
} from "../setup";

describe("ValidatorPool: Collecting logic", async function () {
  let fixture: Fixture;
  let adminSigner: Signer;
  let validators: string[];
  let stakingTokenIds: BigNumber[];

  beforeEach(async function () {
    fixture = await getFixture(false, true, true);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
  });

  it("Should successfully collect profit of validators", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    const eths = ethers.utils.parseEther("4.0").toBigInt();
    const expectedState = await getCurrentState(fixture, validators);
    await fixture.validatorStaking.connect(adminSigner).depositEth(42, {
      value: eths,
    });
    // Expect ValidatorStaking balance to increment by earnings
    expectedState.ValidatorStaking.ETH += eths;
    // Complete ETHDKG Round
    await showState("After deposit:", expectedState);
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await fixture.validatorPool
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .collectProfits();
    // Expect that a fraction of the earnings (1/4 validators) to be transfer from ValidatorStaking to collecting validator
    expectedState.ValidatorStaking.ETH -= eths / BigInt(4);
    const currentState = await getCurrentState(fixture, validators);
    await showState("Expected state after collect profit", expectedState);
    await showState("Current state after collect profit", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Should successfully collect profit of validators even with excess of Eth and Tokens", async function () {
    // Mint a publicStaking and burn it to the ValidatorPool contract. Besides a contract self destructing
    // itself, this is a method to send eth accidentally to the validatorPool contract
    const etherAmount = ethers.utils.parseEther("1");
    const aTokenAmount = ethers.utils.parseEther("2");
    await burnStakeTo(fixture, etherAmount, aTokenAmount, adminSigner);

    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    const eths = ethers.utils.parseEther("4.0").toBigInt();
    const expectedState = await getCurrentState(fixture, validators);
    await fixture.validatorStaking.connect(adminSigner).depositEth(42, {
      value: eths,
    });
    // Expect ValidatorStaking balance to increment by earnings
    expectedState.ValidatorStaking.ETH += eths;
    // Complete ETHDKG Round
    await showState("After deposit:", expectedState);
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await fixture.validatorPool
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .collectProfits();
    // Expect that a fraction of the earnings (1/4 validators) to be transfer from ValidatorStaking to collecting validator
    expectedState.ValidatorStaking.ETH -= eths / BigInt(4);
    const currentState = await getCurrentState(fixture, validators);
    await showState("Expected state after collect profit", expectedState);
    await showState("Current state after collect profit", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Check if profit leftovers are sent back to user when unregistering", async function () {
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    const expectedState = await getCurrentState(fixture, validators);
    const maxNumValidators = validatorsSnapshots.length;
    const eths = ethers.utils.parseEther(`4`).toBigInt();
    const atokens = ethers.utils.parseEther(`4`).toBigInt();
    await fixture.validatorStaking.connect(adminSigner).depositEth(42, {
      value: eths,
    });
    await fixture.aToken
      .connect(adminSigner)
      .approve(fixture.validatorStaking.address, atokens);
    await fixture.validatorStaking
      .connect(adminSigner)
      .depositToken(42, atokens);
    // Expect ValidatorStaking balance to increment by earnings
    expectedState.ValidatorStaking.ETH += eths;
    expectedState.ValidatorStaking.ATK += atokens;
    expectedState.Admin.ATK -= atokens;
    // Complete ETHDKG Round
    let currentState = await getCurrentState(fixture, validators);
    await showState("Expected state after deposit", expectedState);
    await showState("Current state after deposit", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    for (let index = 0; index < validators.length; index++) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshots[index]))
        .collectProfits();
    }
    for (let index = 0; index < expectedState.validators.length; index++) {
      expectedState.ValidatorStaking.ETH -= eths / BigInt(maxNumValidators);
      expectedState.ValidatorStaking.ATK = atokens / BigInt(maxNumValidators);
      expectedState.validators[index].ATK += atokens / BigInt(maxNumValidators);
      expectedState.validators[index].Reg = true;
      expectedState.validators[index].Acc = true;
    }
    currentState = await getCurrentState(fixture, validators);
    await showState("Expected state after collect profit", expectedState);
    await showState("Current state after collect profit", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });
});
