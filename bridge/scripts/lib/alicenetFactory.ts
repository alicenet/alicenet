import {
  BytesLike,
  ContractReceipt,
  ContractTransaction,
  Overrides,
} from "ethers";
import { Artifacts, HardhatEthersHelpers } from "hardhat/types";
import { AliceNetFactory } from "../../typechain-types";
import { PromiseOrValue } from "../../typechain-types/common";
import { encodeMultiCallArgs } from "./alicenetTasks";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  DEPLOYED_PROXY,
  DEPLOYED_RAW,
} from "./constants";
type Ethers = typeof import("../../node_modules/ethers/lib/ethers") &
  HardhatEthersHelpers;

export async function deployFactory(
  constructorArgs: string,
  ethers: Ethers,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
): Promise<AliceNetFactory> {
  const factoryBase = await ethers.getContractFactory(ALICENET_FACTORY);
  return factoryBase.deploy(constructorArgs, overrides);
}

export async function deployUpgradeable(
  contractName: string,
  factoryAddress: string,
  ethers: Ethers,
  artifacts: Artifacts,
  constructorArgs: Array<string>
) {
  const factory = await ethers.getContractAt(ALICENET_FACTORY, factoryAddress);
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
    const proxySalt = await getSalt(contractName, artifacts, ethers);
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
        res.logicAddress,
        ethers
      );
      txResponse = await factory.multiCall(multiCallArgs);
      receipt = await txResponse.wait();
      res.proxyAddress = getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR);
      return res;
    } else {
      console.error(`${contractName} contract missing salt`);
    }
    return res;
  } else {
    throw new Error(`failed to get contract bytecode for ${contractName}`);
  }
}

export async function deployCreateAndRegister(
  contractName: string,
  factoryAddress: string,
  ethers: Ethers,
  constructorArgs: any[],
  overrides?: Overrides & { from?: PromiseOrValue<string> }
): Promise<ContractTransaction> {
  // get a factory instance connected to the factory a
  const factory = await ethers.getContractAt(ALICENET_FACTORY, factoryAddress);
  const logicContract: any = await ethers.getContractFactory(contractName);
  // if not constructor ars is provide and empty array is used to indicate no constructor args
  // encode deployBcode,
  const deployTxData = logicContract.getDeployTransaction(...constructorArgs)
    .data as BytesLike;
  const contractNameByts32 = ethers.utils.formatBytes32String(contractName);
  if (overrides !== undefined) {
    return factory.deployCreateAndRegister(
      deployTxData,
      contractNameByts32,
      overrides
    );
  } else {
    return factory.deployCreateAndRegister(deployTxData, contractNameByts32);
  }
}

export async function upgradeProxy(
  contractName: string,
  factoryAddress: string,
  ethers: Ethers,
  artifacts: Artifacts,
  constructorArgs: string[]
) {
  const factoryBase = await ethers.getContractFactory(ALICENET_FACTORY);
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
    logicAddress: await getEventVar(receipt, DEPLOYED_RAW, CONTRACT_ADDR),
    proxySalt: await getSalt(contractName, artifacts, ethers),
  };
  // upgrade the proxy
  await factory.upgradeProxy(
    res.proxySalt as BytesLike,
    res.logicAddress,
    "0x"
  );
  return res;
}

async function getFullyQualifiedName(
  contractName: string,
  artifacts: Artifacts
) {
  const artifactPaths = await artifacts.getAllFullyQualifiedNames();
  for (let i = 0; i < artifactPaths.length; i++) {
    if (artifactPaths[i].split(":")[1] === contractName) {
      return String(artifactPaths[i]);
    }
  }
  return undefined;
}

/**
 * @description returns everything on the left side of the :
 * ie: src/proxy/Proxy.sol:Mock => src/proxy/Proxy.sol
 * @param qualifiedName the relative path of the contract file + ":" + name of contract
 * @returns the relative path of the contract
 */
export function extractPath(qualifiedName: string) {
  return qualifiedName.split(":")[0];
}

async function getDeployUpgradeableMultiCallArgs(
  factoryAddress: string,
  Salt: BytesLike,
  logicAddress: BytesLike,
  ethers: Ethers,
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
/**
 * @description gets the salt specified in the contracts head with a custom natspec tag @custom:salt
 * @param contractName name of the contract
 * @param artifacts artifacts object from hardhat artifacts
 * @param ethers ethersjs object
 * @returns bytes32 formatted salt
 */
export async function getSalt(
  contractName: string,
  artifacts: Artifacts,
  ethers: Ethers
): Promise<string> {
  const qualifiedName: any = await getFullyQualifiedName(
    contractName,
    artifacts
  );
  const buildInfo = await artifacts.getBuildInfo(qualifiedName);
  if (buildInfo === undefined) {
    throw new Error("Missing custom:salt");
  }
  const path = extractPath(qualifiedName);
  const contractOutput: any = buildInfo?.output.contracts[path][contractName];
  const devdoc: any = contractOutput.devdoc;
  const salt = devdoc["custom:salt"];
  return ethers.utils.formatBytes32String(salt);
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
 *
 * @param factoryAddress address of the factory that deployed the contract
 * @param salt value specified by custom:salt in the contrac
 * @param ethers ethersjs object
 * @returns returns the address of the metamorphic contract deployed with the following metamorphic code "0x6020363636335afa1536363636515af43d36363e3d36f3"
 */
export function getMetamorphicAddress(
  factoryAddress: string,
  salt: string,
  ethers: Ethers
) {
  const initCode = "0x6020363636335afa1536363636515af43d36363e3d36f3";
  return ethers.utils.getCreate2Address(
    factoryAddress,
    salt,
    ethers.utils.keccak256(initCode)
  );
}
