import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import {
  BaseTokensFixture,
  getBaseTokensFixture,
  getTokenIdFromTx,
  mineBlocks,
} from "../../setup";

describe("PublicStaking: Lock and LockWithdrawal", async () => {
  let fixture: BaseTokensFixture;
  let adminSigner: SignerWithAddress;
  let notAdminSigner: SignerWithAddress;
  let tokenID: number;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner, notAdminSigner] = await ethers.getSigners();
    await fixture.madToken.approve(fixture.publicStaking.address, 1000);
    const tx = await fixture.publicStaking.connect(adminSigner).mint(1000);
    tokenID = await getTokenIdFromTx(tx);
  });

  describe("PublicStaking: BurnLock a position", async () => {
    it("Lock own position, wait then burn position", async function () {
      await mineBlocks(3n);
      await fixture.publicStaking.lockOwnPosition(tokenID, 10n);
      await mineBlocks(10n);
      await fixture.publicStaking.connect(adminSigner).burn(tokenID);
    });

    it("Should be able to collect profits from a locked position", async function () {
      await mineBlocks(3n);
      await fixture.publicStaking.lockOwnPosition(tokenID, 10n);
      await fixture.publicStaking.collectEth(tokenID);
      await fixture.publicStaking.collectToken(tokenID);
    });

    it("Should not allow to burn a locked position without waiting", async function () {
      await mineBlocks(3n);
      await fixture.publicStaking.lockOwnPosition(tokenID, 10n);
      await expect(
        fixture.publicStaking.connect(adminSigner).burn(tokenID)
      ).to.revertedWith(
        "PublicStaking: The position is not ready to be burned!"
      );
    });

    it("Should not allow to lock a not owned position", async function () {
      await expect(
        fixture.publicStaking
          .connect(notAdminSigner)
          .lockOwnPosition(tokenID, 10n)
      ).to.revertedWith(
        "PublicStaking: Error, token doesn't exist or doesn't belong to the caller!"
      );
    });

    it("Should not allow to lock a position with a value greater than _MAX_MINT_LOCK", async function () {
      await mineBlocks(3n);
      await expect(
        fixture.publicStaking
          .connect(adminSigner)
          .lockOwnPosition(tokenID, 172801n)
      ).to.revertedWith(
        "PublicStaking: Lock Duration is greater than the amount allowed!"
      );
    });

    it("Should not be able to lock a non-existing position!", async function () {
      await mineBlocks(3n);
      await expect(
        fixture.publicStaking.connect(adminSigner).lockOwnPosition(1000, 10n)
      ).to.revertedWith("ERC721: owner query for nonexistent token");
    });
  });

  describe("PublicStaking: LockWithdrawal", async () => {
    it("Lock Withdrawal of a position, wait, then collect and burn position", async function () {
      await mineBlocks(3n);
      await fixture.publicStaking.lockWithdraw(tokenID, 10n);
      await mineBlocks(10n);
      await fixture.publicStaking.collectEth(tokenID);
      await fixture.publicStaking.collectToken(tokenID);
      await fixture.publicStaking.connect(adminSigner).burn(tokenID);
    });

    it("Should not allow to collect profits from a withdrawalLocked position", async function () {
      await mineBlocks(3n);
      await fixture.publicStaking.lockWithdraw(tokenID, 10n);
      await expect(
        fixture.publicStaking.collectEth(tokenID)
      ).to.be.rejectedWith("PublicStaking: Cannot withdraw at the moment.");
      await expect(
        fixture.publicStaking.collectToken(tokenID)
      ).to.be.rejectedWith("PublicStaking: Cannot withdraw at the moment.");
      await expect(
        fixture.publicStaking.collectEthTo(adminSigner.address, tokenID)
      ).to.be.rejectedWith("PublicStaking: Cannot withdraw at the moment.");
      await expect(
        fixture.publicStaking.collectTokenTo(adminSigner.address, tokenID)
      ).to.be.rejectedWith("PublicStaking: Cannot withdraw at the moment.");
    });

    it("Should not be able to withdrawalLock a non-existing position!", async function () {
      await mineBlocks(3n);
      await expect(
        fixture.publicStaking.connect(adminSigner).lockWithdraw(1000, 10n)
      ).to.revertedWith("ERC721: owner query for nonexistent token");
    });

    it("Should not allow to withdrawalLock a position with a value greater than _MAX_MINT_LOCK", async function () {
      await mineBlocks(3n);
      await expect(
        fixture.publicStaking
          .connect(adminSigner)
          .lockWithdraw(tokenID, 172801n)
      ).to.revertedWith(
        "PublicStaking: Lock Duration is greater than the amount allowed!"
      );
    });

    it("Should not allow to burn a locked position without waiting", async function () {
      await mineBlocks(3n);
      await fixture.publicStaking.lockWithdraw(tokenID, 10n);
      await expect(
        fixture.publicStaking.connect(adminSigner).burn(tokenID)
      ).to.revertedWith(
        "PublicStaking: The position is not ready to be burned!"
      );
    });

    it("Should not allow to lock a not owned position", async function () {
      await expect(
        fixture.publicStaking.connect(notAdminSigner).lockWithdraw(tokenID, 10n)
      ).to.revertedWith(
        "PublicStaking: Error, token doesn't exist or doesn't belong to the caller!"
      );
    });
  });
  describe("PublicStaking: BurnLock and withdrawalLock a position", async () => {
    it("Lock Own position and lockWithdrawal, wait then collect and burn position", async function () {
      await mineBlocks(3n);
      await fixture.publicStaking.lockOwnPosition(tokenID, 10n);
      await fixture.publicStaking.lockWithdraw(tokenID, 10n);
      await mineBlocks(10n);
      await fixture.publicStaking.connect(adminSigner).collectEth(tokenID);
      await fixture.publicStaking.connect(adminSigner).collectToken(tokenID);
      await fixture.publicStaking
        .connect(adminSigner)
        .collectEthTo(adminSigner.address, tokenID);
      await fixture.publicStaking
        .connect(adminSigner)
        .collectTokenTo(adminSigner.address, tokenID);
      await fixture.publicStaking.connect(adminSigner).burn(tokenID);
    });

    it("Should not allow to burn or collect profits from a locked position without waiting", async function () {
      await mineBlocks(3n);
      await fixture.publicStaking.lockOwnPosition(tokenID, 10n);
      await fixture.publicStaking.lockWithdraw(tokenID, 10n);
      await expect(
        fixture.publicStaking.collectEth(tokenID)
      ).to.be.rejectedWith("PublicStaking: Cannot withdraw at the moment.");
      await expect(
        fixture.publicStaking.collectToken(tokenID)
      ).to.be.rejectedWith("PublicStaking: Cannot withdraw at the moment.");
      await expect(
        fixture.publicStaking.collectEthTo(adminSigner.address, tokenID)
      ).to.be.rejectedWith("PublicStaking: Cannot withdraw at the moment.");
      await expect(
        fixture.publicStaking.collectTokenTo(adminSigner.address, tokenID)
      ).to.be.rejectedWith("PublicStaking: Cannot withdraw at the moment.");
      await expect(
        fixture.publicStaking.connect(adminSigner).burn(tokenID)
      ).to.revertedWith(
        "PublicStaking: The position is not ready to be burned!"
      );
    });
  });
});
