import {
  Fixture,
  getValidatorEthAccount,
  getFixture,
  factoryCallAny,
} from "../../setup";
import { expect } from "../../chai-setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import { BigNumber, Signer } from "ethers";
import {
  commitSnapshots,
  createValidators,
  getCurrentState,
  showState,
  stakeValidators,
} from "../setup";

describe("ValidatorPool: Claiming logic", async () => {
  let fixture: Fixture;
  let stakeAmount = 20000;
  let validators: string[];
  let stakingTokenIds: BigNumber[];
  let adminSigner: Signer;

  beforeEach(async function () {
    fixture = await getFixture(false, true, true);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
  });

  it("Should successfully claim exiting NFT positions of all validators", async function () {
    //As this is a complete cycle, expect the initial state to be exactly the same as the final state
    let expectedState = await getCurrentState(fixture, validators);
    validators.map((element, index) => {
      expectedState.Factory.StakeNFT--;
      expectedState.validators[index].NFT++;
    });
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await commitSnapshots(fixture, 4);
    for (const validatorsSnapshot of validatorsSnapshots) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshot))
        .claimExitingNFTPosition();
    }
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

  it("After claiming, register the user again with a new stakenft position", async function () {
    //As this is a complete cycle, expect the initial state to be exactly the same as the final
    //state
    let expectedState = await getCurrentState(fixture, validators);
    validators.map((_, index) => {
      expectedState.Factory.StakeNFT--;
      expectedState.validators[index].NFT++;
      expectedState.validators[index].Acc = true;
      expectedState.validators[index].Reg = true;
      //Validators already start with stakeAmount (see test config)
      expectedState.validators[index].MAD = stakeAmount * 2;
      //New Staking
      expectedState.ValidatorPool.ValNFT++;
      expectedState.Admin.MAD -= stakeAmount * 2;
    });
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
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
    //Re-initialize validators
    let newValidators = await createValidators(fixture, validatorsSnapshots);
    let newStakeNFTIDs = await stakeValidators(fixture, newValidators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      newStakeNFTIDs,
    ]);
    let currentState = await getCurrentState(fixture, validators);
    //Expect that validators funds are transferred again to ValidatorNFT
    expectedState.ValidatorNFT.MAD += stakeAmount * validators.length;
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

  it("Should not allow users to claim position before the right time", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    for (const validator of validatorsSnapshots) {
      expect(
        fixture.validatorPool
          .connect(await getValidatorEthAccount(validator))
          .claimExitingNFTPosition()
      ).to.be.revertedWith(
        "ValidatorPool: The waiting period is not over yet!"
      );
    }
  });

  it("After the claim period, the user should be able to claim its stakenft position", async function () {
    fixture = await getFixture(true, true, true);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await showState(
      "After un -registering",
      await getCurrentState(fixture, validators)
    );
    let expectedState = await getCurrentState(fixture, validators);
    validators.map((_, index) => {
      //New Staking
      // expectedState.ValidatorPool.StakeNFT--;
      // expectedState.Factory.StakeNFT--;
    });
    await commitSnapshots(fixture, 4);
    for (const validator of validatorsSnapshots) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validator))
        .claimExitingNFTPosition();
    }
    await showState(
      "After claiming",
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

  it("Should not allow a non-owner try to get stakenft position in the exitingQueue", async function () {
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
      validators,
    ]);
    await commitSnapshots(fixture, 4);
    await expect(
      fixture.validatorPool.connect(adminSigner).claimExitingNFTPosition()
    ).to.be.revertedWith("ValidatorPool: Address not in the exitingQueue!");
  });
});
