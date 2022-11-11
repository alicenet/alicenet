import { BigNumber, ContractFactory } from "ethers";
import { ethers } from "hardhat";
import { HardhatRuntimeEnvironment } from "hardhat/types";
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
  upgradeProxyGasSafe,
} from "../alicenetFactory";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  EVENT_DEPLOYED_PROXY,
  EVENT_DEPLOYED_RAW,
  FUNCTION_DEPLOY_CREATE,
  ONLY_PROXY,
  UPGRADEABLE_DEPLOYMENT,
} from "../constants";
import {
  ArgData,
  DeployCreateData,
  DeploymentConfig,
  DeploymentConfigWrapper,
  FactoryData,
  InitializerArgsError,
  ProxyData,
} from "./interfaces";
import {
  extractNameFromFullyQualifiedName as extractContractNameFromFullyQualifiedName,
  getBytes32SaltFromContractNSTag,
  getFullyQualifiedName,
  getGasPrices,
  hasConstructorArgs,
  readDeploymentConfig,
  showState,
  verifyContract,
} from "./utils";

type Ethers = typeof ethers;

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
    await getGasPrices(hre.ethers)
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

  const deploymentConfig: DeploymentConfigWrapper = readDeploymentConfig(
    taskArgs.configFile
  );

  const expectedContractFullQualifiedName =
    "contracts/AliceNetFactory.sol:AliceNetFactory";
  const expectedField = "legacyToken_";

  if (
    deploymentConfig[expectedContractFullQualifiedName].constructorArgs[
      expectedField
    ] === undefined
  ) {
    throw new Error(
      `Couldn't find ${expectedField} in the constructor area for` +
        ` ${expectedContractFullQualifiedName} inside ${taskArgs.configFile}`
    );
  }

  const legacyTokenAddress: string = deploymentConfig[
    expectedContractFullQualifiedName
  ].constructorArgs[expectedField] as string;
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

  for (const fullyQualifiedContractName in deploymentConfig) {
    const deploymentConfigForContract =
      deploymentConfig[fullyQualifiedContractName];
    const deployType = deploymentConfig[fullyQualifiedContractName].deployType;
    const salt = deploymentConfig[fullyQualifiedContractName].salt;
    const constructorArgObject =
      deploymentConfig[fullyQualifiedContractName].constructorArgs;

    switch (deployType) {
      case UPGRADEABLE_DEPLOYMENT: {
        const proxyData = await deployUpgradeableProxyTask(
          deploymentConfigForContract,
          taskArgs.waitConfirmation,
          hre,
          undefined,
          factory,
          taskArgs.skipChecks,
          taskArgs.verify
        );
        cumulativeGasUsed = cumulativeGasUsed.add(proxyData.gas);
        break;
      }
      case ONLY_PROXY: {
        const proxyData = await deployOnlyProxyTask(
          taskArgs,
          hre.ethers,
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
  deploymentConfigForContract: DeploymentConfig,
  waitBlocks: number = 0,
  hre: HardhatRuntimeEnvironment,
  factoryAddress?: string,
  factory?: AliceNetFactory,
  skipChecks: boolean = false,
  verify: boolean = false
) {
  const constructorArgs = Object.values(
    deploymentConfigForContract.constructorArgs
  );
  const initializerArgs = Object.values(
    deploymentConfigForContract.initializerArgs
  );

  // if implementationBase is undefined, get it from the artifacts with the contract name
  const implementationBase = (await hre.ethers.getContractFactory(
    deploymentConfigForContract.name
  )) as ContractFactory;

  // if an instance of the factory contract is not provided get it from ethers
  if (factory === undefined) {
    if (factoryAddress === undefined) {
      throw new Error("Either factory or factoryAddress must be provided");
    } else {
      factory = await hre.ethers.getContractAt(
        "AliceNetFactory",
        factoryAddress
      );
    }
  }

  const initCallData: string = await encodeInitCallData(
    implementationBase,
    initializerArgs
  );

  if (hre.network.name === "hardhat") {
    // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
    // being sent as input to the function (the contract bytecode), so we need to increase the block
    // gas limit temporally in order to deploy the template
    await hre.network.provider.send("evm_setBlockGasLimit", [
      "0x3000000000000000",
    ]);
  }

  if (!skipChecks) {
    const constructorDetails = JSON.stringify(
      deploymentConfigForContract.constructorArgs
    );
    const initializerDetails = JSON.stringify(
      deploymentConfigForContract.initializerArgs
    );

    const promptMessage = `Do you want to deploy ${deploymentConfigForContract.name} with  constructor arguments: ${constructorDetails} initializer Args: ${initializerDetails}? (y/n)`;
    await promptCheckDeploymentArgs(promptMessage);
  }
  const txResponse = await deployUpgradeableGasSafe(
    deploymentConfigForContract.name,
    factory,
    hre.ethers,
    initCallData,
    constructorArgs,
    deploymentConfigForContract.salt,
    waitBlocks,
    await getGasPrices(hre.ethers)
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployedLogicAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  if (verify) {
    await verifyContract(hre, deployedLogicAddress, constructorArgs);
  }
  const proxyData: ProxyData = {
    factoryAddress: factory.address,
    logicName: deploymentConfigForContract.name,
    logicAddress: deployedLogicAddress,
    salt: deploymentConfigForContract.salt,
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
      : extractContractNameFromFullyQualifiedName(fullyQaulifiedContractName);
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
 * @param fullyQualifiedContractName
 * @param factory
 * @param implementationBase
 * @param constructorArgObject object with constructor arguments
 * @param salt bytes32 salt to be used for deployCreate2 address
 * @returns
 */
export async function deployCreate2Task(
  taskArgs: any,
  hre: HardhatRuntimeEnvironment,
  fullyQualifiedContractName?: string,
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
    fullyQualifiedContractName === undefined
      ? taskArgs.contractName
      : extractContractNameFromFullyQualifiedName(fullyQualifiedContractName);
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
  fullyQualifiedContractName?: string,
  factory?: AliceNetFactory,
  implementationBase?: ContractFactory,
  constructorArgObject?: ArgData,
  initializerArgObject?: ArgData,
  salt?: string
) {
  let contractName;
  if (fullyQualifiedContractName === undefined) {
    contractName = taskArgs.contractName;
  } else {
    contractName = extractContractNameFromFullyQualifiedName(
      fullyQualifiedContractName
    );
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
    await getGasPrices(hre.ethers)
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
  implementationBase: ContractFactory,
  initializerArgs: any[]
) {
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
      : extractContractNameFromFullyQualifiedName(fullyQualifiedContractName);

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
    await getGasPrices(hre.ethers)
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployCreateData: DeployCreateData = {
    name: taskArgs.contractName,
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

export async function deployOnlyProxyTask(
  taskArgs: any,
  ethers: Ethers,
  factory?: AliceNetFactory,
  salt?: string
) {
  const waitBlocks = taskArgs.waitConfirmation;
  factory =
    factory === undefined
      ? await ethers.getContractAt("AliceNetFactory", taskArgs.factoryAddress)
      : factory;
  if (salt === undefined && taskArgs.salt !== undefined) {
    salt = ethers.utils.formatBytes32String(taskArgs.salt);
  }
  if (salt === undefined) {
    throw new Error("No salt provided");
  }
  const txResponse = await factory.deployProxy(
    salt,
    await getGasPrices(ethers)
  );
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
