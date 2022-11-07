import { HardhatEthersHelpers } from "@nomiclabs/hardhat-ethers/types";
import { BigNumber, BigNumberish, BytesLike, ContractFactory } from "ethers";
import fs from "fs";
import { Artifacts, HardhatRuntimeEnvironment } from "hardhat/types";
import { exit } from "process";
import readline from "readline";
import { AliceNetFactory } from "../../../typechain-types";
import {
  deployCreate,
  deployCreateAndRegister,
  deployFactory,
  deployUpgradeableGasSafe,
  encodeMultiCallArgs,
  getEventVar,
  multiCallDeployUpgradeable,
  upgradeProxyGasSafe,
} from "../alicenetFactory";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  DEFAULT_CONFIG_DIR,
  EVENT_DEPLOYED_PROXY,
  EVENT_DEPLOYED_RAW,
  FUNCTION_DEPLOY_CREATE,
  FUNCTION_INITIALIZE,
  FUNCTION_UPGRADE_PROXY,
  ONLY_PROXY,
  UPGRADEABLE_DEPLOYMENT,
} from "../constants";
import {
  readDeploymentArgs,
  readDeploymentListConfig,
} from "./deploymentConfigUtil";
import {
  DeployCreateData,
  FactoryData,
  ProxyData,
  updateDefaultFactoryData,
  updateDeployCreateList,
  updateExternalContractList,
  updateProxyList,
} from "./factoryStateUtil";

type Ethers = typeof import("../../../node_modules/ethers/lib/ethers") &
  HardhatEthersHelpers;

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

export type DeployProxyMCArgs = {
  contractName: string;
  logicAddress: string;
  waitConfirmation?: number;
  factoryAddress?: string;
  initCallData?: BytesLike;
  outputFolder?: string;
};
export type MultiCallArgsStruct = {
  target: string;
  value: BigNumberish;
  data: BytesLike;
};
export type DeployArgs = {
  contractName: string;
  factoryAddress: string;
  waitConfirmation?: number;
  initCallData?: string;
  constructorArgs?: any;
  outputFolder?: string;
  verify?: boolean;
  standAlone?: boolean;
};

