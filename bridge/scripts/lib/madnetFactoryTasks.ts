import {
  DEPLOYED_PROXY,
  DEPLOYED_RAW,
  CONTRACT_ADDR,
  MADNET_FACTORY,
  DEPLOY_CREATE,
  DEPLOYED_STATIC,
  PROXY,
} from "./constants";
import { task } from "hardhat/config";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import {
  DeployCreateData,
  FactoryConfig,
  FactoryData,
  MetaContractData,
  ProxyData,
  readFactoryStateData,
  TemplateData,
  updateDefaultFactoryData,
  updateDeployCreateList,
  updateMetaList,
  updateProxyList,
  updateTemplateList,
} from "./factoryStateUtils";
import { BytesLike, ContractFactory, ContractReceipt } from "ethers";

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
}

task("getSalt", "gets salt from contract")
  .addParam("contractName", "test contract")
  .setAction(async (taskArgs, hre) => {
    let salt = await getSalt(taskArgs.contractName, hre);
    await showState(salt)
  });

task("getBytes32Salt", "gets the bytes32 version of salt from contract")
  .addParam("contractName", "test contract")
  .setAction(async (taskArgs, hre) => {
    let salt = await getBytes32Salt(taskArgs.contractName, hre);
    await showState(salt)
  });

task(
  "deployFactory",
  "Deploys an instance of a factory contract specified by its name"
)
  .setAction(async (taskArgs, hre) => {

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
    //record the data in a json file to be used in other tasks
    let factoryData: FactoryData = {
      address: factory.address,
      gas: gasCost.toNumber(),
    };
    await updateDefaultFactoryData(factoryData);
    await showState(`Deployed: ${MADNET_FACTORY}, at address: ${factory.address}`)
    return factory.address;
  });

task(
  "deployUpgradeableProxy",
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
    "initialization call data for initializable contracts"
  )
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguments for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    let factoryAddress = await getFactoryAddress(taskArgs);
    //uses the factory Data and logic contractName and returns deploybytecode and any constructor args attached
    let callArgs: DeployArgs = {
      contractName: taskArgs.contractName,
      factoryAddress: factoryAddress,
      constructorArgs: taskArgs.constructorArgs,
    };
    //deploy create the logic contract
    let result: DeployCreateData = await hre.run("deployCreate", callArgs);
    let initCallData = taskArgs.initCallData === undefined ? "0x" : taskArgs.initCallData;
    let mcCallArgs: DeployProxyMCArgs = {
      contractName: taskArgs.contractName,
      logicAddress: result.address,
      initCallData: initCallData,
    };

    let proxyData:ProxyData = await hre.run("multiCallDeployProxy", mcCallArgs);
    return proxyData;
  });

task(
  "deployMetamorphic",
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
    "call data used to initialize initializable contracts"
  )
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguements for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    let factoryData = await getFactoryAddress(taskArgs);
    //uses the factory Data and logic contractName and returns deploybytecode and any constructor args attached
    let callArgs: DeployArgs = await getDeployTemplateArgs(taskArgs);
    //deploy create the logic contract
    let templateAddress = await hre.run("deployTemplate", callArgs);
    callArgs = await getDeployStaticSubtaskArgs(taskArgs);
    let metaContractData = await hre.run("deployStatic", callArgs);
    await showState(`Deployed Metamorphic for ${taskArgs.contractName} at: ${metaContractData.metaAddress}, with logic from, ${metaContractData.templateAddress}, gas used: ${metaContractData.gas}`)
    return metaContractData;
  });

