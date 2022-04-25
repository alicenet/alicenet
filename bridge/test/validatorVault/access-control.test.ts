import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumberish } from "ethers";
import { ethers } from "hardhat";
import { ValidatorPoolMock } from "../../typechain-types";
import { expect } from "../chai-setup";
import { factoryCallAnyFixture, Fixture, getFixture } from "../setup";

describe("ValidatorVault: Testing Access Control", async () => {
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
  });

  describe("A user/contract with the right role should be able to:", async () => {
    it("SkimExcess of Tokens", async function () {
      let rcpt = await factoryCallAnyFixture(
        fixture,
        "validatorVault",
        "skimExcessEth",
        [notAdminSigner.address]
      );
      expect(rcpt.status).to.equal(1);

      rcpt = await factoryCallAnyFixture(
        fixture,
        "validatorVault",
        "skimExcessToken",
        [notAdminSigner.address]
      );
      expect(rcpt.status).to.equal(1);
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
