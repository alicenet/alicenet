import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  BaseTokensFixture,
  factoryCallAnyFixture,
  getBaseTokensFixture,
} from "../setup";

describe("PublicStaking: Testing StakeNFT Access Control", async () => {
  let fixture: BaseTokensFixture;
  let notAdminSigner: SignerWithAddress;
  let adminSigner: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner, notAdminSigner] = await ethers.getSigners();
  });

  describe("A user with admin role should be able to:", async () => {
    it("Trip circuit breaker", async function () {
      const rcpt = await factoryCallAnyFixture(
        fixture,
        "publicStaking",
        "tripCB"
      );
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.publicStaking.circuitBreakerState()).to.be.equals(
        true
      );
    });

    it("Skim excess of Tokens and ETH", async function () {
      let rcpt = await factoryCallAnyFixture(
        fixture,
        "publicStaking",
        "skimExcessEth",
        [adminSigner.address]
      );
      expect(rcpt.status).to.be.equal(1);
      rcpt = await factoryCallAnyFixture(
        fixture,
        "publicStaking",
        "skimExcessToken",
        [adminSigner.address]
      );
      expect(rcpt.status).to.be.equal(1);
    });
  });
  describe("A user without admin role should not be able to:", async () => {
    it("Trip circuit breaker", async function () {
      await expect(fixture.publicStaking.tripCB())
        .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
        .withArgs(adminSigner.address);
    });
    it("Skim excess of Tokens and ETH", async function () {
      await expect(fixture.publicStaking.skimExcessEth(notAdminSigner.address))
        .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
        .withArgs(adminSigner.address);
      await expect(
        fixture.publicStaking.skimExcessToken(notAdminSigner.address)
      )
        .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
        .withArgs(adminSigner.address);
    });
  });
});
