import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import {
  deployFixture,
  deployFixtureWithoutBonusPosition,
  distributeProfits,
  Fixture,
  jumpToPostLockState,
  lockStakedNFT,
  numberOfLockingUsers,
  originalLockedAmount,
  profitALCA,
  profitETH,
} from "./setup";
import { Distribution1, Distribution2 } from "./test.data";

describe("estimateFinalBonusWithProfits", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];

  describe("Without Bonus Position minted", async () => {
    beforeEach(async () => {
      ({ fixture, accounts, stakedTokenIDs } = await loadFixture(
        deployFixtureWithoutBonusPosition
      ));
    });

    it("returns expected amounts with 1 staker", async () => {
      // lock 1 position
      await lockStakedNFT(fixture, accounts[1], stakedTokenIDs[1]);
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
      const expectedPayoutEth = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitETH)
      );
      const expectedPayoutToken = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitALCA)
      );
      const [positionShares, payoutEth, payoutToken] =
        await fixture.lockup.estimateFinalBonusWithProfits(
          stakedTokenIDs[1],
          false
        );
      expect(positionShares).to.equal(
        ethers.utils.parseEther(Distribution1.users.user1.shares)
      );
      expect(payoutEth).to.equal(expectedPayoutEth);
      expect(payoutToken).to.equal(expectedPayoutToken);
    });

    it("returns expected amounts with 1 staker in precision mode", async () => {
      // lock 1 position
      await lockStakedNFT(fixture, accounts[1], stakedTokenIDs[1]);
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
      const expectedPayoutEth = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitETH)
      );
      const expectedPayoutToken = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitALCA)
      );
      const [positionShares, payoutEth, payoutToken] =
        await fixture.lockup.estimateFinalBonusWithProfits(
          stakedTokenIDs[1],
          true
        );
      expect(positionShares).to.equal(
        ethers.utils.parseEther(Distribution1.users.user1.shares)
      );
      expect(payoutEth).to.equal(expectedPayoutEth);
      expect(payoutToken).to.equal(expectedPayoutToken);
    });

    it("returns expected amounts with multiple stakers", async () => {
      for (let i = 1; i <= numberOfLockingUsers; i++) {
        await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
      }
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
      const expectedPayoutEth = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitETH)
      );
      const expectedPayoutToken = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitALCA)
      );
      const [positionShares, payoutEth, payoutToken] =
        await fixture.lockup.estimateFinalBonusWithProfits(
          stakedTokenIDs[1],
          false
        );
      expect(positionShares).to.equal(
        ethers.utils.parseEther(Distribution1.users.user1.shares)
      );
      expect(payoutEth).to.equal(expectedPayoutEth);
      expect(payoutToken).to.equal(expectedPayoutToken);
    });

    it("returns expected amounts with multiple stakers in precision mode", async () => {
      for (let i = 1; i <= numberOfLockingUsers; i++) {
        await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
      }
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
      const expectedPayoutEth = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitETH)
      );
      const expectedPayoutToken = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitALCA)
      );
      const [positionShares, payoutEth, payoutToken] =
        await fixture.lockup.estimateFinalBonusWithProfits(
          stakedTokenIDs[1],
          true
        );
      expect(positionShares).to.equal(
        ethers.utils.parseEther(Distribution1.users.user1.shares)
      );
      expect(payoutEth).to.equal(expectedPayoutEth);
      expect(payoutToken).to.equal(expectedPayoutToken);
    });

    describe("With rewards in the reward pool", async () => {
      beforeEach(async () => {
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
        }
        await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
      });

      it("returns expected amounts with all users calling collectAllProfits", async () => {
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          await fixture.lockup.connect(accounts[i]).collectAllProfits();
        }

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );
          const [rewardEthProfit, rewardTokenProfit] =
            await fixture.rewardPool.estimateRewards(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              false
            );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution1.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            bonusPayoutEth.add(rewardEthProfit),
            "eth"
          );
          expect(payoutToken).to.equal(
            bonusPayoutToken.add(rewardTokenProfit),
            "token"
          );
        }
      });

      it("returns expected amounts with single user calling collectAllProfits", async () => {
        await fixture.lockup.connect(accounts[1]).collectAllProfits();

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const [positionEthProfit, positionTokenProfit] =
            await fixture.publicStaking.estimateAllProfits(stakedTokenIDs[i]);

          const expectedPayoutEth = await fixture.lockup.getReservedAmount(
            positionEthProfit
          );
          const expectedPayoutToken = await fixture.lockup.getReservedAmount(
            positionTokenProfit
          );

          const [rewardEthProfit, rewardTokenProfit] =
            await fixture.rewardPool.estimateRewards(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              false
            );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution1.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            expectedPayoutEth.add(bonusPayoutEth).add(rewardEthProfit),
            "eth"
          );
          expect(payoutToken).to.equal(
            expectedPayoutToken.add(bonusPayoutToken).add(rewardTokenProfit),
            "token"
          );
        }
      });
      it("returns expected amounts with single user calling collectAllProfits in precise mode", async () => {
        await fixture.lockup.connect(accounts[1]).collectAllProfits();

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const expectedPayoutEth = await fixture.lockup.getReservedAmount(
            ethers.utils.parseEther(Distribution1.users[user].profitETH)
          );
          const expectedPayoutToken = await fixture.lockup.getReservedAmount(
            ethers.utils.parseEther(Distribution1.users[user].profitALCA)
          );

          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              true
            );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution1.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            expectedPayoutEth.add(bonusPayoutEth),
            "eth"
          );
          expect(payoutToken).to.equal(
            expectedPayoutToken.add(bonusPayoutToken),
            "token"
          );
        }
      });
    });
  });

  describe("With Bonus Position minted", async () => {
    beforeEach(async () => {
      ({ fixture, accounts, stakedTokenIDs } = await loadFixture(
        deployFixture
      ));
    });

    it("reverts when called with non existant token id", async () => {
      const tokenId = 1234;
      await expect(fixture.lockup.estimateFinalBonusWithProfits(tokenId, false))
        .to.be.revertedWithCustomError(fixture.lockup, "TokenIDNotLocked")
        .withArgs(tokenId);
    });

    it("reverts when called with a token id that has unlocked", async () => {
      const tokenId = stakedTokenIDs[1];
      await lockStakedNFT(fixture, accounts[1], tokenId);
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
      await jumpToPostLockState(fixture);
      await fixture.lockup.aggregateProfits();
      await fixture.lockup
        .connect(accounts[1])
        .unlock(accounts[1].address, false);

      await expect(fixture.lockup.estimateFinalBonusWithProfits(tokenId, false))
        .to.be.revertedWithCustomError(fixture.lockup, "TokenIDNotLocked")
        .withArgs(tokenId);
    });

    it("returns expected amounts with 1 staker", async () => {
      // lock 1 position
      await lockStakedNFT(fixture, accounts[1], stakedTokenIDs[1]);
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
      const expectedPayoutEth = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitETH)
      );
      const expectedPayoutToken = await fixture.lockup.getReservedAmount(
        ethers.utils.parseEther(Distribution1.users.user1.profitALCA)
      );

      const [bonusPayoutEth, bonusPayoutToken] =
        await fixture.bonusPool.estimateBonusAmountWithReward(
          ethers.utils.parseEther(Distribution1.users.user1.shares),
          ethers.utils.parseEther(Distribution1.users.user1.shares)
        );
      const [positionShares, payoutEth, payoutToken] =
        await fixture.lockup.estimateFinalBonusWithProfits(
          stakedTokenIDs[1],
          false
        );
      expect(positionShares).to.equal(
        ethers.utils.parseEther(Distribution1.users.user1.shares)
      );
      expect(payoutEth).to.equal(expectedPayoutEth.add(bonusPayoutEth), "eth");
      expect(payoutToken).to.equal(
        expectedPayoutToken.add(bonusPayoutToken),
        "token"
      );
    });

    it("returns expected amounts with multiple stakers", async () => {
      for (let i = 1; i <= numberOfLockingUsers; i++) {
        await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
      }
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA);

      for (let i = 1; i <= numberOfLockingUsers; i++) {
        const user = "user" + i;

        const expectedPayoutEth = await fixture.lockup.getReservedAmount(
          ethers.utils.parseEther(Distribution1.users[user].profitETH)
        );
        const expectedPayoutToken = await fixture.lockup.getReservedAmount(
          ethers.utils.parseEther(Distribution1.users[user].profitALCA)
        );

        const [bonusPayoutEth, bonusPayoutToken] =
          await fixture.bonusPool.estimateBonusAmountWithReward(
            originalLockedAmount,
            ethers.utils.parseEther(Distribution1.users[user].shares)
          );
        const [positionShares, payoutEth, payoutToken] =
          await fixture.lockup.estimateFinalBonusWithProfits(
            stakedTokenIDs[i],
            false
          );
        expect(positionShares).to.equal(
          ethers.utils.parseEther(Distribution1.users[user].shares),
          "shares"
        );
        expect(payoutEth).to.equal(
          expectedPayoutEth.add(bonusPayoutEth),
          "eth"
        );
        expect(payoutToken).to.equal(
          expectedPayoutToken.add(bonusPayoutToken),
          "token"
        );
      }
    });

    it("returns expected amounts with multiple stakers in precise mode", async () => {
      for (let i = 1; i <= numberOfLockingUsers; i++) {
        await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
      }
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA);

      for (let i = 1; i <= numberOfLockingUsers; i++) {
        const user = "user" + i;

        const expectedPayoutEth = await fixture.lockup.getReservedAmount(
          ethers.utils.parseEther(Distribution1.users[user].profitETH)
        );
        const expectedPayoutToken = await fixture.lockup.getReservedAmount(
          ethers.utils.parseEther(Distribution1.users[user].profitALCA)
        );

        const [bonusPayoutEth, bonusPayoutToken] =
          await fixture.bonusPool.estimateBonusAmountWithReward(
            originalLockedAmount,
            ethers.utils.parseEther(Distribution1.users[user].shares)
          );
        const [positionShares, payoutEth, payoutToken] =
          await fixture.lockup.estimateFinalBonusWithProfits(
            stakedTokenIDs[i],
            true
          );
        expect(positionShares).to.equal(
          ethers.utils.parseEther(Distribution1.users[user].shares),
          "shares"
        );
        expect(payoutEth).to.equal(
          expectedPayoutEth.add(bonusPayoutEth),
          "eth"
        );
        expect(payoutToken).to.equal(
          expectedPayoutToken.add(bonusPayoutToken),
          "token"
        );
      }
    });

    describe("With rewards in the reward pool", async () => {
      beforeEach(async () => {
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i]);
        }
        await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
      });

      it("returns expected amounts with collectAllProfits", async () => {
        for (let i = 1; i <= numberOfLockingUsers; i++) {
          await fixture.lockup.connect(accounts[i]).collectAllProfits();
        }

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const [positionEthProfit, positionTokenProfit] =
            await fixture.publicStaking.estimateAllProfits(stakedTokenIDs[i]);

          const expectedPayoutEth = await fixture.lockup.getReservedAmount(
            positionEthProfit
          );
          const expectedPayoutToken = await fixture.lockup.getReservedAmount(
            positionTokenProfit
          );

          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );
          const [rewardEthProfit, rewardTokenProfit] =
            await fixture.rewardPool.estimateRewards(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              false
            );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution1.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            expectedPayoutEth.add(bonusPayoutEth).add(rewardEthProfit),
            "eth"
          );
          expect(payoutToken).to.equal(
            expectedPayoutToken.add(bonusPayoutToken).add(rewardTokenProfit),
            "token"
          );
        }
      });

      it("returns expected amounts with 1 collectAllProfits", async () => {
        await fixture.lockup.connect(accounts[1]).collectAllProfits();

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const [positionEthProfit, positionTokenProfit] =
            await fixture.publicStaking.estimateAllProfits(stakedTokenIDs[i]);

          const expectedPayoutEth = await fixture.lockup.getReservedAmount(
            positionEthProfit
          );
          const expectedPayoutToken = await fixture.lockup.getReservedAmount(
            positionTokenProfit
          );

          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [rewardEthProfit, rewardTokenProfit] =
            await fixture.rewardPool.estimateRewards(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              false
            );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution1.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            expectedPayoutEth.add(bonusPayoutEth).add(rewardEthProfit),
            "eth"
          );
          expect(payoutToken).to.equal(
            expectedPayoutToken.add(bonusPayoutToken).add(rewardTokenProfit),
            "token"
          );
        }
      });
      it("returns expected amounts with 1 collectAllProfits in precise mode", async () => {
        await fixture.lockup.connect(accounts[1]).collectAllProfits();

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const expectedPayoutEth = await fixture.lockup.getReservedAmount(
            ethers.utils.parseEther(Distribution1.users[user].profitETH)
          );
          const expectedPayoutToken = await fixture.lockup.getReservedAmount(
            ethers.utils.parseEther(Distribution1.users[user].profitALCA)
          );
          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              true
            );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution1.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            expectedPayoutEth.add(bonusPayoutEth),
            "eth"
          );
          expect(payoutToken).to.equal(
            expectedPayoutToken.add(bonusPayoutToken),
            "token"
          );
        }
      });

      it("returns expected amounts with multiple distributions", async () => {
        const numberOfCollectingUsers = 2;
        let expectedPayoutEth = 0n;
        let expectedPayoutToken = 0n;

        for (let i = 1; i <= numberOfCollectingUsers; i++) {
          const user = "user" + i;
          await fixture.lockup.connect(accounts[i]).collectAllProfits();
          expectedPayoutEth += (
            await fixture.lockup.getReservedAmount(
              ethers.utils.parseEther(Distribution1.users[user].profitETH)
            )
          ).toBigInt();
          expectedPayoutToken += (
            await fixture.lockup.getReservedAmount(
              ethers.utils.parseEther(Distribution1.users[user].profitALCA)
            )
          ).toBigInt();
        }

        // compare with reward pool balances
        const ethInRewardPool = await fixture.rewardPool.getEthReserve();
        const tokenInRewardPool = await fixture.rewardPool.getTokenReserve();

        expect(ethInRewardPool).to.equal(expectedPayoutEth);
        expect(tokenInRewardPool).to.equal(expectedPayoutToken);

        await distributeProfits(fixture, accounts[0], profitETH, profitALCA);

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const [positionEthProfit, positionTokenProfit] =
            await fixture.publicStaking.estimateAllProfits(stakedTokenIDs[i]);

          const userExpectedPayoutEth = await fixture.lockup.getReservedAmount(
            positionEthProfit
          );
          const userExpectedPayoutToken =
            await fixture.lockup.getReservedAmount(positionTokenProfit);

          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution2.users[user].shares)
            );

          const [rewardEthProfit, rewardTokenProfit] =
            await fixture.rewardPool.estimateRewards(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution2.users[user].shares)
            );

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              false
            );

          expect(rewardEthProfit).to.equal(
            (expectedPayoutEth *
              ethers.utils
                .parseEther(Distribution2.users[user].shares)
                .toBigInt()) /
              originalLockedAmount
          );
          expect(rewardTokenProfit).to.equal(
            (expectedPayoutToken *
              ethers.utils
                .parseEther(Distribution2.users[user].shares)
                .toBigInt()) /
              originalLockedAmount
          );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution2.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            userExpectedPayoutEth.add(bonusPayoutEth).add(rewardEthProfit),
            "eth"
          );
          expect(payoutToken).to.equal(
            userExpectedPayoutToken
              .add(bonusPayoutToken)
              .add(rewardTokenProfit),
            "token"
          );
        }
      });

      it("returns expected amounts with multiple distributions in precise mode", async () => {
        await fixture.lockup.connect(accounts[1]).collectAllProfits();
        await fixture.lockup.connect(accounts[2]).collectAllProfits();
        await distributeProfits(fixture, accounts[0], profitETH, profitALCA);

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const expectedPayoutEth = await fixture.lockup.getReservedAmount(
            ethers.utils.parseEther(Distribution2.users[user].profitETH)
          );
          const expectedPayoutToken = await fixture.lockup.getReservedAmount(
            ethers.utils.parseEther(Distribution2.users[user].profitALCA)
          );
          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution2.users[user].shares)
            );

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              true
            );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution1.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            expectedPayoutEth.add(bonusPayoutEth),
            "eth"
          );
          expect(payoutToken).to.equal(
            expectedPayoutToken.add(bonusPayoutToken),
            "token"
          );
        }
      });

      it("returns expected amounts with aggregateProfits", async () => {
        await jumpToPostLockState(fixture);
        await fixture.lockup.aggregateProfits();

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const [positionEthProfit, positionTokenProfit] =
            await fixture.publicStaking.estimateAllProfits(stakedTokenIDs[i]);

          const expectedPayoutEth = await fixture.lockup.getReservedAmount(
            positionEthProfit
          );
          const expectedPayoutToken = await fixture.lockup.getReservedAmount(
            positionTokenProfit
          );

          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );
          const [rewardEthProfit, rewardTokenProfit] =
            await fixture.rewardPool.estimateRewards(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [tempEthProfit, tempTokenProfit] =
            await fixture.lockup.getTemporaryRewardBalance(accounts[i].address);

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              false
            );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution1.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            expectedPayoutEth
              .add(bonusPayoutEth)
              .add(rewardEthProfit)
              .add(tempEthProfit),
            "eth"
          );
          expect(payoutToken).to.equal(
            expectedPayoutToken
              .add(bonusPayoutToken)
              .add(rewardTokenProfit)
              .add(tempTokenProfit),
            "token"
          );
        }
      });

      it("returns expected amounts with aggregateProfits in precise mode", async () => {
        await jumpToPostLockState(fixture);
        await fixture.lockup.aggregateProfits();

        for (let i = 1; i <= numberOfLockingUsers; i++) {
          const user = "user" + i;

          const [positionEthProfit, positionTokenProfit] =
            await fixture.publicStaking.estimateAllProfits(stakedTokenIDs[i]);

          const expectedPayoutEth = await fixture.lockup.getReservedAmount(
            positionEthProfit
          );
          const expectedPayoutToken = await fixture.lockup.getReservedAmount(
            positionTokenProfit
          );

          const [bonusPayoutEth, bonusPayoutToken] =
            await fixture.bonusPool.estimateBonusAmountWithReward(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );
          const [rewardEthProfit, rewardTokenProfit] =
            await fixture.rewardPool.estimateRewards(
              originalLockedAmount,
              ethers.utils.parseEther(Distribution1.users[user].shares)
            );

          const [tempEthProfit, tempTokenProfit] =
            await fixture.lockup.getTemporaryRewardBalance(accounts[i].address);

          const [positionShares, payoutEth, payoutToken] =
            await fixture.lockup.estimateFinalBonusWithProfits(
              stakedTokenIDs[i],
              true
            );
          expect(positionShares).to.equal(
            ethers.utils.parseEther(Distribution1.users[user].shares),
            "shares"
          );
          expect(payoutEth).to.equal(
            expectedPayoutEth
              .add(bonusPayoutEth)
              .add(rewardEthProfit)
              .add(tempEthProfit),
            "eth"
          );
          expect(payoutToken).to.equal(
            expectedPayoutToken
              .add(bonusPayoutToken)
              .add(rewardTokenProfit)
              .add(tempTokenProfit),
            "token"
          );
        }
      });
    });
  });
});
