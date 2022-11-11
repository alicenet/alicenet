import { task, types } from "hardhat/config";
import { DEFAULT_CONFIG_FILE_PATH } from "../lib/constants";
import {} from "./alicenetTasks";

import {
  DeploymentConfig,
  DeploymentConfigWrapper,
} from "../lib/deployment/interfaces";
import {
  deployContractsTask,
  deployCreate2Task,
  deployCreateAndRegisterTask,
  deployCreateTask,
  deployFactoryTask,
  deployOnlyProxyTask,
  deployUpgradeableProxyTask,
  upgradeProxyTask,
} from "../lib/deployment/tasks";
import {
  checkUserDirPath,
  extractFullContractInfo,
  generateDeployConfigTemplate,
  getAllContracts,
  getBytes32SaltFromContractNSTag,
  getDeployType,
  getFullyQualifiedName,
  getSortedDeployList,
  populateConstructorArgs,
  populateInitializerArgs,
  showState,
  writeDeploymentConfig,
} from "../lib/deployment/utils";

task(
  "get-network",
  "gets the current network being used from provider"
).setAction(async (_taskArgs, hre) => {
  const network: string = hre.network.name;
  await showState(network);
  return network;
});

task("get-bytes32-salt", "gets the bytes32 version of salt from contract")
  .addParam("contractName", "test contract")
  .setAction(async (taskArgs, hre) => {
    const salt = await getBytes32SaltFromContractNSTag(
      taskArgs.contractName,
      hre.artifacts,
      hre.ethers
    );
    await showState(salt);
  });

task(
  "deploy-factory",
  "Deploys an instance of a factory contract specified by its name"
)
  .addFlag("verify", "try to automatically verify contracts on etherscan")
  .addParam("legacyTokenAddress", "address of legacy token")
  .addOptionalParam(
    "waitConfirmation",
    "wait specified number of blocks between transactions",
    0,
    types.int
  )
  .addOptionalParam(
    "configFile",
    "deployment configuration json file",
    DEFAULT_CONFIG_FILE_PATH,
    types.string
  )
  .setAction(async (taskArgs, hre) => {
    // check constructorArgs length
    const legacyTokenAddress = taskArgs.legacyTokenAddress;

    return await deployFactoryTask(taskArgs, hre, legacyTokenAddress);
  });

task(
  "generate-deployment-configs",
  "default list and arg template will be generated if all optional variables are not specified"
)
  .addFlag("list", "flag to only generate deploy list")
  .addFlag("args", "flag to only generate deploy args template")
  .addOptionalParam(
    "outputFile",
    "output json file to save deployment arg template and list",
    DEFAULT_CONFIG_FILE_PATH,
    types.string
  )
  .addOptionalVariadicPositionalParam(
    "contractNames",
    "custom list of contracts to generate list and arg template for"
  )
  .setAction(async (taskArgs, hre) => {
    await checkUserDirPath(taskArgs.outputFile);
    const file = taskArgs.outputFile;

    let deploymentArgs: DeploymentConfigWrapper = {};
    let contracts: Array<string> = [];
    // no custom path and list input/ writes arg template in default scripts/base-files/deploymentArgs
    if (taskArgs.contractNames === undefined) {
      // create the default list
      // setting list name will specify default configs
      contracts = await getAllContracts(hre.artifacts);
    } // user defined path and list
    else if (taskArgs.contractNames !== undefined) {
      // create deploy list and deployment arg with the specified output path
      const nameList: Array<string> = taskArgs.contractNames;

      for (const name of nameList) {
        const sourceName = (await hre.artifacts.readArtifact(name)).sourceName;
        const fullName = sourceName + ":" + name;
        // this will cause the operation to fail if deployType is not specified on the contract
        await getDeployType(fullName, hre.artifacts);
        contracts.push(fullName);
      }
    } // user defined path, default list
    else {
      throw new Error(
        "you must specify a path to store your custom deploy config files"
      );
    }

    const deploymentList = await getSortedDeployList(
      contracts,
      hre.artifacts,
      hre.ethers
    );

    deploymentArgs = await generateDeployConfigTemplate(
      deploymentList,
      hre.artifacts,
      hre.ethers
    );

    if (taskArgs.list !== true) {
      const savedFile = await writeDeploymentConfig(deploymentArgs, file);
      console.log(`YOU MUST REPLACE THE UNDEFINED VALUES IN ${savedFile}`);
    }
  });

task(
  "deploy-contracts",
  "runs the initial deployment of all AliceNet contracts"
)
  .addFlag(
    "skipChecks",
    "skips initializer and constructor confirmation prompt"
  )
  .addFlag("verify", "try to automatically verify contracts on etherscan")
  .addOptionalParam(
    "factoryAddress",
    "specify if a factory is already deployed, if not specified a new factory will be deployed"
  )
  .addOptionalParam(
    "waitConfirmation",
    "wait specified number of blocks between transactions",
    0,
    types.int
  )
  .addOptionalParam(
    "configFile",
    "deployment configuration json file",
    DEFAULT_CONFIG_FILE_PATH,
    types.string
  )
  .setAction(async (taskArgs, hre) => {
    return await deployContractsTask(taskArgs, hre);
  });

