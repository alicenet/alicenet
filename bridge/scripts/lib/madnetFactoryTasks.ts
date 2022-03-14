import { DEFAULT_CONFIG_OUTPUT_DIR, DEPLOYMENT_ARG_PATH, DEPLOYMENT_LIST_PATH, DEPLOY_METAMORPHIC, DEPLOY_PROXY, DEPLOY_UPGRADEABLE_PROXY, FACTORY_STATE_PATH, INITIALIZER, MULTI_CALL_DEPLOY_PROXY, STATIC_DEPLOYMENT, UPGRADEABLE_DEPLOYMENT, UPGRADE_DEPLOYED_PROXY, UPGRADE_PROXY } from "./constants";
import { DeploymentList, getDeploymentList, getSortedDeployList, transformDeploymentList, writeDeploymentList } from "./deployment/deploymentListUtil";
import { getDeployType, deployUpgradeableProxy, deployStatic, deployFactory, getAllContracts, extractName, ArgData, DeploymentArgs} from "./deployment/deploymentUtil";
import { BytesLike, ContractFactory, ContractReceipt } from "ethers";
import { task } from "hardhat/config";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import {
  CONTRACT_ADDR,
  DEPLOYED_PROXY,
  DEPLOYED_RAW,
  DEPLOYED_STATIC,
  DEPLOY_CREATE,
  MADNET_FACTORY,
  PROXY,
} from "./constants";
import { readFactoryState } from "./deployment/deploymentConfigUtil";
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
import { generateDeployArgTemplate, getConstructorArgsABI, getInitializerArgsABI, writeDeploymentArgs } from "./deployment/deployArgUtil";
import { expect } from "chai";
import fs from "fs";


type DeployProxyMCArgs = {
  contractName: string;
  logicAddress: string;
  factoryAddress?: string;
  initCallData?: BytesLike;
};

type DeployArgs = {
  contractName: string;
  factoryAddress?: string;
  initCallData?: string;
  constructorArgs?: any;
};

export type Args = {
  contractName: string;
  factoryAddress?: string;
  salt?: BytesLike;
  initCallData?: string;
  constructorArgs?: any;
};


task("getNetwork", "gets the current network being used from provider")
  .setAction(async (taskArgs, hre) => {
    let network = hre.network.name
    await showState(network);
    return network
  });

task("getBytes32Salt", "gets the bytes32 version of salt from contract")
  .addParam("contractName", "test contract")
  .setAction(async (taskArgs, hre) => {
    let salt = await getBytes32Salt(taskArgs.contractName, hre);
    await showState(salt);
  });

task(
  "deployFactory",
  "Deploys an instance of a factory contract specified by its name"
).setAction(async (taskArgs, hre) => {
  const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
  const accounts = await getAccounts(hre);
  let txCount = await hre.ethers.provider.getTransactionCount(accounts[0]);
  //calculate the factory address for the constructor arg
  let futureFactoryAddress = hre.ethers.utils.getContractAddress({
    from: accounts[0],
    nonce: txCount,
  });
  let deployTX = factoryBase.getDeployTransaction(futureFactoryAddress);
  let gasCost = await hre.ethers.provider.estimateGas(deployTX);
  //deploys the factory
  let factory = await factoryBase.deploy(futureFactoryAddress);
  await factory.deployTransaction.wait();
  //record the data in a json file to be used in other tasks
  let factoryData: FactoryData = {
    address: factory.address,
    owner: accounts[0],
    gas: gasCost.toNumber(),
  };
  let network = hre.network.name;
  let fileName = process.env.test === "true" ? "testFactoryState" : undefined;
  await updateDefaultFactoryData(network, factoryData, fileName);
  await showState(
    `Deployed: ${MADNET_FACTORY}, at address: ${factory.address}`
  );
  return factory.address;
});

