import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import {
  deployFixture,
  distributeProfits,
  jumpToInlockState,
  jumpToPostLockState,
  lockDuration,
  lockStakedNFT,
  LockupStates,
  numberOfLockingUsers,
  originalLockedAmount,
  profitALCA,
  profitETH,
} from "./setup";
import { Distribution1 } from "./test.data";

xdescribe("Lockup - public accessors", async () => {
  let fixture: any;
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
      fixture.rewardPool.address
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
      expect(await fixture.lockup.getState()).to.be.equal(LockupStates.PreLock);
    });

    it("returns InLock state when lockup start block reached", async () => {
      await jumpToInlockState(fixture);
      expect(await fixture.lockup.getState()).to.be.equal(LockupStates.InLock);
    });

    it("returns PostLock state when lockup end block reached", async () => {
      await jumpToPostLockState(fixture);
      expect(await fixture.lockup.getState()).to.be.equal(
        LockupStates.PostLock
      );
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
      it("returns false before profits have been aggregated", async () => {
        expect(await fixture.lockup.payoutSafe()).to.be.equal(false);
      });

      it("returns true after profits have been aggregated", async () => {
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
        expect(await fixture.lockup.getIndexByTokenId(tokenID)).to.equal(
          BigNumber.from(i)
        );
      }
    });

    describe("getCurrentNumberOfLockedPositions", async () => {
      it("returns correct number of locked positions", async () => {
        expect(
          await fixture.lockup.getCurrentNumberOfLockedPositions()
        ).to.equal(numberOfLockingUsers);
      });

      it("updates amount when positions unlock", async () => {
        await jumpToPostLockState(fixture);
        await fixture.lockup.aggregateProfits();

        const numberOfPositionsToUnlock = 2;
        for (let i = 1; i <= numberOfPositionsToUnlock; i++) {
          await fixture.lockup
            .connect(accounts[i])
            .unlock(accounts[i].address, false);
        }

        expect(
          await fixture.lockup.getCurrentNumberOfLockedPositions()
        ).to.be.equal(numberOfLockingUsers - numberOfPositionsToUnlock);
      });
    });

    describe("getTotalSharesLocked", async () => {
      it("returns correct amount of shares", async () => {
        const expectedShareAmount = originalLockedAmount;

        expect(await fixture.lockup.getTotalSharesLocked()).to.be.equal(
          expectedShareAmount
        );
      });

      it("returns updated amount of shares if positions unlock", async () => {
        await jumpToInlockState(fixture);

        const numberOfPositionsToUnlock = 2;
        let expectedShareAmount = originalLockedAmount;
        for (let i = 1; i <= numberOfPositionsToUnlock; i++) {
          const user = "user" + i;
          await fixture.lockup
            .connect(accounts[i])
            .unlockEarly(Distribution1.users[user].shares, false);
          expectedShareAmount -= BigInt(Distribution1.users[user].shares);
        }

        expect(await fixture.lockup.getTotalSharesLocked()).to.be.equal(
          expectedShareAmount
        );
      });
    });

    describe("getTotalSharesLocked", async () => {
      it("returns correct amount of shares", async () => {
        expect(await fixture.lockup.getTotalSharesLocked()).to.be.equal(
          originalLockedAmount
        );
      });

      it("returns updated amount of shares when positions unlock when in PreLock state", async () => {
        const numberOfPositionsToUnlock = 2;
        let expectedShareAmountUpdated = originalLockedAmount;
        for (let i = 1; i <= numberOfPositionsToUnlock; i++) {
          const user = "user" + i;
          await fixture.lockup
            .connect(accounts[i])
            .unlockEarly(Distribution1.users[user].shares, false);
          expectedShareAmountUpdated -= BigInt(
            Distribution1.users[user].shares
          );
        }

        expect(await fixture.lockup.getTotalSharesLocked()).to.be.equal(
          expectedShareAmountUpdated
        );
      });
    });

    describe("with funds to distribute", async () => {
      beforeEach(async () => {
        await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
        await jumpToPostLockState(fixture);
        await fixture.lockup.aggregateProfits();
      });

      it("getTemporaryRewardBalance returns correct reward balances of eth and tokens", async () => {
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;
          const expectedRewardEth = ethers.utils
            .parseEther(Distribution1.users[user].profitETH)
            .sub(
              await fixture.lockup.getReservedAmount(
                ethers.utils.parseEther(Distribution1.users[user].profitETH)
              )
            );

          const expectedRewardAlca = ethers.utils
            .parseEther(Distribution1.users[user].profitALCA)
            .sub(
              await fixture.lockup.getReservedAmount(
                ethers.utils.parseEther(Distribution1.users[user].profitALCA)
              )
            );
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
    });
  });
});
