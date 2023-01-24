import { Contract, ContractReceipt } from "ethers";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { AliceNetFactory } from "../../typechain-types/contracts/AliceNetFactory";

/**
 *
 * @param factory instance of alicenet factory object
 * @param hre hardhat runtime environment object
 * @returns alicenet factory object connected to imp
 */
export async function impersonateFactorySigner(
  ownerAddress: string,
  factory: AliceNetFactory,
  hre: HardhatRuntimeEnvironment
) {
  await hre.network.provider.request({
    method: "hardhat_impersonateAccount",
    params: [ownerAddress],
  });
  const helpers = require("@nomicfoundation/hardhat-network-helpers");
  await helpers.impersonateAccount(ownerAddress);
  const signer = await hre.ethers.getSigner(ownerAddress);
  return factory.connect(signer);
}
/**
 * @description processes the events emitted by a single contract
 * @param contract: ethers contract instance of the contract that emitted the event
 * @param receipt: ethers contract receipt object containing un parsed events
 */
export const processSingleContractEvent = async (
  contract: Contract,
  receipt: ContractReceipt
) => {
  if (receipt.events !== undefined && receipt.events.length > 0) {
    const events = receipt.events.map((event) => {
      return contract.interface.parseLog(event);
    });
    console.log("events: ", events);
    return events;
  }
};

/**
 * @description processes the events emitted by multiple contracts, and prints in human readable format
 * @param contracts array of ethers contract objects for contracts emitting events
 * @param ethers ethers object
 * @param receipt transaction receipt object
 * @returns
 */
export const processMultipleContractEvents = async (
  contracts: Array<Contract>,
  ethers: HardhatRuntimeEnvironment["ethers"],
  receipt: ContractReceipt
) => {
  let abi: Array<string> = [];
  for (let contract of contracts) {
    for (let event in contract.interface.events) {
      abi.push(event);
    }
  }
  const abiInterface = new ethers.utils.Interface(abi);
  if (receipt.events !== undefined && receipt.events.length > 0) {
    const events = receipt.events.map((event) => {
      return abiInterface.parseLog(event);
    });
    console.log("events: ", events);
    return events;
  }
};
