import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getEthConsumedAsGas, getState, showState, state } from "./setup";

describe("Testing ALCB Deposit methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const minALCBs = 0;
  const eth = 10;
  const alcbs = 10;
  let ethIn: BigNumber;
  let alcbDeposit: BigNumber;

  beforeEach(async function () {
    fixture = await loadFixture(getFixture);
    [admin, user] = await ethers.getSigners();
    showState("Initial", await getState(fixture));
    ethIn = ethers.utils.parseEther(eth.toString());
    alcbDeposit = ethers.utils.parseUnits(alcbs.toString(10));
  });

  it("Only factory should be able to set a new type", async () => {
    await expect(
      fixture.alcb.setAccountType(3, true)
    ).to.be.revertedWithCustomError(fixture.alcb, `OnlyFactory`);
    const recpt = await (
      await fixture.factory
        .connect(admin)
        .callAny(
          fixture.alcb.address,
          0,
          fixture.alcb.interface.encodeFunctionData("setAccountType", [3, true])
        )
    ).wait();
    expect(recpt.status).to.be.equals(1);
  });

  it("Should fail querying for an invalid deposit ID", async () => {
    const depositId = 1000;
    await expect(fixture.alcb.getDeposit(depositId))
      .to.be.revertedWithCustomError(fixture.alcb, `InvalidDepositId`)
      .withArgs(depositId);
  });

  it("Should not deposit to a contract", async () => {
    await expect(fixture.alcb.deposit(1, fixture.alcb.address, 0))
      .to.be.revertedWithCustomError(
        fixture.alcb,
        `ContractsDisallowedDeposits`
      )
      .withArgs(fixture.alcb.address);
  });

  it("Should not deposit to an invalid account type", async () => {
    await fixture.alcb.mint(0, {
      value: ethIn,
    });
    await expect(fixture.alcb.deposit(3, user.address, 100))
      .to.be.revertedWithCustomError(fixture.alcb, "AccountTypeNotSupported")
      .withArgs(3);
  });

  it("Should not deposit to a disabled account type", async () => {
    await fixture.alcb.mint(0, {
      value: ethIn,
    });
    await (
      await fixture.factory
        .connect(admin)
        .callAny(
          fixture.alcb.address,
          0,
          fixture.alcb.interface.encodeFunctionData("setAccountType", [
            1,
            false,
          ])
        )
    ).wait();
    await expect(fixture.alcb.deposit(1, user.address, 100))
      .to.be.revertedWithCustomError(fixture.alcb, "AccountTypeNotSupported")
      .withArgs(1);
  });

  it("Should not mintDeposit to an invalid account type", async () => {
    await expect(
      fixture.alcb.mintDeposit(3, user.address, 0, {
        value: 100,
      })
    )
      .to.be.revertedWithCustomError(fixture.alcb, "AccountTypeNotSupported")
      .withArgs(3);
  });

  it("Should not virtualMintDeposit to an incorrect account type", async () => {
    await expect(
      fixture.factory
        .connect(admin)
        .callAny(
          fixture.alcb.address,
          0,
          fixture.alcb.interface.encodeFunctionData("virtualMintDeposit", [
            3,
            user.address,
            alcbDeposit,
          ])
        )
    )
      .to.be.revertedWithCustomError(fixture.alcb, "AccountTypeNotSupported")
      .withArgs(3);
  });

  it("Should not mintDeposit with 0 eth amount", async () => {
    await expect(
      fixture.alcb.mintDeposit(1, user.address, 0, {
        value: 0,
      })
    )
      .to.be.revertedWithCustomError(fixture.alcb, "MinimumValueNotMet")
      .withArgs(0, await fixture.alcb.getMarketSpread());
  });

  it("Should not deposit with 0 deposit amount", async () => {
    await expect(
      fixture.alcb.deposit(1, user.address, 0)
    ).to.be.revertedWithCustomError(fixture.alcb, `DepositAmountZero`);
  });

  it("Should deposit funds burning tokens hence affecting pool balance", async () => {
    // Mint ATK since a burn will be performed
    await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mint",
      admin,
      [minALCBs],
      ethIn
    );
    await expect(fixture.alcb.deposit(1, user.address, alcbDeposit))
      .to.emit(fixture.alcb, "DepositReceived")
      .withArgs(2, 1, user.address, alcbDeposit);
    expectedState = await getState(fixture);
    const tx = await fixture.alcb.deposit(1, user.address, alcbDeposit);
    expectedState.Balances.alcb.admin -= alcbDeposit.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.alcb.poolBalance = (
      await fixture.alcb.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.alcb.totalSupply -= alcbDeposit.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should be able to deposit with new added type", async () => {
    await fixture.factory
      .connect(admin)
      .callAny(
        fixture.alcb.address,
        0,
        fixture.alcb.interface.encodeFunctionData("setAccountType", [10, true])
      );
    await fixture.alcb.mintDeposit(10, user.address, 0, {
      value: 100,
    });
  });

  it("Should deposit funds without burning tokens hence not affecting balances", async () => {
    await expect(
      fixture.factory
        .connect(admin)
        .callAny(
          fixture.alcb.address,
          0,
          fixture.alcb.interface.encodeFunctionData("virtualMintDeposit", [
            1,
            user.address,
            alcbDeposit,
          ])
        )
    )
      .to.emit(fixture.alcb, "DepositReceived")
      .withArgs(2, 1, user.address, alcbDeposit);
    expectedState = await getState(fixture);
    const tx = await fixture.factory
      .connect(admin)
      .callAny(
        fixture.alcb.address,
        0,
        fixture.alcb.interface.encodeFunctionData("virtualMintDeposit", [
          1,
          user.address,
          alcbDeposit,
        ])
      );
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should deposit funds minting tokens hence affecting balances", async () => {
    // Calculate the amount of alcbs per eth value sent
    const alcbs = await fixture.alcb.getLatestMintedTokensFromEth(ethIn);
    await expect(
      fixture.alcb.mintDeposit(1, user.address, 0, {
        value: ethIn,
      })
    )
      .to.emit(fixture.alcb, "DepositReceived")
      .withArgs(2, 1, user.address, alcbs);
    expectedState = await getState(fixture);
    const tx = await fixture.alcb.mintDeposit(1, user.address, 0, {
      value: ethIn,
    });

    expectedState.Balances.alcb.poolBalance = (
      await fixture.alcb.getPoolBalance()
    ).toBigInt();

    expectedState.Balances.eth.admin -= ethIn.toBigInt();
    expectedState.Balances.eth.alcb += ethIn.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());

    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should get deposit amount by Id", async () => {
    expectedState = await getState(fixture);
    await fixture.factory
      .connect(admin)
      .callAny(
        fixture.alcb.address,
        0,
        fixture.alcb.interface.encodeFunctionData("virtualMintDeposit", [
          1,
          user.address,
          alcbDeposit,
        ])
      );
    const depositId = 2;
    const deposit = await fixture.alcb.getDeposit(depositId);
    expect(deposit.value).to.be.equal(ethIn.toBigInt());
  });

  it("Should distribute after deposit", async () => {
    await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mint",
      admin,
      [minALCBs],
      ethIn
    );
    await expect(fixture.alcb.deposit(1, user.address, alcbDeposit))
      .to.emit(fixture.alcb, "DepositReceived")
      .withArgs(2, 1, user.address, alcbDeposit);
    expectedState = await getState(fixture);
    const tx = await fixture.alcb.deposit(1, user.address, alcbDeposit);
    const [, tx2] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "distribute",
      admin,
      []
    );
    expectedState.Balances.alcb.admin -= alcbDeposit.toBigInt();
    expectedState.Balances.alcb.poolBalance = (
      await fixture.alcb.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.alcb.totalSupply -= alcbDeposit.toBigInt();
    expectedState.Balances.eth.alcb = (
      await fixture.alcb.getPoolBalance()
    ).toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx2.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should distribute after mint deposit", async () => {
    expectedState = await getState(fixture);
    const alcbs = await fixture.alcb.getLatestMintedTokensFromEth(ethIn);
    let tx0;
    await expect(
      (tx0 = fixture.alcb.mintDeposit(1, user.address, 0, {
        value: ethIn,
      }))
    )
      .to.emit(fixture.alcb, "DepositReceived")
      .withArgs(2, 1, user.address, alcbs);
    const tx = await fixture.alcb.mintDeposit(1, user.address, 0, {
      value: ethIn,
    });
    let distributedAmount = await fixture.alcb.getYield();
    const [wasSuccessful, tx2] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "distribute",
      admin,
      []
    );
    expect(wasSuccessful).to.be.equal(true);
    distributedAmount = await fixture.alcb.getYield();
    expect(distributedAmount).to.be.equal(0);
    expectedState.Balances.eth.admin -= ethIn.toBigInt();
    expectedState.Balances.eth.admin -= ethIn.toBigInt();
    expectedState.Balances.eth.alcb -= distributedAmount.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(
      await (await tx0).wait()
    );
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx2.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });
});
