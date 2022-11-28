import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers/lib/ethers";
import { ethers } from "hardhat";
import { Fixture } from "../setup";

let admin: SignerWithAddress;
let user: SignerWithAddress;
export interface state {
  Balances: {
    legacyToken: {
      address: string;
      admin: bigint;
      user: bigint;
      alca: bigint;
    };
    alca: {
      address: string;
      admin: bigint;
      user: bigint;
      legacyToken: bigint;
    };
  };
}

export async function getState(fixture: Fixture) {
  [admin, user] = await ethers.getSigners();
  const state: state = {
    Balances: {
      legacyToken: {
        address: fixture.legacyToken.address.slice(-4),
        admin: (await fixture.legacyToken.balanceOf(admin.address)).toBigInt(),
        user: (await fixture.legacyToken.balanceOf(user.address)).toBigInt(),
        alca: (
          await fixture.legacyToken.balanceOf(fixture.alca.address)
        ).toBigInt(),
      },
      alca: {
        address: fixture.alca.address.slice(-4),
        admin: (await fixture.alca.balanceOf(admin.address)).toBigInt(),
        user: (await fixture.alca.balanceOf(user.address)).toBigInt(),
        legacyToken: (
          await fixture.alca.balanceOf(fixture.legacyToken.address)
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
  return parseInt(ethers.utils.formatUnits(number, 0));
}

export function getUserNotInRoleReason(address: string, role: string) {
  const reason =
    "AccessControl: account " +
    address.toLowerCase() +
    " is missing role " +
    role;
  return reason;
}

export async function init(fixture: Fixture) {
  [admin, user] = await ethers.getSigners();
  await fixture.legacyToken
    .connect(admin)
    .approve(admin.address, ethers.utils.parseUnits("220000000", 18));
  await fixture.legacyToken
    .connect(admin)
    .transferFrom(
      admin.address,
      user.address,
      ethers.utils.parseUnits("220000000", 18)
    );
  showState("Initial", await getState(fixture));
}
