import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import {
  deployFixtureWithoutBonusPosition,
  distributeProfits,
  lockStakedNFT,
  profitALCA,
  profitETH,
} from "./setup";
import { Distribution1 } from "./test.data";

describe("estimateFinalBonusWithProfits", async () => {
  let fixture: any;
  let accounts: SignerWithAddress[];
  let stakedTokenIDs: BigNumber[];
  beforeEach(async () => {
    ({ fixture, accounts, stakedTokenIDs } = await loadFixture(
      deployFixtureWithoutBonusPosition
    ));
  });

  it("Test Prelock before bonusPosition is created", async () => {
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
});
