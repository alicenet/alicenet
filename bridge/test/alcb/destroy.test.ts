import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { callFunctionAndGetReturnValues, Fixture, getFixture } from "../setup";
import { getEthConsumedAsGas, getState, showState, state } from "./setup";

describe("Testing ALCB Destroy methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let fixture: Fixture;
  const eth = 40;
  let alcbs: BigNumber;
  const minALCBs = 0;
  let ethsFromBurning: BigNumber;

  async function deployFixture() {
    const fixture = await getFixture();
    const signers = await ethers.getSigners();
    const [admin, user] = signers;
    const ethForMinting = ethers.utils.parseEther(eth.toString());
    const [alcbs] = await callFunctionAndGetReturnValues(
      fixture.alcb,
      "mint",
      user,
      [minALCBs],
      ethForMinting
    );
    const ethsFromBurning = await fixture.alcb.getLatestEthFromTokensBurn(
      alcbs
    );
    return { fixture, admin, user, ethForMinting, alcbs, ethsFromBurning };
  }

  beforeEach(async function () {
    ({ fixture, admin, user, alcbs, ethsFromBurning } = await loadFixture(
      deployFixture
    ));

    showState("Initial", await getState(fixture));
  });

  it("Should burn alcbs from sender and keep resulting eth on contract", async () => {
    expectedState = await getState(fixture);
    const tx = await fixture.alcb.connect(user).destroyTokens(alcbs);
    expect(alcbs).to.be.equal(BigInt("4020217121585928137263"));
    expectedState.Balances.alcb.user -= alcbs.toBigInt();
    expectedState.Balances.alcb.totalSupply -= alcbs.toBigInt();
    expectedState.Balances.alcb.poolBalance -= ethsFromBurning.toBigInt();
    expectedState.Balances.eth.user -= getEthConsumedAsGas(await tx.wait());
    expect(await getState(fixture)).to.be.deep.equal(expectedState);
  });

  it("Should not destroy more alcbs than supply", async () => {
    await expect(
      fixture.alcb
        .connect(user)
        .destroyTokens((await fixture.alcb.balanceOf(user.address)).add(100))
    )
      .to.be.revertedWithCustomError(fixture.alcb, "BurnAmountExceedsSupply")
      .withArgs(4020217121585928137363n, 4020217121585928137263n);
  });

  it("Should not destroy more ALCBs than on balance", async () => {
    // another user mints
    await fixture.alcb.connect(admin).mint(1n, { value: 1000000000000000000n });
    await expect(
      fixture.alcb
        .connect(user)
        .destroyTokens((await fixture.alcb.balanceOf(user.address)).add(100))
    ).to.be.revertedWith("ERC20: burn amount exceeds balance");
  });

  it("Should not destroy 0 tokens", async () => {
    await expect(fixture.alcb.connect(user).destroyTokens(0n))
      .to.be.revertedWithCustomError(fixture.alcb, `InvalidBurnAmount`)
      .withArgs(0);
  });
});
