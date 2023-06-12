import { task, types } from "hardhat/config";
import {
  getGasPrices,
  parseWaitConfirmationInterval,
  promptCheckDeploymentArgs,
} from "../lib/deployment/utils";
const madTokenABI = [
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
  "migrate-all-madtoken-to-alca",
  "Transfers ALCA from the AliceNet factory to an address"
)
  .addParam(
    "madTokenAddress",
    "address of madtoken contract",
    "0x5B09A0371C1DA44A8E24D36Bf5DEb1141a84d875",
    types.string
  )
  .addParam(
    "alcaAddress",
    "address of the alcb contract defaults to mainnet",
    "0xBb556b0eE2CBd89ed95DdEA881477723A3Aa8F8b",
    types.string
  )
  .addParam(
    "factoryAddress",
    "address of the factory contract",
    "0x4b6dF6B299fB6414f45719E0d9e1889269a7843E",
    types.string
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
    // santize inputs
    if (
      taskArgs.alcaAddress === hre.ethers.constants.AddressZero ||
      taskArgs.alcaAddress === undefined
    ) {
      throw new Error("ALCB address cannot be zero address");
    }
    let alca = await hre.ethers.getContractAt("ALCA", taskArgs.alcaAddress);
    if (
      taskArgs.factoryAddress === hre.ethers.constants.AddressZero ||
      taskArgs.factoryAddress === undefined
    ) {
      throw new Error("invalid factory address");
    }
    if (
      taskArgs.madTokenAddress === hre.ethers.constants.AddressZero ||
      taskArgs.madTokenAddress === undefined
    ) {
      throw new Error("mad token address cannot be zero address");
    }
    let madToken = new hre.ethers.Contract(
      taskArgs.madTokenAddress,
      madTokenABI
    );
    let signer = (await hre.ethers.getSigners())[0];
    if (taskArgs.test) {
      const address = "0xff55549a3ceea32fba4794bf1a649a2363fcda53";
      await hre.network.provider.request({
        method: "hardhat_impersonateAccount",
        params: [address],
      });
      const signers = await hre.ethers.getSigners();
      await signers[0].sendTransaction({
        to: address,
        value: hre.ethers.utils.parseEther("10"),
      });
      const helpers = require("@nomicfoundation/hardhat-network-helpers");
      await helpers.impersonateAccount(address);
      signer = await hre.ethers.getSigner(address);
      madToken = madToken.connect(signer);
      alca = alca.connect(signer);
    }

    // get the madtoken  balance of alicenent deployer
    const balance = await madToken.balanceOf(signer.address);
    // approve the alca contract to spend the balance
    let tx = await madToken.increaseApproval(
      taskArgs.alcaAddress,
      balance,
      await getGasPrices(hre.ethers)
    );
    const alcaInitialMadTokenBalance = await madToken.balanceOf(alca.address);
    const factoryInitialALCA = await alca.balanceOf(taskArgs.factoryAddress);
    const gas = await alca.estimateGas.migrateTo(
      taskArgs.factoryAddress,
      balance,
      { ...(await getGasPrices(hre.ethers)), gasLimit: 3000000 }
    );
    const expectedALCA = await alca.convert(balance);
    const promptMessage = `Do you want to convert ${hre.ethers.utils.formatEther(
      balance
    )} Madtoken to ${hre.ethers.utils.formatEther(
      expectedALCA
    )} for ${gas.toString()} units of gas? (y/n)\n`;
    await promptCheckDeploymentArgs(promptMessage);
    tx = await alca.migrateTo(
      taskArgs.factoryAddress,
      balance,
      await getGasPrices(hre.ethers)
    );
    const receipt = await tx.wait(waitConfirmationsBlocks);
    console.log("receipt:", receipt);
    const factoryEndingBalance = await alca.balanceOf(taskArgs.factoryAddress);
    const alcaMadTokenEndingBalance = await madToken.balanceOf(alca.address);
    console.log(
      "alca initial mad balance",
      hre.ethers.utils.formatEther(alcaInitialMadTokenBalance)
    );
    console.log(
      `alicenet deployer ${
        signer.address
      } initial balance: ${hre.ethers.utils.formatEther(balance)}`
    );
    console.log(`factory initial balance: ${hre.ethers.utils.formatEther(
      factoryInitialALCA
    )}
factory ending balance: ${hre.ethers.utils.formatEther(factoryEndingBalance)}
alca initial madtoken balance: ${hre.ethers.utils.formatEther(
      alcaInitialMadTokenBalance
    )}
alca ending madtoken balance: ${hre.ethers.utils.formatEther(
      alcaMadTokenEndingBalance
    )}
`);
  });
