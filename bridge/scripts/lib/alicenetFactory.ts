// separate deploy of template from deploy of deterministic address

// import { string } from "hardhat/internal/core/params/argumentTypes";

// assume you have to divide the transaction
// estimate gas, observe gas limit,

// return all addresses

// return the logs
import { BytesLike, ContractReceipt } from "ethers";
import { artifacts, ethers } from "hardhat";
import { AliceNetFactory } from "../../typechain-types";
import { encodeMultiCallArgs } from "./alicenetTasks";
import { ALICENET_FACTORY, CONTRACT_ADDR, DEPLOYED_RAW } from "./constants";
const defaultFactoryName = "AliceNetFactory";
const DeployedRawEvent = "DeployedRaw";
const contractAddrVar = "contractAddr";
const DeployedProxyEvent = "DeployedProxy";
const deployedStaticEvent = "DeployedStatic";
const deployedTemplateEvent = "DeployedTemplate";
export async function deployUpgradeable(
  contractName: string,
  factoryAddress: string,
  constructorArgs: Array<string>
) {
  const factory = await ethers.getContractAt(
    defaultFactoryName,
    factoryAddress
  );
  // get an instance of the logic contract interface
  const logicFactory = await ethers.getContractFactory(contractName);
  // get the deployment bytecode from the interface
  const deployTxReq = await logicFactory.getDeployTransaction(
    ...constructorArgs
  );
  const deployBytecode = deployTxReq.data;
  if (deployBytecode !== undefined) {
    // deploy the bytecode using the factory
    let txResponse = await factory.deployCreate(deployBytecode);
    let receipt = await txResponse.wait();
    const proxySalt = await getSalt(contractName);
    const res = <
      {
        logicAddress: string;
        proxyAddress: string;
        proxySalt: string;
      }
    >{
      logicAddress: getEventVar(receipt, DEPLOYED_RAW, CONTRACT_ADDR),
      proxySalt,
    };
    if (proxySalt !== undefined) {
      // multicall deployProxy. upgradeProxy
      const multiCallArgs = await getDeployUpgradeableMultiCallArgs(
        factory.address,
        res.proxySalt,
        res.logicAddress
      );
      txResponse = await factory.multiCall(multiCallArgs);
      receipt = await txResponse.wait();
      res.proxyAddress = getEventVar(
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
  const factoryBase = await ethers.getContractFactory(defaultFactoryName);
  const factory = factoryBase.attach(factoryAddress);
  const logicContractFactory = await ethers.getContractFactory(contractName);
  let deployBCode: BytesLike;
  if (typeof constructorArgs !== "undefined" && constructorArgs.length >= 0) {
    deployBCode = logicContractFactory.getDeployTransaction(...constructorArgs)
      .data as BytesLike;
  } else {
    deployBCode = logicContractFactory.getDeployTransaction().data as BytesLike;
  }
  // instantiate the return object
  const txResponse = await factory.deployCreate(deployBCode);
  const receipt = await txResponse.wait();
  const res = {
    logicAddress: await getEventVar(receipt, DeployedRawEvent, contractAddrVar),
    proxySalt: await getSalt(contractName),
  };
  // upgrade the proxy
  await factory.upgradeProxy(
    res.proxySalt as BytesLike,
    res.logicAddress,
    "0x"
  );
  return res;
}

async function getFullyQualifiedName(contractName: string) {
  const artifactPaths = await artifacts.getAllFullyQualifiedNames();
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
  factoryAddress: string,
  Salt: BytesLike,
  logicAddress: BytesLike,
  initCallData?: BytesLike
) {
  const factoryBase = await ethers.getContractFactory(ALICENET_FACTORY);
  const deployProxy: BytesLike = factoryBase.interface.encodeFunctionData(
    "deployProxy",
    [Salt]
  );
  const upgradeProxy: BytesLike =
    initCallData !== undefined
      ? factoryBase.interface.encodeFunctionData("upgradeProxy", [
          Salt,
          logicAddress,
          initCallData,
        ])
      : factoryBase.interface.encodeFunctionData("upgradeProxy", [
          Salt,
          logicAddress,
          "0x",
        ]);

  return [
    encodeMultiCallArgs(factoryAddress, 0, deployProxy),
    encodeMultiCallArgs(factoryAddress, 0, upgradeProxy),
  ];
}

export async function getSalt(contractName: string) {
  const qualifiedName: any = await getFullyQualifiedName(contractName);
  const buildInfo = await artifacts.getBuildInfo(qualifiedName);
  let contractOutput: any;
  let devdoc: any;
  let salt;
  if (buildInfo !== undefined) {
    const path = extractPath(qualifiedName);
    contractOutput = buildInfo?.output.contracts[path][contractName];
    devdoc = contractOutput.devdoc;
    salt = devdoc["custom:salt"];
    if (salt !== undefined && salt !== "") {
      return ethers.utils.formatBytes32String(salt);
    }
  } else {
    throw new Error("Missing custom:salt");
  }
}

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
