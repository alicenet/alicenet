import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import {
  callFunctionAndGetReturnValues,
  Fixture,
  getFixture,
} from "../../setup";
import { format, getState, init, showState, state } from "./setup";

describe("Testing MadByte Minting methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let eths: BigNumber;
  let fixture: Fixture;
  let eth = 4;
  let mad: bigint;
  let ethIn: BigNumber;
  let minMadBytes = 0;
  const zeroAddress = "0x0000000000000000000000000000000000000000";
  const ONE_MB = 1 * 10 ** 18;

  beforeEach(async function () {
    fixture = await getFixture();
    let signers = await ethers.getSigners();
    [admin, user, user2] = signers;
    await init(fixture);
    showState("Initial", await getState(fixture));
    ethIn = ethers.utils.parseEther(eth.toString());
    mad = ethers.utils.parseUnits(eth.toString()).toBigInt();
  });

  it("Should mint madBytes to sender out of ether value sent", async () => {
    expectedState = await getState(fixture);
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      user,
      [minMadBytes],
      ethIn
    );
    //Eths
    let eths = await fixture.madByte.madByteToEth(
      await fixture.madByte.getPoolBalance(),
      await fixture.madByte.totalSupply(),
      madBytes
    );
    // Round eths to avoid consumed gas on state comparison
    let roundedEths = format(eths);
    expect(madBytes).to.be.equal(BigInt("399028731704364116575"));
    expectedState.Balances.eth.user -= eth;
    expectedState.Balances.madByte.user += madBytes.toBigInt();
    expectedState.Balances.madByte.totalSupply += madBytes.toBigInt();
    expectedState.Balances.madByte.poolBalance += eths.toBigInt();
    expectedState.Balances.eth.madByte += mad;
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
    expectedState = await getState(fixture);
    const [madBytes2] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      user2,
      [minMadBytes],
      ethIn
    );
    //Eths
    let eths2 = await fixture.madByte.madByteToEth(
      await fixture.madByte.getPoolBalance(),
      await fixture.madByte.totalSupply(),
      madBytes2
    );
    expectedState.Balances.eth.user2 -= eth;
    expectedState.Balances.madByte.user2 += madBytes2.toBigInt();
    expectedState.Balances.madByte.totalSupply += madBytes2.toBigInt();
    expectedState.Balances.madByte.poolBalance += eths2.toBigInt();
    expectedState.Balances.eth.madByte += mad;
    expect(madBytes2).to.be.equal(BigInt("399027176702820751481"));
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should mint with huge amount of eth", async () => {
    let eth = 70000000000;
    ethIn = ethers.utils.parseEther(eth.toString());
    expectedState = await getState(fixture);
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      admin,
      [minMadBytes],
      ethIn
    );
    expect(madBytes).to.be.equal(BigInt("17501004975246203818081563855"));
    expect(await fixture.madByte.getPoolBalance()).to.be.equal(
      BigInt("17500000000000000000000000000")
    );
    expect(await fixture.madByte.totalSupply()).to.be.equal(madBytes);
  });

  it("Should mint to an address", async () => {
    expectedState = await getState(fixture);
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mintTo",
      admin,
      [user.address, minMadBytes],
      ethIn
    );
    let eths = await fixture.madByte.madByteToEth(
      await fixture.madByte.getPoolBalance(),
      await fixture.madByte.totalSupply(),
      madBytes
    );
    expect(madBytes).to.be.equal(BigInt("399028731704364116575"));
    expectedState.Balances.eth.admin -= eth;
    expectedState.Balances.madByte.user += madBytes.toBigInt();
    expectedState.Balances.madByte.totalSupply += madBytes.toBigInt();
    expectedState.Balances.madByte.poolBalance += eths.toBigInt();
    expectedState.Balances.eth.madByte += mad;
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
    expectedState = await getState(fixture);
    const [madBytes2] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mintTo",
      admin,
      [user.address, minMadBytes],
      ethIn
    );
    eths = await fixture.madByte.madByteToEth(
      await fixture.madByte.getPoolBalance(),
      await fixture.madByte.totalSupply(),
      madBytes2
    );
    expectedState.Balances.eth.admin -= eth;
    expectedState.Balances.madByte.user += madBytes2.toBigInt();
    expectedState.Balances.madByte.totalSupply += madBytes2.toBigInt();
    expectedState.Balances.madByte.poolBalance += eths.toBigInt();
    expectedState.Balances.eth.madByte += mad;
    expect(madBytes2).to.be.equal(BigInt("399027176702820751481"));
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should mint a big amount of eth to an address", async () => {
    let eth = 70000000000;
    ethIn = ethers.utils.parseEther(eth.toString());
    mad = ethers.utils.parseUnits(eth.toString()).toBigInt();
    expectedState = await getState(fixture);
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mintTo",
      admin,
      [user.address, minMadBytes],
      ethIn
    );
    eths = await fixture.madByte.madByteToEth(
      await fixture.madByte.getPoolBalance(),
      await fixture.madByte.totalSupply(),
      madBytes
    );
    expect(madBytes).to.be.equal(BigInt("17501004975246203818081563855"));
    expect(await fixture.madByte.getPoolBalance()).to.be.equal(
      BigInt("17500000000000000000000000000")
    );
    expectedState.Balances.eth.admin -= eth;
    expectedState.Balances.madByte.user += madBytes.toBigInt();
    expectedState.Balances.madByte.totalSupply += madBytes.toBigInt();
    expectedState.Balances.madByte.poolBalance += eths.toBigInt();
    expectedState.Balances.eth.madByte += mad;
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to mint to 0x0 address", async () => {
    await expect(
      fixture.madByte.connect(admin).mintTo(zeroAddress, minMadBytes, {
        value: ethers.utils.parseEther(eth.toString()),
      })
    ).to.be.revertedWith("ERC20: mint to the zero address");
  });

  it("Should fail to mint with big min MadByte quantity", async () => {
    let minMadBytes = BigInt((900 * ONE_MB).toString());
    await expect(
      fixture.madByte.connect(admin).mint(minMadBytes, {
        value: ethers.utils.parseEther(eth.toString()),
      })
    ).to.be.revertedWith("MadByte: could not mint minimum MadBytes");
  });
});
