import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getState, showState, state } from "./setup";

describe("Testing BToken Destroy methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let eths: BigNumber;
  let fixture: Fixture;
  const eth = 40;
  let ethForMinting: BigNumber;
  let bTokens: BigNumber;
  const minBTokens = 0;
  const marketSpread = 4;
  let ethsFromBurning: BigNumber;

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [admin, user] = signers;
    ethForMinting = ethers.utils.parseEther(eth.toString());
    [bTokens] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      user,
      [minBTokens],
      ethForMinting
    );
    ethsFromBurning = await fixture.bToken.getLatestEthFromBTokensBurn(bTokens);
    showState("Initial", await getState(fixture));
  });

  it("Should burn btokens from sender and keep resulting eth on contract", async () => {
    expectedState = await getState(fixture);
    await fixture.bToken.connect(user).destroyBTokens(bTokens);
    expect(bTokens).to.be.equal(BigInt("3990217121585928137263"));
    expectedState.Balances.bToken.user -= bTokens.toBigInt();
    expectedState.Balances.bToken.totalSupply -= bTokens.toBigInt();
    expectedState.Balances.bToken.poolBalance -= ethsFromBurning.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should burn btokens from address and keep resulting eth on contract", async () => {
    await fixture.bToken.connect(user).approve(admin.address, bTokens);
    expectedState = await getState(fixture);
    await fixture.bToken.destroyPreApprovedBTokens(user.address, bTokens);
    expect(bTokens).to.be.equal(BigInt("3990217121585928137263"));
    expectedState.Balances.bToken.user -= bTokens.toBigInt();
    expectedState.Balances.bToken.totalSupply -= bTokens.toBigInt();
    expectedState.Balances.bToken.poolBalance -= ethsFromBurning.toBigInt();
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should not burn btokens from address if not preapproved", async () => {
    await expect(
      fixture.bToken.destroyPreApprovedBTokens(user.address, bTokens)
    ).to.be.revertedWith("ERC20: insufficient allowance");
  });
});
