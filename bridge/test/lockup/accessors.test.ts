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
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import {
  distributeProfits,
  getImpersonatedSigner,
  jumpToInlockState,
  jumpToPostLockState,
  profitALCA,
  profitETH,
} from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
  lockupStartBlock: number;
  mockFactorySigner: SignerWithAddress;
}

const enrollmentPeriod = 100;
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
    enrollmentPeriod,
    lockDuration,
    totalBonusAmount
  ).data as BytesLike;
  const contractName = ethers.utils.formatBytes32String("Lockup");
  const txResponse = await fixture.factory.deployCreateAndRegister(
    lockupDeployCode,
    contractName
  );

  // get block number from tx
  const tx = await txResponse.wait();
  const blockNumber = tx.blockNumber;

  const lockupStartBlock = blockNumber + enrollmentPeriod;
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
  const asFactory = await getImpersonatedSigner(fixture.factory.address);

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
      lockupStartBlock,
      mockFactorySigner: asFactory,
    },
    accounts: signers,
    stakedTokenIDs: tokenIDs,
  };
}

describe("Lockup - public accessors", async () => {
  // let admin: SignerWithAddress;

  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];
  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(deployFixture));
  });

  it("getLockupStartBlock returns correct locking enrollment start block", async () => {
    expect(await fixture.lockup.getLockupStartBlock()).to.be.equal(
      fixture.lockupStartBlock
    );
  });

  it("getLockupEndBlock returns correct locking enrollment end block", async () => {
    const expectedLockupEndBlock = fixture.lockupStartBlock + lockDuration;

    expect(await fixture.lockup.getLockupEndBlock()).to.be.equal(
      expectedLockupEndBlock
    );
  });

  it("getRewardPoolAddress returns correct reward pool address", async () => {
    expect(await fixture.lockup.getRewardPoolAddress()).to.be.equal(
      rewardPoolAddress
    );
  });

  it("getBonusPoolAddress returns correct reward pool address", async () => {
    expect(await fixture.lockup.getBonusPoolAddress()).to.be.equal(
      fixture.bonusPool.address
    );
  });

  it("SCALING_FACTOR returns expected scaling factor", async () => {
    expect(await fixture.lockup.SCALING_FACTOR()).to.be.equal(
      BigNumber.from("1000000000000000000")
    );
  });

  it("FRACTION_RESERVED returns expected scaling factor", async () => {
    const scalingFactor = await fixture.lockup.SCALING_FACTOR();
    const expectedFractionReserved = scalingFactor.div(5);
    expect(await fixture.lockup.FRACTION_RESERVED()).to.be.equal(
      expectedFractionReserved
    );
  });

  it("getReservedPercentage returns correct amount of shares", async () => {
    const scalingFactor = await fixture.lockup.SCALING_FACTOR();
    const fractionReserved = await fixture.lockup.FRACTION_RESERVED();
    const expectedReservedPercentage = BigNumber.from(100)
      .mul(fractionReserved)
      .div(scalingFactor);
    expect(await fixture.lockup.getReservedPercentage()).to.be.equal(
      expectedReservedPercentage
    );
  });

  it("getReservedAmount returns correct amount based on amount passed", async () => {
    const scalingFactor = await fixture.lockup.SCALING_FACTOR();
    const fractionReserved = await fixture.lockup.FRACTION_RESERVED();
    const amount = ethers.utils.parseEther("1337");
    const expectedReservedPercentage = amount
      .mul(fractionReserved)
      .div(scalingFactor);
    expect(await fixture.lockup.getReservedAmount(amount)).to.be.equal(
      expectedReservedPercentage
    );
  });

  describe("getState", async () => {
    it("returns PreLock state when in PreLock", async () => {
      const expectedState = 0; // PreLock
      expect(await fixture.lockup.getState()).to.be.equal(expectedState);
    });

    it("returns InLock state when in InLock", async () => {
      await jumpToInlockState(fixture);
      const expectedState = 1; // InLock
      expect(await fixture.lockup.getState()).to.be.equal(expectedState);
    });

    it("returns PostLock state when in PostLock", async () => {
      await jumpToPostLockState(fixture);
      const expectedState = 2; // PostLock
      expect(await fixture.lockup.getState()).to.be.equal(expectedState);
    });
  });

  describe("with positions locked", async () => {
    beforeEach(async () => {
      // lock the positions
      for (let i = 1; i <= numberOfLockingUsers; i++) {
        await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
      }
    });

    describe("payoutSafe", async () => {
      it("returns false before profits have been agreggated", async () => {
        expect(await fixture.lockup.payoutSafe()).to.be.equal(false);
      });

      it("returns true after profits have been agreggated", async () => {
        await createBonusStakedPosition(fixture, accounts[0]);
        await jumpToPostLockState(fixture);
        await fixture.lockup.aggregateProfits();
        expect(await fixture.lockup.payoutSafe()).to.be.equal(true);
      });
    });

    it("ownerOf returns correct owner of token", async () => {
      for (let i = 1; i <= numberOfLockingUsers; i++) {
        const account = accounts[i];
        const tokenID = stakedTokenIDs[i];
        expect(await fixture.lockup.ownerOf(tokenID)).to.equal(account.address);
      }
    });

    it("tokenOf returns correct token id for owner", async () => {
      for (let i = 1; i <= numberOfLockingUsers; i++) {
        const account = accounts[i];
        const tokenID = stakedTokenIDs[i];
        expect(await fixture.lockup.tokenOf(account.address)).to.equal(tokenID);
      }
    });

    it("getPositionByIndex returns correct token id", async () => {
      for (let i = 1; i <= numberOfLockingUsers; i++) {
        const tokenID = stakedTokenIDs[i];
        expect(await fixture.lockup.getPositionByIndex(i)).to.equal(tokenID);
      }
    });

    it("getIndexByTokenId returns correct index for token id", async () => {
      for (let i = 1; i <= numberOfLockingUsers; i++) {
        const tokenID = stakedTokenIDs[i];
        expect(await fixture.lockup.getPositionByIndex(tokenID)).to.equal(i);
      }
    });

    it("getCurrentNumberOfLockedPositions returns correct number of locked positions", async () => {
      expect(await fixture.lockup.getCurrentNumberOfLockedPositions()).to.equal(
        numberOfLockingUsers
      );
    });

    it("getTotalCurrentSharesLocked returns correct amount of shares", async () => {
      const expectedShareAmount =
        stakedAmount * BigNumber.from(numberOfLockingUsers).toBigInt();

      expect(await fixture.lockup.getTotalCurrentSharesLocked()).to.be.equal(
        expectedShareAmount
      );
    });

    it("getOriginalLockedShares returns correct amount of shares", async () => {
      const expectedShareAmount =
        stakedAmount * BigNumber.from(numberOfLockingUsers).toBigInt();

      expect(await fixture.lockup.getOriginalLockedShares()).to.be.equal(
        expectedShareAmount
      );
    });

    describe("with funds to distribute", async () => {
      beforeEach(async () => {
        await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
        await createBonusStakedPosition(fixture, accounts[0]);
        await jumpToPostLockState(fixture);
        await fixture.lockup.aggregateProfits();
      });

      it("getTemporaryRewardBalance returns correct reward balances of eth and tokens", async () => {
        const expectedRewardEth = BigNumber.from("8000000000000000000");
        const expectedRewardAlca = BigNumber.from("80000000000000000000000");
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const account = accounts[i];

          const [payoutEth, payoutToken] =
            await fixture.lockup.getTemporaryRewardBalance(account.address);
          expect(payoutEth).to.equal(expectedRewardEth);
          expect(payoutToken).to.equal(expectedRewardAlca);
        }
      });

      it("estimateProfits returns amounts that can be collected from locked positions", async () => {
        const scalingFactor = await fixture.lockup.SCALING_FACTOR();
        const fractionReserved = await fixture.lockup.FRACTION_RESERVED();
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const tokenID = stakedTokenIDs[i];
          const [estimatedPayoutEth, estimatedPayoutToken] =
            await fixture.publicStaking.estimateAllProfits(tokenID);
          const reservedEth = estimatedPayoutEth
            .mul(fractionReserved)
            .div(scalingFactor);
          const reservedToken = estimatedPayoutToken
            .mul(fractionReserved)
            .div(scalingFactor);

          const expectedPayoutEth = estimatedPayoutEth.sub(reservedEth);
          const expectedPayoutToken = estimatedPayoutToken.sub(reservedToken);

          const [payoutEth, payoutToken] = await fixture.lockup.estimateProfits(
            tokenID
          );
          expect(payoutEth).to.equal(expectedPayoutEth);
          expect(payoutToken).to.equal(expectedPayoutToken);
        }
      });

      it("estimateFinalBonusWithProfits returns amounts that can be collected from locked positions", async () => {
        // todo: add test for this
      });
    });
  });
});

async function lockStakedNFT(
  fixture: Fixture,
  account: SignerWithAddress,
  tokenID: BigNumber
): Promise<ContractTransaction> {
  return fixture.publicStaking
    .connect(account)
    ["safeTransferFrom(address,address,uint256,bytes)"](
      account.address,
      fixture.lockup.address,
      tokenID,
      "0x"
    );
}

async function createBonusStakedPosition(
  fixture: Fixture,
  account: SignerWithAddress
) {
  await (
    await fixture.aToken
      .connect(account)
      .transfer(
        fixture.bonusPool.address,
        BigNumber.from("10000000000000000000000")
      )
  ).wait();

  await fixture.bonusPool
    .connect(fixture.mockFactorySigner)
    .createBonusStakedPosition();
}
