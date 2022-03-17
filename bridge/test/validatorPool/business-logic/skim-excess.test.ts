import { Signer } from "ethers";
import { ethers } from "hardhat";
import { ValidatorPool } from "../../../typechain-types";
import { expect } from "../../chai-setup";
import {
  factoryCallAny,
  Fixture,
  getFixture,
  getValidatorEthAccount,
} from "../../setup";
import { validatorsSnapshots } from "../../snapshots/assets/4-validators-snapshots-1";
import { burnStakeTo } from "../setup";

describe("ValidatorPool: Skim excess of ETH and Tokens", async () => {
  let fixture: Fixture;
  let adminSigner: Signer;

  beforeEach(async () => {
    fixture = await getFixture(false, true, true);
    const [admin, , ,] = fixture.namedSigners;
    adminSigner = await getValidatorEthAccount(admin.address);
  });

  it("Factory should be able to skim excess of tokens eth sent to contract", async function () {
    let etherAmount = ethers.utils.parseEther("1");
    let madTokenAmount = ethers.utils.parseEther("2");
    let testAddress = ethers.Wallet.createRandom().address;

    await burnStakeTo(fixture, etherAmount, madTokenAmount, adminSigner);

    // Skimming the excess of eth
    await factoryCallAny(fixture, "validatorPool", "skimExcessEth", [
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

    // skim excess of madtokens
    await factoryCallAny(fixture, "validatorPool", "skimExcessToken", [
      testAddress,
    ]);
    expect(
      (await fixture.madToken.balanceOf(testAddress)).toBigInt()
    ).to.be.equal(
      madTokenAmount.toBigInt(),
      "Test address should have all madToken balance after skim excess token"
    );

    expect(
      (
        await fixture.madToken.balanceOf(fixture.validatorPool.address)
      ).toBigInt()
    ).to.be.equal(
      BigInt(0),
      "ValidatorPool should not have any madToken balance after skim excess token"
    );
  });

  it("Non authorized user should not be able to skim excess of tokens eth sent to contract", async function () {
    let validatorPool = fixture.validatorPool as ValidatorPool;
    await expect(
      validatorPool.skimExcessEth(validatorsSnapshots[0].address)
    ).to.be.rejectedWith("onlyFactory");
    await expect(
      validatorPool.skimExcessToken(validatorsSnapshots[0].address)
    ).to.be.rejectedWith("onlyFactory");
  });
});
