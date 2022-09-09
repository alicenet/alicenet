import toml from "@iarna/toml";
import {
  BigNumber,
  BytesLike,
  ContractReceipt,
  ContractTransaction,
} from "ethers";
import fs from "fs";
import { task } from "hardhat/config";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { encodeMultiCallArgs } from "./alicenetTasks";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  DEFAULT_CONFIG_OUTPUT_DIR,
  DEPLOYED_PROXY,
  DEPLOYED_RAW,
  DEPLOYMENT_ARGS_TEMPLATE_FPATH,
  DEPLOYMENT_ARG_PATH,
  DEPLOYMENT_LIST_FPATH,
  DEPLOY_CREATE,
  DEPLOY_PROXY,
  DEPLOY_UPGRADEABLE_PROXY,
  INITIALIZER,
  MULTICALL_GAS_LIMIT,
  MULTI_CALL_DEPLOY_PROXY,
  ONLY_PROXY,
  PROXY,
  STATIC_DEPLOYMENT,
  UPGRADEABLE_DEPLOYMENT,
  UPGRADE_DEPLOYED_PROXY,
  UPGRADE_PROXY,
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
  DeployArgs,
  deployContractsMulticall,
  DeploymentArgs,
  DeployProxyMCArgs,
  extractName,
  getAllContracts,
  getContractDescriptor,
  getDeployGroup,
  getDeployGroupIndex,
  getDeployMetaArgs,
  getDeployType,
  getDeployUpgradeableMultiCallArgs,
  getDeployUpgradeableProxyArgs,
  getFactoryDeploymentArgs,
  isInitializable,
} from "./deployment/deploymentUtil";
import {
  DeployCreateData,
  FactoryData,
  MetaContractData,
  ProxyData,
  updateDefaultFactoryData,
  updateDeployCreateList,
  updateProxyList,
} from "./deployment/factoryStateUtil";

task(
  "getNetwork",
  "gets the current network being used from provider"
).setAction(async (taskArgs, hre) => {
  const network = hre.network.name;
  await showState(network);
  return network;
});

task("getBytes32Salt", "gets the bytes32 version of salt from contract")
  .addParam("contractName", "test contract")
  .setAction(async (taskArgs, hre) => {
    const salt = await getBytes32Salt(taskArgs.contractName, hre);
    await showState(salt);
  });

task(
  "deployFactory",
  "Deploys an instance of a factory contract specified by its name"
)
  .addFlag("waitConfirmation", "wait 8 blocks between transactions")
  .addOptionalParam("outputFolder", "output folder path to save factoryState")
  .setAction(async (taskArgs, hre) => {
    const waitBlocks = taskArgs.waitConfirmation === true ? 8 : undefined;
    await checkUserDirPath(taskArgs.outputFolder);
    const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
    let constructorArgs = await getFactoryDeploymentArgs(hre.artifacts);
    constructorArgs = constructorArgs === undefined ? [] : constructorArgs;
    const accounts = await getAccounts(hre);
    // calculate the factory address for the constructor arg
    const deployTX = factoryBase.getDeployTransaction(constructorArgs[0]);
    const gasCost = await hre.ethers.provider.estimateGas(deployTX);
    // deploys the factory
    const factory = await factoryBase.deploy(
      constructorArgs[0],
      await getGasPrices(hre)
    );
    await factory.deployTransaction.wait(waitBlocks);
    // record the state in a json file to be used in other tasks
    const factoryData: FactoryData = {
      address: factory.address,
      owner: accounts[0],
      gas: gasCost,
    };
    const network = hre.network.name;
    await updateDefaultFactoryData(network, factoryData, taskArgs.outputFolder);
    await showState(
      `Deployed: ${ALICENET_FACTORY}, at address: ${factory.address}, with ${gasCost} gas`
    );
    return factoryData;
  });

