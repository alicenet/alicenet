import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber, BytesLike } from "ethers";
import { ethers } from "hardhat";
import { CONTRACT_ADDR, DEPLOYED_RAW } from "../../scripts/lib/constants";
import { BonusPool, Lockup, RewardPool } from "../../typechain-types";
import { factory } from "../../typechain-types/contracts/libraries";
import { getEventVar } from "../factory/Setup";
import { BaseTokensFixture, deployFactoryAndBaseTokens, preFixtureSetup } from "../setup";

const startBlock = 100;
const lockDuration = 100;
const stakedAmount = ethers.utils.parseEther("100");
const totalBonusAmount = ethers.utils.parseEther("10000");
async function deployFixture() {
  await preFixtureSetup();
  const signers = await ethers.getSigners();

  const fixture = await deployFactoryAndBaseTokens(signers[0]);
  //deploy lockup contract
  const lockupBase = await ethers.getContractFactory("Lockup");
  const lockupDeployCode = lockupBase.getDeployTransaction(
    startBlock,
    lockDuration,
    totalBonusAmount
  ).data as BytesLike;
  const txResponse = await fixture.factory.deployCreate(lockupDeployCode);
  //get the address from the event
  const lockupAddress = await getEventVar(
    txResponse,
    DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  const lockup = await ethers.getContractAt("Lockup", lockupAddress);
  //get the address of the reward pool from the lockup contract
  const rewardPoolAddress = await lockup.getRewardPoolAddress();
  const rewardPool = await ethers.getContractAt(
    "RewardPool",
    rewardPoolAddress
  );
  //get the address of the bonus pool from the reward pool contract
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);
  let tokenIDs = []
  for(let i = 1; i < 6; i++){
    //transfer 100 ALCA from admin to users
    let txResponse = await fixture.aToken.transfer(signers[i].address, stakedAmount);
    await txResponse.wait();
    //stake the tokens
    txResponse = await fixture.aToken.connect(signers[i]).increaseAllowance(fixture.publicStaking.address, stakedAmount)
    await txResponse.wait()
    txResponse = await fixture.publicStaking.connect(signers[i]).mint(stakedAmount)
    const tokenID = await fixture.publicStaking.tokenOfOwnerByIndex(signers[i].address, 0)
    tokenIDs.push(tokenID)
  }
  
  return {
    fixture: {...fixture,
      rewardPool: rewardPool,
      lockup: lockup,
      bonusPool: bonusPool
    },
    accounts: signers,
    stakedTokenIDs: tokenIDs,
  }
}

describe("lockup", async () => {
  // let admin: SignerWithAddress;
  interface Fixture extends BaseTokensFixture{
    lockup: Lockup;
    rewardPool: RewardPool;
    bonusPool: BonusPool;
  }
  let fixture: Fixture;
  let accounts: SignerWithAddress[]
  let stakedTokenIDs: BigNumber[]
  beforeEach(async () => {
    ({fixture, accounts, stakedTokenIDs} = await loadFixture(deployFixture))
  });

  describe("lockFromApproval",async () => {
    it("approves transfer of nft to lockup, calls lockFromApproval in prelock phase",async () => {
      console.log(await ethers.provider.getBlockNumber())
      //account 1 calls the public staking contract to approve lockup 
      // fixture.publicStaking.connect()
    })
  })

});
