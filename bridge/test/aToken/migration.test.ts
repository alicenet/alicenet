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

  beforeEach(async function () {
    fixture = await getFixture();
    [, user, user2] = await ethers.getSigners();
    await init(fixture);
    expectedState = await getState(fixture);
  });

  describe("Testing Migrate operation", async () => {
    it("Should migrate user legacy tokens", async function () {
      await fixture.legacyToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      expectedState.Balances.legacyToken.user -= amount;
      expectedState.Balances.aToken.user += amount;
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
      await expect(
        aToken.connect(admin).migrate(amount)
      ).to.be.revertedWithCustomError(aToken, "MigrationNotAllowed");
    });

    it("Should not allow migrate user legacy tokens without token", async function () {
      await fixture.legacyToken
        .connect(user2)
        .approve(fixture.aToken.address, amount);
      await expect(
        fixture.aToken.connect(user2).migrate(amount)
      ).to.be.revertedWith("ERC20: transfer amount exceeds balance");
    });
  });
});
