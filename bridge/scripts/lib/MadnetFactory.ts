//separate deploy of template from deploy of deterministic address

//import { string } from "hardhat/internal/core/params/argumentTypes";

//assume you have to divide the transaction
//estimate gas, observe gas limit,

//return all addresses

//return the logs
import { BytesLike, ContractReceipt } from "ethers";
import { ethers, artifacts } from "hardhat";
import { MadnetFactory } from "../../typechain-types";
import { CONTRACT_ADDR, DEPLOYED_RAW } from "./constants";
const defaultFactoryName = "MadnetFactory";
const DeployedRawEvent = "DeployedRaw";
const contractAddrVar = "contractAddr";
const DeployedProxyEvent = "DeployedProxy";
const logicAddrKey = "LogicAddress";
const ProxyAddrKey = "ProxyAddress";
const deployedStaticEvent = "DeployedStatic";
const deployedTemplateEvent = "DeployedTemplate";
const MetaAddrKey = "MetaAddress";
const templateAddrKey = "TemplateAddress";

export async function deployUpgradeable(
  contractName: string,
  factoryAddress: string,
  constructorArgs: Array<string>
) {
  let MadnetFactory = await ethers.getContractFactory(defaultFactoryName);
  let factory = await MadnetFactory.attach(factoryAddress);
  //get an instance of the logic contract interface
  let logicFactory = await ethers.getContractFactory(contractName);
  //get the deployment bytecode from the interface
  let deployTxReq = await logicFactory.getDeployTransaction(...constructorArgs);
  let deployBytecode = deployTxReq.data;
  if(deployBytecode !== undefined){
    //deploy the bytecode using the factory
    let txResponse = await factory.deployCreate(deployBytecode);
    let receipt = await txResponse.wait();
    let proxySalt = await getSalt(contractName);
    let res = <
      {
        logicAddress: string;
        proxyAddress: string;
        proxySalt: string;
      }
      >{
        logicAddress: getEventVar(receipt, DEPLOYED_RAW, CONTRACT_ADDR),
        proxySalt: proxySalt,
      };
    if (proxySalt !== undefined) {
      //multicall deployProxy. upgradeProxy
      let multiCallArgs = await getDeployUpgradeableMultiCallArgs(
        defaultFactoryName,
        res.proxySalt,
        res.logicAddress
      );
      txResponse = await factory.multiCall(multiCallArgs);
      receipt = await txResponse.wait();
      res.proxyAddress = await getEventVar(
        receipt,
        DeployedProxyEvent,
        contractAddrVar
      );
      return res;
    } else {
      console.error(`${contractName} contract missing salt`);
    }
    return res;
  } else {
    throw new Error(`failed to get contract bytecode for ${contractName}`);
  }
}

export async function upgradeProxy(
  contractName: string,
  factoryAddress: string,
  constructorArgs?: string[]
) {
  let factoryBase = await ethers.getContractFactory(defaultFactoryName);
  let factory = factoryBase.attach(factoryAddress);
  let logicContractFactory = await ethers.getContractFactory(contractName);
  let deployBCode: BytesLike;
  if (typeof constructorArgs !== "undefined" && constructorArgs.length >= 0) {
    deployBCode = logicContractFactory.getDeployTransaction(
      ...constructorArgs
    ).data as BytesLike;
  } else {
    deployBCode = logicContractFactory.getDeployTransaction()
      .data as BytesLike;
  }
  //instantiate the return object
  let txResponse = await factory.deployCreate(deployBCode);
  let receipt = await txResponse.wait();
  let res = {
    logicAddress: await getEventVar(
      receipt,
      DeployedRawEvent,
      contractAddrVar
    ),
    proxySalt: await getSalt(contractName),
  };
  //upgrade the proxy
  await factory.upgradeProxy(
    res.proxySalt as BytesLike,
    res.logicAddress,
    "0x"
  );
  return res;
}

export async function deployStatic(
  contractName: string,
  factoryAddress: string
) {
  let MadnetFactory = await ethers.getContractFactory(defaultFactoryName);
  let logicContract = await ethers.getContractFactory(contractName);
  let factory: MadnetFactory = MadnetFactory.attach(factoryAddress);
  let deployBCode = logicContract.bytecode;
  let txResponse = await factory.deployTemplate(deployBCode);
  let receipt = await txResponse.wait();
  let templateAddress: BytesLike = getEventVar(
    receipt,
    deployedTemplateEvent,
    contractAddrVar
  );
  let metaSalt = await getSalt(contractName);
  if (typeof metaSalt === "undefined") {
    throw "Couldn't get the salt for:" + contractName;
  }
  txResponse = await factory.deployStatic(metaSalt, "0x");
  receipt = await txResponse.wait();
  let metaAddress: string = getEventVar(
    receipt,
    deployedStaticEvent,
    contractAddrVar
  );
  return {
    templateAddress,
    metaSalt,
    metaAddress,
  };
}

async function getFullyQualifiedName(contractName: string) {
  let artifactPaths = await artifacts.getAllFullyQualifiedNames();
  for (let i = 0; i < artifactPaths.length; i++) {
    if (artifactPaths[i].split(":")[1] === contractName) {
      return String(artifactPaths[i]);
    }
  }
  return undefined;
}

function extractPath(qualifiedName: string) {
  return qualifiedName.split(":")[0];
}



async function getDeployUpgradeableMultiCallArgs(
  factoryName: string,
  Salt: BytesLike,
  logicAddress: BytesLike,
  initCallData?: BytesLike,
) {
  let factoryBase = await ethers.getContractFactory(factoryName);
  let deployProxy:BytesLike = factoryBase.interface.encodeFunctionData("deployProxy", [Salt]);
  let upgradeProxy: BytesLike = initCallData !== undefined ? 
    factoryBase.interface.encodeFunctionData("upgradeProxy", [Salt, logicAddress, initCallData]) 
    : factoryBase.interface.encodeFunctionData("upgradeProxy", [Salt, logicAddress, "0x"]); 
  
  return [deployProxy, upgradeProxy];
}

async function getSalt(contractName: string) {
  let qualifiedName: any = await getFullyQualifiedName(contractName);
  let buildInfo = await artifacts.getBuildInfo(qualifiedName);
  let contractOutput: any;
  let devdoc: any;
  let salt;
  if (buildInfo !== undefined) {
    let path = extractPath(qualifiedName);
    contractOutput = buildInfo?.output.contracts[path][contractName];
    devdoc = contractOutput.devdoc;
    salt = devdoc["custom:salt"];
    if (salt !== undefined && salt !== "") {
      return ethers.utils.formatBytes32String(salt);
    }
  } else {
    console.error("Missing custom:salt");
  }
}

async function getDeployTypeWithContractName(contractName: string) {
  let qualifiedName: any = await getFullyQualifiedName(contractName);
  let buildInfo = await artifacts.getBuildInfo(qualifiedName);
  let deployType: any;
  if (buildInfo !== undefined) {
    let path = extractPath(qualifiedName);
    deployType = buildInfo?.output.contracts[path][contractName];
  }
  return deployType.devdoc["custom:deploy-type"];
}

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

