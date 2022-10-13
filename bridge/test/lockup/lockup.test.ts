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
  mineBlocks,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";

interface Fixture extends BaseTokensFixture {
  lockup: Lockup;
  rewardPool: RewardPool;
  bonusPool: BonusPool;
}

const startBlock = 100;
const lockDuration = 100;
const stakedAmount = ethers.utils.parseEther("100");
const totalBonusAmount = ethers.utils.parseEther("10000");

async function deployFixture() {
  await preFixtureSetup();
  const signers = await ethers.getSigners();
  const fixture = await deployFactoryAndBaseTokens(signers[0]);
  // deploy lockup contract
  const lockupBase = await ethers.getContractFactory("Lockup");
  const lockupDeployCode = lockupBase.getDeployTransaction(
    startBlock,
    lockDuration,
    totalBonusAmount
  ).data as BytesLike;
  const txResponse = await fixture.factory.deployCreate(lockupDeployCode);
  // get the address from the event
  const lockupAddress = await getEventVar(
    txResponse,
    DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  await posFixtureSetup(fixture.factory, fixture.aToken);
  const lockup = await ethers.getContractAt("Lockup", lockupAddress);
  // get the address of the reward pool from the lockup contract
  const rewardPoolAddress = await lockup.getRewardPoolAddress();
  const rewardPool = await ethers.getContractAt(
    "RewardPool",
    rewardPoolAddress
  );
  // get the address of the bonus pool from the reward pool contract
  const bonusPoolAddress = await rewardPool.getBonusPoolAddress();
  const bonusPool = await ethers.getContractAt("BonusPool", bonusPoolAddress);
  const tokenIDs = [];
  for (let i = 1; i < 6; i++) {
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

  describe("lockFromApproval", async () => {
    it("approves transfer of nft to lockup, calls lockFromApproval in prelock phase", async () => {
      const account = accounts[1];
      const tokenID = stakedTokenIDs[1];
      await expect(lockStakedNFT(fixture, account, tokenID))
        .to.emit(fixture.publicStaking, "Transfer")
        .withArgs(account.address, fixture.lockup.address, tokenID);
    });

    // it("attempts to lockup someone elses tokenID",async () => {
    //   const account = accounts[1]
    //   const tokenID = stakedTokenIDs[2];
    //   await lockStakedNFT(fixture, account, tokenID)
    // })

    it("attempts to lockup 2 tokenID with 1 account", async () => {
      const account1 = accounts[1];
      const account2 = accounts[2];
      const tokenId1 = stakedTokenIDs[1];
      const tokenId2 = stakedTokenIDs[2];
      // give account 1 extra token
      let txResponse = await fixture.publicStaking
        .connect(account2)
        .transferFrom(account2.address, account1.address, tokenId2);
      await txResponse.wait();
      txResponse = await lockStakedNFT(fixture, account1, tokenId1);
      await txResponse.wait();
      await expect(
        lockStakedNFT(fixture, account1, tokenId2)
      ).to.be.revertedWithCustomError(fixture.lockup, "AddressAlreadyLockedUp");
    });

    it("attempts to lockup a tokenID that is already claimed", async () => {
      let account1 = accounts[1];
      let tokenId1 = stakedTokenIDs[1];
      //acct 1 locks tokenid1
      let txResponse = await lockStakedNFT(fixture, account1, tokenId1);
      await txResponse.wait();
      //account 2 attempts to lock tokenId1
      await expect(lockFromTransfer(fixture, account1, tokenId1))
        .to.be.revertedWithCustomError(fixture.lockup, "TokenIDAlreadyClaimed")
        .withArgs(tokenId1);
    });

    it("attempts to lockup a tokenID after the lockup period", async () => {
      await ensureBlockIsAtLeast(startBlock);

      const account1 = accounts[1];
      const tokenId1 = stakedTokenIDs[1];
      await expect(
        lockStakedNFT(fixture, account1, tokenId1)
      ).to.be.revertedWithCustomError(fixture.lockup, "PreLockStateRequired");
    });
  });

  describe("lockFromTransfer", async () => {
    it("succeeds when transfer approved and in prelock phase", async () => {
      const account = accounts[1];
      const tokenID = stakedTokenIDs[1];
      await approveStakingTokenTransfer(fixture, account, tokenID);

      await (
        await fixture.publicStaking
          .connect(account)
          .transferFrom(account.address, fixture.lockup.address, tokenID)
      ).wait();

      await expect(lockFromTransfer(fixture, account, tokenID))
        .to.emit(fixture.lockup, "NewLockup")
        .withArgs(account.address, tokenID);
    });

    it("reverts if token not owned by Lockup contract", async () => {
      const account1 = accounts[1];
      const tokenId1 = stakedTokenIDs[1];

      await expect(
        lockFromTransfer(fixture, account1, tokenId1)
      ).to.be.revertedWithCustomError(
        fixture.lockup,
        "ContractDoesNotOwnTokenID"
      );
    });

    it("reverts if called when state is not in PreLock", async () => {
      await ensureBlockIsAtLeast(startBlock);

      const account1 = accounts[1];
      const tokenId1 = stakedTokenIDs[1];

      await expect(
        lockFromTransfer(fixture, account1, tokenId1)
      ).to.be.revertedWithCustomError(fixture.lockup, "PreLockStateRequired");
    });
  });
  describe("collectAllProfits", async () => {
    it("collects all profits", async () => {
      const account1 = accounts[1];
      const initalACLABalance = await fixture.aToken.balanceOf(
        account1.address
      );
      const initalEthBalance = await ethers.provider.getBalance(
        account1.address
      );
      const txResponse = await mintBtoken(fixture, accounts[0], "1000");
      await txResponse.wait();
      // txResponse = await fixture.lockup.connect(account1).collectAllProfits()
      const currentACLABalance = await fixture.aToken.balanceOf(
        account1.address
      );
      const currentEthBalance = await ethers.provider.getBalance(
        account1.address
      );
      console.log(initalACLABalance, initalEthBalance);
      console.log(currentACLABalance, currentEthBalance);
    });

    it("attempts to collect before postLock phase", async () => {});
  });

  describe("unlockEarly", async () => {});

  describe("aggregateProfits", async () => {});

  describe("unlock", async () => {});

  describe("getter functions", async () => {
    before(async () => {});
  });
});

async function lockStakedNFT(
  fixture: Fixture,
  account: SignerWithAddress,
  tokenID: BigNumber,
  approve: boolean = true
): Promise<ContractTransaction> {
  if (approve) {
    const txResponse = await fixture.publicStaking
      .connect(account)
      .approve(fixture.lockup.address, tokenID);
    await txResponse.wait();
  }
  return fixture.lockup.connect(account).lockFromApproval(tokenID);
}

async function lockFromTransfer(
  fixture: Fixture,
  account: SignerWithAddress,
  tokenID: BigNumber
): Promise<ContractTransaction> {
  return fixture.lockup
    .connect(account)
    .lockFromTransfer(tokenID, account.address);
}

async function approveStakingTokenTransfer(
  fixture: Fixture,
  account: SignerWithAddress,
  tokenID: BigNumber
): Promise<ContractTransaction> {
  return fixture.publicStaking
    .connect(account)
    .approve(fixture.lockup.address, tokenID);
}

async function mintBtoken(
  fixture: Fixture,
  account: SignerWithAddress,
  amount: string
): Promise<ContractTransaction> {
  const btokenAmount = ethers.utils.parseEther(amount);
  const totalBToken = await fixture.bToken.totalSupply();
  const eth = await fixture.bToken.getEthToMintBTokens(
    totalBToken,
    btokenAmount
  );
  return fixture.bToken.connect(account).mint(amount, { value: eth });
}

async function ensureBlockIsAtLeast(targetBlock: number): Promise<void> {
  const currentBlock = await ethers.provider.getBlockNumber();
  if (currentBlock < targetBlock) {
    const blockDelta = targetBlock - currentBlock;
    await mineBlocks(BigInt(blockDelta));
  }
}
