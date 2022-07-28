import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { Signer } from "ethers";
import { ethers } from "hardhat";
import { ValidatorPool } from "../../../typechain-types";
import { expect } from "../../chai-setup";
import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getValidatorEthAccount,
} from "../../setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import { burnStakeTo } from "../setup";

describe("ValidatorPool: Skim excess of ETH and Tokens", async () => {
  let fixture: Fixture;
  let adminSigner: Signer;
  let admin: SignerWithAddress;
  beforeEach(async () => {
    fixture = await getFixture(false, true, true);
    [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
  });

  it("Factory should be able to skim excess of tokens eth sent to contract", async function () {
    const etherAmount = ethers.utils.parseEther("1");
    const aTokenAmount = ethers.utils.parseEther("2");
    const testAddress = ethers.Wallet.createRandom().address;

    await burnStakeTo(fixture, etherAmount, aTokenAmount, adminSigner);

    // Skimming the excess of eth
    await factoryCallAnyFixture(fixture, "validatorPool", "skimExcessEth", [
      testAddress,
    ]);
    expect(
      (await ethers.provider.getBalance(testAddress)).toBigInt()
    ).to.be.equal(
      etherAmount.toBigInt(),
      "Test address should have all eth balance after skim excess token"
    );
    expect(
      (
        await ethers.provider.getBalance(fixture.validatorPool.address)
      ).toBigInt()
    ).to.be.equal(
      BigInt(0),
      "ValidatorPool should not have any eth balance after skim excess token"
    );

    // skim excess of ATokens
    await factoryCallAnyFixture(fixture, "validatorPool", "skimExcessToken", [
      testAddress,
    ]);
    expect(
      (await fixture.aToken.balanceOf(testAddress)).toBigInt()
    ).to.be.equal(
      aTokenAmount.toBigInt(),
      "Test address should have all aToken balance after skim excess token"
    );

    expect(
      (await fixture.aToken.balanceOf(fixture.validatorPool.address)).toBigInt()
    ).to.be.equal(
      BigInt(0),
      "ValidatorPool should not have any aToken balance after skim excess token"
    );
  });

  it("Non authorized user should not be able to skim excess of tokens eth sent to contract", async function () {
    const validatorPool = fixture.validatorPool as ValidatorPool;
    await expect(validatorPool.skimExcessEth(validatorsSnapshots[0].address))
      .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
      .withArgs(admin.address);
    await expect(validatorPool.skimExcessToken(validatorsSnapshots[0].address))
      .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
      .withArgs(admin.address);
  });
});