task(
  "generateDeploymentConfigs",
  "default list and arg template will be generated if all optional variables are not specified"
)
  .addFlag("list", "flag to only generate deploy list")
  .addFlag("args", "flag to only generate deploy args template")
  .addOptionalParam(
    "outputFolder",
    "output folder path to save deployment arg template and list"
  )
  .addOptionalVariadicPositionalParam(
    "contractNames",
    "custom list of contracts to generate list and arg template for"
  )
  .setAction(async (taskArgs, hre) => {
    await checkUserDirPath(taskArgs.outputFolder);
    const path =
      taskArgs.outputFolder === undefined
        ? DEFAULT_CONFIG_OUTPUT_DIR
        : taskArgs.outputFolder;
    let deploymentList: DeploymentList;
    let deploymentArgs: DeploymentArgs = {
      constructor: {},
      initializer: {},
    };
    let list: Array<string>;
    // no custom path and list input/ writes arg template in default scripts/base-files/deploymentArgs
    if (
      taskArgs.outputFolder === undefined &&
      taskArgs.contractNames === undefined
    ) {
      // create the default list
      // setting list name will specify default configs
      const contracts = await getAllContracts(hre.artifacts);
      console.log(contracts);
      deploymentList = await getSortedDeployList(contracts, hre.artifacts);
      console.log(deploymentList);
      list = await transformDeploymentList(deploymentList);
      deploymentArgs = await generateDeployArgTemplate(list, hre.artifacts);
    } // user defined path and list
    else if (
      taskArgs.outputFolder !== undefined &&
      taskArgs.contractNames !== undefined
    ) {
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
      deploymentArgs = await generateDeployArgTemplate(list, hre.artifacts);
    } // user defined path, default list
    else if (
      taskArgs.outputFolder !== undefined &&
      taskArgs.contractNames === undefined
    ) {
      const contracts = await getAllContracts(hre.artifacts);
      deploymentList = await getSortedDeployList(contracts, hre.artifacts);
      list = await transformDeploymentList(deploymentList);
      deploymentArgs = await generateDeployArgTemplate(list, hre.artifacts);
    } else {
      throw new Error(
        "you must specify a path to store your custom deploy config files"
      );
    }
    if (taskArgs.args !== true) {
      let filteredList = [];
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

task("deployContracts", "runs the initial deployment of all AliceNet contracts")
  .addFlag("waitConfirmation", "wait 8 blocks between transactions")
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
    // setting listName undefined will use the default list
    const artifacts = hre.artifacts;
    // deploy the factory first
    let factoryAddress = taskArgs.factoryAddress;
    if (factoryAddress === undefined) {
      const factoryData: FactoryData = await hre.run("deployFactory", {
        outputFolder: taskArgs.outputFolder,
      });
      factoryAddress = factoryData.address;
      cumulativeGasUsed = cumulativeGasUsed.add(factoryData.gas);
    }
    let deployArgs: DeployArgs;
    // get an array of all contracts in the artifacts
    const contracts = await getDeploymentList(taskArgs.inputFolder);
    let metaContractData: MetaContractData;
    let proxyData: ProxyData;
    // let contracts = ["src/tokens/periphery/validatorPool/Snapshots.sol:Snapshots"]
    for (let i = 0; i < contracts.length; i++) {
      const fullyQualifiedName = contracts[i];
      // check the contract for the @custom:deploy-type tag
      const deployType = await getDeployType(fullyQualifiedName, artifacts);
      switch (deployType) {
        case STATIC_DEPLOYMENT: {
          deployArgs = await getDeployMetaArgs(
            fullyQualifiedName,
            taskArgs.waitConfirmation,
            factoryAddress,
            artifacts,
            taskArgs.inputFolder,
            taskArgs.outputFolder
          );
          metaContractData = await hre.run(
            "multiCallDeployMetamorphic",
            deployArgs
          );
          cumulativeGasUsed = cumulativeGasUsed.add(metaContractData.gas);
          break;
        }
        case UPGRADEABLE_DEPLOYMENT: {
          deployArgs = await getDeployUpgradeableProxyArgs(
            fullyQualifiedName,
            factoryAddress,
            artifacts,
            taskArgs.waitConfirmation,
            taskArgs.inputFolder,
            taskArgs.outputFolder
          );
          proxyData = await hre.run("fullMultiCallDeployProxy", deployArgs);
          cumulativeGasUsed = cumulativeGasUsed.add(proxyData.gas);
          break;
        }
        case ONLY_PROXY: {
          const name = extractName(fullyQualifiedName);
          const salt: BytesLike = await getBytes32Salt(name, hre);
          proxyData = await hre.run("deployProxy", {
            factoryAddress,
            salt,
            waitConfirmation: taskArgs.waitConfirmation,
          });
          cumulativeGasUsed = cumulativeGasUsed.add(proxyData.gas);
          break;
        }
        default: {
          break;
        }
      }
    }
    console.log(`total gas used: ${cumulativeGasUsed.toString()}`);
  });

task(
  "fullMultiCallDeployProxy",
  "Multicalls deployCreate, deployProxy, and upgradeProxy, if gas cost exceeds 10 million deployUpgradeableProxy will be used"
)
  .addFlag("waitConfirmation", "wait 8 blocks between transactions")
  .addParam("contractName", "Name of logic contract to point the proxy at")
  .addParam(
    "factoryAddress",
    "address of factory contract to deploy the contract with"
  )
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam("outputFolder", "output folder path to save factory state")
  .addOptionalVariadicPositionalParam("constructorArgs", "")
  .setAction(async (taskArgs, hre) => {
    const waitBlocks = taskArgs.waitConfirmation === true ? 8 : undefined;
    const network = hre.network.name;
    const callArgs: DeployArgs = {
      contractName: taskArgs.contractName,
      factoryAddress: taskArgs.factoryAddress,
      initCallData: taskArgs.initCallData,
      outputFolder: taskArgs.outputFolder,
      constructorArgs: taskArgs.constructorArgs,
    };
    const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
    const factory = factoryBase.attach(taskArgs.factoryAddress);
    const logicFactory: any = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );

    const initArgs =
      taskArgs.initCallData === undefined
        ? []
        : taskArgs.initCallData.replace(/\s+/g, "").split(",");
    const fullname = (await getFullyQualifiedName(
      taskArgs.contractName,
      hre
    )) as string;
    const isInitable = await isInitializable(fullname, hre.artifacts);
    const initCallData = isInitable
      ? logicFactory.interface.encodeFunctionData(INITIALIZER, initArgs)
      : "0x";
    // factory interface pointed to deployed factory contract
    // get the 32byte salt from logic contract file
    const salt: BytesLike = await getBytes32Salt(taskArgs.contractName, hre);
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
    // get the multi call arguements as [deployProxy, upgradeProxy]
    const contractDescriptor = await getContractDescriptor(
      taskArgs.contractName,
      constructorArgs,
      initArgs,
      hre
    );
    const multiCallArgs = await getDeployUpgradeableMultiCallArgs(
      contractDescriptor,
      hre,
      factory.address,
      initCallData
    );
    const estimatedMultiCallGas = await factory.estimateGas.multiCall(
      multiCallArgs
    );
    let txResponse: ContractTransaction;
    let receipt: ContractReceipt;

    if (estimatedMultiCallGas.lt(BigNumber.from(MULTICALL_GAS_LIMIT))) {
      // send the multicall transaction with deployProxy and upgradeProxy
      txResponse = await factory.multiCall(
        multiCallArgs,
        await getGasPrices(hre)
      );
      receipt = await txResponse.wait(waitBlocks);
      const proxyData: ProxyData = {
        factoryAddress: taskArgs.factoryAddress,
        logicName: taskArgs.contractName,
        logicAddress: taskArgs.logicAddress,
        salt,
        proxyAddress: getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR),
        gas: receipt.gasUsed,
        receipt,
        initCallData,
      };
      await showState(
        `Deployed ${proxyData.logicName} with proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`
      );
      await updateProxyList(network, proxyData, taskArgs.outputFolder);
      return proxyData;
    } else {
      return await hre.run(DEPLOY_UPGRADEABLE_PROXY, callArgs);
    }
  });