task("generateDeploymentConfigs", 
"default list and arg template will be generated if all optional variables are not specified"
)
.addFlag('list', "flag to only generate deploy list")
.addFlag('args', "flag to only generate deploy args template")
.addOptionalParam("outputFolder", "output folder path to save deployment arg template and list")
.addOptionalVariadicPositionalParam("contractNames", "custom list of contracts to generate list and arg template for")
.setAction(async (taskArgs, hre) =>{
  if(taskArgs.outputFolder !== undefined){
    if(!fs.existsSync(taskArgs.outputFolder)){
      fs.mkdirSync(taskArgs.outputFolder)
    }
    if( fs.statSync(taskArgs.outputFolder).isFile()){
      throw new Error("outputFolder path should be to a directory not a file")
    }
  }
  let path = taskArgs.outputFolder === undefined ? DEFAULT_CONFIG_OUTPUT_DIR : taskArgs.outputFolder; 
  let deploymentList: DeploymentList
  let deploymentArgs:DeploymentArgs = {
    constructor: {},
    initializer: {}
  };
  let list: Array<string>
  //no custom path and list input/ writes arg template in default scripts/base-files/deploymentArgs
  if(taskArgs.outputFolder === undefined && taskArgs.contractNames === undefined){
    //create the default list
    //setting list name will specify default configs
    let contracts = await getAllContracts(hre.artifacts);
    deploymentList = await getSortedDeployList(contracts, hre.artifacts);
    list = await transformDeploymentList(deploymentList);
    deploymentArgs = await generateDeployArgTemplate(list, hre.artifacts);
  }//user defined path and list 
  else if(taskArgs.outputFolder !== undefined && taskArgs.contractNames !== undefined){
    //create deploy list and deployment arg with the specified output path
    let nameList: Array<string> = taskArgs.contractNames;
    let contracts: Array<string> = []
    for(let name of nameList){
      let sourceName = (await hre.artifacts.readArtifact(name)).sourceName
      let fullName = sourceName + ":" + name
      //this will cause the operation to fail if deployType is not specified on the contract
      await getDeployType(fullName, hre.artifacts)
      contracts.push(fullName)
    }
    deploymentList = await getSortedDeployList(contracts, hre.artifacts);
    list = await transformDeploymentList(deploymentList);
    deploymentArgs = await generateDeployArgTemplate(list, hre.artifacts);
  }//user defined path, default list 
  else if(taskArgs.outputFolder !== undefined && taskArgs.contractNames === undefined){
    let contracts = await getAllContracts(hre.artifacts);
    deploymentList = await getSortedDeployList(contracts, hre.artifacts);
    list = await transformDeploymentList(deploymentList);
    deploymentArgs = await generateDeployArgTemplate(list, hre.artifacts);
  } 
  else{
    throw new Error("you must specify a path to store your custom deploy config files")
  }
  if(taskArgs.args !== true){
    await writeDeploymentList(list, path);
  }
  if(taskArgs.list !== true){
    await writeDeploymentArgs(deploymentArgs, path)
  }
});