//factoryName param doesnt do anything right now
task("deployCreate", "deploys a contract from the factory using create")
  .addParam("contractName", "logic contract name")
  .addOptionalParam("factoryAddress", "factory deploying the contract")
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "array that holds all arguments for constructor"
  )
  .setAction(async (taskArgs, hre) => {
    let factoryAddress = await getFactoryAddress(taskArgs);
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    //get a factory instance connected to the factory a
    const factory = factoryBase.attach(factoryAddress);
    let logicContract: ContractFactory = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    let constructorArgs = taskArgs.constructor === undefined ? [] : taskArgs.constructorArgs
    //encode deployBcode
    let deployTx = logicContract.getDeployTransaction(...constructorArgs);
    if (deployTx.data !== undefined) {
      let txResponse = await factory.deployCreate(deployTx.data);
      let receipt = await txResponse.wait();
      let deployCreateData: DeployCreateData = {
        name: taskArgs.contractName,
        address: getEventVar(receipt, DEPLOYED_RAW, CONTRACT_ADDR),
        factoryAddress: factoryAddress,
        gas: receipt.gasUsed.toNumber(),
        constructorArgs: taskArgs?.constructorArgs,
      };
      await updateDeployCreateList(deployCreateData);
      await showState(`[DEBUG ONLY, DONT USE THIS ADDRESS IN THE SIDE CHAIN, USE THE PROXY INSTEAD!] Deployed logic for ${taskArgs.contractName} contract at: ${deployCreateData.address}, gas: ${receipt.gasUsed}`)
      return deployCreateData;
    } else {
      throw new Error(`failed to get deployment bytecode for ${taskArgs.contractName}`)
    }
  });

task("upgradeDeployedProxy", "deploys a contract from the factory using create")
  .addParam("contractName", "logic contract name")
  .addParam("factoryName", "Name of the factory contract")
  .addParam("factoryAddress", "factory deploying the contract")
  .addParam(
    "logicAddress",
    "address of the new logic contract to upgrade the proxy to"
  )
  .addOptionalParam("initCallData", "data used to initialize initializable contracts")
  .setAction(async (taskArgs, hre) => {
    let factoryAddress = await getFactoryAddress(taskArgs);
    let MadnetFactory = await hre.ethers.getContractFactory(MADNET_FACTORY);
    //grab the salt from the logic contract
    let Salt = await getBytes32Salt(taskArgs.contractName, hre);
    //get logic contract interface
    let logicContract: any = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    const factory = await MadnetFactory.attach(factoryAddress);
    let initCallData: string = taskArgs.initCallData === undefined ? "0x" : taskArgs.initCallData;
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
    await updateProxyList(proxyData);
    return proxyData;
  });

task(
  "deployTemplate",
  "deploys a template contract with the universal code copy constructor that deploys"
)
  .addParam("contractName", "logic contract name")
  .addOptionalParam("factoryAddress", "optional factory address, defaults to config address")
  .addOptionalVariadicPositionalParam(
    "constructorArgs",
    "input constructor args at the end of call"
  )
  .setAction(async (taskArgs, hre) => {
    let factoryAddress = await getFactoryAddress(taskArgs);
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY)
    let logicContract: ContractFactory = await hre.ethers.getContractFactory(
      taskArgs.contractName
    );
    let deployTxReq =
      taskArgs.constructorArgs === undefined ?
        logicContract.getDeployTransaction() : logicContract.getDeployTransaction(...taskArgs.constructorArgs);
    if (deployTxReq.data !== undefined) {
      let deployBytecode = deployTxReq.data;
      //get a factory instance connected to the factory addr
      const factory = factoryBase.attach(factoryAddress);
      if (hre.network.name === "hardhat"){
        // hardhat is not being able to estimate correctly the tx gas due to the massive bytes array
        // being sent as input to the function (the contract bytecode), so we need to increase the block
        // gas limit temporally in order to deploy the template
        await hre.network.provider.send("evm_setBlockGasLimit", ["0x3000000000000000"]);
      }
      let txResponse = await factory.deployTemplate(deployBytecode);
      let receipt = await txResponse.wait();
      let templateData: TemplateData = {
        name: taskArgs.contractName,
        address: getEventVar(receipt, "DeployedTemplate", CONTRACT_ADDR),
        factoryAddress: factoryAddress,
        gas: receipt.gasUsed.toNumber(),
      };
      if (taskArgs.constructorArgs !== undefined) {
        templateData.constructorArgs = taskArgs.constructorArgs;
      }
    //   await showState(`Subtask deployedTemplate for ${taskArgs.contractName} contract at ${templateData.address}, gas: ${receipt.gasUsed}`);
      updateTemplateList(templateData);
      return templateData;
    } else {
      throw new Error(`failed to get contract bytecode for ${taskArgs.contractName}`);
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
    "call data used to initialize initializable contracts"
  )
  .setAction(async (taskArgs, hre) => {
    let factoryAddress = await getFactoryAddress(taskArgs);
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY)
    let initCallData = taskArgs.initCallData === undefined ? "0x" : taskArgs.initCallData;
    let Salt = await getBytes32Salt(taskArgs.contractName, hre);
    //get a factory instance connected to the factory addr
    const factory = factoryBase.attach(factoryAddress);
    //TODO: Reconsider doing this, might get the wrong implementation address
    let tmplAddress = await factory.callStatic.getImplementation();
    let txResponse = await factory.deployStatic(Salt, initCallData);
    let receipt = await txResponse.wait();
    let contractAddr = getEventVar(
      receipt,
      DEPLOYED_STATIC,
      CONTRACT_ADDR
    );
    // await showState(`Subtask deployStatic, ${taskArgs.contractName}, contract at ${contractAddr}, gas: ${receipt.gasUsed}`);
    let outputData: MetaContractData = {
      metaAddress: contractAddr,
      salt: Salt,
      templateName: taskArgs.contractName,
      templateAddress: tmplAddress,
      factoryAddress: factory.address,
      gas: receipt.gasUsed.toNumber(),
      initCallData: initCallData,
    };
    await updateMetaList(outputData);
    return outputData;
  });