task(
  DEPLOY_UPGRADEABLE_PROXY,
  "deploys logic contract, proxy contract, and points the proxy to the logic contract"
)
  .addFlag("waitConfirmation", "wait 8 blocks between transactions")
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
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam("outputFolder", "output folder path to save factory state")
  .addOptionalVariadicPositionalParam("constructorArgs", "")
  .setAction(async (taskArgs, hre) => {
    let cumulativeGas = BigNumber.from("0");
    // uses the factory Data and logic contractName and returns deploybytecode and any constructor args attached
    const callArgs: DeployArgs = {
      contractName: taskArgs.contractName,
      factoryAddress: taskArgs.factoryAddress,
      constructorArgs: taskArgs.constructorArgs,
      outputFolder: taskArgs.outputFolder,
    };
    // deploy create the logic contract
    const deployCreateData: DeployCreateData = await hre.run(
      DEPLOY_CREATE,
      callArgs
    );
    cumulativeGas = cumulativeGas.add(deployCreateData.gas);
    const mcCallArgs: DeployProxyMCArgs = {
      contractName: taskArgs.contractName,
      factoryAddress: taskArgs.factoryAddress,
      logicAddress: deployCreateData.address,
      waitConfirmation: taskArgs.waitConfirmation,
      initCallData: taskArgs.initCallData,
      outputFolder: taskArgs.outputFolder,
    };
    const proxyData: ProxyData = await hre.run(
      MULTI_CALL_DEPLOY_PROXY,
      mcCallArgs
    );
    cumulativeGas = cumulativeGas.add(proxyData.gas);
    proxyData.gas = cumulativeGas;
    await showState(
      `Deployed ${proxyData.logicName} with proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`
    );
    return proxyData;
  });

