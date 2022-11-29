import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getEthConsumedAsGas, getState, showState, state } from "./setup";

describe("Testing ALCB Minting methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 4;
  const marketSpread = 4;
  let ethInWeis: BigNumber;
  const minALCBs = 0;

  beforeEach(async function () {
    fixture = await loadFixture(getFixture);
    const signers = await ethers.getSigners();
    [admin, user, user2] = signers;
    showState("Initial", await getState(fixture));
    ethInWeis = ethers.utils.parseEther(eth.toString());
  });

  it("Should mint alcbs to sender out of ether value sent", async () => {
    expectedState = await getState(fixture);
    const [alcbs, tx] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mint",
      user,
      [minALCBs],
      ethInWeis
    );
    expect(alcbs).to.be.equal(BigInt("402028731704364116575"));
    expectedState.Balances.eth.user -= ethInWeis.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.alcb.user += alcbs.toBigInt();
    expectedState.Balances.alcb.totalSupply += alcbs.toBigInt();
    expectedState.Balances.alcb.poolBalance += ethInWeis
      .div(marketSpread)
      .toBigInt();
    expectedState.Balances.eth.alcb += ethInWeis.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(
      expectedState,
      "State 1 comp"
    );
    expectedState = await getState(fixture);
    const [alcbs2, tx2] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mint",
      user2,
      [minALCBs],
      ethInWeis
    );
    expectedState.Balances.eth.user2 -= ethInWeis.toBigInt();
    expectedState.Balances.eth.user2 -= getEthConsumedAsGas(await tx2.wait());
    expectedState.Balances.alcb.user2 += alcbs2.toBigInt();
    expectedState.Balances.alcb.totalSupply += alcbs2.toBigInt();
    expectedState.Balances.alcb.poolBalance += ethInWeis.div(4).toBigInt();
    expectedState.Balances.eth.alcb += ethInWeis.toBigInt();
    expect(alcbs2).to.be.equal(BigInt("402027176702820751481"));
    expect(await getState(fixture)).to.be.deep.equal(
      expectedState,
      "State 2 comp"
    );
  });

  it("Should mint with huge amount of eth", async () => {
    const eth = 70000000000;
    ethInWeis = ethers.utils.parseEther(eth.toString());
    expectedState = await getState(fixture);
    const [alcbs] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mint",
      admin,
      [minALCBs],
      ethInWeis
    );
    expect(alcbs).to.be.equal(BigInt("70001004975246203818081563855"));
    expect(await fixture.alcb.getPoolBalance()).to.be.equal(
      ethInWeis.div(marketSpread).toBigInt()
    );
    expect(await fixture.alcb.totalSupply()).to.be.equal(alcbs);
  });

  it("Should mint to an address", async () => {
    expectedState = await getState(fixture);
    const [alcbs, tx] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mintTo",
      admin,
      [user.address, minALCBs],
      ethInWeis
    );
    expect(alcbs).to.be.equal(BigInt("402028731704364116575"));
    expectedState.Balances.eth.admin -= ethInWeis.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.alcb.user += alcbs.toBigInt();
    expectedState.Balances.alcb.totalSupply += alcbs.toBigInt();
    expectedState.Balances.alcb.poolBalance += ethInWeis
      .div(marketSpread)
      .toBigInt();
    expectedState.Balances.eth.alcb += ethInWeis.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
    expectedState = await getState(fixture);
    const [alcbs2, tx2] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mintTo",
      admin,
      [user.address, minALCBs],
      ethInWeis
    );
    expectedState.Balances.eth.admin -= ethInWeis.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx2.wait());
    expectedState.Balances.alcb.user += alcbs2.toBigInt();
    expectedState.Balances.alcb.totalSupply += alcbs2.toBigInt();
    expectedState.Balances.alcb.poolBalance += ethInWeis
      .div(marketSpread)
      .toBigInt();
    expectedState.Balances.eth.alcb += ethInWeis.toBigInt();
    expect(alcbs2).to.be.equal(BigInt("402027176702820751481"));
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should mint a big amount of eth to an address", async () => {
    const eth = 70000000000;
    ethInWeis = ethers.utils.parseEther(eth.toString());
    expectedState = await getState(fixture);
    const [alcbs, tx] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mintTo",
      admin,
      [user.address, minALCBs],
      ethInWeis
    );
    expect(alcbs).to.be.equal(BigInt("70001004975246203818081563855"));
    expectedState.Balances.eth.admin -= ethInWeis.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.alcb.user += alcbs.toBigInt();
    expectedState.Balances.alcb.totalSupply += alcbs.toBigInt();
    expectedState.Balances.alcb.poolBalance += ethInWeis
      .div(marketSpread)
      .toBigInt();
    expectedState.Balances.eth.alcb += ethInWeis.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to mint to 0x0 address", async () => {
    await expect(
      fixture.alcb
        .connect(admin)
        .mintTo(ethers.constants.AddressZero, minALCBs, {
          value: ethers.utils.parseEther(eth.toString()),
        })
    ).to.be.revertedWith("ERC20: mint to the zero address");
  });

  it("Should fail to mint with big min ALCB quantity", async () => {
    const oneALCB = ethers.utils.parseUnits("1", 18).toBigInt();
    const minALCBs = 900n * oneALCB;
    const expectedALCBsMintedForEthValue = "402028731704364116575";
    await expect(
      fixture.alcb.connect(admin).mint(minALCBs, {
        value: ethers.utils.parseEther(eth.toString()),
      })
    )
      .to.be.revertedWithCustomError(fixture.alcb, `MinimumMintNotMet`)
      .withArgs(expectedALCBsMintedForEthValue, minALCBs);
  });
});
