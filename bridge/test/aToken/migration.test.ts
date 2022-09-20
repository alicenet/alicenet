import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect, Fixture, getFixture } from "../setup";
import { getState, init, state } from "./setup";

describe("Testing AToken", async () => {
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let currentState: state;
  const amount = 1000n;
  let fixture: Fixture;
  const scaleFactor = 10_000_000_000_000_000_000_000_000_000n;
  const multiplier = 15_555_555_555_555_555_555_555_555_555n;

  async function deployFixture() {
    const fixture = await getFixture();
    const [, user, user2] = await ethers.getSigners();
    await init(fixture);
    return { fixture, user, user2 };
  }

  beforeEach(async function () {
    ({ fixture, user, user2 } = await loadFixture(deployFixture));

    expectedState = await getState(fixture);
  });

  describe("Testing Migrate operation", async () => {
    it("Only factory should be allowed to call finishEarlyStage", async () => {
      await expect(fixture.aToken.connect(user2).finishEarlyStage())
        .to.revertedWithCustomError(fixture.aToken, "OnlyFactory")
        .withArgs(user2.address, fixture.factory.address);
    });

    it("Should be able to get legacy address", async () => {
      expect(await fixture.aToken.getLegacyTokenAddress()).to.be.equal(
        fixture.legacyToken.address
      );
    });

    it("Should migrate user legacy tokens with 1.555555555555555555 multiplier", async function () {
      await fixture.legacyToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      expectedState.Balances.legacyToken.user -= amount;
      const expectedAmount = (amount * multiplier) / scaleFactor;
      expect((await fixture.aToken.convert(amount)).toBigInt()).to.be.equal(
        expectedAmount
      );
      expectedState.Balances.aToken.user += expectedAmount;
      expectedState.Balances.legacyToken.aToken += amount;
      currentState = await getState(fixture);
      expect(currentState).to.be.deep.eq(expectedState);
    });

    it("Mint with and without multiplier", async () => {
      // user minting with multiplier in the earlier stage
      await fixture.legacyToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      expectedState.Balances.legacyToken.user -= amount;
      expectedState.Balances.aToken.user += (amount * multiplier) / scaleFactor;
      expectedState.Balances.legacyToken.aToken += amount;
      currentState = await getState(fixture);
      expect(currentState).to.be.deep.eq(expectedState);

      expectedState = await getState(fixture);
      await fixture.legacyToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      const finishEarlyStage =
        fixture.aToken.interface.encodeFunctionData("finishEarlyStage");
      // end up the earlier stage
      const txResponse = await fixture.factory.callAny(
        fixture.aToken.address,
        0,
        finishEarlyStage
      );
      await txResponse.wait();
      await fixture.aToken.connect(user).migrate(amount);
      expectedState.Balances.aToken.user += amount;
      expectedState.Balances.legacyToken.aToken += amount;
      expectedState.Balances.legacyToken.user -= amount;
      currentState = await getState(fixture);
      expect(currentState).to.be.deep.eq(expectedState);
    });

    it("Should toggle off multiplier and migrate user legacy token without multiplier", async () => {
      expectedState = await getState(fixture);
      await fixture.legacyToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      const finishEarlyStage =
        fixture.aToken.interface.encodeFunctionData("finishEarlyStage");
      const txResponse = await fixture.factory.callAny(
        fixture.aToken.address,
        0,
        finishEarlyStage
      );
      await txResponse.wait();
      await fixture.aToken.connect(user).migrate(amount);
      expectedState.Balances.aToken.user += amount;
      expectedState.Balances.legacyToken.aToken += amount;
      expectedState.Balances.legacyToken.user -= amount;
      currentState = await getState(fixture);
      expect(currentState).to.be.deep.eq(expectedState);
    });

    it("Should not allow migrate user legacy tokens without approval", async function () {
      await expect(
        fixture.aToken.connect(user).migrate(amount)
      ).to.be.revertedWith("ERC20: insufficient allowance");
    });

    it("Should not allow migrate user legacy tokens without token", async function () {
      await fixture.legacyToken
        .connect(user2)
        .approve(fixture.aToken.address, amount);
      await expect(
        fixture.aToken.connect(user2).migrate(amount)
      ).to.be.revertedWith("ERC20: transfer amount exceeds balance");
    });

    it("Should not allow migrate with 0 token", async function () {
      await fixture.legacyToken
        .connect(user2)
        .approve(fixture.aToken.address, amount);
      await expect(
        fixture.aToken.connect(user2).migrate(0)
      ).to.be.revertedWithCustomError(
        fixture.aToken,
        "InvalidConversionAmount"
      );
    });

    it("should convert the full amount of legacy", async () => {
      const cap = 220000000n;
      await fixture.legacyToken
        .connect(user)
        .approve(fixture.aToken.address, cap);
      await fixture.aToken.connect(user).migrate(cap);
      const expectedBalance = (cap * multiplier) / scaleFactor;
      const balance = await fixture.aToken.balanceOf(user.address);
      expect(balance).to.eq(expectedBalance);
    });
  });
});
