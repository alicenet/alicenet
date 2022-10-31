import {
  BigNumber,
  BigNumberish,
  BytesLike,
  ContractFactory,
  ContractReceipt,
  ContractTransaction,
  Overrides,
} from "ethers";
import { Artifacts, HardhatEthersHelpers } from "hardhat/types";
import { AliceNetFactory } from "../../typechain-types";
import { PromiseOrValue } from "../../typechain-types/common";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  DEPLOYED_RAW,
  MULTICALL_GAS_LIMIT,
} from "./constants";
type Ethers = typeof import("../../node_modules/ethers/lib/ethers") &
  HardhatEthersHelpers;

export type MultiCallArgsStruct = {
  target: string;
  value: BigNumberish;
  data: BytesLike;
};

export async function deployFactory(
  constructorArgs: string,
  ethers: Ethers,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
): Promise<AliceNetFactory> {
  const factoryBase = await ethers.getContractFactory(ALICENET_FACTORY);
  return factoryBase.deploy(constructorArgs, overrides);
}

//deploy upgradeable contract

export async function multiCallDeployUpgradeable(
  implementationBase: ContractFactory,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string,
  constructorArgs: any[] = [],
  salt: string,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
) {
  const multiCallArgs = await encodeMultiCallDeployUpgradeableArgs(
    implementationBase,
    factory,
    ethers,
    initCallData,
    constructorArgs,
    salt
  );
  const estimatedMultiCallGas = await factory.estimateGas.multiCall(
    multiCallArgs
  );
  if (estimatedMultiCallGas.gt(BigNumber.from(MULTICALL_GAS_LIMIT))) {
    throw new Error(
      `estimatedGasCost ${estimatedMultiCallGas.toString()} exceeds MULTICALL_GAS_LIMIT ${MULTICALL_GAS_LIMIT}`
    );
  }
  return factory.multiCall(multiCallArgs, overrides);
}

//upgradeProxy
/**
 * @decription uses alicenet factory multicall to deploy a logic contract with deploycreate,
 * then upgrades the proxy with upgradeProxy.
 * @param contractName the name of the logic contract to deploy
 * @param factory instance of alicenetFactory
 * @param ethers instance of ethers
 * @param initCallData encoded init calldata, 0x if no initializer function
 * @param constructorArgs string array of constructor arguments
 * @param salt bytes32 format of the salt of proxy contract to upgrade,
 * if empty string is passed the contract name will be used
 * @param overrides transaction overrides
 * @returns
 */
export async function multiCallDeployLogicDeployProxyUpgradeProxy(
  contractName: string,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string,
  constructorArgs: any[] = [],
  salt: string,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
): Promise<ContractTransaction> {
  const implementationBase = await ethers.getContractFactory(contractName);
  const multiCallArgs = await encodeUpgradeProxyMultiCallArgs(
    implementationBase,
    factory,
    ethers,
    initCallData,
    constructorArgs,
    salt
  );
  return factory.multiCall(multiCallArgs, overrides);
}

export async function multicallDeployAndUpgradeProxy(
  logicAddress: string,
  salt: string,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string
) {
  const multiCallArgs = await encodeMultiCallDeployAndUpgradeProxyArgs(
    logicAddress,
    factory,
    ethers,
    initCallData,
    salt
  );
  return factory.multiCall(multiCallArgs);
}

export async function factoryMultiCall(
  factory: AliceNetFactory,
  multiCallArgs: MultiCallArgsStruct[],
  overrides?: Overrides & { from?: PromiseOrValue<string> }
) {
  return factory.multiCall(multiCallArgs, overrides);
}

export async function encodeMultiCallDeployAndUpgradeProxyArgs(
  implementationAddress: string,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string,
  salt: string
) {
  const deployProxyCallData: BytesLike = factory.interface.encodeFunctionData(
    "deployProxy",
    [salt]
  );
  const deployProxy = encodeMultiCallArgs(
    factory.address,
    0,
    deployProxyCallData
  );
  const upgradeProxyCallData = factory.interface.encodeFunctionData(
    "upgradeProxy",
    [salt, implementationAddress, initCallData]
  );
  const upgradeProxy = encodeMultiCallArgs(
    factory.address,
    0,
    upgradeProxyCallData
  );
  return [deployProxy, upgradeProxy];
}

