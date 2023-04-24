import { expect } from "chai";
import { task, types } from "hardhat/config";
import { parseEvents } from "../lib/deployment/tasks";
import {
  getGasPrices,
  parseWaitConfirmationInterval,
  promptCheckDeploymentArgs,
} from "../lib/deployment/utils";

const MADTOKENABI = [
  {
    constant: true,
    inputs: [],
    name: "name",
    outputs: [{ name: "", type: "string" }],
    payable: false,
    stateMutability: "view",
    type: "function",
  },
  {
    constant: false,
    inputs: [
      { name: "_spender", type: "address" },
      { name: "_value", type: "uint256" },
    ],
    name: "approve",
    outputs: [{ name: "", type: "bool" }],
    payable: false,
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    constant: true,
    inputs: [],
    name: "totalSupply",
    outputs: [{ name: "", type: "uint256" }],
    payable: false,
    stateMutability: "view",
    type: "function",
  },
  {
    constant: false,
    inputs: [
      { name: "_from", type: "address" },
      { name: "_to", type: "address" },
      { name: "_value", type: "uint256" },
    ],
    name: "transferFrom",
    outputs: [{ name: "", type: "bool" }],
    payable: false,
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    constant: true,
    inputs: [],
    name: "INITIAL_SUPPLY",
    outputs: [{ name: "", type: "uint256" }],
    payable: false,
    stateMutability: "view",
    type: "function",
  },
  {
    constant: true,
    inputs: [],
    name: "decimals",
    outputs: [{ name: "", type: "uint8" }],
    payable: false,
    stateMutability: "view",
    type: "function",
  },
  {
    constant: false,
    inputs: [
      { name: "_spender", type: "address" },
      { name: "_subtractedValue", type: "uint256" },
    ],
    name: "decreaseApproval",
    outputs: [{ name: "success", type: "bool" }],
    payable: false,
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    constant: true,
    inputs: [{ name: "_owner", type: "address" }],
    name: "balanceOf",
    outputs: [{ name: "balance", type: "uint256" }],
    payable: false,
    stateMutability: "view",
    type: "function",
  },
  {
    constant: true,
    inputs: [],
    name: "symbol",
    outputs: [{ name: "", type: "string" }],
    payable: false,
    stateMutability: "view",
    type: "function",
  },
  {
    constant: false,
    inputs: [
      { name: "_to", type: "address" },
      { name: "_value", type: "uint256" },
    ],
    name: "transfer",
    outputs: [{ name: "", type: "bool" }],
    payable: false,
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    constant: false,
    inputs: [
      { name: "_spender", type: "address" },
      { name: "_addedValue", type: "uint256" },
    ],
    name: "increaseApproval",
    outputs: [{ name: "success", type: "bool" }],
    payable: false,
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    constant: true,
    inputs: [
      { name: "_owner", type: "address" },
      { name: "_spender", type: "address" },
    ],
    name: "allowance",
    outputs: [{ name: "remaining", type: "uint256" }],
    payable: false,
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    payable: false,
    stateMutability: "nonpayable",
    type: "constructor",
  },
  {
    anonymous: false,
    inputs: [
      { indexed: true, name: "owner", type: "address" },
      { indexed: true, name: "spender", type: "address" },
      { indexed: false, name: "value", type: "uint256" },
    ],
    name: "Approval",
    type: "event",
  },
  {
    anonymous: false,
    inputs: [
      { indexed: true, name: "from", type: "address" },
      { indexed: true, name: "to", type: "address" },
      { indexed: false, name: "value", type: "uint256" },
    ],
    name: "Transfer",
    type: "event",
  },
];

task(
  "send-all-madtoken-to-alca",
  "Transfers ALCA from the AliceNet factory to an address"
)
  .addParam(
    "madTokenAddress",
    "address of madtoken contract",
    "0x5B09A0371C1DA44A8E24D36Bf5DEb1141a84d875"
  )
  .addParam(
    "alcaAddress",
    "address of the alcb contract defaults to mainnet",
    "0xBb556b0eE2CBd89ed95DdEA881477723A3Aa8F8b"
  )
  .addFlag("test", "run in hardhat fork mode")
  .addOptionalParam(
    "waitConfirmation",
    "wait specified number of blocks between transactions",
    0,
    types.int
  )
  .setAction(async (taskArgs, hre) => {
    const waitConfirmationsBlocks = await parseWaitConfirmationInterval(
      taskArgs.waitConfirmation,
      hre
    );
    let madToken = await hre.ethers.getContractAt(
      MADTOKENABI,
      taskArgs.madTokenAddress
    );
    if (taskArgs.test) {
      const address = "0xff55549a3ceea32fba4794bf1a649a2363fcda53";
      await hre.network.provider.request({
        method: "hardhat_impersonateAccount",
        params: [address],
      });
      const helpers = require("@nomicfoundation/hardhat-network-helpers");
      await helpers.impersonateAccount(address);
      const signer = await hre.ethers.getSigner(address);
      madToken = madToken.connect(signer);
    }

    //get the balance of alicenent deployer
    const balance = await madToken.balanceOf(
      await madToken.signer.getAddress()
    );

    //send the balance to alca
    if (taskArgs.alcbAddress === hre.ethers.constants.AddressZero) {
      throw new Error("ALCB address cannot be zero address");
    }
    const gas = await madToken.estimateGas.transfer(
      taskArgs.alcaAddress,
      balance,
      await getGasPrices(hre.ethers)
    );
    const promptMessage = `Do you want to send ${hre.ethers.utils.formatEther(
      balance
    )} Madtoken to the address ${
      taskArgs.alcaAddress
    } for ${gas.toString()} units ? (y/n)\n`;
    await promptCheckDeploymentArgs(promptMessage);
    const originalALCABalance = await madToken.balanceOf(taskArgs.alcaAddress);
    const tx = await madToken.transfer(
      taskArgs.alcaAddress,
      balance,
      await getGasPrices(hre.ethers)
    );
    const receipt = await tx.wait();
    const currentALCABalance = await madToken.balanceOf(taskArgs.alcaAddress);
    expect(currentALCABalance).to.equal(originalALCABalance.add(balance));
    await parseEvents([madToken], hre.ethers, receipt);
    console.log("Successfully transferred ALCA!");
  });