/**
 * deploys a proxy and upgrades it using multicall from factory
 * @returns a proxyData object with logic contract name, address and proxy salt, and address.
 */
task("multiCallDeployProxy", "deploy and upgrade proxy with multicall")
  .addParam(
    "logicAddress",
    "Address of the logic contract to point the proxy to"
  )
  .addParam("contractName", "logic contract name")
  .addOptionalParam("factoryAddress", "factory deploying the contract")
  .addOptionalParam(
    "initCallData",
    "call data used to initialize initializable contracts"
  )
  .addOptionalParam("salt", "unique salt for specifying proxy defaults to salt specified in logic contract")
  .setAction(async (taskArgs, hre) => {
    let gas = 0;
    let factoryAddress = await getFactoryAddress(taskArgs);
    let initCallData = taskArgs.initCallData === undefined ? "0x" : taskArgs.initCallData;
    let factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    //factory interface pointed to deployed factory contract
    let factory = factoryBase.attach(factoryAddress);
    //get the 32byte salt from logic contract file
    let salt: BytesLike = taskArgs.salt === undefined ? await getBytes32Salt(taskArgs.contractName, hre) : hre.ethers.utils.formatBytes32String(taskArgs.salt);
    //get the multi call arguements as [deployProxy, upgradeProxy]
    let multiCallArgs = await getProxyMultiCallArgs(salt, taskArgs, hre);
    //send the multicall transaction with deployProxy and upgradeProxy
    let txResponse = await factory.multiCall(multiCallArgs);
    let receipt = await txResponse.wait();
    //Data to return to the main task
    let proxyData: ProxyData = {
      factoryAddress: factoryAddress,
      logicName: taskArgs.contractName,
      logicAddress: taskArgs.logicAddress,
      salt: salt,
      proxyAddress: getEventVar(
        receipt,
        DEPLOYED_PROXY,
        CONTRACT_ADDR
      ),
      gas: receipt.gasUsed.toNumber(),
      initCallData: initCallData,
    };
    await showState(`Deployed ${proxyData.logicName} with proxy at ${proxyData.proxyAddress}, gasCost: ${proxyData.gas}`);
    updateProxyList(proxyData);
    return proxyData;
  });

