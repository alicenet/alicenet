import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import {
  BigNumber,
  BytesLike,
  ContractReceipt,
  ContractTransaction,
} from "ethers/lib/ethers";
import hre, { ethers, expect } from "hardhat";
import { CONTRACT_ADDR, DEPLOYED_RAW } from "../../scripts/lib/constants";
import { getEventVar } from "../factory/Setup";
import {
  BaseTokensFixture,
  deployFactoryAndBaseTokens,
  Fixture,
  mineBlocks,
  posFixtureSetup,
  preFixtureSetup,
} from "../setup";
import { Distribution1 } from "./test.data";

export const numberOfLockingUsers = 5;
export const stakedAmount = ethers.utils.parseEther("100000000").toBigInt();
export const totalBonusAmount = ethers.utils.parseEther("2000000");
export const startBlock = 100;
export const lockDuration = 100;
export const LockupStates = {
  PreLock: 0,
  InLock: 1,
  PostLock: 2,
};
export const example = {
  totalStakedALCA: 100000000,
  distribution: Distribution1,
};

export const profitETH = ethers.utils.parseEther(
  example.distribution.profitETH
);
export const profitALCA = ethers.utils.parseEther(
  example.distribution.profitALCA
);

interface UserDistribution {
  owner: string;
  tokenID: number;
  shares: string;
  percentageFromTotal: number;
  profitETH: string;
  profitALCA: string;
}

interface UsersDistribution {
  [key: string]: UserDistribution;
}
export interface Distribution {
  profitETH: string;
  profitALCA: string;
  users: UsersDistribution;
}

export interface UnlockState {
  [key: string]: {
    tokenID: number;
    shares: number;
    percentageFromTotalLockup: number;
    bonusSharesALCA: number;
    bonusProfitEth: number;
    bonusProfitALCA: number;
    rewardHeldEth: number;
    rewardHeldALCA: number;
    lockupAggregatedEth: number;
    lockupAggregatedALCA: number;
    totalEarnedEth: string;
    totalEarnedALCA: string;
  };
}

interface UsersState {
  [key: string]: {
    address: string;
    alca: bigint;
    eth: bigint;
    tokenId: bigint;
    tokenOwner: string;
    position: bigint;
  };
}

interface LockupPositionsState {
  [key: string]: {
    index: bigint;
    owner: string;
    tokenId: bigint;
  };
}

interface StakingPositionsState {
  [key: string]: {
    shares: bigint;
  };
}

interface ContractsState {
  [key: string]: {
    alca: bigint;
    eth: bigint;
    lockedPositions?: bigint;
  };
}

export interface State {
  contracts: ContractsState;
  users: UsersState;
  lockupPositions: LockupPositionsState;
  stakingPositions: StakingPositionsState;
}

