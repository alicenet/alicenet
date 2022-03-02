import { Fixture, getFixture } from "../setup";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";

import { BigNumberish } from "ethers";
import { ValidatorPoolMock } from "../../typechain-types";

describe("ValidatorNFT: Testing ValidatorNFT Access Control", async () => {
  let fixture: Fixture;
  let notAdminSigner: SignerWithAddress;
  let lockTime = 1;
  let amount: BigNumberish;
  let validatorPool: ValidatorPoolMock;

  beforeEach(async function () {
    fixture = await getFixture(true, true);
    const [, notAdmin] = fixture.namedSigners;
    notAdminSigner = await ethers.getSigner(notAdmin.address);
    validatorPool = fixture.validatorPool as ValidatorPoolMock;
    amount = await validatorPool.getStakeAmount();
    await fixture.madToken.approve(validatorPool.address, amount);
  });

  describe("A user with admin role should be able to:", async () => {
    it("Mint a token", async function () {
      let rcpt = await (await validatorPool.mintValidatorNFT()).wait();
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.validatorNFT.ownerOf(1)).to.be.eq(
        validatorPool.address
      );
    });

    it("Burn a token", async function () {
      await (await validatorPool.mintValidatorNFT()).wait();
      let rcpt = await (await validatorPool.burnValidatorNFT(1)).wait();
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.madToken.balanceOf(validatorPool.address)).to.be.eq(
        amount
      );
    });

    it("Mint a token to an address", async function () {
      let rcpt = await (
        await validatorPool.mintToValidatorNFT(notAdminSigner.address)
      ).wait();
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.validatorNFT.ownerOf(1)).to.be.eq(
        notAdminSigner.address
      );
    });

    it("Burn a token from an address", async function () {
      await (await validatorPool.mintValidatorNFT()).wait();
      let rcpt = await (
        await validatorPool.burnToValidatorNFT(1, notAdminSigner.address)
      ).wait();
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.madToken.balanceOf(notAdminSigner.address)).to.be.eq(
        amount
      );
    });
  });

  describe("A user without admin role should not be able to:", async function () {
    it("Mint a token", async function () {
      await expect(
        fixture.validatorNFT.connect(notAdminSigner).mint(amount)
      ).to.be.revertedWith("onlyValidatorPool");
    });

    it("Burn a token", async function () {
      expect(
        fixture.validatorNFT.connect(notAdminSigner).burn(42) //nonexistent
      ).to.be.revertedWith("onlyValidatorPool");
    });

    it("Mint a token to an address", async function () {
      expect(
        fixture.validatorNFT
          .connect(notAdminSigner)
          .mintTo(notAdminSigner.address, amount, lockTime)
      ).to.be.revertedWith("onlyValidatorPool");
    });

    it("Burn a token from an address", async function () {
      expect(
        fixture.validatorNFT
          .connect(notAdminSigner)
          .burnTo(notAdminSigner.address, 42) //nonexistent
      ).to.be.revertedWith("onlyValidatorPool");
    });
  });
});
