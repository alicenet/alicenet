import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { HardhatEthersSigner } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import {
  BaseTokensFixture,
  factoryCallAnyFixture,
  getBaseTokensFixture,
} from "../setup";

describe("PublicStaking: Testing StakeNFT Access Control", async () => {
  let fixture: BaseTokensFixture;
  let notAdminSigner: HardhatEthersSigner;
  let adminSigner: HardhatEthersSigner;

  beforeEach(async function () {
    fixture = await loadFixture(getBaseTokensFixture);
    [adminSigner, notAdminSigner] = await ethers.getSigners();
  });

  describe("A user with admin role should be able to:", async () => {
    it("Trip circuit breaker", async function () {
      const rcpt = await factoryCallAnyFixture(
        fixture,
        "publicStaking",
        "tripCB"
      );
      expect(rcpt!.status).to.be.equal(1);
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
      expect(rcpt!.status).to.be.equal(1);
      rcpt = await factoryCallAnyFixture(
        fixture,
        "publicStaking",
        "skimExcessToken",
        [adminSigner.address]
      );
      expect(rcpt!.status).to.be.equal(1);
    });
  });
});
