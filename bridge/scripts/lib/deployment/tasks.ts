import { BigNumber, ContractFactory } from "ethers";
import { ethers } from "hardhat";
import { HardhatRuntimeEnvironment } from "hardhat/types";
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
  DeployCreateData,
  DeploymentConfig,
  DeploymentConfigWrapper,
  FactoryData,
  ProxyData,
} from "./interfaces";
import {
  encodeInitCallData,
  getGasPrices,
  promptCheckDeploymentArgs,
  readDeploymentConfig,
  showState,
  verifyContract,
} from "./utils";

type Ethers = typeof ethers;

// AliceNet Factory Task Functions

export async function deployFactoryTask(
  legacyTokenAddress: string,
  hre: HardhatRuntimeEnvironment,
  waitBlocks: number = 0,
  verify: boolean = false
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
  await factory.deployTransaction.wait(waitBlocks);
  // record the state in a json file to be used in other tasks
  const factoryData: FactoryData = {
    address: factory.address,
    owner: signers[0].address,
    gas: gasCost,
  };
  if (verify) {
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
  configFile: string,
  hre: HardhatRuntimeEnvironment,
  factoryAddress?: string,
  waitBlocks: number = 0,
  skipChecks: boolean = false,
  verify: boolean = false
) {
  let cumulativeGasUsed = BigNumber.from("0");

  const deploymentConfig: DeploymentConfigWrapper =
    readDeploymentConfig(configFile);

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
        ` ${expectedContractFullQualifiedName} inside ${configFile}`
    );
  }

  const legacyTokenAddress: string = deploymentConfig[
    expectedContractFullQualifiedName
  ].constructorArgs[expectedField] as string;

  if (factoryAddress === undefined) {
    const factoryData: FactoryData = await deployFactoryTask(
      legacyTokenAddress,
      hre,
      waitBlocks,
      verify
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

    switch (deployType) {
      case UPGRADEABLE_DEPLOYMENT: {
        const proxyData = await deployUpgradeableProxyTask(
          deploymentConfigForContract,
          hre,
          waitBlocks,
          factory,
          undefined,
          skipChecks,
          verify
        );
        cumulativeGasUsed = cumulativeGasUsed.add(proxyData.gas);
        break;
      }
      case ONLY_PROXY: {
        const proxyData = await deployOnlyProxyTask(
          salt,
          hre.ethers,
          factory,
          undefined,
          waitBlocks
        );
        cumulativeGasUsed = cumulativeGasUsed.add(proxyData.gas);
        break;
      }
      case FUNCTION_DEPLOY_CREATE: {
        const deployCreateData = await deployCreateAndRegisterTask(
          deploymentConfigForContract,
          hre,
          factory,
          undefined,
          waitBlocks,
          skipChecks,
          verify
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
  hre: HardhatRuntimeEnvironment,
  waitBlocks: number = 0,
  factory?: AliceNetFactory,
  factoryAddress?: string,
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

  const initCallData: string = encodeInitCallData(
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
  deploymentConfigForContract: DeploymentConfig,
  hre: HardhatRuntimeEnvironment,
  factory?: AliceNetFactory,
  factoryAddress?: string,
  waitBlocks: number = 0,
  verify: boolean = false,
  standAlone: boolean = false
) {
  const constructorArgs = Object.values(
    deploymentConfigForContract.constructorArgs
  );

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
  } else {
    factoryAddress = factory.address;
  }
  const txResponse = await deployCreate(
    deploymentConfigForContract.name,
    factory,
    hre.ethers,
    constructorArgs
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployCreateData: DeployCreateData = {
    name: deploymentConfigForContract.name,
    address: getEventVar(receipt, EVENT_DEPLOYED_RAW, CONTRACT_ADDR),
    factoryAddress: factoryAddress,
    gas: receipt.gasUsed,
    constructorArgs: deploymentConfigForContract.constructorArgs,
  };
  if (verify) {
    await verifyContract(hre, factory.address, constructorArgs);
  }
  if (!standAlone) {
    await showState(
      `[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${deploymentConfigForContract.name} contract at: ${deployCreateData.address}, gas: ${receipt.gasUsed}`
    );
  } else {
    await showState(
      `Deployed ${deployCreateData.name} at ${deployCreateData.address}, gasCost: ${deployCreateData.gas}`
    );
  }
  deployCreateData.receipt = receipt;
  return deployCreateData;
}

export async function deployCreate2Task(
  deploymentConfigForContract: DeploymentConfig,
  hre: HardhatRuntimeEnvironment,
  waitBlocks: number = 0,
  factory?: AliceNetFactory,
  factoryAddress?: string,
  verify: boolean = false,
  standAlone: boolean = false
) {
  const constructorArgs = Object.values(
    deploymentConfigForContract.constructorArgs
  );

  if (factory === undefined) {
    if (factoryAddress === undefined) {
      throw new Error("Either factory or factoryAddress must be provided");
    } else {
      factory = await hre.ethers.getContractAt(
        "AliceNetFactory",
        factoryAddress
      );
    }
  } else {
    factoryAddress = factory.address;
  }
  const txResponse = await deployCreate2(
    deploymentConfigForContract.name,
    factory,
    hre.ethers,
    constructorArgs,
    deploymentConfigForContract.salt
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployCreate2Data: any = {
    name: deploymentConfigForContract.name,
    address: getEventVar(receipt, EVENT_DEPLOYED_RAW, CONTRACT_ADDR),
    factoryAddress: factoryAddress,
    gas: receipt.gasUsed,
    constructorArgs: deploymentConfigForContract.constructorArgs,
  };
  if (verify) {
    await verifyContract(hre, factory.address, constructorArgs);
  }
  if (!standAlone) {
    await showState(
      `[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${deploymentConfigForContract.name} contract at: ${deployCreate2Data.address}, gas: ${receipt.gasUsed}`
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
  deploymentConfigForContract: DeploymentConfig,
  hre: HardhatRuntimeEnvironment,
  waitBlocks: number = 0,
  factory?: AliceNetFactory,
  factoryAddress?: string,
  implementationBase?: ContractFactory
) {
  const constructorArgs = Object.values(
    deploymentConfigForContract.constructorArgs
  );
  const initializerArgs = Object.values(
    deploymentConfigForContract.initializerArgs
  );

  if (factory === undefined) {
    if (factoryAddress === undefined) {
      throw new Error("Either factory or factoryAddress must be provided");
    } else {
      factory = await hre.ethers.getContractAt(
        "AliceNetFactory",
        factoryAddress
      );
    }
  } else {
    factoryAddress = factory.address;
  }

  implementationBase =
    implementationBase === undefined
      ? ((await hre.ethers.getContractFactory(
          deploymentConfigForContract.name
        )) as ContractFactory)
      : implementationBase;

  const initCallData: string = encodeInitCallData(
    implementationBase,
    initializerArgs
  );
  const txResponse = await upgradeProxyGasSafe(
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
  const implementationAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  console.log("events", receipt.events);
  const proxyAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_PROXY,
    CONTRACT_ADDR
  );
  await showState(
    `Updating logic for the ${deploymentConfigForContract.name} proxy at ${proxyAddress} to point to implementation at ${implementationAddress}, gasCost: ${receipt.gasUsed}`
  );
  const proxyData: ProxyData = {
    factoryAddress: factoryAddress,
    logicName: deploymentConfigForContract.name,
    logicAddress: implementationAddress,
    salt: deploymentConfigForContract.salt,
    proxyAddress,
    gas: receipt.gasUsed.toNumber(),
    receipt,
    initCallData,
  };
  return proxyData;
}

export async function deployCreateAndRegisterTask(
  deploymentConfigForContract: DeploymentConfig,
  hre: HardhatRuntimeEnvironment,
  factory?: AliceNetFactory,
  factoryAddress?: string,
  waitBlocks: number = 0,
  skipChecks: boolean = false,
  verify: boolean = false,
  standAlone: boolean = false
) {
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
  } else {
    factoryAddress = factory.address;
  }

  const constructorArgs = Object.values(
    deploymentConfigForContract.constructorArgs
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
    const promptMessage = `Do you want to deploy ${deploymentConfigForContract.name} with constructorArgs: ${constructorDetails}, salt: ${deploymentConfigForContract.salt}? (y/n)`;
    await promptCheckDeploymentArgs(promptMessage);
  }
  const txResponse = await deployCreateAndRegister(
    deploymentConfigForContract.name,
    factory,
    hre.ethers,
    constructorArgs,
    deploymentConfigForContract.salt,
    await getGasPrices(hre.ethers)
  );
  const receipt = await txResponse.wait(waitBlocks);
  const deployCreateData: DeployCreateData = {
    name: deploymentConfigForContract.name,
    address: getEventVar(receipt, EVENT_DEPLOYED_RAW, CONTRACT_ADDR),
    factoryAddress,
    gas: receipt.gasUsed,
    constructorArgs: deploymentConfigForContract.constructorArgs,
  };

  if (verify) {
    await verifyContract(hre, factory.address, constructorArgs);
  }

  if (!standAlone) {
    await showState(
      `[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${deploymentConfigForContract.name} contract at: ${deployCreateData.address}, gas: ${receipt.gasUsed}`
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
  salt: string,
  ethers: Ethers,
  factory?: AliceNetFactory,
  factoryAddress?: string,
  waitBlocks: number = 0
) {
  // if an instance of the factory contract is not provided get it from ethers
  if (factory === undefined) {
    if (factoryAddress === undefined) {
      throw new Error("Either factory or factoryAddress must be provided");
    } else {
      factory = await ethers.getContractAt("AliceNetFactory", factoryAddress);
    }
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
