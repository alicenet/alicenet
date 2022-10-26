import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert, expect } from "chai";
import { ethers } from "hardhat";
import {
  deployFixture,
  distributeProfits,
  example,
  getImpersonatedSigner,
  getSimulatedStakingPositions,
  getState,
  jumpToPostLockState,
  lockStakedNFT,
  profitALCA,
  profitETH,
} from "./setup";

describe("Testing Staking Distribution", async () => {
  let fixture: any;
  let accounts: SignerWithAddress[];
  beforeEach(async () => {
    ({ fixture, accounts } = await loadFixture(deployFixture));
  });

  it("aggregate profits", async () => {
    const stakedTokenIDs = await getSimulatedStakingPositions(
      fixture,
      accounts,
      5
    );
    for (let i = 1; i <= NUM_USERS; i++) {
      const txResponse = await lockStakedNFT(
        fixture,
        accounts[i],
        stakedTokenIDs[i],
        true
      );
      await txResponse.wait();
      const tokenID = await fixture.lockup.tokenOf(accounts[i].address);
    }

    await jumpToPostLockState(fixture);

    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    let expectedState = await getState(fixture);
    const txResponse = await fixture.lockup.aggregateProfits({
      gasLimit: 10000000,
    });
    await txResponse.wait();
    expectedState = await getState(fixture);
    expect(expectedState.contracts.lockup.lockedPositions).to.eq(
      BigInt(NUM_USERS)
    );
    console.log(expectedState);
    for (let i = 1; i <= NUM_USERS; i++) {
      const user = ("user" + i) as string;
      expectedState.users[user].rewardToken = ethers.utils
        .parseEther(example.distribution.users[user].profitALCA)
        .toBigInt();
      expectedState.users[user].rewardEth = ethers.utils
        .parseEther(example.distribution.users[user].profitETH)
        .toBigInt();
    }

    const currentState = await getState(fixture);
    console.log(currentState);
    const totalLockupReward = currentState.contracts.lockup.alca;
    const totalSharesLocked = await fixture.lockup.getTotalSharesLocked();
    assert.deepEqual(await getState(fixture), expectedState);
  });
  it("creates 100 positions in 100 accounts", async () => {
    const approvalAmount = ethers.utils.parseEther("200000000");
    let txResponse = await fixture.aToken
      .connect(accounts[0])
      .increaseAllowance(fixture.publicStaking.address, approvalAmount);
    await txResponse.wait();
    const transactions: Promise<any>[] = [];
    for (let i = 0; i <= 100; i++) {
      // make a random acount
      const randomAccount = ethers.Wallet.createRandom();
      // ethers.providers.JsonRpcSigner()
      const recipient = await getImpersonatedSigner(randomAccount.address);
      // stake random account
      transactions.push(generateLockedPosition(fixture, accounts, recipient));
    }
    await Promise.allSettled(transactions);
    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    await jumpToPostLockState(fixture);
    txResponse = await fixture.lockup.aggregateProfits({ gasLimit: 10000000 });
    // console.log(await fixture.aToken.balanceOf(fixture.lockup.address))
    await txResponse.wait();
    console.log(await fixture.aToken.balanceOf(fixture.lockup.address));
  });
});

export async function generateLockedPosition(
  fixture: any,
  signers: SignerWithAddress[],
  recipient: SignerWithAddress
) {
  // stake a random amount between 1 and 11 million
  const randomAmount = (Math.random() * 11 + 1) * 100000;
  const stakedAmount = ethers.utils.parseEther(randomAmount.toString(10));
  await fixture.publicStaking
    .connect(signers[0])
    .mintTo(recipient.address, stakedAmount, 0);
  const tokenID = await fixture.publicStaking.tokenOfOwnerByIndex(
    recipient.address,
    0
  );
  await lockStakedNFT(fixture, recipient, tokenID, true);
}
