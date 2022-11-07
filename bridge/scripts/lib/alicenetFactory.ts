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
import {
  AliceNetFactory,
  AliceNetFactory__factory,
} from "../../typechain-types";
import { PromiseOrValue } from "../../typechain-types/common";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  EVENT_DEPLOYED_RAW,
  MULTICALL_GAS_LIMIT,
} from "./constants";
import { getBytes32SaltFromContractNSTag } from "./deployment/deploymentUtil";
type Ethers = typeof import("../../node_modules/ethers/lib/ethers") &
  HardhatEthersHelpers;

export type MultiCallArgsStruct = {
  target: string;
  value: BigNumberish;
  data: BytesLike;
};
export class MultiCallGasError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "MultiCallGasError";
  }
}
export async function deployFactory(
  constructorArgs: string,
  ethers: Ethers,
  factoryBase?: AliceNetFactory__factory,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
): Promise<AliceNetFactory> {
  factoryBase =
    factoryBase === undefined
      ? await ethers.getContractFactory(ALICENET_FACTORY)
      : factoryBase;
  return factoryBase.deploy(constructorArgs, overrides);
}

//multicall deploy logic, proxy, and upgrade proxy
/**
 * @description uses multicall to deploy logic contract with deployCreate, deploys proxy with deployProxy, and upgrades proxy with upgradeProxy
 * @dev since upgradeable contracts go through proxies, constructor args can only be used to set immutable variables
 * this function will fail if gas cost exceeds 10 million gas units
 * @param implementationBase ethers contract factory for the implementation contract
 * @param factory an instance of a deployed and connected factory
 * @param ethers ethers object
 * @param initCallData encoded initialization call data for contracts with a initialize function
 * @param constructorArgs a list of arguements to pass to the constructor of the implementation contract, only for immutable variables
 * @param salt bytes32 formatted salt used for deploycreate2 and to reference the contract in lookup
 * @param overrides
 * @returns a promise that resolves to the deployed contracts
 */
