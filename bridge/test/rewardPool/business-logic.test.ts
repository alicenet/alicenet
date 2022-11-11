import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { BonusPool, Foundation, RewardPool } from "../../typechain-types";

import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import {
  calculateExpectedProportions,
  depositEthForStakingRewards,
  depositEthToAddress,
  depositTokensForStakingRewards,
  depositTokensToAddress,
  getImpersonatedSigner,
} from "./setup";

interface Fixture extends BaseTokensFixture {
  rewardPool: RewardPool;
  bonusPool: BonusPool;
  foundation: Foundation;
  lockupAddress: string;
  rewardPoolAddress: string;
  totalBonusAmount: BigNumber;
  mockFactorySigner: SignerWithAddress;
  mockLockupSigner: SignerWithAddress;
  mockBonusPoolSigner: SignerWithAddress;
  mockPublicStakingSigner: SignerWithAddress;
}

async function deployFixture() {
  await preFixtureSetup();
  const signers = await ethers.getSigners();
  const fixture = await deployFactoryAndBaseTokens(signers[0]);

  const foundation = (await deployUpgradeableWithFactory(
    fixture.factory,
    "Foundation",
    undefined
  )) as Foundation;

  await posFixtureSetup(fixture.factory, fixture.aToken);

  // get the address of the reward pool from the lockup contract
  const lockupAddress = signers[5].address;
  const aliceNetFactoryAddress = fixture.factory.address;

  const asFactory = await getImpersonatedSigner(fixture.factory.address);
  const asPublicStaking = await getImpersonatedSigner(
    fixture.publicStaking.address
  );
  await depositEthForStakingRewards(
    signers,
    fixture.publicStaking,
    ethers.utils.parseEther("1000000")
  );
  await depositTokensForStakingRewards(
    signers,
    fixture.aToken,
    fixture.publicStaking,
    ethers.utils.parseEther("1000000")
  );

  const asLockup = await getImpersonatedSigner(lockupAddress);
  const totalBonusAmount = ethers.utils.parseEther("1000000");
  // deploy reward pool
  const rewardPool = await (await ethers.getContractFactory("RewardPool"))
    .connect(asLockup)
    .deploy(fixture.aToken.address, aliceNetFactoryAddress, totalBonusAmount);

  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);

  const asBonusPool = await getImpersonatedSigner(bonusPoolAddress);

  return {
    fixture: {
      ...fixture,
      rewardPool,
      bonusPool,
      foundation,
      lockupAddress,
      rewardPoolAddress: rewardPool.address,
      totalBonusAmount,
      mockFactorySigner: asFactory,
      mockLockupSigner: asLockup,
      mockBonusPoolSigner: asBonusPool,
      mockPublicStakingSigner: asPublicStaking,
    },
    accounts: signers,
  };
}

