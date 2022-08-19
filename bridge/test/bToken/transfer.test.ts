import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getEthConsumedAsGas, getState, showState, state } from "./setup";

describe("Testing BToken Transfer methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 4;
  let ethIn: BigNumber;
  const minBTokens = 0;
  let bTokens: BigNumber;

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [admin, user, user2] = signers;
    showState("Initial", await getState(fixture));
    ethIn = ethers.utils.parseEther(eth.toString());
    // Mint some ATokens for testing
    [bTokens] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      user,
      [minBTokens],
      ethIn
    );
    expectedState = await getState(fixture);
  });

  it("Should transfer from sender to specified address", async () => {
    const tx = await fixture.bToken
      .connect(user)
      .transfer(admin.address, bTokens);
    expectedState.Balances.bToken.admin += bTokens.toBigInt();
    expectedState.Balances.bToken.user -= bTokens.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should transfer from specified address to specified address", async () => {
    const tx = await fixture.bToken
      .connect(user)
      .approve(admin.address, bTokens);
    const tx2 = await fixture.bToken
      .connect(admin)
      .transferFrom(user.address, user2.address, bTokens);
    expectedState.Balances.bToken.user -= bTokens.toBigInt();
    expectedState.Balances.bToken.user2 += bTokens.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx2.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to transfer from specified address to specified address without allowance", async () => {
    await expect(
      fixture.bToken
        .connect(admin)
        .transferFrom(user.address, user2.address, bTokens)
    ).to.be.revertedWith("ERC20: insufficient allowance");
  });

  it("Should fail to transfer from specified address to specified address without enough allowance", async () => {
    await fixture.bToken
      .connect(user)
      .approve(admin.address, BigNumber.from(bTokens).sub(1));
    await expect(
      fixture.bToken
        .connect(admin)
        .transferFrom(user.address, user2.address, bTokens)
    ).to.be.revertedWith("ERC20: insufficient allowance");
  });
});