export async function getState(
  fixture: Fixture | BaseTokensFixture,
  stakingInitialPosition?: number
) {
  const signers = await ethers.getSigners();
  const contracts = [
    fixture.lockup,
    fixture.publicStaking,
    fixture.bonusPool,
    fixture.rewardPool,
    fixture.factory,
  ];
  const contractNames = [
    "lockup",
    "publicStaking",
    "bonusPool",
    "rewardPool",
    "factory",
  ];
  const contractsState: ContractsState = {};
  const usersState: UsersState = {};
  const lastMintedPosition = (
    await fixture.publicStaking.getLatestMintedPositionID()
  ).toNumber();
  if (stakingInitialPosition === undefined)
    stakingInitialPosition = lastMintedPosition;
  for (let i = 0; i < contracts.length; i++) {
    if (contractNames[i] === "lockup")
      contractsState[contractNames[i]] = {
        alca: (await fixture.aToken.balanceOf(contracts[i].address)).toBigInt(),
        eth: (
          await ethers.provider.getBalance(contracts[i].address)
        ).toBigInt(),
        lockedPositions: (
          await fixture.lockup.getCurrentNumberOfLockedPositions()
        ).toBigInt(),
      };
    else {
      contractsState[contractNames[i]] = {
        alca: (await fixture.aToken.balanceOf(contracts[i].address)).toBigInt(),
        eth: (
          await ethers.provider.getBalance(contracts[i].address)
        ).toBigInt(),
      };
    }
  }
  for (let i = 1; i <= numberOfLockingUsers; i++) {
    usersState["user" + i] = {
      address: signers[i].address,
      alca: (await fixture.aToken.balanceOf(signers[i].address)).toBigInt(),
      eth: (await ethers.provider.getBalance(signers[i].address)).toBigInt(),
      tokenId: (await fixture.lockup.tokenOf(signers[i].address)).toBigInt(),
      tokenOwner: await fixture.lockup.ownerOf(
        await fixture.lockup.tokenOf(signers[i].address)
      ),
      position: (
        await fixture.lockup.getPositionByIndex(
          await fixture.lockup.tokenOf(signers[i].address)
        )
      ).toBigInt(),
    };
  }
  usersState.bonusPool = {
    address: fixture.bonusPool.address,
    alca: (
      await fixture.aToken.balanceOf(fixture.bonusPool.address)
    ).toBigInt(),
    eth: (
      await ethers.provider.getBalance(fixture.bonusPool.address)
    ).toBigInt(),
    tokenId: (
      await fixture.lockup.tokenOf(fixture.bonusPool.address)
    ).toBigInt(),
    tokenOwner: await fixture.lockup.ownerOf(
      await fixture.lockup.tokenOf(fixture.bonusPool.address)
    ),
    position: (
      await fixture.lockup.getPositionByIndex(
        await fixture.lockup.tokenOf(fixture.bonusPool.address)
      )
    ).toBigInt(),
  };

  const positionsState: LockupPositionsState = {};
  const positions = await fixture.lockup.getCurrentNumberOfLockedPositions();
  for (let i = 1; i <= positions; i++) {
    const position = (await fixture.lockup.getPositionByIndex(i)).toBigInt();
    const owner = await fixture.lockup.ownerOf(position);
    const tokenId = (await fixture.lockup.tokenOf(owner)).toBigInt();
    positionsState[i] = {
      owner: await fixture.lockup.ownerOf(position),
      tokenId: (await fixture.lockup.tokenOf(owner)).toBigInt(),
      index: (await fixture.lockup.getIndexByTokenId(tokenId)).toBigInt(),
    };
  }

  const stakingsState: StakingPositionsState = {};
  for (let i = stakingInitialPosition + 1; i <= lastMintedPosition; i++) {
    const [position2Shares, , ,] = await fixture.publicStaking.getPosition(
      lastMintedPosition
    );
    stakingsState[i] = {
      shares: position2Shares.toBigInt(),
    };
  }

  const state: State = {
    contracts: contractsState,
    users: usersState,
    lockupPositions: positionsState,
    stakingPositions: stakingsState,
  };
  return state;
}

export function showState(title: string, state: State) {
  if (process.env.npm_config_detailed === "true") {
    // execute "npm --detailed=true test" to see this output
    console.log(title, state);
  }
}

export const getEthConsumedAsGas = (receipt: ContractReceipt): bigint => {
  return receipt.cumulativeGasUsed.mul(receipt.effectiveGasPrice).toBigInt();
};

export const getImpersonatedSigner = async (
  addressToImpersonate: string
): Promise<any> => {
  const [admin] = await ethers.getSigners();
  const testUtils = await (
    await (await ethers.getContractFactory("TestUtils")).deploy()
  ).deployed();
  await admin.sendTransaction({
    to: testUtils.address,
    value: ethers.utils.parseEther("1"),
  });
  await testUtils.payUnpayable(addressToImpersonate);
  await hre.network.provider.request({
    method: "hardhat_impersonateAccount",
    params: [addressToImpersonate],
  });
  return ethers.provider.getSigner(addressToImpersonate);
};