// factoryName param doesnt do anything right now
task(DEPLOY_CREATE, "deploys a contract from the factory using create")
  .addFlag("waitConfirmation", "wait 8 blocks between transactions")
  .addParam("contractName", "logic contract name")
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .addOptionalParam("outputFolder", "output folder path to save factory state")
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguments for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    const waitBlocks = taskArgs.waitConfirmation === true ? 8 : undefined;
    const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
    // get a factory instance connected to the factory a
    const factory = factoryBase.attach(taskArgs.factoryAddress);
    const logicContract: any = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    const constructorArgs =
      taskArgs.constructorArgs === undefined ? [] : taskArgs.constructorArgs;
    // encode deployBcode
    const deployTx = logicContract.getDeployTransaction(...constructorArgs);
    if (hre.network.name === "hardhat") {
      // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
      // being sent as input to the function (the contract bytecode), so we need to increase the block
      // gas limit temporally in order to deploy the template
      await hre.network.provider.send("evm_setBlockGasLimit", [
        "0x3000000000000000",
      ]);
    }
    if (deployTx.data !== undefined) {
      const txResponse = await factory.deployCreate(
        deployTx.data,
        await getGasPrices(hre)
      );
      const receipt = await txResponse.wait(waitBlocks);
      const deployCreateData: DeployCreateData = {
        name: taskArgs.contractName,
        address: getEventVar(receipt, DEPLOYED_RAW, CONTRACT_ADDR),
        factoryAddress: taskArgs.factoryAddress,
        gas: receipt.gasUsed,
        constructorArgs: taskArgs?.constructorArgs,
      };
      const network = hre.network.name;
      await updateDeployCreateList(
        network,
        deployCreateData,
        taskArgs.outputFolder
      );
      await showState(
        `[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${taskArgs.contractName} contract at: ${deployCreateData.address}, gas: ${receipt.gasUsed}`
      );
      deployCreateData.receipt = receipt;
      return deployCreateData;
    } else {
      throw new Error(
        `failed to get deployment bytecode for ${taskArgs.contractName}`
      );
    }
  });

