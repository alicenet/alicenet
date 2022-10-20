import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert, expect } from "chai";
import { BigNumber, BytesLike, ContractTransaction } from "ethers";
import { ethers } from "hardhat";
import { CONTRACT_ADDR, DEPLOYED_RAW } from "../../scripts/lib/constants";
import {
  BonusPool,
  Foundation,
  Lockup,
  RewardPool,
} from "../../typechain-types";
import { getEventVar } from "../factory/Setup";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import {
  getEthConsumedAsGas,
  getImpersonatedSigner,
  getState,
  jumpToInlockState,
  jumpToPostLockState,
  numberOfLockingUsers,
  showState,
} from "./setup";
import { Distribution1, example1, example2, example3 } from "./test.data";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

const startBlock = 100;
const lockDuration = 100;
const stakedAmount = ethers.utils.parseEther("100000000").toBigInt();
const totalBonusAmount = ethers.utils.parseEther("2000000");
const lockedAmount = ethers.utils.parseEther("20000000").toBigInt();

let rewardPoolAddress: any;
let asFactory: SignerWithAddress;

async function deployFixture() {
  await preFixtureSetup();

  const signers = await ethers.getSigners();
  const fixture = await deployFactoryAndBaseTokens(signers[0]);
  // deploying foundation so terminate doesn't fail
  (await deployUpgradeableWithFactory(
    fixture.factory,
    "Foundation",
    undefined
  )) as Foundation;

  // deploy lockup contract
  const lockupBase = await ethers.getContractFactory("Lockup");
  const lockupDeployCode = lockupBase.getDeployTransaction(
    startBlock,
    lockDuration,
    totalBonusAmount
  ).data as BytesLike;
  const txResponse = await fixture.factory.deployCreate(lockupDeployCode);
  // get the address from the event
  const lockupAddress = await getEventVar(
    txResponse,
    DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  await posFixtureSetup(fixture.factory, fixture.aToken);
  const lockup = await ethers.getContractAt("Lockup", lockupAddress);
  // get the address of the reward pool from the lockup contract
  rewardPoolAddress = await lockup.getRewardPoolAddress();
  const rewardPool = await ethers.getContractAt(
    "RewardPool",
    rewardPoolAddress
  );
  // get the address of the bonus pool from the reward pool contract
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);
  const tokenIDs = [];
  await fixture.aToken
    .connect(signers[0])
    .increaseAllowance(fixture.publicStaking.address, stakedAmount);
  for (let i = 1; i <= numberOfLockingUsers * 10; i++) {
    if (i % 10 === 0) {
      // stake test positions only for tokens 10,20,30,40 & 50
      const index = i / 10;
      const user = ("user" + index) as string;
      const stakedAmount = ethers.utils.parseEther(
        Distribution1.users[user].shares
      );
      await fixture.publicStaking
        .connect(signers[0])
        .mintTo(signers[index].address, stakedAmount, 0);
      const tokenID = await fixture.publicStaking.tokenOfOwnerByIndex(
        signers[index].address,
        0
      );
      tokenIDs[index] = tokenID;
    } else {
      if (i % 2 === 0) {
        // for the rest stake 1M if even
        await fixture.publicStaking
          .connect(signers[0])
          .mintTo(signers[0].address, ethers.utils.parseEther("1000000"), 0);
      } else {
        // or 500K if odd
        await fixture.publicStaking
          .connect(signers[0])
          .mintTo(signers[0].address, ethers.utils.parseEther("500000"), 0);
      }
    }
  }
  asFactory = await getImpersonatedSigner(fixture.factory.address);
  await fixture.aToken
    .connect(signers[0])
    .transfer(bonusPoolAddress, totalBonusAmount);
  await bonusPool.connect(asFactory).createBonusStakedPosition();
  const bonusPosition = await fixture.publicStaking.getPosition(
    await bonusPool.getBonusStakedPosition()
  );
  expect(bonusPosition.shares).to.be.equals(totalBonusAmount);
  const leftOver =
    stakedAmount - (await fixture.publicStaking.getTotalShares()).toBigInt();
  await fixture.publicStaking
    .connect(signers[0])
    .mintTo(signers[0].address, leftOver, 0);
  expect(
    (await fixture.publicStaking.getTotalShares()).toBigInt()
  ).to.be.equals(stakedAmount);

  return {
    fixture: {
      ...fixture,
      rewardPool,
      lockup,
      bonusPool,
    },
    accounts: signers,
    stakedTokenIDs: tokenIDs,
  };
}

