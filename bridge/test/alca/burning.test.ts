import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect, factoryCallAnyFixture, Fixture, getFixture } from "../setup";
import { getState, init, state } from "./setup";
describe("Testing ALCA", async () => {
  let user: SignerWithAddress;
  let admin: SignerWithAddress;
  let expectedState: state;
  let currentState: state;
  const amount = 1000n;
  let fixture: Fixture;

  async function deployFixture() {
    const fixture = await getFixture();
    const [admin, user] = await ethers.getSigners();
    await init(fixture);
    return { fixture, admin, user };
  }

  beforeEach(async function () {
    ({ fixture, admin, user } = await loadFixture(deployFixture));

    expectedState = await getState(fixture);
  });

  describe("Testing burning operation", async () => {
    describe("Methods with onlyALCABurner modifier", async () => {
      it("Should not burn when called by external address not identified as burner", async function () {
        await expect(fixture.alca.externalBurn(user.address, amount))
          .to.be.revertedWithCustomError(fixture.alca, `OnlyALCABurner`)
          .withArgs(admin.address, fixture.alcaBurner.address);

        await expect(
          fixture.alca.connect(admin).externalBurn(user.address, amount)
        )
          .to.be.revertedWithCustomError(fixture.alca, `OnlyALCABurner`)
          .withArgs(admin.address, fixture.alcaBurner.address);
      });
    });
    describe("Business methods with onlyFactory modifier", async () => {
      it("Should burn when called by external identified as burner impersonating factory", async function () {
        // migrate some tokens for burning
        await fixture.legacyToken
          .connect(user)
          .approve(fixture.alca.address, amount);
        await fixture.alca.connect(user).migrate(amount);
        expectedState = await getState(fixture);
        // burn
        await factoryCallAnyFixture(fixture, "alcaBurner", "burn", [
          user.address,
          amount,
        ]);
        expectedState.Balances.alca.user -= amount;
        currentState = await getState(fixture);
        expect(currentState).to.be.deep.eq(expectedState);
      });

      it("Should not allow to burn when called by external identified as burner not impersonating factory", async function () {
        // migrate some tokens for burning
        await fixture.legacyToken
          .connect(user)
          .approve(fixture.alca.address, amount);
        await fixture.alca.connect(user).migrate(amount);
        expectedState = await getState(fixture);
        // burn
        await expect(fixture.alcaBurner.burn(user.address, amount))
          .to.be.revertedWithCustomError(fixture.alcaBurner, `OnlyFactory`)
          .withArgs(admin.address, fixture.factory.address);
      });
    });
  });
});