task(
  "deploy-upgradeable-proxy",
  "deploys logic contract, proxy contract, and points the proxy to the logic contract"
)
  .addFlag(
    "skipChecks",
    "skips initializer and constructor confirmation prompt"
  )
  .addFlag("verify", "try to automatically verify contracts on etherscan")
  .addParam(
    "contractName",
    "Name of logic contract to point the proxy at",
    "string"
  )
  .addParam(
    "factoryAddress",
    "address of factory contract to deploy the contract with"
  )
  .addOptionalParam(
    "initializerArgs",
    "input initializer arguments as comma separated string values, eg: --initializerArgs 'arg1, arg2'"
  )
  .addOptionalParam(
    "waitConfirmation",
    "wait specified number of blocks between transactions",
    0,
    types.int
  )
  .addOptionalParam("outputFolder", "output folder path to save factory state")
  .addOptionalVariadicPositionalParam("constructorArgs", "constructor argfu")
  .setAction(async (taskArgs, hre) => {
    const fullyQualifiedName = await getFullyQualifiedName(
      taskArgs.contractName,
      hre.artifacts
    );

    const deploymentConfigForContract: DeploymentConfig =
      await extractFullContractInfo(
        fullyQualifiedName,
        hre.artifacts,
        hre.ethers
      );

    if (
      taskArgs.initializerArgs === undefined &&
      Object.keys(deploymentConfigForContract.initializerArgs).length > 0
    ) {
      throw new Error(
        "initializerArgs must be specified for contract: " +
          taskArgs.contractName
      );
    }

    if (
      taskArgs.constructorArgs === undefined &&
      Object.keys(deploymentConfigForContract.constructorArgs).length > 0
    ) {
      throw new Error(
        "constructorArgs must be specified for contract: " +
          taskArgs.contractName
      );
    }

    if (taskArgs.initializerArgs !== undefined) {
      const initializerArgsArray = taskArgs.initializerArgs.split(",");
      populateInitializerArgs(
        initializerArgsArray,
        deploymentConfigForContract
      );
    }

    if (taskArgs.constructorArgs !== undefined) {
      const constructorArgsArray = taskArgs.constructorArgs.split(",");
      populateConstructorArgs(
        constructorArgsArray,
        deploymentConfigForContract
      );
    }

    return await deployUpgradeableProxyTask(
      deploymentConfigForContract,
      taskArgs.waitConfirmation,
      hre,
      taskArgs.factoryAddress,
      undefined,
      taskArgs.skipChecks,
      taskArgs.verify
    );
  });

// factoryName param doesnt do anything right now
task("deploy-create2", "deploys a contract from the factory using create2")
  .addFlag(
    "standAlone",
    "flag to specify that this is not a template for a proxy"
  )
  .addFlag("verify", "try to automatically verify contracts on etherscan")
  .addParam("contractName", "logic contract name")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .addParam("salt", "salt for create2")
  .addOptionalParam(
    "waitConfirmation",
    "wait specified number of blocks between transactions",
    0,
    types.int
  )
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguments for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    return await deployCreate2Task(taskArgs, hre);
  });

// factoryName param doesnt do anything right now
task("deploy-create", "deploys a contract from the factory using create")
  .addFlag(
    "standAlone",
    "flag to specify that this is not a template for a proxy"
  )
  .addFlag("verify", "try to automatically verify contracts on etherscan")
  .addParam("contractName", "logic contract name")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .addOptionalParam(
    "waitConfirmation",
    "wait specified number of blocks between transactions",
    0,
    types.int
  )
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguments for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    return await deployCreateTask(taskArgs, hre);
  });

task(
  "deploy-create-and-register",
  "deploys a contract from the factory using create and records the address to the external contract mapping for lookup, for deploying contracts outside of deterministic address"
)
  .addFlag(
    "skipChecks",
    "skips initializer and constructor confirmation prompt"
  )
  .addFlag("verify", "try to automatically verify contracts on etherscan")
  .addParam("contractName", "logic contract name")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .addOptionalParam(
    "waitConfirmation",
    "wait specified number of blocks between transactions",
    0,
    types.int
  )

  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguments for constructor, defaults to empty array"
  )
  .setAction(async (taskArgs, hre) => {
    return await deployCreateAndRegisterTask(taskArgs, hre);
  });

task(
  "deploy-only-proxy",
  "deploys a proxy from the factory, without implementation"
)
  .addParam(
    "salt",
    "salt used to specify logicContract and proxy address calculation"
  )
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .addOptionalParam(
    "waitConfirmation",
    "wait specified number of blocks between transactions",
    0,
    types.int
  )
  .setAction(async (taskArgs, hre) => {
    return deployOnlyProxyTask(taskArgs, hre.ethers);
  });

task(
  "upgrade-proxy",
  "for upgrading existing proxy with new logic uses factory multicall function to deployCreate logic and upgrade"
)
  .addParam("contractName", "logic contract name")
  .addParam(
    "factoryAddress",
    "address of factory contract to deploy the contract with"
  )
  .addOptionalParam(
    "initializerArgs",
    "input initializer arguments as comma separated string values, eg: --initializerArgs 'arg1, arg2'"
  )
  .addOptionalParam(
    "configFile",
    "deployment configuration json file",
    DEFAULT_CONFIG_FILE_PATH,
    types.string
  )
  .addOptionalParam(
    "waitConfirmation",
    "wait specified number of blocks between transactions",
    0,
    types.int
  )
  .addOptionalVariadicPositionalParam("constructorArgs")
  .setAction(async (taskArgs, hre) => {
    return await upgradeProxyTask(taskArgs, hre);
  });
