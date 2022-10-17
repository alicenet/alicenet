import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BytesLike } from "ethers";
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
import { getImpersonatedSigner } from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

const startBlock = 100;
const lockDuration = 2;
const totalBonusAmount = ethers.utils.parseEther("2000000");
let rewardPoolAddress: any;
let asPublicStaking: SignerWithAddress;
let asRewardPool: SignerWithAddress;

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
  asPublicStaking = await getImpersonatedSigner(fixture.publicStaking.address);

  asRewardPool = await getImpersonatedSigner(rewardPool.address);
  return {
    fixture: {
      ...fixture,
      rewardPool,
      lockup,
      bonusPool,
    },
    accounts: signers,
  };
}

describe("Testing Lockup Access Control", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];

  beforeEach(async () => {
    ({ fixture, accounts } = await loadFixture(deployFixture));
  });

  it("BonusPool should not receive ETH from address diffrent that PublicStaking or RewardPool Contract", async () => {
    await expect(
      accounts[0].sendTransaction({
        to: fixture.lockup.address,
        value: ethers.utils.parseEther("1"),
      })
    ).to.be.revertedWithCustomError(
      fixture.lockup,
      "AddressNotAllowedToSendEther"
    );
  });

  it("should receive ETH from PublicStaking contract", async () => {
    await asPublicStaking.sendTransaction({
      to: fixture.lockup.address,
      value: 1,
    });
  });

  it("should receive ETH from RewardPool contract", async () => {
    await asRewardPool.sendTransaction({
      to: fixture.lockup.address,
      value: 1,
    });
  });

  it("should not receive ETH from address diffrent that PublicStaking or RewardPool Contract", async () => {
    await expect(
      accounts[0].sendTransaction({
        to: fixture.bonusPool.address,
        value: ethers.utils.parseEther("1"),
      })
    ).to.be.revertedWithCustomError(
      fixture.lockup,
      "AddressNotAllowedToSendEther"
    );
  });
});
