import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BytesLike } from "ethers";
import { ethers } from "hardhat";
import { deployCreateAndRegister } from "../../scripts/lib/alicenetFactory";
import {
  CONTRACT_ADDR,
  DEPLOYED_RAW,
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
const stakedAmount = ethers.utils.parseEther("100").toBigInt();
const totalBonusAmount = ethers.utils.parseEther("10000");
const migrationAmount = ethers.utils.parseEther("100");
let rewardPoolAddress: any;
const numberOfLockingUsers = 5;

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
  //deploy Lockup
  let contractName = ethers.utils.formatBytes32String("Lockup");
  let txResponse = await fixture.factory.deployCreateAndRegister(
    lockupDeployCode,
    contractName
  );
  // get the address from the event
  const lockupAddress = await getEventVar(
    txResponse,
    DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  //deploy staking router
  const stakingRouterBase = await ethers.getContractFactory("StakingRouterV1");
  contractName = ethers.utils.formatBytes32String(STAKING_ROUTER_V1);
  txResponse = await deployCreateAndRegister(
    STAKING_ROUTER_V1,
    fixture.factory.address,
    ethers,
    [],
    undefined
  );
  // get the address from the event
  const stakingRouterAddress = await getEventVar(
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
  //connect and instance of the staking router
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

describe("StakingRouterV1", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  beforeEach(async () => {
    ({ fixture, accounts } = await loadFixture(deployFixture));
  });

  describe("migrateand stake", async () => {
    it("successsfully migrate Madtoken and stakes", async () => {
      const tokenOwner = accounts[1];
      let txResponse = await fixture.legacyToken.increaseAllowance(
        fixture.stakingRouterV1.address,
        migrationAmount
      );
      await txResponse.wait();
      txResponse = await fixture.stakingRouterV1.migrateAndStake(
        tokenOwner.address,
        migrationAmount,
        stakedAmount
      );
      const tokenID = await fixture.publicStaking.tokenOfOwnerByIndex(
        tokenOwner.address,
        0
      );
      const position = await fixture.publicStaking.getPosition(tokenID);
      console.log(position);
    });
  });
});
