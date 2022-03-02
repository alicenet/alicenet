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
import { BigNumber } from "ethers";
import {
  createValidators,
  getCurrentState,
  showState,
  stakeValidators,
} from "../setup";

describe("ValidatorPool: Unregistration logic", async () => {
  let fixture: Fixture;
  let maxNumValidators = 4; //default number
  let stakeAmount = 20000;

  beforeEach(async function () {
    fixture = await getFixture(false, true, true);
  });

  it("Should not allow unregistering of non-validators (even in the middle of array of validators)", async function () {
    let validators = await createValidators(fixture, validatorsSnapshots);
    let stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let newValidators = validators
    //Set a non validator address in the middle of array for un-registering
    newValidators[1] = "0x000000000000000000000000000000000000dEaD";
    await expect(
      factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
        newValidators,
      ])
    ).to.be.revertedWith("ValidatorPool: Address is not a validator_!");
  });

  it("Should not allow unregistering if consensus or an ETHDKG round is running", async function () {
    let validators = await createValidators(fixture, validatorsSnapshots);
    let stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    // Complete ETHDKG Round
    await factoryCallAny(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await expect(
      factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
        validators,
      ])
    ).to.be.revertedWith(
      "ValidatorPool: Error Madnet Consensus should be halted!"
    );
  });

  it("Should not allow unregistering more addresses that in the pool", async function () {
    let validators = await createValidators(fixture, validatorsSnapshots);
    let stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    //Add an extraother validator to unregister array
    validators.push("0x000000000000000000000000000000000000dEaD");
    stakingTokenIds.push(BigNumber.from(0));
    await expect(
      factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
        validators,
      ])
    ).to.be.revertedWith(
      "ValidatorPool: There are not enough validators to be removed!"
    );
  });

  it("Should not allow registering an address that was unregistered and didn’t claim is stakeNFT position (i.e still in the exitingQueue).", async function () {
    let validators = await createValidators(fixture, validatorsSnapshots);
    let stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    showState("after registering", await getCurrentState(fixture, validators));
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    showState(
      "after un-registering",
      await getCurrentState(fixture, validators)
    );
    let newValidators = await createValidators(fixture, validatorsSnapshots);
    let newStakeNFTIds = await stakeValidators(fixture, newValidators);
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        validators,
        newStakeNFTIds,
      ])
    ).to.be.revertedWith(
      "ValidatorPool: Address is already a validator or it is in the exiting line!"
    );
  });

  it("Should successfully unregister validators if all conditions are met", async function () {
    let validators = await createValidators(fixture, validatorsSnapshots);
    let stakingTokenIds = await stakeValidators(fixture, validators);
    let expectedState = await getCurrentState(fixture, validators);
    //Expect that NFT are transferred from ValidatorPool to Factory
    validators.map((_, index) => {
      expectedState.ValidatorPool.StakeNFT++;
      expectedState.Factory.StakeNFT--;
      expectedState.validators[index].Acc = true
      expectedState.validators[index].ExQ = true
    });
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    //  await showState("Current state after registering", await getCurrentState(fixture))
    await showState("after registering", expectedState);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await showState("after unregistering", expectedState);
    let currentState = await getCurrentState(fixture, validators);
    await showState(
      "Expected state after registering/un-registering",
      expectedState
    );
    await showState(
      "Current state after registering/un-registering",
      currentState
    );
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("The validatorNFT position was correctly burned and the state of the contract was correctly set (e.g user should not be a validator but should be accusable).", async function () {
    let validators = await createValidators(fixture, validatorsSnapshots);
    let stakingTokenIds = await stakeValidators(fixture, validators);
    let expectedState = await getCurrentState(fixture, validators);
    //Expect that NFT are transferred to ValidatorPool as StakeNFTs
    validators.map((_, index) => {
      expectedState.ValidatorPool.StakeNFT++;
      expectedState.Factory.StakeNFT--;
      expectedState.validators[index].Acc = true
      expectedState.validators[index].ExQ = true
    });
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "unregisterAllValidators");
    let currentState = await getCurrentState(fixture, validators);
    await showState(
      "Expected state after unregister all validators",
      currentState
    );
    await showState(
      "Expected state after unregister all validators",
      expectedState
    );
    expect(currentState).to.be.deep.equal(expectedState);
    expect(
      await fixture.validatorPool.isAccusable(validatorsSnapshots[0].address)
    ).to.be.true;
  });

  it("Do an ether and Madtoken deposit for the VALIDATORNFT contract before unregistering, but don’t collect the profits", async function () {
    let validators = await createValidators(fixture, validatorsSnapshots);
    let stakingTokenIds = await stakeValidators(fixture, validators);
    const [admin, , ,] = fixture.namedSigners;
    let adminSigner = await getValidatorEthAccount(admin.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await showState(
      "After registering:",
      await getCurrentState(fixture, validators)
    );
    let eths = 4;
    let mads = 4;
    await fixture.validatorNFT.connect(adminSigner).depositEth(42, {
      value: ethers.utils.parseEther(eths.toString()),
    });
    await fixture.madToken
      .connect(adminSigner)
      .approve(fixture.validatorNFT.address, ethers.utils.parseEther(mads.toString()));
    await fixture.validatorNFT
      .connect(adminSigner)
      .depositToken(42, ethers.utils.parseEther(mads.toString()));
    let expectedState = await getCurrentState(fixture, validators);
    validators.map((element, index) => {
      expectedState.ValidatorNFT.ETH -= eths / maxNumValidators;
      expectedState.validators[index].ETH += eths / maxNumValidators;
      expectedState.ValidatorNFT.MAD -= mads / maxNumValidators;
      expectedState.validators[index].MAD += mads / maxNumValidators;
      expectedState.validators[index].Reg = false;
      expectedState.validators[index].ExQ = true;
    });
    expectedState.ValidatorNFT.MAD -= stakeAmount * maxNumValidators;
    expectedState.StakeNFT.MAD += stakeAmount * maxNumValidators;
    expectedState.ValidatorPool.ValNFT -= maxNumValidators;
    expectedState.ValidatorPool.StakeNFT += maxNumValidators;

    await showState(
      "After deposit:",
      await getCurrentState(fixture, validators)
    );
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await showState(
      "After unregistering:",
      await getCurrentState(fixture, validators)
    );
    let currentState = await getCurrentState(fixture, validators);
    await showState(
      "Expected state after claiming exiting NFT position",
      expectedState
    );
    await showState(
      "Current state after claiming exiting NFT position",
      currentState
    );
    expect(currentState).to.be.deep.equal(expectedState);
  });
});