task("deployContracts", "runs the initial deployment of all madnet contracts")
  .addFlag("factory", "flag to indicate deployment, will deploy the factory first if set")
  .addOptionalParam("configDir", "output folder path to save deployment arg template and list")
  .setAction(async (taskArgs, hre) =>{
    if(taskArgs.outputFolder !== undefined && fs.statSync(taskArgs.outputFolder).isFile()){
      throw new Error("outputFolder path should be to a directory not a file")
    }
    let path = taskArgs.configDir === undefined ? DEFAULT_CONFIG_OUTPUT_DIR : taskArgs.configDir; 
    //setting listName undefined will use the default list 
    let ethers = hre.ethers
    let artifacts = hre.artifacts;
    let run = hre.run;
    //deploy the factory first
    if(taskArgs.factory === true){
      await deployFactory(run);
    }
    //get an array of all contracts in the artifacts
    let contracts = await getDeploymentList(path);
    //let contracts = ["src/tokens/periphery/validatorPool/Snapshots.sol:Snapshots"]
    for (let i = 0; i < contracts.length; i++) {
      let fullyQualifiedName = contracts[i];
      //check the contract for the @custom:deploy-type tag
      let deployType = await getDeployType(fullyQualifiedName, artifacts);
      switch (deployType) {
        case STATIC_DEPLOYMENT:
          await deployStatic(fullyQualifiedName, artifacts, ethers, run, path);
          break;
        case UPGRADEABLE_DEPLOYMENT:
          await deployUpgradeableProxy(fullyQualifiedName, artifacts, ethers, run, path);
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
    "address of the factory deploying the contract"
  )
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    ""
  )
  .setAction(async (taskArgs, hre) => {
    let network = hre.network.name;
    let factoryAddress = await getFactoryAddress(network, taskArgs);
    //uses the factory Data and logic contractName and returns deploybytecode and any constructor args attached
    let callArgs: DeployArgs = {
      contractName: taskArgs.contractName,
      factoryAddress: factoryAddress,
      constructorArgs: taskArgs.constructorArgs,
    };
    //deploy create the logic contract
    let result: DeployCreateData = await hre.run(DEPLOY_CREATE, callArgs);
    let mcCallArgs: DeployProxyMCArgs = {
      contractName: taskArgs.contractName,
      logicAddress: result.address,
      initCallData: taskArgs.initCallData,
    };
    let proxyData: ProxyData = await hre.run(MULTI_CALL_DEPLOY_PROXY, mcCallArgs);
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
    "address of the factory deploying the contract"
  )
  .addOptionalParam(
    "initCallData",
    "input initCallData args in a string list, eg: --initCallData 'arg1, arg2'"
  )
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguements for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    let network = hre.network.name;
    let factoryData = await getFactoryAddress(network, taskArgs);
    //uses the factory Data and logic contractName and returns deploybytecode and any constructor args attached
    let factoryAddress = await getFactoryAddress(network, taskArgs);
    let callArgs: DeployArgs = {
      contractName: taskArgs.contractName,
      factoryAddress: factoryAddress,
      constructorArgs: taskArgs?.constructorArgs,
    };
    //deploy create the logic contract
    let templateData: TemplateData = await hre.run("deployTemplate", callArgs);
    callArgs = {
      contractName: taskArgs.contractName,
      factoryAddress: factoryAddress,
      initCallData: taskArgs.initCallData,
    };
    let metaContractData = await hre.run("deployStatic", callArgs);
    await showState(
      `Deployed Metamorphic for ${taskArgs.contractName} at: ${metaContractData.metaAddress}, with logic from, ${metaContractData.templateAddress}, gas used: ${metaContractData.gas}`
    );
    return metaContractData;
  });

//factoryName param doesnt do anything right now
task(DEPLOY_CREATE, "deploys a contract from the factory using create")
  .addParam("contractName", "logic contract name")
  .addOptionalParam("factoryAddress", "factory deploying the contract")
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguments for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    let network = hre.network.name;
    let factoryAddress = await getFactoryAddress(network, taskArgs);
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    //get a factory instance connected to the factory a
    const factory = factoryBase.attach(factoryAddress);
    let logicContract: ContractFactory = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    let constructorArgs =
      taskArgs.constructorArgs === undefined ? [] : taskArgs.constructorArgs;
    //encode deployBcode
    let deployTx = logicContract.getDeployTransaction(...constructorArgs);
    if (hre.network.name === "hardhat") {
      // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
      // being sent as input to the function (the contract bytecode), so we need to increase the block
      // gas limit temporally in order to deploy the template
      await hre.network.provider.send("evm_setBlockGasLimit", [
        "0x3000000000000000",
      ]);
    }
    if (deployTx.data !== undefined) {
      let txResponse = await factory.deployCreate(deployTx.data);
      let receipt = await txResponse.wait();
      let deployCreateData: DeployCreateData = {
        name: taskArgs.contractName,
        address: getEventVar(receipt, DEPLOYED_RAW, CONTRACT_ADDR),
        factoryAddress: factoryAddress,
        gas: receipt.gasUsed.toNumber(),
        receipt: receipt,
        constructorArgs: taskArgs?.constructorArgs,
      };
      let network = hre.network.name;
      await updateDeployCreateList(network, deployCreateData);
      await showState(
        `[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${taskArgs.contractName} contract at: ${deployCreateData.address}, gas: ${receipt.gasUsed}`
      );
      return deployCreateData;
    } else {
      throw new Error(
        `failed to get deployment bytecode for ${taskArgs.contractName}`
      );
    }
  });

