import { HardhatEthersHelpers } from "@nomiclabs/hardhat-ethers/types";
import { BigNumber, BigNumberish, BytesLike, ContractFactory, ContractReceipt } from "ethers";
import fs from "fs";
import { Artifacts, HardhatRuntimeEnvironment } from "hardhat/types";
import { exit } from "process";
import readline from "readline";
import { AliceNetFactory } from "../../../typechain-types";
import {
  deployCreate,
  deployCreate2,
  deployCreateAndRegister,
  deployFactory,
  deployUpgradeableGasSafe,
  getEventVar,
  multiCallDeployUpgradeable,
  upgradeProxyGasSafe,
} from "../alicenetFactory";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  DEFAULT_CONFIG_FILE_PATH,
  EVENT_DEPLOYED_PROXY,
  EVENT_DEPLOYED_RAW,
  FUNCTION_DEPLOY_CREATE,
  FUNCTION_INITIALIZE,
  ONLY_PROXY,
  UPGRADEABLE_DEPLOYMENT,
} from "../constants";

type Ethers = typeof import("../../../node_modules/ethers/lib/ethers") &
  HardhatEthersHelpers;

export class InitializerArgsError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "InitializerArgsError";
  }
}

export interface ArgData {
  [key: string]:
    | string
    | number
    | boolean
    | BigNumberish
    | BytesLike
    | ArrayLike<string | number | boolean | BigNumberish | BytesLike>;
}
export interface ContractArgs {
  [key: string]: ArgData;
}
export interface DeploymentArgs {
  constructor: ContractArgs;
  initializer: ContractArgs;
}

export interface ContractDescriptor {
  name: string;
  fullyQualifiedName: string;
  deployGroup: string;
  deployGroupIndex: number;
  deployType: string;
  constructorArgs: Array<any>;
  initializerArgs: Array<any>;
}

export interface DeploymentConfig {
  name: string;
  fullyQualifiedName: string;
  salt: string;
  deployGroup: string;
  deployGroupIndex: string;
  deployType: string;
  constructorArgs: ArgData;
  initializerArgs: ArgData;
}

export interface DeploymentConfigWrapper {
  [key: string]: DeploymentConfig;
}

export type DeploymentList = {
  [key: string]: Array<DeploymentConfig>;
};

export type FactoryData = {
  address: string;
  owner?: string;
  gas: BigNumber;
};

export type DeployCreateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: BigNumber;
  constructorArgs?: any;
  receipt?: ContractReceipt;
};
export type MetaContractData = {
  metaAddress: string;
  salt: string;
  templateName: string;
  templateAddress: string;
  factoryAddress: string;
  gas: BigNumber;
  initCallData: string;
  receipt?: ContractReceipt;
};
export type TemplateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: BigNumber;
  receipt?: ContractReceipt;
  constructorArgs?: string;
};

export interface FactoryConfig {
  [key: string]: any;
}
export type ProxyData = {
  proxyAddress: string;
  salt: BytesLike;
  logicName?: string;
  logicAddress?: string;
  factoryAddress: string;
  gas: BigNumberish;
  receipt?: ContractReceipt;
  initCallData?: BytesLike;
};

// AliceNet Factory Task Functions

export async function deployFactoryTask(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment,
  legacyTokenAddress: string
) {
  const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
  if (!hre.ethers.utils.isAddress(legacyTokenAddress)) {
    throw new Error("legacyTokenAddress is not an address");
  }
  const signers = await hre.ethers.getSigners();
  const deployTX = factoryBase.getDeployTransaction(legacyTokenAddress);
  const gasCost = await hre.ethers.provider.estimateGas(deployTX);
  // deploys the factory
  const factory = await deployFactory(
    legacyTokenAddress,
    hre.ethers,
    factoryBase,
    await getGasPrices(hre)
  );
  // wait for deployment confirmation
  await factory.deployTransaction.wait(taskArgs.waitConfirmation);
  // record the state in a json file to be used in other tasks
  const factoryData: FactoryData = {
    address: factory.address,
    owner: signers[0].address,
    gas: gasCost,
  };
  if (taskArgs.verify) {
    await verifyContract(hre, factory.address, [legacyTokenAddress]);
  }
  const network = hre.network.name;

  await showState(
    `Deployed ${ALICENET_FACTORY} at address: ${factory.address}, gasCost: ${gasCost}`
  );
  await showState(
    `Deployed ALCA at address: ${await factory.lookup(
      hre.ethers.utils.formatBytes32String("AToken")
    )}, gasCost: ${gasCost}`
  );
  return factoryData;
}