export type Args = {
  contractName: string;
  factoryAddress?: string;
  salt?: BytesLike;
  initCallData?: string;
  constructorArgs?: any;
  outputFolder?: string;
};
export interface InitData {
  constructorArgs: { [key: string]: any };
  initializerArgs: { [key: string]: any };
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

export async function getContractDescriptor(
  contractName: string,
  constructorArgs: Array<any>,
  initializerArgs: Array<any>,
  hre: HardhatRuntimeEnvironment
): Promise<ContractDescriptor> {
  const fullyQualifiedName = await getFullyQualifiedName(
    contractName,
    hre.artifacts
  );
  return {
    name: contractName,
    fullyQualifiedName,
    deployGroup: await getDeployGroup(fullyQualifiedName, hre.artifacts),
    deployGroupIndex: parseInt(
      await getDeployGroupIndex(fullyQualifiedName, hre.artifacts),
      10
    ),
    deployType: await getDeployType(fullyQualifiedName, hre.artifacts),
    constructorArgs,
    initializerArgs,
  };
}

// AliceNet Factory Task Functions

export async function deployFactoryTask(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment,
  constructorArgs?: Array<any>
) {
  await checkUserDirPath(taskArgs.outputFolder);
  const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
  // if the user didnt specify constructor args, get from the default deployment config list

  constructorArgs =
    constructorArgs === undefined ? taskArgs.constructorArgs : constructorArgs;
  if (constructorArgs === undefined) {
    constructorArgs = [];
  }
  const signers = await hre.ethers.getSigners();
  const deployTX = factoryBase.getDeployTransaction(constructorArgs[0]);
  const gasCost = await hre.ethers.provider.estimateGas(deployTX);
  // deploys the factory
  const factory = await deployFactory(
    constructorArgs[0],
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
    await verifyContract(hre, factory.address, constructorArgs);
  }
  const network = hre.network.name;
  await updateDefaultFactoryData(network, factoryData, taskArgs.outputFolder);
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
  await checkUserDirPath(taskArgs.outputFolder);
  const deploymentListConfig = await readDeploymentListConfig(
    taskArgs.inputFolder
  );
  // setting listName undefined will use the default list
  // deploy the factory first
  const keys = Object.keys(deploymentListConfig);

  const constructorArgs = [
    deploymentListConfig[keys[0]].constructorArgs.legacyToken_,
  ];
  let factoryAddress = taskArgs.factoryAddress;
  if (factoryAddress === undefined) {
    const factoryData: FactoryData = await deployFactoryTask(
      taskArgs,
      hre,
      constructorArgs
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
    constructorArgs = taskArgs.constructorArgs;
  }
  if (initializerArgObject !== undefined) {
    initializerArgs = Object.values(initializerArgObject);
  } else {
    initializerArgs = taskArgs.initCallData.split(",");
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
  let txResponse = await deployCreate(
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
  }
  if (initializerArgObject !== undefined) {
    initializerArgs = Object.values(initializerArgObject);
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
    taskArgs.initializerArgs
  );
  let txResponse = await upgradeProxyGasSafe(
    contractName,
    factory,
    hre.ethers,
    initCallData,
    taskArgs.constructorArgs,
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
    logicName: taskArgs.contractName,
    logicAddress: taskArgs.logicAddress,
    salt,
    proxyAddress,
    gas: receipt.gasUsed.toNumber(),
    receipt,
    initCallData,
  };
  return proxyData;
}

// export async function muiltiCallDeployImplementationAndUpgradeProxyTask(
//   taskArgs: any,
//   hre: HardhatRuntimeEnvironment
// ) {
//   const contractName = taskArgs.contractName;
//   const factory = await hre.ethers.getContractAt(
//     "AliceNetFactory",
//     taskArgs.factoryAddress
//   );
//   const implementationBase = (await hre.ethers.getContractFactory(
//     contractName
//   )) as ContractFactory;
//   const salt = await getBytes32SaltFromContractNSTag(
//     contractName,
//     hre.artifacts,
//     hre.ethers
//   );
//   const initCallData: string = await encodeInitCallData(
//     taskArgs,
//     implementationBase,
//     taskArgs.initializerArgs
//   );
//   const txResponse = await multiCallUpgradeProxy(
//     contractName,
//     factory,
//     hre.ethers,
//     initCallData,
//     taskArgs.constructorArgs,
//     salt
//   );
//   const receipt = await txResponse.wait(taskArgs.waitConfirmation);
//   const implementationAddress = getEventVar(
//     receipt,
//     EVENT_DEPLOYED_RAW,
//     CONTRACT_ADDR
//   );
//   const proxyAddress = getEventVar(
//     receipt,
//     EVENT_DEPLOYED_PROXY,
//     CONTRACT_ADDR
//   );
//   await showState(
//     `Updating logic for the ${taskArgs.contractName} proxy at ${proxyAddress} to point to implementation at ${implementationAddress}, gasCost: ${receipt.gasUsed}`
//   );
//   const proxyData: ProxyData = {
//     factoryAddress: taskArgs.factoryAddress,
//     logicName: taskArgs.contractName,
//     logicAddress: taskArgs.logicAddress,
//     salt,
//     proxyAddress,
//     gas: receipt.gasUsed.toNumber(),
//     receipt,
//     initCallData,
//   };
//   return proxyData;
// }

export async function encodeInitCallData(
  taskArgs: any,
  implementationBase: ContractFactory,
  initializerArgs?: any[]
) {
  if (initializerArgs === undefined) {
    initializerArgs =
      taskArgs.initCallData === undefined
        ? ""
        : taskArgs.initCallData.split(",");
  }
  try {
    return implementationBase.interface.encodeFunctionData(
      "initialize",
      initializerArgs
    );
  } catch (err: any) {
    if (err.reason === "no matching function" && err.value === "initialize") {
      return "0x";
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
  fullyQaulifiedContractName?: string,
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
    fullyQaulifiedContractName === undefined
      ? taskArgs.contractName
      : extractName(fullyQaulifiedContractName);

  fullyQaulifiedContractName =
    fullyQaulifiedContractName === undefined
      ? await getFullyQualifiedName(contractName, hre.artifacts)
      : fullyQaulifiedContractName;
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
  if (await hasConstructorArgs(fullyQaulifiedContractName, hre.artifacts)) {
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
  await updateDeployCreateList(
    network,
    deployCreateData,
    taskArgs.outputFolder
  );
  if (taskArgs.standAlone !== true) {
    await showState(
      `[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${taskArgs.contractName} contract at: ${deployCreateData.address}, gas: ${receipt.gasUsed}`
    );
  } else {
    await showState(
      `Deployed ${deployCreateData.name} at ${deployCreateData.address}, gasCost: ${deployCreateData.gas}`
    );
    await updateExternalContractList(
      network,
      deployCreateData,
      taskArgs.outputFolder
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
    taskArgs.initCallData === undefined
      ? []
      : taskArgs.initCallData.replace(/\s+/g, "").split(",");
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
  await updateProxyList(network, proxyData, taskArgs.outputFolder);
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

export async function getDeployMetaArgs(
  fullyQualifiedName: string,
  waitConfirmation: number,
  factoryAddress: string,
  artifacts: Artifacts,
  inputFolder?: string,
  outputFolder?: string,
  verify?: boolean
): Promise<DeployArgs> {
  let initCallData;
  // check if contract needs to be initialized
  const initAble = await isInitializable(fullyQualifiedName, artifacts);
  if (initAble) {
    const initializerArgs = await getDeploymentInitializerArgs(
      fullyQualifiedName,
      inputFolder
    );
    initCallData = await getEncodedInitCallData(initializerArgs);
  }
  const hasConArgs = await hasConstructorArgs(fullyQualifiedName, artifacts);
  const constructorArgs = hasConArgs
    ? await getDeploymentConstructorArgs(fullyQualifiedName, inputFolder)
    : undefined;
  return {
    contractName: extractName(fullyQualifiedName),
    waitConfirmation,
    factoryAddress,
    initCallData,
    constructorArgs,
    outputFolder,
    verify,
  };
}

export async function getFactoryDeploymentArgs(
  artifacts: Artifacts,
  inputFolder?: string
) {
  const fullyQualifiedName = await getFullyQualifiedName(
    "AliceNetFactory",
    artifacts
  );
  const hasConArgs = await hasConstructorArgs(fullyQualifiedName, artifacts);
  const constructorArgs = hasConArgs
    ? await getDeploymentConstructorArgs(fullyQualifiedName, inputFolder)
    : [];
  return constructorArgs;
}

// export async function getDeployUpgradeableProxyArgs(
//   fullyQualifiedName: string,
//   factoryAddress: string,
//   artifacts: Artifacts,
//   waitConfirmation?: number,
//   inputFolder?: string,
//   outputFolder?: string,
//   verify?: boolean
// ): Promise<DeployArgs> {
//   let initCallData;
//   const initAble = await isInitializable(fullyQualifiedName, artifacts);
//   if (initAble) {
//     const initializerArgs = await getDeploymentInitializerArgs(
//       fullyQualifiedName,
//       inputFolder
//     );
//     initCallData = await getEncodedInitCallData(initializerArgs);
//   }
//   const hasConArgs = await hasConstructorArgs(fullyQualifiedName, artifacts);
//   const constructorArgs = hasConArgs
//     ? await getDeploymentConstructorArgs(fullyQualifiedName, inputFolder)
//     : undefined;
//   return {
//     contractName: extractName(fullyQualifiedName),
//     waitConfirmation,
//     factoryAddress,
//     initCallData,
//     constructorArgs,
//     outputFolder,
//     verify,
//   };
// }

// export async function getDeployCreateArgs(
//   fullyQualifiedName: string,
//   factoryAddress: string,
//   artifacts: Artifacts,
//   waitConfirmation?: number,
//   inputFolder?: string,
//   outputFolder?: string,
//   verify?: boolean,
//   standAlone?: boolean
// ): Promise<DeployArgs> {
//   let initCallData;
//   const initAble = await isInitializable(fullyQualifiedName, artifacts);
//   if (initAble) {
//     const initializerArgs = await getDeploymentInitializerArgs(
//       fullyQualifiedName,
//       inputFolder
//     );
//     initCallData = await getEncodedInitCallData(initializerArgs);
//   }
//   const hasConArgs = await hasConstructorArgs(fullyQualifiedName, artifacts);
//   const constructorArgs = hasConArgs
//     ? await getDeploymentConstructorArgs(fullyQualifiedName, inputFolder)
//     : undefined;
//   return {
//     contractName: extractName(fullyQualifiedName),
//     waitConfirmation,
//     factoryAddress,
//     initCallData,
//     constructorArgs,
//     outputFolder,
//     verify,
//     standAlone,
//   };
// }

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

// return a list of constructor inputs for each contract
export async function getDeploymentConstructorArgs(
  fullName: string,
  configDirPath?: string
) {
  let output: Array<ArgData> = [];
  // get the deployment args
  const path =
    configDirPath === undefined
      ? DEFAULT_CONFIG_DIR + "/deploymentArgsTemplate"
      : configDirPath + "/deploymentArgsTemplate";
  const deploymentConfig: any = await readDeploymentArgs(path);
  if (deploymentConfig !== undefined) {
    const deploymentArgs: DeploymentArgs = {
      constructor: deploymentConfig.constructor,
      initializer: deploymentConfig.initializer,
    };
    if (
      deploymentArgs.constructor !== undefined &&
      deploymentArgs.constructor[fullName] !== undefined
    ) {
      output = extractArgs(deploymentArgs.constructor[fullName]);
    }
  } else {
    output = [];
  }
  return output;
}

export function extractArgs(input: ArgData) {
  const output: Array<any> = [];
  for (const key in input) {
    output.push(input[key]);
  }
  return output;
}

// return a list of initializer inputs for each contract
export async function getDeploymentInitializerArgs(
  fullName: string,
  configDirPath?: string
) {
  let output: Array<string> | undefined;
  const path =
    configDirPath === undefined
      ? DEFAULT_CONFIG_DIR + "/deploymentArgsTemplate"
      : configDirPath + "/deploymentArgsTemplate";
  const deploymentConfig: any = await readDeploymentArgs(path);
  if (deploymentConfig !== undefined) {
    const deploymentArgs: DeploymentArgs = deploymentConfig;
    if (
      deploymentArgs.initializer !== undefined &&
      deploymentArgs.initializer[fullName] !== undefined
    ) {
      output = extractArgs(deploymentArgs.initializer[fullName]);
    }
  } else {
    output = undefined;
  }
  return output;
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

export async function getDeployUpgradeableMultiCallArgs(
  contractDescriptor: ContractDescriptor,
  hre: HardhatRuntimeEnvironment,
  factoryAddr: string,
  inputedInitCallData?: string
) {
  const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
  const factory = factoryBase.attach(factoryAddr);
  const logicContract: ContractFactory = await hre.ethers.getContractFactory(
    contractDescriptor.name
  );
  const logicFactory = await hre.ethers.getContractFactory(
    contractDescriptor.name
  );
  const deployTxReq = logicContract.getDeployTransaction(
    ...contractDescriptor.constructorArgs
  );
  let initCallData = "0x";
  if (inputedInitCallData) {
    initCallData = inputedInitCallData;
  }
  if (contractDescriptor.initializerArgs.length > 0)
    initCallData = logicFactory.interface.encodeFunctionData(
      FUNCTION_INITIALIZE,
      contractDescriptor.initializerArgs
    );
  const salt = await getBytes32Salt(
    contractDescriptor.name,
    hre.artifacts,
    hre.ethers
  );
  const nonce = await hre.ethers.provider.getTransactionCount(factory.address);
  const logicAddress = hre.ethers.utils.getContractAddress({
    from: factory.address,
    nonce,
  });

  // encode deploy create
  const deployCreateCallData: BytesLike =
    factoryBase.interface.encodeFunctionData(FUNCTION_DEPLOY_CREATE, [
      deployTxReq.data,
    ]);
  // encode the deployProxy function call with Salt as arg
  const deployProxyCallData: BytesLike =
    factoryBase.interface.encodeFunctionData("deployProxy", [salt]);
  // encode upgrade proxy multicall
  const upgradeProxyCallData: BytesLike =
    factoryBase.interface.encodeFunctionData(FUNCTION_UPGRADE_PROXY, [
      salt,
      logicAddress,
      initCallData,
    ]);
  const deployCreate = encodeMultiCallArgs(
    factory.address,
    0,
    deployCreateCallData
  );
  const deployProxy = encodeMultiCallArgs(
    factory.address,
    0,
    deployProxyCallData
  );
  const upgradeProxy = encodeMultiCallArgs(
    factory.address,
    0,
    upgradeProxyCallData
  );
  const multiCallArgs = [deployCreate, deployProxy, upgradeProxy];
  return multiCallArgs;
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

export async function deployContractsMulticall(
  contracts: ContractDescriptor[],
  hre: HardhatRuntimeEnvironment,
  factoryAddr: string
) {
  const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
  const factory = factoryBase.attach(factoryAddr);
  for (let i = 0; i < contracts.length; i++) {
    const contract = contracts[i];
    const deployType = contract.deployType;
    switch (deployType) {
      case UPGRADEABLE_DEPLOYMENT: {
        const multiCallArgsArray = [];
        const [deployCreate, deployProxy, upgradeProxy] =
          await getDeployUpgradeableMultiCallArgs(
            contract,
            hre,
            factory.address
          );
        multiCallArgsArray.push(deployCreate);
        multiCallArgsArray.push(deployProxy);
        multiCallArgsArray.push(upgradeProxy);
        const txResponse = await factory.multiCall(multiCallArgsArray);
        const receipt = await txResponse.wait();
        const address = getEventVar(
          receipt,
          EVENT_DEPLOYED_PROXY,
          CONTRACT_ADDR
        );
        console.log(contract.name, "deployed at:", address);
        break;
      }
      case ONLY_PROXY: {
        const name = extractName(contract.fullyQualifiedName);
        const salt: BytesLike = await getBytes32Salt(
          name,
          hre.artifacts,
          hre.ethers
        );
        const factoryAddress = factory.address;
        await hre.run("deploy-proxy", {
          factoryAddress,
          salt,
        });
        break;
      }
      default: {
        break;
      }
    }
  }
}
