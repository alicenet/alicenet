import { BytesLike, ContractFactory, ContractReceipt } from "ethers";
import fs from "fs";
import { task } from "hardhat/config";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import {
  CONTRACT_ADDR,
  DEFAULT_CONFIG_OUTPUT_DIR,
  DEPLOYED_PROXY,
  DEPLOYED_RAW,
  DEPLOYED_STATIC,
  DEPLOY_CREATE,
  DEPLOY_METAMORPHIC,
  DEPLOY_PROXY,
  DEPLOY_UPGRADEABLE_PROXY,
  FACTORY_STATE_PATH,
  INITIALIZER,
  MADNET_FACTORY,
  MULTI_CALL_DEPLOY_PROXY,
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
import { readFactoryState } from "./deployment/deploymentConfigUtil";
import {
  DeploymentList,
  getDeploymentList,
  getSortedDeployList,
  transformDeploymentList,
  writeDeploymentList,
} from "./deployment/deploymentListUtil";
import {
  deployFactory,
  DeploymentArgs,
  deployStatic,
  deployUpgradeableProxy,
  getAllContracts,
  getDeployType,
  isInitializable,
} from "./deployment/deploymentUtil";
import {
  DeployCreateData,
  FactoryData,
  MetaContractData,
  ProxyData,
  TemplateData,
  updateDefaultFactoryData,
  updateDeployCreateList,
  updateMetaList,
  updateProxyList,
  updateTemplateList,
} from "./deployment/factoryStateUtil";

type DeployProxyMCArgs = {
  contractName: string;
  logicAddress: string;
  factoryAddress?: string;
  initCallData?: BytesLike;
  outputFolder?: string;
};

type DeployArgs = {
  contractName: string;
  factoryAddress?: string;
  initCallData?: string;
  constructorArgs?: any;
  outputFolder?: string;
};

export type Args = {
  contractName: string;
  factoryAddress?: string;
  salt?: BytesLike;
  initCallData?: string;
  constructorArgs?: any;
  outputFolder?: string;
};

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
  .addOptionalParam("outputFolder")
  .setAction(async (taskArgs, hre) => {
    await checkUserDirPath(taskArgs.outputFolder);
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    const accounts = await getAccounts(hre);
    const txCount = await hre.ethers.provider.getTransactionCount(accounts[0]);
    // calculate the factory address for the constructor arg
    const futureFactoryAddress = hre.ethers.utils.getContractAddress({
      from: accounts[0],
      nonce: txCount,
    });
    const deployTX = factoryBase.getDeployTransaction(futureFactoryAddress);
    const gasCost = await hre.ethers.provider.estimateGas(deployTX);
    // deploys the factory
    const factory = await factoryBase.deploy(futureFactoryAddress);
    await factory.deployTransaction.wait();
    // record the data in a json file to be used in other tasks
    const factoryData: FactoryData = {
      address: factory.address,
      owner: accounts[0],
      gas: gasCost.toNumber(),
    };
    const network = hre.network.name;
    await updateDefaultFactoryData(network, factoryData, taskArgs.outputFolder);
    await showState(
      `Deployed: ${MADNET_FACTORY}, at address: ${factory.address}`
    );
    return factory.address;
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
      deploymentList = await getSortedDeployList(contracts, hre.artifacts);
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
      await writeDeploymentList(list, path);
    }
    if (taskArgs.list !== true) {
      await writeDeploymentArgs(deploymentArgs, path);
      console.log(
        `YOU MUST REPLACE THE UNDEFINED VALUES IN ${path}/deploymentArgsTemplate`
      );
    }
  });

