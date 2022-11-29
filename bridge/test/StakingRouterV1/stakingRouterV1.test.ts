import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BytesLike } from "ethers";
import { ethers } from "hardhat";
import { deployCreateAndRegister } from "../../scripts/lib/alicenetFactory";
import {
  CONTRACT_ADDR,
  EVENT_DEPLOYED_RAW,
  STAKING_ROUTER_V1,
} from "../../scripts/lib/constants";
import {
  BonusPool,
  Lockup,
  RewardPool,
  StakingRouterV1,
} from "../../typechain-types";
import { getEventVar } from "../factory/Setup";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import { getImpersonatedSigner } from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
  stakingRouterV1: StakingRouterV1;
  mockFactorySigner: SignerWithAddress;
}

const startBlock = 100;
const lockDuration = 2;
const stakedAmount = ethers.utils.parseEther("100").toBigInt();
const totalBonusAmount = ethers.utils.parseEther("10000");
const migrationAmount = ethers.utils.parseEther("100");
let rewardPoolAddress: any;

async function deployFixture() {
  await preFixtureSetup();
  const signers = await ethers.getSigners();
  const fixture = await deployFactoryAndBaseTokens(signers[0]);
  // deploy lockup contract
  const lockupBase = await ethers.getContractFactory("Lockup");
  const lockupDeployCode = lockupBase.getDeployTransaction(
    startBlock,
    lockDuration,
    totalBonusAmount
  ).data as BytesLike;
  // deploy Lockup
  let contractName = ethers.utils.formatBytes32String("Lockup");
  let txResponse = await fixture.factory.deployCreateAndRegister(
    lockupDeployCode,
    contractName
  );
  // get the address from the event
  const lockupAddress = await getEventVar(
    txResponse,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  // deploy staking router
  await ethers.getContractFactory("StakingRouterV1");
  contractName = ethers.utils.formatBytes32String(STAKING_ROUTER_V1);
  txResponse = await deployCreateAndRegister(
    STAKING_ROUTER_V1,
    fixture.factory,
    ethers,
    [],
    contractName
  );
  // get the address from the event
  const stakingRouterAddress = await getEventVar(
    txResponse,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  await posFixtureSetup(fixture.factory, fixture.alca);
  const lockup = await ethers.getContractAt("Lockup", lockupAddress);
  // get the address of the reward pool from the lockup contract
  rewardPoolAddress = await lockup.getRewardPoolAddress();
  const rewardPool = await ethers.getContractAt(
    "RewardPool",
    rewardPoolAddress
  );
  // get the address of the bonus pool from the reward pool contract
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);
  // connect and instance of the staking router
  const stakingRouterV1 = await ethers.getContractAt(
    STAKING_ROUTER_V1,
    stakingRouterAddress
  );
  const asFactory = await getImpersonatedSigner(fixture.factory.address);

  return {
    fixture: {
      ...fixture,
      rewardPool,
      lockup,
      bonusPool,
      stakingRouterV1,
      mockFactorySigner: asFactory,
    },
    accounts: signers,
  };
}

describe("StakingRouterV1", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  beforeEach(async () => {
    ({ fixture, accounts } = await loadFixture(deployFixture));
  });

  describe("migrateAndStake", async () => {
    it("reverts if stakingAmount is greater than migratedAmount", async () => {
      const shortMigrationAmount = stakedAmount - 1n;
      // ensure no bonus tokens are minted
      await fixture.alca.connect(fixture.mockFactorySigner).finishEarlyStage();
      await (
        await fixture.legacyToken.increaseAllowance(
          fixture.stakingRouterV1.address,
          migrationAmount
        )
      ).wait();

      const tokenOwner = accounts[1];
      await expect(
        fixture.stakingRouterV1.migrateAndStake(
          tokenOwner.address,
          shortMigrationAmount,
          stakedAmount
        )
      )
        .to.be.revertedWithCustomError(
          fixture.stakingRouterV1,
          "InvalidStakingAmount"
        )
        .withArgs(stakedAmount, shortMigrationAmount);
    });

    it("successfully migrates legacy token from sender, stakes amount and transfers position to address specified and remainder to sender", async () => {
      const sender = accounts[0];
      const recipient = accounts[1];
      const expectedMigrationAmountAfterConversion = await (
        await fixture.alca.convert(migrationAmount)
      ).toBigInt();

      const expectedRemainder =
        expectedMigrationAmountAfterConversion - stakedAmount;

      const balanceOfSenderBefore = await fixture.legacyToken.balanceOf(
        sender.address
      );
      expect(await fixture.alca.balanceOf(recipient.address)).to.equal(0);

      expect(
        await fixture.publicStaking.balanceOf(recipient.address)
      ).to.be.equal(0);
      await (
        await fixture.legacyToken
          .connect(sender)
          .increaseAllowance(fixture.stakingRouterV1.address, migrationAmount)
      ).wait();

      await expect(
        fixture.stakingRouterV1
          .connect(sender)
          .migrateAndStake(recipient.address, migrationAmount, stakedAmount)
      ).to.not.be.reverted;

      expect(await fixture.legacyToken.balanceOf(sender.address)).to.be.equal(
        balanceOfSenderBefore.sub(migrationAmount)
      );
      expect(await fixture.alca.balanceOf(recipient.address)).to.be.equal(
        expectedRemainder
      );
      expect(
        await fixture.publicStaking.balanceOf(recipient.address)
      ).to.be.equal(1);

      const tokenID = await fixture.publicStaking.tokenOfOwnerByIndex(
        recipient.address,
        0
      );
      const [shares, , , , ,] = await fixture.publicStaking.getPosition(
        tokenID
      );

      expect(shares).to.be.equal(stakedAmount);
    });
  });

  describe("migrateStakeAndLock", async () => {
    it("reverts if stakingAmount is greater than migratedAmount", async () => {
      const shortMigrationAmount = stakedAmount - 1n;
      // ensure no bonus tokens are minted
      await fixture.alca.connect(fixture.mockFactorySigner).finishEarlyStage();
      await (
        await fixture.legacyToken.increaseAllowance(
          fixture.stakingRouterV1.address,
          migrationAmount
        )
      ).wait();

      const tokenOwner = accounts[1];
      await expect(
        fixture.stakingRouterV1.migrateStakeAndLock(
          tokenOwner.address,
          shortMigrationAmount,
          stakedAmount
        )
      )
        .to.be.revertedWithCustomError(
          fixture.stakingRouterV1,
          "InvalidStakingAmount"
        )
        .withArgs(stakedAmount, shortMigrationAmount);
    });

    it("successfully migrates legacy token from sender, stakes amount, locks stake and transfers position to address specified and remainder to sender", async () => {
      const sender = accounts[0];
      const recipient = accounts[1];
      const expectedMigrationAmountAfterConversion = await (
        await fixture.alca.convert(migrationAmount)
      ).toBigInt();

      const expectedRemainder =
        expectedMigrationAmountAfterConversion - stakedAmount;

      const balanceOfSenderBefore = await fixture.legacyToken.balanceOf(
        sender.address
      );
      expect(await fixture.alca.balanceOf(recipient.address)).to.equal(0);

      expect(await fixture.lockup.tokenOf(recipient.address)).to.be.equal(0);
      await (
        await fixture.legacyToken
          .connect(sender)
          .increaseAllowance(fixture.stakingRouterV1.address, migrationAmount)
      ).wait();

      await expect(
        fixture.stakingRouterV1
          .connect(sender)
          .migrateStakeAndLock(recipient.address, migrationAmount, stakedAmount)
      ).to.not.be.reverted;

      expect(await fixture.legacyToken.balanceOf(sender.address)).to.be.equal(
        balanceOfSenderBefore.sub(migrationAmount)
      );
      expect(await fixture.alca.balanceOf(recipient.address)).to.be.equal(
        expectedRemainder
      );
      const tokenID = await fixture.lockup.tokenOf(recipient.address);
      expect(tokenID).to.not.be.equal(0);

      const [shares, , , , ,] = await fixture.publicStaking.getPosition(
        tokenID
      );

      expect(shares).to.be.equal(stakedAmount);
    });
  });

  describe("stakeAndLock", async () => {
    it("successfully stakes amount from sender and transfers position to specified account", async () => {
      const sender = accounts[0];
      const recipient = accounts[1];

      const balanceOfSenderBefore = await fixture.alca.balanceOf(
        sender.address
      );
      expect(await fixture.alca.balanceOf(recipient.address)).to.equal(0);

      expect(await fixture.lockup.tokenOf(recipient.address)).to.be.equal(0);

      await (
        await fixture.alca
          .connect(sender)
          .increaseAllowance(fixture.stakingRouterV1.address, stakedAmount)
      ).wait();

      await expect(
        fixture.stakingRouterV1
          .connect(sender)
          .stakeAndLock(recipient.address, stakedAmount)
      ).to.not.be.reverted;

      expect(await fixture.alca.balanceOf(sender.address)).to.be.equal(
        balanceOfSenderBefore.sub(stakedAmount)
      );

      const tokenID = await fixture.lockup.tokenOf(recipient.address);
      expect(tokenID).to.not.be.equal(0);

      const [shares, , , , ,] = await fixture.publicStaking.getPosition(
        tokenID
      );

      expect(shares).to.be.equal(stakedAmount);
    });
  });
});
