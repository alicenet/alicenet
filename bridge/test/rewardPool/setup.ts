import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import hre, { ethers } from "hardhat";
import { AToken, PublicStaking } from "../../typechain-types";
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

export async function depositTokensToAddress(
  accountFrom: SignerWithAddress,
  aToken: AToken,
  accountTo: string,
  alca: BigNumber
): Promise<void> {
  await (await aToken.connect(accountFrom).transfer(accountTo, alca)).wait();
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

export function calculateExpectedProportions(
  ethReserveAmount: BigNumber,
  tokenReserveAmount: BigNumber,
  userShares: BigNumber,
  totalShares: BigNumber
) {
  return [
    ethReserveAmount.mul(userShares).div(totalShares),
    tokenReserveAmount.mul(userShares).div(totalShares),
  ];
}
export async function ensureBlockIsAtLeast(targetBlock: number): Promise<void> {
  const currentBlock = await ethers.provider.getBlockNumber();
  if (currentBlock < targetBlock) {
    const blockDelta = targetBlock - currentBlock;
    await mineBlocks(BigInt(blockDelta));
  }
}
