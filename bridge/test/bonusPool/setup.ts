import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import hre, { ethers } from "hardhat";
import { AToken, BonusPool, PublicStaking } from "../../typechain-types";
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
  aToken: AToken,
  publicStaking: PublicStaking,
  alca: BigNumber
): Promise<void> {
  await (
    await aToken
      .connect(accounts[0])
      .increaseAllowance(publicStaking.address, alca)
  ).wait();

  await (
    await publicStaking.connect(accounts[0]).depositToken(42, alca)
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
  aToken: AToken,
  bonusPool: BonusPool,
  mockFactorySigner: SignerWithAddress
) {
  await (
    await aToken
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
  finalTotalSharesLocked: BigNumber,
  originalTotalSharesLocked: BigNumber,
  tokenId: BigNumber,
  bonusPool: BonusPool,
  publicStaking: PublicStaking
): Promise<[BigNumber, BigNumber, BigNumber, BigNumber, BigNumber]> {
  const scalingFactor = await bonusPool.SCALING_FACTOR();
  const bonusRate = await bonusPool.getScaledBonusRate(
    originalTotalSharesLocked
  );
  const overallProportion = finalTotalSharesLocked
    .mul(scalingFactor)
    .div(originalTotalSharesLocked);

  // estimate all profits does not include the original stake amount, hence no need to subtract it here
  const [estimatedPayoutEth, estimatedPayoutToken] =
    await publicStaking.estimateAllProfits(tokenId);

  const expectedBonusShares = bonusRate
    .mul(finalTotalSharesLocked)
    .div(scalingFactor);
  const expectedBonusRewardEth = overallProportion
    .mul(estimatedPayoutEth)
    .div(scalingFactor);
  const expectedBonusRewardToken = overallProportion
    .mul(estimatedPayoutToken)
    .div(scalingFactor);

  return [
    estimatedPayoutEth,
    estimatedPayoutToken,
    expectedBonusShares,
    expectedBonusRewardEth,
    expectedBonusRewardToken.add(expectedBonusShares),
  ];
}

export async function calculateUserProfits(
  userSharesLocked: BigNumber,
  currentTotalSharesLocked: BigNumber,
  originalTotalSharesLocked: BigNumber,
  tokenId: BigNumber,
  bonusPool: BonusPool,
  publicStaking: PublicStaking
): Promise<[BigNumber, BigNumber, BigNumber]> {
  const scalingFactor = await bonusPool.SCALING_FACTOR();
  const bonusRate = await bonusPool.getScaledBonusRate(
    originalTotalSharesLocked
  );
  const overallProportion = currentTotalSharesLocked
    .mul(scalingFactor)
    .div(originalTotalSharesLocked);
  const userProportion = userSharesLocked
    .mul(scalingFactor)
    .div(currentTotalSharesLocked);

  const [estimatedPayoutEth, estimatedPayoutToken] =
    await publicStaking.estimateAllProfits(tokenId);

  const totalExpectedBonusRewardEth = overallProportion
    .mul(estimatedPayoutEth)
    .div(scalingFactor);
  const totalExpectedBonusRewardToken = overallProportion
    .mul(estimatedPayoutToken)
    .div(scalingFactor);

  const expectedUserBonusShares = bonusRate
    .mul(userSharesLocked)
    .div(scalingFactor);
  const userExpectedBonusRewardEth = userProportion
    .mul(totalExpectedBonusRewardEth)
    .div(scalingFactor);
  const userExpectedBonusRewardToken = userProportion
    .mul(totalExpectedBonusRewardToken)
    .div(scalingFactor);

  return [
    expectedUserBonusShares,
    userExpectedBonusRewardEth,
    userExpectedBonusRewardToken,
  ];
}
