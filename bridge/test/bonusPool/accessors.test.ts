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
  calculateUserProfits,
  depositEthForStakingRewards,
  depositTokensForStakingRewards,
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
  const asLockup = await getImpersonatedSigner(lockupAddress);

  const totalBonusAmount = ethers.utils.parseEther("1000000");
  // deploy reward pool
  const rewardPool = await (await ethers.getContractFactory("RewardPool"))
    .connect(asLockup)
    .deploy(fixture.alca.address, aliceNetFactoryAddress, totalBonusAmount);
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
    it("getLockupContractAddress returns lockup contract address [ @skip-on-coverage ]", async () => {
      expect(await fixture.bonusPool.getLockupContractAddress()).to.equal(
        fixture.lockupAddress
      );
    });

    it("getRewardPoolAddress returns lockup contract address [ @skip-on-coverage ]", async () => {
      expect(await fixture.bonusPool.getRewardPoolAddress()).to.equal(
        fixture.rewardPoolAddress
      );
    });

    it("getBonusStakedPosition returns token id of staked position [ @skip-on-coverage ]", async () => {
      const tokenId = await mintBonusPosition(
        accounts,
        fixture.totalBonusAmount,
        fixture.alca,
        fixture.bonusPool,
        fixture.mockFactorySigner
      );
      expect(await fixture.bonusPool.getBonusStakedPosition()).to.equal(
        tokenId
      );
    });
  });

  describe("estimateBonusAmountWithReward", async () => {
    it("Returns expected amount based on bonus rate [ @skip-on-coverage ]", async () => {
      const tokenId = await mintBonusPosition(
        accounts,
        fixture.totalBonusAmount,
        fixture.alca,
        fixture.bonusPool,
        fixture.mockFactorySigner
      );

      const alcaRewards = ethers.utils.parseEther("1000000");
      const ethRewards = ethers.utils.parseEther("10");

      await depositEthForStakingRewards(
        accounts,
        fixture.publicStaking,
        ethRewards
      );
      await depositTokensForStakingRewards(
        accounts,
        fixture.alca,
        fixture.publicStaking,
        alcaRewards
      );

      const currentSharesLocked = BigNumber.from(8000);
      const userSharesLocked = BigNumber.from(4000);

      const [userExpectedBonusRewardEth, userExpectedBonusRewardToken] =
        await calculateUserProfits(
          userSharesLocked,
          currentSharesLocked,
          tokenId,
          fixture.publicStaking
        );

      const [bonusRewardEth, bonusRewardToken] =
        await fixture.bonusPool.estimateBonusAmountWithReward(
          currentSharesLocked,
          userSharesLocked
        );

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