task(DEPLOY_PROXY, "deploys a proxy from the factory")
  .addParam("salt", "salt used to specify logicContract and proxy address calculation")
  .addOptionalParam("factoryAddress")
  .setAction(async (taskArgs, hre) => {
    let network = hre.network.name;
    let factoryAddress = await getFactoryAddress(network, taskArgs);
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    let factory = factoryBase.attach(factoryAddress);
    let txResponse = await factory.deployProxy(taskArgs.salt);
    let receipt = await txResponse.wait();
    let proxyAddr = getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR)
    let proxyData: ProxyData = {
      proxyAddress: proxyAddr,
      salt: taskArgs.salt,
      factoryAddress: factoryAddress,
      gas: receipt.gasUsed,
      receipt: receipt,
    }
    return proxyData
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
  .setAction(async (taskArgs, hre) => {
    let network = hre.network.name;
    let factoryAddress = await getFactoryAddress(network, taskArgs);
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    //grab the salt from the logic contract
    let Salt = await getBytes32Salt(taskArgs.contractName, hre);
    //get logic contract interface
    let contractFactory = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    let initCallData: string =
      taskArgs.initCallData === undefined ? "0x" : contractFactory.interface.encodeFunctionData(INITIALIZER, taskArgs.initCallData.split(", "));
    const factory = factoryBase.attach(factoryAddress);
    let txResponse = await factory.upgradeProxy(
      Salt,
      taskArgs.logicAddress,
      initCallData
    );
    let receipt = await txResponse.wait();
    //Data to return to the main task
    let proxyData: ProxyData = {
      proxyAddress: getMetamorphicAddress(factoryAddress, Salt, hre),
      salt: Salt,
      logicName: taskArgs.contractName,
      logicAddress: taskArgs.logicAddress,
      factoryAddress: taskArgs.factoryAddress,
      gas: receipt.gasUsed.toNumber(),
      receipt: receipt,
      initCallData: initCallData,
    };
    await showState(
      `Updated logic with upgradeDeployedProxy for the
      ${taskArgs.contractName}
      contract at
      ${proxyData.proxyAddress}
      gas:
      ${receipt.gasUsed}`
    );
    await updateProxyList(network, proxyData);
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
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "input constructor args at the end of call"
  )
  .setAction(async (taskArgs, hre) => {
    let network = hre.network.name;
    if (network === "hardhat") {
      // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
      // being sent as input to the function (the contract bytecode), so we need to increase the block
      // gas limit temporally in order to deploy the template
      await hre.network.provider.send("evm_setBlockGasLimit", [
        "0x3000000000000000",
      ]);
    }
    let factoryAddress = await getFactoryAddress(network, taskArgs);
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    const factory = factoryBase.attach(factoryAddress);
    let logicContract: ContractFactory = await hre.ethers.getContractFactory(taskArgs.contractName);
    let constructorArgs = taskArgs.constructorArgs === undefined ? [] : taskArgs.constructorArgs  
    let deployTxReq = logicContract.getDeployTransaction(...constructorArgs)
    if (deployTxReq.data !== undefined) {
      let deployBytecode = deployTxReq.data;
      let txResponse = await factory.deployTemplate(deployBytecode);
      let receipt = await txResponse.wait();
      let templateData: TemplateData = {
        name: taskArgs.contractName,
        address: getEventVar(receipt, "DeployedTemplate", CONTRACT_ADDR),
        factoryAddress: factoryAddress,
        gas: receipt.gasUsed.toNumber(),
        receipt: receipt,
        constructorArgs: constructorArgs,
      };
      //   await showState(`Subtask deployedTemplate for ${taskArgs.contractName} contract at ${templateData.address}, gas: ${receipt.gasUsed}`);
      updateTemplateList(network, templateData);
      return templateData;
    } else {
      throw new Error(
        `failed to get contract bytecode for ${taskArgs.contractName}`
      );
    }
  });

//takes in optional
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
  .setAction(async (taskArgs, hre) => {
    let network = hre.network.name;
    let factoryAddress = await getFactoryAddress(network, taskArgs);
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    let logicFactory = await hre.ethers.getContractFactory(taskArgs.contractName);
    let initCallData =
      taskArgs.initCallData === undefined ? "0x" : logicFactory.interface.encodeFunctionData(INITIALIZER, taskArgs.initCallData.split(", "));
    let Salt = await getBytes32Salt(taskArgs.contractName, hre);
    //get a factory instance connected to the factory addr
    const factory = factoryBase.attach(factoryAddress);
    //TODO: Reconsider doing this, might get the wrong implementation address
    let tmplAddress = await factory.callStatic.getImplementation();
    let txResponse = await factory.deployStatic(Salt, initCallData);
    let receipt = await txResponse.wait();
    let contractAddr = getEventVar(receipt, DEPLOYED_STATIC, CONTRACT_ADDR);
    // await showState(`Subtask deployStatic, ${taskArgs.contractName}, contract at ${contractAddr}, gas: ${receipt.gasUsed}`);
    let outputData: MetaContractData = {
      metaAddress: contractAddr,
      salt: Salt,
      templateName: taskArgs.contractName,
      templateAddress: tmplAddress,
      factoryAddress: factory.address,
      gas: receipt.gasUsed.toNumber(),
      receipt: receipt,
      initCallData: initCallData,
    };
    await updateMetaList(network, outputData);
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
    "salt",
    "unique salt for specifying proxy defaults to salt specified in logic contract"
  )
  .setAction(async (taskArgs, hre) => {
    let network = hre.network.name;
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    let factoryAddress = await getFactoryAddress(network, taskArgs);
    let factory = factoryBase.attach(factoryAddress);
    let logicFactory = await hre.ethers.getContractFactory(taskArgs.contractName);
    let initCallData =
      taskArgs.initCallData === undefined ? "0x" : logicFactory.interface.encodeFunctionData(INITIALIZER, taskArgs.initCallData.split(", "));
    //factory interface pointed to deployed factory contract
    //get the 32byte salt from logic contract file
    let salt: BytesLike = taskArgs.salt === undefined ? await getBytes32Salt(taskArgs.contractName, hre) : hre.ethers.utils.formatBytes32String(taskArgs.salt);
    //encode the deployProxy function call with Salt as arg
    let deployProxy: BytesLike = factoryBase.interface.encodeFunctionData(
      DEPLOY_PROXY,
      [salt]
    );
    //encode upgrade proxy multicall
    let upgradeProxy: BytesLike = factoryBase.interface.encodeFunctionData(
      UPGRADE_PROXY,
      [salt, taskArgs.logicAddress, initCallData]
    );
    //get the multi call arguements as [deployProxy, upgradeProxy]
    let multiCallArgs = [deployProxy, upgradeProxy];
    //send the multicall transaction with deployProxy and upgradeProxy
    let txResponse = await factory.multiCall(multiCallArgs);
    let receipt = await txResponse.wait();
    //Data to return to the main task
    let proxyData: ProxyData = {
      factoryAddress: factoryAddress,
      logicName: taskArgs.contractName,
      logicAddress: taskArgs.logicAddress,
      salt: salt,
      proxyAddress: getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR),
      gas: receipt.gasUsed.toNumber(),
      receipt: receipt,
      initCallData: initCallData,
    };
    await showState(
      `Deployed ${proxyData.logicName} with proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`
    );
    updateProxyList(network, proxyData);
    return proxyData;
  });

task(
  "multiCallUpgradeProxy",
  "multi call to deploy logic and upgrade proxy through factory"
)
  .addParam("contractName", "logic contract name")
  .addOptionalParam("factoryAddress", "factory deploying the contract")
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
    let network = hre.network.name;
    let factoryAddress = await getFactoryAddress(network, taskArgs);
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    const factory = factoryBase.attach(factoryAddress);
    const logicFactory: ContractFactory = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    let initCallData =
      taskArgs.initCallData === undefined ? "0x" : logicFactory.interface.encodeFunctionData(INITIALIZER, taskArgs.initCallData.split(", "));
      
    let deployTx = logicFactory.getDeployTransaction(
      ...taskArgs.constructorArgs
    );
    let deployCreate = factoryBase.interface.encodeFunctionData(DEPLOY_CREATE, [
      deployTx.data,
    ]);
    let salt: string =
      taskArgs.salt === undefined
        ? await getBytes32Salt(taskArgs.contractName, hre)
        : hre.ethers.utils.formatBytes32String(taskArgs.salt);
    let txCount = await hre.ethers.provider.getTransactionCount(
      factory.address
    );
    let implAddress = hre.ethers.utils.getContractAddress({
      from: factory.address,
      nonce: txCount,
    });
    let upgradeProxy = factoryBase.interface.encodeFunctionData(DEPLOY_CREATE, [
      salt,
      implAddress,
      initCallData,
    ]);
    const PROXY_FACTORY = await hre.ethers.getContractFactory(PROXY);
    let proxyAddress = getMetamorphicAddress(factoryAddress, salt, hre);
    let proxyContract = await PROXY_FACTORY.attach(proxyAddress);
    let oldImpl = await proxyContract.getImplementationAddress();
    let txResponse = await factory.multiCall([deployCreate, upgradeProxy]);
    let receipt = await txResponse.wait();
    await showState(
      `Updating logic for the ${taskArgs.contractName} proxy at ${proxyAddress} from ${oldImpl} to ${implAddress}, gasCost: ${receipt.gasUsed}`
    );
    let proxyData: ProxyData = {
      factoryAddress: factoryAddress,
      logicName: taskArgs.contractName,
      logicAddress: taskArgs.logicAddress,
      salt: salt,
      proxyAddress: getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR),
      gas: receipt.gasUsed.toNumber(),
      receipt: receipt,
      initCallData: initCallData,
    };
    return proxyData;
  });

