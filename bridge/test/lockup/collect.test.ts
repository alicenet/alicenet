import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert, expect } from "chai";
import { BigNumber } from "ethers";
import { BonusPool, Lockup, RewardPool } from "../../typechain-types";
import { BaseTokensFixture } from "../setup";
import {
  deployFixture,
  distributeProfits,
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

xdescribe("Testing Collect", async () => {
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

  it("should collect for calling user", async () => {
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

      //  Add estimateProfits in Lockup assertion
      showState("After Collecting", await getState(fixture));
      // account for used gas
      expectedState.users["user" + i].eth -= getEthConsumedAsGas(
        await tx.wait()
      );
      assert.deepEqual(await getState(fixture), expectedState);
    }
  });

  it("should not collect for calling user with previous unlocking-early", async () => {
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const userLockingInfo = await getUserLockingInfo(fixture, i);
      const exitAmount = userLockingInfo.userInitialShares;
      await fixture.lockup.connect(accounts[i]).unlockEarly(exitAmount, false);
      await expect(
        fixture.lockup.connect(accounts[i]).collectAllProfits()
      ).to.be.revertedWithCustomError(fixture.lockup, "UserHasNoPosition");
    }
  });

  it("should not collect in post lock state", async () => {
    await jumpToPostLockState(fixture);
    await expect(
      fixture.lockup.connect(accounts[1]).collectAllProfits()
    ).to.be.revertedWithCustomError(fixture.lockup, "PostLockStateNotAllowed");
  });
});
