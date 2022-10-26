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
  lockStakedNFT,
  numberOfLockingUsers,
  originalLockedAmount,
  profitALCA,
  profitETH,
} from "./setup";
import { Distribution1 } from "./test.data";

describe("estimateFinalBonusWithProfits", async () => {
  let fixture: Fixture;
  let accounts: SignerWithAddress[];
  let asFactory: SignerWithAddress;
  let stakedTokenIDs: BigNumber[];

  describe("Without Bonus Position minted", async () => {
    beforeEach(async () => {
      ({ fixture, accounts, stakedTokenIDs, asFactory } = await loadFixture(
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
        await fixture.lockup.estimateFinalBonusWithProfits(stakedTokenIDs[1]);
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
        await fixture.lockup.estimateFinalBonusWithProfits(stakedTokenIDs[1]);
      expect(positionShares).to.equal(
        ethers.utils.parseEther(Distribution1.users.user1.shares)
      );
      expect(payoutEth).to.equal(expectedPayoutEth);
      expect(payoutToken).to.equal(expectedPayoutToken);
    });
  });

  describe("With Bonus Position minted", async () => {
    beforeEach(async () => {
      ({ fixture, accounts, stakedTokenIDs, asFactory } = await loadFixture(
        deployFixture
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

      const [bonusPayoutEth, bonusPayoutToken] =
        await fixture.bonusPool.estimateBonusAmountWithReward(
          ethers.utils.parseEther(Distribution1.users.user1.shares),
          ethers.utils.parseEther(Distribution1.users.user1.shares)
        );
      const [positionShares, payoutEth, payoutToken] =
        await fixture.lockup.estimateFinalBonusWithProfits(stakedTokenIDs[1]);
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
          await fixture.lockup.estimateFinalBonusWithProfits(stakedTokenIDs[i]);
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
