import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import {
  BaseTokensFixture,
  factoryCallAny,
  getBaseTokensFixture,
} from "../../setup";

describe("PublicStaking: Call functions with Circuit Breaker tripped", async () => {
  let fixture: BaseTokensFixture;
  let notAdminSigner: SignerWithAddress;
  let adminSigner: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [adminSigner, notAdminSigner] = await ethers.getSigners();
    await fixture.madToken.approve(fixture.publicStaking.address, 1000);
    await fixture.publicStaking.connect(adminSigner).mint(1000);
    await factoryCallAny(fixture, "publicStaking", "tripCB");
  });

  describe("Users should not be able to:", async () => {
    it("Lock Position", async function () {
      await expect(
        fixture.publicStaking.lockPosition(adminSigner.address, 1, 1)
      ).to.be.rejectedWith("CircuitBreaker: The Circuit breaker is opened!");
    });
    it("Lock Own Position", async function () {
      await expect(
        fixture.publicStaking.lockOwnPosition(1, 1)
      ).to.be.rejectedWith("CircuitBreaker: The Circuit breaker is opened!");
    });
    it("Lock Withdraw", async function () {
      await expect(fixture.publicStaking.lockWithdraw(1, 1)).to.be.rejectedWith(
        "CircuitBreaker: The Circuit breaker is opened!"
      );
    });
    it("DepositToken", async function () {
      await expect(
        fixture.publicStaking.depositToken(42, 10)
      ).to.be.rejectedWith("CircuitBreaker: The Circuit breaker is opened!");
    });
    it("DepositEth", async function () {
      await expect(
        fixture.publicStaking.connect(adminSigner).depositEth(42, { value: 10 })
      ).to.be.rejectedWith("CircuitBreaker: The Circuit breaker is opened!");
    });
    it("Mint", async function () {
      await expect(
        fixture.publicStaking.connect(adminSigner).mint(100)
      ).to.be.rejectedWith("CircuitBreaker: The Circuit breaker is opened!");
    });
    it("MintTo", async function () {
      await expect(
        fixture.publicStaking
          .connect(adminSigner)
          .mintTo(notAdminSigner.address, 100, 1)
      ).to.be.rejectedWith("CircuitBreaker: The Circuit breaker is opened!");
    });
  });
  describe("Users should be able to:", async () => {
    it("Burn", async function () {
      await fixture.publicStaking.connect(adminSigner).burn(1);
    });
    it("BurnTo", async function () {
      await fixture.publicStaking
        .connect(adminSigner)
        .burnTo(notAdminSigner.address, 1);
    });
    it("Collect Eth and Tokens", async function () {
      await fixture.publicStaking.connect(adminSigner).collectEth(1);
      await fixture.publicStaking.connect(adminSigner).collectToken(1);
    });
    it("CollectTo Eth and Tokens ", async function () {
      await fixture.publicStaking
        .connect(adminSigner)
        .collectEthTo(notAdminSigner.address, 1);
      await fixture.publicStaking
        .connect(adminSigner)
        .collectTokenTo(notAdminSigner.address, 1);
    });
  });
});
