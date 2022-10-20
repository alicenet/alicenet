import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
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
HardhatRuntimeEnvironment.ethers: typeof import("/home/zj/work/alicenet/bridge/node_modules/ethers/lib/ethers") & HardhatEthersHelpers

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

const startBlock = 100;
const lockDuration = 2;
const stakedAmount = ethers.utils.parseEther("100").toBigInt();
const totalBonusAmount = ethers.utils.parseEther("10000");

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
    let contractName = ethers.utils.formatBytes32String("Lockup")
    let txResponse = await fixture.factory.deployCreateAndRegister(lockupDeployCode, contractName);
    // get the address from the event
    const lockupAddress = await getEventVar(
      txResponse,
      DEPLOYED_RAW,
      CONTRACT_ADDR
    );
    //deploy staking router
    const stakingRouterBase = await ethers.getContractFactory("StakingRouterV1")
    const stakingRouterDeployCode = stakingRouterBase.getDeployTransaction().data as BytesLike
    contractName = ethers.utils.formatBytes32String("StakingRouterV1")
    txResponse = await fixture.factory.deployCreateAndRegister(stakingRouterDeployCode, contractName);
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
    const tokenIDs = [];
    for (let i = 1; i <= numberOfLockingUsers; i++) {
      // transfer 100 ALCA from admin to users
      let txResponse = await fixture.aToken
        .connect(signers[0])
        .transfer(signers[i].address, stakedAmount);
      await txResponse.wait();
      // stake the tokens
      txResponse = await fixture.aToken
        .connect(signers[i])
        .increaseAllowance(fixture.publicStaking.address, stakedAmount);
      await txResponse.wait();
      txResponse = await fixture.publicStaking
        .connect(signers[i])
        .mint(stakedAmount);
      const tokenID = await fixture.publicStaking.tokenOfOwnerByIndex(
        signers[i].address,
        0
      );
      tokenIDs[i] = tokenID;
    }
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
  