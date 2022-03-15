import { BigNumber, ContractTransaction, Signer } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import {
  factoryCallAny,
  Fixture,
  getFixture,
  getValidatorEthAccount,
} from "../../setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import {
  burnStakeTo,
  commitSnapshots,
  createValidators,
  getCurrentState,
  getStakeNFTFromMinorSlashEvent,
  showState,
  stakeValidators,
} from "../setup";

describe("ValidatorPool: Slashing logic", async () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  let validators: string[];
  let stakingTokenIds: BigNumber[];
  let stakeAmount: bigint;

  beforeEach(async () => {
    fixture = await getFixture(false, true, true);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
    validators = await createValidators(fixture, validatorsSnapshots);
    stakingTokenIds = await stakeValidators(fixture, validators);
    stakeAmount = (await fixture.validatorPool.getStakeAmount()).toBigInt();
  });

  it("Minor slash a validator", async function () {
    let reward = ethers.utils.parseEther("1").toBigInt();
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    await showState(
      "After registering",
      await getCurrentState(fixture, validators)
    );
    await ethdkg.minorSlash(validators[0], validators[1]);
    await showState(
      "After minor slashing",
      await getCurrentState(fixture, validators)
    );
    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.StakeNFT++;

    // Expect infringer to loose the validator position
    expectedState.StakeNFT.MAD += stakeAmount;
    expectedState.ValidatorNFT.MAD -= stakeAmount;
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.StakeNFT.MAD -= reward;
    expectedState.validators[1].MAD += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Minor slash a validator then he leaves the pool", async function () {
    let reward = ethers.utils.parseEther("1").toBigInt();
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    await showState(
      "After registering",
      await getCurrentState(fixture, validators)
    );
    let tx = await ethdkg.minorSlash(validators[0], validators[1]);
    let newStakeNFT = await getStakeNFTFromMinorSlashEvent(tx);
    expect(newStakeNFT).to.be.gt(
      BigInt(0),
      "New StakeNFT position was not created properly!"
    );
    await showState(
      "After minor slashing",
      await getCurrentState(fixture, validators)
    );
    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.StakeNFT++;

    // Expect infringer to loose the validator position
    expectedState.StakeNFT.MAD += stakeAmount;
    expectedState.ValidatorNFT.MAD -= stakeAmount;
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.StakeNFT.MAD -= reward;
    expectedState.validators[1].MAD += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed assert after minor slash"
    );

    await commitSnapshots(fixture, 4);

    let claimTx = (await fixture.validatorPool
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .claimExitingNFTPosition()) as ContractTransaction;
    let receipt = await ethers.provider.getTransactionReceipt(claimTx.hash);
    let exitingNFT = BigNumber.from(receipt.logs[0].topics[3]);
    let blockNumber = receipt.blockNumber;

    expect(exitingNFT.toBigInt()).to.be.equal(
      newStakeNFT,
      "Failed claiming NFT position!"
    );

    expect(
      (await fixture.stakeNFT.ownerOf(newStakeNFT)).toLowerCase()
    ).to.be.equal(validatorsSnapshots[0].address.toLowerCase());

    let position = await fixture.stakeNFT.getPosition(newStakeNFT);
    expect(position.freeAfter.toBigInt()).to.be.equal(
      (await fixture.validatorPool.POSITION_LOCK_PERIOD()).toBigInt() +
        BigInt(blockNumber)
    );
  });

  it("Minor slash a validator then major slash it", async function () {
    let reward = ethers.utils.parseEther("10000").toBigInt();
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    let tx = await ethdkg.minorSlash(validators[0], validators[1]);
    let newStakeNFT = await getStakeNFTFromMinorSlashEvent(tx);
    expect(newStakeNFT).to.be.gt(
      BigInt(0),
      "New StakeNFT position was not created properly!"
    );
    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.StakeNFT++;

    // Expect infringer to loose the validator position
    expectedState.StakeNFT.MAD += stakeAmount - reward;
    expectedState.ValidatorNFT.MAD -= stakeAmount;
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.validators[1].MAD += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed assert after minor slash"
    );

    expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[2]);
    currentState = await getCurrentState(fixture, validators);
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.StakeNFT--;
    // Expect reward to be transferred from ValidatorNFT to disputer
    expectedState.StakeNFT.MAD -= stakeAmount - reward;
    // the stakeamount minus the 2 rewards given to the 2 accusators should be redistributed to all
    // validators
    expectedState.ValidatorNFT.MAD += stakeAmount - BigInt(2) * reward;
    expectedState.validators[2].MAD += reward;
    // Expect infringer to be unregistered, not in exiting queue and not accusable
    expectedState.validators[0].Reg = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Acc = false;

    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed checking state after major slashing!"
    );
  });

  it("Should be able to Minor slash and major slash even with excess of eth and Token", async function () {
    // Mint a stakeNFT and burn it to the ValidatorPool contract. Besides a contract self destructing
    // itself, this is a method to send eth accidentally to the validatorPool contract
    let etherAmount = ethers.utils.parseEther("1");
    let madTokenAmount = ethers.utils.parseEther("2");
    await burnStakeTo(fixture, etherAmount, madTokenAmount, adminSigner);

    let reward = ethers.utils.parseEther("10000").toBigInt();
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    let tx = await ethdkg.minorSlash(validators[0], validators[1]);
    let newStakeNFT = await getStakeNFTFromMinorSlashEvent(tx);
    expect(newStakeNFT).to.be.gt(
      BigInt(0),
      "New StakeNFT position was not created properly!"
    );
    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.StakeNFT++;

    // Expect infringer to loose the validator position
    expectedState.StakeNFT.MAD += stakeAmount - reward;
    expectedState.ValidatorNFT.MAD -= stakeAmount;
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.validators[1].MAD += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed assert after minor slash"
    );

    expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[2]);
    currentState = await getCurrentState(fixture, validators);
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.StakeNFT--;
    // Expect reward to be transferred from ValidatorNFT to disputer
    expectedState.StakeNFT.MAD -= stakeAmount - reward;
    // the stakeamount minus the 2 rewards given to the 2 accusators should be redistributed to all
    // validators
    expectedState.ValidatorNFT.MAD += stakeAmount - BigInt(2) * reward;
    expectedState.validators[2].MAD += reward;
    // Expect infringer to be unregistered, not in exiting queue and not accusable
    expectedState.validators[0].Reg = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Acc = false;

    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed checking state after major slashing!"
    );
  });

  it("Minor slash a validator then major slash it with funds distribution in the middle", async function () {
    let reward = ethers.utils.parseEther("10000").toBigInt();
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);

    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let eths = 4;
    let mads = 20;
    await fixture.validatorNFT.connect(adminSigner).depositEth(42, {
      value: ethers.utils.parseEther(`${eths}`),
    });
    await fixture.madToken
      .connect(adminSigner)
      .approve(
        fixture.validatorNFT.address,
        ethers.utils.parseEther(`${mads}`)
      );
    await fixture.validatorNFT
      .connect(adminSigner)
      .depositToken(42, ethers.utils.parseEther(`${mads}`));
    let expectedState = await getCurrentState(fixture, validators);
    let tx = await ethdkg.minorSlash(validators[0], validators[1]);
    let newStakeNFT = await getStakeNFTFromMinorSlashEvent(tx);
    expect(newStakeNFT).to.be.gt(
      BigInt(0),
      "New StakeNFT position was not created properly!"
    );
    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.StakeNFT++;

    // Expect infringer to loose the validator position
    expectedState.StakeNFT.MAD += stakeAmount - reward;
    expectedState.ValidatorNFT.MAD -=
      stakeAmount + ethers.utils.parseEther(`${mads / 4}`).toBigInt();
    expectedState.ValidatorNFT.ETH -= ethers.utils
      .parseEther(`${eths / 4}`)
      .toBigInt();
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.validators[1].MAD +=
      reward + ethers.utils.parseEther("5").toBigInt(); // result from distribution
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed assert after minor slash"
    );
    // the validatorPool shouldn't have any balance
    expect(
      (
        await ethers.provider.getBalance(fixture.validatorPool.address)
      ).toBigInt()
    ).to.be.equal(BigInt(0), "ValidatorPool shouldn't have any balance");

    // minting 1 stakeNFt so all the profit funds are not moved from the stakenft contract at slashing
    // time
    let newStakeAmount = (stakeAmount - reward) * BigInt(3);
    await fixture.madToken
      .connect(adminSigner)
      .approve(fixture.stakeNFT.address, newStakeAmount);
    await fixture.stakeNFT.connect(adminSigner).mint(newStakeAmount);

    // Deposit now in the stakeNFT contract
    await fixture.stakeNFT.connect(adminSigner).depositEth(42, {
      value: ethers.utils.parseEther(`${eths}`),
    });
    await fixture.madToken
      .connect(adminSigner)
      .approve(fixture.stakeNFT.address, ethers.utils.parseEther(`${mads}`));
    await fixture.stakeNFT
      .connect(adminSigner)
      .depositToken(42, ethers.utils.parseEther(`${mads}`));

    expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[2]);
    currentState = await getCurrentState(fixture, validators);
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.StakeNFT--;
    expectedState.StakeNFT.ETH -= ethers.utils
      .parseEther(`${eths / 4}`)
      .toBigInt();
    // Expect reward to be transferred from ValidatorNFT to disputer
    expectedState.StakeNFT.MAD -=
      stakeAmount - reward + ethers.utils.parseEther(`${mads / 4}`).toBigInt();
    // the stakeamount minus the 2 rewards given to the 2 accusators should be redistributed to all
    // validators
    expectedState.ValidatorNFT.MAD += stakeAmount - BigInt(2) * reward;
    expectedState.validators[2].MAD +=
      reward + ethers.utils.parseEther(`${mads / 4}`).toBigInt();
    // Expect infringer to be unregistered, not in exiting queue and not accusable
    expectedState.validators[0].Reg = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Acc = false;

    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed checking state after major slashing!"
    );

    expect(
      (
        await ethers.provider.getBalance(fixture.validatorPool.address)
      ).toBigInt()
    ).to.be.equal(
      BigInt(0),
      "ValidatorPool shouldn't have any balance after major slash"
    );

    await ethdkg.setConsensusRunning();
    expectedState = await getCurrentState(fixture, validators);
    let collectedAmount =
      (stakeAmount - BigInt(2) * reward) /
        BigInt(validatorsSnapshots.length - 1) +
      ethers.utils.parseEther(`${mads / 4}`).toBigInt();
    for (let index = 1; index < validatorsSnapshots.length; index++) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshots[index]))
        .collectProfits();
      expectedState.validators[index].MAD += collectedAmount;
      expectedState.ValidatorNFT.ETH -= ethers.utils
        .parseEther(`${eths / 4}`)
        .toBigInt();
      expectedState.ValidatorNFT.MAD -= collectedAmount;
    }
    currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed in the assertion after collect Profits"
    );
  });

  it("Major slash a validator then validators collect profit", async function () {
    let reward = ethers.utils.parseEther("5000").toBigInt();
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
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
    await showState(
      "After major slashing",
      await getCurrentState(fixture, validators)
    );
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.ValNFT--;
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
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed checking state after major slashing!"
    );

    await ethdkg.setConsensusRunning();
    expectedState = await getCurrentState(fixture, validators);
    let collectedAmount =
      (stakeAmount - reward) / BigInt(validatorsSnapshots.length - 1);
    for (let index = 1; index < validatorsSnapshots.length; index++) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshots[index]))
        .collectProfits();
      expectedState.validators[index].MAD += collectedAmount;
      expectedState.ValidatorNFT.MAD -= collectedAmount;
    }
    currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed in the assertion after collect Profits"
    );
  });

  it("Should not allow major slash a validator twice", async function () {
    let reward = ethers.utils.parseEther("5000").toBigInt();
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
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
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.ValNFT--;
    // Expect reward to be transferred from ValidatorNFT to disputer
    expectedState.ValidatorNFT.MAD -= reward;
    expectedState.validators[1].MAD += reward;
    // Expect infringer to be unregistered, not in exiting queue and not accusable
    expectedState.validators[0].Reg = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Acc = false;
    let currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed checking state after major slashing!"
    );
    await expect(
      ethdkg.majorSlash(validators[0], validators[1])
    ).to.be.revertedWith(
      "ValidatorPool: DishonestValidator should be a validator or be in the exiting line!"
    );
  });

  it("Should not allow major/minor slash a person that it's not a validator", async function () {
    let reward = ethers.utils.parseEther("5000").toBigInt();
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    await expect(
      ethdkg.majorSlash(await adminSigner.getAddress(), validators[1])
    ).to.be.revertedWith(
      "ValidatorPool: DishonestValidator should be a validator or be in the exiting line!"
    );

    await expect(
      ethdkg.minorSlash(await adminSigner.getAddress(), validators[1])
    ).to.be.revertedWith(
      "ValidatorPool: DishonestValidator should be a validator or be in the exiting line!"
    );
  });

  it("Major slash a validator with disputer reward greater than stake Amount", async function () {
    let reward = (await fixture.validatorPool.getStakeAmount()).toBigInt();
    reward *= BigInt(2);
    //Set reward to 1 MadToken
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
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
    await showState(
      "After major slashing",
      await getCurrentState(fixture, validators)
    );
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.ValNFT--;
    // Expect reward to be transferred from ValidatorNFT to disputer
    expectedState.ValidatorNFT.MAD -= stakeAmount;
    expectedState.validators[1].MAD += stakeAmount;
    // Expect infringer to be unregistered, not in exiting queue and not accusable
    expectedState.validators[0].Reg = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Acc = false;
    let currentState = await getCurrentState(fixture, validators);
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed checking state after major slashing!"
    );

    await ethdkg.setConsensusRunning();
    expectedState = await getCurrentState(fixture, validators);
    let collectedAmount = BigInt(0);
    for (let index = 1; index < validatorsSnapshots.length; index++) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshots[index]))
        .collectProfits();
      expectedState.validators[index].MAD += collectedAmount;
      expectedState.ValidatorNFT.MAD -= collectedAmount;
    }
    currentState = await getCurrentState(fixture, validators);
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed in the assertion after collect Profits"
    );
  });

  it("Minor slash a validator until he has no more funds", async function () {
    //Set reward to 1 MadToken
    let reward = (await fixture.validatorPool.getStakeAmount())
      .div(4)
      .add(1)
      .toBigInt();
    // Set infringer and disputer validators
    let infringer = validators[0];
    let disputer = validators[1];
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    await showState("After registering", expectedState);

    await ethdkg.minorSlash(infringer, disputer);
    let currentState = await getCurrentState(fixture, validators);
    await showState("After minor slashing", currentState);

    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.StakeNFT++;

    // Expect infringer to loose the validator position
    expectedState.StakeNFT.MAD += stakeAmount - reward;
    expectedState.ValidatorNFT.MAD -= stakeAmount;

    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.validators[1].MAD += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);

    // burn the guy more 2 times
    for (let i = 0; i < 2; i++) {
      let expectedState = await getCurrentState(fixture, validators);
      await ethdkg.minorSlash(infringer, disputer);
      let currentState = await getCurrentState(fixture, validators);
      expectedState.StakeNFT.MAD -= reward;
      expectedState.validators[1].MAD += reward;
      expectedState.validators[0].Acc = true;
      expectedState.validators[0].ExQ = true;
      expectedState.validators[0].Reg = false;
      expect(currentState).to.be.deep.equal(
        expectedState,
        `After minor slashing: ${i}`
      );
    }
    let finalReward = stakeAmount - BigInt(3) * reward;
    // After last minor slash the guy should have any funds to generate a new stakeNFT position
    expectedState = await getCurrentState(fixture, validators);
    await ethdkg.minorSlash(infringer, disputer);
    currentState = await getCurrentState(fixture, validators);
    expectedState.StakeNFT.MAD -= finalReward;
    expectedState.ValidatorPool.StakeNFT--;
    expectedState.validators[1].MAD += finalReward;
    expectedState.validators[0].Acc = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Reg = false;
    expect(currentState).to.be.deep.equal(
      expectedState,
      `Failed on final minor slashing`
    );
    await expect(ethdkg.minorSlash(infringer, disputer)).to.be.revertedWith(
      "ValidatorPool: DishonestValidator should be a validator or be in the exiting line!"
    );
  });

  it("Minor slash a validator with reward equals to the staking amount", async function () {
    let reward = (await fixture.validatorPool.getStakeAmount()).toBigInt();
    // Set reward to be equal to the staked amount (in this case minor will be equal to major slashes)
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    await showState(
      "After registering",
      await getCurrentState(fixture, validators)
    );
    await ethdkg.minorSlash(validators[0], validators[1]);
    await showState(
      "After minor slashing",
      await getCurrentState(fixture, validators)
    );

    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;

    // Expect infringer to loose the validator position and the accusator to have all the funds gained
    // as reward
    expectedState.ValidatorNFT.MAD -= stakeAmount;
    expectedState.validators[1].MAD += reward;
    expectedState.validators[0].Acc = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Minor slash a validator with reward is greater than the staking amount", async function () {
    let reward = (await fixture.validatorPool.getStakeAmount()).toBigInt();
    reward *= BigInt(2);
    // Set reward to be equal to the staked amount (in this case minor will be equal to major slashes)
    await factoryCallAny(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    let ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAny(fixture, "validatorPool", "registerValidators", [
      validators,
      stakingTokenIds,
    ]);
    let expectedState = await getCurrentState(fixture, validators);
    await showState(
      "After registering",
      await getCurrentState(fixture, validators)
    );
    await ethdkg.minorSlash(validators[0], validators[1]);
    await showState(
      "After minor slashing",
      await getCurrentState(fixture, validators)
    );

    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister and not StakeNFT position to be created
    expectedState.ValidatorPool.ValNFT--;

    // Expect infringer to loose the validator position and the accusator to have all the funds gained
    // as reward
    expectedState.ValidatorNFT.MAD -= stakeAmount;
    expectedState.validators[1].MAD += stakeAmount;
    expectedState.validators[0].Acc = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });
});
