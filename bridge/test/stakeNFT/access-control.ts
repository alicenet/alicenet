import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  BaseTokensFixture,
  factoryCallAny,
  getBaseTokensFixture,
} from "../setup";

describe("StakeNFT: Testing StakeNFT Access Control", async () => {
  let fixture: BaseTokensFixture;
  let notAdminSigner: SignerWithAddress;
  let adminSigner: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner, notAdminSigner] = await ethers.getSigners();
  });

  describe("A user with admin role should be able to:", async () => {
    it("Trip circuit breaker", async function () {
      const rcpt = await factoryCallAny(fixture, "stakeNFT", "tripCB");
      expect(rcpt.status).to.be.equal(1);
      expect(await fixture.stakeNFT.circuitBreakerState()).to.be.equals(true);
    });

    it("Skim excess of Tokens and ETH", async function () {
      let rcpt = await factoryCallAny(fixture, "stakeNFT", "skimExcessEth", [
        adminSigner.address,
      ]);
      expect(rcpt.status).to.be.equal(1);
      rcpt = await factoryCallAny(fixture, "stakeNFT", "skimExcessToken", [
        adminSigner.address,
      ]);
      expect(rcpt.status).to.be.equal(1);
    });
  });
  describe("A user without admin role should not be able to:", async () => {
    it("Trip circuit breaker", async function () {
      await expect(fixture.stakeNFT.tripCB()).to.be.revertedWith("onlyFactory");
    });
    it("Skim excess of Tokens and ETH", async function () {
      await expect(
        fixture.stakeNFT.skimExcessEth(notAdminSigner.address)
      ).to.be.revertedWith("onlyFactory");
      await expect(
        fixture.stakeNFT.skimExcessToken(notAdminSigner.address)
      ).to.be.revertedWith("onlyFactory");
    });
  });
});
