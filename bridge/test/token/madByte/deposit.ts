import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { MadByte } from "../../../typechain-types";
import { expect } from "../../chai-setup";
import {
  callFunctionAndGetReturnValues,
  factoryCallAny,
  Fixture,
  getFixture,
} from "../../setup";
import { getState, init, showState, state } from "./setup";

describe("Testing MadByte Deposit methods", async () => {
  let madByte: MadByte;
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  let minMadBytes = 0;
  let marketSpread = 4;
  let eth = 10;
  let mad = 10;
  let ethIn: BigNumber;
  let madDeposit: BigNumber;

  beforeEach(async function () {
    fixture = await getFixture();
    let signers = await ethers.getSigners();
    [admin, user, user2] = signers;
    await init(fixture);
    // let expectedState = await getState(contractAddresses, userAddresses);
    showState("Initial", await getState(fixture));
    await factoryCallAny(fixture, "madByte", "setAdmin", [admin.address]);
    ethIn = ethers.utils.parseEther(eth.toString());
    madDeposit = ethers.utils.parseUnits(mad.toString());
  });

  it("Should fail querying for an invalid deposit ID", async () => {
    await expect(fixture.madByte.getDeposit(1000)).to.be.revertedWith(
      "MadByte: Invalid deposit ID!"
    );
  });

  it("Should not deposit to a contract", async () => {
    await expect(
      fixture.madByte.deposit(1, fixture.madByte.address, 0)
    ).to.be.revertedWith("MadByte: Contracts cannot make MadBytes deposits!");
  });

  it("Should not deposit with 0 eth amount", async () => {
    await expect(
      fixture.madByte.mintDeposit(1, user.address, 0, {
        value: 0,
      })
    ).to.be.revertedWith("MadByte: requires at least 4 WEI");
  });

  it("Should not deposit with 0 deposit amount", async () => {
    await expect(
      fixture.madByte.deposit(1, user.address, 0)
    ).to.be.revertedWith(
      "MadByte: The deposit amount must be greater than zero!"
    );
  });

  it("Should deposit funds on side-chain burning main-chain tokens then affecting pool balance", async () => {
    // Mint MAD since a burn will be performed
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      admin,
      [minMadBytes],
      ethIn
    );
    expectedState = await getState(fixture);
    await expect(fixture.madByte.deposit(1, user.address, madDeposit))
      .to.emit(fixture.madByte, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, madDeposit);
    expectedState.Balances.madByte.admin -= madDeposit.toBigInt();
    expectedState.Balances.madByte.poolBalance = (
      await fixture.madByte.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.madByte.totalSupply -= madDeposit.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should deposit funds on side-chain without burning main-chain tokens then not affecting balances", async () => {
    expectedState = await getState(fixture);
    await expect(
      fixture.madByte.virtualMintDeposit(1, user.address, madDeposit)
    )
      .to.emit(fixture.madByte, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, madDeposit);
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should deposit funds on side-chain without burning main-chain tokens then not affecting balances", async () => {
    expectedState = await getState(fixture);
    // Calculate the amount of bytes per eth value sent
    let madBytes = await fixture.madByte.ethToMadByte(
      await fixture.madByte.getPoolBalance(),
      ethIn.div(marketSpread)
    );
    await expect(
      fixture.madByte.mintDeposit(1, user.address, 0, {
        value: ethIn,
      })
    )
      .to.emit(fixture.madByte, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, madBytes);
    expectedState.Balances.madByte.poolBalance = (
      await fixture.madByte.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.eth.admin -= eth;
    expectedState.Balances.eth.madByte += ethIn.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should get deposit amount by Id", async () => {
    expectedState = await getState(fixture);
    const [depositId] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "virtualMintDeposit",
      admin,
      [1, user.address, madDeposit]
    );
    let deposit = await fixture.madByte.getDeposit(depositId);
    expect(deposit.value).to.be.equal(ethIn.toBigInt());
  });

  it("Should distribute after deposit", async () => {
    // Mint MAD since a burn will be performed
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      admin,
      [minMadBytes],
      ethIn
    );
    expectedState = await getState(fixture);
    await expect(fixture.madByte.deposit(1, user.address, madDeposit))
      .to.emit(fixture.madByte, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, madDeposit);
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "distribute",
      admin,
      []
    );
    const distributedAmount = distribution.minerAmount
      .add(distribution.stakingAmount)
      .add(distribution.lpStakingAmount)
      .add(distribution.foundationAmount);
    expectedState.Balances.madByte.admin -= madDeposit.toBigInt();
    expectedState.Balances.madByte.poolBalance = (
      await fixture.madByte.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.madByte.totalSupply -= madDeposit.toBigInt();
    expectedState.Balances.eth.madByte = madDeposit
      .sub(distributedAmount)
      .toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should distribute after mint deposit", async () => {
    let madBytes = await fixture.madByte.ethToMadByte(
      await fixture.madByte.getPoolBalance(),
      ethIn.div(marketSpread)
    );
    expectedState = await getState(fixture);
    await expect(
      fixture.madByte.mintDeposit(1, user.address, 0, {
        value: ethIn,
      })
    )
      .to.emit(fixture.madByte, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, madBytes);
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "distribute",
      admin,
      []
    );
    const distributedAmount = distribution.minerAmount
      .add(distribution.stakingAmount)
      .add(distribution.lpStakingAmount)
      .add(distribution.foundationAmount);
    expectedState.Balances.eth.admin -= eth;
    expectedState.Balances.madByte.poolBalance = madDeposit
      .sub(distributedAmount)
      .toBigInt();
    expectedState.Balances.eth.madByte = madDeposit
      .sub(distributedAmount)
      .toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });
});
