import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert, expect } from "chai";
import { BigNumber, BytesLike } from "ethers";
import { ethers } from "hardhat";
import { CONTRACT_ADDR, DEPLOYED_RAW } from "../../scripts/lib/constants";
import { BonusPool, Lockup, RewardPool } from "../../typechain-types";
import { getEventVar } from "../factory/Setup";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import {
  getEthConsumedAsGas,
  getImpersonatedSigner,
  getSimulatedStakingPositions,
  getState,
  numberOfLockingUsers,
  showState,
  example,
  totalBonusAmount,
  stakedAmount,
  deployLockupContract,
  deployFixture,
  distributeProfits
} from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

let rewardPoolAddress: any;
let asFactory: SignerWithAddress;


describe("Testing Staking Distribution", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];

  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(deployFixture));
  });

  it("attempts to distribute a first round", async () => {
    showState("Initial distribution", await getState(fixture));
    await fixture.aToken
      .connect(accounts[0])
      .increaseAllowance(
        fixture.publicStaking.address,
        ethers.utils.parseEther(example.distribution.profitALCA)
      );
    await fixture.publicStaking.connect(accounts[0]).depositEth(42, {
      value: ethers.utils.parseEther(example.distribution.profitETH),
    });
    await fixture.publicStaking
      .connect(accounts[0])
      .depositToken(
        42,
        ethers.utils.parseEther(example.distribution.profitALCA)
      );
    showState("After distribution 1", await getState(fixture));
    const expectedState = await getState(fixture);
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

  it("attempts to distribute a second round (v2)", async () => {
    showState("Initial distribution", await getState(fixture));
    const expectedState = await getState(fixture);
    await distributeProfits(fixture, accounts[0]);
    expectedState.contracts.publicStaking.eth += ethers.utils
    .parseEther(example.distribution.profitETH)
    .toBigInt();
  expectedState.contracts.publicStaking.alca += ethers.utils
    .parseEther(example.distribution.profitALCA)
    .toBigInt();
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
    await distributeProfits(fixture, accounts[0]);
    expectedState.contracts.publicStaking.eth += ethers.utils
      .parseEther(example.distribution.profitETH)
      .toBigInt();
    expectedState.contracts.publicStaking.alca += ethers.utils
      .parseEther(example.distribution.profitALCA)
      .toBigInt();
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


