import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { BonusPool, Lockup, RewardPool } from "../../typechain-types";
import { BaseTokensFixture } from "../setup";
import {
  blockNumberAtLockupDeployment,
  deployFixture,
  distributeProfits,
  example,
  jumpToPostLockState,
  lockDuration,
  lockStakedNFT,
  numberOfLockingUsers,
  profitALCA,
  profitETH,
  startBlock,
} from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

describe("getter functions", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];

  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(deployFixture));
  });

  it("should get locking enrollment start block", async () => {
    expect(await fixture.lockup.getLockupStartBlock()).to.be.equal(
      blockNumberAtLockupDeployment + 1 + startBlock
    );
  });

  it("should get locking enrollment end block", async () => {
    expect(await fixture.lockup.getLockupEndBlock()).to.be.equal(
      blockNumberAtLockupDeployment + 1 + startBlock + lockDuration
    );
  });

  it("should get temporary reward balance", async () => {
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    await jumpToPostLockState(fixture);
    await fixture.lockup.aggregateProfits();
    const [ethRewards, aclRewards] = await fixture.lockup
      .connect(accounts[1])
      .getTemporaryRewardBalance(accounts[1].address);
    expect(ethRewards).to.be.equal(ethers.utils.parseEther("4"));
    expect(aclRewards).to.be.equal(ethers.utils.parseEther("40000"));
  });

  it("should get reward pool address", async () => {
    expect(await fixture.lockup.getRewardPoolAddress()).to.be.equal(
      fixture.rewardPool.address
    );
  });

  it("should get bonus pool address", async () => {
    expect(await fixture.lockup.getBonusPoolAddress()).to.be.equal(
      fixture.bonusPool.address
    );
  });

  it("should get current number of locked positions", async () => {
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    expect(
      await fixture.lockup.getCurrentNumberOfLockedPositions()
    ).to.be.equal(numberOfLockingUsers);
    await jumpToPostLockState(fixture);
    await fixture.lockup.aggregateProfits();
    await fixture.lockup
      .connect(accounts[1])
      .unlock(accounts[1].address, false);
    expect(
      await fixture.lockup.getCurrentNumberOfLockedPositions()
    ).to.be.equal(numberOfLockingUsers - 1);
  });

  it("should get index in token array for token ID", async () => {
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      expect(await fixture.lockup.getIndexByTokenId(i * 10)).to.be.equal(i);
    }
  });

  it("should get position for index in token array", async () => {
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      expect(await fixture.lockup.getPositionByIndex(i)).to.be.equal(i * 10);
    }
  });

  it("should get original number of locked shares", async () => {
    let originalLockedShares = 0n;
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
      originalLockedShares += ethers.utils
        .parseEther(example.distribution.users["user" + i].shares)
        .toBigInt();
    }
    expect(await fixture.lockup.getOriginalLockedShares()).to.be.equal(
      originalLockedShares
    );
  });

  it("should get reserved percentage amount", async () => {
    expect(await fixture.lockup.getReservedPercentage()).to.be.equal(20);
  });
});
