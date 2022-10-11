import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { BytesLike } from "ethers";
import { ethers } from "hardhat";
import { CONTRACT_ADDR, DEPLOYED_RAW } from "../../scripts/lib/constants";
import { getEventVar } from "../factory/Setup";
import { deployFactoryAndBaseTokens, Fixture, preFixtureSetup } from "../setup";

const startBlock = 100;
const lockDuration = 100;
const totalBonusAmount = ethers.utils.parseEther("100");
async function deployFixture() {
  await preFixtureSetup();
  const [admin] = await ethers.getSigners();

  const fixture = await deployFactoryAndBaseTokens(admin);
  //deploy lockup contract
  const lockupBase = await ethers.getContractFactory("Lockup");
  const lockupDeployCode = lockupBase.getDeployTransaction(
    startBlock,
    lockDuration,
    totalBonusAmount
  ).data as BytesLike;
  const txResponse = await fixture.factory.deployCreate(lockupDeployCode);
  const lockupAddress = await getEventVar(
    txResponse,
    DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  const lockup = await ethers.getContractAt("Lockup", lockupAddress);
  const rewardPoolAddress = await lockup.getRewardPoolAddress();
  const rewardPool = await ethers.getContractAt(
    "RewardPool",
    rewardPoolAddress
  );
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);
  return {
    ...fixture,
    rewardPool: rewardPool,
    lockup: lockup,
    bonusPool: bonusPool,
  };
}

describe("lockup", async () => {
  // let admin: SignerWithAddress;
  let fixture: Fixture;
  beforeEach(async () => {
    let fixture = await loadFixture(deployFixture);
  });
});
