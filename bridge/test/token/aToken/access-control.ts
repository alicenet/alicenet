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
    describe("Testing Access Control operation", async () => {
      describe("Methods with onlyFactory modifier", async () => {
        it("Should not be able to set admin when not impersonating factory", async function () {
          await expect(
            fixture.aTokenMinter.connect(admin).setAdmin(user.address)
          ).to.be.revertedWith("onlyFactory");
        });
        it("Should be able set admin when impersonating factory", async function () {
          await factoryCallAny(fixture, "aTokenMinter", "setAdmin", [
            admin.address,
          ]);
          await fixture.aTokenMinter.connect(admin).setMinter(user.address);
        });
      });

      describe("Methods with onlyAdmin modifier", async () => {
        it("Should not be able to set minter if not admin", async function () {
          await expect(
            fixture.aTokenMinter.connect(admin).setMinter(user.address)
          ).to.be.revertedWith("onlyAdmin");
        });
        it("Should be able set minter if admin", async function () {
          await factoryCallAny(fixture, "aTokenMinter", "setAdmin", [
            admin.address,
          ]);
          await fixture.aTokenMinter.connect(admin).setMinter(user.address);
        });
      });
    });
  });
});
