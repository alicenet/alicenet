import {
  Fixture,
  getValidatorEthAccount,
  getFixture,
  factoryCallAny,
} from "../../setup";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import { BigNumber, ContractTransaction, Signer } from "ethers";
import {
  commitSnapshots,
  createValidators,
  getCurrentState,
  showState,
  stakeValidators,
} from "../setup";

describe("ValidatorPool: Slashing logic", async () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  let maxNumValidators = 4; //default number
  let stakeAmount = 20000;
  let validators: string[];
  let stakingTokenIds: BigNumber[];

  beforeEach(async function () {
    fixture = await getFixture(false, true, true);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
  });

  it("Minor slash a validator", async function () {
    let reward = 1;
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      ethers.utils.parseEther("" + reward),
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    await showState("After registering", await getCurrentState(fixture, validators));
    await ethdkg.minorSlash(validators[0], validators[1]);
    await showState("After minor slashing", await getCurrentState(fixture, validators));
    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT -= 1;
    expectedState.ValidatorPool.StakeNFT += 1;

    // Expect infringer to loose the validator position
    expectedState.StakeNFT.MAD += stakeAmount;
    expectedState.ValidatorNFT.MAD -= stakeAmount;
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.StakeNFT.MAD -= 1;
    expectedState.validators[1].MAD += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Major slash a validator", async function () {
    let reward = 3;
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      ethers.utils.parseEther("" + reward),
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[1]);
    await showState("After major slashing", await getCurrentState(fixture, validators));
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.ValNFT -= 1;
    // Expect reward to be transferred from ValidatorNFT to disputer
    expectedState.ValidatorNFT.MAD -= reward;
    expectedState.validators[1].MAD += reward;
    // Expect infringer to be unregistered, not in exiting queue and not accusable
    expectedState.validators[0].Reg = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Acc = false;
    let currentState = await getCurrentState(fixture, validators);
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
    await ethdkg.setConsensusRunning();
    expectedState = await getCurrentState(fixture, validators);
    for (let index = 1; index <= 3; index++) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshots[index]))
        .collectProfits();
      expectedState.validators[index].MAD +=
        (stakeAmount - reward) / (maxNumValidators - 1);
    }
    expectedState.ValidatorNFT.MAD -= stakeAmount - reward;
    currentState = await getCurrentState(fixture, validators);
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Minor slash a validator until he has no more funds", async function () {
    //Set reward to 1 MadToken
    let reward = 1;
    // Set infringer and disputer validators
    let infringer = validators[0];
    let disputer = validators[1];
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      ethers.utils.parseEther("" + reward),
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    await showState("After registering", await getCurrentState(fixture, validators));
    await ethdkg.minorSlash(infringer, disputer);
    await showState("After minor slashing", await getCurrentState(fixture, validators));
    await commitSnapshots(fixture, 4);
    // Prepare arrays for re-registering
    validatorsSnapshots.length = 1;
    stakingTokenIds = [];
    for (const validator of validatorsSnapshots) {
      let claimTx = (await fixture.validatorPool
        .connect(await getValidatorEthAccount(validator))
        .claimExitingNFTPosition()) as ContractTransaction;
      await showState("After claiming", await getCurrentState(fixture, validators));
      let receipt = await ethers.provider.getTransactionReceipt(claimTx.hash);
      //When a token is claimed it gets burned an minted so validator gets a new tokenId so we need to update array for re-registration
      const tokenId = BigNumber.from(receipt.logs[0].topics[3]);
      stakingTokenIds.push(tokenId);
      // Transfer to factory for re-registering
      await fixture.stakeNFT
        .connect(await getValidatorEthAccount(validator))
        .transferFrom(validator.address, fixture.factory.address, tokenId);
      await showState("After transferring", await getCurrentState(fixture, validators));
      // Approve new tokensIds
      await factoryCallAny(fixture, "stakeNFT", "approve", [
        fixture.validatorPool.address,
        tokenId,
      ]);
    }
    validators.length = 1;
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        validators,
        stakingTokenIds,
      ])
    ).to.be.revertedWith(
      "ValidatorStakeNFT: Error, the Stake position doesn't have enough funds!"
    );
  });
});
