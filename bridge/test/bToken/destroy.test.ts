import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
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
  let bTokens: BigNumber;
  const minBTokens = 0;
  let ethsFromBurning: BigNumber;

  async function deployFixture() {
    const fixture = await getFixture();
    const signers = await ethers.getSigners();
    const [admin, user] = signers;
    const ethForMinting = ethers.utils.parseEther(eth.toString());
    const [bTokens] = await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      user,
      [minBTokens],
      ethForMinting
    );
    const ethsFromBurning = await fixture.bToken.getLatestEthFromBTokensBurn(
      bTokens
    );
    return { fixture, admin, user, ethForMinting, bTokens, ethsFromBurning };
  }

  beforeEach(async function () {
    ({ fixture, admin, user, bTokens, ethsFromBurning } = await loadFixture(
      deployFixture
    ));

    showState("Initial", await getState(fixture));
  });

  it("Should burn btokens from sender and keep resulting eth on contract", async () => {
    expectedState = await getState(fixture);
    const tx = await fixture.bToken.connect(user).destroyBTokens(bTokens);
    expect(bTokens).to.be.equal(BigInt("4020217121585928137263"));
    expectedState.Balances.bToken.user -= bTokens.toBigInt();
    expectedState.Balances.bToken.totalSupply -= bTokens.toBigInt();
    expectedState.Balances.bToken.poolBalance -= ethsFromBurning.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should not destroy more bTokens than supply", async () => {
    await expect(
      fixture.bToken
        .connect(user)
        .destroyBTokens((await fixture.bToken.balanceOf(user.address)).add(100))
    )
      .to.be.revertedWithCustomError(fixture.bToken, "BurnAmountExceedsSupply")
      .withArgs(4020217121585928137363n, 4020217121585928137263n);
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