task("deployContracts", "runs the initial deployment of all madnet contracts")
  .addFlag(
    "deployFactory",
    "flag to indicate deployment, will deploy the factory first if set"
  )
  .addOptionalParam("inputFolder", "path to location containing deploymentArgsTemplate, and deploymentList")
  .addOptionalParam(
    "outputFolder",
    "output folder path to save factory state"
  )
  .setAction(async (taskArgs, hre) => {
    await checkUserDirPath(taskArgs.outputFolder);
    // setting listName undefined will use the default list
    const ethers = hre.ethers;
    const artifacts = hre.artifacts;
    const run = hre.run;
    // deploy the factory first
    if (taskArgs.deployFactory === true) {
      await deployFactory(run, taskArgs.outputFolder);
    }
    // get an array of all contracts in the artifacts
    const contracts = await getDeploymentList(taskArgs.inputFolder);
    // let contracts = ["src/tokens/periphery/validatorPool/Snapshots.sol:Snapshots"]
    for (let i = 0; i < contracts.length; i++) {
      const fullyQualifiedName = contracts[i];
      // check the contract for the @custom:deploy-type tag
      const deployType = await getDeployType(fullyQualifiedName, artifacts);
      switch (deployType) {
        case STATIC_DEPLOYMENT:
          await deployStatic(
            fullyQualifiedName,
            artifacts,
            ethers,
            run,
            taskArgs.inputFolder,
            taskArgs.outputFolder
          );
          break;
        case UPGRADEABLE_DEPLOYMENT:
          await deployUpgradeableProxy(
            fullyQualifiedName,
            artifacts,
            ethers,
            run,
            taskArgs.outputFolder
          );
          break;
        default:
          break;
      }
    }
  });

task(
  DEPLOY_UPGRADEABLE_PROXY,
  "deploys logic contract, proxy contract, and points the proxy to the logic contract"
)
  .addParam(
    "contractName",
    "Name of logic contract to point the proxy at",
    "string"
  )
  .addOptionalParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam("outputFolder", "output folder path to save factory state")
  .addOptionalVariadicPositionalParam("constructorArgs", "")
  .setAction(async (taskArgs, hre) => {
    const network = hre.network.name;
    const factoryAddress = await getFactoryAddress(network, taskArgs);
    // uses the factory Data and logic contractName and returns deploybytecode and any constructor args attached
    const callArgs: DeployArgs = {
      contractName: taskArgs.contractName,
      factoryAddress,
      constructorArgs: taskArgs.constructorArgs,
      outputFolder: taskArgs.outputFolder,
    };
    // deploy create the logic contract
    const result: DeployCreateData = await hre.run(DEPLOY_CREATE, callArgs);
    const mcCallArgs: DeployProxyMCArgs = {
      contractName: taskArgs.contractName,
      factoryAddress,
      logicAddress: result.address,
      initCallData: taskArgs.initCallData,
      outputFolder: taskArgs.outputFolder,
    };
    const proxyData: ProxyData = await hre.run(
      MULTI_CALL_DEPLOY_PROXY,
      mcCallArgs
    );
    return proxyData;
  });

task(
  DEPLOY_METAMORPHIC,
  "deploys template contract, and then deploys metamorphic contract, and points the proxy to the logic contract"
)
  .addParam(
    "contractName",
    "Name of logic contract to point the proxy at",
    "string"
  )
  .addOptionalParam(
    "factoryAddress",
    "the default factory address from factoryState will be used if not set"
  )
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam("outputFolder", "output folder path to save factory state")
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguements for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    const network = hre.network.name;
    const factoryAddress = await getFactoryAddress(network, taskArgs);
    let callArgs: DeployArgs = {
      contractName: taskArgs.contractName,
      factoryAddress,
      constructorArgs: taskArgs?.constructorArgs,
      outputFolder: taskArgs.outputFolder,
    };
    // deploy create the logic contract
    await hre.run("deployTemplate", callArgs);
    callArgs = {
      contractName: taskArgs.contractName,
      factoryAddress,
      initCallData: taskArgs.initCallData,
      outputFolder: taskArgs.outputFolder,
    };
    const metaContractData = await hre.run("deployStatic", callArgs);
    await showState(
      `Deployed Metamorphic for ${taskArgs.contractName} at: ${metaContractData.metaAddress}, with logic from, ${metaContractData.templateAddress}, gas used: ${metaContractData.gas}`
    );
    return metaContractData;
  });

