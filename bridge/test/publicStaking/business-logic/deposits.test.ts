import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { BaseTokensFixture, getBaseTokensFixture } from "../../setup";
import { assertTotalReserveAndZeroExcess } from "../setup";

describe("PublicStaking: Deposit Tokens and ETH", async () => {
  let fixture: BaseTokensFixture;
  let notAdminSigner: SignerWithAddress;
  let adminSigner: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner, notAdminSigner] = await ethers.getSigners();
    await fixture.madToken.approve(fixture.publicStaking.address, 100000);
  });
  it("Make successful deposits of tokens and ETH", async function () {
    const ethAmount = ethers.utils.parseEther("10").toBigInt();
    const tokenAmount = BigInt(100000);
    await fixture.publicStaking.depositToken(42, tokenAmount);
    await fixture.publicStaking.depositEth(42, { value: ethAmount });
    expect(
      (
        await fixture.madToken.balanceOf(fixture.publicStaking.address)
      ).toBigInt()
    ).to.be.equals(tokenAmount);
    expect(
      (
        await ethers.provider.getBalance(fixture.publicStaking.address)
      ).toBigInt()
    ).to.be.equals(ethAmount);
    assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      tokenAmount,
      ethAmount
    );
  });

  it("Should not allow deposits with wrong magic number", async function () {
    const ethAmount = ethers.utils.parseEther("10").toBigInt();
    const tokenAmount = BigInt(100000);
    await expect(
      fixture.publicStaking.depositToken(41, tokenAmount)
    ).to.be.revertedWith("BAD MAGIC");
    await expect(
      fixture.publicStaking.depositEth(41, { value: ethAmount })
    ).to.be.revertedWith("BAD MAGIC");
  });
});
