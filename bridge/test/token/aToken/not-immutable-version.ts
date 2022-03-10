import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect, factoryCallAny, Fixture, getFixture } from "../../setup";
import { getState, init, state } from "./setup";

describe("Testing AToken", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let currentState: state;
  let amount = 1000;
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture();
    [admin, user] = await ethers.getSigners();
    await init(fixture);
    expectedState = await getState(fixture);
  });

  describe.skip("Testing AToken Not Immutable Version", async () => {
    it("Should migrate user legacy tokens", async function () {
      await fixture.madToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      expectedState.Balances.madToken.user -= amount;
      expectedState.Balances.aToken.user += amount;
      expectedState.Balances.madToken.aToken += amount;
      currentState = await getState(fixture);
      expect(currentState).to.be.deep.eq(expectedState);
    });

    it("Should mint when called by external identified as minter", async function () {
      await factoryCallAny(fixture, "aToken", "setMinter", [user.address]);
      await fixture.aToken.connect(user).externalMint(user.address, amount);
      expectedState.Balances.aToken.user += amount;
      currentState = await getState(fixture);
      expect(currentState).to.be.deep.eq(expectedState);
    });

    it("Should not mint when called by external not identified as minter", async function () {
      await expect(
        fixture.aToken.connect(user).externalMint(user.address, amount)
      ).to.be.revertedWith("AToken: onlyMinter role allowed");
    });

    it("Should burn when called by external identified as burner", async function () {
      await fixture.madToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      expectedState = await getState(fixture);
      await factoryCallAny(fixture, "aToken", "setBurner", [user.address]);
      await fixture.aToken.connect(user).externalBurn(user.address, amount);
      expectedState.Balances.aToken.user -= amount;
      currentState = await getState(fixture);
      expect(currentState).to.be.deep.eq(expectedState);
    });

    it("Should not burn when called by external not identified as burner", async function () {
      await fixture.madToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      await expect(
        fixture.aToken.connect(user).externalBurn(user.address, amount)
      ).to.be.revertedWith("AToken: onlyBurner role allowed");
    });
  });
});