export async function deployContractsTask(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment
) {
  let cumulativeGasUsed = BigNumber.from("0");

  const deploymentListConfig = await readDeploymentConfig(taskArgs.configFile);
  // setting listName undefined will use the default list
  // deploy the factory first
  const keys = Object.keys(deploymentListConfig);

  const legacyTokenAddress = deploymentListConfig[keys[0]].constructorArgs
    .legacyToken_ as string;
  let factoryAddress = taskArgs.factoryAddress;
  if (factoryAddress === undefined) {
    const factoryData: FactoryData = await deployFactoryTask(
      taskArgs,
      hre,
      legacyTokenAddress
    );
    factoryAddress = factoryData.address;
    cumulativeGasUsed = cumulativeGasUsed.add(factoryData.gas);
  }
  // connect an instance of the factory
  const factory = await hre.ethers.getContractAt(
    "AliceNetFactory",
    factoryAddress
  );

  for (const fullyQualifiedContractName in deploymentListConfig) {
    const deployType =
      deploymentListConfig[fullyQualifiedContractName].deployType;
    const salt = deploymentListConfig[fullyQualifiedContractName].salt;
    const constructorArgObject =
      deploymentListConfig[fullyQualifiedContractName].constructorArgs;
    const initializerArgObject =
      deploymentListConfig[fullyQualifiedContractName].initializerArgs;

    switch (deployType) {
      case UPGRADEABLE_DEPLOYMENT: {
        const proxyData = await deployUpgradeableProxyTask(
          taskArgs,
          hre,
          fullyQualifiedContractName,
          factory,
          undefined,
          constructorArgObject,
          initializerArgObject,
          salt
        );
        cumulativeGasUsed = cumulativeGasUsed.add(proxyData.gas);
        break;
      }
      case ONLY_PROXY: {
        const proxyData = await deployOnlyProxyTask(
          taskArgs,
          hre,
          factory,
          salt
        );
        cumulativeGasUsed = cumulativeGasUsed.add(proxyData.gas);
        break;
      }
      case FUNCTION_DEPLOY_CREATE: {
        const deployCreateData = await deployCreateAndRegisterTask(
          taskArgs,
          hre,
          fullyQualifiedContractName,
          factory,
          undefined,
          constructorArgObject,
          salt
        );
        cumulativeGasUsed = cumulativeGasUsed.add(deployCreateData.gas);
        break;
      }
      default: {
        break;
      }
    }
  }
  console.log(`total gas used: ${cumulativeGasUsed.toString()}`);
}
export async function deployUpgradeableProxyTask(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment,
  fullyQaulifiedContractName?: string,
  factory?: AliceNetFactory,
  implementationBase?: ContractFactory,
  constructorArgObject?: ArgData,
  initializerArgObject?: ArgData,
  salt?: string
) {
  let constructorArgs;
  let initializerArgs;
  if (constructorArgObject !== undefined) {
    constructorArgs = Object.values(constructorArgObject);
  } else {
    constructorArgs =
      taskArgs.constructorArgs !== undefined ? taskArgs.constructorArgs : [];
  }
  if (initializerArgObject !== undefined) {
    initializerArgs = Object.values(initializerArgObject);
  } else {
    initializerArgs =
      taskArgs.initializerArgs !== undefined
        ? taskArgs.initializerArgs.split(",")
        : [];
  }
  const waitBlocks = taskArgs.waitConfirmation;
  const contractName =
    fullyQaulifiedContractName === undefined
      ? taskArgs.contractName
      : extractName(fullyQaulifiedContractName);
  // if implementationBase is undefined, get it from the artifacts with the contract name
  implementationBase =
    implementationBase === undefined
      ? ((await hre.ethers.getContractFactory(contractName)) as ContractFactory)
      : implementationBase;
  // if an instance of the factory contract is not provided get it from ethers
  factory =
    factory === undefined
      ? await hre.ethers.getContractAt(
          "AliceNetFactory",
          taskArgs.factoryAddress
        )
      : factory;
  // if the fully qualified contract name is not provided, get it from the artifacts
  fullyQaulifiedContractName =
    fullyQaulifiedContractName === undefined
      ? await getFullyQualifiedName(taskArgs.contractName, hre.artifacts)
      : fullyQaulifiedContractName;
  constructorArgs =
    constructorArgs === undefined ? taskArgs.constructorArgs : constructorArgs;
  const initCallData: string = await encodeInitCallData(
    taskArgs,
    implementationBase,
    initializerArgs
  );

  // if salt is not parsed, get it from the contract itself
  if (salt === undefined) {
    if (taskArgs.salt === undefined) {
      salt = contractName as string;
    } else {
      salt = taskArgs.salt as string;
    }
    salt = hre.ethers.utils.formatBytes32String(salt);
  }
  if (hre.network.name === "hardhat") {
    // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
    // being sent as input to the function (the contract bytecode), so we need to increase the block
    // gas limit temporally in order to deploy the template
    await hre.network.provider.send("evm_setBlockGasLimit", [
      "0x3000000000000000",
    ]);
  }
  let constructorDetails;
  let initializerDetails;
  if (constructorArgObject !== undefined) {
    constructorDetails = JSON.stringify(constructorArgObject);
  } else {
    constructorDetails = constructorArgs;
  }
  if (initializerArgObject !== undefined) {
    initializerDetails = JSON.stringify(initializerArgObject);
  } else {
    initializerDetails = initializerArgs;
  }
  if (taskArgs.skipChecks !== true) {
    const promptMessage = `Do you want to deploy ${contractName} with  constructor arguemnets: ${constructorDetails} initializer Args: ${initializerDetails}? (y/n)`;
    await promptCheckDeploymentArgs(promptMessage);
  }
  const txResponse = await deployUpgradeableGasSafe(
    contractName,
    factory,
    hre.ethers,
    initCallData,
    constructorArgs,
    salt,
    waitBlocks,
    await getGasPrices(hre)
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployedLogicAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  if (taskArgs.verify) {
    await verifyContract(hre, deployedLogicAddress, constructorArgs);
  }
  const proxyData: ProxyData = {
    factoryAddress: factory.address,
    logicName: contractName,
    logicAddress: deployedLogicAddress,
    salt,
    proxyAddress: getEventVar(receipt, EVENT_DEPLOYED_PROXY, CONTRACT_ADDR),
    gas: receipt.gasUsed,
    receipt,
    initCallData,
  };
  await showState(
    `Deployed ${proxyData.logicName} with proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`
  );
  return proxyData;
}

export async function deployCreateTask(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment,
  fullyQaulifiedContractName?: string,
  factory?: AliceNetFactory,
  implementationBase?: ContractFactory,
  constructorArgObject?: ArgData
) {
  const waitBlocks = taskArgs.waitConfirmation;
  let constructorArgs;
  if (constructorArgObject !== undefined) {
    constructorArgs = Object.values(constructorArgObject);
  } else {
    constructorArgs = taskArgs.constructorArgs;
  }
  const contractName =
    fullyQaulifiedContractName === undefined
      ? taskArgs.contractName
      : extractName(fullyQaulifiedContractName);
  factory =
    factory === undefined
      ? await hre.ethers.getContractAt(
          "AliceNetFactory",
          taskArgs.factoryAddress
        )
      : factory;
  const txResponse = await deployCreate(
    contractName,
    factory,
    hre.ethers,
    constructorArgs
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployCreateData: DeployCreateData = {
    name: contractName,
    address: getEventVar(receipt, EVENT_DEPLOYED_RAW, CONTRACT_ADDR),
    factoryAddress: taskArgs.factoryAddress,
    gas: receipt.gasUsed,
    constructorArgs: taskArgs?.constructorArgs,
  };
  if (taskArgs.verify) {
    await verifyContract(hre, factory.address, constructorArgs);
  }
  if (taskArgs.standAlone !== true) {
    await showState(
      `[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${taskArgs.contractName} contract at: ${deployCreateData.address}, gas: ${receipt.gasUsed}`
    );
  } else {
    await showState(
      `Deployed ${deployCreateData.name} at ${deployCreateData.address}, gasCost: ${deployCreateData.gas}`
    );
  }
  deployCreateData.receipt = receipt;
  return deployCreateData;
}
/**
 *
 * @param taskArgs
 * @param hre
 * @param fullyQaulifiedContractName
 * @param factory
 * @param implementationBase
 * @param constructorArgObject object with constructor arguments
 * @param salt bytes32 salt to be used for deployCreate2 address
 * @returns
 */
export async function deployCreate2Task(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment,
  fullyQaulifiedContractName?: string,
  factory?: AliceNetFactory,
  constructorArgObject?: ArgData,
  salt?: string
) {
  const waitBlocks = taskArgs.waitConfirmation;
  let constructorArgs;
  if (constructorArgObject !== undefined) {
    constructorArgs = Object.values(constructorArgObject);
  } else {
    constructorArgs = taskArgs.constructorArgs;
  }
  salt =
    salt === undefined
      ? hre.ethers.utils.formatBytes32String(taskArgs.salt)
      : salt;
  const contractName =
    fullyQaulifiedContractName === undefined
      ? taskArgs.contractName
      : extractName(fullyQaulifiedContractName);
  factory =
    factory === undefined
      ? await hre.ethers.getContractAt(
          "AliceNetFactory",
          taskArgs.factoryAddress
        )
      : factory;
  const txResponse = await deployCreate2(
    contractName,
    factory,
    hre.ethers,
    constructorArgs,
    salt
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployCreate2Data: any = {
    name: contractName,
    address: getEventVar(receipt, EVENT_DEPLOYED_RAW, CONTRACT_ADDR),
    factoryAddress: taskArgs.factoryAddress,
    gas: receipt.gasUsed,
    constructorArgs: taskArgs?.constructorArgs,
  };
  if (taskArgs.verify) {
    await verifyContract(hre, factory.address, constructorArgs);
  }
  if (taskArgs.standAlone !== true) {
    await showState(
      `[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${taskArgs.contractName} contract at: ${deployCreate2Data.address}, gas: ${receipt.gasUsed}`
    );
  } else {
    await showState(
      `Deployed ${deployCreate2Data.name} at ${deployCreate2Data.address}, gasCost: ${deployCreate2Data.gas}`
    );
  }
  deployCreate2Data.receipt = receipt;
  return deployCreate2Data;
}

export async function upgradeProxyTask(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment,
  fullyQaulifiedContractName?: string,
  factory?: AliceNetFactory,
  implementationBase?: ContractFactory,
  constructorArgObject?: ArgData,
  initializerArgObject?: ArgData,
  salt?: string
) {
  let contractName;
  if (fullyQaulifiedContractName === undefined) {
    contractName = taskArgs.contractName;
  } else {
    contractName = extractName(fullyQaulifiedContractName);
  }
  let constructorArgs;
  let initializerArgs;
  if (constructorArgObject !== undefined) {
    constructorArgs = Object.values(constructorArgObject);
  } else {
    constructorArgs = taskArgs.constructorArgs;
  }
  if (initializerArgObject !== undefined) {
    initializerArgs = Object.values(initializerArgObject);
  } else {
    initializerArgs = taskArgs.initializerArgs.split(",");
  }
  factory =
    factory === undefined
      ? await hre.ethers.getContractAt(
          "AliceNetFactory",
          taskArgs.factoryAddress
        )
      : factory;

  implementationBase =
    implementationBase === undefined
      ? ((await hre.ethers.getContractFactory(contractName)) as ContractFactory)
      : implementationBase;

  if (salt === undefined) {
    if (taskArgs.salt === undefined) {
      salt = contractName as string;
    } else {
      salt = taskArgs.salt as string;
    }
    salt = hre.ethers.utils.formatBytes32String(salt);
  }
  const initCallData: string = await encodeInitCallData(
    taskArgs,
    implementationBase,
    initializerArgs
  );
  const txResponse = await upgradeProxyGasSafe(
    contractName,
    factory,
    hre.ethers,
    initCallData,
    constructorArgs,
    salt,
    taskArgs.waitConfirmation,
    await getGasPrices(hre)
  );

  const receipt = await txResponse.wait(taskArgs.waitConfirmation);
  const implementationAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  const proxyAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_PROXY,
    CONTRACT_ADDR
  );
  await showState(
    `Updating logic for the ${taskArgs.contractName} proxy at ${proxyAddress} to point to implementation at ${implementationAddress}, gasCost: ${receipt.gasUsed}`
  );
  const proxyData: ProxyData = {
    factoryAddress: taskArgs.factoryAddress,
    logicName: contractName,
    logicAddress: implementationAddress,
    salt,
    proxyAddress,
    gas: receipt.gasUsed.toNumber(),
    receipt,
    initCallData,
  };
  return proxyData;
}

export async function encodeInitCallData(
  taskArgs: any,
  implementationBase: ContractFactory,
  initializerArgs?: any[]
) {
  if (initializerArgs === undefined) {
    const initializerArgsString: string =
      taskArgs.initializerArgs === undefined ? "" : taskArgs.initializerArgs;
    initializerArgs = initializerArgsString.split(",");
  }
  try {
    return implementationBase.interface.encodeFunctionData(
      "initialize",
      initializerArgs
    );
  } catch (err: any) {
    if (err.reason === "no matching function" && err.value === "initialize") {
      return "0x";
    } else if (err.reason === "types/values length mismatch") {
      throw new InitializerArgsError(
        "Initializer args provided do not match the initializer function"
      );
    } else {
      throw err;
    }
  }
}

export async function promptCheckDeploymentArgs(message: string) {
  let missingInput = true;
  if (process.env.silencer === "true") {
    missingInput = false;
  }

  let dynamicSuggestion = message;
  const defaultSuggestion = dynamicSuggestion;
  while (missingInput) {
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });
    const prompt = (query: any) =>
      new Promise((resolve) => rl.question(query, resolve));
    const answer = await prompt(dynamicSuggestion);
    if (
      answer === "y" ||
      answer === "Y" ||
      answer === "yes" ||
      answer === "Yes" ||
      answer === "YES"
    ) {
      missingInput = false;
      break;
    } else if (
      answer === "n" ||
      answer === "N" ||
      answer === "no" ||
      answer === "No" ||
      answer === "NO"
    ) {
      missingInput = false;
      exit();
    } else {
      if (dynamicSuggestion === defaultSuggestion) {
        dynamicSuggestion =
          "invalid input, enter one of the following: Y, y, yes, Yes, YES, N, n, no, No, NO";
      }
    }
    rl.close();
  }
}

