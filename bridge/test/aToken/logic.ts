import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers/lib/ethers";
import { ethers } from "hardhat";
import { MadToken } from "../../typechain-types";
import { AToken } from "../../typechain-types/AToken";
import { expect, factoryCallAny, Fixture, getFixture } from "../setup";

describe("Testing AToken", async () => {
  let aToken: AToken;
  let madToken: MadToken;
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let expectedState: state;
  let currentState: state;
  let amount = 1000;
  let fixture: Fixture;
  interface state {
    Balances: {
      madToken: {
        address: string;
        admin: number;
        user: number;
        aToken: number;
      };
      aToken: {
        address: string;
        admin: number;
        user: number;
        madToken: number;
      };
    };
  }

  beforeEach(async function () {
    fixture = await getFixture();
    const contractName = await fixture.aToken.name();
    [admin, user] = await ethers.getSigners();
    await fixture.madToken.connect(admin).approve(admin.address, 1000);
    await fixture.madToken
      .connect(admin)
      .transferFrom(admin.address, user.address, 1000);
    showState("Initial", await getState());
    expectedState = await getState();
  });

  describe("Testing AToken Immutable Version (deterministic)", async () => {
    it("Should migrate user legacy tokens", async function () {
      await fixture.madToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      expectedState.Balances.madToken.user -= amount;
      expectedState.Balances.aToken.user += amount;
      expectedState.Balances.madToken.aToken += amount;
      currentState = await getState();
      expect(currentState).to.be.deep.eq(expectedState);
    });

    describe("Testing Access Control with onlyFactory modifier", async () => {
      it("Should mint when called by external identified as minter impersonating factory", async function () {
        factoryCallAny(fixture, "aTokenMinter", "mint", [user.address, amount]);
        expectedState.Balances.aToken.user += amount;
        currentState = await getState();
        expect(currentState).to.be.deep.eq(expectedState);
      });

      it("Should not mint when called by external identified as minter not impersonating factory", async function () {
        await expect(
          fixture.aTokenMinter.mint(user.address, amount)
        ).to.be.revertedWith("onlyFactory");
      });
    });

    describe.only("Testing Access Control with onlyRole(MINTER_ROLE) modifier", async () => {
      it.only("Should mint when called by external identified as minter with onlyRole(MINTER_ROLE)", async function () {
        await factoryCallAny(fixture, "aTokenMinter", "setAdmin", [
          admin.address,
        ]);
        await fixture.aTokenMinter.connect(admin).setMinter(user.address);
        await fixture.aTokenMinter.connect(user).mint(user.address, amount);
        expectedState.Balances.aToken.user += amount;
        currentState = await getState();
        expect(currentState).to.be.deep.eq(expectedState);
      });

      it("Should not mint when called by external identified as minter with no admin role", async function () {
        let reason = getUserNotInRoleReason(
          user.address,
          await fixture.aTokenMinter.MINTER_ROLE()
        );
        await expect(
          fixture.aTokenMinter.connect(user).mint(user.address, amount)
        ).to.be.revertedWith(reason);
      });
    });

    it("Should not mint when called by external not identified as minter", async function () {
      await expect(
        fixture.aToken.externalMint(user.address, amount)
      ).to.be.revertedWith("onlyATokenMinter");
    });

    it("Should burn when called by external identified as burner", async function () {
      await fixture.madToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      expectedState = await getState();
      await fixture.aTokenBurner.burn(user.address, amount);
      expectedState.Balances.aToken.user -= amount;
      currentState = await getState();
      expect(currentState).to.be.deep.eq(expectedState);
    });

    it("Should not burn when called by external not identified as burner", async function () {
      await fixture.madToken
        .connect(user)
        .approve(fixture.aToken.address, amount);
      await fixture.aToken.connect(user).migrate(amount);
      await expect(
        fixture.aToken.connect(user).externalBurn(user.address, amount)
      ).to.be.revertedWith("onlyATokenBurner");
    });
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
      currentState = await getState();
      expect(currentState).to.be.deep.eq(expectedState);
    });

    it("Should mint when called by external identified as minter", async function () {
      await factoryCallAny(fixture, "aToken", "setMinter", [user.address]);
      await fixture.aToken.connect(user).externalMint(user.address, amount);
      expectedState.Balances.aToken.user += amount;
      currentState = await getState();
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
      expectedState = await getState();
      await factoryCallAny(fixture, "aToken", "setBurner", [user.address]);
      await fixture.aToken.connect(user).externalBurn(user.address, amount);
      expectedState.Balances.aToken.user -= amount;
      currentState = await getState();
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

  async function getState() {
    let state: state = {
      Balances: {
        madToken: {
          address: fixture.madToken.address.slice(-4),
          admin: format(await fixture.madToken.balanceOf(admin.address)),
          user: format(await fixture.madToken.balanceOf(user.address)),
          aToken: format(
            await fixture.madToken.balanceOf(fixture.aToken.address)
          ),
        },
        aToken: {
          address: fixture.aToken.address.slice(-4),
          admin: format(await fixture.aToken.balanceOf(admin.address)),
          user: format(await fixture.aToken.balanceOf(user.address)),
          madToken: format(
            await fixture.aToken.balanceOf(fixture.madToken.address)
          ),
        },
      },
    };
    return state;
  }

  function showState(title: string, state: state) {
    if (process.env.npm_config_detailed == "true") {
      // execute "npm --detailed=true test" to see this output
      console.log(title, state);
    }
  }

  function format(number: BigNumber) {
    return parseInt(ethers.utils.formatUnits(number, 0));
  }

  function getUserNotInRoleReason(address: string, role: string) {
    let reason =
      "AccessControl: account " +
      address.toLowerCase() +
      " is missing role " +
      role;
    return reason;
  }
});
