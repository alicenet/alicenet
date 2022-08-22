import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect, factoryCallAnyFixture, Fixture, getFixture } from "../setup";
import { getState, init, state } from "./setup";

describe("Testing AToken", async () => {
  let user: SignerWithAddress;
  let admin: SignerWithAddress;
  let expectedState: state;
  let currentState: state;
  const amount = 1000n;
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
          await expect(fixture.aToken.externalMint(user.address, amount))
            .to.be.revertedWithCustomError(fixture.aToken, `OnlyATokenMinter`)
            .withArgs(admin.address, fixture.aTokenMinter.address);

          await expect(
            fixture.aToken.connect(admin).externalMint(user.address, amount)
          )
            .to.be.revertedWithCustomError(fixture.aToken, `OnlyATokenMinter`)
            .withArgs(admin.address, fixture.aTokenMinter.address);
        });
      });

      describe("Business methods with onlyFactory modifier", async () => {
        it("Should mint when called by external identified as minter impersonating factory", async function () {
          await factoryCallAnyFixture(fixture, "aTokenMinter", "mint", [
            user.address,
            amount,
          ]);
          expectedState.Balances.aToken.user += amount;
          currentState = await getState(fixture);
          expect(currentState).to.be.deep.eq(expectedState);
        });

        it("Should not mint when called by external identified as minter not impersonating factory", async function () {
          await expect(fixture.aTokenMinter.mint(user.address, amount))
            .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
            .withArgs(admin.address, fixture.factory.address);
        });
      });
    });
  });
});
