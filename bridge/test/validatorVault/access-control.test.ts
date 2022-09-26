import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { factoryCallAnyFixture, Fixture, getFixture } from "../setup";

describe("ValidatorVault: Testing Access Control", async () => {
  let fixture: Fixture;
  let notAdminSigner: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getFixture(true, true);
    const [, notAdmin] = fixture.namedSigners;
    notAdminSigner = await ethers.getSigner(notAdmin.address);
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
});
