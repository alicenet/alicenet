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
      const userShares = ethers.utils.parseEther(
        example.distribution.users[user].shares
      );
      const userUnlockInfo = await getUserUnlockInfo(
        fixture,
        unlockingEarlyUsers[userId],
        userShares,
        50
      );
      // new lockup position is created for user's remaining staking balance
      const newPositionId = (
        await fixture.publicStaking.getLatestMintedPositionID()
      )
        .add(1)
        .toBigInt();
      expectedState.lockupPositions[userUnlockInfo.index.toString()].tokenId =
        newPositionId;
      // new position is assigned to user
      expectedState.users[user].tokenId = newPositionId;
      expectedState.users[user].tokenOwner = userUnlockInfo.owner.address;
      // user receives unlocked staking amount
      expectedState.contracts.publicStaking.alca -= userUnlockInfo.exitAmount;
      expectedState.users[user].alca += userUnlockInfo.exitAmount;
      // user receives unlocked ALCA & ETH profits
      expectedState.contracts.publicStaking.alca -=
        userUnlockInfo.profitALCAUser;
      expectedState.users[user].alca += userUnlockInfo.profitALCAUser;
      expectedState.contracts.publicStaking.eth -= userUnlockInfo.profitETHUser;
      expectedState.users[user].eth += userUnlockInfo.profitETHUser;
      // user and staking reserved ALCA and ETH go to reward pool
      expectedState.users[user].alca -= userUnlockInfo.reservedProfitALCAUser;
      expectedState.users[user].eth -= userUnlockInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.alca +=
        userUnlockInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.eth +=
        userUnlockInfo.reservedProfitETHUser;
      const tx = await fixture.lockup
        .connect(userUnlockInfo.owner)
        .unlockEarly(userUnlockInfo.exitAmount, false);
      // account for used gas
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
    }

    showState("After Unlock early", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("should unlock-early user for 50% of initial position re-staking unlocked shares", async () => {
    const lastStakingPosition = (
      await fixture.publicStaking.getLatestMintedPositionID()
    ).toNumber();
    const expectedState = await getState(fixture, lastStakingPosition);
    const unlockingEarlyUsers = [2];
    for (let userId = 0; userId < unlockingEarlyUsers.length; userId++) {
      const user = "user" + unlockingEarlyUsers[userId];
      const userShares = ethers.utils.parseEther(
        example.distribution.users[user].shares
      );
      const userUnlockInfo = await getUserUnlockInfo(
        fixture,
        unlockingEarlyUsers[userId],
        userShares,
        50
      );
      // new lockup position is created for user's remaining staking balance
      const newPositionId = (
        await fixture.publicStaking.getLatestMintedPositionID()
      )
        .add(1)
        .toBigInt();
      expectedState.lockupPositions[userUnlockInfo.index.toString()].index =
        userUnlockInfo.index.toBigInt();
      expectedState.lockupPositions[userUnlockInfo.index.toString()].owner =
        userUnlockInfo.owner.address;
      expectedState.lockupPositions[userUnlockInfo.index.toString()].tokenId =
        newPositionId;
      // new position is assigned to user
      expectedState.users[user].tokenId = newPositionId;
      expectedState.users[user].tokenOwner = userUnlockInfo.owner.address;
      // user only earns free ETH from staking since unlocked ALCA is re-staked
      expectedState.contracts.publicStaking.eth -= userUnlockInfo.profitETHUser;
      expectedState.users[user].eth += userUnlockInfo.profitETHUser;
      // user's reserved ETH goes to reward pool
      expectedState.users[user].eth -= userUnlockInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.eth +=
        userUnlockInfo.reservedProfitETHUser;
      // staking's reserved ALCA goes to reward pool
      expectedState.contracts.publicStaking.alca -=
        userUnlockInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.alca +=
        userUnlockInfo.reservedProfitALCAUser;
      // proceed to unlock early
      const tx = await fixture.lockup
        .connect(userUnlockInfo.owner)
        .unlockEarly(userUnlockInfo.exitAmount, true);
      // account for used gas
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
    }
    showState(
      "After Unlock early",
      await getState(fixture, lastStakingPosition)
    );
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("should unlock-early user for 100% of initial position in two phases (50%+50%) re-staking unlocked shares", async () => {
    const lastStakingPosition = (
      await fixture.publicStaking.getLatestMintedPositionID()
    ).toNumber();
    const expectedState = await getState(fixture, lastStakingPosition);
    const unlockingEarlyUsers = [2];
    for (let userId = 0; userId < unlockingEarlyUsers.length; userId++) {
      const user = "user" + unlockingEarlyUsers[userId];
      const userShares = ethers.utils.parseEther(
        example.distribution.users[user].shares
      );
      let userUnlockInfo = await getUserUnlockInfo(
        fixture,
        unlockingEarlyUsers[userId],
        userShares,
        50
      );
      // new lockup position is created for user's remaining staking balance
      const newPositionId = (
        await fixture.publicStaking.getLatestMintedPositionID()
      )
        .add(1)
        .toBigInt();
      expectedState.lockupPositions[userUnlockInfo.index.toString()].index =
        userUnlockInfo.index.toBigInt();
      expectedState.lockupPositions[userUnlockInfo.index.toString()].owner =
        userUnlockInfo.owner.address;
      expectedState.lockupPositions[userUnlockInfo.index.toString()].tokenId =
        newPositionId;
      // new position is assigned to user
      expectedState.users[user].tokenId = newPositionId;
      expectedState.users[user].tokenOwner = userUnlockInfo.owner.address;
      // user only earns free ETH from staking since unlocked ALCA is re-staked
      expectedState.contracts.publicStaking.eth -= userUnlockInfo.profitETHUser;
      expectedState.users[user].eth += userUnlockInfo.profitETHUser;
      // user's reserved ETH goes to reward pool
      expectedState.users[user].eth -= userUnlockInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.eth +=
        userUnlockInfo.reservedProfitETHUser;
      // staking's reserved ALCA goes to reward pool
      expectedState.contracts.publicStaking.alca -=
        userUnlockInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.alca +=
        userUnlockInfo.reservedProfitALCAUser;
      // proceed to unlock early
      const tx1 = await fixture.lockup
        .connect(userUnlockInfo.owner)
        .unlockEarly(userUnlockInfo.exitAmount, true);
      // account for used gas
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx1.wait());
      showState(
        "After Unlock early phase 1 ",
        await getState(fixture, lastStakingPosition)
      );
      assert.deepEqual(await getState(fixture), expectedState);
      // fast forward to staking free after
      await mineBlocks(90n);
      userUnlockInfo = await getUserUnlockInfo(
        fixture,
        unlockingEarlyUsers[userId],
        userShares.div(2),
        100
      );
      /*       // user only earns free ETH since ALCA is re-staked
      expectedState.contracts.publicStaking.eth -= userUnlockInfo.profitETHUser;
      expectedState.users[user].eth += userUnlockInfo.profitETHUser */
      /*       // user's reserved ETH goes to reward pool
      expectedState.users[user].eth -= userUnlockInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.eth += userUnlockInfo.reservedProfitETHUser; */
      // user looses lock position
      expectedState.users[user].tokenId = ethers.constants.Zero.toBigInt();
      expectedState.users[user].tokenOwner = ethers.constants.AddressZero;
      // user lockup position is reused with last one
      const lastPosition = (
        await fixture.lockup.getCurrentNumberOfLockedPositions()
      ).toString();
      expectedState.lockupPositions[lastPosition].index =
        userUnlockInfo.index.toBigInt();
      expectedState.lockupPositions[userUnlockInfo.index.toString()] =
        expectedState.lockupPositions[lastPosition];
      // last position is deleted
      delete expectedState.lockupPositions[lastPosition];
      /*       // staking reserved ALCA goes to rewardpool 
      expectedState.contracts.publicStaking.alca -= userUnlockInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.alca += userUnlockInfo.reservedProfitALCAUser;
 */ // proceed with unlock early
      const tx2 = await fixture.lockup
        .connect(userUnlockInfo.owner)
        .unlockEarly(userUnlockInfo.exitAmount, true);
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx2.wait());
      // number of locked positions is decreased
      if (expectedState.contracts.lockup.lockedPositions !== undefined)
        expectedState.contracts.lockup.lockedPositions -= BigInt(
          unlockingEarlyUsers.length
        );
      showState("After Unlock early phase 2", await getState(fixture));
      assert.deepEqual(await getState(fixture), expectedState);
    }
  });

  it("should unlock-early user for 100% of initial position without re-staking unlocked shares", async () => {
    const expectedState = await getState(fixture);
    const lastStakingPosition = (
      await fixture.publicStaking.getLatestMintedPositionID()
    ).toNumber();
    const unlockingEarlyUsers = [3];
    for (let userId = 0; userId < unlockingEarlyUsers.length; userId++) {
      const user = "user" + unlockingEarlyUsers[userId];
      const userShares = ethers.utils.parseEther(
        example.distribution.users[user].shares
      );
      const userUnlockInfo = await getUserUnlockInfo(
        fixture,
        unlockingEarlyUsers[userId],
        userShares,
        100
      );
      // user receives unlocked shares
      expectedState.contracts.publicStaking.alca -= userUnlockInfo.exitAmount;
      expectedState.users[user].alca += userUnlockInfo.exitAmount;
      // user receives ALCA & ETH profits
      expectedState.contracts.publicStaking.alca -=
        userUnlockInfo.profitALCAUser;
      expectedState.users[user].alca += userUnlockInfo.profitALCAUser;
      expectedState.contracts.publicStaking.eth -= userUnlockInfo.profitETHUser;
      expectedState.users[user].eth += userUnlockInfo.profitETHUser;
      // user looses lockup position
      expectedState.users[user].tokenId = ethers.constants.Zero.toBigInt();
      expectedState.users[user].tokenOwner = ethers.constants.AddressZero;
      // user lockup position is reused with last position
      const lastPosition = (
        await fixture.lockup.getCurrentNumberOfLockedPositions()
      ).toString();
      expectedState.lockupPositions[lastPosition].index =
        userUnlockInfo.index.toBigInt();
      expectedState.lockupPositions[userUnlockInfo.index.toString()] =
        expectedState.lockupPositions[lastPosition];
      // last position is deleted
      delete expectedState.lockupPositions[lastPosition];
      // reserved ALCA & ETH goes to reward pool
      expectedState.contracts.rewardPool.alca +=
        userUnlockInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.eth +=
        userUnlockInfo.reservedProfitETHUser;
      expectedState.users[user].alca -= userUnlockInfo.reservedProfitALCAUser;
      expectedState.users[user].eth -= userUnlockInfo.reservedProfitETHUser;
      // proceed with unlock
      const tx = await fixture.lockup
        .connect(userUnlockInfo.owner)
        .unlockEarly(userUnlockInfo.exitAmount, false);
      // accoun for used gas
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
    }
    // unlocked positions is decreased
    if (expectedState.contracts.lockup.lockedPositions !== undefined)
      expectedState.contracts.lockup.lockedPositions -= BigInt(
        unlockingEarlyUsers.length
      );
    showState(
      "After Unlock early",
      await getState(fixture, lastStakingPosition)
    );
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("should unlock-early user for 100% of initial position re-staking unlocked shares", async () => {
    const expectedState = await getState(fixture);
    const lastStakingPosition = (
      await fixture.publicStaking.getLatestMintedPositionID()
    ).toNumber();
    const unlockingEarlyUsers = [3];
    for (let i = 0; i < unlockingEarlyUsers.length; i++) {
      const user = "user" + unlockingEarlyUsers[i];
      const userShares = ethers.utils.parseEther(
        example.distribution.users[user].shares
      );
      const userUnlockInfo = await getUserUnlockInfo(
        fixture,
        unlockingEarlyUsers[i],
        userShares,
        100
      );
      // user only earns free ETH since ALCA is re-staked
      expectedState.contracts.publicStaking.eth -= userUnlockInfo.profitETHUser;
      expectedState.users[user].eth += userUnlockInfo.profitETHUser;
      // user's reserved ETH goes to reward pool
      expectedState.users[user].eth -= userUnlockInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.eth +=
        userUnlockInfo.reservedProfitETHUser;
      // user looses lock position
      expectedState.users[user].tokenId = ethers.constants.Zero.toBigInt();
      expectedState.users[user].tokenOwner = ethers.constants.AddressZero;
      // user lockup position is reused with last one
      const lastPosition = (
        await fixture.lockup.getCurrentNumberOfLockedPositions()
      ).toString();
      expectedState.lockupPositions[lastPosition].index =
        userUnlockInfo.index.toBigInt();
      expectedState.lockupPositions[userUnlockInfo.index.toString()] =
        expectedState.lockupPositions[lastPosition];
      // last position is deleted
      delete expectedState.lockupPositions[lastPosition];
      // staking reserved ALCA goes to rewardpool
      expectedState.contracts.publicStaking.alca -=
        userUnlockInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.alca +=
        userUnlockInfo.reservedProfitALCAUser;
      // proceed with unlock early
      const tx = await fixture.lockup
        .connect(userUnlockInfo.owner)
        .unlockEarly(userUnlockInfo.exitAmount, true);
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
    }
    // number of locked positions is decreased
    if (expectedState.contracts.lockup.lockedPositions !== undefined)
      expectedState.contracts.lockup.lockedPositions -= BigInt(
        unlockingEarlyUsers.length
      );
    showState(
      "After Unlock early",
      await getState(fixture, lastStakingPosition)
    );
    assert.deepEqual(await getState(fixture), expectedState);
  });
});

async function getUserUnlockInfo(
  fixture: Fixture,
  userId: number,
  userShares: BigNumber,
  percentageOfUnlock: number
) {
  const totalShares = await fixture.publicStaking.getTotalShares();
  const signers = await ethers.getSigners();
  const owner_ = signers[userId];
  const tokenId = await fixture.lockup.tokenOf(owner_.address);
  const index_ = await fixture.lockup.getIndexByTokenId(tokenId);
  const exitAmount_ = userShares.div(100).mul(percentageOfUnlock).toBigInt();
  const profitALCAUser_ = profitALCA
    .mul(userShares)
    .div(totalShares)
    .toBigInt();
  const profitETHUser_ = profitETH.mul(userShares).div(totalShares).toBigInt();
  const reservedProfitALCAUser_ = (
    await fixture.lockup.getReservedAmount(profitALCAUser_)
  ).toBigInt();
  const reservedProfitETHUser_ = (
    await fixture.lockup.getReservedAmount(profitETHUser_)
  ).toBigInt();
  return {
    exitAmount: exitAmount_,
    profitALCAUser: profitALCAUser_,
    reservedProfitALCAUser: reservedProfitALCAUser_,
    profitETHUser: profitETHUser_,
    reservedProfitETHUser: reservedProfitETHUser_,
    index: index_,
    owner: owner_,
  };
}
