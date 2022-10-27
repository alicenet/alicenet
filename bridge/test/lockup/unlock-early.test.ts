import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert, expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { BonusPool, Lockup, RewardPool } from "../../typechain-types";
import { BaseTokensFixture, mineBlocks } from "../setup";
import {
  deployFixture,
  distributeProfits,
  example,
  getEthConsumedAsGas,
  getState,
  getUserLockingInfo,
  lockStakedNFT,
  numberOfLockingUsers,
  profitALCA,
  profitETH,
  showState,
} from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

describe("Testing Unlock Early", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];

  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(deployFixture));
    showState("Initial State with staked positions", await getState(fixture));
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    showState("After Locking", await getState(fixture));
    const expectedState = await getState(fixture);

    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    expectedState.contracts.publicStaking.eth += profitETH.toBigInt();
    expectedState.contracts.publicStaking.alca += profitALCA.toBigInt();
    showState("After Distribution", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("should not unlock early for a value greater than locked", async () => {
    const user1Shares = ethers.utils.parseEther(
      example.distribution.users.user1.shares
    );
    await expect(
      fixture.lockup.connect(accounts[1]).unlockEarly(user1Shares.add(1), false)
    )
      .to.be.revertedWithCustomError(
        fixture.lockup,
        "InsufficientBalanceForEarlyExit"
      )
      .withArgs(user1Shares.add(1), user1Shares);
  });

  it("should unlock-early user for 50% of initial position without re-staking unlocked shares", async () => {
    const expectedState = await getState(fixture);
    const unlockingEarlyUsers = [2];
    for (let userId = 0; userId < unlockingEarlyUsers.length; userId++) {
      const user = "user" + unlockingEarlyUsers[userId];
      const userLockingInfo = await getUserLockingInfo(
        fixture,
        unlockingEarlyUsers[userId]
      );
      const exitAmount = userLockingInfo.userCurrentShares / 2n;
      // new lockup position is created for user's remaining staking balance
      const newPositionId = (
        await fixture.publicStaking.getLatestMintedPositionID()
      )
        .add(1)
        .toBigInt();
      expectedState.lockupPositions[userLockingInfo.index.toString()].tokenId =
        newPositionId;
      // new position is assigned to user
      expectedState.users[user].tokenId = newPositionId;
      expectedState.users[user].tokenOwner = userLockingInfo.owner.address;
      // user receives unlocked staking amount
      expectedState.contracts.publicStaking.alca -= exitAmount;
      expectedState.users[user].alca += exitAmount;
      // user receives unlocked ALCA & ETH profits
      expectedState.contracts.publicStaking.alca -=
        userLockingInfo.profitALCAUser;
      expectedState.users[user].alca += userLockingInfo.profitALCAUser;
      expectedState.contracts.publicStaking.eth -=
        userLockingInfo.profitETHUser;
      expectedState.users[user].eth += userLockingInfo.profitETHUser;
      // staking position is updated with remaining shares
      expectedState.stakingPositions[user].shares =
        userLockingInfo.userCurrentShares - exitAmount;
      expectedState.stakingPositions[user].tokenId = newPositionId;
      // user and staking reserved ALCA and ETH go to reward pool
      expectedState.users[user].alca -= userLockingInfo.reservedProfitALCAUser;
      expectedState.users[user].eth -= userLockingInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.alca +=
        userLockingInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.eth +=
        userLockingInfo.reservedProfitETHUser;
      const tx = await fixture.lockup
        .connect(userLockingInfo.owner)
        .unlockEarly(exitAmount, false);
      // account for used gas
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
    }
    showState("After Unlock early", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("should unlock-early user for 50% of initial position re-staking unlocked shares", async () => {
    const expectedState = await getState(fixture);
    const unlockingEarlyUsers = [2];
    for (let userId = 0; userId < unlockingEarlyUsers.length; userId++) {
      const user = "user" + unlockingEarlyUsers[userId];
      const userLockingInfo = await getUserLockingInfo(
        fixture,
        unlockingEarlyUsers[userId]
      );
      const exitAmount = userLockingInfo.userCurrentShares / 2n;
      // new lockup position is created for user's remaining staking balance
      const newPositionId = (
        await fixture.publicStaking.getLatestMintedPositionID()
      )
        .add(1)
        .toBigInt();
      expectedState.lockupPositions[userLockingInfo.index.toString()].index =
        userLockingInfo.index.toBigInt();
      expectedState.lockupPositions[userLockingInfo.index.toString()].owner =
        userLockingInfo.owner.address;
      expectedState.lockupPositions[userLockingInfo.index.toString()].tokenId =
        newPositionId;
      // new position is assigned to user
      expectedState.users[user].tokenId = newPositionId;
      expectedState.users[user].tokenOwner = userLockingInfo.owner.address;
      // staking position is updated with remaining shares
      expectedState.stakingPositions[user].shares =
        userLockingInfo.userCurrentShares - exitAmount;
      expectedState.stakingPositions[user].tokenId = newPositionId;
      // user only earns free ETH from staking since unlocked ALCA is re-staked
      expectedState.contracts.publicStaking.eth -=
        userLockingInfo.profitETHUser;
      expectedState.users[user].eth += userLockingInfo.profitETHUser;
      // user's reserved ETH goes to reward pool
      expectedState.users[user].eth -= userLockingInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.eth +=
        userLockingInfo.reservedProfitETHUser;
      // staking's reserved ALCA goes to reward pool
      expectedState.contracts.publicStaking.alca -=
        userLockingInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.alca +=
        userLockingInfo.reservedProfitALCAUser;
      // proceed to unlock early
      const tx = await fixture.lockup
        .connect(userLockingInfo.owner)
        .unlockEarly(exitAmount, true);
      // account for used gas
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
    }
    showState("After Unlock early", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("should unlock-early user for 100% of initial position without re-staking unlocked shares", async () => {
    const expectedState = await getState(fixture);
    const unlockingEarlyUsers = [3];
    for (let userId = 0; userId < unlockingEarlyUsers.length; userId++) {
      const user = "user" + unlockingEarlyUsers[userId];
      const userLockingInfo = await getUserLockingInfo(
        fixture,
        unlockingEarlyUsers[userId]
      );
      const exitAmount = userLockingInfo.userCurrentShares;
      // user receives unlocked shares
      expectedState.contracts.publicStaking.alca -= exitAmount;
      expectedState.users[user].alca += exitAmount;
      // user receives ALCA & ETH profits
      expectedState.contracts.publicStaking.alca -=
        userLockingInfo.profitALCAUser;
      expectedState.users[user].alca += userLockingInfo.profitALCAUser;
      expectedState.contracts.publicStaking.eth -=
        userLockingInfo.profitETHUser;
      expectedState.users[user].eth += userLockingInfo.profitETHUser;
      // user looses lockup position
      expectedState.users[user].tokenId = ethers.constants.Zero.toBigInt();
      expectedState.users[user].tokenOwner = ethers.constants.AddressZero;
      // user lockup position is re-used with last position
      const lastPosition = (
        await fixture.lockup.getCurrentNumberOfLockedPositions()
      ).toString();
      expectedState.lockupPositions[lastPosition].index =
        userLockingInfo.index.toBigInt();
      expectedState.lockupPositions[userLockingInfo.index.toString()] =
        expectedState.lockupPositions[lastPosition];
      // last position is deleted
      delete expectedState.lockupPositions[lastPosition];
      // staking position is deleted
      delete expectedState.stakingPositions[user];
      // reserved ALCA & ETH goes to reward pool
      expectedState.contracts.rewardPool.alca +=
        userLockingInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.eth +=
        userLockingInfo.reservedProfitETHUser;
      expectedState.users[user].alca -= userLockingInfo.reservedProfitALCAUser;
      expectedState.users[user].eth -= userLockingInfo.reservedProfitETHUser;
      // proceed with unlock
      const tx = await fixture.lockup
        .connect(userLockingInfo.owner)
        .unlockEarly(exitAmount, false);
      // accoun for used gas
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
    }
    // unlocked positions is decreased
    if (expectedState.contracts.lockup.lockedPositions !== undefined)
      expectedState.contracts.lockup.lockedPositions -= BigInt(
        unlockingEarlyUsers.length
      );
    showState("After Unlock early", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("should unlock-early user for 100% of initial position re-staking unlocked shares", async () => {
    const expectedState = await getState(fixture);
    const unlockingEarlyUsers = [3];
    for (let userId = 0; userId < unlockingEarlyUsers.length; userId++) {
      const user = "user" + unlockingEarlyUsers[userId];
      const userLockingInfo = await getUserLockingInfo(
        fixture,
        unlockingEarlyUsers[userId]
      );
      const exitAmount = userLockingInfo.userCurrentShares;
      // user only earns free ETH since ALCA is re-staked
      expectedState.contracts.publicStaking.eth -=
        userLockingInfo.profitETHUser;
      expectedState.users[user].eth += userLockingInfo.profitETHUser;
      // user's reserved ETH goes to reward pool
      expectedState.users[user].eth -= userLockingInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.eth +=
        userLockingInfo.reservedProfitETHUser;
      // user looses lock position
      expectedState.users[user].tokenId = ethers.constants.Zero.toBigInt();
      expectedState.users[user].tokenOwner = ethers.constants.AddressZero;
      // user lockup position is re-used with last one
      const lastPosition = (
        await fixture.lockup.getCurrentNumberOfLockedPositions()
      ).toString();
      expectedState.lockupPositions[lastPosition].index =
        userLockingInfo.index.toBigInt();
      expectedState.lockupPositions[userLockingInfo.index.toString()] =
        expectedState.lockupPositions[lastPosition];
      // last position is deleted
      delete expectedState.lockupPositions[lastPosition];
      // staking position is deleted
      delete expectedState.stakingPositions[user];
      // staking reserved ALCA goes to rewardpool
      expectedState.contracts.publicStaking.alca -=
        userLockingInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.alca +=
        userLockingInfo.reservedProfitALCAUser;
      // proceed with unlock early
      const tx = await fixture.lockup
        .connect(userLockingInfo.owner)
        .unlockEarly(exitAmount, true);
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
    }
    // number of locked positions is decreased
    if (expectedState.contracts.lockup.lockedPositions !== undefined)
      expectedState.contracts.lockup.lockedPositions -= BigInt(
        unlockingEarlyUsers.length
      );
    showState("After Unlock early", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("should unlock-early user for 100% of initial position in two phases (50%+50%) without second distribution and re-staking unlocked shares", async () => {
    const expectedState = await getState(fixture);
    const unlockingEarlyUsers = [2];
    for (let userId = 0; userId < unlockingEarlyUsers.length; userId++) {
      const user = "user" + unlockingEarlyUsers[userId];
      let userLockingInfo = await getUserLockingInfo(
        fixture,
        unlockingEarlyUsers[userId]
      );
      let exitAmount = userLockingInfo.userInitialShares / 2n;
      // new lockup position is created for user's remaining staking balance
      const newPositionId = (
        await fixture.publicStaking.getLatestMintedPositionID()
      )
        .add(1)
        .toBigInt();
      expectedState.lockupPositions[userLockingInfo.index.toString()].tokenId =
        newPositionId;
      expectedState.lockupPositions[userLockingInfo.index.toString()].index =
        userLockingInfo.index.toBigInt();
      expectedState.lockupPositions[userLockingInfo.index.toString()].owner =
        userLockingInfo.owner.address;
      // staking position is updated with remaining shares
      expectedState.stakingPositions[user].shares =
        userLockingInfo.userCurrentShares - exitAmount;
      expectedState.stakingPositions[user].tokenId = newPositionId;
      // new position is assigned to user
      expectedState.users[user].tokenId = newPositionId;
      expectedState.users[user].tokenOwner = userLockingInfo.owner.address;
      // user only earns free ETH from staking since unlocked ALCA is re-staked
      expectedState.contracts.publicStaking.eth -=
        userLockingInfo.profitETHUser;
      expectedState.users[user].eth += userLockingInfo.profitETHUser;
      // user's reserved ETH goes to reward pool
      expectedState.users[user].eth -= userLockingInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.eth +=
        userLockingInfo.reservedProfitETHUser;
      // staking's reserved ALCA goes to reward pool
      expectedState.contracts.publicStaking.alca -=
        userLockingInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.alca +=
        userLockingInfo.reservedProfitALCAUser;
      // proceed to unlock early
      const tx1 = await fixture.lockup
        .connect(userLockingInfo.owner)
        .unlockEarly(exitAmount, true);
      // account for used gas
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx1.wait());
      showState("After Unlock early phase 1", await getState(fixture));
      assert.deepEqual(await getState(fixture), expectedState);
      // fast forward to staking free after
      await mineBlocks(1n);
      userLockingInfo = await getUserLockingInfo(
        fixture,
        unlockingEarlyUsers[userId]
      );
      exitAmount = userLockingInfo.userInitialShares / 2n;
      // user looses lock position
      expectedState.users[user].tokenId = ethers.constants.Zero.toBigInt();
      expectedState.users[user].tokenOwner = ethers.constants.AddressZero;
      // user lockup position is re-used with last one
      const lastPosition = (
        await fixture.lockup.getCurrentNumberOfLockedPositions()
      ).toString();
      expectedState.lockupPositions[lastPosition].index =
        userLockingInfo.index.toBigInt();
      expectedState.lockupPositions[userLockingInfo.index.toString()] =
        expectedState.lockupPositions[lastPosition];
      delete expectedState.lockupPositions[lastPosition];
      // staking position is deleted
      delete expectedState.stakingPositions[user];
      if (expectedState.contracts.lockup.lockedPositions !== undefined)
        expectedState.contracts.lockup.lockedPositions -= BigInt(
          unlockingEarlyUsers.length
        );
      // proceed with unlock early
      const tx2 = await fixture.lockup
        .connect(userLockingInfo.owner)
        .unlockEarly(exitAmount, true);
      // account for used gas
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx2.wait());
      showState("After Unlock early phase 2", await getState(fixture));
      assert.deepEqual(await getState(fixture), expectedState);
    }
  });
});
