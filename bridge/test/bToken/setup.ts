import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber, ContractReceipt } from "ethers/lib/ethers";
import { ethers } from "hardhat";
import { BaseTokensFixture, Fixture } from "../setup";

let admin: SignerWithAddress;
let user: SignerWithAddress;
let user2: SignerWithAddress;

export interface state {
  Balances: {
    bToken: {
      address: string;
      admin: bigint;
      user: bigint;
      user2: bigint;
      totalSupply: bigint;
      poolBalance: bigint;
    };
    eth: {
      address: string;
      // We leave user balance as number to round and avoid comparing of gas consumed
      admin: bigint;
      user: bigint;
      user2: bigint;
      bToken: bigint;
    };
  };
}

export async function getState(fixture: Fixture | BaseTokensFixture) {
  [admin, user, user2] = await ethers.getSigners();
  const state: state = {
    Balances: {
      bToken: {
        address: fixture.bToken.address.slice(-4),
        admin: (await fixture.bToken.balanceOf(admin.address)).toBigInt(),
        user: (await fixture.bToken.balanceOf(user.address)).toBigInt(),
        user2: (await fixture.bToken.balanceOf(user2.address)).toBigInt(),
        totalSupply: (await fixture.bToken.totalSupply()).toBigInt(),
        poolBalance: (await fixture.bToken.getPoolBalance()).toBigInt(),
      },
      eth: {
        address: "0000",
        admin: (await ethers.provider.getBalance(admin.address)).toBigInt(),
        user: (await ethers.provider.getBalance(user.address)).toBigInt(),
        user2: (await ethers.provider.getBalance(user2.address)).toBigInt(),
        bToken: (
          await ethers.provider.getBalance(fixture.bToken.address)
        ).toBigInt(),
      },
    },
  };
  return state;
}

export function showState(title: string, state: state) {
  if (process.env.npm_config_detailed === "true") {
    // execute "npm --detailed=true test" to see this output
    console.log(title, state);
  }
}

export function format(number: BigNumber) {
  return parseFloat((+ethers.utils.formatEther(number)).toFixed(0));
}

export function formatBigInt(number: BigNumber) {
  return BigInt(parseFloat((+ethers.utils.formatEther(number)).toFixed(0)));
}

export async function getResultsFromTx(tx: any) {
  const abi = [
    "event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)",
  ];
  const iface = new ethers.utils.Interface(abi);
  const receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  console.log("Logs", receipt);
  const logs =
    typeof receipt.logs[2] !== "undefined" ? receipt.logs[2] : receipt.logs[0];
  const log = iface.parseLog(logs);
  return log.args[2];
}

export const getBridgeRouterSalt = (version: number): string => {
  return ethers.utils.keccak256(
    ethers.utils.solidityPack(
      ["bytes32", "bytes32"],
      [
        ethers.utils.solidityKeccak256(["string"], ["BridgeRouter"]),
        ethers.utils.solidityKeccak256(["uint16"], [version]),
      ]
    )
  );
};

export const getEthConsumedAsGas = (receipt: ContractReceipt): bigint => {
  return receipt.cumulativeGasUsed.mul(receipt.effectiveGasPrice).toBigInt();
};
