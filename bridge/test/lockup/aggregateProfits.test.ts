import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert, expect } from "chai";
import { ContractTransaction } from "ethers";
import { ethers } from "hardhat";
import {
  deployFixtureForAggregateProfits,
  distributeProfits,
  example,
  Fixture,
  getImpersonatedSigner,
  getSimulatedStakingPositions,
  getState,
  jumpToPostLockState,
  lockStakedNFT,
  numberOfLockingUsers,
  profitALCA,
  profitETH,
  showState,
  showVariable,
  totalBonusAmount,
} from "./setup";

describe("Testing Staking Distribution", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let asFactory: SignerWithAddress;
  beforeEach(async () => {
    ({ fixture, accounts, asFactory } = await loadFixture(
      deployFixtureForAggregateProfits
    ));
  });

  it("aggregate profits", async () => {
    const stakedTokenIDs = await getSimulatedStakingPositions(
      fixture,
      accounts,
      5,
      true
    );
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const txResponse = await lockStakedNFT(
        fixture,
        accounts[i],
        stakedTokenIDs[i]
      );
      await txResponse.wait();
    }

    let currentState = await getState(fixture);
    showState("state after lockup", currentState);
    const totalSharesLocked = await fixture.lockup.getTotalSharesLocked();

    showVariable("totalshares at lockup", totalSharesLocked);

    await jumpToPostLockState(fixture);
    currentState = await getState(fixture);
    showState("state before distribute profit", currentState);
    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    currentState = await getState(fixture);
    showState("state after distribute profit", currentState);
    const expectedState = await getState(fixture);
    currentState = await getState(fixture);
    showState("state after distribute profit", currentState);
    const [bonusEth, bonusToken] =
      await fixture.bonusPool.estimateBonusAmountWithReward(1, 1);
    const txResponse = await fixture.lockup.aggregateProfits({
      gasLimit: 10000000,
    });
    await txResponse.wait();
    currentState = await getState(fixture);
    showState("state after aggregate profit", currentState);
    expect(expectedState.contracts.lockup.lockedPositions).to.eq(
      BigInt(numberOfLockingUsers)
    );
    // update the bonus pool balance
    let lockupEthBalance = BigInt(0);
    let lockupALCABalanca = BigInt(0);
    let rewardPoolEthBalance = BigInt(0);
    let rewardPoolALCABalance = BigInt(0);
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const user = "user" + i;
      const expectedALCA = ethers.utils
        .parseEther(example.distribution.users[user].profitALCA)
        .toBigInt();
      const expectedETH = ethers.utils
        .parseEther(example.distribution.users[user].profitETH)
        .toBigInt();
      const expectedETHReward = (expectedETH * BigInt(8)) / BigInt(10);
      const expectedALCAReward = (expectedALCA * BigInt(8)) / BigInt(10);
      rewardPoolALCABalance += expectedALCA - expectedALCAReward;
      rewardPoolEthBalance += expectedETH - expectedETHReward;
      lockupEthBalance += expectedETHReward;
      lockupALCABalanca += expectedALCAReward;
      // accumulate lockup contract balances
      expectedState.users[user].rewardToken = expectedALCAReward;
      expectedState.users[user].rewardEth = expectedETHReward;
    }
    rewardPoolALCABalance += bonusToken.toBigInt();
    rewardPoolEthBalance += bonusEth.toBigInt();
    const totalPayoutETH = lockupEthBalance + rewardPoolEthBalance;
    const totalPayoutALCA = lockupALCABalanca + rewardPoolALCABalance;
    expectedState.contracts.publicStaking.alca -= totalPayoutALCA;
    expectedState.contracts.publicStaking.eth -= totalPayoutETH;
    // const totalLockupReward = currentState.contracts.lockup.alca
    // TODO update test to reflect new bonus splits

    expectedState.contracts.lockup.alca = lockupALCABalanca;
    expectedState.contracts.lockup.eth = lockupEthBalance;
    expectedState.contracts.rewardPool.alca += rewardPoolALCABalance;
    expectedState.contracts.rewardPool.eth += rewardPoolEthBalance;
    showState("final expected state", expectedState);
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("attempt to call aggregate profit without a staked bonus position", async () => {
    const approvalAmount = ethers.utils.parseEther("200000000");
    // approve public staking to spend account 1 alca
    const txResponse = await fixture.aToken
      .connect(accounts[0])
      .increaseAllowance(fixture.publicStaking.address, approvalAmount);
    await txResponse.wait();
    // mint one position and lock
    await generateLockedPosition(fixture, accounts, accounts[1]);
    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    await jumpToPostLockState(fixture);
    // call aggregate profit with 10 mil gas limit
    await expect(
      fixture.lockup.aggregateProfits({ gasLimit: 10000000 })
    ).to.revertedWithCustomError(fixture.bonusPool, "BonusTokenNotCreated");
  });

  it("creates 100 positions in 100 accounts", async () => {
    const approvalAmount = ethers.utils.parseEther("200000000");
    // approve public staking to spend account 1 alca
    let txResponse = await fixture.aToken
      .connect(accounts[0])
      .increaseAllowance(fixture.publicStaking.address, approvalAmount);
    await txResponse.wait();
    txResponse = await fixture.aToken
      .connect(accounts[0])
      .transfer(fixture.bonusPool.address, totalBonusAmount);
    await txResponse.wait();
    txResponse = await fixture.bonusPool
      .connect(asFactory)
      .createBonusStakedPosition();
    await txResponse.wait();
    // mint 100 positions to 100 random accounts
    const transactions: Promise<ContractTransaction>[] = [];
    const addresses: string[] = [];
    for (let i = 0; i <= 100; i++) {
      // make a random acount
      const randomAccount = ethers.Wallet.createRandom();
      addresses.push(randomAccount.address);
      const recipient = await getImpersonatedSigner(randomAccount.address);
      // stake random account
      transactions.push(generateLockedPosition(fixture, accounts, recipient));
    }
    for (let i = 0; i < transactions.length; i++) {
      const txResponse = await transactions[i];
      await txResponse.wait();
    }
    // let currentState = await getState(fixture)
    // showState("after lockup", currentState)
    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    await jumpToPostLockState(fixture);
    // call aggregate profit with 10 mil gas limit
    txResponse = await fixture.lockup.aggregateProfits({ gasLimit: 10000000 });
    await txResponse.wait();
    const payoutState = await fixture.lockup.payoutSafe();
    expect(payoutState).to.eq(false);
    // showState("after first aggregate", currentState);
    const lockupALCABalance = await fixture.aToken.balanceOf(
      fixture.lockup.address
    );
    const lockupEthBalance = await ethers.provider.getBalance(
      fixture.lockup.address
    );
    const expectedALCA = ethers.utils.parseEther(
      example.distribution.profitALCA
    );
    const expectedETH = ethers.utils.parseEther(example.distribution.profitETH);

    expect(lockupALCABalance).to.be.lessThan(expectedALCA);
    expect(lockupEthBalance).to.be.lessThan(expectedETH);
    txResponse = await fixture.lockup.aggregateProfits({ gasLimit: 6000000 });
    await txResponse.wait();
    // check if the last position got paid
    const [ethbal, alcabal] = await fixture.lockup.getTemporaryRewardBalance(
      addresses[99]
    );
    expect(ethbal.toBigInt()).to.be.greaterThan(BigInt(0));
    expect(alcabal.toBigInt()).to.be.greaterThan(BigInt(0));
    // check if bonus pool balances zero out
  });
});

export async function generateLockedPosition(
  fixture: any,
  signers: SignerWithAddress[],
  recipient: SignerWithAddress
) {
  // stake a random amount between 1 and 11 million
  const randomAmount = (Math.random() * 11 + 1) * 100000;
  const stakedAmount = await ethers.utils.parseEther(randomAmount.toString(10));
  await fixture.publicStaking
    .connect(signers[0])
    .mintTo(recipient.address, stakedAmount, 0);
  const tokenID = await fixture.publicStaking.tokenOfOwnerByIndex(
    recipient.address,
    0
  );
  return lockStakedNFT(fixture, recipient, tokenID);
}
