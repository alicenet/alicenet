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
    describe.skip("Business methods with onlyFactory modifier", async () => {
      it("Should burn when called by external identified as burner impersonating factory", async function () {
        factoryCallAny(fixture, "aTokenBurn", "burn", [user.address, amount]);
        expectedState.Balances.aToken.user -= amount;
        currentState = await getState(fixture);
        expect(currentState).to.be.deep.eq(expectedState);
      });

      it("Should not burn when called by external identified as burner not impersonating factory", async function () {
        await expect(
          fixture.aTokenBurner.burn(user.address, amount)
        ).to.be.revertedWith("onlyFactory");
      });
    });

    describe("Methods with onlyATokenBurner modifier", async () => {
      it("Should not burn when called by external address not identified as burner", async function () {
        await expect(
          fixture.aToken.externalBurn(user.address, amount)
        ).to.be.revertedWith("onlyATokenBurner");
      });
    });
  });
});
