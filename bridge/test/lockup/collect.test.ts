import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert, expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { BonusPool, Lockup, RewardPool } from "../../typechain-types";
import { BaseTokensFixture } from "../setup";
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
    showState("After Locking", await getState(fixture));
    const expectedState = await getState(fixture);

    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    expectedState.contracts.publicStaking.eth += profitETH.toBigInt();
    expectedState.contracts.publicStaking.alca += profitALCA.toBigInt();
    showState("After Distribution", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("should collect for user without previous unlocking of another user", async () => {
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userLockingInfo = await getUserLockingInfo(fixture, i);
      const expectedState = await getState(fixture);
      const tx = await fixture.lockup.connect(accounts[i]).collectAllProfits();
      // user receives free unlocked staking amounts
      expectedState.contracts.publicStaking.alca -=
        userLockingInfo.profitALCAUser;
      expectedState.users["user" + i].alca += userLockingInfo.profitALCAUser;
      expectedState.contracts.publicStaking.eth -=
        userLockingInfo.profitETHUser;
      expectedState.users["user" + i].eth += userLockingInfo.profitETHUser;
      // reserved locking amounts go to reward pool
      expectedState.users["user" + i].eth -=
        userLockingInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.eth +=
        userLockingInfo.reservedProfitETHUser;
      expectedState.users["user" + i].alca -=
        userLockingInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.alca +=
        userLockingInfo.reservedProfitALCAUser;
      showState("After Collecting", await getState(fixture));
      // account for used gas
      expectedState.users["user" + i].eth -= getEthConsumedAsGas(
        await tx.wait()
      );
      assert.deepEqual(await getState(fixture), expectedState);
    }
  });

  it.skip("should collect for user with previous unlocking of other user", async () => {
    const userShares = ethers.utils.parseEther(
      example.distribution.users.user1.shares
    );
    await fixture.lockup.connect(accounts[1]).unlockEarly(userShares, false);
    for (let i = 2; i <= numberOfLockingUsers; i++) {
      const userLockingInfo = await getUserLockingInfo(fixture, i);
      const expectedState = await getState(fixture);
      const tx = await fixture.lockup.connect(accounts[i]).collectAllProfits();
      // user receives free unlocked staking amounts
      expectedState.contracts.publicStaking.alca -=
        userLockingInfo.profitALCAUser;
      expectedState.users["user" + i].alca += userLockingInfo.profitALCAUser;
      expectedState.contracts.publicStaking.eth -=
        userLockingInfo.profitETHUser;
      expectedState.users["user" + i].eth += userLockingInfo.profitETHUser;
      // reserved locking amounts go to reward pool
      expectedState.users["user" + i].eth -=
        userLockingInfo.reservedProfitETHUser;
      expectedState.contracts.rewardPool.eth +=
        userLockingInfo.reservedProfitETHUser;
      expectedState.users["user" + i].alca -=
        userLockingInfo.reservedProfitALCAUser;
      expectedState.contracts.rewardPool.alca +=
        userLockingInfo.reservedProfitALCAUser;
      showState("After Collecting", await getState(fixture));
      // account for used gas
      expectedState.users["user" + i].eth -= getEthConsumedAsGas(
        await tx.wait()
      );
      assert.deepEqual(await getState(fixture), expectedState);
    }
  });

  it("should not collect for user with previous unlocking of the same user ", async () => {
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userLockingInfo = await getUserLockingInfo(fixture, i);
      const exitAmount = userLockingInfo.userShares;
      await fixture.lockup.connect(accounts[i]).unlockEarly(exitAmount, false);
      await expect(
        fixture.lockup.connect(accounts[i]).collectAllProfits()
      ).to.be.revertedWithCustomError(fixture.lockup, "UserHasNoPosition");
    }
  });

  it.skip("should collect for user with no previous unlocking", async () => {
    await jumpToPostLockState(fixture);
    const lastStakingPosition = (
      await fixture.publicStaking.getLatestMintedPositionID()
    ).toNumber();
    console.log("lastStakingPosition", lastStakingPosition);
    const bonusPoolShares = ethers.utils.parseEther(
      "20000000000000000000000000"
    );
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userShares = ethers.utils.parseEther(
        example.distribution.users["user" + i].shares
      );
      console.log(bonusPoolShares, bonusPoolShares, userShares);
      console.log(
        "estimated!",
        await fixture.bonusPool.estimateBonusAmountWithReward(
          bonusPoolShares,
          bonusPoolShares,
          userShares
        )
      );
    }
    const expectedState = await getState(fixture);
    await fixture.lockup.aggregateProfits();
    /*     await fixture.lockup.connect(accounts[1]).unlock(accounts[1].address,false)
     */

    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userLockingInfo = await getUserLockingInfo(fixture, i);

      // user receives free unlocked staking amounts
      expectedState.contracts.publicStaking.alca -=
        userLockingInfo.profitALCAUser;
      expectedState.users["user" + i].alca += userLockingInfo.profitALCAUser;
      expectedState.contracts.publicStaking.eth -=
        userLockingInfo.profitETHUser;
      expectedState.users["user" + i].eth += userLockingInfo.profitETHUser;
      /*      // reserving locking amounts go to reward pool
      expectedState.users["user" + i].eth -= userLockingInfo.reservedProfitETHUser
      expectedState.contracts.rewardPool.eth += userLockingInfo.reservedProfitETHUser
      expectedState.users["user" + i].alca -= userLockingInfo.reservedProfitALCAUser
      expectedState.contracts.rewardPool.alca += userLockingInfo.reservedProfitALCAUser */
      // account for used gas
      /*       expectedState.users["user" + i].eth -= getEthConsumedAsGas(await tx.wait());
       */
    }
    console.log(
      expectedState.contracts.publicStaking.alca -
        (await getState(fixture)).contracts.publicStaking.alca
    );

    assert.deepEqual(await getState(fixture), expectedState);
  });
});
