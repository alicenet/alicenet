import {
  Fixture,
  getValidatorEthAccount,
  getFixture,
  factoryCallAny,
} from "../../setup";
import { completeETHDKGRound } from "../../ethdkg/setup";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import { BigNumber, Signer } from "ethers";
import {
  createValidators,
  getCurrentState,
  showState,
  stakeValidators,
  commitSnapshots,
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
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    //Simulate 4 ETH from admin earned by pool after validators registration
    expectedState.Admin.ETH -= 4;
    await fixture.validatorNFT.connect(adminSigner).depositEth(42, {
      value: ethers.utils.parseEther("4.0"),
    });
    //Expect ValidatorNFT balance to increment by earnings
    expectedState.ValidatorNFT.ETH += 4;
    // Complete ETHDKG Round
    await showState("After deposit:", expectedState);
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await fixture.validatorPool
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .collectProfits();
    // Expect that a fraction of the earnings (1/4 validators) to be transfer from ValidatorNFT to collecting validator
    expectedState.ValidatorNFT.ETH -= 1;
    expectedState.validators[0].ETH += 1;
    let currentState = await getCurrentState(fixture, validators);
    await showState("Expected state after collect profit", expectedState);
    await showState("Current state after collect profit", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Check if profit leftovers are sent back to user when unregistering", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    let maxNumValidators = validatorsSnapshots.length;
    let eths = 4;
    let mads = 4;
    await fixture.validatorNFT.connect(adminSigner).depositEth(42, {
      value: ethers.utils.parseEther(`${eths}`),
    });
    await fixture.madToken
      .connect(adminSigner)
      .approve(fixture.validatorNFT.address, ethers.utils.parseEther("4"));
    await fixture.validatorNFT
      .connect(adminSigner)
      .depositToken(42, ethers.utils.parseEther(`${mads}`));
    //Expect ValidatorNFT balance to increment by earnings
    expectedState.ValidatorNFT.ETH += 4;
    expectedState.ValidatorNFT.MAD += 4;
    expectedState.Admin.MAD -= 4;
    expectedState.Admin.ETH -= 4;
    // Complete ETHDKG Round
    let currentState = await getCurrentState(fixture, validators);
    await showState("Expected state after deposit", expectedState);
    await showState("Current state after deposit", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    for (let index = 0; index < validators.length; index++) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshots[index]))
        .collectProfits();
    }
    validators.map((_, index) => {
      expectedState.ValidatorNFT.ETH -= eths / maxNumValidators;
      expectedState.validators[index].ETH += eths / maxNumValidators;
      expectedState.ValidatorNFT.MAD -= mads / maxNumValidators;
      expectedState.validators[index].MAD += mads / maxNumValidators;
      expectedState.validators[index].Reg = true;
      expectedState.validators[index].Acc = true;
    });
    currentState = await getCurrentState(fixture, validators);
    await showState("Expected state after collect profit", expectedState);
    await showState("Current state after collect profit", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  xit("Skim excess of tokens end ether in the contract", async function () {
    let expectedState = await getCurrentState(fixture, validators);
    // await factoryCallAny(fixture, "validatorPool", "registerValidators", [
    //   validators,
    //   stakingTokenIds,
    // ]);
    // await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
    //   validators,
    // ]);
    // await commitSnapshots(fixture, 4);
    // for (const validator of validatorsSnapshots) {
    //   await fixture.validatorPool
    //     .connect(await getValidatorEthAccount(validator))
    //     .claimExitingNFTPosition();
    // }
    await showState(
      "After claiming:",
      await getCurrentState(fixture, validators)
    );
    console.log(
      await fixture.stakeNFT
        .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
        .ownerOf(1)
    );

    await fixture.stakeNFT
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .collectEthTo(validatorsSnapshots[0].address, 1);
    console.log(1);
    await fixture.stakeNFT.collectTokenTo(validatorsSnapshots[1].address, 2);
    let currentState = await getCurrentState(fixture, validators);
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });
});
