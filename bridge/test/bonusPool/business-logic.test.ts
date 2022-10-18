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
  mineBlocks,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import { getImpersonatedSigner } from "./setup";

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

  await posFixtureSetup(fixture.factory, fixture.aToken);

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
    .deploy(fixture.aToken.address, aliceNetFactoryAddress, totalBonusAmount);
  // Deploy the bonus pool standalone
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);

  const alcaRewards = ethers.utils.parseEther("1000000");
  const ethRewards = ethers.utils.parseEther("10");

  await depositEthForStakingRewards(signers, fixture, ethRewards);
  await depositTokensForStakingRewards(signers, fixture, alcaRewards);

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
        await fixture.aToken
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
        await fixture.aToken
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
        await fixture.aToken
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
      const initialTotalLocked = 1234;
      const finalSharesLocked = 1234;
      await expect(
        fixture.bonusPool
          .connect(fixture.mockLockupSigner)
          .terminate(finalSharesLocked, initialTotalLocked)
      ).to.be.revertedWithCustomError(
        fixture.bonusPool,
        "BonusTokenNotCreated"
      );
    });

    it("Reverts if called before minted bonus stake free after time not reached", async () => {
      await mintBonusPosition(accounts, fixture);
      const originalTotalSharesLocked = BigNumber.from(8000);
      const finalTotalSharesLocked = BigNumber.from(8000);

      await expect(
        fixture.bonusPool
          .connect(fixture.mockLockupSigner)
          .terminate(finalTotalSharesLocked, originalTotalSharesLocked)
      ).to.be.revertedWithCustomError(
        fixture.publicStaking,
        "FreeAfterTimeNotReached"
      );
    });

    describe("with bonus position minted + rewards available", async () => {
      let tokenId: BigNumber;

      beforeEach(async () => {
        tokenId = await mintBonusPosition(accounts, fixture);
      });

      it("Reverts if original shares locked is 0", async () => {
        const originalTotalSharesLocked = 0;
        const finalSharesLocked = 1234;
        await expect(
          fixture.bonusPool
            .connect(fixture.mockLockupSigner)
            .terminate(finalSharesLocked, originalTotalSharesLocked)
        ).to.be.revertedWithCustomError(
          fixture.bonusPool,
          "InvalidOriginalSharesValue"
        );
      });

      it("Reverts if original shares locked less that final shares locked", async () => {
        const originalTotalSharesLocked = 1233;
        const finalSharesLocked = 1234;
        await expect(
          fixture.bonusPool
            .connect(fixture.mockLockupSigner)
            .terminate(finalSharesLocked, originalTotalSharesLocked)
        ).to.be.revertedWithCustomError(
          fixture.bonusPool,
          "InvalidOriginalSharesValue"
        );
      });

      it("Distributes all bonus eth/tokens to reward pool when final total equals original total", async () => {
        const originalTotalSharesLocked = BigNumber.from(8000);
        const finalTotalSharesLocked = BigNumber.from(8000);

        const [, freeAfter, , , ,] = await fixture.publicStaking.getPosition(
          tokenId
        );

        await ensureBlockIsAtLeast(freeAfter.toNumber() + 1);

        const rewardPoolEthBalanceBefore = await ethers.provider.getBalance(
          fixture.rewardPoolAddress
        );
        const rewardPoolTokenBalanceBefore = await fixture.aToken.balanceOf(
          fixture.rewardPoolAddress
        );

        const foundationEthBalanceBefore = await ethers.provider.getBalance(
          fixture.foundation.address
        );
        const factoryTokenBalanceBefore = await fixture.aToken.balanceOf(
          fixture.factory.address
        );

        const [, , , expectedBonusRewardEth, expectedBonusRewardToken] =
          await calculateTerminationProfits(
            finalTotalSharesLocked,
            originalTotalSharesLocked,
            tokenId,
            fixture
          );

        await (
          await fixture.bonusPool
            .connect(fixture.mockLockupSigner)
            .terminate(finalTotalSharesLocked, originalTotalSharesLocked)
        ).wait();

        const rewardPoolEthBalanceAfter = await ethers.provider.getBalance(
          fixture.rewardPoolAddress
        );
        const rewardPoolTokenBalanceAfter = await fixture.aToken.balanceOf(
          fixture.rewardPoolAddress
        );

        const rewardEthDiff = rewardPoolEthBalanceAfter.sub(
          rewardPoolEthBalanceBefore
        );
        const rewardTokenDiff = rewardPoolTokenBalanceAfter.sub(
          rewardPoolTokenBalanceBefore
        );

        const foundationEthBalanceAfter = await ethers.provider.getBalance(
          fixture.foundation.address
        );
        const factoryTokenBalanceAfter = await fixture.aToken.balanceOf(
          fixture.factory.address
        );

        const foundationEthDiff = foundationEthBalanceAfter.sub(
          foundationEthBalanceBefore
        );
        const factoryTokenDiff = factoryTokenBalanceAfter.sub(
          factoryTokenBalanceBefore
        );

        expect(rewardEthDiff).to.equal(expectedBonusRewardEth);
        expect(rewardTokenDiff).to.equal(expectedBonusRewardToken);
        expect(foundationEthDiff).to.equal(BigNumber.from(0));
        expect(factoryTokenDiff).to.equal(BigNumber.from(0));
      });

      it("Distributes remaining tokens to factory and remaining eth to foundation when final shares less than original", async () => {
        const originalTotalSharesLocked = BigNumber.from(8000);
        const finalTotalSharesLocked = BigNumber.from(7000);

        const [, freeAfter, , , ,] = await fixture.publicStaking.getPosition(
          tokenId
        );

        await ensureBlockIsAtLeast(freeAfter.toNumber() + 1);

        const rewardPoolEthBalanceBefore = await ethers.provider.getBalance(
          fixture.rewardPoolAddress
        );
        const rewardPoolTokenBalanceBefore = await fixture.aToken.balanceOf(
          fixture.rewardPoolAddress
        );

        const foundationEthBalanceBefore = await ethers.provider.getBalance(
          fixture.foundation.address
        );
        const factoryTokenBalanceBefore = await fixture.aToken.balanceOf(
          fixture.factory.address
        );

        const [
          estimatedPayoutEth,
          estimatedPayoutToken,
          ,
          expectedBonusRewardEth,
          expectedBonusRewardToken,
        ] = await calculateTerminationProfits(
          finalTotalSharesLocked,
          originalTotalSharesLocked,
          tokenId,
          fixture
        );

        await (
          await fixture.bonusPool
            .connect(fixture.mockLockupSigner)
            .terminate(finalTotalSharesLocked, originalTotalSharesLocked)
        ).wait();

        const rewardPoolEthBalanceAfter = await ethers.provider.getBalance(
          fixture.rewardPoolAddress
        );
        const rewardPoolTokenBalanceAfter = await fixture.aToken.balanceOf(
          fixture.rewardPoolAddress
        );

        const rewardEthDiff = rewardPoolEthBalanceAfter.sub(
          rewardPoolEthBalanceBefore
        );
        const rewardTokenDiff = rewardPoolTokenBalanceAfter.sub(
          rewardPoolTokenBalanceBefore
        );

        const foundationEthBalanceAfter = await ethers.provider.getBalance(
          fixture.foundation.address
        );
        const factoryTokenBalanceAfter = await fixture.aToken.balanceOf(
          fixture.factory.address
        );

        const foundationEthDiff = foundationEthBalanceAfter.sub(
          foundationEthBalanceBefore
        );
        const factoryTokenDiff = factoryTokenBalanceAfter.sub(
          factoryTokenBalanceBefore
        );

        expect(rewardEthDiff).to.equal(expectedBonusRewardEth);
        expect(rewardTokenDiff).to.equal(expectedBonusRewardToken);
        expect(foundationEthDiff).to.equal(
          estimatedPayoutEth.sub(expectedBonusRewardEth)
        );
        expect(factoryTokenDiff).to.equal(
          estimatedPayoutToken
            .add(fixture.totalBonusAmount)
            .sub(expectedBonusRewardToken)
        );
      });
    });
  });
});