task(DEPLOY_PROXY, "deploys a proxy from the factory")
  .addFlag("waitConfirmation", "wait 8 blocks between transactions")
  .addParam(
    "salt",
    "salt used to specify logicContract and proxy address calculation"
  )
  .addParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .setAction(async (taskArgs, hre) => {
    const waitBlocks = taskArgs.waitConfirmation === true ? 8 : undefined;
    const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
    const factory = factoryBase.attach(taskArgs.factoryAddress);
    const txResponse = await factory.deployProxy(
      taskArgs.salt,
      await getGasPrices(hre)
    );
    const receipt = await txResponse.wait(waitBlocks);
    const proxyAddr = getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR);
    const proxyData: ProxyData = {
      proxyAddress: proxyAddr,
      salt: taskArgs.salt,
      factoryAddress: taskArgs.factoryAddress,
      gas: receipt.gasUsed,
      receipt,
    };
    const salt = hre.ethers.utils.parseBytes32String(taskArgs.salt);
    await showState(
      `Deployed ${salt} proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`
    );
    return proxyData;
  });

task(UPGRADE_DEPLOYED_PROXY, "deploys a contract from the factory using create")
  .addFlag("waitConfirmation", "wait 8 blocks between transactions")
  .addParam("contractName", "logic contract name")
  .addParam(
    "logicAddress",
    "address of the new logic contract to upgrade the proxy to"
  )
  .addParam(
    "factoryAddress",
    "address of factory contract to deploy the contract with"
  )
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam("outputFolder", "output folder path to save factory state")
  .setAction(async (taskArgs, hre) => {
    const waitBlocks = taskArgs.waitConfirmation === true ? 8 : undefined;
    const network = hre.network.name;
    const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
    // grab the salt from the logic contract
    const Salt = await getBytes32Salt(taskArgs.contractName, hre);
    // get logic contract interface
    const logicFactory: any = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    const initArgs =
      taskArgs.initCallData === undefined
        ? []
        : taskArgs.initCallData.replace(/\s+/g, "").split(",");
    const fullname = (await getFullyQualifiedName(
      taskArgs.contractName,
      hre
    )) as string;
    const isInitable = await isInitializable(fullname, hre.artifacts);
    const initCallData = isInitable
      ? logicFactory.interface.encodeFunctionData(INITIALIZER, initArgs)
      : "0x";
    const factory = factoryBase.attach(taskArgs.factoryAddress);
    const txResponse = await factory.upgradeProxy(
      Salt,
      taskArgs.logicAddress,
      initCallData,
      await getGasPrices(hre)
    );
    const receipt = await txResponse.wait(waitBlocks);
    // Data to return to the main task
    const proxyData: ProxyData = {
      proxyAddress: getMetamorphicAddress(taskArgs.factoryAddress, Salt, hre),
      salt: Salt,
      logicName: taskArgs.contractName,
      logicAddress: taskArgs.logicAddress,
      factoryAddress: taskArgs.factoryAddress,
      gas: receipt.gasUsed.toNumber(),
      receipt,
      initCallData,
    };
    await showState(
      `Updated logic with upgradeDeployedProxy for the
      ${taskArgs.contractName}
      contract at
      ${proxyData.proxyAddress}
      gas:
      ${receipt.gasUsed}`
    );
    await updateProxyList(network, proxyData, taskArgs.outputFolder);
    return proxyData;
  });

/**
 * deploys a proxy and upgrades it using multicall from factory
 * @returns a proxyData object with logic contract name, address and proxy salt, and address.
 */
