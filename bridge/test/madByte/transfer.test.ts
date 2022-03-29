import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  BaseTokensFixture,
  callFunctionAndGetReturnValues,
  getBaseTokensFixture,
} from "../setup";
import { getState, showState, state } from "./setup";

describe("Testing MadByte Transfer methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let fixture: BaseTokensFixture;
  const eth = 4;
  let ethIn: BigNumber;
  const minMadBytes = 0;
  let madBytes: BigNumber;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    const signers = await ethers.getSigners();
    [admin, user, user2] = signers;
    showState("Initial", await getState(fixture));
    ethIn = ethers.utils.parseEther(eth.toString());
    // Mint some MB for testing
    [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      user,
      [minMadBytes],
      ethIn
    );
    expectedState = await getState(fixture);
  });

  it("Should transfer from sender to specified address", async () => {
    await fixture.madByte.connect(user).transfer(admin.address, madBytes);
    expectedState.Balances.madByte.admin += madBytes.toBigInt();
    expectedState.Balances.madByte.user -= madBytes.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should transfer from specified address to specified address", async () => {
    await fixture.madByte.connect(user).approve(admin.address, madBytes);
    await fixture.madByte
      .connect(admin)
      .transferFrom(user.address, user2.address, madBytes);
    expectedState.Balances.madByte.user -= madBytes.toBigInt();
    expectedState.Balances.madByte.user2 += madBytes.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to transfer from specified address to specified address without allowance", async () => {
    await expect(
      fixture.madByte
        .connect(admin)
        .transferFrom(user.address, user2.address, madBytes)
    ).to.be.revertedWith("ERC20: insufficient allowance");
  });

  it("Should fail to transfer from specified address to specified address without enough allowance", async () => {
    await fixture.madByte
      .connect(user)
      .approve(admin.address, BigNumber.from(madBytes).sub(1));
    await expect(
      fixture.madByte
        .connect(admin)
        .transferFrom(user.address, user2.address, madBytes)
    ).to.be.revertedWith("ERC20: insufficient allowance");
  });
});
