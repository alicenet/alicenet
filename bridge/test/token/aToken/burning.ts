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

  describe("Testing AToken Immutable Version (deterministic)", async () => {
    describe("Testing burning operation", async () => {
      describe("Methods with onlyBurner modifier", async () => {
        it("Should burn when called by external identified as burner and user in burner role", async function () {
          await fixture.madToken
            .connect(user)
            .approve(fixture.aToken.address, amount);
          await fixture.aToken.connect(user).migrate(amount);
          expectedState = await getState(fixture);
          await factoryCallAny(fixture, "aTokenBurner", "setAdmin", [
            admin.address,
          ]);
          await fixture.aTokenBurner.connect(admin).setBurner(user.address);
          await fixture.aTokenBurner.connect(user).burn(user.address, amount);
          expectedState.Balances.aToken.user -= amount;
          currentState = await getState(fixture);
          expect(currentState).to.be.deep.eq(expectedState);
        });

        it("Should not burn when called by external identified as burner with user not in burner role", async function () {
          await expect(
            fixture.aTokenBurner.connect(user).burn(user.address, amount)
          ).to.be.revertedWith("onlyBurner");
        });
      });

      describe("Methods with onlyAdmin modifier", async () => {
        it("Should not be able to set burner if not admin", async function () {
          await expect(
            fixture.aTokenBurner.connect(admin).setBurner(user.address)
          ).to.be.revertedWith("onlyAdmin");
        });
        it("Should be able set burner if admin", async function () {
          await factoryCallAny(fixture, "aTokenBurner", "setAdmin", [
            admin.address,
          ]);
          await fixture.aTokenBurner.connect(admin).setBurner(user.address);
        });
      });

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

      describe("Methods with onlyBurner modifier", async () => {
        it("Should burn when called by external identified as burner with user in burner role", async function () {
          await fixture.madToken
            .connect(user)
            .approve(fixture.aToken.address, amount);
          await fixture.aToken.connect(user).migrate(amount);
          expectedState = await getState(fixture);
          await factoryCallAny(fixture, "aTokenBurner", "setAdmin", [
            admin.address,
          ]);
          await fixture.aTokenBurner.connect(admin).setBurner(user.address);
          await fixture.aTokenBurner.connect(user).burn(user.address, amount);
          expectedState.Balances.aToken.user -= amount;
          currentState = await getState(fixture);
          expect(currentState).to.be.deep.eq(expectedState);
        });

        it("Should not mint when called by external identified as burner with user not in minter role", async function () {
          await expect(
            fixture.aTokenBurner.connect(user).burn(user.address, amount)
          ).to.be.revertedWith("onlyBurner");
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
});
