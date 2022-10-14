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

    it("SCALING_FACTOR returns expected value", async () => {
      expect(await fixture.bonusPool.SCALING_FACTOR()).to.equal(
        BigNumber.from("1000000000000000000")
      );
    });

    it("getOriginalSharesLocked returns 0 before bonus rate set", async () => {
      expect(await fixture.bonusPool.getOriginalSharesLocked()).to.equal(0);
    });

    it("getOriginalSharesLocked returns correct value once bonus rate set", async () => {
      const initialTotalLocked = 1234;
      await (
        await fixture.bonusPool
          .connect(fixture.mockLockupSigner)
          .setBonusRate(initialTotalLocked)
      ).wait();

      expect(await fixture.bonusPool.getOriginalSharesLocked()).to.equal(
        initialTotalLocked
      );
    });

    it("getScaledBonusRate returns 0 before bonus rate set", async () => {
      expect(await fixture.bonusPool.getScaledBonusRate()).to.equal(0);
    });

    it("getScaledBonusRate returns correct value once bonus rate set", async () => {
      const initialTotalLocked = 1234;
      await setBonusRate(fixture, initialTotalLocked);

      const expectedBonusRate = fixture.totalBonusAmount
        .mul(await fixture.bonusPool.SCALING_FACTOR())
        .div(initialTotalLocked);

      expect(await fixture.bonusPool.getScaledBonusRate()).to.equal(
        expectedBonusRate
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

  describe("estimateBonusAmount", async () => {
    it("Returns expected amount based on set bonus rate", async () => {
      // set bonus rate
      const initialTotalLocked = 8000;
      await setBonusRate(fixture, initialTotalLocked);

      const shares = BigNumber.from(4000);
      const expectedEstimatedBonusAmount = shares
        .mul(await fixture.bonusPool.getScaledBonusRate())
        .div(await fixture.bonusPool.SCALING_FACTOR());

      const estimatedBonusAmount = await fixture.bonusPool.estimateBonusAmount(
        shares
      );

      expect(estimatedBonusAmount).to.equal(expectedEstimatedBonusAmount);
    });
  });

  describe("estimateBonusAmountWithReward", async () => {
    it("reverts if bonus token not minted", async () => {
      const currentSharesLocked = BigNumber.from(8000);
      const userSharesLocked = BigNumber.from(4000);

      await expect(
        fixture.bonusPool.estimateBonusAmountWithReward(
          currentSharesLocked,
          userSharesLocked
        )
      ).to.be.revertedWithCustomError(
        fixture.bonusPool,
        "BonusTokenNotCreated"
      );
    });

    it("Returns expected amount based on set bonus rate", async () => {
      const tokenId = await mintBonusPosition(accounts, fixture);
      // set bonus rate
      const initialTotalLocked = 8000;
      await setBonusRate(fixture, initialTotalLocked);
      const scalingFactor = await fixture.bonusPool.SCALING_FACTOR();
      const bonusRate = await fixture.bonusPool.getScaledBonusRate();

      const currentSharesLocked = BigNumber.from(8000);
      const userSharesLocked = BigNumber.from(4000);
      const proportion = currentSharesLocked
        .mul(scalingFactor)
        .div(initialTotalLocked);
      const userProportion = userSharesLocked
        .mul(scalingFactor)
        .div(currentSharesLocked);

      const [estimatedPayoutEth, estimatedPayoutToken] =
        await fixture.publicStaking.estimateAllProfits(tokenId);

      console.log("estimatedPayoutEth", estimatedPayoutEth.toString());
      console.log("estimatedPayoutToken", estimatedPayoutToken.toString());

      const totalExpectedBonusRewardEth = proportion
        .mul(estimatedPayoutEth)
        .div(scalingFactor);
      const totalExpectedBonusRewardToken = proportion
        .mul(estimatedPayoutToken)
        .div(scalingFactor);

      const expectedUserBonusShares = bonusRate
        .mul(userSharesLocked)
        .div(scalingFactor);
      const userExpectedBonusRewardEth = userProportion
        .mul(totalExpectedBonusRewardEth)
        .div(scalingFactor);
      const userExpectedBonusRewardToken = userProportion
        .mul(totalExpectedBonusRewardToken)
        .div(scalingFactor);

      const [bonusShares, bonusRewardEth, bonusRewardToken] =
        await fixture.bonusPool.estimateBonusAmountWithReward(
          currentSharesLocked,
          userSharesLocked
        );

      expect(bonusShares).to.equal(expectedUserBonusShares);
      expect(bonusRewardEth).to.equal(userExpectedBonusRewardEth);
      expect(bonusRewardToken).to.equal(userExpectedBonusRewardToken);
    });
  });

  describe("terminate", async () => {
    it("Reverts if called from non lockup address", async () => {
      const finalSharesLocked = 1234;
      await expect(
        fixture.bonusPool.connect(accounts[1]).terminate(finalSharesLocked)
      ).to.be.revertedWithCustomError(fixture.bonusPool, "CallerNotLockup");
    });

    it("Reverts if bonus NFT is not created", async () => {
      const finalSharesLocked = 1234;
      await expect(
        fixture.bonusPool
          .connect(fixture.mockLockupSigner)
          .terminate(finalSharesLocked)
      ).to.be.revertedWithCustomError(
        fixture.bonusPool,
        "BonusTokenNotCreated"
      );
    });
  });
});

async function mintBonusPosition(
  accounts: SignerWithAddress[],
  fixture: Fixture
) {
  const exactStakeAmount = fixture.totalBonusAmount;
  await (
    await fixture.aToken
      .connect(accounts[0])
      .transfer(fixture.bonusPool.address, exactStakeAmount)
  ).wait();

  const receipt = await (
    await fixture.bonusPool
      .connect(fixture.mockFactorySigner)
      .createBonusStakedPosition()
  ).wait();

  const createdEvent = receipt.events?.find(
    (event) => event.event === "BonusPositionCreated"
  );

  return createdEvent?.args?.tokenID;
}

async function setBonusRate(
  fixture: Fixture,
  initialTotalLocked: number
): Promise<void> {
  await (
    await fixture.bonusPool
      .connect(fixture.mockLockupSigner)
      .setBonusRate(initialTotalLocked)
  ).wait();
}
