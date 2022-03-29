import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { MadByte } from "../../../typechain-types";
import { expect } from "../../chai-setup";
import {
  callFunctionAndGetReturnValues,
  Fixture,
  getFixture,
} from "../../setup";
import { format, getState, init, showState, state } from "./setup";

describe("Testing MadByte Burning methods", async () => {
  let madByte: MadByte;
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let currentState: state;
  let amount = 1000;
  let eths: BigNumber;
  let fixture: Fixture;
  let userAddresses: string[];
  let contractAddresses: string[];
  let minerAmount: BigNumber;
  let stakingAmount: BigNumber;
  let lpStakingAmount: BigNumber;
  let foundationAmount: BigNumber;
  let eth = 40;
  let mad = 10;
  let ethIn: BigNumber;
  let madDeposit: BigNumber;
  let burnQuantity: BigNumber;
  let madBytes: BigNumber;
  let minMadBytes = 0;

  let marketSpread = 4;
  const zeroAddress = "0x0000000000000000000000000000000000000000";
  const ONE_MB = 1 * 10 ** 18;

  beforeEach(async function () {
    fixture = await getFixture();
    let signers = await ethers.getSigners();
    [admin, user, user2] = signers;
    await init(fixture);
    // let expectedState = await getState(contractAddresses, userAddresses);
    showState("Initial", await getState(fixture));
    ethIn = ethers.utils.parseEther(eth.toString());
    [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      user,
      [minMadBytes],
      ethIn
    );
    burnQuantity = madBytes;
    expectedState = await getState(fixture);
  });

  it("Should burn", async () => {
    let remaining = ethers.utils.parseUnits("100").toBigInt();
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
    let eth = 70000000000;
    let ethIn = ethers.utils.parseEther(eth.toString());
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mintTo",
      admin,
      [user.address, minMadBytes],
      ethIn
    );
    let burnQuantity = madBytes.toBigInt();
    expectedState = await getState(fixture);
    [eths] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burn",
      user,
      [burnQuantity, 0]
    );
    let roundedEths = format(eths);
    expectedState.Balances.madByte.user -= burnQuantity;
    expectedState.Balances.eth.user += roundedEths;
    expectedState.Balances.madByte.poolBalance -= eths.toBigInt();
    expectedState.Balances.madByte.totalSupply -= burnQuantity;
    expectedState.Balances.eth.madByte -= eths.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to burn more than possible", async () => {
    // Try to burn one more than minted
    let burnQuantity = BigNumber.from(madBytes).add(1);
    await expect(
      fixture.madByte.connect(user).burn(burnQuantity, 0)
    ).to.be.revertedWith(
      "MadByte: The number of tokens to be burned is greater than the Total Supply!"
    );
  });

  it("Should fail to burn 0 tokens", async () => {
    // Try to burn 0
    let burnQuantity = 0;
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
    let burnQuantity = madBytes;
    // Round eths to avoid consumed gas on state comparison
    expectedState = await getState(fixture);
    //Eths to receive
    [eths] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "burnTo",
      user,
      [admin.address, burnQuantity, 0]
    );
    let roundedEths = format(eths);
    expectedState.Balances.madByte.user -= burnQuantity.toBigInt();
    expectedState.Balances.eth.admin += roundedEths;
    expectedState.Balances.madByte.poolBalance -= eths.toBigInt();
    expectedState.Balances.madByte.totalSupply -= burnQuantity.toBigInt();
    expectedState.Balances.eth.madByte -= eths.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to burn to an address more than possible", async () => {
    // Try to burn one more than minted
    let burnQuantity = BigNumber.from(madBytes).add(1);
    await expect(
      fixture.madByte.connect(user).burnTo(admin.address, burnQuantity, 0)
    ).to.be.revertedWith(
      "MadByte: The number of tokens to be burned is greater than the Total Supply!"
    );
  });

  it("Should fail to burn 0 tokens to an address", async () => {
    let burnQuantity = 0;
    await expect(
      fixture.madByte.connect(user).burnTo(admin.address, burnQuantity, 0)
    ).to.be.revertedWith(
      "MadByte: The number of MadBytes to be burn should be greater than 0!"
    );
  });

  it("Should fail to burn without fulfilling min eth amount", async () => {
    let minEth = ethIn.add(1);
    let burnQuantity = madBytes;
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
    let burnQuantity = madBytes;
    //Eths to receive
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
    let totalEth = 40;
    let it = 10;
    let eth = totalEth / it;
    for (let i = 0; i < it; i++) {
      const [madBytes] = await callFunctionAndGetReturnValues(
        fixture.madByte,
        "mint",
        user,
        [minMadBytes],
        ethIn
      );
      let burnQuantity = madBytes.toBigInt();
      expectedState = await getState(fixture);
      //Eths to receive
      expectedState = await getState(fixture);
      [eths] = await callFunctionAndGetReturnValues(
        fixture.madByte,
        "burn",
        user,
        [burnQuantity, 0]
      );
      // Round eths to avoid consumed gas on state comparison
      let roundedEths = format(eths);
      expectedState.Balances.madByte.user -= burnQuantity;
      expectedState.Balances.eth.user += roundedEths;
      expectedState.Balances.madByte.poolBalance -= eths.toBigInt();
      expectedState.Balances.madByte.totalSupply -= burnQuantity;
      expectedState.Balances.eth.madByte -= eths.toBigInt();
      expect(await getState(fixture)).to.be.deep.equal(expectedState);
    }
  });

  it("Should burn it all", async () => {
    let burnQuantity = await fixture.madByte.totalSupply();
    expectedState = await getState(fixture);
    //Eths to receive
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
    let ethIn = ethers.utils.parseEther("8000");
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

  xit("Should keep ether sent value and (mint/burn) value relation greater or equal to market spread value with subsequent minting", async () => {
    // If some consequent mints are performed this relation is not kept between two first operations
    let ethIn = ethers.utils.parseEther("4");
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
    expect(ethIn.div(ethOut1).toNumber()).to.be.greaterThanOrEqual(
      marketSpread
    );
    expect(ethIn.div(ethOut2).toNumber()).to.be.greaterThanOrEqual(
      marketSpread
    );
    expect(ethIn.div(ethOut3).toNumber()).to.be.greaterThanOrEqual(
      marketSpread
    );
  });
});
