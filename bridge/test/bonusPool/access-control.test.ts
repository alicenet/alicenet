import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { BonusPool, Foundation } from "../../typechain-types";

import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  deployUpgradeableWithFactory,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import { getImpersonatedSigner } from "./setup";

interface Fixture extends BaseTokensFixture {
  bonusPool: BonusPool;
  foundation: Foundation;
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

  const foundation = (await deployUpgradeableWithFactory(
    fixture.factory,
    "Foundation",
    undefined
  )) as Foundation;

  await posFixtureSetup(fixture.factory, fixture.aToken);

  // get the address of the reward pool from the lockup contract
  const lockupAddress = signers[5].address;
  const aliceNetFactoryAddress = fixture.factory.address;

  const asFactory = await getImpersonatedSigner(fixture.factory.address);
  const asLockup = await getImpersonatedSigner(lockupAddress);

  const totalBonusAmount = ethers.utils.parseEther("1000000");
  // deploy reward pool
  const rewardPool = await (await ethers.getContractFactory("RewardPool"))
    .connect(asLockup)
    .deploy(fixture.aToken.address, aliceNetFactoryAddress, totalBonusAmount);
  // Deploy the bonus pool standalone
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);

  return {
    fixture: {
      ...fixture,
      bonusPool,
      foundation,
      lockupAddress,
      rewardPoolAddress: rewardPool.address,
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

  describe("createBonusStakedPosition", async () => {
    it("Reverts if called from non factory address", async () => {
      await expect(
        fixture.bonusPool.connect(accounts[1]).createBonusStakedPosition()
      )
        .to.be.revertedWithCustomError(fixture.bonusPool, "OnlyFactory")
        .withArgs(accounts[1].address, fixture.factory.address);
    });
  });

  describe("terminate", async () => {
    it("Reverts if called from non lockup address", async () => {
      const initialTotalLocked = 1234;
      const finalSharesLocked = 1234;
      await expect(
        fixture.bonusPool
          .connect(accounts[1])
          .terminate(finalSharesLocked, initialTotalLocked)
      ).to.be.revertedWithCustomError(fixture.bonusPool, "CallerNotLockup");
    });
  });
});
