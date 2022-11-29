import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { BonusPool, Foundation } from "../../typechain-types";

import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import {
  calculateTerminationProfits,
  depositEthForStakingRewards,
  depositEthToAddress,
  depositTokensForStakingRewards,
  ensureBlockIsAtLeast,
  getImpersonatedSigner,
  mintBonusPosition,
} from "./setup";

interface Fixture extends BaseTokensFixture {
  bonusPool: BonusPool;
  foundation: Foundation;
  lockupAddress: string;
  rewardPoolAddress: string;
  totalBonusAmount: BigNumber;
  mockFactorySigner: SignerWithAddress;
  mockLockupSigner: SignerWithAddress;
  mockPublicStaking: SignerWithAddress;
  alcaRewards: BigNumber;
  ethRewards: BigNumber;
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
  const asLockup = await getImpersonatedSigner(lockupAddress);

  const totalBonusAmount = ethers.utils.parseEther("1000000");
  // deploy reward pool
  const rewardPool = await (await ethers.getContractFactory("RewardPool"))
    .connect(asLockup)
    .deploy(fixture.alca.address, aliceNetFactoryAddress, totalBonusAmount);
  // Deploy the bonus pool standalone
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);

  const alcaRewards = ethers.utils.parseEther("1000000");
  const ethRewards = ethers.utils.parseEther("10");

  await depositEthForStakingRewards(signers, fixture.publicStaking, ethRewards);
  await depositTokensForStakingRewards(
    signers,
    fixture.alca,
    fixture.publicStaking,
    alcaRewards
  );

  return {
    fixture: {
      ...fixture,
      bonusPool,
      foundation,
      lockupAddress,
      rewardPoolAddress: rewardPool.address,
      totalBonusAmount,
      mockFactorySigner: asFactory,
      mockLockupSigner: asLockup,
      mockPublicStaking: asPublicStaking,
      alcaRewards,
      ethRewards,
    },
    accounts: signers,
  };
}

describe("BonusPool", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];

  beforeEach(async () => {
    ({ fixture, accounts } = await loadFixture(deployFixture));
  });

  describe("createBonusStakedPosition", async () => {
    it("Reverts if insufficient ALCA to stake", async () => {
      const shortStakeAmount = fixture.totalBonusAmount.sub(1);
      await (
        await fixture.alca
          .connect(accounts[0])
          .transfer(fixture.bonusPool.address, shortStakeAmount)
      ).wait();

      await expect(
        fixture.bonusPool
          .connect(fixture.mockFactorySigner)
          .createBonusStakedPosition()
      )
        .to.be.revertedWithCustomError(
          fixture.bonusPool,
          "NotEnoughALCAToStake"
        )
        .withArgs(shortStakeAmount, fixture.totalBonusAmount);
    });
    it("Succeeds if called from factory address and has enough ALCA", async () => {
      const exactStakeAmount = fixture.totalBonusAmount;
      await (
        await fixture.alca
          .connect(accounts[0])
          .transfer(fixture.bonusPool.address, exactStakeAmount)
      ).wait();

      const expectedTokenId = 1;
      await expect(
        fixture.bonusPool
          .connect(fixture.mockFactorySigner)
          .createBonusStakedPosition()
      )
        .to.emit(fixture.bonusPool, "BonusPositionCreated")
        .withArgs(expectedTokenId);
    });

    it("Reverts if bonus position already created", async () => {
      const exactStakeAmount = fixture.totalBonusAmount;
      await (
        await fixture.alca
          .connect(accounts[0])
          .transfer(fixture.bonusPool.address, exactStakeAmount)
      ).wait();

      await (
        await fixture.bonusPool
          .connect(fixture.mockFactorySigner)
          .createBonusStakedPosition()
      ).wait();

      await expect(
        fixture.bonusPool
          .connect(fixture.mockFactorySigner)
          .createBonusStakedPosition()
      ).to.be.revertedWithCustomError(
        fixture.bonusPool,
        "BonusTokenAlreadyCreated"
      );
    });
  });

  describe("receive", async () => {
    it("Reverts if sent from non public staking address", async () => {
      const ethAmount = BigNumber.from(1234);

      await expect(
        depositEthToAddress(accounts[0], fixture.bonusPool.address, ethAmount)
      ).to.be.revertedWithCustomError(
        fixture.bonusPool,
        "AddressNotAllowedToSendEther"
      );
    });

    it("Succeeds if sent from public staking address", async () => {
      const ethAmount = BigNumber.from(1234);

      const bonusPoolEthBalanceBefore = await ethers.provider.getBalance(
        fixture.bonusPool.address
      );

      await depositEthToAddress(
        fixture.mockPublicStaking,
        fixture.bonusPool.address,
        ethAmount
      );

      const bonusPoolEthBalanceAfter = await ethers.provider.getBalance(
        fixture.bonusPool.address
      );

      expect(bonusPoolEthBalanceAfter).to.equal(
        bonusPoolEthBalanceBefore.add(ethAmount)
      );
    });
  });

  describe("terminate", async () => {
    it("Reverts if bonus NFT is not created", async () => {
      await expect(
        fixture.bonusPool.connect(fixture.mockLockupSigner).terminate()
      ).to.be.revertedWithCustomError(
        fixture.bonusPool,
        "BonusTokenNotCreated"
      );
    });

    it("Reverts if called before minted bonus stake free after time not reached", async () => {
      await mintBonusPosition(
        accounts,
        fixture.totalBonusAmount,
        fixture.alca,
        fixture.bonusPool,
        fixture.mockFactorySigner
      );

      await expect(
        fixture.bonusPool.connect(fixture.mockLockupSigner).terminate()
      ).to.be.revertedWithCustomError(
        fixture.publicStaking,
        "FreeAfterTimeNotReached"
      );
    });

    describe("with bonus position minted + rewards available", async () => {
      let tokenId: BigNumber;

      beforeEach(async () => {
        tokenId = await mintBonusPosition(
          accounts,
          fixture.totalBonusAmount,
          fixture.alca,
          fixture.bonusPool,
          fixture.mockFactorySigner
        );
      });

      it("Distributes all bonus eth/tokens to reward pool", async () => {
        const [, freeAfter, , , ,] = await fixture.publicStaking.getPosition(
          tokenId
        );

        await ensureBlockIsAtLeast(freeAfter.toNumber() + 1);

        const [expectedBonusRewardEth, expectedBonusRewardToken] =
          await calculateTerminationProfits(tokenId, fixture.publicStaking);

        const rewardPoolEthBalanceBefore = await ethers.provider.getBalance(
          fixture.rewardPoolAddress
        );
        const rewardPoolTokenBalanceBefore = await fixture.alca.balanceOf(
          fixture.rewardPoolAddress
        );

        await (
          await fixture.bonusPool.connect(fixture.mockLockupSigner).terminate()
        ).wait();

        const rewardPoolEthBalanceAfter = await ethers.provider.getBalance(
          fixture.rewardPoolAddress
        );
        const rewardPoolTokenBalanceAfter = await fixture.alca.balanceOf(
          fixture.rewardPoolAddress
        );

        const rewardEthDiff = rewardPoolEthBalanceAfter.sub(
          rewardPoolEthBalanceBefore
        );
        const rewardTokenDiff = rewardPoolTokenBalanceAfter.sub(
          rewardPoolTokenBalanceBefore
        );

        expect(rewardEthDiff).to.equal(expectedBonusRewardEth);
        expect(rewardTokenDiff).to.equal(expectedBonusRewardToken);
      });
    });
  });
});
