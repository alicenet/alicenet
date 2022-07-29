import { BigNumber, ContractTransaction, Signer } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
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
  commitSnapshots,
  createValidators,
  getCurrentState,
  getPublicStakingFromMinorSlashEvent,
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
    const reward = ethers.utils.parseEther("1").toBigInt();
    // Set reward to 1 AToken
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    const expectedState = await getCurrentState(fixture, validators);
    await showState(
      "After registering",
      await getCurrentState(fixture, validators)
    );
    await ethdkg.minorSlash(validators[0], validators[1]);
    await showState(
      "After minor slashing",
      await getCurrentState(fixture, validators)
    );
    const currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.PublicStaking++;

    // Expect infringer to loose the validator position
    expectedState.PublicStaking.ATK += stakeAmount;
    expectedState.ValidatorStaking.ATK -= stakeAmount;
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.PublicStaking.ATK -= reward;
    expectedState.validators[1].ATK += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });

  it("Minor slash a validator then he leaves the pool", async function () {
    const reward = ethers.utils.parseEther("1").toBigInt();
    // Set reward to 1 AToken
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    const expectedState = await getCurrentState(fixture, validators);
    await showState(
      "After registering",
      await getCurrentState(fixture, validators)
    );
    const tx = await ethdkg.minorSlash(validators[0], validators[1]);
    const newPublicStaking = await getPublicStakingFromMinorSlashEvent(tx);
    expect(newPublicStaking).to.be.gt(
      BigInt(0),
      "New PublicStaking position was not created properly!"
    );
    await showState(
      "After minor slashing",
      await getCurrentState(fixture, validators)
    );
    const currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.PublicStaking++;

    // Expect infringer to loose the validator position
    expectedState.PublicStaking.ATK += stakeAmount;
    expectedState.ValidatorStaking.ATK -= stakeAmount;
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.PublicStaking.ATK -= reward;
    expectedState.validators[1].ATK += reward;
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

    const claimTx = (await fixture.validatorPool
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .claimExitingNFTPosition()) as ContractTransaction;
    const receipt = await ethers.provider.getTransactionReceipt(claimTx.hash);
    const exitingNFT = BigNumber.from(receipt.logs[0].topics[3]);
    const blockNumber = receipt.blockNumber;

    expect(exitingNFT.toBigInt()).to.be.equal(
      newPublicStaking,
      "Failed claiming NFT position!"
    );

    expect(
      (await fixture.publicStaking.ownerOf(newPublicStaking)).toLowerCase()
    ).to.be.equal(validatorsSnapshots[0].address.toLowerCase());

    const position = await fixture.publicStaking.getPosition(newPublicStaking);
    expect(position.freeAfter.toBigInt()).to.be.equal(
      (await fixture.validatorPool.POSITION_LOCK_PERIOD()).toBigInt() +
        BigInt(blockNumber)
    );
  });

  it("Minor slash a validator then major slash it", async function () {
    const reward = ethers.utils.parseEther("10000").toBigInt();
    // Set reward to 1 AToken
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    let expectedState = await getCurrentState(fixture, validators);
    const tx = await ethdkg.minorSlash(validators[0], validators[1]);
    const newPublicStaking = await getPublicStakingFromMinorSlashEvent(tx);
    expect(newPublicStaking).to.be.gt(
      BigInt(0),
      "New PublicStaking position was not created properly!"
    );
    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.PublicStaking++;

    // Expect infringer to loose the validator position
    expectedState.PublicStaking.ATK += stakeAmount - reward;
    expectedState.ValidatorStaking.ATK -= stakeAmount;
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.validators[1].ATK += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed assert after minor slash"
    );
    await mineBlocks(1n);
    expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[2]);
    currentState = await getCurrentState(fixture, validators);
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.PublicStaking--;
    // Expect reward to be transferred from ValidatorStaking to disputer
    expectedState.PublicStaking.ATK -= stakeAmount - reward;
    // the stakeamount minus the 2 rewards given to the 2 accusators should be redistributed to all
    // validators
    expectedState.ValidatorStaking.ATK += stakeAmount - BigInt(2) * reward;
    expectedState.validators[2].ATK += reward;
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
    // Mint a publicStaking and burn it to the ValidatorPool contract. Besides a contract self destructing
    // itself, this is a method to send eth accidentally to the validatorPool contract
    const etherAmount = ethers.utils.parseEther("1");
    const aTokenAmount = ethers.utils.parseEther("2");
    await burnStakeTo(fixture, etherAmount, aTokenAmount, adminSigner);

    const reward = ethers.utils.parseEther("10000").toBigInt();
    // Set reward to 1 AToken
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    let expectedState = await getCurrentState(fixture, validators);
    const tx = await ethdkg.minorSlash(validators[0], validators[1]);
    const newPublicStaking = await getPublicStakingFromMinorSlashEvent(tx);
    expect(newPublicStaking).to.be.gt(
      BigInt(0),
      "New PublicStaking position was not created properly!"
    );
    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.PublicStaking++;

    // Expect infringer to loose the validator position
    expectedState.PublicStaking.ATK += stakeAmount - reward;
    expectedState.ValidatorStaking.ATK -= stakeAmount;
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.validators[1].ATK += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed assert after minor slash"
    );
    await mineBlocks(1n);
    expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[2]);
    currentState = await getCurrentState(fixture, validators);
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.PublicStaking--;
    // Expect reward to be transferred from ValidatorStaking to disputer
    expectedState.PublicStaking.ATK -= stakeAmount - reward;
    // the stakeamount minus the 2 rewards given to the 2 accusators should be redistributed to all
    // validators
    expectedState.ValidatorStaking.ATK += stakeAmount - BigInt(2) * reward;
    expectedState.validators[2].ATK += reward;
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
    const reward = ethers.utils.parseEther("10000").toBigInt();
    // Set reward to 1 AToken
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);

    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    const eths = 4;
    const atokens = 20;
    await fixture.validatorStaking.connect(adminSigner).depositEth(42, {
      value: ethers.utils.parseEther(`${eths}`),
    });
    await fixture.aToken
      .connect(adminSigner)
      .approve(
        fixture.validatorStaking.address,
        ethers.utils.parseEther(`${atokens}`)
      );
    await fixture.validatorStaking
      .connect(adminSigner)
      .depositToken(42, ethers.utils.parseEther(`${atokens}`));
    let expectedState = await getCurrentState(fixture, validators);
    const tx = await ethdkg.minorSlash(validators[0], validators[1]);
    const newPublicStaking = await getPublicStakingFromMinorSlashEvent(tx);
    expect(newPublicStaking).to.be.gt(
      BigInt(0),
      "New PublicStaking position was not created properly!"
    );
    let currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.PublicStaking++;

    // Expect infringer to loose the validator position
    expectedState.PublicStaking.ATK += stakeAmount - reward;
    expectedState.ValidatorStaking.ATK -=
      stakeAmount + ethers.utils.parseEther(`${atokens / 4}`).toBigInt();
    expectedState.ValidatorStaking.ETH -= ethers.utils
      .parseEther(`${eths / 4}`)
      .toBigInt();
    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.validators[1].ATK +=
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

    // minting 1 PublicStaking so all the profit funds are not moved from the PublicStaking contract at slashing
    // time
    const newStakeAmount = (stakeAmount - reward) * BigInt(3);
    await fixture.aToken
      .connect(adminSigner)
      .approve(fixture.publicStaking.address, newStakeAmount);
    await fixture.publicStaking.connect(adminSigner).mint(newStakeAmount);

    // Deposit now in the publicStaking contract
    await fixture.publicStaking.connect(adminSigner).depositEth(42, {
      value: ethers.utils.parseEther(`${eths}`),
    });
    await fixture.aToken
      .connect(adminSigner)
      .approve(
        fixture.publicStaking.address,
        ethers.utils.parseEther(`${atokens}`)
      );
    await fixture.publicStaking
      .connect(adminSigner)
      .depositToken(42, ethers.utils.parseEther(`${atokens}`));

    expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[2]);
    currentState = await getCurrentState(fixture, validators);
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.PublicStaking--;
    expectedState.PublicStaking.ETH -= ethers.utils
      .parseEther(`${eths / 4}`)
      .toBigInt();
    // Expect reward to be transferred from ValidatorStaking to disputer
    expectedState.PublicStaking.ATK -=
      stakeAmount -
      reward +
      ethers.utils.parseEther(`${atokens / 4}`).toBigInt();
    // the stakeamount minus the 2 rewards given to the 2 accusators should be redistributed to all
    // validators
    expectedState.ValidatorStaking.ATK += stakeAmount - BigInt(2) * reward;
    expectedState.validators[2].ATK +=
      reward + ethers.utils.parseEther(`${atokens / 4}`).toBigInt();
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
    const collectedAmount =
      (stakeAmount - BigInt(2) * reward) /
        BigInt(validatorsSnapshots.length - 1) +
      ethers.utils.parseEther(`${atokens / 4}`).toBigInt();
    for (let index = 1; index < validatorsSnapshots.length; index++) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshots[index]))
        .collectProfits();
      expectedState.validators[index].ATK += collectedAmount;
      expectedState.ValidatorStaking.ETH -= ethers.utils
        .parseEther(`${eths / 4}`)
        .toBigInt();
      expectedState.ValidatorStaking.ATK -= collectedAmount;
    }
    currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed in the assertion after collect Profits"
    );
  });

  it("Major slash a validator then validators collect profit", async function () {
    const reward = ethers.utils.parseEther("5000").toBigInt();
    // Set reward to 1 AToken
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    let expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[1]);
    await showState(
      "After major slashing",
      await getCurrentState(fixture, validators)
    );
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.ValNFT--;
    // Expect reward to be transferred from ValidatorStaking to disputer
    expectedState.ValidatorStaking.ATK -= reward;
    expectedState.validators[1].ATK += reward;
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
    const collectedAmount =
      (stakeAmount - reward) / BigInt(validatorsSnapshots.length - 1);
    for (let index = 1; index < validatorsSnapshots.length; index++) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshots[index]))
        .collectProfits();
      expectedState.validators[index].ATK += collectedAmount;
      expectedState.ValidatorStaking.ATK -= collectedAmount;
    }
    currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed in the assertion after collect Profits"
    );
  });

  it("Should not allow major slash a validator twice", async function () {
    const reward = ethers.utils.parseEther("5000").toBigInt();
    // Set reward to 1 AToken
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    const expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[1]);
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.ValNFT--;
    // Expect reward to be transferred from ValidatorStaking to disputer
    expectedState.ValidatorStaking.ATK -= reward;
    expectedState.validators[1].ATK += reward;
    // Expect infringer to be unregistered, not in exiting queue and not accusable
    expectedState.validators[0].Reg = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Acc = false;
    const currentState = await getCurrentState(fixture, validators);
    expect(currentState).to.be.deep.equal(
      expectedState,
      "Failed checking state after major slashing!"
    );
    await expect(ethdkg.majorSlash(validators[0], validators[1]))
      .to.be.revertedWithCustomError(
        fixture.validatorPool,
        "DishonestValidatorNotAccusable"
      )
      .withArgs(ethers.utils.getAddress(validators[0]));
  });

  it("Should not allow major/minor slash a person that it's not a validator", async function () {
    const reward = ethers.utils.parseEther("5000").toBigInt();
    // Set reward to 1 AToken
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await expect(
      ethdkg.majorSlash(await adminSigner.getAddress(), validators[1])
    )
      .to.be.revertedWithCustomError(
        fixture.validatorPool,
        "DishonestValidatorNotAccusable"
      )
      .withArgs(await adminSigner.getAddress());

    await expect(
      ethdkg.minorSlash(await adminSigner.getAddress(), validators[1])
    )
      .to.be.revertedWithCustomError(
        fixture.validatorPool,
        "DishonestValidatorNotAccusable"
      )
      .withArgs(await adminSigner.getAddress());
  });

  it("Major slash a validator with disputer reward greater than stake Amount", async function () {
    let reward = (await fixture.validatorPool.getStakeAmount()).toBigInt();
    reward *= BigInt(2);
    // Set reward to 1 AToken
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    let expectedState = await getCurrentState(fixture, validators);
    await ethdkg.majorSlash(validators[0], validators[1]);
    await showState(
      "After major slashing",
      await getCurrentState(fixture, validators)
    );
    // Expect infringer unregister the validator position
    expectedState.ValidatorPool.ValNFT--;
    // Expect reward to be transferred from ValidatorStaking to disputer
    expectedState.ValidatorStaking.ATK -= stakeAmount;
    expectedState.validators[1].ATK += stakeAmount;
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
    const collectedAmount = BigInt(0);
    for (let index = 1; index < validatorsSnapshots.length; index++) {
      await fixture.validatorPool
        .connect(await getValidatorEthAccount(validatorsSnapshots[index]))
        .collectProfits();
      expectedState.validators[index].ATK += collectedAmount;
      expectedState.ValidatorStaking.ATK -= collectedAmount;
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
    // Set reward to 1 AToken
    const reward = (await fixture.validatorPool.getStakeAmount())
      .div(4)
      .add(1)
      .toBigInt();
    // Set infringer and disputer validators
    const infringer = validators[0];
    const disputer = validators[1];
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    let expectedState = await getCurrentState(fixture, validators);
    await showState("After registering", expectedState);

    await ethdkg.minorSlash(infringer, disputer);
    let currentState = await getCurrentState(fixture, validators);
    await showState("After minor slashing", currentState);
    await mineBlocks(1n);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;
    expectedState.ValidatorPool.PublicStaking++;

    // Expect infringer to loose the validator position
    expectedState.PublicStaking.ATK += stakeAmount - reward;
    expectedState.ValidatorStaking.ATK -= stakeAmount;

    // Expect infringer to loose reward on his staking position and be transferred to accusator
    expectedState.validators[1].ATK += reward;
    expectedState.validators[0].Acc = true;
    expectedState.validators[0].ExQ = true;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);

    // burn the guy more 2 times
    for (let i = 0; i < 2; i++) {
      const expectedState = await getCurrentState(fixture, validators);
      await ethdkg.minorSlash(infringer, disputer);
      const currentState = await getCurrentState(fixture, validators);
      expectedState.PublicStaking.ATK -= reward;
      expectedState.validators[1].ATK += reward;
      expectedState.validators[0].Acc = true;
      expectedState.validators[0].ExQ = true;
      expectedState.validators[0].Reg = false;
      expect(currentState).to.be.deep.equal(
        expectedState,
        `After minor slashing: ${i}`
      );
      await mineBlocks(1n);
    }
    const finalReward = stakeAmount - BigInt(3) * reward;
    // After last minor slash the guy should have any funds to generate a new publicStaking position
    expectedState = await getCurrentState(fixture, validators);
    await ethdkg.minorSlash(infringer, disputer);
    currentState = await getCurrentState(fixture, validators);
    expectedState.PublicStaking.ATK -= finalReward;
    expectedState.ValidatorPool.PublicStaking--;
    expectedState.validators[1].ATK += finalReward;
    expectedState.validators[0].Acc = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Reg = false;
    expect(currentState).to.be.deep.equal(
      expectedState,
      `Failed on final minor slashing`
    );
    await expect(ethdkg.minorSlash(infringer, disputer))
      .to.be.revertedWithCustomError(
        fixture.validatorPool,
        "DishonestValidatorNotAccusable"
      )
      .withArgs(ethers.utils.getAddress(infringer));
  });

  it("Minor slash a validator with reward equals to the staking amount", async function () {
    const reward = (await fixture.validatorPool.getStakeAmount()).toBigInt();
    // Set reward to be equal to the staked amount (in this case minor will be equal to major slashes)
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    const expectedState = await getCurrentState(fixture, validators);
    await showState(
      "After registering",
      await getCurrentState(fixture, validators)
    );
    await ethdkg.minorSlash(validators[0], validators[1]);
    await showState(
      "After minor slashing",
      await getCurrentState(fixture, validators)
    );

    const currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister
    expectedState.ValidatorPool.ValNFT--;

    // Expect infringer to loose the validator position and the accusator to have all the funds gained
    // as reward
    expectedState.ValidatorStaking.ATK -= stakeAmount;
    expectedState.validators[1].ATK += reward;
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
    await factoryCallAnyFixture(fixture, "validatorPool", "setDisputerReward", [
      reward,
    ]);
    // Obtain ETHDKG Mock
    const ETHDKGMockFactory = await ethers.getContractFactory("ETHDKGMock");
    const ethdkg = ETHDKGMockFactory.attach(fixture.ethdkg.address);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await mineBlocks(1n);
    const expectedState = await getCurrentState(fixture, validators);
    await showState(
      "After registering",
      await getCurrentState(fixture, validators)
    );
    await ethdkg.minorSlash(validators[0], validators[1]);
    await showState(
      "After minor slashing",
      await getCurrentState(fixture, validators)
    );

    const currentState = await getCurrentState(fixture, validators);
    // Expect infringer validator position to be unregister and not PublicStaking position to be created
    expectedState.ValidatorPool.ValNFT--;

    // Expect infringer to loose the validator position and the accusator to have all the funds gained
    // as reward
    expectedState.ValidatorStaking.ATK -= stakeAmount;
    expectedState.validators[1].ATK += stakeAmount;
    expectedState.validators[0].Acc = false;
    expectedState.validators[0].ExQ = false;
    expectedState.validators[0].Reg = false;
    await showState("Expected state", expectedState);
    await showState("Current state", currentState);
    expect(currentState).to.be.deep.equal(expectedState);
  });
});