export async function multiCallDeployUpgradeable(
  implementationBase: ContractFactory,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string,
  constructorArgs: any[] = [],
  salt: string,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
): Promise<ContractTransaction> {
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
    throw new MultiCallGasError(
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
  const multiCallArgs = await encodeMultiCallUpgradeProxyArgs(
    implementationBase,
    factory,
    ethers,
    initCallData,
    constructorArgs,
    salt
  );
  return factory.multiCall(multiCallArgs, overrides);
}
/**
 * @description uses factory multicall to deploy a proxy contract with deployProxy, then upgrades the proxy with upgradeProxy
 * @param logicAddress address of the logic contract already deployed
 * @param factory instance of deployed and connected alicenetFactory
 * @param salt bytes32 formatted salt used for deployCreate2 and to reference the contract in lookup
 * @param initCallData
 * @returns
 */
export async function multicallDeployAndUpgradeProxy(
  logicAddress: string,
  factory: AliceNetFactory,
  salt: string,
  initCallData: string
) {
  const multiCallArgs = await encodeMultiCallDeployAndUpgradeProxyArgs(
    logicAddress,
    factory,
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

/**
 * @description encodes multicall for deployProxy and upgradeProxy
 * @dev this function is used if the logic contract is too big to be deployed with a full multicall
 * this just deploys a proxy and upgrade the proxy to point to the implementationAddress
 * @param implementationAddress address of the implementation contract
 * @param factory instance of deployed and connected alicenet factory
 * @param initCallData encoded init calldata, 0x if no initializer function
 * @param salt bytes32 format of the salt that references the proxy contract
 * @returns
 */
export async function encodeMultiCallDeployAndUpgradeProxyArgs(
  implementationAddress: string,
  factory: AliceNetFactory,
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
/**
 * @description multicall deployCreate and upgradeProxy
 * @param contractName name of logic contract to deploy
 * @param factory ethers connected instance of alicenet factory
 * @param ethers ethers js object
 * @param constructorArgs array of constructor arguments
 * @param initCallData encoded init calldata, 0x if no initializer function
 * @param salt salt used for deployCreate2 and to reference the proxy contract in lookup
 * @param overrides transaction overrides
 * @returns
 */
export async function multiCallUpgradeProxy(
  contractName: string,
  factory: AliceNetFactory,
  ethers: Ethers,
  constructorArgs: any[],
  initCallData: string,
  salt: string,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
) {
  const implementationBase = await ethers.getContractFactory(contractName);
  const multiCallArgs = await encodeMultiCallUpgradeProxyArgs(
    implementationBase,
    factory,
    ethers,
    initCallData,
    constructorArgs,
    salt
  );
  return factory.multiCall(multiCallArgs, overrides);
}

/**
 * @description encodes the arguments for alicenet factory multicall to
 * deploy a logic contract with deploycreate,
 * deploy a proxy with deployProxy,
 * and upgrade the proxy with upgradeProxy
 * @param implementationBase ethers contract factory for the implementation contract
 * @param factory instance of deployed and connected alicenetFactory
 * @param ethers instance of ethers
 * @param initCallData encoded call data for the initialize function of the implementation contract
 * @param constructorArgs string array of constructor arguments, only used to set immutable variables
 * @param salt bytes32 formatted salt used for deploycreate2 and to reference the contract in lookup
 * @returns an array of encoded multicall data for deployCreate, deployProxy, and upgradeProxy
 */
export async function encodeMultiCallDeployUpgradeableArgs(
  implementationBase: ContractFactory,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string,
  constructorArgs: string[] = [],
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
  const [deployCreate, upgradeProxy] = await encodeMultiCallUpgradeProxyArgs(
    implementationBase,
    factory,
    ethers,
    initCallData,
    constructorArgs,
    salt
  );
  return [deployCreate, deployProxy, upgradeProxy];
}
/**
 * @decription encodes a multicall for deploying a logic contract with deployCreate, and upgradeProxy to point to the newly deployed implementation contract
 * @param implementationBase ethers contract instance of the implementation contract
 * @param factory connected instance of alicenetFactory
 * @param ethers instance of hardhat ethers
 * @param initCallData encoded call data for the initialize function of the implementation contract
 * @param constructorArgs encoded constructor arguments
 * @param salt bytes32 formatted salt used to deploy the proxy
 * @returns
 */
export async function encodeMultiCallUpgradeProxyArgs(
  implementationBase: ContractFactory,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string,
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
    [salt, implementationContractAddress, initCallData]
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

export async function deployUpgradeableGasSafe(
  contractName: string,
  factory: AliceNetFactory,
  ethers: Ethers,
  initCallData: string,
  constructorArgs: any[],
  salt: string,
  waitConfirmantions: number = 0,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
) {
  const ImplementationBase = await ethers.getContractFactory(contractName);
  try {
    return await multiCallDeployUpgradeable(
      ImplementationBase,
      factory,
      ethers,
      initCallData,
      constructorArgs,
      salt,
      overrides
    );
  } catch (err) {
    if (err instanceof MultiCallGasError) {
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
    throw err;
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
/**
 *
 * @param contractName name of the contract to deploy
 * @param factory instance of deployed and connected alicenetFactory
 * @param ethers ethers js object
 * @param initCallData encoded call data for the initialize function of the implementation contract
 * @param constructorArgs constructor arguments for the implementation contract
 * @param salt bytes32 formatted salt used to deploy the proxy
 * @param waitConfirmantion number of confirmations to wait for before returning the transaction
 * @param overrides
 * @returns
 */
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
  //use mutlticall to deploy proxy and upgrade proxy
  const multiCallArgs = await encodeMultiCallDeployAndUpgradeProxyArgs(
    implementationContractAddress,
    factory,
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
  salt: string,
  overrides?: Overrides & { from?: PromiseOrValue<string> }
): Promise<ContractTransaction> {
  const logicContract: any = await ethers.getContractFactory(contractName);
  // if not constructor ars is provide and empty array is used to indicate no constructor args
  // encode deployBcode,
  const deployTxData = logicContract.getDeployTransaction(...constructorArgs)
    .data as BytesLike;
  return factory.deployCreateAndRegister(deployTxData, salt, overrides);
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
    logicAddress: await getEventVar(receipt, EVENT_DEPLOYED_RAW, CONTRACT_ADDR),
    proxySalt: await getBytes32SaltFromContractNSTag(
      contractName,
      artifacts,
      ethers
    ),
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
