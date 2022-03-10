import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers/lib/ethers";
import { ethers } from "hardhat";
import { MadToken } from "../../../typechain-types";
import { AToken } from "../../../typechain-types/AToken";
import { expect, factoryCallAny, Fixture, getFixture } from "../../setup";

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
    describe("Testing Migrate operation", async () => {
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
    });
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

    describe("Testing burning operation", async () => {
      describe("Methods with onlyBurner modifier", async () => {
        it("Should burn when called by external identified as burner and user in burner role", async function () {
          await fixture.madToken
            .connect(user)
            .approve(fixture.aToken.address, amount);
          await fixture.aToken.connect(user).migrate(amount);
          expectedState = await getState();
          await factoryCallAny(fixture, "aTokenBurner", "setAdmin", [
            admin.address,
          ]);
          await fixture.aTokenBurner.connect(admin).setBurner(user.address);
          await fixture.aTokenBurner.connect(user).burn(user.address, amount);
          expectedState.Balances.aToken.user -= amount;
          currentState = await getState();
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
          currentState = await getState();
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
          expectedState = await getState();
          await factoryCallAny(fixture, "aTokenBurner", "setAdmin", [
            admin.address,
          ]);
          await fixture.aTokenBurner.connect(admin).setBurner(user.address);
          await fixture.aTokenBurner.connect(user).burn(user.address, amount);
          expectedState.Balances.aToken.user -= amount;
          currentState = await getState();
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

    describe("Testing minting operation", async () => {
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

      describe("Methods with onlyATokenMinter modifier", async () => {
        it("Should not mint when called by external address not identified as minter", async function () {
          await expect(
            fixture.aToken.externalMint(user.address, amount)
          ).to.be.revertedWith("onlyATokenMinter");
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

      describe.skip("Business methods with onlyFactory modifier", async () => {
        it("Should mint when called by external identified as minter impersonating factory", async function () {
          factoryCallAny(fixture, "aTokenMinter", "mint", [
            user.address,
            amount,
          ]);
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

      describe("Methods with onlyMinter modifier", async () => {
        it("Should mint when called by external identified as minter with user in minter role", async function () {
          await factoryCallAny(fixture, "aTokenMinter", "setAdmin", [
            admin.address,
          ]);
          await fixture.aTokenMinter.connect(admin).setMinter(user.address);
          await fixture.aTokenMinter.connect(user).mint(user.address, amount);
          expectedState.Balances.aToken.user += amount;
          currentState = await getState();
          expect(currentState).to.be.deep.eq(expectedState);
        });

        it("Should not mint when called by external identified as minter with user not in minter role", async function () {
          await expect(
            fixture.aTokenMinter.connect(user).mint(user.address, amount)
          ).to.be.revertedWith("onlyMinter");
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
