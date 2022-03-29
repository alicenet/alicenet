import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { format, getState, state } from "./setup";

describe("Testing MadByte Burning methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let eths: BigNumber;
  let fixture: Fixture;
  const eth = 40;
  let ethIn: BigNumber;
  let madBytes: BigNumber;
  const minMadBytes = 0;
  const marketSpread = 4;

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [admin, user] = signers;
    ethIn = ethers.utils.parseEther(eth.toString());
    [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      user,
      [minMadBytes],
      ethIn
    );
    expectedState = await getState(fixture);
  });

  it("Should burn", async () => {
    const remaining = ethers.utils.parseUnits("100").toBigInt();
    let burnQuantity = BigNumber.from(madBytes).sub(remaining).toBigInt();
    expectedState = await getState(fixture);
    [eths] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [burnQuantity, 0]
    );
    let roundedEths = format(eths);
    expect(madBytes).to.be.equal(BigInt("3990217121585928137263"));
    expect(eths).to.be.equal(BigInt("9749391845405398553"));
    expectedState.Balances.madByte.user -= burnQuantity;
    expectedState.Balances.eth.user += roundedEths;
    expectedState.Balances.madByte.poolBalance -= eths.toBigInt();
    expectedState.Balances.madByte.totalSupply -= burnQuantity;
    expectedState.Balances.eth.madByte -= eths.toBigInt();
    // await fixture.madByte.connect(user).burn(burnQuantity, 0);
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
    burnQuantity = remaining;
    expectedState = await getState(fixture);
    [eths] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [burnQuantity, 0]
    );
    roundedEths = format(eths);
    expect(eths).to.be.equal(BigInt("250608154594601447"));
    expectedState.Balances.madByte.user -= burnQuantity;
    expectedState.Balances.eth.user += roundedEths;
    expectedState.Balances.madByte.poolBalance -= eths.toBigInt();
    expectedState.Balances.madByte.totalSupply -= burnQuantity;
    expectedState.Balances.eth.madByte -= eths.toBigInt();
    // await fixture.madByte.connect(user).burn(remaining, 0);
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should burn a huge amount of MadBytes", async () => {
    const eth = 70000000000;
    const ethIn = ethers.utils.parseEther(eth.toString());
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mintTo",
      admin,
      [user.address, minMadBytes],
      ethIn
    );
    const burnQuantity = madBytes.toBigInt();
    expectedState = await getState(fixture);
    [eths] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [burnQuantity, 0]
    );
    const roundedEths = format(eths);
    expectedState.Balances.madByte.user -= burnQuantity;
    expectedState.Balances.eth.user += roundedEths;
    expectedState.Balances.madByte.poolBalance -= eths.toBigInt();
    expectedState.Balances.madByte.totalSupply -= burnQuantity;
    expectedState.Balances.eth.madByte -= eths.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to burn more than possible", async () => {
    // Try to burn one more than minted
    const burnQuantity = BigNumber.from(madBytes).add(1);
    await expect(
      fixture.madByte.connect(user).burn(burnQuantity, 0)
    ).to.be.revertedWith(
      "MadByte: The number of tokens to be burned is greater than the Total Supply!"
    );
  });

  it("Should fail to burn 0 tokens", async () => {
    // Try to burn 0
    const burnQuantity = 0;
    await fixture.madByte.connect(user).mint(minMadBytes, {
      value: ethers.utils.parseEther(eth.toString()),
    });
    await expect(
      fixture.madByte.connect(user).burn(burnQuantity, 0)
    ).to.be.revertedWith(
      "MadByte: The number of MadBytes to be burn should be greater than 0!"
    );
  });

  it("Should burn to an address", async () => {
    const burnQuantity = madBytes;
    // Round eths to avoid consumed gas on state comparison
    expectedState = await getState(fixture);
    // Eths to receive
    [eths] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burnTo",
      user,
      [admin.address, burnQuantity, 0]
    );
    const roundedEths = format(eths);
    expectedState.Balances.madByte.user -= burnQuantity.toBigInt();
    expectedState.Balances.eth.admin += roundedEths;
    expectedState.Balances.madByte.poolBalance -= eths.toBigInt();
    expectedState.Balances.madByte.totalSupply -= burnQuantity.toBigInt();
    expectedState.Balances.eth.madByte -= eths.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to burn to an address more than possible", async () => {
    // Try to burn one more than minted
    const burnQuantity = BigNumber.from(madBytes).add(1);
    await expect(
      fixture.madByte.connect(user).burnTo(admin.address, burnQuantity, 0)
    ).to.be.revertedWith(
      "MadByte: The number of tokens to be burned is greater than the Total Supply!"
    );
  });

  it("Should fail to burn 0 tokens to an address", async () => {
    const burnQuantity = 0;
    await expect(
      fixture.madByte.connect(user).burnTo(admin.address, burnQuantity, 0)
    ).to.be.revertedWith(
      "MadByte: The number of MadBytes to be burn should be greater than 0!"
    );
  });

  it("Should fail to burn without fulfilling min eth amount", async () => {
    const minEth = ethIn.add(1);
    const burnQuantity = madBytes;
    await expect(
      fixture.madByte
        .connect(user)
        .burnTo(
          admin.address,
          burnQuantity,
          ethers.utils.parseEther(minEth.toString())
        )
    ).to.be.revertedWith("MadByte: Couldn't burn the minEth amount");
  });

  it("Should burn and keep market spread", async () => {
    const burnQuantity = madBytes;
    // Eths to receive
    [eths] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expectedState.Balances.madByte.user -= burnQuantity.toBigInt();
    expectedState.Balances.eth.user += eth / marketSpread;
    expectedState.Balances.madByte.poolBalance -= eths.toBigInt();
    expectedState.Balances.madByte.totalSupply -= burnQuantity.toBigInt();
    expectedState.Balances.eth.madByte -= eths.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should mint and burn a lot", async () => {
    const it = 10;
    for (let i = 0; i < it; i++) {
      const [madBytes] = await callFunctionAndGetReturnValues(
        fixture.madByte,
        "mint",
        user,
        [minMadBytes],
        ethIn
      );
      const burnQuantity = madBytes.toBigInt();
      expectedState = await getState(fixture);
      // Eths to receive
      expectedState = await getState(fixture);
      [eths] = await callFunctionAndGetReturnValues(
        fixture.madByte,
        "burn",
        user,
        [burnQuantity, 0]
      );
      // Round eths to avoid consumed gas on state comparison
      const roundedEths = format(eths);
      expectedState.Balances.madByte.user -= burnQuantity;
      expectedState.Balances.eth.user += roundedEths;
      expectedState.Balances.madByte.poolBalance -= eths.toBigInt();
      expectedState.Balances.madByte.totalSupply -= burnQuantity;
      expectedState.Balances.eth.madByte -= eths.toBigInt();
      expect(await getState(fixture)).to.be.deep.equal(expectedState);
    }
  });

  it("Should burn it all", async () => {
    const burnQuantity = await fixture.madByte.totalSupply();
    expectedState = await getState(fixture);
    // Eths to receive
    expectedState = await getState(fixture);
    [eths] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [burnQuantity, 0]
    );
    expect(await fixture.madByte.totalSupply()).to.be.equal(0);
  });

  it("Should keep ether sent value and (mint/burn) value relation greater or equal to market spread value", async () => {
    const ethIn = ethers.utils.parseEther("8000");
    const [madBytes1] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mintTo",
      admin,
      [user.address, minMadBytes],
      ethIn
    );
    const [ethOut1] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [madBytes1, 0]
    );
    expect(ethIn.div(ethOut1).toNumber()).to.be.greaterThanOrEqual(
      marketSpread
    );
  });

  it("Should keep ether sent value and (mint/burn) value relation greater or equal to market spread value with subsequent minting", async () => {
    // If some consequent mints are performed this relation is not kept between two first operations
    const ethIn = ethers.utils.parseEther("4");
    const [madBytes1] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mintTo",
      admin,
      [user.address, minMadBytes],
      ethIn
    );
    const [madBytes2] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mintTo",
      admin,
      [user.address, minMadBytes],
      ethIn
    );
    const [madBytes3] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mintTo",
      admin,
      [user.address, minMadBytes],
      ethIn
    );
    const [ethOut1] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [madBytes1, 0]
    );
    const [ethOut2] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [madBytes2, 0]
    );
    const [ethOut3] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [madBytes3, 0]
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
