import toml from "@iarna/toml";
import { BigNumber } from "ethers";
import fs from "fs";
import { task, types } from "hardhat/config";
import {} from "./alicenetTasks";
import {
  ALICENET_FACTORY,
  DEFAULT_CONFIG_DIR,
  DEFAULT_FACTORY_STATE_OUTPUT_DIR,
  DEPLOYMENT_ARGS_TEMPLATE_FPATH,
  DEPLOYMENT_ARG_PATH,
  DEPLOYMENT_LIST_FPATH,
} from "./constants";
import {
  generateDeployArgTemplate,
  writeDeploymentArgs,
} from "./deployment/deployArgUtil";
import {
  DeploymentList,
  getDeploymentList,
  getSortedDeployList,
  transformDeploymentList,
  writeDeploymentList,
} from "./deployment/deploymentListUtil";
import {
  checkUserDirPath,
  deployContractsMulticall,
  deployContractsTask,
  deployCreate2Task,
  deployCreateAndRegisterTask,
  deployCreateTask,
  deployFactoryTask,
  DeploymentConfigWrapper,
  deployOnlyProxyTask,
  deployUpgradeableProxyTask,
  getAllContracts,
  getBytes32SaltFromContractNSTag,
  getDeployGroup,
  getDeployGroupIndex,
  getDeployType,
  showState,
  upgradeProxyTask,
} from "./deployment/deploymentUtil";
import { FactoryData } from "./deployment/factoryStateUtil";

