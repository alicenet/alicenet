import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import hre, { ethers } from "hardhat";
import { ALCA, BonusPool, PublicStaking } from "../../typechain-types";
import { mineBlocks } from "../setup";

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

export async function depositEthForStakingRewards(
  accounts: SignerWithAddress[],
  publicStaking: PublicStaking,
  eth: BigNumber
): Promise<void> {
  await (
    await publicStaking.connect(accounts[0]).depositEth(42, { value: eth })
  ).wait();
}

export async function depositEthToAddress(
  accountFrom: SignerWithAddress,
  accountTo: string,
  eth: BigNumber
): Promise<void> {
  await accountFrom.sendTransaction({
    to: accountTo,
    value: eth,
  });
}

export async function depositTokensForStakingRewards(
  accounts: SignerWithAddress[],
  alca: ALCA,
  publicStaking: PublicStaking,
  alcaAmount: BigNumber
): Promise<void> {
  await (
    await alca
      .connect(accounts[0])
      .increaseAllowance(publicStaking.address, alcaAmount)
  ).wait();

  await (
    await publicStaking.connect(accounts[0]).depositToken(42, alcaAmount)
  ).wait();
}

export async function ensureBlockIsAtLeast(targetBlock: number): Promise<void> {
  const currentBlock = await ethers.provider.getBlockNumber();
  if (currentBlock < targetBlock) {
    const blockDelta = targetBlock - currentBlock;
    await mineBlocks(BigInt(blockDelta));
  }
}

export async function mintBonusPosition(
  accounts: SignerWithAddress[],
  exactStakeAmount: BigNumber,
  alca: ALCA,
  bonusPool: BonusPool,
  mockFactorySigner: SignerWithAddress
) {
  await (
    await alca
      .connect(accounts[0])
      .transfer(bonusPool.address, exactStakeAmount)
  ).wait();

  const receipt = await (
    await bonusPool.connect(mockFactorySigner).createBonusStakedPosition()
  ).wait();

  const createdEvent = receipt.events?.find(
    (event) => event.event === "BonusPositionCreated"
  );

  return createdEvent?.args?.tokenID;
}

export async function calculateTerminationProfits(
  tokenId: BigNumber,
  publicStaking: PublicStaking
): Promise<[BigNumber, BigNumber]> {
  const [estimatedPayoutEth, estimatedPayoutToken] =
    await publicStaking.estimateAllProfits(tokenId);
  const [shares, , , ,] = await publicStaking.getPosition(tokenId);

  return [estimatedPayoutEth, estimatedPayoutToken.add(shares)];
}

export async function calculateUserProfits(
  userSharesLocked: BigNumber,
  currentTotalSharesLocked: BigNumber,
  tokenId: BigNumber,
  publicStaking: PublicStaking
): Promise<[BigNumber, BigNumber]> {
  const [estimatedPayoutEth, estimatedPayoutToken] =
    await publicStaking.estimateAllProfits(tokenId);
  const [shares, , , ,] = await publicStaking.getPosition(tokenId);

  const userExpectedBonusRewardEth = estimatedPayoutEth
    .mul(userSharesLocked)
    .div(currentTotalSharesLocked);

  const userExpectedBonusRewardToken = estimatedPayoutToken
    .add(shares)
    .mul(userSharesLocked)
    .div(currentTotalSharesLocked);

  return [userExpectedBonusRewardEth, userExpectedBonusRewardToken];
}
