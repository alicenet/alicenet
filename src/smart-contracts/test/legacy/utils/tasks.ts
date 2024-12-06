import { ContractFactory } from "ethers";
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
} from "./alicenetFactory";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  CREATE_AND_REGISTER_DEPLOYMENT,
  EVENT_DEPLOYED_PROXY,
  EVENT_DEPLOYED_RAW,
  ONLY_PROXY,
  UPGRADEABLE_DEPLOYMENT,
} from "./constants";
import {
  DeployCreateData,
  DeploymentConfig,
  DeploymentConfigWrapper,
  FactoryData,
  ProxyData,
} from "./interfaces";
import {
  encodeInitCallData,
  promptCheckDeploymentArgs,
  readDeploymentConfig,
  showState,
  verifyContract,
} from "./utils";
import * as helpers from "@nomicfoundation/hardhat-toolbox/network-helpers";

type Ethers = typeof ethers;

// AliceNet Factory Task Functions

export async function deployFactoryTask(
  legacyTokenAddress: string,
  hre: HardhatRuntimeEnvironment,
  waitBlocks: number = 0,
  verify: boolean = false,
  skipChecks: boolean = false
) {
  const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
  if (!hre.ethers.isAddress(legacyTokenAddress)) {
    throw new Error("legacyTokenAddress is not an address");
  }
  const signers = await hre.ethers.getSigners();
  const deployTX = await factoryBase.getDeployTransaction(legacyTokenAddress);
  if (!skipChecks) {
    const promptMessage = `Do you want to deploy the factory and ALCA with constructor argument: { legacyTokenAddress: ${legacyTokenAddress} } ? (y/n)\n`;
    await promptCheckDeploymentArgs(promptMessage);
  }
  const gasCost = await hre.ethers.provider.estimateGas(deployTX);
  // deploys the factory
  const factory = await deployFactory(
    legacyTokenAddress,
    hre.ethers,
    factoryBase
  );
  // wait for deployment confirmation
  await factory.deploymentTransaction()?.wait(waitBlocks);
  // record the state in a json file to be used in other tasks
  const factoryData: FactoryData = {
    address: await factory.getAddress(),
    owner: signers[0].address,
    gas: gasCost,
  };
  const alcaAddress = await factory.lookup(
    hre.ethers.encodeBytes32String("ALCA")
  );
  if (verify) {
    await verifyContract(hre, await factory.getAddress(), [legacyTokenAddress]);
    await verifyContract(hre, alcaAddress, [legacyTokenAddress]);
  }

  await showState(
    `Deployed ${ALICENET_FACTORY} at address: ${await factory.getAddress()}, gasCost: ${gasCost}`
  );
  await showState(
    `Deployed ALCA at address: ${await factory.lookup(
      hre.ethers.encodeBytes32String("ALCA")
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
  let cumulativeGasUsed = 0n;

  const deploymentConfig: DeploymentConfigWrapper =
    readDeploymentConfig(configFile);

  if (factoryAddress === undefined) {
    const expectedContractFullQualifiedName =
      "contracts/legacy/AliceNetFactory.sol:AliceNetFactory";
    const expectedField = "legacyToken_";

    if (
      deploymentConfig[expectedContractFullQualifiedName] === undefined ||
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
    const factoryData: FactoryData = await deployFactoryTask(
      legacyTokenAddress,
      hre,
      waitBlocks,
      verify,
      skipChecks
    );
    factoryAddress = factoryData.address;
    cumulativeGasUsed = cumulativeGasUsed + factoryData.gas;
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
    if (salt !== undefined && salt !== "") {
      const existingAddress = await factory.lookup(salt);
      if (existingAddress !== hre.ethers.ZeroAddress) {
        console.log(
          `Skipping deployment for contract ${fullyQualifiedContractName} since it already exists at address ${existingAddress}`
        );
        continue;
      }
    }

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
        cumulativeGasUsed = cumulativeGasUsed + proxyData.gas;
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
        cumulativeGasUsed = cumulativeGasUsed + proxyData.gas;
        break;
      }
      case CREATE_AND_REGISTER_DEPLOYMENT: {
        const deployCreateData = await deployCreateAndRegisterTask(
          deploymentConfigForContract,
          hre,
          factory,
          undefined,
          waitBlocks,
          skipChecks,
          verify
        );
        cumulativeGasUsed = cumulativeGasUsed + deployCreateData.gas;
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
  let initCallData = "0x";
  if (initializerArgs.length > 0) {
    initCallData = encodeInitCallData(implementationBase, initializerArgs);
  }

  if (hre.network.name === "hardhat") {
    // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
    // being sent as input to the function (the contract bytecode), so we need to increase the block
    // gas limit temporally in order to deploy the template
    await helpers.setBlockGasLimit(3000000000000000);
  }

  if (!skipChecks) {
    const constructorDetails = JSON.stringify(
      deploymentConfigForContract.constructorArgs
    );
    const initializerDetails = JSON.stringify(
      deploymentConfigForContract.initializerArgs
    );

    const promptMessage = `Do you want to deploy ${deploymentConfigForContract.name} with constructor arguments: ${constructorDetails} initializer Args: ${initializerDetails}? (y/n)\n`;
    await promptCheckDeploymentArgs(promptMessage);
  }

  const txResponse = await deployUpgradeableGasSafe(
    deploymentConfigForContract.name,
    factory,
    hre.ethers,
    initCallData,
    constructorArgs,
    deploymentConfigForContract.salt,
    waitBlocks
  );
  const receipt = await txResponse.wait(waitBlocks);
  if (receipt === null) {
    throw new Error("receipt is null");
  }
  const deployedLogicAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  if (verify) {
    await verifyContract(hre, deployedLogicAddress, constructorArgs);
  }
  const proxyData: ProxyData = {
    factoryAddress: await factory.getAddress(),
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
    factoryAddress = await factory.getAddress();
  }
  const txResponse = await deployCreate(
    deploymentConfigForContract.name,
    factory,
    hre.ethers,
    constructorArgs
  );
  const receipt = await txResponse.wait(waitBlocks);
  if (receipt === null) {
    throw new Error("receipt is null");
  }
  const contractAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  const deployCreateData: DeployCreateData = {
    name: deploymentConfigForContract.name,
    address: contractAddress,
    factoryAddress,
    gas: receipt.gasUsed,
    constructorArgs: deploymentConfigForContract.constructorArgs,
  };
  if (verify) {
    await verifyContract(hre, contractAddress, constructorArgs);
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
    factoryAddress = await factory.getAddress();
  }
  const txResponse = await deployCreate2(
    deploymentConfigForContract.name,
    factory,
    hre.ethers,
    constructorArgs,
    deploymentConfigForContract.salt
  );
  const receipt = await txResponse.wait(waitBlocks);
  if (receipt === null) {
    throw new Error("receipt is null");
  }
  const contractAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  const deployCreate2Data: any = {
    name: deploymentConfigForContract.name,
    address: contractAddress,
    factoryAddress,
    gas: receipt.gasUsed,
    constructorArgs: deploymentConfigForContract.constructorArgs,
  };
  if (verify) {
    await verifyContract(hre, contractAddress, constructorArgs);
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
  implementationBase?: ContractFactory,
  skipInitializer?: boolean,
  test?: boolean
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
    factoryAddress = await factory.getAddress();
  }

  if (test) {
    factory = await impersonateFactoryOwner(hre, undefined, undefined, factory);
  }

  implementationBase =
    implementationBase === undefined
      ? ((await hre.ethers.getContractFactory(
          deploymentConfigForContract.name
        )) as ContractFactory)
      : implementationBase;
  let initCallData = "0x";
  if (!skipInitializer) {
    initCallData = encodeInitCallData(implementationBase, initializerArgs);
  }

  const txResponse = await upgradeProxyGasSafe(
    deploymentConfigForContract.name,
    factory,
    hre.ethers,
    initCallData,
    constructorArgs,
    deploymentConfigForContract.salt,
    waitBlocks
  );

  const receipt = await txResponse.wait(waitBlocks);
  if (receipt === null) {
    throw new Error("receipt is null");
  }
  const implementationAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  const proxyAddress = await factory.lookup(deploymentConfigForContract.salt);
  await showState(
    `Updating logic for the ${deploymentConfigForContract.name} proxy at ${proxyAddress} to point to implementation at ${implementationAddress}, gasCost: ${receipt.gasUsed}`
  );
  const proxyData: ProxyData = {
    factoryAddress,
    logicName: deploymentConfigForContract.name,
    logicAddress: implementationAddress,
    salt: deploymentConfigForContract.salt,
    proxyAddress,
    gas: receipt.gasUsed,
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
  verify: boolean = false
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
    factoryAddress = await factory.getAddress();
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
    const promptMessage = `Do you want to deploy ${deploymentConfigForContract.name} with constructorArgs: ${constructorDetails}, salt: ${deploymentConfigForContract.salt}? (y/n)\n`;
    await promptCheckDeploymentArgs(promptMessage);
  }
  const txResponse = await deployCreateAndRegister(
    deploymentConfigForContract.name,
    factory,
    hre.ethers,
    constructorArgs,
    deploymentConfigForContract.salt
  );
  const receipt = await txResponse.wait(waitBlocks);
  if (receipt === null) {
    throw new Error("receipt is null");
  }
  const contractAddress = getEventVar(
    receipt,
    EVENT_DEPLOYED_RAW,
    CONTRACT_ADDR
  );
  const deployCreateData: DeployCreateData = {
    name: deploymentConfigForContract.name,
    address: contractAddress,
    factoryAddress,
    gas: receipt.gasUsed,
    constructorArgs: deploymentConfigForContract.constructorArgs,
  };

  if (verify) {
    await verifyContract(hre, contractAddress, constructorArgs);
  }

  await showState(
    `Deployed ${deployCreateData.name} at ${deployCreateData.address}, gasCost: ${deployCreateData.gas}`
  );

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

  const txResponse = await factory.deployProxy(salt);
  const receipt = await txResponse.wait(waitBlocks);
  if (receipt === null) {
    throw new Error("receipt is null");
  }
  const proxyAddr = getEventVar(receipt, EVENT_DEPLOYED_PROXY, CONTRACT_ADDR);
  const proxyData: ProxyData = {
    proxyAddress: proxyAddr,
    salt,
    factoryAddress: await factory.getAddress(),
    gas: receipt.gasUsed,
    receipt,
  };

  await showState(
    `Deployed ${salt} proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`
  );
  return proxyData;
}

export async function impersonateFactoryOwner(
  hre: HardhatRuntimeEnvironment,
  address: string = "0xff55549a3ceea32fba4794bf1a649a2363fcda53",
  factoryAddress: string = "0x4b6df6b299fb6414f45719e0d9e1889269a7843e",
  factory?: AliceNetFactory
) {
  if (factory === undefined) {
    factory = await hre.ethers.getContractAt("AliceNetFactory", factoryAddress);
  }
  await hre.network.provider.request({
    method: "hardhat_impersonateAccount",
    params: [address],
  });
  const signers = await hre.ethers.getSigners();
  const tx = await signers[0].populateTransaction({
    to: address,
    value: hre.ethers.parseEther("2"),
  });
  await signers[0].sendTransaction(tx);
  await helpers.impersonateAccount(address);
  const signer = await hre.ethers.getSigner(address);
  return factory.connect(signer);
}