task("multiCallUpgradeProxy", "multi call to deploy logic and upgrade proxy through factory")
  .addParam("contractName", "logic contract name")
  .addOptionalParam("initCallData", "abi encoded initialization data with selector")
  .addOptionalParam("salt", "unique salt for specifying proxy defaults to salt specified in logic contract")
  .addOptionalVariadicPositionalParam("constructorArgs")
  .setAction(async (taskArgs, hre) => {
    let factoryAddress = await getFactoryAddress(taskArgs);
    const factoryBase = await hre.ethers.getContractFactory(MADNET_FACTORY);
    const factory = factoryBase.attach(factoryAddress)
    const logicFactory: ContractFactory = await hre.ethers.getContractFactory(taskArgs.contractName)
    let initCallData = taskArgs.initCallData === undefined ? "0x" : taskArgs.initCallData;
    let deployTx = logicFactory.getDeployTransaction(...taskArgs.constructorArgs);
    let deployCreate = factoryBase.interface.encodeFunctionData(DEPLOY_CREATE, [deployTx.data]);
    let salt: string = taskArgs.salt === undefined ? await getBytes32Salt(taskArgs.contractName, hre) : hre.ethers.utils.formatBytes32String(taskArgs.salt);;
    let txCount = await hre.ethers.provider.getTransactionCount(factory.address)
    let implAddress = hre.ethers.utils.getContractAddress({ from: factory.address, nonce: txCount });
    let upgradeProxy = factoryBase.interface.encodeFunctionData(DEPLOY_CREATE, [salt, implAddress, initCallData]);
    const PROXY_FACTORY = await hre.ethers.getContractFactory(PROXY)
    let proxyAddress = getMetamorphicAddress(factoryAddress, salt, hre)
    let proxyContract = await PROXY_FACTORY.attach(proxyAddress)
    let oldImpl = await proxyContract.getImplementationAddress()
    let txResponse = await factory.multiCall([deployCreate, upgradeProxy]);
    let receipt = await txResponse.wait();
    await showState(`Updating logic for the ${taskArgs.contractName} proxy at ${proxyAddress} from ${oldImpl} to ${implAddress}, gasCost: ${receipt.gasUsed}`);
    return receipt
  });

async function getFactoryAddress(taskArgs: any) {
  //get Factory data from factoryConfig.json
  const factoryConfig = (await readFactoryStateData()) as FactoryConfig;
  const configFactoryAddress = factoryConfig.defaultFactoryData.address;
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
  let initCallData: BytesLike = taskArgs.initCallData === undefined ? "0x" : taskArgs.initCallData;
  //encode the deployProxy function call with Salt as arg
  let deployProxy: BytesLike = factoryBase.interface.encodeFunctionData(
    "deployProxy",
    [Salt]
  );
  //encode upgrade proxy multicall
  let upgradeProxy: BytesLike = factoryBase.interface.encodeFunctionData(
    "upgradeProxy",
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
async function getDeployTemplateArgs(taskArgs: any) {
  let factoryAddress = await getFactoryAddress(taskArgs);
  return <DeployArgs>{
    contractName: taskArgs.contractName,
    factoryAddress: factoryAddress,
    constructorArgs: taskArgs?.constructorArgs,
  };
}

/**
 * @description parses config and task args for deployStatic subtask call args
 * @param taskArgs arguements provided to the task
 * @returns object with call data for deployStatic subtask
 */
async function getDeployStaticSubtaskArgs(taskArgs: any) {
  let factoryAddress = await getFactoryAddress(taskArgs);
  let initCallData = taskArgs.initCallData === undefined ? "0x" : taskArgs.initCallData;
  return <DeployArgs>{
    contractName: taskArgs.contractName,
    factoryAddress: factoryAddress,
    initCallData: initCallData,
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
function getEventVar(receipt: ContractReceipt, eventName: string, varName: string) {
  let result = "0x";
  if (receipt.events !== undefined) {
    let events = receipt.events
    for (let i = 0; i < events.length; i++) {
      //look for the event
      if (events[i].event === eventName) {
        if (events[i].args !== undefined) {
          let args = events[i].args
          //extract the deployed mock logic contract address from the event
          result = args !== undefined ? args[varName] : undefined;
          if (result !== undefined) {
            return result;
          }
        } else {
          throw new Error(`failed to extract ${varName} from event: ${eventName}`)
        }
      }
    }
  }
  throw new Error(`failed to find event: ${eventName}`)
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

export const showState = async (message:string):Promise<void> => {
  if (process.env.silencer === undefined || process.env.silencer === "false") {
    console.log(message);
  }
}
