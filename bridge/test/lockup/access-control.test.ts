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
import { deployFixture, getImpersonatedSigner } from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

const startBlock = 100;
const lockDuration = 2;
const totalBonusAmount = ethers.utils.parseEther("2000000");
let rewardPoolAddress: any;
let asFactory: SignerWithAddress;
let asPublicStaking: SignerWithAddress;
let asRewardPool: SignerWithAddress;


describe("Testing Lockup Access Control", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];

  beforeEach(async () => {
    ({ fixture, accounts, asFactory, asPublicStaking, asRewardPool } = await loadFixture(deployFixture));
  });

  it("BonusPool should not receive ETH from address different that PublicStaking or RewardPool contracts", async () => {
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


});
