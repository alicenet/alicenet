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
  jumpToPostLockState,
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

describe("Testing Collect", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];

  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(deployFixture));
    showState("Initial State with staked positions", await getState(fixture));
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    showState("After Locking", await getState(fixture, 0));
    const expectedState = await getState(fixture, 0);

    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    expectedState.contracts.publicStaking.eth += profitETH.toBigInt();
    expectedState.contracts.publicStaking.alca += profitALCA.toBigInt();
    showState("After Distribution", await getState(fixture, 0));
    assert.deepEqual(await getState(fixture, 0), expectedState);
  });

  it("should collect with no previous unlocking", async () => {
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userShares = ethers.utils.parseEther(
        example.distribution.users["user" + i].shares
      );
      const userUnlockInfo = await getUserLockingInfo(
        fixture,
        i,
        userShares,
        100
      );
      const expectedState = await getState(fixture, 0);
      const tx = await fixture.lockup.connect(accounts[i]).collectAllProfits()
      // user receives free unlocked staking amounts
      expectedState.contracts.publicStaking.alca -= userUnlockInfo.profitALCAUser
      expectedState.users["user" + i].alca += userUnlockInfo.profitALCAUser
      expectedState.contracts.publicStaking.eth -= userUnlockInfo.profitETHUser
      expectedState.users["user" + i].eth += userUnlockInfo.profitETHUser
      // reserving locking amounts go to reward pool
      expectedState.users["user" + i].eth -= userUnlockInfo.reservedProfitETHUser
      expectedState.contracts.rewardPool.eth += userUnlockInfo.reservedProfitETHUser
      expectedState.users["user" + i].alca -= userUnlockInfo.reservedProfitALCAUser
      expectedState.contracts.rewardPool.alca += userUnlockInfo.reservedProfitALCAUser
      showState("After Collecting", await getState(fixture, 0));
      // account for used gas
      expectedState.users["user" + i].eth -= getEthConsumedAsGas(await tx.wait());
      assert.deepEqual(await getState(fixture, 0), expectedState);
    }
  });

  it.only("should collect with no previous unlocking", async () => {
    await jumpToPostLockState(fixture);
    await fixture.lockup.aggregateProfits()
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userShares = ethers.utils.parseEther(
        example.distribution.users["user" + i].shares
      );
      await fixture.lockup.connect(accounts[1]).unlock(accounts[1].address,false)
      const userUnlockInfo = await getUserLockingInfo(
        fixture,
        i,
        userShares,
        100
      );
      const expectedState = await getState(fixture, 0);
      const tx = await fixture.lockup.connect(accounts[i]).collectAllProfits()
      // user receives free unlocked staking amounts
      expectedState.contracts.publicStaking.alca -= userUnlockInfo.profitALCAUser
      expectedState.users["user" + i].alca += userUnlockInfo.profitALCAUser
      expectedState.contracts.publicStaking.eth -= userUnlockInfo.profitETHUser
      expectedState.users["user" + i].eth += userUnlockInfo.profitETHUser
      // reserving locking amounts go to reward pool
      expectedState.users["user" + i].eth -= userUnlockInfo.reservedProfitETHUser
      expectedState.contracts.rewardPool.eth += userUnlockInfo.reservedProfitETHUser
      expectedState.users["user" + i].alca -= userUnlockInfo.reservedProfitALCAUser
      expectedState.contracts.rewardPool.alca += userUnlockInfo.reservedProfitALCAUser
      showState("After Collecting", await getState(fixture, 0));
      // account for used gas
      expectedState.users["user" + i].eth -= getEthConsumedAsGas(await tx.wait());
      assert.deepEqual(await getState(fixture, 0), expectedState);
    }
  });

});