async function lockStakedNFT(
  fixture: Fixture,
  account: SignerWithAddress,
  tokenID: BigNumber,
  approve: boolean = true
): Promise<ContractTransaction> {
  if (approve) {
    const txResponse = await fixture.publicStaking
      .connect(account)
      .approve(fixture.lockup.address, tokenID);
    await txResponse.wait();
  }
  return fixture.lockup.connect(account).lockFromApproval(tokenID);
}

async function distributeProfits(fixture: Fixture, admin: SignerWithAddress) {
  await fixture.aToken
    .connect(admin)
    .increaseAllowance(
      fixture.publicStaking.address,
      ethers.utils.parseEther(Distribution1.profitALCA)
    );
  await fixture.publicStaking.connect(admin).depositEth(42, {
    value: ethers.utils.parseEther(Distribution1.profitETH),
  });
  await fixture.publicStaking
    .connect(admin)
    .depositToken(42, ethers.utils.parseEther(Distribution1.profitALCA));
}

describe("Testing Unlock", async () => {
  // let admin: SignerWithAddress;

  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];
  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(deployFixture));
  });

  it("unlock all positions with no early exits", async () => {
    showState("Initial State with staked position", await getState(fixture));
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    await jumpToInlockState(fixture);
    showState("After Locking", await getState(fixture));
    await distributeProfits(fixture, accounts[0]);
    showState("After Distribution", await getState(fixture));
    await distributeProfits(fixture, accounts[0]);
    showState("After Distribution2", await getState(fixture));
    await jumpToPostLockState(fixture);
    await fixture.lockup.aggregateProfits();
    showState("After Aggregate", await getState(fixture));
    const expectedState = await getState(fixture);
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const user = "user" + i;
      const tx = await fixture.lockup
        .connect(accounts[i])
        .unlock(accounts[i].address, false);
      expectedState.users[user].eth += BigNumber.from(
        example1[user].totalEarnedEth
      ).toBigInt();
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.users[user].alca += BigNumber.from(
        example1[user].totalEarnedALCA
      ).toBigInt();
      expectedState.users[user].tokenOf = ethers.constants.Zero.toBigInt();
      expectedState.users[user].ownerOf = ethers.constants.AddressZero;
    }
    // lockup should has to be distributed all the assets
    expectedState.contracts.lockup.alca = 0n;
    expectedState.contracts.lockup.eth = 0n;
    expectedState.contracts.rewardPool.alca = 0n;
    expectedState.contracts.rewardPool.eth = 0n;
    // since all positions were burn, the alca from publicStaking shares should go to the user
    expectedState.contracts.publicStaking.alca -= lockedAmount;
    // Expected state definition
    assert.deepEqual(await getState(fixture), expectedState);
    showState("After Unlocking", await getState(fixture));
  });

  it("unlock all positions with biggest position exiting early", async () => {
    showState("Initial State with staked position", await getState(fixture));
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    await jumpToInlockState(fixture);
    showState("After Locking", await getState(fixture));
    await distributeProfits(fixture, accounts[0]);
    showState("After Distribution", await getState(fixture));
    // user1 exists (unlock) his full position early
    await fixture.lockup
      .connect(accounts[1])
      .unlockEarly(
        ethers.utils.parseEther(Distribution1.users.user1.shares),
        false
      );
    // user1 re-stake the shares again so it gains some of the profits from the latter distributions
    // and also makes the math better to test (since total staked in the publicStaking goes back to
    // 100m alca)
    const leftOver =
      stakedAmount - (await fixture.publicStaking.getTotalShares()).toBigInt();
    await fixture.aToken
      .connect(accounts[1])
      .approve(fixture.publicStaking.address, leftOver);
    await fixture.publicStaking.connect(accounts[1]).mint(leftOver);
    await distributeProfits(fixture, accounts[0]);
    showState("After Distribution2", await getState(fixture));
    await jumpToPostLockState(fixture);
    await fixture.publicStaking
      .connect(accounts[1])
      .collectAllProfits(
        await fixture.publicStaking.tokenOfOwnerByIndex(accounts[1].address, 0)
      );
    await fixture.lockup.aggregateProfits();
    showState("After Aggregate", await getState(fixture));
    const expectedState = await getState(fixture);
    for (let i = 2; i <= numberOfLockingUsers; i++) {
      const user = "user" + i;
      const tx = await fixture.lockup
        .connect(accounts[i])
        .unlock(accounts[i].address, false);
      expectedState.users[user].eth += BigNumber.from(
        example2[user].totalEarnedEth
      ).toBigInt();
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.users[user].alca += BigNumber.from(
        example2[user].totalEarnedALCA
      ).toBigInt();
      expectedState.users[user].tokenOf = ethers.constants.Zero.toBigInt();
      expectedState.users[user].ownerOf = ethers.constants.AddressZero;
    }
    // lockup should has to be distributed all the assets
    expectedState.contracts.lockup.alca = 0n;
    expectedState.contracts.lockup.eth = 0n;
    expectedState.contracts.rewardPool.alca = 0n;
    expectedState.contracts.rewardPool.eth = 0n;
    // since all positions were burn, the alca from publicStaking shares should go to the user
    expectedState.contracts.publicStaking.alca -=
      lockedAmount -
      ethers.utils.parseEther(Distribution1.users.user1.shares).toBigInt();
    // Expected state definition
    showState("After Unlocking", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it("unlock all positions with 2 position exiting early", async () => {
    showState("Initial State with staked position", await getState(fixture));
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    await jumpToInlockState(fixture);
    showState("After Locking", await getState(fixture));
    await distributeProfits(fixture, accounts[0]);
    showState("After Distribution", await getState(fixture));
    // user4 and user5 exists (unlock) their full position early and re-stake so it gains some of
    // the profits from latter the distributions
    await fixture.lockup
      .connect(accounts[4])
      .unlockEarly(
        ethers.utils.parseEther(Distribution1.users.user4.shares),
        false
      );
    await fixture.lockup
      .connect(accounts[5])
      .unlockEarly(
        ethers.utils.parseEther(Distribution1.users.user5.shares),
        false
      );
    // user4 and 5 re-stake the shares again so it gains some of the profits from the latter
    // distributions and also makes the math better to test (since total staked in the publicStaking
    // goes back to 100m alca)
    const shares4 = ethers.utils.parseEther(Distribution1.users.user4.shares);
    await fixture.aToken
      .connect(accounts[4])
      .approve(fixture.publicStaking.address, shares4);
    await fixture.publicStaking.connect(accounts[4]).mint(shares4);
    const shares5 = ethers.utils.parseEther(Distribution1.users.user5.shares);
    await fixture.aToken
      .connect(accounts[5])
      .approve(fixture.publicStaking.address, shares5);
    await fixture.publicStaking.connect(accounts[5]).mint(shares5);
    await distributeProfits(fixture, accounts[0]);
    showState("After Distribution2", await getState(fixture));
    await jumpToPostLockState(fixture);
    await fixture.publicStaking
      .connect(accounts[4])
      .collectAllProfits(
        await fixture.publicStaking.tokenOfOwnerByIndex(accounts[4].address, 0)
      );
    await fixture.publicStaking
      .connect(accounts[5])
      .collectAllProfits(
        await fixture.publicStaking.tokenOfOwnerByIndex(accounts[5].address, 0)
      );
    await fixture.lockup.aggregateProfits();
    showState("After Aggregate", await getState(fixture));
    const expectedState = await getState(fixture);
    // unlock the users that stayed
    for (let i = 1; i <= numberOfLockingUsers - 2; i++) {
      const user = "user" + i;
      const tx = await fixture.lockup
        .connect(accounts[i])
        .unlock(accounts[i].address, false);
      expectedState.users[user].eth += BigNumber.from(
        example3[user].totalEarnedEth
      ).toBigInt();
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.users[user].alca += BigNumber.from(
        example3[user].totalEarnedALCA
      ).toBigInt();
      // workaround to discount integer division errors.
      expectedState.users[user].eth =
        (expectedState.users[user].eth / 100n) * 100n;
      expectedState.users[user].alca =
        (expectedState.users[user].alca / 100n) * 100n;
      expectedState.users[user].tokenOf = ethers.constants.Zero.toBigInt();
      expectedState.users[user].ownerOf = ethers.constants.AddressZero;
    }
    // lockup should has to be distributed all the assets
    expectedState.contracts.lockup.alca = 0n;
    expectedState.contracts.lockup.eth = 0n;
    expectedState.contracts.rewardPool.alca = 0n;
    expectedState.contracts.rewardPool.eth = 0n;
    // since all positions were burn, the alca from publicStaking shares should go to the user
    expectedState.contracts.publicStaking.alca -=
      lockedAmount -
      ethers.utils
        .parseEther(Distribution1.users.user4.shares)
        .add(ethers.utils.parseEther(Distribution1.users.user5.shares))
        .toBigInt();
    // Expected state definition
    showState("After Unlocking", await getState(fixture));
    const currentState = await getState(fixture);
    // workaround to discount integer division errors.
    for (let i = 1; i <= numberOfLockingUsers - 2; i++) {
      const user = "user" + i;
      currentState.users[user].eth =
        (currentState.users[user].eth / 100n) * 100n;
      currentState.users[user].alca =
        (currentState.users[user].alca / 100n) * 100n;
    }
    assert.deepEqual(currentState, expectedState);
  });
});
