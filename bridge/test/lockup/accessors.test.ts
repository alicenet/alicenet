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
import { jumpToInlockState, jumpToPostLockState } from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
  lockupStartBlock: number;
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
    },
    accounts: signers,
    stakedTokenIDs: tokenIDs,
  };
}

describe("lockup", async () => {
  // let admin: SignerWithAddress;

  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];
  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(deployFixture));
  });

  describe("Public accessors", async () => {
    before(async () => {});

    it("getLockupStartBlock returns correct locking enrollment start block", async () => {
      expect(await fixture.lockup.getLockupStartBlock()).to.be.equal(
        fixture.lockupStartBlock
      );
    });

    it("getLockupEndBlock should return correct locking enrollment end block", async () => {
      const expectedLockupEndBlock = fixture.lockupStartBlock + lockDuration;

      expect(await fixture.lockup.getLockupEndBlock()).to.be.equal(
        expectedLockupEndBlock
      );
    });

    it("getRewardPoolAddress should return correct reward pool address", async () => {
      expect(await fixture.lockup.getRewardPoolAddress()).to.be.equal(
        rewardPoolAddress
      );
    });

    it("getBonusPoolAddress should return correct reward pool address", async () => {
      expect(await fixture.lockup.getBonusPoolAddress()).to.be.equal(
        fixture.bonusPool.address
      );
    });

    it("SCALING_FACTOR should return expected scaling factor", async () => {
      expect(await fixture.lockup.SCALING_FACTOR()).to.be.equal(
        BigNumber.from("1000000000000000000")
      );
    });

    it("FRACTION_RESERVED should return expected scaling factor", async () => {
      const scalingFactor = await fixture.lockup.SCALING_FACTOR();
      const expectedFractionReserved = scalingFactor.div(5);
      expect(await fixture.lockup.FRACTION_RESERVED()).to.be.equal(
        expectedFractionReserved
      );
    });

    it("getReservedPercentage should return correct amount of shares", async () => {
      const scalingFactor = await fixture.lockup.SCALING_FACTOR();
      const fractionReserved = await fixture.lockup.FRACTION_RESERVED();
      const expectedReservedPercentage = BigNumber.from(100)
        .mul(fractionReserved)
        .div(scalingFactor);
      expect(await fixture.lockup.getReservedPercentage()).to.be.equal(
        expectedReservedPercentage
      );
    });

    describe("getState", async () => {
      it("should return PreLock state when in PreLock", async () => {
        const expectedState = 0; // PreLock
        expect(await fixture.lockup.getState()).to.be.equal(expectedState);
      });

      it("should return InLock state when in InLock", async () => {
        await jumpToInlockState(fixture);
        const expectedState = 1; // InLock
        expect(await fixture.lockup.getState()).to.be.equal(expectedState);
      });

      it("should return PostLock state when in PostLock", async () => {
        await jumpToPostLockState(fixture);
        const expectedState = 2; // PostLock
        expect(await fixture.lockup.getState()).to.be.equal(expectedState);
      });
    });

    describe("payoutSafe", async () => {
      it("returns false before profits have been agreggated", async () => {
        expect(await fixture.lockup.payoutSafe()).to.be.equal(false);
      });

      it("returns true after profits have been agreggated", async () => {
        // todo - add test
      });
    });

    describe("with positions locked", async () => {
      beforeEach(async () => {
        // lock the positions
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
        }
      });

      it("ownerOf returns correct owner of token", async () => {
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const account = accounts[i];
          const tokenID = stakedTokenIDs[i];
          expect(await fixture.lockup.ownerOf(tokenID)).to.equal(
            account.address
          );
        }
      });

      it("tokenOf returns correct token id for owner", async () => {
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const account = accounts[i];
          const tokenID = stakedTokenIDs[i];
          expect(await fixture.lockup.tokenOf(account.address)).to.equal(
            tokenID
          );
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
        expect(
          await fixture.lockup.getCurrentNumberOfLockedPositions()
        ).to.equal(numberOfLockingUsers);
      });

      it("getTotalCurrentSharesLocked should return correct amount of shares", async () => {
        const expectedShareAmount =
          stakedAmount * BigNumber.from(numberOfLockingUsers).toBigInt();

        expect(await fixture.lockup.getTotalCurrentSharesLocked()).to.be.equal(
          expectedShareAmount
        );
      });

      it("getOriginalLockedShares should return correct amount of shares", async () => {
        const expectedShareAmount =
          stakedAmount * BigNumber.from(numberOfLockingUsers).toBigInt();

        expect(await fixture.lockup.getOriginalLockedShares()).to.be.equal(
          expectedShareAmount
        );
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