async function getFactoryAddress(network: string, taskArgs: any) {
  //get Factory data from factoryConfig.json
  let path = process.env.test === "true" ? FACTORY_STATE_PATH.split("factoryState")[0] + "testFactoryState" : FACTORY_STATE_PATH
  const factoryConfig = await readFactoryState(path);
  const configFactoryAddress = factoryConfig[network].defaultFactoryAddress;
  let cliFactoryAddress = taskArgs.factoryAddress;
  //object to store data to update config var
  let factoryAddress: string;
  //check if the user provided factory Data in call
  if (cliFactoryAddress !== undefined) {
    factoryAddress = cliFactoryAddress;
  }
  //if the user did not provide factory data check for factory data in factoryConfig
  else if (configFactoryAddress !== undefined) {
    factoryAddress = configFactoryAddress;
  }
  //if no factoryData provided in call and in userConfig throw error
  else {
    throw new Error(
      "Insufficient Factory Data: specify factory name and address in call or HardhatUserConfig.defaultFactory"
    );
  }
  return factoryAddress;
}

/**
 * @description function used to get the arguements for deploy proxy upgrade proxy contract multicall
 * @param taskArgs.factoryName name of factory contract used
 * @param Salt Salt specified in the logic contract file
 * @param taskArgs.logicAddress address of the logic contract
 * @param hre hardhat runtime environment
 * @returns array of solidity function calls
 */
