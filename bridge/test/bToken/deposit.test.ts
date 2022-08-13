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
import { getEthConsumedAsGas, getState, showState, state } from "./setup";

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

  it.only("Should not deposit with 0 eth amount", async () => {
    await expect(
      fixture.bToken.mintDeposit(1, user.address, 0, {
        value: 0,
      })
    )
      .to.be.revertedWithCustomError(fixture.bToken, "MinimumValueNotMet")
      .withArgs(0, await fixture.bToken.getMarketSpread());
  });

  it("Should not deposit with 0 deposit amount", async () => {
    await expect(
      fixture.bToken.deposit(1, user.address, 0)
    ).to.be.revertedWithCustomError(fixture.bToken, `DepositAmountZero`);
  });

  it("Should deposit funds burning tokens hence affecting pool balance", async () => {
    // Mint ATK since a burn will be performed
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    await expect(fixture.bToken.deposit(1, user.address, bTokenDeposit))
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokenDeposit);
    expectedState = await getState(fixture);
    const tx = await fixture.bToken.deposit(1, user.address, bTokenDeposit);
    expectedState.Balances.bToken.admin -= bTokenDeposit.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.bToken.poolBalance = (
      await fixture.bToken.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.bToken.totalSupply -= bTokenDeposit.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should deposit funds without burning tokens hence not affecting balances", async () => {
    await expect(
      fixture.bToken.virtualMintDeposit(1, user.address, bTokenDeposit)
    )
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokenDeposit);
    expectedState = await getState(fixture);
    const tx = await fixture.bToken.virtualMintDeposit(
      1,
      user.address,
      bTokenDeposit
    );
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should deposit funds minting tokens hence affecting balances", async () => {
    // Calculate the amount of bTokens per eth value sent
    const bTokens = await fixture.bToken.getLatestMintedBTokensFromEth(
      ethIn.div(marketSpread)
    );
    await expect(
      fixture.bToken.mintDeposit(1, user.address, 0, {
        value: ethIn,
      })
    )
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokens);
    expectedState = await getState(fixture);
    const tx = await fixture.bToken.mintDeposit(1, user.address, 0, {
      value: ethIn,
    });

    expectedState.Balances.bToken.poolBalance = (
      await fixture.bToken.getPoolBalance()
    ).toBigInt();

    expectedState.Balances.eth.admin -= ethIn.toBigInt();
    expectedState.Balances.eth.bToken += ethIn.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());

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
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    await expect(fixture.bToken.deposit(1, user.address, bTokenDeposit))
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokenDeposit);
    expectedState = await getState(fixture);
    const tx = await fixture.bToken.deposit(1, user.address, bTokenDeposit);
    const [, tx2] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    expectedState.Balances.bToken.admin -= bTokenDeposit.toBigInt();
    expectedState.Balances.bToken.poolBalance = (
      await fixture.bToken.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.bToken.totalSupply -= bTokenDeposit.toBigInt();
    expectedState.Balances.eth.bToken = (
      await fixture.bToken.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx2.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should distribute after mint deposit", async () => {
    const bTokens = await fixture.bToken.getLatestMintedBTokensFromEth(
      ethIn.div(marketSpread)
    );
    await expect(
      fixture.bToken.mintDeposit(1, user.address, 0, {
        value: ethIn,
      })
    )
      .to.emit(fixture.bToken, "DepositReceived")
      .withArgs(1 || 2, 1, user.address, bTokens);
    expectedState = await getState(fixture);
    const tx = await fixture.bToken.mintDeposit(1, user.address, 0, {
      value: ethIn,
    });
    const [distribution, tx2] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    const distributedAmount = distribution.minerAmount
      .add(distribution.stakingAmount)
      .add(distribution.lpStakingAmount)
      .add(distribution.foundationAmount);
    expectedState.Balances.eth.admin -= ethIn.toBigInt();
    expectedState.Balances.eth.bToken -= distributedAmount.div(2).toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx2.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });
});
