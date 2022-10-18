import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber, BytesLike } from "ethers";
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
import { deployFixture, distributeProfits, getImpersonatedSigner, jumpToInlockState, jumpToPostLockState, lockStakedNFT, State } from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

describe("Testing Lockup Access Control", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let asFactory: SignerWithAddress;
  let asPublicStaking: SignerWithAddress;
  let asRewardPool: SignerWithAddress;
  let stakedTokenIDs: BigNumber[] = []

  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs, asFactory, asPublicStaking, asRewardPool } = await loadFixture(deployFixture));
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

  describe("Testing onlyPreLock functions", async () => {

    it("attempts to use onERC721Received", async () => {
      expect(await fixture.lockup.getState()).to.be.equals(State.PreLock)
      await expect(
        fixture.lockup.connect(asPublicStaking).onERC721Received(
          ethers.constants.AddressZero,
          ethers.constants.AddressZero,
          0,
          []
        )
      ).to.be.revertedWith("ERC721: invalid token ID");
      await jumpToInlockState(fixture);
      await expect(
        fixture.lockup.connect(asPublicStaking).onERC721Received(
          ethers.constants.AddressZero,
          ethers.constants.AddressZero,
          0,
          []
        )
      ).to.be.revertedWithCustomError(fixture.lockup, "PreLockStateRequired");
      await jumpToPostLockState(fixture);
      await expect(
        fixture.lockup.connect(asPublicStaking).onERC721Received(
          ethers.constants.AddressZero,
          ethers.constants.AddressZero,
          0,
          []
        )
      ).to.be.revertedWithCustomError(fixture.lockup, "PreLockStateRequired");

    });

    it("attempts to use lockFromTransfer", async () => {

      expect(await fixture.lockup.getState()).to.be.equals(State.PreLock)
      await expect(
        fixture.lockup.lockFromTransfer(
          0,
          ethers.constants.AddressZero,
        )
      ).to.be.revertedWith("ERC721: invalid token ID");
      await jumpToInlockState(fixture);
      await expect(
        fixture.lockup.lockFromTransfer(
          0,
          ethers.constants.AddressZero,
        )
      ).to.be.revertedWithCustomError(fixture.lockup, "PreLockStateRequired");
      await jumpToPostLockState(fixture);
      await expect(
        fixture.lockup.lockFromTransfer(
          0,
          ethers.constants.AddressZero,
        )
      ).to.be.revertedWithCustomError(fixture.lockup, "PreLockStateRequired");
    });

  });

  describe("Testing excludePreLock functions", async () => {

    it("attempts to use estimateFinalBonusWithProfits", async () => {
      expect(await fixture.lockup.getState()).to.be.equals(State.PreLock)
      await expect(
        fixture.lockup.estimateFinalBonusWithProfits(
          0)
      ).to.be.revertedWithCustomError(fixture.lockup, "PreLockStateNotAllowed");
      await jumpToInlockState(fixture);
      await expect(
        fixture.lockup.estimateFinalBonusWithProfits(
          0
        )
      ).to.be.revertedWithCustomError(fixture.lockup, "TokenIDNotLocked");
      await jumpToPostLockState(fixture);
      await expect(
        fixture.lockup.estimateFinalBonusWithProfits(
          0
        )
      ).to.be.revertedWithCustomError(fixture.lockup, "TokenIDNotLocked");
    });

  });

  describe("Testing onlyPostLock functions", async () => {

    it("attempts to use aggregateProfits", async () => {
      expect(await fixture.lockup.getState()).to.be.equals(State.PreLock)
      await expect(
        fixture.lockup.aggregateProfits(
        )
      ).to.be.revertedWithCustomError(fixture.lockup, "PostLockStateRequired");
      await jumpToInlockState(fixture);
      await expect(
        fixture.lockup.aggregateProfits(
        )
      ).to.be.revertedWithCustomError(fixture.lockup, "PostLockStateRequired");
      await jumpToPostLockState(fixture);
      await expect(
        fixture.lockup.aggregateProfits(
        )
      ).to.be.revertedWithCustomError(fixture.bonusPool, "InvalidOriginalSharesValue");

    });

    it("attempts to use unlock", async () => {
      expect(await fixture.lockup.getState()).to.be.equals(State.PreLock)
      await expect(
        fixture.lockup.unlock(ethers.constants.AddressZero, false)
      ).to.be.revertedWithCustomError(fixture.lockup, "PostLockStateRequired");
      await jumpToInlockState(fixture);
      await expect(
        fixture.lockup.unlock(ethers.constants.AddressZero, false)

      ).to.be.revertedWithCustomError(fixture.lockup, "PostLockStateRequired");
      await jumpToPostLockState(fixture);
      await expect(
        fixture.lockup.unlock(ethers.constants.AddressZero, false)
      ).to.be.revertedWithCustomError(fixture.lockup, "PayoutUnsafe");
    });

  });


  describe("Testing onlyPayoutSafe functions", async () => {

    it("attempts to use unlock", async () => {
      await lockStakedNFT(fixture, accounts[1], stakedTokenIDs[1]);
      await jumpToPostLockState(fixture);
       await expect(
        fixture.lockup.unlock(ethers.constants.AddressZero, true)
      ).to.be.revertedWithCustomError(fixture.lockup, "PayoutUnsafe");
       await
        fixture.lockup.aggregateProfits()
      await expect(
        fixture.lockup.unlock(ethers.constants.AddressZero, true)
      ).to.be.revertedWithCustomError(fixture.lockup, "UserHasNoPosition");
    });

  });

  describe("Testing onlyPayoutUnSafe functions", async () => {

    it("attempts to use aggregateProfits", async () => {
      await lockStakedNFT(fixture, accounts[1], stakedTokenIDs[1]);
      await jumpToPostLockState(fixture);
       await 
        fixture.lockup.aggregateProfits()

      await expect(
        fixture.lockup.aggregateProfits()
      ).to.be.revertedWithCustomError(fixture.lockup, "PayoutSafe");
    });

  });

});


