import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import {
  expect,
  factoryCallAnyFixture,
  factoryCallAnyTX,
  Fixture,
  getFixture,
} from "../setup";
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
    [admin, user] = await ethers.getSigners();
    await init(fixture);
    return { fixture, admin, user };
  }

  beforeEach(async function () {
    ({ fixture, admin, user } = await loadFixture(deployFixture));

    expectedState = await getState(fixture);
  });

  describe("Testing minting operation", async () => {
    describe("Methods with onlyFactory modifier", async () => {
      describe("Methods with onlyALCAMinter modifier", async () => {
        it("Should not mint when called by external address not identified as minter", async function () {
          await expect(fixture.alca.externalMint(user.address, amount))
            .to.be.revertedWithCustomError(fixture.alca, `OnlyALCAMinter`)
            .withArgs(admin.address, fixture.alcaMinter.address);

          await expect(
            fixture.alca.connect(admin).externalMint(user.address, amount)
          )
            .to.be.revertedWithCustomError(fixture.alca, `OnlyALCAMinter`)
            .withArgs(admin.address, fixture.alcaMinter.address);
        });
      });
      describe("supply cap tests", async () => {
        it("should mint 1 billion tokens", async function () {
          await factoryCallAnyFixture(fixture, "alcaMinter", "mint", [
            user.address,
            1000000000n,
          ]);
          expectedState.Balances.alca.user += 1000000000n;
          currentState = await getState(fixture);
          expect(currentState).to.be.deep.eq(expectedState);
        });
        it("should mint 1 billion tokens and fail on 1 wei more", async function () {
          // get the current supply
          const maxSupply = ethers.utils.parseEther("1000000000");
          let currentSupply = await fixture.alca.totalSupply();
          const mintAmount = maxSupply.sub(currentSupply);
          await factoryCallAnyFixture(fixture, "alcaMinter", "mint", [
            user.address,
            mintAmount,
          ]);
          currentSupply = await fixture.alca.totalSupply();
          expect(currentSupply).to.be.eq(maxSupply);
          const tx = factoryCallAnyTX(
            fixture.factory,
            fixture.alcaMinter,
            "mint",
            [user.address, 1]
          );
          await expect(tx)
            .to.be.revertedWithCustomError(
              fixture.alcaMinter,
              "MintingExceeds1Billion"
            )
            .withArgs(currentSupply);
        });

        it("should not mint more than 1 billion tokens", async function () {
          const currentSupply = await fixture.alca.totalSupply();
          const tx = factoryCallAnyTX(
            fixture.factory,
            fixture.alcaMinter,
            "mint",
            [user.address, ethers.utils.parseEther("1000000001")]
          );
          await expect(tx)
            .to.be.revertedWithCustomError(
              fixture.alcaMinter,
              "MintingExceeds1Billion"
            )
            .withArgs(currentSupply);
        });
      });
      describe("Business methods with onlyFactory modifier", async () => {
        it("Should mint when called by external identified as minter impersonating factory", async function () {
          await factoryCallAnyFixture(fixture, "alcaMinter", "mint", [
            user.address,
            amount,
          ]);
          expectedState.Balances.alca.user += amount;
          currentState = await getState(fixture);
          expect(currentState).to.be.deep.eq(expectedState);
        });

        it("Should not mint when called by external identified as minter not impersonating factory", async function () {
          await expect(fixture.alcaMinter.mint(user.address, amount))
            .to.be.revertedWithCustomError(fixture.alcb, `OnlyFactory`)
            .withArgs(admin.address, fixture.factory.address);
        });
      });
    });
  });
});
