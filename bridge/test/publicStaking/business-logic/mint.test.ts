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

describe("PublicStaking: Only Mint", async () => {
  let fixture: BaseTokensFixture;
  let notAdminSigner: SignerWithAddress;
  let adminSigner: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner, notAdminSigner] = await ethers.getSigners();
  });

  it("Mint a NFT position", async function () {
    await fixture.aToken.approve(fixture.publicStaking.address, 1000);
    const tx = await fixture.publicStaking.connect(adminSigner).mint(1000);
    const blockNumber = BigInt(tx.blockNumber as number);
    const tokenID = await getTokenIdFromTx(tx);
    expect(tokenID).to.be.equal(1);
    expect(await fixture.publicStaking.getLatestMintedPositionID()).to.be.equal(
      1
    );
    await assertPositions(
      fixture.publicStaking,
      tokenID,
      newPosition(1000n, blockNumber + 1n, blockNumber + 1n, 0n, 0n),
      adminSigner.address,
      1n,
      1000n,
      0n
    );
  });

  it("Mint many NFT positions for a user", async function () {
    await fixture.aToken.approve(fixture.publicStaking.address, 1000);
    for (let i = 1; i <= 10; i++) {
      const tx = await fixture.publicStaking.connect(adminSigner).mint(100);
      const blockNumber = BigInt(tx.blockNumber as number);
      const tokenID = await getTokenIdFromTx(tx);
      expect(tokenID).to.be.equal(i);
      expect(
        await fixture.publicStaking.getLatestMintedPositionID()
      ).to.be.equal(i);
      await assertPositions(
        fixture.publicStaking,
        tokenID,
        newPosition(100n, blockNumber + 1n, blockNumber + 1n, 0n, 0n),
        adminSigner.address
      );
    }
    expect(
      await fixture.publicStaking.balanceOf(adminSigner.address)
    ).to.be.equals(10);
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      BigInt(1000),
      BigInt(0)
    );
  });

  it("Should not be able to mint NFT positions without calling ERC20 approve before", async function () {
    await expect(
      fixture.publicStaking.connect(adminSigner).mint(1000)
    ).to.revertedWith("ERC20: insufficient allowance");
  });

  it("Should not be able to mintTo a NFT positions without calling ERC20 approve before", async function () {
    await expect(
      fixture.publicStaking
        .connect(adminSigner)
        .mintTo(notAdminSigner.address, 1000, 0)
    ).to.revertedWith("ERC20: insufficient allowance");
  });

  it("Should not be able to mint NFT positions without having tokens", async function () {
    await fixture.aToken
      .connect(notAdminSigner)
      .approve(fixture.publicStaking.address, 1000);
    await expect(
      fixture.publicStaking.connect(notAdminSigner).mint(1000)
    ).to.revertedWith("ERC20: transfer amount exceeds balance");
  });

  it("Should not be able to mint NFT position with more tokens than will ever exist", async function () {
    await expect(
      fixture.publicStaking.connect(adminSigner).mint(2n ** 224n)
    ).to.be.revertedWithCustomError(
      fixture.publicStaking,
      "MintAmountExceedsMaximumSupply"
    );
  });

  it("Should not be able to mintTo a NFT position with more tokens than will ever exist", async function () {
    await expect(
      fixture.publicStaking
        .connect(adminSigner)
        .mintTo(notAdminSigner.address, 2n ** 224n, 1)
    ).to.be.revertedWithCustomError(
      fixture.publicStaking,
      "MintAmountExceedsMaximumSupply"
    );
  });

  it("MintTo a NFT position to another user without lock", async function () {
    await fixture.aToken.approve(fixture.publicStaking.address, 1000);
    const tx = await fixture.publicStaking
      .connect(adminSigner)
      .mintTo(notAdminSigner.address, 1000, 0);
    const blockNumber = BigInt(tx.blockNumber as number);
    const tokenID = await getTokenIdFromTx(tx);
    expect(tokenID).to.be.equal(1);
    expect(await fixture.publicStaking.getLatestMintedPositionID()).to.be.equal(
      1
    );
    await assertPositions(
      fixture.publicStaking,
      tokenID,
      newPosition(1000n, blockNumber + 1n, blockNumber + 1n, 0n, 0n),
      notAdminSigner.address,
      1n,
      1000n,
      0n
    );
  });

  it("MintTo a NFT position to another user with time lock", async function () {
    await fixture.aToken.approve(fixture.publicStaking.address, 1000);
    const tx = await fixture.publicStaking
      .connect(adminSigner)
      .mintTo(notAdminSigner.address, 1000, 10);
    const blockNumber = BigInt(tx.blockNumber as number);
    const tokenID = await getTokenIdFromTx(tx);
    expect(tokenID).to.be.equal(1);
    expect(await fixture.publicStaking.getLatestMintedPositionID()).to.be.equal(
      1
    );
    const expectedPosition = newPosition(
      1000n,
      blockNumber + 10n,
      blockNumber + 1n,
      0n,
      0n
    );
    await assertPositions(
      fixture.publicStaking,
      tokenID,
      expectedPosition,
      notAdminSigner.address,
      1n,
      1000n,
      0n
    );
  });

  it("Should not be able to mintTo a NFT position with lock duration greater than _MAX_MINT_LOCK", async function () {
    await fixture.aToken.approve(fixture.publicStaking.address, 1000);
    await expect(
      fixture.publicStaking
        .connect(adminSigner)
        .mintTo(
          notAdminSigner.address,
          1000,
          (await fixture.publicStaking.getMaxMintLock()).toBigInt() + 1n
        )
    ).to.be.revertedWithCustomError(
      fixture.publicStaking,
      "LockDurationGreaterThanMintLock"
    );
  });
});