export async function deployCreateAndRegisterTask(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment,
  fullyQualifiedContractName?: string,
  factory?: AliceNetFactory,
  implementationBase?: ContractFactory,
  constructorArgsObject?: ArgData,
  salt?: string
) {
  factory =
    factory === undefined
      ? await hre.ethers.getContractAt(
          "AliceNetFactory",
          taskArgs.factoryAddress
        )
      : factory;
  const waitBlocks = taskArgs.waitConfirmation;
  const contractName =
    fullyQualifiedContractName === undefined
      ? taskArgs.contractName
      : extractName(fullyQualifiedContractName);

  fullyQualifiedContractName =
    fullyQualifiedContractName === undefined
      ? await getFullyQualifiedName(contractName, hre.artifacts)
      : fullyQualifiedContractName;
  if (salt === undefined) {
    salt =
      taskArgs.salt !== undefined
        ? hre.ethers.utils.formatBytes32String(taskArgs.salt)
        : await getBytes32SaltFromContractNSTag(
            contractName,
            hre.artifacts,
            hre.ethers
          );
  }
  let constructorArgs;
  if (constructorArgsObject !== undefined) {
    constructorArgs = Object.values(constructorArgsObject);
  }
  if (await hasConstructorArgs(fullyQualifiedContractName, hre.artifacts)) {
    constructorArgs =
      constructorArgs === undefined
        ? taskArgs.constructorArgs
        : constructorArgs;
    if (constructorArgs === undefined) {
      throw new Error(`No constructor args provided for ${contractName}`);
    }
  } else {
    constructorArgs = [];
  }
  if (hre.network.name === "hardhat") {
    // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
    // being sent as input to the function (the contract bytecode), so we need to increase the block
    // gas limit temporally in order to deploy the template
    await hre.network.provider.send("evm_setBlockGasLimit", [
      "0x3000000000000000",
    ]);
  }
  let constructorDetails;
  if (constructorArgsObject !== undefined) {
    constructorDetails = JSON.stringify(constructorArgsObject);
  } else {
    constructorDetails = constructorArgs;
  }
  if (taskArgs.skipChecks !== true) {
    const promptMessage = `Do you want to deploy ${contractName} with constructorArgs: ${constructorDetails}, salt: ${salt}? (y/n)`;
    await promptCheckDeploymentArgs(promptMessage);
  }
  const txResponse = await deployCreateAndRegister(
    contractName,
    factory,
    hre.ethers,
    constructorArgs,
    salt,
    await getGasPrices(hre)
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployCreateData: DeployCreateData = {
    name: taskArgs.contractName,
    address: getEventVar(receipt, EVENT_DEPLOYED_RAW, CONTRACT_ADDR),
    factoryAddress: taskArgs.factoryAddress,
    gas: receipt.gasUsed,
    constructorArgs: taskArgs?.constructorArgs,
  };
  const network = hre.network.name;
  if (taskArgs.verify) {
    await verifyContract(hre, factory.address, constructorArgs);
  }

  if (taskArgs.standAlone !== true) {
    await showState(
      `[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${taskArgs.contractName} contract at: ${deployCreateData.address}, gas: ${receipt.gasUsed}`
    );
  } else {
    await showState(
      `Deployed ${deployCreateData.name} at ${deployCreateData.address}, gasCost: ${deployCreateData.gas}`
    );
  }
  deployCreateData.receipt = receipt;
  return deployCreateData;
}

export async function deployOnlyProxyTask(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment,
  factory?: AliceNetFactory,
  salt?: string
) {
  const waitBlocks = taskArgs.waitConfirmation;
  factory =
    factory === undefined
      ? await hre.ethers.getContractAt(
          "AliceNetFactory",
          taskArgs.factoryAddress
        )
      : factory;
  if (salt === undefined && taskArgs.salt !== undefined) {
    salt = hre.ethers.utils.formatBytes32String(taskArgs.salt);
  }
  if (salt === undefined) {
    throw new Error("No salt provided");
  }
  const txResponse = await factory.deployProxy(salt, await getGasPrices(hre));
  const receipt = await txResponse.wait(waitBlocks);
  const proxyAddr = getEventVar(receipt, EVENT_DEPLOYED_PROXY, CONTRACT_ADDR);
  const proxyData: ProxyData = {
    proxyAddress: proxyAddr,
    salt,
    factoryAddress: factory.address,
    gas: receipt.gasUsed,
    receipt,
  };

  await showState(
    `Deployed ${salt} proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`
  );
  return proxyData;
}

export async function multiCallDeployUpgradeableTask(taskArgs: any, hre: any) {
  const waitBlocks = taskArgs.waitConfirmation;
  const implementationBase = (await hre.ethers.getContractFactory(
    taskArgs.contractName
  )) as ContractFactory;
  const network = hre.network.name;
  const factory = await hre.ethers.getContractAt(
    "AliceNetFactory",
    taskArgs.factoryAddress
  );
  const initArgs =
    taskArgs.initializerArgs === undefined
      ? []
      : taskArgs.initializerArgs.replace(/\s+/g, "").split(",");
  const fullname = (await getFullyQualifiedName(
    taskArgs.contractName,
    hre.artifacts
  )) as string;

  const initCallData = (await isInitializable(fullname, hre.artifacts))
    ? implementationBase.interface.encodeFunctionData(
        FUNCTION_INITIALIZE,
        initArgs
      )
    : "0x";
  const salt: BytesLike = await getBytes32SaltFromContractNSTag(
    taskArgs.contractName,
    hre.artifacts,
    hre.ethers
  );
  const constructorArgs =
    taskArgs.constructorArgs === undefined ? [] : taskArgs.constructorArgs;
  // encode deployBcode
  if (hre.network.name === "hardhat") {
    // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
    // being sent as input to the function (the contract bytecode), so we need to increase the block
    // gas limit temporally in order to deploy the template
    await hre.network.provider.send("evm_setBlockGasLimit", [
      "0x3000000000000000",
    ]);
  }
  const txResponse = await multiCallDeployUpgradeable(
    implementationBase,
    factory,
    hre.ethers,
    initCallData,
    constructorArgs,
    salt,
    await getGasPrices(hre)
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployedLogicAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  if (taskArgs.verify) {
    await verifyContract(hre, deployedLogicAddress, constructorArgs);
  }
  const proxyData: ProxyData = {
    factoryAddress: taskArgs.factoryAddress,
    logicName: taskArgs.contractName,
    logicAddress: deployedLogicAddress,
    salt,
    proxyAddress: getEventVar(receipt, EVENT_DEPLOYED_PROXY, CONTRACT_ADDR),
    gas: receipt.gasUsed,
    receipt,
    initCallData,
  };
  await showState(
    `Deployed ${proxyData.logicName} with proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`
  );

  return proxyData;
}

export async function checkUserDirPath(path: string) {
  if (path !== undefined) {
    if (!fs.existsSync(path)) {
      console.log(
        "Creating Folder at" + path + " since it didn't exist before!"
      );
      fs.mkdirSync(path);
    }
    if (fs.statSync(path).isFile()) {
      throw new Error("outputFolder path should be to a directory not a file");
    }
  }
}

export async function isInitializable(
  fullyQualifiedName: string,
  artifacts: Artifacts
) {
  const buildInfo: any = await artifacts.getBuildInfo(fullyQualifiedName);
  const path = extractPath(fullyQualifiedName);
  const name = extractName(fullyQualifiedName);
  const methods = buildInfo.output.contracts[path][name].abi;
  for (const method of methods) {
    if (method.name === FUNCTION_INITIALIZE) {
      return true;
    }
  }
  return false;
}

export async function hasConstructorArgs(
  fullName: string,
  artifacts: Artifacts
) {
  const buildInfo: any = await artifacts.getBuildInfo(fullName);
  const path = extractPath(fullName);
  const name = extractName(fullName);
  const methods = buildInfo.output.contracts[path][name].abi;
  for (const method of methods) {
    if (method.type === "constructor") {
      return method.inputs.length > 0;
    }
  }
  return false;
}
/**
 * @description encodes init call state input to be used by the custom hardhat tasks
 * @param args values of the init call state as an array of strings where each string represents variable value
 * @returns the args array as a comma delimited string
 */
export async function getEncodedInitCallData(
  args: Array<string> | undefined
): Promise<string | undefined> {
  if (args !== undefined) {
    return args.toString();
  }
}

export async function getContract(name: string, artifacts: Artifacts) {
  const artifactPaths = await artifacts.getAllFullyQualifiedNames();
  for (let i = 0; i < artifactPaths.length; i++) {
    if (artifactPaths[i].split(":")[1] === name) {
      return String(artifactPaths[i]);
    }
  }
}

export async function getAllContracts(artifacts: Artifacts) {
  // get a list with all the contract names
  return await artifacts.getAllFullyQualifiedNames();
}

export function extractPath(fullName: string) {
  return fullName.split(":")[0];
}

export function extractName(fullName: string) {
  return fullName.split(":")[1];
}

export async function getCustomNSTag(
  fullyQaulifiedContractName: string,
  tagName: string,
  artifacts: Artifacts
): Promise<string> {
  const buildInfo = await artifacts.getBuildInfo(fullyQaulifiedContractName);
  if (buildInfo !== undefined) {
    const name = extractName(fullyQaulifiedContractName);
    const path = extractPath(fullyQaulifiedContractName);
    const info: any = buildInfo?.output.contracts[path][name];
    return info.devdoc[`custom:${tagName}`];
  } else {
    throw new Error(`Failed to get natspec tag ${tagName}`);
  }
}

/**
 * @description gets the salt specified in the contracts head with a custom natspec tag @custom:salt
 * @param contractName name of the contract
 * @param artifacts artifacts object from hardhat artifacts
 * @param ethers ethersjs object
 * @returns bytes32 formatted salt
 */
export async function getBytes32SaltFromContractNSTag(
  contractName: string,
  artifacts: Artifacts,
  ethers: Ethers,
  fullyQualifiedName?: string
): Promise<string> {
  fullyQualifiedName =
    fullyQualifiedName === undefined
      ? await getFullyQualifiedName(contractName, artifacts)
      : fullyQualifiedName;
  const salt = await getSalt(fullyQualifiedName, artifacts);
  return ethers.utils.formatBytes32String(salt);
}

export async function getSalt(fullName: string, artifacts: Artifacts) {
  return await getCustomNSTag(fullName, "salt", artifacts);
}

export async function getBytes32Salt(
  contractName: string,
  artifacts: Artifacts,
  ethers: Ethers
) {
  const fullName = await getFullyQualifiedName(contractName, artifacts);
  const salt: string = await getSalt(fullName, artifacts);
  return ethers.utils.formatBytes32String(salt);
}

export async function getFullyQualifiedName(
  contractName: string,
  artifacts: Artifacts
) {
  const contractArtifact = await artifacts.readArtifact(contractName);
  const path = contractArtifact.sourceName;
  return path + ":" + contractName;
}
export async function getDeployType(fullName: string, artifacts: Artifacts) {
  return await getCustomNSTag(fullName, "deploy-type", artifacts);
}

export async function getDeployGroup(fullName: string, artifacts: Artifacts) {
  return await getCustomNSTag(fullName, "deploy-group", artifacts);
}

export async function getDeployGroupIndex(
  fullName: string,
  artifacts: Artifacts
) {
  return await getCustomNSTag(fullName, "deploy-group-index", artifacts);
}

export async function getGasPrices(hre: HardhatRuntimeEnvironment) {
  // get the latest block
  const latestBlock = await hre.ethers.provider.getBlock("latest");
  // get the previous basefee from the latest block
  const _blockBaseFee = latestBlock.baseFeePerGas;
  if (_blockBaseFee === undefined || _blockBaseFee === null) {
    throw new Error("undefined block base fee per gas");
  }
  const blockBaseFee = _blockBaseFee.toBigInt();
  // miner tip
  let maxPriorityFeePerGas: bigint;
  const network = await hre.ethers.provider.getNetwork();
  const minValue = hre.ethers.utils.parseUnits("2.0", "gwei").toBigInt();
  if (network.chainId === 1337) {
    maxPriorityFeePerGas = minValue;
  } else {
    maxPriorityFeePerGas = BigInt(
      await hre.network.provider.send("eth_maxPriorityFeePerGas")
    );
  }
  maxPriorityFeePerGas = (maxPriorityFeePerGas * 125n) / 100n;
  maxPriorityFeePerGas =
    maxPriorityFeePerGas < minValue ? minValue : maxPriorityFeePerGas;
  const maxFeePerGas = 2n * blockBaseFee + maxPriorityFeePerGas;
  return { maxPriorityFeePerGas, maxFeePerGas };
}

export async function verifyContract(
  hre: HardhatRuntimeEnvironment,
  deployedContractAddress: string,
  constructorArgs: Array<any>
) {
  let result;
  try {
    result = await hre.run("verify", {
      network: hre.network.name,
      address: deployedContractAddress,
      constructorArgsParams: constructorArgs,
    });
  } catch (error) {
    console.log(
      `Failed to automatically verify ${deployedContractAddress} please do it manually!`
    );
    console.log(error);
  }
  return result;
}

export const showState = async (message: string): Promise<void> => {
  if (process.env.silencer === undefined || process.env.silencer === "false") {
    console.log(message);
  }
};

export async function generateDeployConfigTemplate(
  list: DeploymentList,
  artifacts: Artifacts,
  ethers: Ethers
): Promise<DeploymentConfigWrapper> {
  const deploymentArgs: DeploymentConfigWrapper = {};
  const factoryName = await getFullyQualifiedName("AliceNetFactory", artifacts);

  const factoryContractInfo = await extractFullContractInfo(
    factoryName,
    artifacts,
    ethers
  );
  deploymentArgs[factoryName] = factoryContractInfo;

  // iterate over the deployment list
  for (const key of Object.keys(list)) {
    const arrayOfConfigs = list[key];
    for (const deploymentConfig of arrayOfConfigs) {
      deploymentArgs[deploymentConfig.fullyQualifiedName] = deploymentConfig;
    }
  }

  return deploymentArgs;
}

export async function getSortedDeployList(
  contracts: Array<string>,
  artifacts: Artifacts,
  ethers: Ethers
) {
  const deploymentList: DeploymentList = {};
  for (const contract of contracts) {
    const contractInfo = await extractFullContractInfo(
      contract,
      artifacts,
      ethers
    );

    const deployType: string | undefined = contractInfo?.deployType;
    let group: string | undefined = contractInfo?.deployGroup;

    if (group !== undefined) {
      if (contractInfo?.deployGroupIndex === undefined) {
        throw new Error(
          "If deploy-group-index is specified a deploy-group-index also should be!"
        );
      }
      try {
        // check deploy group index exists
        parseInt(contractInfo?.deployGroupIndex);
      } catch (error) {
        throw new Error(
          `Failed to convert deploy-group-index for contract ${contract}! deploy-group-index should be an integer!`
        );
      }
    } else {
      group = "general";
    }
    if (deployType !== undefined) {
      if (deploymentList[group] === undefined) {
        deploymentList[group] = [];
      }
      deploymentList[group].push(contractInfo);
    }
  }
  for (const key in deploymentList) {
    if (key !== "general") {
      deploymentList[key].sort((contractA, contractB) => {
        const indexA = parseInt(contractA.deployGroupIndex);
        const indexB = parseInt(contractB.deployGroupIndex);
        return indexA - indexB;
      });
    }
  }
  return deploymentList;
}

export async function extractFullContractInfo(
  fullName: string,
  artifacts: Artifacts,
  ethers: Ethers
): Promise<DeploymentConfig> {
  let constructorArgs: ArgData = {};
  let initializerArgs: ArgData = {};

  const buildInfo = await artifacts.getBuildInfo(fullName);
  if (buildInfo !== undefined) {
    const name = extractName(fullName);
    const path = extractPath(fullName);
    const info: any = buildInfo?.output.contracts[path][name];

    const methods = buildInfo.output.contracts[path][name].abi;
    for (const method of methods) {
      if (method.type === "constructor" || method.name === "initialize") {
        const args: ArgData = {};
        for (const input of method.inputs) {
          args[input.name] = "UNDEFINED";
        }

        if (method.type === "constructor") {
          constructorArgs = args;
        } else {
          initializerArgs = args;
        }
      }
    }

    const salt =
      info.devdoc["custom:salt"] !== undefined
        ? ethers.utils.formatBytes32String(info.devdoc["custom:salt"])
        : "";
    const deployGroup = info.devdoc["custom:deploy-group"];
    const deployGroupIndex = info.devdoc["custom:deploy-group-index"];
    const deployType = info.devdoc["custom:deploy-type"];
    const deploymentConfig: DeploymentConfig = {
      name,
      fullyQualifiedName: fullName,
      salt,
      deployGroup,
      deployGroupIndex,
      deployType,
      constructorArgs,
      initializerArgs,
    };
    return deploymentConfig;
  } else {
    throw new Error(`failed to fetch ${fullName} info`);
  }
}

export async function writeDeploymentConfig(
  deploymentArgs: DeploymentConfigWrapper,
  configFile?: string
) {
  const file = configFile === undefined ? DEFAULT_CONFIG_FILE_PATH : configFile;
  const jsonData = JSON.stringify(deploymentArgs, null, 2);
  fs.writeFileSync(file, jsonData);
  return file;
}

export async function readDeploymentConfig(
  file: string
): Promise<DeploymentConfigWrapper> {
  return await readJSON(file);
}

export async function readJSON(file: string) {
  const rawData = fs.readFileSync(file);
  return JSON.parse(rawData.toString("utf8"));
}
