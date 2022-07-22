import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers/lib/ethers";
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
      admin: number;
      user: number;
      user2: number;
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
        admin: format(await ethers.provider.getBalance(admin.address)),
        user: format(await ethers.provider.getBalance(user.address)),
        user2: format(await ethers.provider.getBalance(user2.address)),
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