import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { AToken, LegacyToken } from "../../typechain-types";
import {
  deployAliceNetFactory,
  deployStaticWithFactory,
  expect,
  Fixture,
  getFixture,
} from "../setup";
import { getState, init, state } from "./setup";

describe("Testing AToken", async () => {
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let expectedState: state;
  let currentState: state;
  const amount = 1000;
  let fixture: Fixture;
  const multiplier = 1.5555555555556;
  beforeEach(async function () {
    fixture = await getFixture();
    [, user, user2] = await ethers.getSigners();
    await init(fixture);
    expectedState = await getState(fixture);
  });

  describe("Testing Migrate operation", async () => {
    it("Should migrate user legacy tokens with 1.555555555555555556 multiplier", async function () {
      await fixture.legacyToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      expectedState.Balances.legacyToken.user -= amount;
      expectedState.Balances.aToken.user += Math.floor(amount * multiplier);
      expectedState.Balances.legacyToken.aToken += amount;
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

    it("Should toggle on multiplier and migrate user legacy token without multiplier", async () => {
      expectedState = await getState(fixture);
      console.log(expectedState);
      await fixture.legacyToken
        .connect(user)
        .approve(fixture.aToken.address, 2 * amount);
      const finishEarlyStage =
        fixture.aToken.interface.encodeFunctionData("finishEarlyStage");
      let txResponse = await fixture.factory.callAny(
        fixture.aToken.address,
        0,
        finishEarlyStage
      );
      await txResponse.wait();
      txResponse = await fixture.aToken.connect(user).migrate(amount);
      await txResponse.wait();
      expectedState.Balances.aToken.user += amount;
      expectedState.Balances.legacyToken.aToken += amount;
      expectedState.Balances.legacyToken.user -= amount;
      currentState = await getState(fixture);
      expect(currentState).to.be.deep.eq(expectedState);
      expectedState = await getState(fixture);
      const toggleMultiplierOn =
        fixture.aToken.interface.encodeFunctionData("toggleMultiplierOn");
      txResponse = await fixture.factory.callAny(
        fixture.aToken.address,
        0,
        toggleMultiplierOn
      );
      await txResponse.wait();
      await fixture.aToken.connect(user).migrate(amount);
      expectedState.Balances.legacyToken.user -= amount;
      expectedState.Balances.aToken.user += Math.floor(amount * multiplier);
      expectedState.Balances.legacyToken.aToken += amount;
      currentState = await getState(fixture);
      expect(currentState).to.be.deep.eq(expectedState);
    });

    it("Should not allow migrate user legacy tokens without approval", async function () {
      await expect(
        fixture.aToken.connect(user).migrate(amount)
      ).to.be.revertedWith("ERC20: insufficient allowance");
    });

    it("Should not allow migrate if migration is locked", async function () {
      const [admin] = await ethers.getSigners();
      const factory = await deployAliceNetFactory(admin);

      // LegacyToken
      const legacyToken = (await deployStaticWithFactory(
        factory,
        "LegacyToken"
      )) as LegacyToken;

      const aToken = (await deployStaticWithFactory(
        factory,
        "AToken",
        "AToken",
        undefined,
        [legacyToken.address]
      )) as AToken;

      await factory.callAny(
        legacyToken.address,
        0,
        aToken.interface.encodeFunctionData("transfer", [
          admin.address,
          ethers.utils.parseEther("100000000"),
        ])
      );
      await expect(aToken.connect(admin).migrate(amount)).to.be.revertedWith(
        "MadTokens migration not allowed"
      );
    });

    it("Should not allow migrate user legacy tokens without token", async function () {
      await fixture.legacyToken
        .connect(user2)
        .approve(fixture.aToken.address, amount);
      await expect(
        fixture.aToken.connect(user2).migrate(amount)
      ).to.be.revertedWith("ERC20: transfer amount exceeds balance");
    });
    it("should convert the full amount of legacy", async () => {
      const cap = 220000000;
      await fixture.legacyToken
        .connect(user)
        .approve(fixture.aToken.address, cap);
      await fixture.aToken.connect(user).migrate(cap);
      const expectedBalance = Math.floor(cap * multiplier);
      const balance = await fixture.aToken.balanceOf(user.address);
      console.log(balance);
      console.log(expectedBalance);
      expect(balance).to.eq(expectedBalance);
    });
  });
});
