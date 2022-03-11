import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { MadToken } from "../../../typechain-types";
import { AToken } from "../../../typechain-types/AToken";
import { expect, factoryCallAny, Fixture, getFixture } from "../../setup";
import { getState, init, state } from "./setup";

describe("Testing AToken", async () => {
  let aToken: AToken;
  let madToken: MadToken;
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

  describe("Testing burning operation", async () => {
    describe("Methods with onlyATokenBurner modifier", async () => {
      it("Should burn when called by external address not identified as burner", async function () {
        await expect(
          fixture.aToken.externalBurn(user.address, amount)
        ).to.be.revertedWith("onlyATokenBurner");
      });
    });
    describe("Business methods with onlyFactory modifier", async () => {
      it("Should burn when called by external identified as burner impersonating factory", async function () {
        //migrate some tokens for burning
        await fixture.madToken
          .connect(user)
          .approve(fixture.aToken.address, amount);
        await fixture.aToken.connect(user).migrate(amount);
        expectedState = await getState(fixture);
        //burn
        await factoryCallAny(fixture, "aTokenBurner", "burn", [
          user.address,
          amount,
        ]);
        expectedState.Balances.aToken.user -= amount;
        currentState = await getState(fixture);
        expect(currentState).to.be.deep.eq(expectedState);
      });

      it.only("Should burn when called by external identified as burner not impersonating factory", async function () {
        //migrate some tokens for burning
        await fixture.madToken
          .connect(user)
          .approve(fixture.aToken.address, amount);
        await fixture.aToken.connect(user).migrate(amount);
        expectedState = await getState(fixture);
        //burn
        fixture.aTokenBurner.burn(user.address, amount);
        expectedState.Balances.aToken.user -= amount;
        currentState = await getState(fixture);
        expect(currentState).to.be.deep.eq(expectedState);
      });
    });
  });
});
