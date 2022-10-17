import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert, expect } from "chai";
import { BigNumber, BytesLike, ContractTransaction } from "ethers";
import { ethers } from "hardhat";
import { CONTRACT_ADDR, DEPLOYED_RAW } from "../../scripts/lib/constants";
import { BonusPool, Lockup, RewardPool } from "../../typechain-types";
import { getEventVar } from "../factory/Setup";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  mineBlocks,
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
const lockDuration = 100;
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

describe("Testing Unlock", async () => {
  // let admin: SignerWithAddress;

  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];
  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(deployFixture));
  });

  it("attempts to totally unlock early with no rewards", async () => {
    showState("Initial State with staked position", await getState(fixture));
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
    }
    let blocksToMine = (await fixture.lockup.getLockupStartBlock())
      .sub(await ethers.provider.getBlockNumber())
      .toBigInt();
    await mineBlocks(blocksToMine + 1n);
    expect(await ethers.provider.getBlockNumber()).to.be.equals(
      (await fixture.lockup.getLockupStartBlock()).add(1)
    );
    showState("After Locking", await getState(fixture));
    await distributeProfits(fixture, accounts[0]);
    showState("After Distribution", await getState(fixture));
    await distributeProfits(fixture, accounts[0]);
    showState("After Distribution2", await getState(fixture));
    blocksToMine = (await fixture.lockup.getLockupEndBlock())
      .sub(await ethers.provider.getBlockNumber())
      .toBigInt();
    await mineBlocks(blocksToMine + 1n);
    expect(await ethers.provider.getBlockNumber()).to.be.equals(
      (await fixture.lockup.getLockupEndBlock()).add(1)
    );
    await fixture.bonusPool;
    await fixture.lockup.aggregateProfits();
    showState("After Aggregate", await getState(fixture));
    const expectedState = await getState(fixture);
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      const tx = await fixture.lockup
        .connect(accounts[i])
        .unlock(accounts[i].address, false);
      expectedState.users.user1.eth -= getEthConsumedAsGas(await tx.wait());
      expectedState.users.user1.alca += stakedAmount;
      expectedState.users.user1.tokenOf = ethers.constants.Zero.toBigInt();
      expectedState.users.user1.ownerOf = ethers.constants.AddressZero;
    }
    // Expected state definition
    assert.deepEqual(await getState(fixture), expectedState);
    showState("After Unlocking", await getState(fixture));
  });
});

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
      ethers.utils.parseEther(example.distribution1.profitALCA)
    );
  await fixture.publicStaking.connect(admin).depositEth(42, {
    value: ethers.utils.parseEther(example.distribution1.profitETH),
  });
  await fixture.publicStaking
    .connect(admin)
    .depositToken(
      42,
      ethers.utils.parseEther(example.distribution1.profitALCA)
    );
}
