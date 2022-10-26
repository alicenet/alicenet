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

  it("should collect for user without other user previous unlocking", async () => {
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userShares = ethers.utils.parseEther(
        example.distribution.users["user" + i].shares
      );
      const userLockingInfo = await getUserLockingInfo(
        fixture,
        i,
      );
      const expectedState = await getState(fixture, 0);
      const tx = await fixture.lockup.connect(accounts[i]).collectAllProfits()
      // user receives free unlocked staking amounts
      expectedState.contracts.publicStaking.alca -= userLockingInfo.profitALCAUser
      expectedState.users["user" + i].alca += userLockingInfo.profitALCAUser
      expectedState.contracts.publicStaking.eth -= userLockingInfo.profitETHUser
      expectedState.users["user" + i].eth += userLockingInfo.profitETHUser
      // reserved locking amounts go to reward pool
      expectedState.users["user" + i].eth -= userLockingInfo.reservedProfitETHUser
      expectedState.contracts.rewardPool.eth += userLockingInfo.reservedProfitETHUser
      expectedState.users["user" + i].alca -= userLockingInfo.reservedProfitALCAUser
      expectedState.contracts.rewardPool.alca += userLockingInfo.reservedProfitALCAUser
      showState("After Collecting", await getState(fixture, 0));
      // account for used gas
      expectedState.users["user" + i].eth -= getEthConsumedAsGas(await tx.wait());
      assert.deepEqual(await getState(fixture, 0), expectedState);
    }
  });

  it.skip("should collect for user with other user previous unlocking", async () => {
    const lastStakingPosition = (
      await fixture.publicStaking.getLatestMintedPositionID()
    ).toNumber();

    const userShares = ethers.utils.parseEther(
      example.distribution.users["user1"].shares
    );   await fixture.lockup.connect(accounts[1]).unlockEarly(userShares,false)

    for (let i = 2; i <= numberOfLockingUsers; i++) {
      const userShares = ethers.utils.parseEther(
        example.distribution.users["user" + i].shares
      );
      const userLockingInfo = await getUserLockingInfo(
        fixture,
        i,
      );
      const expectedState = await getState(fixture, 1, lastStakingPosition-2);
      const tx = await fixture.lockup.connect(accounts[i]).collectAllProfits()
      // user receives free unlocked staking amounts
      expectedState.contracts.publicStaking.alca -= userLockingInfo.profitALCAUser
      expectedState.users["user" + i].alca += userLockingInfo.profitALCAUser
      expectedState.contracts.publicStaking.eth -= userLockingInfo.profitETHUser
      expectedState.users["user" + i].eth += userLockingInfo.profitETHUser
      // reserved locking amounts go to reward pool
      expectedState.users["user" + i].eth -= userLockingInfo.reservedProfitETHUser
      expectedState.contracts.rewardPool.eth += userLockingInfo.reservedProfitETHUser
      expectedState.users["user" + i].alca -= userLockingInfo.reservedProfitALCAUser
      expectedState.contracts.rewardPool.alca += userLockingInfo.reservedProfitALCAUser
      showState("After Collecting", await getState(fixture, 0, lastStakingPosition-2));
      // account for used gas
      expectedState.users["user" + i].eth -= getEthConsumedAsGas(await tx.wait());
      assert.deepEqual(await getState(fixture, 2, lastStakingPosition-2), expectedState);
    }
  });

  it("should not collect for user with user previous unlocking", async () => {
     
    const lastStakingPosition = (
      await fixture.publicStaking.getLatestMintedPositionID()
    ).toNumber();

    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userShares = ethers.utils.parseEther(
        example.distribution.users["user" + i].shares
      );
      const userLockingInfo = await getUserLockingInfo(
        fixture,
        i,
      );
      const exitAmount = userLockingInfo.userShares 
      await fixture.lockup.connect(accounts[i]).unlockEarly(exitAmount,false)
      const expectedState = await getState(fixture, undefined, lastStakingPosition - 2);
      await expect(
       fixture.lockup.connect(accounts[i]).collectAllProfits()).to.be.revertedWithCustomError(fixture.lockup, "UserHasNoPosition")
    }
  });

  it.skip("should collect for user with no previous unlocking", async () => {

    await jumpToPostLockState(fixture);
    const lastStakingPosition = (
      await fixture.publicStaking.getLatestMintedPositionID()
    ).toNumber();
    console.log("lastStakingPosition",lastStakingPosition)
    const bonusPoolShares = ethers.utils.parseEther("20000000000000000000000000");
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userShares = ethers.utils.parseEther(
        example.distribution.users["user" + i].shares
      );
      console.log(bonusPoolShares, bonusPoolShares,userShares)
      console.log("estimated!",await fixture.bonusPool.estimateBonusAmountWithReward(bonusPoolShares, bonusPoolShares,userShares ))
      }
      assert(true,false)
    let  expectedState = await getState(fixture,0, lastStakingPosition- 2)
    await fixture.lockup.aggregateProfits()
    /*     await fixture.lockup.connect(accounts[1]).unlock(accounts[1].address,false)
     */



    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userShares = ethers.utils.parseEther(
        example.distribution.users["user" + i].shares
      );
      const userLockingInfo = await getUserLockingInfo(
        fixture,
        i,
        userShares,
        100
      );
       // user receives free unlocked staking amounts
      expectedState.contracts.publicStaking.alca -= userLockingInfo.profitALCAUser
      expectedState.users["user" + i].alca += userLockingInfo.profitALCAUser
      expectedState.contracts.publicStaking.eth -= userLockingInfo.profitETHUser
      expectedState.users["user" + i].eth += userLockingInfo.profitETHUser
/*      // reserving locking amounts go to reward pool
      expectedState.users["user" + i].eth -= userLockingInfo.reservedProfitETHUser
      expectedState.contracts.rewardPool.eth += userLockingInfo.reservedProfitETHUser
      expectedState.users["user" + i].alca -= userLockingInfo.reservedProfitALCAUser
      expectedState.contracts.rewardPool.alca += userLockingInfo.reservedProfitALCAUser */
      // account for used gas
/*       expectedState.users["user" + i].eth -= getEthConsumedAsGas(await tx.wait());
 */    }
 console.log(expectedState.contracts.publicStaking.alca- (await getState(fixture, 0, lastStakingPosition-2)).contracts.publicStaking.alca)

    assert.deepEqual(await getState(fixture), expectedState)
  });


});