import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert } from "chai";
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

describe("Testing Staking Distribution", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];

  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(deployFixture));
  });

  it("attempts to distribute a first round [ @skip-on-coverage ]", async () => {
    showState("Initial distribution", await getState(fixture));
    const expectedState = await getState(fixture);
    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    expectedState.contracts.publicStaking.eth += profitETH.toBigInt();
    expectedState.contracts.publicStaking.alca += profitALCA.toBigInt();
    showState("After distribution 1", await getState(fixture));
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const user = ("user" + i) as string;
      const tx = await fixture.publicStaking
        .connect(accounts[i])
        .collectAllProfits(stakedTokenIDs[i]);
      expectedState.users[user].alca += ethers.utils
        .parseEther(example.distribution.users[user].profitALCA)
        .toBigInt();
      expectedState.users[user].eth += ethers.utils
        .parseEther(example.distribution.users[user].profitETH)
        .toBigInt();
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.contracts.publicStaking.eth -= ethers.utils
        .parseEther(example.distribution.users[user].profitETH)
        .toBigInt();
      expectedState.contracts.publicStaking.alca -= ethers.utils
        .parseEther(example.distribution.users[user].profitALCA)
        .toBigInt();
    }
    showState("After collect 1", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("attempts to distribute a second round (v2) [ @skip-on-coverage ]", async () => {
    showState("Initial distribution", await getState(fixture));
    const expectedState = await getState(fixture);
    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    expectedState.contracts.publicStaking.eth += profitETH.toBigInt();
    expectedState.contracts.publicStaking.alca += profitALCA.toBigInt();
    showState("After distribution 1", await getState(fixture));
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const user = ("user" + i) as string;
      const tx = await fixture.publicStaking
        .connect(accounts[i])
        .collectAllProfits(stakedTokenIDs[i]);
      expectedState.users[user].alca += ethers.utils
        .parseEther(example.distribution.users[user].profitALCA)
        .toBigInt();
      expectedState.users[user].eth += ethers.utils
        .parseEther(example.distribution.users[user].profitETH)
        .toBigInt();
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.contracts.publicStaking.eth -= ethers.utils
        .parseEther(example.distribution.users[user].profitETH)
        .toBigInt();
      expectedState.contracts.publicStaking.alca -= ethers.utils
        .parseEther(example.distribution.users[user].profitALCA)
        .toBigInt();
    }
    showState("After collect 1", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
    await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
    expectedState.contracts.publicStaking.eth += profitETH.toBigInt();
    expectedState.contracts.publicStaking.alca += profitALCA.toBigInt();
    showState("After distribution 2", await getState(fixture));
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const user = ("user" + i) as string;
      const tx = await fixture.publicStaking
        .connect(accounts[i])
        .collectAllProfits(stakedTokenIDs[i]);
      expectedState.users[user].alca += ethers.utils
        .parseEther(example.distribution.users[user].profitALCA)
        .toBigInt();
      expectedState.users[user].eth += ethers.utils
        .parseEther(example.distribution.users[user].profitETH)
        .toBigInt();
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.contracts.publicStaking.eth -= ethers.utils
        .parseEther(example.distribution.users[user].profitETH)
        .toBigInt();
      expectedState.contracts.publicStaking.alca -= ethers.utils
        .parseEther(example.distribution.users[user].profitALCA)
        .toBigInt();
    }
    showState("After collect 2", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });
});
