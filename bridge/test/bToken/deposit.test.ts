import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  callFunctionAndGetReturnValues,
  factoryCallAnyFixture,
  Fixture,
  getFixture,
} from "../setup";
import { getState, showState, state } from "./setup";

describe("Testing BToken Deposit methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const minBTokens = 0;
  const marketSpread = 4;
  const eth = 10;
  const bTokens = 10;
  let ethIn: BigNumber;
  let bTokenDeposit: BigNumber;

  beforeEach(async function () {
    fixture = await getFixture();
    [admin, user] = await ethers.getSigners();
    showState("Initial", await getState(fixture));
    await factoryCallAnyFixture(fixture, "bToken", "setAdmin", [admin.address]);
    ethIn = ethers.utils.parseEther(eth.toString());
    bTokenDeposit = ethers.utils.parseUnits(bTokens.toString());
  });

  it("Should fail querying for an invalid deposit ID", async () => {
    const depositId = 1000;
    await expect(fixture.bToken.getDeposit(depositId))
      .to.be.revertedWithCustomError(fixture.bToken, `InvalidDepositId`)
      .withArgs(depositId);
  });

  it("Should not deposit to a contract", async () => {
    await expect(fixture.bToken.deposit(1, fixture.bToken.address, 0))
      .to.be.revertedWithCustomError(
        fixture.bToken,
        `ContractsDisallowedDeposits`
      )
      .withArgs(fixture.bToken.address);
  });

  it("Should not deposit with 0 eth amount", async () => {
    await expect(
      fixture.bToken.mintDeposit(1, user.address, 0, {
        value: 0,
      })
    )
      .to.be.revertedWithCustomError(fixture.bToken, `MarketSpreadTooLow`)
      .withArgs(0);
  });

  it("Should not deposit with 0 deposit amount", async () => {
    await expect(
      fixture.bToken.deposit(1, user.address, 0)
    ).to.be.revertedWithCustomError(fixture.bToken, `DepositAmountZero`);
  });

  it("Should deposit funds on side-chain burning main-chain tokens then affecting pool balance", async () => {
    // Mint ATK since a burn will be performed
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    expectedState = await getState(fixture);
    await expect(fixture.bToken.deposit(1, user.address, bTokenDeposit))
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokenDeposit);
    expectedState.Balances.bToken.admin -= bTokenDeposit.toBigInt();
    expectedState.Balances.bToken.poolBalance = (
      await fixture.bToken.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.bToken.totalSupply -= bTokenDeposit.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should deposit funds on side-chain without burning main-chain tokens then not affecting balances", async () => {
    expectedState = await getState(fixture);
    await expect(
      fixture.bToken.virtualMintDeposit(1, user.address, bTokenDeposit)
    )
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokenDeposit);
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should deposit funds on side-chain without burning main-chain tokens then not affecting balances", async () => {
    expectedState = await getState(fixture);
    // Calculate the amount of bytes per eth value sent
    const bTokens = await fixture.bToken.ethToBTokens(
      await fixture.bToken.getPoolBalance(),
      ethIn.div(marketSpread)
    );
    await expect(
      fixture.bToken.mintDeposit(1, user.address, 0, {
        value: ethIn,
      })
    )
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokens);
    expectedState.Balances.bToken.poolBalance = (
      await fixture.bToken.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.eth.admin -= eth;
    expectedState.Balances.eth.bToken += ethIn.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should get deposit amount by Id", async () => {
    expectedState = await getState(fixture);
    const [depositId] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "virtualMintDeposit",
      admin,
      [1, user.address, bTokenDeposit]
    );
    const deposit = await fixture.bToken.getDeposit(depositId);
    expect(deposit.value).to.be.equal(ethIn.toBigInt());
  });

  it("Should distribute after deposit", async () => {
    // Mint ATK since a burn will be performed
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    expectedState = await getState(fixture);
    await expect(fixture.bToken.deposit(1, user.address, bTokenDeposit))
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokenDeposit);
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    const distributedAmount = distribution.minerAmount
      .add(distribution.stakingAmount)
      .add(distribution.lpStakingAmount)
      .add(distribution.foundationAmount);
    expectedState.Balances.bToken.admin -= bTokenDeposit.toBigInt();
    expectedState.Balances.bToken.poolBalance = (
      await fixture.bToken.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.bToken.totalSupply -= bTokenDeposit.toBigInt();
    expectedState.Balances.eth.bToken = bTokenDeposit
      .sub(distributedAmount)
      .toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should distribute after mint deposit", async () => {
    const bTokens = await fixture.bToken.ethToBTokens(
      await fixture.bToken.getPoolBalance(),
      ethIn.div(marketSpread)
    );
    expectedState = await getState(fixture);
    await expect(
      fixture.bToken.mintDeposit(1, user.address, 0, {
        value: ethIn,
      })
    )
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokens);
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    const distributedAmount = distribution.minerAmount
      .add(distribution.stakingAmount)
      .add(distribution.lpStakingAmount)
      .add(distribution.foundationAmount);
    expectedState.Balances.eth.admin -= eth;
    expectedState.Balances.bToken.poolBalance = bTokenDeposit
      .sub(distributedAmount)
      .toBigInt();
    expectedState.Balances.eth.bToken = bTokenDeposit
      .sub(distributedAmount)
      .toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });
});
