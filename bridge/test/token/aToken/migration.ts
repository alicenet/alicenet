import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { MadToken } from "../../../typechain-types";
import { AToken } from "../../../typechain-types/AToken";
import { expect, Fixture, getFixture } from "../../setup";
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

  describe("Testing Migrate operation", async () => {
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
  });
});
