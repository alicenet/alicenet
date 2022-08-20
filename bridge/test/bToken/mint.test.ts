import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getEthConsumedAsGas, getState, showState, state } from "./setup";

describe("Testing BToken Minting methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 4;
  const marketSpread = 4;
  let ethInWeis: BigNumber;
  const minBTokens = 0;

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [admin, user, user2] = signers;
    showState("Initial", await getState(fixture));
    ethInWeis = ethers.utils.parseEther(eth.toString());
  });

  it("Should mint bTokens to sender out of ether value sent", async () => {
    expectedState = await getState(fixture);
    const [bTokens, tx] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      user,
      [minBTokens],
      ethInWeis
    );
    expect(bTokens).to.be.equal(BigInt("399028731704364116575"));
    expectedState.Balances.eth.user -= ethInWeis.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.bToken.user += bTokens.toBigInt();
    expectedState.Balances.bToken.totalSupply += bTokens.toBigInt();
    expectedState.Balances.bToken.poolBalance += ethInWeis
      .div(marketSpread)
      .toBigInt();
    expectedState.Balances.eth.bToken += ethInWeis.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(
      expectedState,
      "State 1 comp"
    );
    expectedState = await getState(fixture);
    const [bTokens2, tx2] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      user2,
      [minBTokens],
      ethInWeis
    );
    expectedState.Balances.eth.user2 -= ethInWeis.toBigInt();
    expectedState.Balances.eth.user2 -= getEthConsumedAsGas(await tx2.wait());
    expectedState.Balances.bToken.user2 += bTokens2.toBigInt();
    expectedState.Balances.bToken.totalSupply += bTokens2.toBigInt();
    expectedState.Balances.bToken.poolBalance += ethInWeis.div(4).toBigInt();
    expectedState.Balances.eth.bToken += ethInWeis.toBigInt();
    expect(bTokens2).to.be.equal(BigInt("399027176702820751481"));
    expect(await getState(fixture)).to.be.deep.equal(
      expectedState,
      "State 2 comp"
    );
  });

  it("Should mint with huge amount of eth", async () => {
    const eth = 70000000000;
    ethInWeis = ethers.utils.parseEther(eth.toString());
    expectedState = await getState(fixture);
    const [bTokens] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethInWeis
    );
    expect(bTokens).to.be.equal(BigInt("17501004975246203818081563855"));
    expect(await fixture.bToken.getPoolBalance()).to.be.equal(
      ethInWeis.div(marketSpread).toBigInt()
    );
    expect(await fixture.bToken.totalSupply()).to.be.equal(bTokens);
  });

  it("Should mint to an address", async () => {
    expectedState = await getState(fixture);
    const [bTokens, tx] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      admin,
      [user.address, minBTokens],
      ethInWeis
    );
    expect(bTokens).to.be.equal(BigInt("399028731704364116575"));
    expectedState.Balances.eth.admin -= ethInWeis.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.bToken.user += bTokens.toBigInt();
    expectedState.Balances.bToken.totalSupply += bTokens.toBigInt();
    expectedState.Balances.bToken.poolBalance += ethInWeis
      .div(marketSpread)
      .toBigInt();
    expectedState.Balances.eth.bToken += ethInWeis.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
    expectedState = await getState(fixture);
    const [bTokens2, tx2] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      admin,
      [user.address, minBTokens],
      ethInWeis
    );
    expectedState.Balances.eth.admin -= ethInWeis.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx2.wait());
    expectedState.Balances.bToken.user += bTokens2.toBigInt();
    expectedState.Balances.bToken.totalSupply += bTokens2.toBigInt();
    expectedState.Balances.bToken.poolBalance += ethInWeis
      .div(marketSpread)
      .toBigInt();
    expectedState.Balances.eth.bToken += ethInWeis.toBigInt();
    expect(bTokens2).to.be.equal(BigInt("399027176702820751481"));
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should mint a big amount of eth to an address", async () => {
    const eth = 70000000000;
    ethInWeis = ethers.utils.parseEther(eth.toString());
    expectedState = await getState(fixture);
    const [bTokens, tx] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mintTo",
      admin,
      [user.address, minBTokens],
      ethInWeis
    );
    expect(bTokens).to.be.equal(BigInt("17501004975246203818081563855"));
    expectedState.Balances.eth.admin -= ethInWeis.toBigInt();
    expectedState.Balances.eth.admin -= getEthConsumedAsGas(await tx.wait());
    expectedState.Balances.bToken.user += bTokens.toBigInt();
    expectedState.Balances.bToken.totalSupply += bTokens.toBigInt();
    expectedState.Balances.bToken.poolBalance += ethInWeis
      .div(marketSpread)
      .toBigInt();
    expectedState.Balances.eth.bToken += ethInWeis.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should fail to mint to 0x0 address", async () => {
    await expect(
      fixture.bToken
        .connect(admin)
        .mintTo(ethers.constants.AddressZero, minBTokens, {
          value: ethers.utils.parseEther(eth.toString()),
        })
    ).to.be.revertedWith("ERC20: mint to the zero address");
  });

  it("Should fail to mint with big min BToken quantity", async () => {
    const oneBToken = ethers.utils.parseUnits("1", 18).toBigInt();
    const minBTokens = 900n * oneBToken;
    const expectedBTokensMintedForEthValue = "399028731704364116575";
    await expect(
      fixture.bToken.connect(admin).mint(minBTokens, {
        value: ethers.utils.parseEther(eth.toString()),
      })
    )
      .to.be.revertedWithCustomError(fixture.bToken, `MinimumMintNotMet`)
      .withArgs(expectedBTokensMintedForEthValue, minBTokens);
  });
});