task("multiCallDeployProxy", "deploy and upgrade proxy with multicall")
  .addFlag("waitConfirmation", "wait 8 blocks between transactions")
  .addParam("contractName", "logic contract name")
  .addParam(
    "logicAddress",
    "Address of the logic contract to point the proxy to"
  )
  .addParam(
    "factoryAddress",
    "address of factory contract to deploy the contract with"
  )
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam("outputFolder", "output folder path to save factoryState")
  .addOptionalParam(
    "salt",
    "unique salt for specifying proxy defaults to salt specified in logic contract"
  )
  .setAction(async (taskArgs, hre) => {
    const waitBlocks = taskArgs.waitConfirmation === true ? 8 : undefined;
    const network = hre.network.name;
    const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
    const factory = factoryBase.attach(taskArgs.factoryAddress);
    const logicFactory: any = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    const initArgs =
      taskArgs.initCallData === undefined
        ? []
        : taskArgs.initCallData.replace(/\s+/g, "").split(",");
    const fullname = (await getFullyQualifiedName(
      taskArgs.contractName,
      hre
    )) as string;
    const isInitable = await isInitializable(fullname, hre.artifacts);
    const initCallData = isInitable
      ? logicFactory.interface.encodeFunctionData(INITIALIZER, initArgs)
      : "0x";
    // factory interface pointed to deployed factory contract
    // get the 32byte salt from logic contract file
    const salt: BytesLike =
      taskArgs.salt === undefined
        ? await getBytes32Salt(taskArgs.contractName, hre)
        : hre.ethers.utils.formatBytes32String(taskArgs.salt);
    // encode the deployProxy function call with Salt as arg
    const deployProxyCallData: BytesLike =
      factoryBase.interface.encodeFunctionData(DEPLOY_PROXY, [salt]);
    // encode upgrade proxy multicall
    const upgradeProxyCallData: BytesLike =
      factoryBase.interface.encodeFunctionData(UPGRADE_PROXY, [
        salt,
        taskArgs.logicAddress,
        initCallData,
      ]);
    // get the multi call arguements as [deployProxy, upgradeProxy]

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
    const multiCallArgs = [deployProxy, upgradeProxy];
    // send the multicall transaction with deployProxy and upgradeProxy
    const txResponse = await factory.multiCall(
      multiCallArgs,
      await getGasPrices(hre)
    );
    const receipt = await txResponse.wait(waitBlocks);
    // Data to return to the main task
    const proxyData: ProxyData = {
      factoryAddress: taskArgs.factoryAddress,
      logicName: taskArgs.contractName,
      logicAddress: taskArgs.logicAddress,
      salt,
      proxyAddress: getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR),
      gas: receipt.gasUsed.toNumber(),
      receipt,
      initCallData,
    };

    await updateProxyList(network, proxyData, taskArgs.outputFolder);
    return proxyData;
  });

