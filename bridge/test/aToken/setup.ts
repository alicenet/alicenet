import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers/lib/ethers";
import { ethers } from "hardhat";
import { Fixture } from "../setup";

let admin: SignerWithAddress;
let user: SignerWithAddress;
export interface state {
  Balances: {
    madToken: {
      address: string;
      admin: number;
      user: number;
      aToken: number;
    };
    aToken: {
      address: string;
      admin: number;
      user: number;
      madToken: number;
    };
  };
}

export async function getState(fixture: Fixture) {
  [admin, user] = await ethers.getSigners();
  const state: state = {
    Balances: {
      madToken: {
        address: fixture.madToken.address.slice(-4),
        admin: format(await fixture.madToken.balanceOf(admin.address)),
        user: format(await fixture.madToken.balanceOf(user.address)),
        aToken: format(
          await fixture.madToken.balanceOf(fixture.aToken.address)
        ),
      },
      aToken: {
        address: fixture.aToken.address.slice(-4),
        admin: format(await fixture.aToken.balanceOf(admin.address)),
        user: format(await fixture.aToken.balanceOf(user.address)),
        madToken: format(
          await fixture.aToken.balanceOf(fixture.madToken.address)
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
  await fixture.madToken.connect(admin).approve(admin.address, 1000);
  await fixture.madToken
    .connect(admin)
    .transferFrom(admin.address, user.address, 1000);
  showState("Initial", await getState(fixture));
}
