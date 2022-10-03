import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { factoryCallAnyFixture, Fixture, getFixture } from "../setup";

describe("ValidatorVault: Testing Access Control", async () => {
  let fixture: Fixture;
  let notAdminSigner: SignerWithAddress;

  async function deployFixture() {
    const fixture = await getFixture(true, true, false, true);
    const [, notAdmin] = fixture.namedSigners;
    const notAdminSigner = await ethers.getSigner(notAdmin.address);
    return {
      fixture,
      notAdminSigner,
    };
  }

  beforeEach(async function () {
    ({ fixture, notAdminSigner } = await loadFixture(deployFixture));
  });

  describe("A user/contract with the right role should be able to:", async () => {
    it("Skim excess of Eth", async function () {
      const rcpt = await factoryCallAnyFixture(
        fixture,
        "validatorVault",
        "skimExcessEth",
        [notAdminSigner.address]
      );
      expect(rcpt.status).to.equal(1);
    });

    it("Skim excess of Tokens", async function () {
      const rcpt = await factoryCallAnyFixture(
        fixture,
        "validatorVault",
        "skimExcessToken",
        [notAdminSigner.address]
      );
      expect(rcpt.status).to.equal(1);
    });
  });

  describe("A user/contract without the right role should not be able to:", async () => {
    it("call skimExcessEth", async function () {
      const validatorVault = fixture.validatorVault.connect(notAdminSigner);

      await expect(validatorVault.skimExcessEth(notAdminSigner.address))
        .to.revertedWithCustomError(validatorVault, "OnlyFactory")
        .withArgs(notAdminSigner.address, fixture.factory.address);
    });

    it("call skimExcessToken", async function () {
      const validatorVault = fixture.validatorVault.connect(notAdminSigner);

      await expect(validatorVault.skimExcessToken(notAdminSigner.address))
        .to.revertedWithCustomError(validatorVault, "OnlyFactory")
        .withArgs(notAdminSigner.address, fixture.factory.address);
    });

    it("call depositDilutionAdjustment", async function () {
      const validatorVault = fixture.validatorVault.connect(notAdminSigner);
      const adjustmentAmount = 1234;
      await expect(validatorVault.depositDilutionAdjustment(adjustmentAmount))
        .to.revertedWithCustomError(validatorVault, "OnlyATokenMinter")
        .withArgs(notAdminSigner.address, fixture.aTokenMinter.address);
    });

    it("call depositStake", async function () {
      const validatorVault = fixture.validatorVault.connect(notAdminSigner);
      const position = 4321;
      const amount = 1234;
      await expect(validatorVault.depositStake(position, amount))
        .to.revertedWithCustomError(validatorVault, "OnlyValidatorPool")
        .withArgs(notAdminSigner.address, fixture.validatorPool.address);
    });

    it("call withdrawStake", async function () {
      const validatorVault = fixture.validatorVault.connect(notAdminSigner);
      const position = 4321;

      await expect(validatorVault.withdrawStake(position))
        .to.revertedWithCustomError(validatorVault, "OnlyValidatorPool")
        .withArgs(notAdminSigner.address, fixture.validatorPool.address);
    });
  });
});