describe("RewardPool - business logic", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];

  beforeEach(async () => {
    ({ fixture, accounts } = await loadFixture(deployFixture));
  });

  describe("deposit", async () => {
    it("Transfers eth and updates token reserve value", async () => {
      const tokenAmount = ethers.utils.parseEther("123456");
      const ethAmount = ethers.utils.parseEther("789");

      const ethBalanceBefore = await ethers.provider.getBalance(
        fixture.rewardPool.address
      );
      const ethReserveValueBefore = await fixture.rewardPool.getEthReserve();
      const tokenReserveValueBefore =
        await fixture.rewardPool.getTokenReserve();

      await depositEthToAddress(accounts[0], fixture.lockupAddress, ethAmount);
      await depositTokensToAddress(
        accounts[0],
        fixture.aToken,
        fixture.lockupAddress,
        tokenAmount
      );

      await fixture.rewardPool
        .connect(fixture.mockLockupSigner)
        .deposit(tokenAmount, {
          value: ethAmount,
        });

      const ethBalanceAfter = await ethers.provider.getBalance(
        fixture.rewardPool.address
      );
      const ethReserveAfter = await fixture.rewardPool.getEthReserve();
      const tokenReserveAfter = await fixture.rewardPool.getTokenReserve();

      expect(ethReserveAfter).to.equal(ethReserveValueBefore.add(ethAmount));
      expect(tokenReserveAfter).to.equal(
        tokenReserveValueBefore.add(tokenAmount)
      );
      expect(ethBalanceAfter).to.equal(ethBalanceBefore.add(ethAmount));
    });
  });

  describe("payout", async () => {
    const tokenAmount = ethers.utils.parseEther("10000000");
    const ethAmount = ethers.utils.parseEther("5000");

    beforeEach(async () => {
      // ensure reward pool has enough eth and tokens to pay out
      await depositTokensToAddress(
        accounts[0],
        fixture.aToken,
        fixture.rewardPoolAddress,
        tokenAmount
      );

      await depositEthToAddress(accounts[0], fixture.lockupAddress, ethAmount);
      await fixture.rewardPool
        .connect(fixture.mockLockupSigner)
        .deposit(tokenAmount, {
          value: ethAmount,
        });

      expect(await fixture.rewardPool.getEthReserve()).to.equal(ethAmount);
      expect(await fixture.rewardPool.getTokenReserve()).to.equal(tokenAmount);
    });

    it("Returns correct proportion values of tokens/eth when not last position", async () => {
      const totalShares = ethers.utils.parseEther("5000");
      const userShares = ethers.utils.parseEther("2500");
      const isLastPosition = false;

      const [expectedProportionEth, expectedProportionTokens] =
        calculateExpectedProportions(
          ethAmount,
          tokenAmount,
          userShares,
          totalShares
        );

      const [proportionalEth, proportionalTokens] = await fixture.rewardPool
        .connect(fixture.mockLockupSigner)
        .callStatic.payout(totalShares, userShares, isLastPosition);

      expect(proportionalEth).to.equal(expectedProportionEth);
      expect(proportionalTokens).to.equal(expectedProportionTokens);
    });

    it("Returns correct proportion values of tokens/eth when last position", async () => {
      const totalShares = ethers.utils.parseEther("5000");
      const userShares = ethers.utils.parseEther("2500");
      const isLastPosition = true;

      const expectedProportionEth = ethAmount;
      const expectedProportionTokens = tokenAmount;

      const [proportionalEth, proportionalTokens] = await fixture.rewardPool
        .connect(fixture.mockLockupSigner)
        .callStatic.payout(totalShares, userShares, isLastPosition);

      expect(proportionalEth).to.equal(expectedProportionEth);
      expect(proportionalTokens).to.equal(expectedProportionTokens);
    });

    it("Transfers proportion of tokens/eth to lockup if not last position", async () => {
      const totalShares = ethers.utils.parseEther("5000");
      const userShares = ethers.utils.parseEther("2500");
      const isLastPosition = false;

      const [expectedProportionEth, expectedProportionTokens] =
        calculateExpectedProportions(
          ethAmount,
          tokenAmount,
          userShares,
          totalShares
        );

      const ethBalanceBefore = await ethers.provider.getBalance(
        fixture.lockupAddress
      );
      const tokenBalanceBefore = await fixture.aToken.balanceOf(
        fixture.lockupAddress
      );

      const receipt = await (
        await fixture.rewardPool
          .connect(fixture.mockLockupSigner)
          .payout(totalShares, userShares, isLastPosition)
      ).wait();

      const ethBalanceAfter = await ethers.provider.getBalance(
        fixture.lockupAddress
      );
      const tokenBalanceAfter = await fixture.aToken.balanceOf(
        fixture.lockupAddress
      );

      expect(ethBalanceAfter).to.equal(
        ethBalanceBefore
          .add(expectedProportionEth)
          .sub(receipt.cumulativeGasUsed.mul(receipt.effectiveGasPrice))
      );
      expect(tokenBalanceAfter).to.equal(
        tokenBalanceBefore.add(expectedProportionTokens)
      );
    });

    it("Transfers balance of tokens/eth to lockup if last position", async () => {
      const totalShares = ethers.utils.parseEther("5000");
      const userShares = ethers.utils.parseEther("2500");
      const isLastPosition = true;

      const expectedProportionEth = ethAmount;
      const expectedProportionTokens = tokenAmount;

      const ethBalanceBefore = await ethers.provider.getBalance(
        fixture.lockupAddress
      );
      const tokenBalanceBefore = await fixture.aToken.balanceOf(
        fixture.lockupAddress
      );

      const receipt = await (
        await fixture.rewardPool
          .connect(fixture.mockLockupSigner)
          .payout(totalShares, userShares, isLastPosition)
      ).wait();

      const ethBalanceAfter = await ethers.provider.getBalance(
        fixture.lockupAddress
      );
      const tokenBalanceAfter = await fixture.aToken.balanceOf(
        fixture.lockupAddress
      );

      expect(ethBalanceAfter).to.equal(
        ethBalanceBefore
          .add(expectedProportionEth)
          .sub(receipt.cumulativeGasUsed.mul(receipt.effectiveGasPrice))
      );
      expect(tokenBalanceAfter).to.equal(
        tokenBalanceBefore.add(expectedProportionTokens)
      );
    });
  });
});