task(
  "multiCallUpgradeProxy",
  "multi call to deploy logic and upgrade proxy through factory"
)
  .addFlag("waitConfirmation", "wait 8 blocks between transactions")
  .addParam("contractName", "logic contract name")
  .addParam(
    "factoryAddress",
    "address of factory contract to deploy the contract with"
  )
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam(
    "salt",
    "unique salt for specifying proxy defaults to salt specified in logic contract"
  )
  .addOptionalVariadicPositionalParam("constructorArgs")
  .setAction(async (taskArgs, hre) => {
    const waitBlocks = taskArgs.waitConfirmation === true ? 8 : undefined;
    const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
    const factory = factoryBase.attach(taskArgs.factoryAddress);
    const logicFactory: any = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    const initArgs =
      taskArgs.initCallData === undefined
        ? []
        : taskArgs.initCallData.replace(/\s+/g, "").split(",");
    const fullname = (await getFullyQualifiedName(
      taskArgs.contractName,
      hre
    )) as string;
    const isInitable = await isInitializable(fullname, hre.artifacts);
    const initCallData = isInitable
      ? logicFactory.interface.encodeFunctionData(INITIALIZER, initArgs)
      : "0x";
    const deployTx = logicFactory.getDeployTransaction(
      ...taskArgs.constructorArgs
    );
    const deployCreateCallData = factoryBase.interface.encodeFunctionData(
      DEPLOY_CREATE,
      [deployTx.data]
    );
    const salt: string =
      taskArgs.salt === undefined
        ? await getBytes32Salt(taskArgs.contractName, hre)
        : hre.ethers.utils.formatBytes32String(taskArgs.salt);
    const txCount = await hre.ethers.provider.getTransactionCount(
      factory.address
    );
    const implAddress = hre.ethers.utils.getContractAddress({
      from: factory.address,
      nonce: txCount,
    });
    const upgradeProxyCallData = factoryBase.interface.encodeFunctionData(
      UPGRADE_PROXY,
      [salt, implAddress, initCallData]
    );
    const PROXY_FACTORY = await hre.ethers.getContractFactory(PROXY);
    const proxyAddress = getMetamorphicAddress(
      taskArgs.factoryAddress,
      salt,
      hre
    );
    const proxyContract = await PROXY_FACTORY.attach(proxyAddress);
    const oldImpl = await proxyContract.getImplementationAddress();
    const deployCreate = await encodeMultiCallArgs(
      factory.address,
      0,
      deployCreateCallData
    );
    const upgradeProxy = await encodeMultiCallArgs(
      factory.address,
      0,
      upgradeProxyCallData
    );
    const txResponse = await factory.multiCall(
      [deployCreate, upgradeProxy],
      await getGasPrices(hre)
    );
    const receipt = await txResponse.wait(waitBlocks);
    await showState(
      `Updating logic for the ${taskArgs.contractName} proxy at ${proxyAddress} from ${oldImpl} to ${implAddress}, gasCost: ${receipt.gasUsed}`
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
  });

// Generate a json file with all deployment information
task(
  "generateContractsDescriptor",
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
        ? DEFAULT_CONFIG_OUTPUT_DIR
        : taskArgs.outputFolder;
    const path =
      configDirPath === undefined
        ? DEFAULT_CONFIG_OUTPUT_DIR + DEPLOYMENT_LIST_FPATH + ".json"
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
  "deployContractsFromDescriptor",
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
        ? DEFAULT_CONFIG_OUTPUT_DIR
        : taskArgs.outputFolder;
    const path =
      configDirPath === undefined
        ? DEFAULT_CONFIG_OUTPUT_DIR + DEPLOYMENT_LIST_FPATH + ".json"
        : configDirPath + DEPLOYMENT_LIST_FPATH + ".json";
    if (!fs.existsSync(path)) {
      const error =
        "Could not find " +
        DEFAULT_CONFIG_OUTPUT_DIR +
        DEPLOYMENT_LIST_FPATH +
        ".json file. It must be generated first with generateContractsDescriptor task";
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
      const factoryData: FactoryData = await hre.run("deployFactory", {
        outputFolder: taskArgs.outputFolder,
      });
      factoryAddress = factoryData.address;
      cumulativeGasUsed = cumulativeGasUsed.add(factoryData.gas);
    }
    const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
    const factory = factoryBase.attach(factoryAddress);
    const txCount = await hre.ethers.provider.getTransactionCount(
      factory.address
    );
    const contracts = json.contracts;
    await deployContractsMulticall(
      contracts,
      hre,
      factory.address,
      txCount,
      taskArgs.inputFolder,
      taskArgs.outputFolder
    );
    console.log(`total gas used: ${cumulativeGasUsed.toString()}`);
  });

