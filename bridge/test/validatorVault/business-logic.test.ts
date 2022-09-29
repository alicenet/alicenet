import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { BigNumber } from "ethers";
import { ATokenMinterMock, ValidatorPoolMock } from "../../typechain-types";
import { expect } from "../chai-setup";

import {
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  mineBlocks,
} from "../setup";
import { validatorsSnapshots } from "../snapshots/assets/4-validators-snapshots-1";
import { createValidators, stakeValidators } from "../validatorPool/setup";
describe("ValidatorVault: Testing Business Logic", async () => {
  let fixture: Fixture;

  let stakeAmount: bigint;

  let stakingTokenIds: BigNumber[];

  describe("With no validators registered", async () => {
    async function deployFixture() {
      return await getFixture(true, true, false, true);
    }

    beforeEach(async function () {
      fixture = await loadFixture(deployFixture);
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

    describe("Stake management", async () => {
      it("depositStake should transfer tokens from validator pool to validator vault", async function () {
        const stakePosition = 1;
        const amount = 1000;

        const aTokenMinterMock = fixture.aTokenMinter as ATokenMinterMock;
        await aTokenMinterMock.mint(fixture.validatorPool.address, amount);

        const validatorVaultBalanceBefore = await fixture.aToken.balanceOf(
          fixture.validatorVault.address
        );
        const validatorPoolBalanceBefore = await fixture.aToken.balanceOf(
          fixture.validatorPool.address
        );
        const validatorPoolMock = fixture.validatorPool as ValidatorPoolMock;

        validatorPoolMock.depositStake(stakePosition, amount);

        const validatorVaultBalanceAfter = await fixture.aToken.balanceOf(
          fixture.validatorVault.address
        );
        const validatorPoolBalanceAfter = await fixture.aToken.balanceOf(
          fixture.validatorPool.address
        );
        const stakedAmount = await fixture.validatorVault.estimateStakedAmount(
          stakePosition
        );

        expect(stakedAmount).to.equal(amount);

        expect(validatorVaultBalanceAfter).to.equal(
          validatorVaultBalanceBefore.add(amount)
        );

        expect(validatorPoolBalanceAfter).to.equal(
          validatorPoolBalanceBefore.sub(amount)
        );
      });

      it("withdrawStake should transfer tokens from validator vault to validator pool", async function () {
        const stakePosition = 1;
        const amount = 1000;

        const aTokenMinterMock = fixture.aTokenMinter as ATokenMinterMock;
        await aTokenMinterMock.mint(fixture.validatorPool.address, amount);

        const validatorPoolMock = fixture.validatorPool as ValidatorPoolMock;
        validatorPoolMock.depositStake(stakePosition, amount);
        const validatorVaultBalanceBefore = await fixture.aToken.balanceOf(
          fixture.validatorVault.address
        );
        const validatorPoolBalanceBefore = await fixture.aToken.balanceOf(
          fixture.validatorPool.address
        );

        validatorPoolMock.withdrawStake(stakePosition);
        const validatorVaultBalanceAfter = await fixture.aToken.balanceOf(
          fixture.validatorVault.address
        );
        const validatorPoolBalanceAfter = await fixture.aToken.balanceOf(
          fixture.validatorPool.address
        );

        const stakedAmount = await fixture.validatorVault.estimateStakedAmount(
          stakePosition
        );

        expect(stakedAmount).to.equal(0);

        expect(validatorVaultBalanceAfter).to.equal(
          validatorVaultBalanceBefore.sub(amount)
        );

        expect(validatorPoolBalanceAfter).to.equal(
          validatorPoolBalanceBefore.add(amount)
        );
      });
    });
  });

  describe("With validators registered", async () => {
    const adjustmentAmount = 5348;
    const userAdjustmentAmount = 1337; // adjustmentAmount / validators.length (4)

    async function deployFixture() {
      const fixture = await getFixture(false, true, false, true);
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
        stakingTokenIds,
      };
    }

    beforeEach(async function () {
      ({ fixture, stakeAmount, stakingTokenIds } = await loadFixture(
        deployFixture
      ));
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
