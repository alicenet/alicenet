import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { assert, expect } from "chai";
import { BigNumber, Signer, Wallet } from "ethers";
import { ethers } from "hardhat";
import { BonusPool, Lockup, RewardPool } from "../../typechain-types";
import { math } from "../../typechain-types/contracts/libraries";
import { BaseTokensFixture, deployFactoryAndBaseTokens, preFixtureSetup } from "../setup";
import {
  deployLockupContract,
  distributeProfits,
  example,
  getEthConsumedAsGas,
  getImpersonatedSigner,
  getSimulatedStakingPositions,
  getState,
  jumpToPostLockState,
  lockStakedNFT,
  numberOfLockingUsers,
  profitALCA,
  profitETH,
  showState,
  stakedAmount,
  totalBonusAmount,
} from "./setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}
const NUM_USERS = 5;

export async function deployFixture() {
  await preFixtureSetup();
  const signers = await ethers.getSigners();
  const baseTokensFixture = await deployFactoryAndBaseTokens(signers[0]);
  const lockup = await deployLockupContract(baseTokensFixture, 1000);
  // get the address of the reward pool from the lockup contract
  const rewardPoolAddress = await lockup.getRewardPoolAddress();
  const rewardPool = await ethers.getContractAt(
    "RewardPool",
    rewardPoolAddress
  );
  // get the address of the bonus pool from the reward pool contract
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);
  const factorySigner = await getImpersonatedSigner(
    baseTokensFixture.factory.address
  );
  const pblicStakingSigner = await getImpersonatedSigner(
    baseTokensFixture.publicStaking.address
  );
  const rewardPoolSigner = await getImpersonatedSigner(rewardPoolAddress);
  const fixture = {
    ...baseTokensFixture,
    rewardPool,
    lockup,
    bonusPool,
    factorySigner,
    pblicStakingSigner,
    rewardPoolSigner,
  };

  return {
    fixture,
    accounts: signers,
    asFactory: factorySigner,
    asPublicStaking: pblicStakingSigner,
    asRewardPool: rewardPoolSigner,
  };
}


describe("Testing Staking Distribution", async () => {
    let fixture: Fixture;
    let accounts: SignerWithAddress[];
    let asFactory:SignerWithAddress
    beforeEach(async () => {
      ({ fixture, accounts, asFactory } = await loadFixture(deployFixture)); 
    });

    it("aggregate profits",async () => {
      const stakedTokenIDs = await getSimulatedStakingPositions(fixture, accounts, 5);
      for(let i = 1; i <= NUM_USERS; i++) {
        let txResponse = await lockStakedNFT(fixture, accounts[i], stakedTokenIDs[i], true);
        await txResponse.wait()
        const tokenID = (await fixture.lockup.tokenOf(accounts[i].address))
      }
      
      await jumpToPostLockState(fixture)
      
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA);
      let expectedState = await getState(fixture);
      let txResponse = await fixture.lockup.aggregateProfits({gasLimit:10000000});
      await txResponse.wait();
      expectedState = await getState(fixture);
      expect(expectedState.contracts.lockup.lockedPositions).to.eq(BigInt(NUM_USERS))
      console.log(expectedState)
      for(let i = 1; i <= NUM_USERS; i++){
        const user = ("user" + i) as string;
        expectedState.users[user].rewardToken = ethers.utils
          .parseEther(example.distribution.users[user].profitALCA)
          .toBigInt();
        expectedState.users[user].rewardEth = ethers.utils
          .parseEther(example.distribution.users[user].profitETH)
          .toBigInt();
      }
      
      const currentState = await getState(fixture)
      console.log(currentState)
      const totalLockupReward = currentState.contracts.lockup.alca
      const totalSharesLocked = await fixture.lockup.getTotalCurrentSharesLocked()
      assert.deepEqual(await getState(fixture), expectedState);  
    })
    it("creates 100 positions in 100 accounts",async () => {
      const approvalAmount = ethers.utils.parseEther("200000000")
      let txResponse= await fixture.aToken.connect(accounts[0]).increaseAllowance(fixture.publicStaking.address, approvalAmount);
      await txResponse.wait();
      let transactions: Promise<any>[] = []
      for(let i = 0; i <= 100; i++){
        //make a random acount 
        const randomAccount = ethers.Wallet.createRandom()
        // ethers.providers.JsonRpcSigner()
        const recipient = await getImpersonatedSigner(randomAccount.address)
        //stake random account 
         transactions.push(generateLockedPosition(fixture, accounts, recipient))
      }
      await Promise.allSettled(transactions)
      await distributeProfits(fixture, accounts[0], profitETH, profitALCA)
      await jumpToPostLockState(fixture)
      txResponse = await fixture.lockup.aggregateProfits({gasLimit:10000000});
      // console.log(await fixture.aToken.balanceOf(fixture.lockup.address))
      await txResponse.wait()
      console.log(await fixture.aToken.balanceOf(fixture.lockup.address))
    })
})  

export async function generateLockedPosition(
  fixture: Fixture,
  signers: SignerWithAddress[],
  recipient: SignerWithAddress,

) {
    //stake a random amount between 1 and 11 million
    const randomAmount = ((Math.random() * 11) + 1) * 100000 
    const stakedAmount = ethers.utils.parseEther(randomAmount.toString(10));
    await fixture.publicStaking
      .connect(signers[0])
      .mintTo(recipient.address, stakedAmount, 0);
    const tokenID = await fixture.publicStaking.tokenOfOwnerByIndex(
      recipient.address,
      0
    );
    await lockStakedNFT(fixture, recipient, tokenID, true)
  }
