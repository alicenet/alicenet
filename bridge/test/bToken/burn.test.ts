import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getEthConsumedAsGas, getState, state } from "./setup";

describe("Testing BToken Burning methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 40;
  let ethIn: BigNumber;
  let bTokens: BigNumber;
  const minBTokens = 0;
  const marketSpread = 4;

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [admin, user] = signers;
    ethIn = ethers.utils.parseEther(eth.toString());
    [bTokens] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      user,
      [minBTokens],
      ethIn
    );
    expectedState = await getState(fixture);
  });

  it("Should burn", async () => {
    const remaining = ethers.utils.parseUnits("100").toBigInt();
    let burnQuantity = BigNumber.from(bTokens).sub(remaining).toBigInt();
    expectedState = await getState(fixture);
    const [eths, tx] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expect(bTokens).to.be.equal(BigInt("3990217121585928137263"));
    expect(eths).to.be.equal(BigInt("9749391845405398553"));
    expectedState.Balances.bToken.user -= burnQuantity;
    expectedState.Balances.eth.user += eths.toBigInt();
    expectedState.Balances.bToken.poolBalance -= eths.toBigInt();
    expectedState.Balances.bToken.totalSupply -= burnQuantity;
    expectedState.Balances.eth.bToken -= eths.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
    burnQuantity = remaining;
    expectedState = await getState(fixture);
    const [eths2, tx2] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expect(eths).to.be.equal(BigInt("9749391845405398553"));
    expectedState.Balances.bToken.user -= burnQuantity;
    expectedState.Balances.eth.user += eths2.toBigInt();
    expectedState.Balances.bToken.poolBalance -= eths2.toBigInt();
    expectedState.Balances.bToken.totalSupply -= burnQuantity;
    expectedState.Balances.eth.bToken -= eths2.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx2.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should burn a huge amount of BTokens", async () => {
    const eth = 70000000000;
    const ethIn = ethers.utils.parseEther(eth.toString());
    const [bTokens] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      admin,
      [user.address, minBTokens],
      ethIn
    );
    const burnQuantity = bTokens.toBigInt();
    expectedState = await getState(fixture);
    const [eths, tx] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expectedState.Balances.bToken.user -= burnQuantity;
    expectedState.Balances.eth.user += eths.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.bToken.poolBalance -= eths.toBigInt();
    expectedState.Balances.bToken.totalSupply -= burnQuantity;
    expectedState.Balances.eth.bToken -= eths.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to burn more than possible", async () => {
    // Try to burn one more than minted
    const burnQuantity = BigNumber.from(bTokens).add(1);

    await expect(fixture.bToken.connect(user).burn(burnQuantity, 0))
      .to.be.revertedWithCustomError(fixture.bToken, `BurnAmountExceedsSupply`)
      .withArgs(burnQuantity, bTokens);
  });

  it("Should fail to burn 0 tokens", async () => {
    // Try to burn 0
    const burnQuantity = 0;
    await fixture.bToken.connect(user).mint(minBTokens, {
      value: ethers.utils.parseEther(eth.toString()),
    });

    await expect(fixture.bToken.connect(user).burn(burnQuantity, 0))
      .to.be.revertedWithCustomError(fixture.bToken, `InvalidBurnAmount`)
      .withArgs(burnQuantity);
  });

  it("Should burn to an address", async () => {
    const burnQuantity = bTokens;
    // Round eths to avoid consumed gas on state comparison
    expectedState = await getState(fixture);
    // Eths to receive
    const [eths, tx] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "burnTo",
      user,
      [admin.address, burnQuantity, 0]
    );
    expectedState.Balances.bToken.user -= burnQuantity.toBigInt();
    expectedState.Balances.eth.admin += eths.toBigInt();
    expectedState.Balances.bToken.poolBalance -= eths.toBigInt();
    expectedState.Balances.bToken.totalSupply -= burnQuantity.toBigInt();
    expectedState.Balances.eth.bToken -= eths.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to burn to an address more than possible", async () => {
    // Try to burn one more than minted
    const burnQuantity = BigNumber.from(bTokens).add(1);

    await expect(fixture.bToken.connect(user).burn(burnQuantity, 0))
      .to.be.revertedWithCustomError(fixture.bToken, `BurnAmountExceedsSupply`)
      .withArgs(burnQuantity, bTokens);
  });

  it("Should fail to burn 0 tokens to an address", async () => {
    const burnQuantity = 0;

    await expect(
      fixture.bToken.connect(user).burnTo(admin.address, burnQuantity, 0)
    )
      .to.be.revertedWithCustomError(fixture.bToken, `InvalidBurnAmount`)
      .withArgs(burnQuantity);
  });

  it("Should fail to burn without fulfilling min eth amount", async () => {
    const minEth = ethIn.add(1);
    const burnQuantity = bTokens;

    await expect(
      fixture.bToken
        .connect(user)
        .burnTo(
          admin.address,
          burnQuantity,
          ethers.utils.parseEther(minEth.toString())
        )
    )
      .to.be.revertedWithCustomError(fixture.bToken, `MinimumBurnNotMet`)
      .withArgs(
        "10000000000000000000",
        ethers.utils.parseEther(minEth.toString())
      );
  });

  it("Should burn and keep market spread", async () => {
    const burnQuantity = bTokens;
    const [eths, tx] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expectedState.Balances.bToken.user -= burnQuantity.toBigInt();
    expectedState.Balances.eth.user += ethIn.div(marketSpread).toBigInt();
    expectedState.Balances.bToken.poolBalance -= eths.toBigInt();
    expectedState.Balances.bToken.totalSupply -= burnQuantity.toBigInt();
    expectedState.Balances.eth.bToken -= eths.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should mint and burn a lot", async () => {
    const it = 10;
    for (let i = 0; i < it; i++) {
      const [bTokens] = await callFunctionAndGetReturnValues(
        fixture.bToken,
        "mint",
        user,
        [minBTokens],
        ethIn
      );
      const burnQuantity = bTokens.toBigInt();
      expectedState = await getState(fixture);
      const [eths, tx] = await callFunctionAndGetReturnValues(
        fixture.bToken,
        "burn",
        user,
        [burnQuantity, 0]
      );
      // Round eths to avoid consumed gas on state comparison
      expectedState.Balances.bToken.user -= burnQuantity;
      expectedState.Balances.eth.user += eths.toBigInt();
      expectedState.Balances.bToken.poolBalance -= eths.toBigInt();
      expectedState.Balances.bToken.totalSupply -= burnQuantity;
      expectedState.Balances.eth.bToken -= eths.toBigInt();
      expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
      expect(await getState(fixture)).to.be.deep.equal(expectedState);
    }
  });

  it("Should burn it all", async () => {
    const burnQuantity = await fixture.bToken.totalSupply();
    expectedState = await getState(fixture);
    // Eths to receive
    expectedState = await getState(fixture);
    await callFunctionAndGetReturnValues(fixture.bToken, "burn", user, [
      burnQuantity,
      0,
    ]);
    expect(await fixture.bToken.totalSupply()).to.be.equal(0);
  });

  it("Should keep ether sent value and (mint/burn) value relation greater or equal to market spread value", async () => {
    const ethIn = ethers.utils.parseEther("8000");
    const [bTokens1] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      admin,
      [user.address, minBTokens],
      ethIn
    );
    const [ethOut1] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "burn",
      user,
      [bTokens1, 0]
    );
    expect(ethIn.div(ethOut1).toNumber()).to.be.greaterThanOrEqual(
      marketSpread
    );
  });

  it("Should keep ether sent value and (mint/burn) value relation greater or equal to market spread value with subsequent minting", async () => {
    // If some consequent mints are performed this relation is not kept between two first operations
    const ethIn = ethers.utils.parseEther("4");
    const [bTokens1] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      admin,
      [user.address, minBTokens],
      ethIn
    );
    const [bTokens2] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      admin,
      [user.address, minBTokens],
      ethIn
    );
    const [bTokens3] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      admin,
      [user.address, minBTokens],
      ethIn
    );
    const [ethOut1] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "burn",
      user,
      [bTokens1, 0]
    );
    const [ethOut2] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "burn",
      user,
      [bTokens2, 0]
    );
    const [ethOut3] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "burn",
      user,
      [bTokens3, 0]
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
