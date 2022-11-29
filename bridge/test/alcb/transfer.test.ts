import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getEthConsumedAsGas, getState, state } from "./setup";

describe("Testing ALCB Transfer methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 4;
  const minALCBs = 0;
  let alcbs: BigNumber;

  async function deployFixture() {
    const fixture = await getFixture();
    const signers = await ethers.getSigners();
    const [admin, user, user2] = signers;

    const ethIn = ethers.utils.parseEther(eth.toString());
    // Mint some ALCAs for testing
    const [alcbs] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mint",
      user,
      [minALCBs],
      ethIn
    );
    return { fixture, admin, user, user2, ethIn, alcbs };
  }

  beforeEach(async function () {
    ({ fixture, admin, user, user2, alcbs } = await loadFixture(deployFixture));

    expectedState = await getState(fixture);
  });

  it("Should transfer from sender to specified address", async () => {
    const tx = await fixture.alcb.connect(user).transfer(admin.address, alcbs);
    expectedState.Balances.alcb.admin += alcbs.toBigInt();
    expectedState.Balances.alcb.user -= alcbs.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should transfer from specified address to specified address", async () => {
    const tx = await fixture.alcb.connect(user).approve(admin.address, alcbs);
    const tx2 = await fixture.alcb
      .connect(admin)
      .transferFrom(user.address, user2.address, alcbs);
    expectedState.Balances.alcb.user -= alcbs.toBigInt();
    expectedState.Balances.alcb.user2 += alcbs.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx2.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to transfer from specified address to specified address without allowance", async () => {
    await expect(
      fixture.alcb
        .connect(admin)
        .transferFrom(user.address, user2.address, alcbs)
    ).to.be.revertedWith("ERC20: insufficient allowance");
  });

  it("Should fail to transfer from specified address to specified address without enough allowance", async () => {
    await fixture.alcb
      .connect(user)
      .approve(admin.address, BigNumber.from(alcbs).sub(1));
    await expect(
      fixture.alcb
        .connect(admin)
        .transferFrom(user.address, user2.address, alcbs)
    ).to.be.revertedWith("ERC20: insufficient allowance");
  });
});
