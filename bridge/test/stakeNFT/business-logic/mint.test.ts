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

describe("StakeNFT: Only Mint", async () => {
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
    const blockNumber = BigInt(tx.blockNumber as number);
    const tokenID = await getTokenIdFromTx(tx);
    await assertPositions(
      fixture.stakeNFT,
      tokenID,
      newPosition(1000n, blockNumber + 1n, blockNumber + 1n, 0n, 0n),
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
      const blockNumber = BigInt(tx.blockNumber as number);
      const tokenID = await getTokenIdFromTx(tx);
      await assertPositions(
        fixture.stakeNFT,
        tokenID,
        newPosition(100n, blockNumber + 1n, blockNumber + 1n, 0n, 0n),
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

  it("Should not be able to mintTo a NFT positions without calling ERC20 approve before", async function () {
    await expect(
      fixture.stakeNFT
        .connect(adminSigner)
        .mintTo(notAdminSigner.address, 1000, 0)
    ).to.revertedWith("ERC20: insufficient allowance");
  });

  it("Should not be able to mint NFT positions without having tokens", async function () {
    await fixture.madToken
      .connect(notAdminSigner)
      .approve(fixture.stakeNFT.address, 1000);
    await expect(
      fixture.stakeNFT.connect(notAdminSigner).mint(1000)
    ).to.revertedWith("ERC20: transfer amount exceeds balance");
  });

  it("Should not be able to mint NFT position with more tokens than will ever exist", async function () {
    await expect(
      fixture.stakeNFT.connect(adminSigner).mint(2n ** 224n)
    ).to.revertedWith(
      "StakeNFT: The amount exceeds the maximum number of MadTokens that will ever exist!"
    );
  });

  it("Should not be able to mintTo a NFT position with more tokens than will ever exist", async function () {
    await expect(
      fixture.stakeNFT
        .connect(adminSigner)
        .mintTo(notAdminSigner.address, 2n ** 224n, 1)
    ).to.revertedWith(
      "StakeNFT: The amount exceeds the maximum number of MadTokens that will ever exist!"
    );
  });

  it("MintTo a NFT position to another user without lock", async function () {
    await fixture.madToken.approve(fixture.stakeNFT.address, 1000);
    const tx = await fixture.stakeNFT
      .connect(adminSigner)
      .mintTo(notAdminSigner.address, 1000, 0);
    const blockNumber = BigInt(tx.blockNumber as number);
    const tokenID = await getTokenIdFromTx(tx);
    await assertPositions(
      fixture.stakeNFT,
      tokenID,
      newPosition(1000n, blockNumber + 1n, blockNumber + 1n, 0n, 0n),
      notAdminSigner.address,
      1n,
      1000n,
      0n
    );
  });

  it("MintTo a NFT position to another user with time lock", async function () {
    await fixture.madToken.approve(fixture.stakeNFT.address, 1000);
    const tx = await fixture.stakeNFT
      .connect(adminSigner)
      .mintTo(notAdminSigner.address, 1000, 10);
    const blockNumber = BigInt(tx.blockNumber as number);
    const tokenID = await getTokenIdFromTx(tx);
    const expectedPosition = newPosition(
      1000n,
      blockNumber + 10n,
      blockNumber + 1n,
      0n,
      0n
    );
    await assertPositions(
      fixture.stakeNFT,
      tokenID,
      expectedPosition,
      notAdminSigner.address,
      1n,
      1000n,
      0n
    );
  });

  it("Should not be able to mintTo a NFT position with lock duration greater than _MAX_MINT_LOCK", async function () {
    await fixture.madToken.approve(fixture.stakeNFT.address, 1000);
    await expect(
      fixture.stakeNFT
        .connect(adminSigner)
        .mintTo(
          notAdminSigner.address,
          1000,
          (await fixture.stakeNFT.getMaxMintLock()).toBigInt() + 1n
        )
    ).to.revertedWith(
      "StakeNFT: The lock duration must be less or equal than the maxMintLock!"
    );
  });
});
