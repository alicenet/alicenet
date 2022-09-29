import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber, Signer } from "ethers";
import { ethers } from "hardhat";
import { ATokenMinterMock } from "../../typechain-types";
import { expect } from "../chai-setup";

import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getValidatorEthAccount,
  mineBlocks,
} from "../setup";
import { validatorsSnapshots } from "../snapshots/assets/4-validators-snapshots-1";
import { createValidators, stakeValidators } from "../validatorPool/setup";
describe("ValidatorVault: Testing Business Logic", async () => {
  let fixture: Fixture;
  let notAdminSigner: SignerWithAddress;
  let notAdminValidator: Signer;
  let stakeAmount: bigint;
  let validators: string[];
  let stakingTokenIds: BigNumber[];
  let adminSigner: SignerWithAddress;
  let adminValidator: Signer;

  describe("With no validators registered", async () => {
    async function deployFixture() {
      const fixture = await getFixture(false, true, false, true);
      const [, notAdmin] = fixture.namedSigners;
      const notAdminSigner = await ethers.getSigner(notAdmin.address);
      return {
        fixture,
        notAdminSigner,
      };
    }

    beforeEach(async function () {
      ({ fixture, notAdminSigner } = await loadFixture(deployFixture));
    });

    describe("Public functions", async () => {
      it("getAdjustmentPrice returns 0", async function () {
        const adjustmentAmount = 1234;
        const adjustmentPrice = await fixture.validatorVault.getAdjustmentPrice(
          adjustmentAmount
        );

        expect(adjustmentPrice).to.equal(0);
      });

      it("getAdjustmentPerUser returns 0", async function () {
        const adjustmentAmount = 1234;
        const [adjustmentPrice, numValidators] =
          await fixture.validatorVault.getAdjustmentPerUser(adjustmentAmount);

        expect(adjustmentPrice).to.equal(0);
        expect(numValidators).to.equal(0);
      });

      it("estimateStakedAmount returns 0", async function () {
        const stakePosition = 1;
        const stakeAmount = await fixture.validatorVault.estimateStakedAmount(
          stakePosition
        );

        expect(stakeAmount).to.equal(0);
      });
    });

    it("depositDilutionAdjustment fails if tokens not present", async function () {
      const adjustmentAmount = 5348;

      const aTokenMinterMock = fixture.aTokenMinter as ATokenMinterMock;
      await expect(
        aTokenMinterMock.depositDilutionAdjustment(adjustmentAmount)
      ).to.revertedWith("ERC20: transfer amount exceeds balance");
    });
  });

  describe("With validators registered", async () => {
    const adjustmentAmount = 5348;
    const userAdjustmentAmount = 1337; // adjustmentAmount / validators.length (4)

    async function deployFixture() {
      const fixture = await getFixture(false, true, false, true);
      const [admin, notAdmin, ,] = fixture.namedSigners;
      const adminSigner = await ethers.getSigner(admin.address);
      const adminValidator = await getValidatorEthAccount(admin.address);
      const notAdminValidator = await getValidatorEthAccount(notAdmin.address);
      const notAdminSigner = await ethers.getSigner(notAdmin.address);
      const validators = await createValidators(fixture, validatorsSnapshots);
      const stakingTokenIds = await stakeValidators(fixture, validators);
      const stakeAmount = (
        await fixture.validatorPool.getStakeAmount()
      ).toBigInt();
      await factoryCallAnyFixture(
        fixture,
        "validatorPool",
        "registerValidators",
        [validators, stakingTokenIds]
      );
      await mineBlocks(1n);

      const aTokenMinterMock = fixture.aTokenMinter as ATokenMinterMock;
      await aTokenMinterMock.mint(aTokenMinterMock.address, adjustmentAmount);

      return {
        fixture,
        stakeAmount,
        validators,
        stakingTokenIds,
        adminSigner,
        adminValidator,
        notAdminSigner,
        notAdminValidator,
      };
    }

    beforeEach(async function () {
      ({
        fixture,
        stakeAmount,
        validators,
        stakingTokenIds,
        adminSigner,
        adminValidator,
        notAdminSigner,
        notAdminValidator,
      } = await loadFixture(deployFixture));
    });

    describe("Token dilution", async () => {
      it("depositDilutionAdjustment succeeds if tokens present", async function () {
        const reserveBefore = await fixture.validatorVault.totalReserve();
        const globalAccumulatorBefore =
          await fixture.validatorVault.globalAccumulator();
        const stakeAmountBefore = await fixture.validatorPool.getStakeAmount();

        const aTokenMinterMock = fixture.aTokenMinter as ATokenMinterMock;

        const rcpt = await (
          await aTokenMinterMock.depositDilutionAdjustment(adjustmentAmount)
        ).wait();

        expect(rcpt.status).to.equal(1);

        const reserveAfter = await fixture.validatorVault.totalReserve();
        const globalAccumulatorAfter =
          await fixture.validatorVault.globalAccumulator();
        const stakeAmountAfter = await fixture.validatorPool.getStakeAmount();
        expect(reserveAfter).to.equal(reserveBefore.add(adjustmentAmount));
        expect(globalAccumulatorAfter).to.equal(
          globalAccumulatorBefore.add(userAdjustmentAmount)
        );
        expect(stakeAmountAfter).to.equal(
          stakeAmountBefore.add(userAdjustmentAmount)
        );
      });
    });

    describe("Public functions", async () => {
      it("getAdjustmentPerUser returns expected amount", async function () {
        const expectedValidatorCount = 4;
        const [adjustmentPrice, validatorCount] =
          await fixture.validatorVault.getAdjustmentPerUser(adjustmentAmount);

        expect(adjustmentPrice).to.equal(userAdjustmentAmount);
        expect(validatorCount).to.equal(expectedValidatorCount);
      });

      it("getAdjustmentPrice returns expected amount", async function () {
        const expectedAdjustmentPrice = adjustmentAmount;
        const adjustmentPrice = await fixture.validatorVault.getAdjustmentPrice(
          adjustmentAmount
        );

        expect(adjustmentPrice).to.equal(expectedAdjustmentPrice);
      });

      it("estimateStakedAmount returns expected amount", async function () {
        // expected amount is the original stake amount minus 1 wei due to it being staked in the validator pool
        const expectedStakedAmount = BigNumber.from(stakeAmount).sub(1);
        const stakedAmount = await fixture.validatorVault.estimateStakedAmount(
          stakingTokenIds[0]
        );

        expect(stakedAmount).to.equal(expectedStakedAmount);
      });
    });
  });
});
