import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect, factoryCallAny, Fixture, getFixture } from "../../setup";
import { getState, init, state } from "./setup";

describe("Testing AToken", async () => {
  let user: SignerWithAddress;
  let expectedState: state;
  let currentState: state;
  const amount = 1000;
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture();
    [, user] = await ethers.getSigners();
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

      describe("Business methods with onlyFactory modifier", async () => {
        it("Should mint when called by external identified as minter impersonating factory", async function () {
          await factoryCallAny(fixture, "aTokenMinter", "mint", [
            user.address,
            amount,
          ]);
          expectedState.Balances.aToken.user += amount;
          currentState = await getState(fixture);
          expect(currentState).to.be.deep.eq(expectedState);
        });

        it("Should mint when called by external identified as minter not impersonating factory", async function () {
          await fixture.aTokenMinter.mint(user.address, amount);
          expectedState.Balances.aToken.user += amount;
          currentState = await getState(fixture);
          expect(currentState).to.be.deep.eq(expectedState);
        });
      });
    });
  });
});
