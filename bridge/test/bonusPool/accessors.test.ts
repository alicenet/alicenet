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
import { getImpersonatedSigner } from "./setup";

interface Fixture extends BaseTokensFixture {
  bonusPool: BonusPool;
  foundation: Foundation;
  lockupAddress: string;
  rewardPoolAddress: string;
  totalBonusAmount: BigNumber;
  mockFactorySigner: SignerWithAddress;
  mockLockupSigner: SignerWithAddress;
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
  const asLockup = await getImpersonatedSigner(lockupAddress);

  const totalBonusAmount = ethers.utils.parseEther("1000000");
  // deploy reward pool
  const rewardPool = await (await ethers.getContractFactory("RewardPool"))
    .connect(asLockup)
    .deploy(fixture.aToken.address, aliceNetFactoryAddress, totalBonusAmount);
  // Deploy the bonus pool standalone
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);

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

  describe("Public accessors", async () => {
    it("getLockupContractAddress returns lockup contract address", async () => {
      expect(await fixture.bonusPool.getLockupContractAddress()).to.equal(
        fixture.lockupAddress
      );
    });

    it("getRewardPoolAddress returns lockup contract address", async () => {
      expect(await fixture.bonusPool.getRewardPoolAddress()).to.equal(
        fixture.rewardPoolAddress
      );
    });

    it("SCALING_FACTOR returns expected value", async () => {
      expect(await fixture.bonusPool.SCALING_FACTOR()).to.equal(
        BigNumber.from("1000000000000000000")
      );
    });

    it("getBonusStakedPosition returns token id of staked position", async () => {
      const tokenId = await mintBonusPosition(accounts, fixture);
      expect(await fixture.bonusPool.getBonusStakedPosition()).to.equal(
        tokenId
      );
    });

    it("getScaledBonusRate returns correct value once bonus rate set", async () => {
      const initialTotalLocked = 1234;

      const expectedBonusRate = fixture.totalBonusAmount
        .mul(await fixture.bonusPool.SCALING_FACTOR())
        .div(initialTotalLocked);

      expect(
        await fixture.bonusPool.getScaledBonusRate(initialTotalLocked)
      ).to.equal(expectedBonusRate);
    });
  });

  describe("estimateBonusAmountWithReward", async () => {
    it("reverts if bonus token not minted", async () => {
      const initialTotalLocked = BigNumber.from(8000);
      const currentSharesLocked = BigNumber.from(8000);
      const userSharesLocked = BigNumber.from(4000);

      await expect(
        fixture.bonusPool.estimateBonusAmountWithReward(
          currentSharesLocked,
          initialTotalLocked,
          userSharesLocked
        )
      ).to.be.revertedWithCustomError(
        fixture.bonusPool,
        "BonusTokenNotCreated"
      );
    });

    it("Returns expected amount based on bonus rate", async () => {
      const tokenId = await mintBonusPosition(accounts, fixture);

      const alcaRewards = ethers.utils.parseEther("1000000");
      const ethRewards = ethers.utils.parseEther("10");

      await depositEthForStakingRewards(accounts, fixture, ethRewards);
      await depositTokensForStakingRewards(accounts, fixture, alcaRewards);

      const initialTotalLocked = BigNumber.from(8000);
      const currentSharesLocked = BigNumber.from(8000);
      const userSharesLocked = BigNumber.from(4000);

      const [
        expectedUserBonusShares,
        userExpectedBonusRewardEth,
        userExpectedBonusRewardToken,
      ] = await calculateUserProfits(
        userSharesLocked,
        currentSharesLocked,
        initialTotalLocked,
        tokenId,
        fixture
      );

      const [bonusShares, bonusRewardEth, bonusRewardToken] =
        await fixture.bonusPool.estimateBonusAmountWithReward(
          currentSharesLocked,
          initialTotalLocked,
          userSharesLocked
        );

      expect(bonusShares).to.equal(expectedUserBonusShares);
      expect(bonusRewardEth).to.equal(
        userExpectedBonusRewardEth,
        "bonusRewardEth are not equal"
      );
      expect(bonusRewardToken).to.equal(
        userExpectedBonusRewardToken,
        "bonusRewardToken are not equal"
      );
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
  fixture: Fixture,
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
  fixture: Fixture,
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

async function calculateUserProfits(
  userSharesLocked: BigNumber,
  currentTotalSharesLocked: BigNumber,
  originalTotalSharesLocked: BigNumber,
  tokenId: BigNumber,
  fixture: Fixture
): Promise<[BigNumber, BigNumber, BigNumber]> {
  const scalingFactor = await fixture.bonusPool.SCALING_FACTOR();
  const bonusRate = await fixture.bonusPool.getScaledBonusRate(
    originalTotalSharesLocked
  );
  const overallProportion = currentTotalSharesLocked
    .mul(scalingFactor)
    .div(originalTotalSharesLocked);
  const userProportion = userSharesLocked
    .mul(scalingFactor)
    .div(currentTotalSharesLocked);

  const [estimatedPayoutEth, estimatedPayoutToken] =
    await fixture.publicStaking.estimateAllProfits(tokenId);

  const totalExpectedBonusRewardEth = overallProportion
    .mul(estimatedPayoutEth)
    .div(scalingFactor);
  const totalExpectedBonusRewardToken = overallProportion
    .mul(estimatedPayoutToken)
    .div(scalingFactor);

  const expectedUserBonusShares = bonusRate
    .mul(userSharesLocked)
    .div(scalingFactor);
  const userExpectedBonusRewardEth = userProportion
    .mul(totalExpectedBonusRewardEth)
    .div(scalingFactor);
  const userExpectedBonusRewardToken = userProportion
    .mul(totalExpectedBonusRewardToken)
    .div(scalingFactor);

  return [
    expectedUserBonusShares,
    userExpectedBonusRewardEth,
    userExpectedBonusRewardToken,
  ];
}
