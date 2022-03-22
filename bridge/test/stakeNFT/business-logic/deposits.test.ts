import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { BaseTokensFixture, getBaseTokensFixture } from "../../setup";
import { assertTotalReserveAndZeroExcess } from "../setup";

describe("StakeNFT: Deposit Tokens and ETH", async () => {
  let fixture: BaseTokensFixture;
  let notAdminSigner: SignerWithAddress;
  let adminSigner: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner, notAdminSigner] = await ethers.getSigners();
    await fixture.madToken.approve(fixture.stakeNFT.address, 100000);
  });
  it("Make successful deposits of tokens and ETH", async function () {
    let ethAmount = ethers.utils.parseEther("10").toBigInt();
    let tokenAmount = BigInt(100000);
    await fixture.stakeNFT.depositToken(42, tokenAmount);
    await fixture.stakeNFT.depositEth(42, { value: ethAmount });
    expect(
      (await fixture.madToken.balanceOf(fixture.stakeNFT.address)).toBigInt()
    ).to.be.equals(tokenAmount);
    expect(
      (await ethers.provider.getBalance(fixture.stakeNFT.address)).toBigInt()
    ).to.be.equals(ethAmount);
    assertTotalReserveAndZeroExcess(fixture.stakeNFT, tokenAmount, ethAmount);
  });

  it("Should not allow deposits with wrong magic number", async function () {
    let ethAmount = ethers.utils.parseEther("10").toBigInt();
    let tokenAmount = BigInt(100000);
    await expect(
      fixture.stakeNFT.depositToken(41, tokenAmount)
    ).to.be.revertedWith("BAD MAGIC");
    await expect(
      fixture.stakeNFT.depositEth(41, { value: ethAmount })
    ).to.be.revertedWith("BAD MAGIC");
  });
});
