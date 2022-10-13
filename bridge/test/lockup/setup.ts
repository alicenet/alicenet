import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ContractReceipt } from "ethers/lib/ethers";
import hre, { ethers } from "hardhat";
import { BaseTokensFixture, Fixture } from "../setup";

let admin: SignerWithAddress;
let user1: SignerWithAddress;
let user2: SignerWithAddress;

export interface Balances {
  alca: bigint;
  eth: bigint;
}

export interface Lockup {
  tokenOf: bigint;
  ownerOf: string;
  ethRewards: bigint;
  tokenRewards: bigint;
}

export interface LockState {
  balances: Balances;
  lockup: Lockup;
}

export interface State {
  lockupContract: Balances;
  user1: LockState;
}

export async function getState(fixture: Fixture | BaseTokensFixture) {
  [, user1] = await ethers.getSigners();
  const state: State = {
    lockupContract: {
      alca: (await fixture.aToken.balanceOf(fixture.lockup.address)).toBigInt(),
      eth: (
        await ethers.provider.getBalance(fixture.lockup.address)
      ).toBigInt(),
    },
    user1: {
      balances: {
        alca: (await fixture.aToken.balanceOf(user1.address)).toBigInt(),
        eth: (await ethers.provider.getBalance(user1.address)).toBigInt(),
      },
      lockup: {
        tokenOf: (await fixture.lockup.tokenOf(user1.address)).toBigInt(),
        ownerOf: await fixture.lockup.ownerOf(
          await fixture.lockup.tokenOf(user1.address)
        ),
        ethRewards: (
          await fixture.lockup.connect(user1.address).getEthRewardBalance()
        ).toBigInt(),
        tokenRewards: (
          await fixture.lockup.connect(user1.address).getTokenRewardBalance()
        ).toBigInt(),
      },
    },
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
