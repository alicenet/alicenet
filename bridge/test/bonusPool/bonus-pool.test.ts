import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { BonusPool } from "../../typechain-types";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import { getImpersonatedSigner } from "./setup";

interface Fixture extends BaseTokensFixture {
  bonusPool: BonusPool;
  lockupAddress: string;
  rewardPoolAddress: string;
  totalBonusAmount: BigNumber;
  mockFactorySigner: SignerWithAddress;
  mockLockupSigner: SignerWithAddress;
}

async function deployFixture() {
  await preFixtureSetup();
  const signers = await ethers.getSigners();
  const fixture = await deployFactoryAndBaseTokens(signers[0]);

  await posFixtureSetup(fixture.factory, fixture.aToken);

  // get the address of the reward pool from the lockup contract
  const rewardPoolAddress = signers[4].address;
  const lockupAddress = signers[5].address;
  const aliceNetFactoryAddress = fixture.factory.address;
  const totalBonusAmount = ethers.utils.parseEther("1000000");
  // Deploy the bonus pool standalone
  const bonusPool = await (
    await ethers.getContractFactory("BonusPool")
  ).deploy(
    aliceNetFactoryAddress,
    lockupAddress,
    rewardPoolAddress,
    totalBonusAmount
  );

  const asFactory = await getImpersonatedSigner(fixture.factory.address);
  const asLockup = await getImpersonatedSigner(lockupAddress);

  return {
    fixture: {
      ...fixture,
      bonusPool,
      lockupAddress,
      rewardPoolAddress,
      totalBonusAmount,
      mockFactorySigner: asFactory,
      mockLockupSigner: asLockup,
    },
    accounts: signers,
  };
}

describe("BonusPool", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];

  beforeEach(async () => {
    ({ fixture, accounts } = await loadFixture(deployFixture));
  });

  describe("getters", async () => {
    it("getLockupContractAddress returns lockup contract address", async () => {
      expect(await fixture.bonusPool.getLockupContractAddress()).to.equal(
        fixture.lockupAddress
      );
    });

    it("getRewardPoolAddress returns lockup contract address", async () => {
      expect(await fixture.bonusPool.getRewardPoolAddress()).to.equal(
        fixture.rewardPoolAddress
      );
    });
  });

  describe("createBonusStakedPosition", async () => {
    it("Reverts if called from non factory address", async () => {
      await expect(
        fixture.bonusPool.connect(accounts[1]).createBonusStakedPosition()
      )
        .to.be.revertedWithCustomError(fixture.bonusPool, "OnlyFactory")
        .withArgs(accounts[1].address, fixture.factory.address);
    });

    it("Reverts if insufficient ALCA to stake", async () => {
      const shortStakeAmount = fixture.totalBonusAmount.sub(1);
      await (
        await fixture.aToken
          .connect(accounts[0])
          .transfer(fixture.bonusPool.address, shortStakeAmount)
      ).wait();

      await expect(
        fixture.bonusPool
          .connect(fixture.mockFactorySigner)
          .createBonusStakedPosition()
      )
        .to.be.revertedWithCustomError(
          fixture.bonusPool,
          "NotEnoughALCAToStake"
        )
        .withArgs(shortStakeAmount, fixture.totalBonusAmount);
    });
    it("Succeeds if called from factory address and has enough ALCA", async () => {
      const exactStakeAmount = fixture.totalBonusAmount;
      await (
        await fixture.aToken
          .connect(accounts[0])
          .transfer(fixture.bonusPool.address, exactStakeAmount)
      ).wait();

      const expectedTokenId = 1;
      await expect(
        fixture.bonusPool
          .connect(fixture.mockFactorySigner)
          .createBonusStakedPosition()
      )
        .to.emit(fixture.bonusPool, "BonusPositionCreated")
        .withArgs(expectedTokenId);
    });

    it("Reverts if bonus position already created", async () => {
      const exactStakeAmount = fixture.totalBonusAmount;
      await (
        await fixture.aToken
          .connect(accounts[0])
          .transfer(fixture.bonusPool.address, exactStakeAmount)
      ).wait();

      await (
        await fixture.bonusPool
          .connect(fixture.mockFactorySigner)
          .createBonusStakedPosition()
      ).wait();

      await expect(
        fixture.bonusPool
          .connect(fixture.mockFactorySigner)
          .createBonusStakedPosition()
      ).to.be.revertedWithCustomError(
        fixture.bonusPool,
        "BonusTokenAlreadyCreated"
      );
    });
  });
  describe("setBonusRate", async () => {
    it("Reverts if called from non lockup address", async () => {
      const initialTotalLocked = 1234;
      await expect(
        fixture.bonusPool.connect(accounts[1]).setBonusRate(initialTotalLocked)
      ).to.be.revertedWithCustomError(fixture.bonusPool, "CallerNotLockup");
    });

    it("Succeeds when called from lockup address", async () => {
      const initialTotalLocked = 1234;
      await expect(
        fixture.bonusPool
          .connect(fixture.mockLockupSigner)
          .setBonusRate(initialTotalLocked)
      ).to.not.be.reverted;
    });
    it("Reverts if called from lockup address when bonus already set", async () => {
      const initialTotalLocked = 1234;
      await (
        await fixture.bonusPool
          .connect(fixture.mockLockupSigner)
          .setBonusRate(initialTotalLocked)
      ).wait();

      await expect(
        fixture.bonusPool
          .connect(fixture.mockLockupSigner)
          .setBonusRate(initialTotalLocked)
      ).to.be.revertedWithCustomError(fixture.bonusPool, "BonusRateAlreadySet");
    });
  });
});