async function mintBonusPosition(
  accounts: SignerWithAddress[],
  fixture: Fixture
) {
  const exactStakeAmount = fixture.totalBonusAmount;
  await (
    await fixture.aToken
      .connect(accounts[0])
      .transfer(fixture.bonusPool.address, exactStakeAmount)
  ).wait();

  const receipt = await (
    await fixture.bonusPool
      .connect(fixture.mockFactorySigner)
      .createBonusStakedPosition()
  ).wait();

  const createdEvent = receipt.events?.find(
    (event) => event.event === "BonusPositionCreated"
  );

  return createdEvent?.args?.tokenID;
}

async function depositEthForStakingRewards(
  accounts: SignerWithAddress[],
  fixture: BaseTokensFixture,
  eth: BigNumber
): Promise<void> {
  await (
    await fixture.publicStaking
      .connect(accounts[0])
      .depositEth(42, { value: eth })
  ).wait();
}

async function depositTokensForStakingRewards(
  accounts: SignerWithAddress[],
  fixture: BaseTokensFixture,
  alca: BigNumber
): Promise<void> {
  await (
    await fixture.aToken
      .connect(accounts[0])
      .increaseAllowance(fixture.publicStaking.address, alca)
  ).wait();

  await (
    await fixture.publicStaking.connect(accounts[0]).depositToken(42, alca)
  ).wait();
}

async function calculateTerminationProfits(
  finalTotalSharesLocked: BigNumber,
  originalTotalSharesLocked: BigNumber,
  tokenId: BigNumber,
  fixture: Fixture
): Promise<[BigNumber, BigNumber, BigNumber, BigNumber, BigNumber]> {
  const scalingFactor = await fixture.bonusPool.SCALING_FACTOR();
  const bonusRate = await fixture.bonusPool.getScaledBonusRate(
    originalTotalSharesLocked
  );
  const overallProportion = finalTotalSharesLocked
    .mul(scalingFactor)
    .div(originalTotalSharesLocked);

  // estimate all profits does not include the original stake amount, hence no need to subtract it here
  const [estimatedPayoutEth, estimatedPayoutToken] =
    await fixture.publicStaking.estimateAllProfits(tokenId);

  const expectedBonusShares = bonusRate
    .mul(finalTotalSharesLocked)
    .div(scalingFactor);
  const expectedBonusRewardEth = overallProportion
    .mul(estimatedPayoutEth)
    .div(scalingFactor);
  const expectedBonusRewardToken = overallProportion
    .mul(estimatedPayoutToken)
    .div(scalingFactor);

  return [
    estimatedPayoutEth,
    estimatedPayoutToken,
    expectedBonusShares,
    expectedBonusRewardEth,
    expectedBonusRewardToken.add(expectedBonusShares),
  ];
}

async function ensureBlockIsAtLeast(targetBlock: number): Promise<void> {
  const currentBlock = await ethers.provider.getBlockNumber();
  if (currentBlock < targetBlock) {
    const blockDelta = targetBlock - currentBlock;
    await mineBlocks(BigInt(blockDelta));
  }
}

async function depositEthToAddress(
  accountFrom: SignerWithAddress,
  accountTo: string,
  eth: BigNumber
): Promise<void> {
  await accountFrom.sendTransaction({
    to: accountTo,
    value: eth,
  });
}
