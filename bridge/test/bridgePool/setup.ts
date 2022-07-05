import { BigNumber } from "ethers";
import { defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { BridgePoolV1 } from "../../typechain-types";
import { Fixture } from "../setup";

export interface state {
  Balances: {
    aToken: {
      address: string;
      admin: bigint;
      user: bigint;
      bridgePool: bigint;
      totalSupply: bigint;
    };
    bToken: {
      address: string;
      admin: bigint;
      user: bigint;
      bridgePool: bigint;
      totalSupply: bigint;
    };
    eth: {
      address: string;
      // We leave user balance as number to round values and avoid consumed gas comparison
      admin: number;
      user: number;
      bridgePool: bigint;
      aToken: bigint;
      bToken: bigint;
    };
  };
}

export async function getState(fixture: Fixture, bridgePool: BridgePoolV1) {
  const [admin, user] = await ethers.getSigners();
  const state: state = {
    Balances: {
      aToken: {
        address: fixture.aToken.address.slice(-4),
        admin: (await fixture.aToken.balanceOf(admin.address)).toBigInt(),
        user: (await fixture.aToken.balanceOf(user.address)).toBigInt(),
        bridgePool: (
          await fixture.aToken.balanceOf(bridgePool.address)
        ).toBigInt(),
        totalSupply: (await fixture.aToken.totalSupply()).toBigInt(),
      },
      bToken: {
        address: fixture.bToken.address.slice(-4),
        admin: (await fixture.bToken.balanceOf(admin.address)).toBigInt(),
        user: (await fixture.bToken.balanceOf(user.address)).toBigInt(),
        bridgePool: (
          await fixture.bToken.balanceOf(bridgePool.address)
        ).toBigInt(),
        totalSupply: (await fixture.bToken.totalSupply()).toBigInt(),
      },
      eth: {
        address: "0000",
        admin: format(await ethers.provider.getBalance(admin.address)),
        user: format(await ethers.provider.getBalance(user.address)),
        bridgePool: (
          await ethers.provider.getBalance(bridgePool.address)
        ).toBigInt(),
        aToken: (
          await ethers.provider.getBalance(fixture.aToken.address)
        ).toBigInt(),
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
    // execute "npm run --detailed=true test" to see this output
    console.log(title, state);
  }
}

export function format(number: BigNumber) {
  return parseFloat((+ethers.utils.formatEther(number)).toFixed(0));
}

export function formatBigInt(number: BigNumber) {
  return BigInt(parseFloat((+ethers.utils.formatEther(number)).toFixed(0)));
}

export function getMockBlockClaimsForStateRoot(stateRoot: string) {
  return defaultAbiCoder.encode(
    ["uint32", "uint32", "uint32", "bytes32", "bytes32", "bytes32", "bytes32"],
    [
      0,
      0,
      0,
      "0x0000000000000000000000000000000000000000000000000000000000000000",
      "0x0000000000000000000000000000000000000000000000000000000000000000",
      stateRoot,
      "0x0000000000000000000000000000000000000000000000000000000000000000",
    ]
  );
}
