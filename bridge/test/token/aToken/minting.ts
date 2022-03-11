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

  describe("Testing minting operation", async () => {
    describe("Methods with onlyFactory modifier", async () => {
      describe("Methods with onlyATokenMinter modifier", async () => {
        it("Should not mint when called by external address not identified as minter", async function () {
          await expect(
            fixture.aToken.externalMint(user.address, amount)
          ).to.be.revertedWith("onlyATokenMinter");
        });
      });

      describe.skip("Business methods with onlyFactory modifier", async () => {
        it("Should mint when called by external identified as minter impersonating factory", async function () {
          factoryCallAny(fixture, "aTokenMinter", "mint", [
            user.address,
            amount,
          ]);
          expectedState.Balances.aToken.user += amount;
          currentState = await getState(fixture);
          expect(currentState).to.be.deep.eq(expectedState);
        });

        it("Should not mint when called by external identified as minter not impersonating factory", async function () {
          await expect(
            fixture.aTokenMinter.mint(user.address, amount)
          ).to.be.revertedWith("onlyFactory");
        });
      });

      describe("Testing Access Control with onlyATokenMinter modifier", async () => {
        it("Should not mint when called by external address not identified as minter", async function () {
          await expect(
            fixture.aToken.externalMint(user.address, amount)
          ).to.be.revertedWith("onlyATokenMinter");
        });
      });
    });
  });
});
