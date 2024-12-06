import { ethers } from "hardhat";
import { Fixture } from "../setup";
import { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";

let admin: HardhatEthersSigner;
let user: HardhatEthersSigner;
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
        address: (await fixture.legacyToken.getAddress()).slice(-4),
        admin: await fixture.legacyToken.balanceOf(admin.address),
        user: await fixture.legacyToken.balanceOf(user.address),
        alca: await fixture.legacyToken.balanceOf(
          await fixture.alca.getAddress()
        ),
      },
      alca: {
        address: (await fixture.alca.getAddress()).slice(-4),
        admin: await fixture.alca.balanceOf(admin.address),
        user: await fixture.alca.balanceOf(user.address),
        legacyToken: await fixture.alca.balanceOf(
          await fixture.legacyToken.getAddress()
        ),
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

export function format(number: bigint) {
  return parseInt(ethers.formatUnits(number, 0));
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
    .approve(admin.address, ethers.parseUnits("220000000", 18));
  await fixture.legacyToken
    .connect(admin)
    .transferFrom(
      admin.address,
      user.address,
      ethers.parseUnits("220000000", 18)
    );
  showState("Initial", await getState(fixture));
}
