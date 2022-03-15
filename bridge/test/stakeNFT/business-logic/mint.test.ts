import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import {
  BaseTokensFixture,
  getBaseTokensFixture,
  getTokenIdFromTx,
} from "../../setup";
import {
  assertPositions,
  assertTotalReserveAndZeroExcess,
  newPosition,
} from "../setup";

describe("StakeNFT: Mint and Burn", async () => {
  let fixture: BaseTokensFixture;
  let notAdminSigner: SignerWithAddress;
  let adminSigner: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner, notAdminSigner] = await ethers.getSigners();
  });

  it("Mint a NFT position", async function () {
    await fixture.madToken.approve(fixture.stakeNFT.address, 1000);
    const tx = await fixture.stakeNFT.connect(adminSigner).mint(1000);
    const tokenID = await getTokenIdFromTx(tx);
    await assertPositions(
      fixture.stakeNFT,
      tokenID,
      newPosition(1000n, 1n, 1n, 0n, 0n),
      adminSigner.address,
      1n,
      1000n,
      0n
    );
  });

  it("Mint many NFT positions for a user", async function () {
    await fixture.madToken.approve(fixture.stakeNFT.address, 1000);
    for (let i = 0; i < 10; i++) {
      const tx = await fixture.stakeNFT.connect(adminSigner).mint(100);
      const tokenID = await getTokenIdFromTx(tx);
      await assertPositions(
        fixture.stakeNFT,
        tokenID,
        newPosition(100n, 1n, 1n, 0n, 0n),
        adminSigner.address
      );
    }
    expect(await fixture.stakeNFT.balanceOf(adminSigner.address)).to.be.equals(
      10
    );
    await assertTotalReserveAndZeroExcess(
      fixture.stakeNFT,
      BigInt(1000),
      BigInt(0)
    );
  });

  it("Should not be able to mint NFT positions without calling ERC20 approve before", async function () {
    await expect(
      fixture.stakeNFT.connect(adminSigner).mint(1000)
    ).to.revertedWith("ERC20: insufficient allowance");
  });

  it("Should not be able to mint NFT position with more tokens than will ever exist", async function () {
    await expect(
      fixture.stakeNFT.connect(adminSigner).mint(2n ** 224n)
    ).to.revertedWith(
      "StakeNFT: The amount exceeds the maximum number of MadTokens that will ever exist!"
    );
  });

  it("MintTo a NFT position to another user", async function () {
    await fixture.madToken.approve(fixture.stakeNFT.address, 1000);
    const tx = await fixture.stakeNFT
      .connect(adminSigner)
      .mintTo(notAdminSigner.address, 1000, 0);
    const tokenID = await getTokenIdFromTx(tx);
    await assertPositions(
      fixture.stakeNFT,
      tokenID,
      newPosition(1000n, 1n, 1n, 0n, 0n),
      notAdminSigner.address,
      1n,
      1000n,
      0n
    );
  });
});