export async function encodeMultiCallDeployUpgradeableArgs(
  implementationBase: ContractFactory,
  factory: AliceNetFactory,
  ethers: Ethers,
  initializationCallData: string,
  constructorArgs: any[] = [],
  salt: string
) {
  const deployProxyCallData: BytesLike = factory.interface.encodeFunctionData(
    "deployProxy",
    [salt]
  );
  const deployProxy = encodeMultiCallArgs(
    factory.address,
    0,
    deployProxyCallData
  );
  const [deployCreate, upgradeProxy] = await encodeUpgradeProxyMultiCallArgs(
    implementationBase,
    factory,
    ethers,
    initializationCallData,
    constructorArgs,
    salt
  );
  return [deployCreate, deployProxy, upgradeProxy];
}
export async function encodeUpgradeProxyMultiCallArgs(
  implementationBase: ContractFactory,
  factory: AliceNetFactory,
  ethers: Ethers,
  initializationCallData: string,
  constructorArgs: any[] = [],
  salt: string
) {
  const deployTxData = implementationBase.getDeployTransaction(
    ...constructorArgs
  ).data as BytesLike;
  const deployCreateCallData = factory.interface.encodeFunctionData(
    "deployCreate",
    [deployTxData]
  );
  const implementationContractAddress = await calculateDeployCreateAddress(
    factory.address,
    ethers
  );

  const upgradeProxyCallData = factory.interface.encodeFunctionData(
    "upgradeProxy",
    [salt, implementationContractAddress, initializationCallData]
  );
  const deployCreate = encodeMultiCallArgs(
    factory.address,
    0,
    deployCreateCallData
  );
  const upgradeProxy = encodeMultiCallArgs(
    factory.address,
    0,
    upgradeProxyCallData
  );
  return [deployCreate, upgradeProxy];
}
export function encodeMultiCallArgs(
  targetAddress: string,
  value: BigNumberish,
  callData: BytesLike
): MultiCallArgsStruct {
  const output: MultiCallArgsStruct = {
    target: targetAddress,
    value,
    data: callData,
  };
  return output;
}

export async function calculateDeployCreateAddress(
  deployerAddress: string,
  ethers: Ethers
) {
  const factoryNonce = await ethers.provider.getTransactionCount(
    deployerAddress
  );
  return ethers.utils.getContractAddress({
    from: deployerAddress,
    nonce: factoryNonce,
  });
}

export async function deployUpgradeableSafe(
  contractName: string,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string,
  constructorArgs: any[] = [],
  salt: string,
  waitConfirmantions: number = 0,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
) {
  const ImplementationBase = await ethers.getContractFactory(contractName);
  const multiCallCallData = await encodeMultiCallDeployUpgradeableArgs(
    ImplementationBase,
    factory,
    ethers,
    initCallData,
    constructorArgs,
    salt
  );
  let estimatedGas = await factory.estimateGas.multiCall(
    multiCallCallData,
    overrides
  );
  if (estimatedGas.lt(MULTICALL_GAS_LIMIT)) {
    return factory.multiCall(multiCallCallData, overrides);
  } else {
    return deployUpgradeable(
      contractName,
      factory,
      ethers,
      initCallData,
      constructorArgs,
      salt,
      waitConfirmantions,
      overrides
    );
  }
}

export async function deployCreate(
  contractName: string,
  factory: AliceNetFactory,
  ethers: Ethers,
  constructorArgs: any[] = [],
  overrides?: Overrides & { from?: PromiseOrValue<string> }
) {
  const implementationBase = await ethers.getContractFactory(contractName);
  const deployTxData = implementationBase.getDeployTransaction(
    ...constructorArgs
  ).data as BytesLike;
  return factory.deployCreate(deployTxData, overrides);
}

export async function deployUpgradeable(
  contractName: string,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string,
  constructorArgs: Array<string>,
  salt: string = "",
  waitConfirmantion: number = 0,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
) {
  let txResponse = await deployCreate(
    contractName,
    factory,
    ethers,
    constructorArgs,
    overrides
  );
  let receipt = await txResponse.wait(waitConfirmantion);
  const implementationContractAddress = await getEventVar(
    receipt,
    "deployedRaw",
    "contractAddr"
  );
  salt =
    salt.length === 0 ? ethers.utils.formatBytes32String(contractName) : salt;
  const multiCallArgs = await encodeMultiCallDeployAndUpgradeProxyArgs(
    implementationContractAddress,
    factory,
    ethers,
    initCallData,
    salt
  );
  return factory.multiCall(multiCallArgs, overrides);
}

export async function deployCreateAndRegister(
  contractName: string,
  factory: AliceNetFactory,
  ethers: Ethers,
  constructorArgs: any[],
  overrides?: Overrides & { from?: PromiseOrValue<string> }
): Promise<ContractTransaction> {
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
  factory: AliceNetFactory,
  ethers: Ethers,
  artifacts: Artifacts,
  constructorArgs: string[]
) {
  const logicContractFactory = await ethers.getContractFactory(contractName);
  let deployBCode: BytesLike;
  if (typeof constructorArgs !== "undefined" && constructorArgs.length >= 0) {
    deployBCode = logicContractFactory.getDeployTransaction(...constructorArgs)
      .data as BytesLike;
  } else {
    deployBCode = logicContractFactory.getDeployTransaction().data as BytesLike;
  }
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
): Promise<string> {
  const artifactPaths = await artifacts.getAllFullyQualifiedNames();
  const fullName = artifactPaths.find((path) => path.endsWith(contractName));
  if (fullName === undefined) {
    throw new Error(`contract ${contractName} not found`);
  }
  return fullName;
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
