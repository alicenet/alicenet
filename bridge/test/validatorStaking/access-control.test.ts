import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumberish } from "ethers";
import { ethers } from "hardhat";
import { ValidatorPoolMock } from "../../typechain-types";
import { expect } from "../chai-setup";
import { Fixture, getFixture, mineBlocks } from "../setup";

describe("ValidatorStaking: Testing ValidatorStaking Access Control", async () => {
  let fixture: Fixture;
  let notAdminSigner: SignerWithAddress;
  const lockTime = 1;
  let amount: BigNumberish;
  let validatorPool: ValidatorPoolMock;

  beforeEach(async function () {
    fixture = await getFixture(true, true);
    const [, notAdmin] = fixture.namedSigners;
    notAdminSigner = await ethers.getSigner(notAdmin.address);
    validatorPool = fixture.validatorPool as ValidatorPoolMock;
    amount = await validatorPool.getStakeAmount();
    await fixture.aToken.approve(validatorPool.address, amount);
  });

  describe("A user with admin role should be able to:", async () => {
    it("Mint a token", async function () {
      const rcpt = await (await validatorPool.mintValidatorStaking()).wait();
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.validatorStaking.ownerOf(1)).to.be.eq(
        validatorPool.address
      );
    });

    it("Burn a token", async function () {
      await (await validatorPool.mintValidatorStaking()).wait();
      await mineBlocks(1n);
      const rcpt = await (await validatorPool.burnValidatorStaking(1)).wait();
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.aToken.balanceOf(validatorPool.address)).to.be.eq(
        amount
      );
    });

    it("Mint a token to an address", async function () {
      const rcpt = await (
        await validatorPool.mintToValidatorStaking(notAdminSigner.address)
      ).wait();
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.validatorStaking.ownerOf(1)).to.be.eq(
        notAdminSigner.address
      );
    });

    it("Burn a token from an address", async function () {
      await (await validatorPool.mintValidatorStaking()).wait();
      await mineBlocks(1n);
      const rcpt = await (
        await validatorPool.burnToValidatorStaking(1, notAdminSigner.address)
      ).wait();
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.aToken.balanceOf(notAdminSigner.address)).to.be.eq(
        amount
      );
    });
  });

  describe("A user without admin role should not be able to:", async function () {
    it("Mint a token", async function () {
      await expect(
        fixture.validatorStaking.connect(notAdminSigner).mint(amount)
      ).to.be.revertedWith("2010");
    });

    it("Burn a token", async function () {
      await expect(
        fixture.validatorStaking.connect(notAdminSigner).burn(42) // nonexistent
      ).to.be.revertedWith("2010");
    });

    it("Mint a token to an address", async function () {
      await expect(
        fixture.validatorStaking
          .connect(notAdminSigner)
          .mintTo(notAdminSigner.address, amount, lockTime)
      ).to.be.revertedWith("2010");
    });

    it("Burn a token from an address", async function () {
      await expect(
        fixture.validatorStaking
          .connect(notAdminSigner)
          .burnTo(notAdminSigner.address, 42) // nonexistent
      ).to.be.revertedWith("2010");
    });
  });
});