// factoryName param doesnt do anything right now
task(DEPLOY_CREATE, "deploys a contract from the factory using create")
  .addParam("contractName", "logic contract name")
  .addOptionalParam("factoryAddress", "the default factory address from factoryState will be used if not set")
  .addOptionalParam("outputFolder", "output folder path to save factory state")
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguments for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    const network = hre.network.name;
    const factoryAddress = await getFactoryAddress(network, taskArgs);
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    // get a factory instance connected to the factory a
    const factory = factoryBase.attach(factoryAddress);
    const logicContract: ContractFactory = await hre.ethers.getContractFactory(
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
      const txResponse = await factory.deployCreate(deployTx.data);
      const receipt = await txResponse.wait();
      const deployCreateData: DeployCreateData = {
        name: taskArgs.contractName,
        address: getEventVar(receipt, DEPLOYED_RAW, CONTRACT_ADDR),
        factoryAddress,
        gas: receipt.gasUsed.toNumber(),
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
  .addParam(
    "salt",
    "salt used to specify logicContract and proxy address calculation"
  )
  .addOptionalParam("factoryAddress", "the default factory address from factoryState will be used if not set")
  .setAction(async (taskArgs, hre) => {
    const network = hre.network.name;
    const factoryAddress = await getFactoryAddress(network, taskArgs);
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    const factory = factoryBase.attach(factoryAddress);
    const txResponse = await factory.deployProxy(taskArgs.salt);
    const receipt = await txResponse.wait();
    const proxyAddr = getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR);
    const proxyData: ProxyData = {
      proxyAddress: proxyAddr,
      salt: taskArgs.salt,
      factoryAddress,
      gas: receipt.gasUsed,
      receipt,
    };
    return proxyData;
  });

task(UPGRADE_DEPLOYED_PROXY, "deploys a contract from the factory using create")
  .addParam("contractName", "logic contract name")
  .addParam(
    "logicAddress",
    "address of the new logic contract to upgrade the proxy to"
  )
  .addOptionalParam("factoryAddress", "factory deploying the contract")
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam(
    "outputFolder",
    "output folder path to save factory state"
  )
  .setAction(async (taskArgs, hre) => {
    const network = hre.network.name;
    const factoryAddress = await getFactoryAddress(network, taskArgs);
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    // grab the salt from the logic contract
    const Salt = await getBytes32Salt(taskArgs.contractName, hre);
    // get logic contract interface
    const logicFactory = await hre.ethers.getContractFactory(
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
    const factory = factoryBase.attach(factoryAddress);
    const txResponse = await factory.upgradeProxy(
      Salt,
      taskArgs.logicAddress,
      initCallData
    );
    const receipt = await txResponse.wait();
    // Data to return to the main task
    const proxyData: ProxyData = {
      proxyAddress: getMetamorphicAddress(factoryAddress, Salt, hre),
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

task(
  "deployTemplate",
  "deploys a template contract with the universal code copy constructor that deploys"
)
  .addParam("contractName", "logic contract name")
  .addOptionalParam(
    "factoryAddress",
    "optional factory address, defaults to config address"
  )
  .addOptionalParam(
    "outputFolder",
    "output folder path to save factory state"
  )
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "input constructor args at the end of call"
  )
  .setAction(async (taskArgs, hre) => {
    const network = hre.network.name;
    if (network === "hardhat") {
      // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
      // being sent as input to the function (the contract bytecode), so we need to increase the block
      // gas limit temporally in order to deploy the template
      await hre.network.provider.send("evm_setBlockGasLimit", [
        "0x3000000000000000",
      ]);
    }
    const factoryAddress = await getFactoryAddress(network, taskArgs);
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    const factory = factoryBase.attach(factoryAddress);
    const logicContract: ContractFactory = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    const constructorArgs =
      taskArgs.constructorArgs === undefined ? [] : taskArgs.constructorArgs;
    const deployTxReq = logicContract.getDeployTransaction(...constructorArgs);
    if (deployTxReq.data !== undefined) {
      const deployBytecode = deployTxReq.data;
      const txResponse = await factory.deployTemplate(deployBytecode);
      const receipt = await txResponse.wait();
      const templateData: TemplateData = {
        name: taskArgs.contractName,
        address: getEventVar(receipt, "DeployedTemplate", CONTRACT_ADDR),
        factoryAddress,
        gas: receipt.gasUsed.toNumber(),
        receipt,
        constructorArgs,
      };
      //   await showState(`Subtask deployedTemplate for ${taskArgs.contractName} contract at ${templateData.address}, gas: ${receipt.gasUsed}`);
      await updateTemplateList(network, templateData, taskArgs.outputFolder);
      return templateData;
    } else {
      throw new Error(
        `failed to get contract bytecode for ${taskArgs.contractName}`
      );
    }
  });

// takes in optional
task(
  "deployStatic",
  "deploys a template contract with the universal code copy constructor that deploys"
)
  .addParam("contractName", "logic contract name")
  .addOptionalParam("factoryAddress", "factory deploying the contract")
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam(
    "outputFolder",
    "output folder path to save factoryState"
  )
  .setAction(async (taskArgs, hre) => {
    const network = hre.network.name;
    const factoryAddress = await getFactoryAddress(network, taskArgs);
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    const logicFactory = await hre.ethers.getContractFactory(
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
    const Salt = await getBytes32Salt(taskArgs.contractName, hre);
    // get a factory instance connected to the factory addr
    const factory = factoryBase.attach(factoryAddress);
    // TODO: Reconsider doing this, might get the wrong implementation address
    const tmplAddress = await factory.callStatic.getImplementation();
    const txResponse = await factory.deployStatic(Salt, initCallData);
    const receipt = await txResponse.wait();
    const contractAddr = getEventVar(receipt, DEPLOYED_STATIC, CONTRACT_ADDR);
    // await showState(`Subtask deployStatic, ${taskArgs.contractName}, contract at ${contractAddr}, gas: ${receipt.gasUsed}`);
    const outputData: MetaContractData = {
      metaAddress: contractAddr,
      salt: Salt,
      templateName: taskArgs.contractName,
      templateAddress: tmplAddress,
      factoryAddress: factory.address,
      gas: receipt.gasUsed.toNumber(),
      receipt,
      initCallData,
    };
    await updateMetaList(network, outputData, taskArgs.outputFolder);
    return outputData;
  });

/**
 * deploys a proxy and upgrades it using multicall from factory
 * @returns a proxyData object with logic contract name, address and proxy salt, and address.
 */
task("multiCallDeployProxy", "deploy and upgrade proxy with multicall")
  .addParam("contractName", "logic contract name")
  .addParam(
    "logicAddress",
    "Address of the logic contract to point the proxy to"
  )
  .addOptionalParam("factoryAddress", "factory deploying the contract")
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalParam(
    "outputFolder",
    "output folder path to save factoryState"
  )
  .addOptionalParam(
    "salt",
    "unique salt for specifying proxy defaults to salt specified in logic contract"
  )
  .setAction(async (taskArgs, hre) => {
    const network = hre.network.name;
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    const factoryAddress = await getFactoryAddress(network, taskArgs);
    const factory = factoryBase.attach(factoryAddress);
    const logicFactory = await hre.ethers.getContractFactory(
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
    const deployProxy: BytesLike = factoryBase.interface.encodeFunctionData(
      DEPLOY_PROXY,
      [salt]
    );
    // encode upgrade proxy multicall
    const upgradeProxy: BytesLike = factoryBase.interface.encodeFunctionData(
      UPGRADE_PROXY,
      [salt, taskArgs.logicAddress, initCallData]
    );
    // get the multi call arguements as [deployProxy, upgradeProxy]
    const multiCallArgs = [deployProxy, upgradeProxy];
    // send the multicall transaction with deployProxy and upgradeProxy
    const txResponse = await factory.multiCall(multiCallArgs);
    const receipt = await txResponse.wait();
    // Data to return to the main task
    const proxyData: ProxyData = {
      factoryAddress,
      logicName: taskArgs.contractName,
      logicAddress: taskArgs.logicAddress,
      salt,
      proxyAddress: getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR),
      gas: receipt.gasUsed.toNumber(),
      receipt,
      initCallData,
    };
    await showState(
      `Deployed ${proxyData.logicName} with proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`
    );
    await updateProxyList(network, proxyData, taskArgs.outputFolder);
    return proxyData;
  });

task(
  "multiCallUpgradeProxy",
  "multi call to deploy logic and upgrade proxy through factory"
)
  .addParam("contractName", "logic contract name")
  .addOptionalParam("factoryAddress", "the default factory address from factoryState will be used if not set")
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
    const network = hre.network.name;
    const factoryAddress = await getFactoryAddress(network, taskArgs);
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    const factory = factoryBase.attach(factoryAddress);
    const logicFactory: ContractFactory = await hre.ethers.getContractFactory(
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
    const deployCreate = factoryBase.interface.encodeFunctionData(
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
    const upgradeProxy = factoryBase.interface.encodeFunctionData(
      DEPLOY_CREATE,
      [salt, implAddress, initCallData]
    );
    const PROXY_FACTORY = await hre.ethers.getContractFactory(PROXY);
    const proxyAddress = getMetamorphicAddress(factoryAddress, salt, hre);
    const proxyContract = await PROXY_FACTORY.attach(proxyAddress);
    const oldImpl = await proxyContract.getImplementationAddress();
    const txResponse = await factory.multiCall([deployCreate, upgradeProxy]);
    const receipt = await txResponse.wait();
    await showState(
      `Updating logic for the ${taskArgs.contractName} proxy at ${proxyAddress} from ${oldImpl} to ${implAddress}, gasCost: ${receipt.gasUsed}`
    );
    const proxyData: ProxyData = {
      factoryAddress,
      logicName: taskArgs.contractName,
      logicAddress: taskArgs.logicAddress,
      salt,
      proxyAddress: getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR),
      gas: receipt.gasUsed.toNumber(),
      receipt,
      initCallData,
    };
    return proxyData;
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

async function getFactoryAddress(network: string, taskArgs: any) {
  // get Factory data from factoryConfig.json
  const path =
    taskArgs.outputFolder === undefined
      ? FACTORY_STATE_PATH
      : taskArgs.outputFolder.replace(/\/+$/, "") + "/factoryState";
  const factoryConfig = await readFactoryState(path);
  const configFactoryAddress = factoryConfig[network].defaultFactoryAddress;
  const cliFactoryAddress = taskArgs.factoryAddress;
  // object to store data to update config var
  let factoryAddress: string;
  // check if the user provided factory Data in call
  if (cliFactoryAddress !== undefined) {
    factoryAddress = cliFactoryAddress;
  }
  // if the user did not provide factory data check for factory data in factoryConfig
  else if (configFactoryAddress !== undefined) {
    factoryAddress = configFactoryAddress;
  }
  // if no factoryData provided in call and in userConfig throw error
  else {
    throw new Error(
      "Failed to find factory address via cli arguement or factory state file"
    );
  }
  return factoryAddress;
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
  const artifactPaths = await hre.artifacts.getAllFullyQualifiedNames();
  for (let i = 0; i < artifactPaths.length; i++) {
    if (artifactPaths[i].split(":")[1] === contractName) {
      return String(artifactPaths[i]);
    }
  }
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
function getEventVar(
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
 * @description converts
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