export async function deployLockupContract(
  baseTokensFixture: BaseTokensFixture
) {
  const lockupBase = await ethers.getContractFactory("Lockup");
  const lockupDeployCode = lockupBase.getDeployTransaction(
    startBlock,
    lockDuration,
    totalBonusAmount
  ).data as BytesLike;
  const txResponse = await baseTokensFixture.factory.deployCreate(
    lockupDeployCode
  );
  // get the address from the event
  const lockupAddress = await getEventVar(
    txResponse,
    DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  await posFixtureSetup(baseTokensFixture.factory, baseTokensFixture.aToken);
  return await ethers.getContractAt("Lockup", lockupAddress);
}

export async function getSimulatedStakingPositions(
  fixture: BaseTokensFixture,
  signers: SignerWithAddress[],
  numberOfUsers: number
) {
  const tokenIDs = [];
  const asFactory = await getImpersonatedSigner(fixture.factory.address);
  await fixture.aToken
    .connect(signers[0])
    .increaseAllowance(fixture.publicStaking.address, stakedAmount);
  await fixture.aToken
    .connect(signers[0])
    .transfer(fixture.bonusPool.address, totalBonusAmount);
  for (let i = 1; i <= numberOfUsers * 10; i++) {
    if (i % 10 === 0) {
      // stake test positions only for tokens 10,20,30,40 & 50
      const index = i / 10;
      const user = ("user" + index) as string;
      const stakedAmount = ethers.utils.parseEther(
        example.distribution.users[user].shares
      );
      await fixture.publicStaking
        .connect(signers[0])
        .mintTo(signers[index].address, stakedAmount, 0);
      const tokenID = await fixture.publicStaking.tokenOfOwnerByIndex(
        signers[index].address,
        0
      );
      tokenIDs[index] = tokenID;
    } else {
      if (i % 2 === 0) {
        // for the rest stake 1M if even
        await fixture.publicStaking
          .connect(signers[0])
          .mintTo(signers[0].address, ethers.utils.parseEther("1000000"), 0);
      } else {
        // or 500K if odd
        await fixture.publicStaking
          .connect(signers[0])
          .mintTo(signers[0].address, ethers.utils.parseEther("500000"), 0);
      }
    }
  }
  await fixture.bonusPool.connect(asFactory).createBonusStakedPosition();
  const leftOver =
    stakedAmount - (await fixture.publicStaking.getTotalShares()).toBigInt();
  await fixture.publicStaking
    .connect(signers[0])
    .mintTo(signers[0].address, leftOver, 0);
  tokenIDs[tokenIDs.length] = await fixture.publicStaking.tokenOfOwnerByIndex(
    fixture.bonusPool.address,
    0
  );
  return tokenIDs;
}

export async function deployFixture() {
  await preFixtureSetup();
  const signers = await ethers.getSigners();
  const baseTokensFixture = await deployFactoryAndBaseTokens(signers[0]);
  const lockup = await deployLockupContract(baseTokensFixture);
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
  const tokenIDs = await getSimulatedStakingPositions(fixture, signers, 5);
  expect(
    (await fixture.publicStaking.getTotalShares()).toBigInt()
  ).to.be.equals(stakedAmount);
  return {
    fixture,
    accounts: signers,
    stakedTokenIDs: tokenIDs,
    asFactory: factorySigner,
    asPublicStaking: pblicStakingSigner,
    asRewardPool: rewardPoolSigner,
  };
}

export async function distributeProfits(
  fixture: BaseTokensFixture,
  fundsSourceAddress: SignerWithAddress,
  profitETH: BigNumber,
  profitALCA: BigNumber
) {
  await fixture.aToken
    .connect(fundsSourceAddress)
    .increaseAllowance(fixture.publicStaking.address, profitALCA);
  await fixture.publicStaking.connect(fundsSourceAddress).depositEth(42, {
    value: profitETH,
  });
  await fixture.publicStaking
    .connect(fundsSourceAddress)
    .depositToken(42, profitALCA);
}

export async function lockStakedNFT(
  fixture: BaseTokensFixture,
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

export async function jumpToInlockState(fixture: BaseTokensFixture) {
  const blocksToMine = (await fixture.lockup.getLockupStartBlock())
    .sub(await ethers.provider.getBlockNumber())
    .toBigInt();
  await mineBlocks(blocksToMine + 1n);
  expect(await fixture.lockup.getState()).to.be.equals(LockupStates.InLock);
}

export async function jumpToPostLockState(fixture: BaseTokensFixture) {
  const blocksToMine = (await fixture.lockup.getLockupEndBlock())
    .sub(await ethers.provider.getBlockNumber())
    .toBigInt();
  await mineBlocks(blocksToMine + 1n);
  expect(await fixture.lockup.getState()).to.be.equals(LockupStates.PostLock);
}
