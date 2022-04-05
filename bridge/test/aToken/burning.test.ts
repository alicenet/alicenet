import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect, factoryCallAnyFixture, Fixture, getFixture } from "../setup";
import { getState, init, state } from "./setup";

describe("Testing AToken", async () => {
  let user: SignerWithAddress;
  let admin: SignerWithAddress;
  let expectedState: state;
  let currentState: state;
  const amount = 1000;
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture();
    [admin, user] = await ethers.getSigners();
    await init(fixture);
    expectedState = await getState(fixture);
  });

  describe("Testing burning operation", async () => {
    describe("Methods with onlyATokenBurner modifier", async () => {
      it("Should not burn when called by external address not identified as burner", async function () {
        await expect(
          fixture.aToken.externalBurn(user.address, amount)
        ).to.be.revertedWith("2012");

        await expect(
          fixture.aToken.connect(admin).externalBurn(user.address, amount)
        ).to.be.revertedWith("2012");
      });
    });
    describe("Business methods with onlyFactory modifier", async () => {
      it("Should burn when called by external identified as burner impersonating factory", async function () {
        // migrate some tokens for burning
        await fixture.legacyToken
          .connect(user)
          .approve(fixture.aToken.address, amount);
        await fixture.aToken.connect(user).migrate(amount);
        expectedState = await getState(fixture);
        // burn
        await factoryCallAnyFixture(fixture, "aTokenBurner", "burn", [
          user.address,
          amount,
        ]);
        expectedState.Balances.aToken.user -= amount;
        currentState = await getState(fixture);
        expect(currentState).to.be.deep.eq(expectedState);
      });

      it("Should not allow to burn when called by external identified as burner not impersonating factory", async function () {
        // migrate some tokens for burning
        await fixture.legacyToken
          .connect(user)
          .approve(fixture.aToken.address, amount);
        await fixture.aToken.connect(user).migrate(amount);
        expectedState = await getState(fixture);
        // burn
        await expect(
          fixture.aTokenBurner.burn(user.address, amount)
        ).to.be.revertedWith("2000");
      });
    });
  });
});