task(
  "get-network",
  "gets the current network being used from provider"
).setAction(async (taskArgs, hre) => {
  const network = hre.network.name;
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
  .addOptionalParam("outputFolder", "output folder path to save factoryState")
  .addOptionalParam(
    "inputFolder",
    "input folder path for deploymentArgsTemplate"
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
    "outputFolder",
    "output folder path to save deployment arg template and list",
    DEFAULT_CONFIG_DIR,
    types.string
  )
  .addOptionalVariadicPositionalParam(
    "contractNames",
    "custom list of contracts to generate list and arg template for"
  )
  .setAction(async (taskArgs, hre) => {
    await checkUserDirPath(taskArgs.outputFolder);
    const path = taskArgs.outputFolder;
    let deploymentList: DeploymentList;
    let deploymentArgs: DeploymentConfigWrapper = {};
    let list: Array<string>;
    // no custom path and list input/ writes arg template in default scripts/base-files/deploymentArgs
    if (taskArgs.contractNames === undefined) {
      // create the default list
      // setting list name will specify default configs
      const contracts = await getAllContracts(hre.artifacts);
      deploymentList = await getSortedDeployList(contracts, hre.artifacts);
      list = await transformDeploymentList(deploymentList);
      deploymentArgs = await generateDeployArgTemplate(
        list,
        hre.artifacts,
        hre.ethers
      );
    } // user defined path and list
    else if (taskArgs.contractNames !== undefined) {
      // create deploy list and deployment arg with the specified output path
      const nameList: Array<string> = taskArgs.contractNames;
      const contracts: Array<string> = [];
      for (const name of nameList) {
        const sourceName = (await hre.artifacts.readArtifact(name)).sourceName;
        const fullName = sourceName + ":" + name;
        // this will cause the operation to fail if deployType is not specified on the contract
        await getDeployType(fullName, hre.artifacts);
        contracts.push(fullName);
      }
      deploymentList = await getSortedDeployList(contracts, hre.artifacts);
      list = await transformDeploymentList(deploymentList);
      deploymentArgs = await generateDeployArgTemplate(
        list,
        hre.artifacts,
        hre.ethers
      );
    } // user defined path, default list
    else {
      throw new Error(
        "you must specify a path to store your custom deploy config files"
      );
    }
    if (taskArgs.args !== true) {
      const filteredList = [];
      for (const name of list) {
        if (name.includes("AliceNetFactory")) {
          continue;
        }
        filteredList.push(name);
      }
      await writeDeploymentList(filteredList, path);
    }
    if (taskArgs.list !== true) {
      await writeDeploymentArgs(deploymentArgs, path);
      console.log(
        `YOU MUST REPLACE THE UNDEFINED VALUES IN ${path}/deploymentArgsTemplate`
      );
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
    "inputFolder",
    "path to location containing deploymentArgsTemplate, and deploymentList",
    DEFAULT_CONFIG_DIR,
    types.string
  )
  .addOptionalParam(
    "outputFolder",
    "output folder path to save factory state",
    DEFAULT_FACTORY_STATE_OUTPUT_DIR,
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
    return await deployUpgradeableProxyTask(taskArgs, hre);
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
  .addOptionalParam(
    "outputFolder",
    "output folder path to save factory state",
    DEFAULT_FACTORY_STATE_OUTPUT_DIR,
    types.string
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
  .addOptionalParam(
    "outputFolder",
    "output folder path to save factory state",
    DEFAULT_FACTORY_STATE_OUTPUT_DIR,
    types.string
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
  .addOptionalParam(
    "outputFolder",
    "output folder path to save factory state",
    DEFAULT_FACTORY_STATE_OUTPUT_DIR,
    types.string
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
    return deployOnlyProxyTask(taskArgs, hre);
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
    "inputFolder",
    "path to location containing deploymentArgsTemplate, and deploymentList",
    DEFAULT_CONFIG_DIR,
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

// Generate a json file with all deployment information
task(
  "generate-contracts-descriptor",
  "Generates deploymentList.json file for faster contract deployment (requires deploymentList and deploymentArgsTemplate files to be already generated)"
)
  .addOptionalParam(
    "outputFolder",
    "output folder path to save deployment arg template and list"
  )
  .setAction(async (taskArgs, hre) => {
    await checkUserDirPath(taskArgs.outputFolder);
    const configDirPath =
      taskArgs.outputFolder === undefined
        ? DEFAULT_CONFIG_DIR
        : taskArgs.outputFolder;
    const path =
      configDirPath === undefined
        ? DEFAULT_CONFIG_DIR + DEPLOYMENT_LIST_FPATH + ".json"
        : configDirPath + DEPLOYMENT_LIST_FPATH + ".json";
    const deploymentArgsPath =
      configDirPath === undefined
        ? DEPLOYMENT_ARG_PATH + DEPLOYMENT_ARGS_TEMPLATE_FPATH
        : configDirPath + DEPLOYMENT_ARGS_TEMPLATE_FPATH;
    const contractsArray: any = [];
    const json = { contracts: contractsArray };
    const contracts = await getDeploymentList(taskArgs.inputFolder);
    const deploymentArgsFile = fs.readFileSync(deploymentArgsPath);
    const tomlFile: any = toml.parse(deploymentArgsFile.toLocaleString());
    for (let i = 0; i < contracts.length; i++) {
      const contract = contracts[i];
      const contractName = contract.split(":")[1];
      const tomlConstructorArgs = tomlFile.constructor[
        contract
      ] as toml.JsonArray;
      const constructorArgs: any = [];
      if (tomlConstructorArgs !== undefined) {
        tomlConstructorArgs.forEach((jsonObject) => {
          constructorArgs.push(JSON.stringify(jsonObject).split('"')[3]);
        });
      }
      const tomlInitializerArgs = tomlFile.initializer[
        contract
      ] as toml.JsonArray;
      const initializerArgs: any = [];
      if (tomlInitializerArgs !== undefined) {
        tomlInitializerArgs.forEach((jsonObject) => {
          initializerArgs.push(JSON.stringify(jsonObject).split('"')[3]);
        });
      }
      const deployType = await getDeployType(contract, hre.artifacts);
      const deployGroup = await getDeployGroup(contract, hre.artifacts);
      const deployGroupIndex = await getDeployGroupIndex(
        contract,
        hre.artifacts
      );
      if (deployType !== undefined) {
        const object = {
          name: contractName,
          fullyQualifiedName: contract,
          deployGroup:
            deployGroup !== undefined && deployGroup ? deployGroup : "general",
          deployGroupIndex:
            deployGroupIndex !== undefined && deployGroupIndex
              ? deployGroupIndex
              : "0",
          deployType,
          constructorArgs,
          initializerArgs,
        };
        json.contracts.push(object);
      }
    }
    fs.writeFileSync(path, JSON.stringify(json, null, 4));
  });

task(
  "deploy-contracts-from-descriptor",
  "Deploys ALL AliceNet contracts reading deploymentList.json"
)
  .addOptionalParam(
    "factoryAddress",
    "specify if a factory is already deployed, if not specified a new factory will be deployed"
  )
  .addOptionalParam(
    "inputFolder",
    "path to location containing deploymentArgsTemplate, and deploymentList"
  )
  .addOptionalParam("outputFolder", "output folder path to save factory state")
  .setAction(async (taskArgs, hre) => {
    let cumulativeGasUsed = BigNumber.from("0");
    await checkUserDirPath(taskArgs.outputFolder);
    const configDirPath =
      taskArgs.outputFolder === undefined
        ? DEFAULT_CONFIG_DIR
        : taskArgs.outputFolder;
    const path =
      configDirPath === undefined
        ? DEFAULT_CONFIG_DIR + DEPLOYMENT_LIST_FPATH + ".json"
        : configDirPath + DEPLOYMENT_LIST_FPATH + ".json";
    if (!fs.existsSync(path)) {
      const error =
        "Could not find " +
        DEFAULT_CONFIG_DIR +
        DEPLOYMENT_LIST_FPATH +
        ".json file. It must be generated first with generate-contracts-descriptor task";
      throw new Error(error);
    }
    const rawdata = fs.readFileSync(path);
    const json = JSON.parse(rawdata.toLocaleString());
    if (hre.network.name === "hardhat") {
      // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
      // being sent as input to the function (the contract bytecode), so we need to increase the block
      // gas limit temporally in order to deploy the template
      await hre.network.provider.send("evm_setBlockGasLimit", [
        "0x9000000000000000",
      ]);
    }
    // deploy the factory first
    let factoryAddress = taskArgs.factoryAddress;
    if (factoryAddress === undefined) {
      const factoryData: FactoryData = await hre.run("deploy-factory", {
        outputFolder: taskArgs.outputFolder,
        inputFolder: taskArgs.inputFolder,
      });
      factoryAddress = factoryData.address;
      cumulativeGasUsed = cumulativeGasUsed.add(factoryData.gas);
    }
    const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
    const factory = factoryBase.attach(factoryAddress);
    const contracts = json.contracts;
    await deployContractsMulticall(contracts, hre, factory.address);
    console.log(`total gas used: ${cumulativeGasUsed.toString()}`);
  });
