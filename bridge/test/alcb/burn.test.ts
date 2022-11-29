import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getEthConsumedAsGas, getState, state } from "./setup";

describe("Testing ALCB Burning methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 40;
  let ethIn: BigNumber;
  let alcbs: BigNumber;
  const minALCBs = 0;
  const marketSpread = 4;

  async function deployFixture() {
    const fixture = await getFixture();
    const signers = await ethers.getSigners();
    const [admin, user] = signers;
    const ethIn = ethers.utils.parseEther(eth.toString());
    const [alcbs] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mint",
      user,
      [minALCBs],
      ethIn
    );
    return { fixture, admin, user, ethIn, alcbs };
  }

  beforeEach(async function () {
    ({ fixture, admin, user, ethIn, alcbs } = await loadFixture(deployFixture));

    expectedState = await getState(fixture);
  });

  it("Should burn", async () => {
    const remaining = ethers.utils.parseUnits("100").toBigInt();
    let burnQuantity = BigNumber.from(alcbs).sub(remaining).toBigInt();
    expectedState = await getState(fixture);
    const [eths, tx] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expect(alcbs).to.be.equal(BigInt("4020217121585928137263"));
    expect(eths).to.be.equal(BigInt("9751261920046697614"));
    expectedState.Balances.alcb.user -= burnQuantity;
    expectedState.Balances.eth.user += eths.toBigInt();
    expectedState.Balances.alcb.poolBalance -= eths.toBigInt();
    expectedState.Balances.alcb.totalSupply -= burnQuantity;
    expectedState.Balances.eth.alcb -= eths.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
    burnQuantity = remaining;
    expectedState = await getState(fixture);
    const [eths2, tx2] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expect(eths).to.be.equal(BigInt("9751261920046697614"));
    expectedState.Balances.alcb.user -= burnQuantity;
    expectedState.Balances.eth.user += eths2.toBigInt();
    expectedState.Balances.alcb.poolBalance -= eths2.toBigInt();
    expectedState.Balances.alcb.totalSupply -= burnQuantity;
    expectedState.Balances.eth.alcb -= eths2.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx2.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should burn a huge amount of ALCBs", async () => {
    const eth = 70000000000;
    const ethIn = ethers.utils.parseEther(eth.toString());
    const [alcbs] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mintTo",
      admin,
      [user.address, minALCBs],
      ethIn
    );
    const burnQuantity = alcbs.toBigInt();
    expectedState = await getState(fixture);
    const [eths, tx] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expectedState.Balances.alcb.user -= burnQuantity;
    expectedState.Balances.eth.user += eths.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.alcb.poolBalance -= eths.toBigInt();
    expectedState.Balances.alcb.totalSupply -= burnQuantity;
    expectedState.Balances.eth.alcb -= eths.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to burn more than possible", async () => {
    // Try to burn one more than minted
    const burnQuantity = BigNumber.from(alcbs).add(1);

    await expect(fixture.alcb.connect(user).burn(burnQuantity, 0))
      .to.be.revertedWithCustomError(fixture.alcb, `BurnAmountExceedsSupply`)
      .withArgs(burnQuantity, alcbs);
  });

  it("Should fail to burn 0 tokens", async () => {
    // Try to burn 0
    const burnQuantity = 0;
    await fixture.alcb.connect(user).mint(minALCBs, {
      value: ethers.utils.parseEther(eth.toString()),
    });

    await expect(fixture.alcb.connect(user).burn(burnQuantity, 0))
      .to.be.revertedWithCustomError(fixture.alcb, `InvalidBurnAmount`)
      .withArgs(burnQuantity);
  });

  it("Should burn to an address", async () => {
    const burnQuantity = alcbs;
    // Round eths to avoid consumed gas on state comparison
    expectedState = await getState(fixture);
    // Eths to receive
    const [eths, tx] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "burnTo",
      user,
      [admin.address, burnQuantity, 0]
    );
    expectedState.Balances.alcb.user -= burnQuantity.toBigInt();
    expectedState.Balances.eth.admin += eths.toBigInt();
    expectedState.Balances.alcb.poolBalance -= eths.toBigInt();
    expectedState.Balances.alcb.totalSupply -= burnQuantity.toBigInt();
    expectedState.Balances.eth.alcb -= eths.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to burn to an address more than possible", async () => {
    // Try to burn one more than minted
    const burnQuantity = BigNumber.from(alcbs).add(1);

    await expect(fixture.alcb.connect(user).burn(burnQuantity, 0))
      .to.be.revertedWithCustomError(fixture.alcb, `BurnAmountExceedsSupply`)
      .withArgs(burnQuantity, alcbs);
  });

  it("Should fail to burn 0 tokens to an address", async () => {
    const burnQuantity = 0;

    await expect(
      fixture.alcb.connect(user).burnTo(admin.address, burnQuantity, 0)
    )
      .to.be.revertedWithCustomError(fixture.alcb, `InvalidBurnAmount`)
      .withArgs(burnQuantity);
  });

  it("Should fail to burn without fulfilling min eth amount", async () => {
    const minEth = ethIn.add(1);
    const burnQuantity = alcbs;

    await expect(
      fixture.alcb
        .connect(user)
        .burnTo(
          admin.address,
          burnQuantity,
          ethers.utils.parseEther(minEth.toString())
        )
    )
      .to.be.revertedWithCustomError(fixture.alcb, `MinimumBurnNotMet`)
      .withArgs(
        "10000000000000000000",
        ethers.utils.parseEther(minEth.toString())
      );
  });

  it("Should burn and keep market spread", async () => {
    const burnQuantity = alcbs;
    const [eths, tx] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expectedState.Balances.alcb.user -= burnQuantity.toBigInt();
    expectedState.Balances.eth.user += ethIn.div(marketSpread).toBigInt();
    expectedState.Balances.alcb.poolBalance -= eths.toBigInt();
    expectedState.Balances.alcb.totalSupply -= burnQuantity.toBigInt();
    expectedState.Balances.eth.alcb -= eths.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should mint and burn a lot", async () => {
    const it = 10;
    for (let i = 0; i < it; i++) {
      const [alcbs] = await callFunctionAndGetReturnValues(
        fixture.alcb,
        "mint",
        user,
        [minALCBs],
        ethIn
      );
      const burnQuantity = alcbs.toBigInt();
      expectedState = await getState(fixture);
      const [eths, tx] = await callFunctionAndGetReturnValues(
        fixture.alcb,
        "burn",
        user,
        [burnQuantity, 0]
      );
      // Round eths to avoid consumed gas on state comparison
      expectedState.Balances.alcb.user -= burnQuantity;
      expectedState.Balances.eth.user += eths.toBigInt();
      expectedState.Balances.alcb.poolBalance -= eths.toBigInt();
      expectedState.Balances.alcb.totalSupply -= burnQuantity;
      expectedState.Balances.eth.alcb -= eths.toBigInt();
      expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
      expect(await getState(fixture)).to.be.deep.equal(expectedState);
    }
  });

  it("Should burn it all", async () => {
    const burnQuantity = await fixture.alcb.totalSupply();
    expectedState = await getState(fixture);
    // Eths to receive
    expectedState = await getState(fixture);
    await callFunctionAndGetReturnValues(fixture.alcb, "burn", user, [
      burnQuantity,
      0,
    ]);
    expect(await fixture.alcb.totalSupply()).to.be.equal(0);
  });

  it("Should keep ether sent value and (mint/burn) value relation greater or equal to market spread value", async () => {
    const ethIn = ethers.utils.parseEther("8000");
    const [alcbs1] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mintTo",
      admin,
      [user.address, minALCBs],
      ethIn
    );
    const [ethOut1] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "burn",
      user,
      [alcbs1, 0]
    );
    expect(ethIn.div(ethOut1).toNumber()).to.be.greaterThanOrEqual(
      marketSpread
    );
  });

  it("Should keep ether sent value and (mint/burn) value relation greater or equal to market spread value with subsequent minting", async () => {
    // If some consequent mints are performed this relation is not kept between two first operations
    const ethIn = ethers.utils.parseEther("4");
    const [alcbs1] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mintTo",
      admin,
      [user.address, minALCBs],
      ethIn
    );
    const [alcbs2] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mintTo",
      admin,
      [user.address, minALCBs],
      ethIn
    );
    const [alcbs3] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mintTo",
      admin,
      [user.address, minALCBs],
      ethIn
    );
    const [ethOut1] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "burn",
      user,
      [alcbs1, 0]
    );
    const [ethOut2] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "burn",
      user,
      [alcbs2, 0]
    );
    const [ethOut3] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "burn",
      user,
      [alcbs3, 0]
    );
    const expectedAmount =
      (ethIn.toBigInt() * 3n) /
      (ethOut3.toBigInt() + ethOut2.toBigInt() + ethOut1.toBigInt());
    expect(expectedAmount).to.be.equals(
      BigInt(marketSpread),
      "Failed at assertion"
    );
  });
});
