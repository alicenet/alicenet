import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { BonusPool, Foundation, RewardPool } from "../../typechain-types";

import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import { calculateExpectedProportions, getImpersonatedSigner } from "./setup";

interface Fixture extends BaseTokensFixture {
  rewardPool: RewardPool;
  bonusPool: BonusPool;
  foundation: Foundation;
  lockupAddress: string;
  rewardPoolAddress: string;
  totalBonusAmount: BigNumber;
  mockFactorySigner: SignerWithAddress;
  mockLockupSigner: SignerWithAddress;
  mockBonusPoolSigner: SignerWithAddress;
}

async function deployFixture() {
  await preFixtureSetup();
  const signers = await ethers.getSigners();
  const fixture = await deployFactoryAndBaseTokens(signers[0]);

  const foundation = (await deployUpgradeableWithFactory(
    fixture.factory,
    "Foundation",
    undefined
  )) as Foundation;

  await posFixtureSetup(fixture.factory, fixture.alca);

  // get the address of the reward pool from the lockup contract
  const lockupAddress = signers[5].address;
  const aliceNetFactoryAddress = fixture.factory.address;

  const asFactory = await getImpersonatedSigner(fixture.factory.address);
  const asLockup = await getImpersonatedSigner(lockupAddress);
  const totalBonusAmount = ethers.utils.parseEther("1000000");
  // deploy reward pool
  const rewardPool = await (await ethers.getContractFactory("RewardPool"))
    .connect(asLockup)
    .deploy(fixture.alca.address, aliceNetFactoryAddress, totalBonusAmount);

  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);

  const asBonusPool = await getImpersonatedSigner(bonusPoolAddress);

  return {
    fixture: {
      ...fixture,
      rewardPool,
      bonusPool,
      foundation,
      lockupAddress,
      rewardPoolAddress: rewardPool.address,
      totalBonusAmount,
      mockFactorySigner: asFactory,
      mockLockupSigner: asLockup,
      mockBonusPoolSigner: asBonusPool,
    },
  };
}

describe("RewardPool - public accessors", async () => {
  let fixture: Fixture;

  beforeEach(async () => {
    ({ fixture } = await loadFixture(deployFixture));
  });

  it("getBonusPoolAddress returns bonus pool contract address", async () => {
    expect(await fixture.rewardPool.getBonusPoolAddress()).to.equal(
      fixture.bonusPool.address
    );
  });

  it("getLockupContractAddress returns lockup contract address", async () => {
    expect(await fixture.rewardPool.getLockupContractAddress()).to.equal(
      fixture.lockupAddress
    );
  });

  describe("with deposit", async () => {
    const tokenAmount = ethers.utils.parseEther("123456");
    const ethAmount = ethers.utils.parseEther("789");
    beforeEach(async () => {
      await fixture.rewardPool
        .connect(fixture.mockLockupSigner)
        .deposit(tokenAmount, {
          value: ethAmount,
        });
    });

    it("getTokenReserve returns token reserve amount", async () => {
      expect(await fixture.rewardPool.getTokenReserve()).to.equal(tokenAmount);
    });

    it("getEthReserve returns token eth amount", async () => {
      expect(await fixture.rewardPool.getEthReserve()).to.equal(ethAmount);
    });

    it("estimateRewards returns correct share proportions", async () => {
      const totalShares = BigNumber.from(100);
      const userShares = BigNumber.from(10);

      const [expectedProportionEth, expectedProportionTokens] =
        calculateExpectedProportions(
          ethAmount,
          tokenAmount,
          userShares,
          totalShares
        );
      const [proportionEth, proportionTokens] =
        await fixture.rewardPool.estimateRewards(totalShares, userShares);
      expect(proportionEth).to.equal(expectedProportionEth);
      expect(proportionTokens).to.equal(expectedProportionTokens);
    });

    it("estimateRewards reverts if total shares less than user shares", async () => {
      const totalShares = BigNumber.from(100);
      const userShares = BigNumber.from(101);

      await expect(
        fixture.rewardPool.estimateRewards(totalShares, userShares)
      ).to.be.revertedWithCustomError(
        fixture.rewardPool,
        "InvalidTotalSharesValue"
      );
    });

    it("estimateRewards reverts if total shares is 0", async () => {
      const totalShares = BigNumber.from(0);
      const userShares = BigNumber.from(101);

      await expect(
        fixture.rewardPool.estimateRewards(totalShares, userShares)
      ).to.be.revertedWithCustomError(
        fixture.rewardPool,
        "InvalidTotalSharesValue"
      );
    });
  });
});
