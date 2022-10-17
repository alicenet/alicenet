import { BigNumber, ContractReceipt } from "ethers/lib/ethers";
import hre, { ethers } from "hardhat";
import { BaseTokensFixture, Fixture } from "../setup";

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
interface UserState {
  alca: bigint;
  eth: bigint;
  tokenOf: bigint;
  ownerOf: string;
  ethRewards: bigint;
  tokenRewards: bigint;
}

interface UsersState {
  [key: string]: UserState;
}

interface ContractState {
  alca: bigint;
  eth: bigint;
}
interface ContractsState {
  [key: string]: ContractState;
}

export interface State {
  contracts: ContractsState;
  users: UsersState;
}

export const numberOfLockingUsers = 5;

export async function getState(fixture: Fixture | BaseTokensFixture) {
  const signers = await ethers.getSigners();
  const state: State = {
    contracts: {
      lockup: {
        alca: (
          await fixture.aToken.balanceOf(fixture.lockup.address)
        ).toBigInt(),
        eth: (
          await ethers.provider.getBalance(fixture.lockup.address)
        ).toBigInt(),
      },
      publicStaking: {
        alca: (
          await fixture.aToken.balanceOf(fixture.publicStaking.address)
        ).toBigInt(),
        eth: (
          await ethers.provider.getBalance(fixture.publicStaking.address)
        ).toBigInt(),
      },
      factory: {
        alca: (
          await fixture.aToken.balanceOf(fixture.factory.address)
        ).toBigInt(),
        eth: (
          await ethers.provider.getBalance(fixture.factory.address)
        ).toBigInt(),
      },
      bonusPool: {
        alca: (
          await fixture.aToken.balanceOf(fixture.bonusPool.address)
        ).toBigInt(),
        eth: (
          await ethers.provider.getBalance(fixture.bonusPool.address)
        ).toBigInt(),
      },
    },
    users: {
      user1: {
        alca: (await fixture.aToken.balanceOf(signers[1].address)).toBigInt(),
        eth: (await ethers.provider.getBalance(signers[1].address)).toBigInt(),
        tokenOf: (await fixture.lockup.tokenOf(signers[1].address)).toBigInt(),
        ownerOf: await fixture.lockup.ownerOf(
          await fixture.lockup.tokenOf(signers[1].address)
        ),
        ethRewards: BigNumber.from(0).toBigInt(),
        tokenRewards: BigNumber.from(0).toBigInt(),
      },
      user2: {
        alca: (await fixture.aToken.balanceOf(signers[2].address)).toBigInt(),
        eth: (await ethers.provider.getBalance(signers[2].address)).toBigInt(),
        tokenOf: (await fixture.lockup.tokenOf(signers[2].address)).toBigInt(),
        ownerOf: await fixture.lockup.ownerOf(
          await fixture.lockup.tokenOf(signers[2].address)
        ),
        ethRewards: BigNumber.from(0).toBigInt(),
        tokenRewards: BigNumber.from(0).toBigInt(),
      },

      user3: {
        alca: (await fixture.aToken.balanceOf(signers[3].address)).toBigInt(),
        eth: (await ethers.provider.getBalance(signers[3].address)).toBigInt(),
        tokenOf: (await fixture.lockup.tokenOf(signers[3].address)).toBigInt(),
        ownerOf: await fixture.lockup.ownerOf(
          await fixture.lockup.tokenOf(signers[3].address)
        ),
        ethRewards: BigNumber.from(0).toBigInt(),
        tokenRewards: BigNumber.from(0).toBigInt(),
      },

      user4: {
        alca: (await fixture.aToken.balanceOf(signers[4].address)).toBigInt(),
        eth: (await ethers.provider.getBalance(signers[4].address)).toBigInt(),
        tokenOf: (await fixture.lockup.tokenOf(signers[4].address)).toBigInt(),
        ownerOf: await fixture.lockup.ownerOf(
          await fixture.lockup.tokenOf(signers[4].address)
        ),
        ethRewards: BigNumber.from(0).toBigInt(),
        tokenRewards: BigNumber.from(0).toBigInt(),
      },
      user5: {
        alca: (await fixture.aToken.balanceOf(signers[5].address)).toBigInt(),
        eth: (await ethers.provider.getBalance(signers[5].address)).toBigInt(),
        tokenOf: (await fixture.lockup.tokenOf(signers[5].address)).toBigInt(),
        ownerOf: await fixture.lockup.ownerOf(
          await fixture.lockup.tokenOf(signers[5].address)
        ),
        ethRewards: BigNumber.from(0).toBigInt(),
        tokenRewards: BigNumber.from(0).toBigInt(),
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
