import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers/lib/ethers";
import { ethers } from "hardhat";
import { Fixture } from "../../setup";

let admin: SignerWithAddress;
let user: SignerWithAddress;
let user2: SignerWithAddress;

export interface state {
  Balances: {
    madByte: {
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
      madByte: bigint;
    };
  };
}

export async function getState(fixture: Fixture) {
  [admin, user, user2] = await ethers.getSigners();
  let state: state = {
    Balances: {
      madByte: {
        address: fixture.madByte.address.slice(-4),
        admin: (await fixture.madByte.balanceOf(admin.address)).toBigInt(),
        user: (await fixture.madByte.balanceOf(user.address)).toBigInt(),
        user2: (await fixture.madByte.balanceOf(user2.address)).toBigInt(),
        totalSupply: (await fixture.madByte.totalSupply()).toBigInt(),
        poolBalance: (await fixture.madByte.getPoolBalance()).toBigInt(),
      },
      eth: {
        address: "0000",
        admin: format(await ethers.provider.getBalance(admin.address)),
        user: format(await ethers.provider.getBalance(user.address)),
        user2: format(await ethers.provider.getBalance(user2.address)),
        madByte: (
          await ethers.provider.getBalance(fixture.madByte.address)
        ).toBigInt(),
      },
    },
  };
  return state;
}

// export async function getState(contracts: stateActor[], users: stateActor[]) {
//   let state: state;
//   let actorStates: stateActorBalance;
//   contracts.map(async (contract, i) => {
//     users.map(async (user, i) => {
//       let stateActorBalance: stateActorBalance = {
//         actor: user,
//         balance: await (contract.actor as MadByte).balanceOf(
//           user.actor.address
//         ),
//       };
//       state.balances.push(actorStates);
//     });
//   });

//   return state;
// }

export function showState(title: string, state: state) {
  if (process.env.npm_config_detailed == "true") {
    // execute "npm --detailed=true test" to see this output
    console.log(title, state);
  }
}

export function format(number: BigNumber) {
  return parseFloat((+ethers.utils.formatEther(number)).toFixed(0));
  // return parseInt(ethers.utils.formatUnits(number, 0));
}

export function formatBigInt(number: BigNumber) {
  return BigInt(parseFloat((+ethers.utils.formatEther(number)).toFixed(0)));
  // return parseInt(ethers.utils.formatUnits(number, 0));
}

export function getUserNotInRoleReason(address: string, role: string) {
  let reason =
    "AccessControl: account " +
    address.toLowerCase() +
    " is missing role " +
    role;
  return reason;
}

export async function init(fixture: Fixture) {
  // const contractName = await fixture.aToken.name();
  // [admin, user] = await ethers.getSigners();
  // await fixture.madToken.connect(admin).approve(admin.address, 1000);
  // await fixture.madToken
  //   .connect(admin)
  //   .transferFrom(admin.address, user.address, 1000);
  // showState("Initial", await getState(fixture));
}

export async function getResultsFromTx(tx: any) {
  let abi = [
    "event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)",
  ];
  let iface = new ethers.utils.Interface(abi);
  let receipt = await ethers.provider.getTransactionReceipt(tx.hash);
  console.log("Logs", receipt);
  let logs =
    typeof receipt.logs[2] !== "undefined" ? receipt.logs[2] : receipt.logs[0];
  let log = iface.parseLog(logs);
  const { from, to, tokenId } = log.args;
  return tokenId;
}
