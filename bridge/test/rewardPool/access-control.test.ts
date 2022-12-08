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

  await posFixtureSetup(fixture.factory, fixture.alca);

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
    fixture.alca,
    fixture.publicStaking,
    ethers.utils.parseEther("1000000")
  );
  const asLockup = await getImpersonatedSigner(lockupAddress);
  const totalBonusAmount = ethers.utils.parseEther("1000000");
  // deploy reward pool
  const rewardPool = await (await ethers.getContractFactory("RewardPool"))
    .connect(asLockup)
    .deploy(fixture.alca.address, aliceNetFactoryAddress, totalBonusAmount);

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

describe("RewardPool - access control", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];

  beforeEach(async () => {
    ({ fixture, accounts } = await loadFixture(deployFixture));
  });

  describe("deposit", async () => {
    it("Reverts if called from address that is not lockup or bonus", async () => {
      const tokenAmount = ethers.utils.parseEther("123456");
      const ethAmount = ethers.utils.parseEther("789");
      await expect(
        fixture.rewardPool.connect(accounts[1]).deposit(tokenAmount, {
          value: ethAmount,
        })
      ).to.be.revertedWithCustomError(
        fixture.rewardPool,
        "CallerNotLockupOrBonus"
      );
    });

    it("Succeeds if called from lockup", async () => {
      const tokenAmount = ethers.utils.parseEther("123456");
      const ethAmount = ethers.utils.parseEther("789");
      await depositEthToAddress(accounts[0], fixture.lockupAddress, ethAmount);
      await depositTokensToAddress(
        accounts[0],
        fixture.alca,
        fixture.lockupAddress,
        tokenAmount
      );
      await expect(
        fixture.rewardPool
          .connect(fixture.mockLockupSigner)
          .deposit(tokenAmount, {
            value: ethAmount,
          })
      ).to.not.be.reverted;
    });

    it("Succeeds if called from bonus pool", async () => {
      const tokenAmount = ethers.utils.parseEther("123456");
      const ethAmount = ethers.utils.parseEther("789");
      await depositEthToAddress(
        fixture.mockPublicStakingSigner,
        fixture.bonusPool.address,
        ethAmount
      );
      await depositTokensToAddress(
        fixture.mockPublicStakingSigner,
        fixture.alca,
        fixture.bonusPool.address,
        tokenAmount
      );
      await expect(
        fixture.rewardPool
          .connect(fixture.mockBonusPoolSigner)
          .deposit(tokenAmount, {
            value: ethAmount,
          })
      ).to.not.be.reverted;
    });
  });
  describe("payout", async () => {
    it("Reverts if called from address that is not lockup", async () => {
      const totalShares = ethers.utils.parseEther("1337");
      const userShares = ethers.utils.parseEther("1337");
      const isLastPosition = false;

      await expect(
        fixture.rewardPool
          .connect(accounts[1])
          .payout(totalShares, userShares, isLastPosition)
      ).to.be.revertedWithCustomError(fixture.rewardPool, "CallerNotLockup");
    });

    it("Reverts if total shares is 0", async () => {
      const totalShares = 0;
      const userShares = ethers.utils.parseEther("1337");
      const isLastPosition = false;

      await expect(
        fixture.rewardPool
          .connect(fixture.mockLockupSigner)
          .payout(totalShares, userShares, isLastPosition)
      ).to.be.revertedWithCustomError(
        fixture.rewardPool,
        "InvalidTotalSharesValue"
      );
    });

    it("Reverts if total shares is less than user shares", async () => {
      const totalShares = ethers.utils.parseEther("1336");
      const userShares = ethers.utils.parseEther("1337");
      const isLastPosition = false;

      await expect(
        fixture.rewardPool
          .connect(fixture.mockLockupSigner)
          .payout(totalShares, userShares, isLastPosition)
      ).to.be.revertedWithCustomError(
        fixture.rewardPool,
        "InvalidTotalSharesValue"
      );
    });
  });
});
