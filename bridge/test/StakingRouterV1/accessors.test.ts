import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { BytesLike } from "ethers";
import { ethers } from "hardhat";
import { deployCreateAndRegister } from "../../scripts/lib/alicenetFactory";
import {
  CONTRACT_ADDR,
  EVENT_DEPLOYED_RAW,
  STAKING_ROUTER_V1,
} from "../../scripts/lib/constants";
import {
  BonusPool,
  Lockup,
  RewardPool,
  StakingRouterV1,
} from "../../typechain-types";
import { getEventVar } from "../factory/Setup";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
  stakingRouterV1: StakingRouterV1;
}

const startBlock = 100;
const lockDuration = 2;
const totalBonusAmount = ethers.utils.parseEther("10000");
let rewardPoolAddress: any;

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
  // deploy Lockup
  let contractName = ethers.utils.formatBytes32String("Lockup");
  let txResponse = await fixture.factory.deployCreateAndRegister(
    lockupDeployCode,
    contractName
  );
  // get the address from the event
  const lockupAddress = await getEventVar(
    txResponse,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  // deploy staking router
  await ethers.getContractFactory("StakingRouterV1");
  contractName = ethers.utils.formatBytes32String(STAKING_ROUTER_V1);
  txResponse = await deployCreateAndRegister(
    STAKING_ROUTER_V1,
    fixture.factory,
    ethers,
    [],
    contractName
  );
  // get the address from the event
  const stakingRouterAddress = await getEventVar(
    txResponse,
    EVENT_DEPLOYED_RAW,
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
  // connect and instance of the staking router
  const stakingRouterV1 = await ethers.getContractAt(
    STAKING_ROUTER_V1,
    stakingRouterAddress
  );

  return {
    fixture: {
      ...fixture,
      rewardPool,
      lockup,
      bonusPool,
      stakingRouterV1,
    },
    accounts: signers,
  };
}

describe("StakingRouterV1 - accessors", async () => {
  let fixture: Fixture;

  beforeEach(async () => {
    ({ fixture } = await loadFixture(deployFixture));
  });

  it("getLegacyTokenAddress returns expected address", async () => {
    expect(await fixture.stakingRouterV1.getLegacyTokenAddress()).to.equal(
      await fixture.legacyToken.address
    );
  });
});
