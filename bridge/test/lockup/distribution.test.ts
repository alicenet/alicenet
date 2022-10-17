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
  getState,
  numberOfLockingUsers,
  showState,
} from "./setup";
import { example } from "./test.data";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

const startBlock = 100;
const lockDuration = 2;
const stakedAmount = ethers.utils.parseEther("100000000").toBigInt();
const totalBonusAmount = ethers.utils.parseEther("2000000");
let rewardPoolAddress: any;
let asFactory: SignerWithAddress;

async function deployFixture() {
  await preFixtureSetup();

  const signers = await ethers.getSigners();
  const fixture = await deployFactoryAndBaseTokens(signers[0]);

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
        example.distribution1.users[user].shares
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
        ethers.utils.parseEther(example.distribution1.profitALCA)
      );
    await fixture.publicStaking.connect(accounts[0]).depositEth(42, {
      value: ethers.utils.parseEther(example.distribution1.profitETH),
    });
    await fixture.publicStaking
      .connect(accounts[0])
      .depositToken(
        42,
        ethers.utils.parseEther(example.distribution1.profitALCA)
      );
    showState("After distribution", await getState(fixture));
    const expectedState = await getState(fixture);
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const user = ("user" + i) as string;
      const tx = await fixture.publicStaking
        .connect(accounts[i])
        .collectAllProfits(stakedTokenIDs[i]);
      expectedState.users[user].alca += ethers.utils
        .parseEther(example.distribution1.users[user].profitALCA)
        .toBigInt();
      expectedState.users[user].eth += ethers.utils
        .parseEther(example.distribution1.users[user].profitETH)
        .toBigInt();
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.contracts.publicStaking.eth -= ethers.utils
        .parseEther(example.distribution1.users[user].profitETH)
        .toBigInt();
      expectedState.contracts.publicStaking.alca -= ethers.utils
        .parseEther(example.distribution1.users[user].profitALCA)
        .toBigInt();
    }
    showState("After collecting", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });

  it.only("attempts to distribute a second round (v2)", async () => {
    showState("Initial distribution", await getState(fixture));
    await fixture.aToken
      .connect(accounts[0])
      .increaseAllowance(
        fixture.publicStaking.address,
        ethers.utils.parseEther(example.distribution1.profitALCA)
      );
    await fixture.publicStaking.connect(accounts[0]).depositEth(42, {
      value: ethers.utils.parseEther(example.distribution1.profitETH),
    });
    await fixture.publicStaking
      .connect(accounts[0])
      .depositToken(
        42,
        ethers.utils.parseEther(example.distribution1.profitALCA)
      );
    showState("After distribution", await getState(fixture));
    const expectedState = await getState(fixture);
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const user = ("user" + i) as string;
      const tx = await fixture.publicStaking
        .connect(accounts[i])
        .collectAllProfits(stakedTokenIDs[i]);
      expectedState.users[user].alca += ethers.utils
        .parseEther(example.distribution1.users[user].profitALCA)
        .toBigInt();
      expectedState.users[user].eth += ethers.utils
        .parseEther(example.distribution1.users[user].profitETH)
        .toBigInt();
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.contracts.publicStaking.eth -= ethers.utils
        .parseEther(example.distribution1.users[user].profitETH)
        .toBigInt();
      expectedState.contracts.publicStaking.alca -= ethers.utils
        .parseEther(example.distribution1.users[user].profitALCA)
        .toBigInt();
    }
    await fixture.aToken
      .connect(accounts[0])
      .increaseAllowance(
        fixture.publicStaking.address,
        ethers.utils.parseEther(example.distribution2.profitALCA)
      );
    await fixture.publicStaking.connect(accounts[0]).depositEth(42, {
      value: ethers.utils.parseEther(example.distribution2.profitETH),
    });
    await fixture.publicStaking
      .connect(accounts[0])
      .depositToken(
        42,
        ethers.utils.parseEther(example.distribution2.profitALCA)
      );
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const user = ("user" + i) as string;
      const tx = await fixture.publicStaking
        .connect(accounts[i])
        .collectAllProfits(stakedTokenIDs[i]);
      expectedState.users[user].alca += ethers.utils
        .parseEther(example.distribution1.users[user].profitALCA)
        .toBigInt();
      expectedState.users[user].eth += ethers.utils
        .parseEther(example.distribution1.users[user].profitETH)
        .toBigInt();
      expectedState.users[user].eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.contracts.publicStaking.eth -= ethers.utils
        .parseEther(example.distribution1.users[user].profitETH)
        .toBigInt();
      expectedState.contracts.publicStaking.alca -= ethers.utils
        .parseEther(example.distribution1.users[user].profitALCA)
        .toBigInt();
    }

    showState("After collecting", await getState(fixture));
    assert.deepEqual(await getState(fixture), expectedState);
  });
});