async function getProxyMultiCallArgs(
  Salt: BytesLike,
  taskArgs: any,
  hre: HardhatRuntimeEnvironment
): Promise<BytesLike[]> {
  //get factory object from ethers for encoding function calls
  let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
  let initCallData: BytesLike =
    taskArgs.initCallData === undefined ? "0x" : taskArgs.initCallData;
  //encode the deployProxy function call with Salt as arg
  let deployProxy: BytesLike = factoryBase.interface.encodeFunctionData(
    DEPLOY_PROXY,
    [Salt]
  );
  //encode upgrade proxy multicall
  let upgradeProxy: BytesLike = factoryBase.interface.encodeFunctionData(
    UPGRADE_PROXY,
    [Salt, taskArgs.logicAddress, initCallData]
  );
  return [deployProxy, upgradeProxy];
}

//0x39cab472536e617073686f74730000000000000000000000000000000000000000000000;

/**
 * @description parses config and task args for deployTemplate subtask call args
 * @param taskArgs arguements provided to the task
 * @returns object with call data for deployTemplate subtask
 */
async function getDeployTemplateArgs(network: string, taskArgs: any) {
  let factoryAddress = await getFactoryAddress(network, taskArgs);
  return <DeployArgs>{
    contractName: taskArgs.contractName,
    factoryAddress: factoryAddress,
    constructorArgs: taskArgs?.constructorArgs,
  };
}

function getConstructorArgCount(contract: any) {
  for (let funcObj of contract.abi) {
    if (funcObj.type === "constructor") {
      return funcObj.inputs.length;
    }
  }
  return 0;
}

async function getAccounts(hre: HardhatRuntimeEnvironment) {
  let signers = await hre.ethers.getSigners();
  let accounts: string[] = [];
  for (let signer of signers) {
    accounts.push(signer.address);
  }
  return accounts;
}

async function getFullyQualifiedName(
  contractName: string,
  hre: HardhatRuntimeEnvironment
) {
  let artifactPaths = await hre.artifacts.getAllFullyQualifiedNames();
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
    let events = receipt.events;
    for (let i = 0; i < events.length; i++) {
      //look for the event
      if (events[i].event === eventName) {
        if (events[i].args !== undefined) {
          let args = events[i].args;
          //extract the deployed mock logic contract address from the event
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
  let qualifiedName: any = await getFullyQualifiedName(contractName, hre);
  let buildInfo = await hre.artifacts.getBuildInfo(qualifiedName);
  let contractOutput: any;
  let devdoc: any;
  let salt: string = "";
  if (buildInfo !== undefined) {
    let path = extractPath(qualifiedName);
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
  let salt: string = await getSalt(contractName, hre);
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
  let initCode = "0x6020363636335afa1536363636515af43d36363e3d36f3";
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
