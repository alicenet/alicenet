import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getEthConsumedAsGas, getState, showState, state } from "./setup";

describe("Testing BToken Destroy methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 40;
  let ethForMinting: BigNumber;
  let bTokens: BigNumber;
  const minBTokens = 0;
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
    const tx = await fixture.bToken.connect(user).destroyBTokens(bTokens);
    expect(bTokens).to.be.equal(BigInt("3990217121585928137263"));
    expectedState.Balances.bToken.user -= bTokens.toBigInt();
    expectedState.Balances.bToken.totalSupply -= bTokens.toBigInt();
    expectedState.Balances.bToken.poolBalance -= ethsFromBurning.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should not destroy more btokens than suply", async () => {
    await expect(
      fixture.bToken
        .connect(user)
        .destroyBTokens((await fixture.bToken.balanceOf(user.address)).add(100))
    )
      .to.be.revertedWithCustomError(fixture.bToken, "BurnAmountExceedsSupply")
      .withArgs(3990217121585928137363n, 3990217121585928137263n);
  });

  it("Should not destroy more BTokens than on balance", async () => {
    // another user mints
    await fixture.bToken
      .connect(admin)
      .mint(1n, { value: 1000000000000000000n });
    await expect(
      fixture.bToken
        .connect(user)
        .destroyBTokens((await fixture.bToken.balanceOf(user.address)).add(100))
    ).to.be.revertedWith("ERC20: burn amount exceeds balance");
  });

  it("Should not destroy 0 tokens", async () => {
    await expect(fixture.bToken.connect(user).destroyBTokens(0n))
      .to.be.revertedWithCustomError(fixture.bToken, `InvalidBurnAmount`)
      .withArgs(0);
  });
});