async function checkUserDirPath(path: string) {
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

async function getAccounts(hre: HardhatRuntimeEnvironment) {
  const signers = await hre.ethers.getSigners();
  const accounts: string[] = [];
  for (const signer of signers) {
    accounts.push(signer.address);
  }
  return accounts;
}

async function getFullyQualifiedName(
  contractName: string,
  hre: HardhatRuntimeEnvironment
) {
  const contractArtifact = await hre.artifacts.readArtifact(contractName);
  const path = contractArtifact.sourceName;
  return path + ":" + contractName;
}
/**
 * @description returns everything on the left side of the :
 * ie: src/proxy/Proxy.sol:Mock => src/proxy/Proxy.sol
 * @param qualifiedName the relative path of the contract file + ":" + name of contract
 * @returns the relative path of the contract
 */
function extractPath(qualifiedName: string) {
  return qualifiedName.split(":")[0];
}

/**
 * @description goes through the receipt from the
 * transaction and extract the specified event name and variable
 * @param receipt tx object returned from the tran
 * @param eventName
 * @param varName
 * @returns
 */
export function getEventVar(
  receipt: ContractReceipt,
  eventName: string,
  varName: string
) {
  let result = "0x";
  if (receipt.events !== undefined) {
    const events = receipt.events;
    for (let i = 0; i < events.length; i++) {
      // look for the event
      if (events[i].event === eventName) {
        if (events[i].args !== undefined) {
          const args = events[i].args;
          // extract the deployed mock logic contract address from the event
          result = args !== undefined ? args[varName] : undefined;
          if (result !== undefined) {
            return result;
          }
        } else {
          throw new Error(
            `failed to extract ${varName} from event: ${eventName}`
          );
        }
      }
    }
  }
  throw new Error(`failed to find event: ${eventName}`);
}
/**
 * @description extracts the value specified by custom:salt from the contracts artifacts
 * buildInfo
 * @param contractName the name of the contract to get the salt for
 * @param hre hardhat runtime environment
 * @returns the string version of salt specified by custom:salt
 *  NatSpec tag in the contract
 */
async function getSalt(
  contractName: string,
  hre: HardhatRuntimeEnvironment
): Promise<string> {
  const qualifiedName: any = await getFullyQualifiedName(contractName, hre);
  const buildInfo = await hre.artifacts.getBuildInfo(qualifiedName);
  let contractOutput: any;
  let devdoc: any;
  let salt: string = "";
  if (buildInfo !== undefined) {
    const path = extractPath(qualifiedName);
    contractOutput = buildInfo.output.contracts[path][contractName];
    devdoc = contractOutput.devdoc;
    salt = devdoc["custom:salt"];
    return salt;
  } else {
    console.error("missing salt");
  }
  return salt;
}

/**
 * @description gets the salt from the contract natspec tag and converts it to bytes32
 * @param contractName the name of the contract to get the salt for
 * @param hre hardhat runtime environment
 * @returns the string that represents the 32Bytes version
 * of the salt specified by custom:salt
 */
export async function getBytes32Salt(
  contractName: string,
  hre: HardhatRuntimeEnvironment
) {
  const salt: string = await getSalt(contractName, hre);
  return hre.ethers.utils.formatBytes32String(salt);
}

/**
 *
 * @param factoryAddress address of the factory that deployed the contract
 * @param salt value specified by custom:salt in the contrac
 * @param hre hardhat runtime environment
 * @returns returns the address of the metamorphic contract
 */
function getMetamorphicAddress(
  factoryAddress: string,
  salt: string,
  hre: HardhatRuntimeEnvironment
) {
  const initCode = "0x6020363636335afa1536363636515af43d36363e3d36f3";
  return hre.ethers.utils.getCreate2Address(
    factoryAddress,
    salt,
    hre.ethers.utils.keccak256(initCode)
  );
}

export const showState = async (message: string): Promise<void> => {
  if (process.env.silencer === undefined || process.env.silencer === "false") {
    console.log(message);
  }
};

export async function getGasPrices(hre: HardhatRuntimeEnvironment) {
  //get the latest block
  let blockBaseFee = await hre.network.provider.send("eth_gasPrice");
  //get the previous basefee from the latest block
  blockBaseFee = BigNumber.from(blockBaseFee);
  blockBaseFee = blockBaseFee.toBigInt();
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
