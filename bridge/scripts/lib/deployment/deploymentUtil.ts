import { HardhatEthersHelpers } from "@nomiclabs/hardhat-ethers/types";
import { BigNumberish, BytesLike, ContractFactory } from "ethers";
import {
  Artifacts,
  HardhatRuntimeEnvironment,
  RunTaskFunction,
} from "hardhat/types";
// import {
//   AliceNetFactory,
//   AliceNetFactory__factory,
// } from "../../../typechain-types";
import { getEventVar } from "../alicenetFactoryTasks";
import { encodeMultiCallArgs } from "../alicenetTasks";
import {
  ALICENET_FACTORY,
  CONTRACT_ADDR,
  DEFAULT_CONFIG_OUTPUT_DIR,
  DEPLOYED_PROXY,
  DEPLOY_CREATE,
  DEPLOY_PROXY,
  INITIALIZER,
  ONLY_PROXY,
  UPGRADEABLE_DEPLOYMENT,
  UPGRADE_PROXY,
} from "../constants";
import { readDeploymentArgs } from "./deploymentConfigUtil";
import { ProxyData } from "./factoryStateUtil";

type Ethers = typeof import("../../../node_modules/ethers/lib/ethers") &
  HardhatEthersHelpers;

export interface ArgData {
  [key: string]: string;
}
export interface ContractArgs {
  [key: string]: Array<ArgData>;
}
export interface DeploymentArgs {
  constructor: ContractArgs;
  initializer: ContractArgs;
}

export type DeployProxyMCArgs = {
  contractName: string;
  logicAddress: string;
  waitConfirmation?: boolean;
  factoryAddress?: string;
  initCallData?: BytesLike;
  outputFolder?: string;
};
export type MultiCallArgsStruct = {
  target: string;
  value: BigNumberish;
  data: BytesLike;
};
export type DeployArgs = {
  contractName: string;
  factoryAddress: string;
  waitConfirmation?: boolean;
  initCallData?: string;
  constructorArgs?: any;
  outputFolder?: string;
  verify?: boolean;
};

export type Args = {
  contractName: string;
  factoryAddress?: string;
  salt?: BytesLike;
  initCallData?: string;
  constructorArgs?: any;
  outputFolder?: string;
};
export interface InitData {
  constructorArgs: { [key: string]: any };
  initializerArgs: { [key: string]: any };
}

export interface ContractDescriptor {
  name: string;
  fullyQualifiedName: string;
  deployGroup: string;
  deployGroupIndex: number;
  deployType: string;
  constructorArgs: Array<any>;
  initializerArgs: Array<any>;
}

export async function getContractDescriptor(
  contractName: string,
  constructorArgs: Array<any>,
  initializerArgs: Array<any>,
  hre: HardhatRuntimeEnvironment
): Promise<ContractDescriptor> {
  const fullyQualifiedName = await getFullyQualifiedName(
    contractName,
    hre.artifacts
  );
  return {
    name: contractName,
    fullyQualifiedName: await getFullyQualifiedName(
      contractName,
      hre.artifacts
    ),
    deployGroup: await getDeployGroup(fullyQualifiedName, hre.artifacts),
    deployGroupIndex: parseInt(
      await getDeployGroupIndex(fullyQualifiedName, hre.artifacts),
      10
    ),
    deployType: await getDeployType(fullyQualifiedName, hre.artifacts),
    constructorArgs: constructorArgs,
    initializerArgs: initializerArgs,
  };
}
// function to deploy the factory
export async function deployFactory(run: RunTaskFunction, usrPath?: string) {
  return await run("deployFactory", { outputFolder: usrPath });
}

export async function getDeployMetaArgs(
  fullyQualifiedName: string,
  waitConfirmation: boolean,
  factoryAddress: string,
  artifacts: Artifacts,
  inputFolder?: string,
  outputFolder?: string,
  verify?: boolean
): Promise<DeployArgs> {
  let initCallData;
  // check if contract needs to be initialized
  const initAble = await isInitializable(fullyQualifiedName, artifacts);
  if (initAble) {
    const initializerArgs = await getDeploymentInitializerArgs(
      fullyQualifiedName,
      inputFolder
    );
    initCallData = await getEncodedInitCallData(initializerArgs);
  }
  const hasConArgs = await hasConstructorArgs(fullyQualifiedName, artifacts);
  const constructorArgs = hasConArgs
    ? await getDeploymentConstructorArgs(fullyQualifiedName, inputFolder)
    : undefined;
  return {
    contractName: extractName(fullyQualifiedName),
    waitConfirmation: waitConfirmation,
    factoryAddress: factoryAddress,
    initCallData: initCallData,
    constructorArgs: constructorArgs,
    outputFolder: outputFolder,
    verify: verify,
  };
}
export async function getFactoryDeploymentArgs(
  artifacts: Artifacts,
  inputFolder?: string
) {
  const fullyQualifiedName = await getFullyQualifiedName(
    "AliceNetFactory",
    artifacts
  );
  const hasConArgs = await hasConstructorArgs(fullyQualifiedName, artifacts);
  const constructorArgs = hasConArgs
    ? await getDeploymentConstructorArgs(fullyQualifiedName, inputFolder)
    : undefined;
  return constructorArgs;
}
export async function getDeployUpgradeableProxyArgs(
  fullyQualifiedName: string,
  factoryAddress: string,
  artifacts: Artifacts,
  waitConfirmation?: boolean,
  inputFolder?: string,
  outputFolder?: string,
  verify?: boolean
): Promise<DeployArgs> {
  let initCallData;
  const initAble = await isInitializable(fullyQualifiedName, artifacts);
  if (initAble) {
    const initializerArgs = await getDeploymentInitializerArgs(
      fullyQualifiedName,
      inputFolder
    );
    initCallData = await getEncodedInitCallData(initializerArgs);
  }
  const hasConArgs = await hasConstructorArgs(fullyQualifiedName, artifacts);
  const constructorArgs = hasConArgs
    ? await getDeploymentConstructorArgs(fullyQualifiedName, inputFolder)
    : undefined;
  return {
    contractName: extractName(fullyQualifiedName),
    waitConfirmation: waitConfirmation,
    factoryAddress: factoryAddress,
    initCallData: initCallData,
    constructorArgs: constructorArgs,
    outputFolder: outputFolder,
    verify: verify,
  };
}

export async function isInitializable(
  fullyQualifiedName: string,
  artifacts: Artifacts
) {
  const buildInfo: any = await artifacts.getBuildInfo(fullyQualifiedName);
  const path = extractPath(fullyQualifiedName);
  const name = extractName(fullyQualifiedName);
  const methods = buildInfo.output.contracts[path][name].abi;
  for (const method of methods) {
    if (method.name === INITIALIZER) {
      return true;
    }
  }
  return false;
}

export async function hasConstructorArgs(
  fullName: string,
  artifacts: Artifacts
) {
  const buildInfo: any = await artifacts.getBuildInfo(fullName);
  const path = extractPath(fullName);
  const name = extractName(fullName);
  const methods = buildInfo.output.contracts[path][name].abi;
  for (const method of methods) {
    if (method.type === "constructor") {
      return method.inputs.length > 0;
    }
  }
  return false;
}
/**
 * @description encodes init call state input to be used by the custom hardhat tasks
 * @param args values of the init call state as an array of strings where each string represents variable value
 * @returns the args array as a comma delimited string
 */
export async function getEncodedInitCallData(
  args: Array<string> | undefined
): Promise<string | undefined> {
  if (args !== undefined) {
    return args.toString();
  }
}

export async function getContract(name: string, artifacts: Artifacts) {
  const artifactPaths = await artifacts.getAllFullyQualifiedNames();
  for (let i = 0; i < artifactPaths.length; i++) {
    if (artifactPaths[i].split(":")[1] === name) {
      return String(artifactPaths[i]);
    }
  }
}

export async function getAllContracts(artifacts: Artifacts) {
  // get a list with all the contract names
  return await artifacts.getAllFullyQualifiedNames();
}

export function extractPath(fullName: string) {
  return fullName.split(":")[0];
}

export function extractName(fullName: string) {
  return fullName.split(":")[1];
}

export async function getCustomNSTag(
  fullyQaulifiedContractName: string,
  tagName: string,
  artifacts: Artifacts
): Promise<string> {
  const buildInfo = await artifacts.getBuildInfo(fullyQaulifiedContractName);
  if (buildInfo !== undefined) {
    const name = extractName(fullyQaulifiedContractName);
    const path = extractPath(fullyQaulifiedContractName);
    const info: any = buildInfo?.output.contracts[path][name];
    return info.devdoc[`custom:${tagName}`];
  } else {
    throw new Error(`Failed to get natspec tag ${tagName}`);
  }
}

// return a list of constructor inputs for each contract
export async function getDeploymentConstructorArgs(
  fullName: string,
  configDirPath?: string
) {
  let output: Array<string> = [];
  // get the deployment args
  const path =
    configDirPath === undefined
      ? DEFAULT_CONFIG_OUTPUT_DIR + "/deploymentArgsTemplate"
      : configDirPath + "/deploymentArgsTemplate";
  const deploymentConfig: any = await readDeploymentArgs(path);
  if (deploymentConfig !== undefined) {
    const deploymentArgs: DeploymentArgs = {
      constructor: deploymentConfig.constructor,
      initializer: deploymentConfig.initializer,
    };
    if (
      deploymentArgs.constructor !== undefined &&
      deploymentArgs.constructor[fullName] !== undefined
    ) {
      output = extractArgs(deploymentArgs.constructor[fullName]);
    }
  } else {
    output = [];
  }
  return output;
}

export function extractArgs(input: Array<ArgData>) {
  const output: Array<string> = [];
  for (let i = 0; i < input.length; i++) {
    const argName = Object.keys(input[i])[0];
    const argData = input[i];
    output.push(argData[argName]);
  }
  return output;
}

// return a list of initializer inputs for each contract
export async function getDeploymentInitializerArgs(
  fullName: string,
  configDirPath?: string
) {
  let output: Array<string> | undefined;
  const path =
    configDirPath === undefined
      ? DEFAULT_CONFIG_OUTPUT_DIR + "/deploymentArgsTemplate"
      : configDirPath + "/deploymentArgsTemplate";
  const deploymentConfig: any = await readDeploymentArgs(path);
  if (deploymentConfig !== undefined) {
    const deploymentArgs: DeploymentArgs = deploymentConfig;
    if (
      deploymentArgs.initializer !== undefined &&
      deploymentArgs.initializer[fullName] !== undefined
    ) {
      output = extractArgs(deploymentArgs.initializer[fullName]);
    }
  } else {
    output = undefined;
  }
  return output;
}

export async function getSalt(fullName: string, artifacts: Artifacts) {
  return await getCustomNSTag(fullName, "salt", artifacts);
}

export async function getBytes32Salt(
  contractName: string,
  artifacts: Artifacts,
  ethers: Ethers
) {
  const fullName = await getFullyQualifiedName(contractName, artifacts);
  const salt: string = await getSalt(fullName, artifacts);
  return ethers.utils.formatBytes32String(salt);
}
export async function getFullyQualifiedName(
  contractName: string,
  artifacts: Artifacts
) {
  const contractArtifact = await artifacts.readArtifact(contractName);
  const path = contractArtifact.sourceName;
  return path + ":" + contractName;
}
export async function getDeployType(fullName: string, artifacts: Artifacts) {
  return await getCustomNSTag(fullName, "deploy-type", artifacts);
}

export async function getDeployGroup(fullName: string, artifacts: Artifacts) {
  return await getCustomNSTag(fullName, "deploy-group", artifacts);
}

export async function getDeployGroupIndex(
  fullName: string,
  artifacts: Artifacts
) {
  return await getCustomNSTag(fullName, "deploy-group-index", artifacts);
}

export async function getDeployUpgradeableMultiCallArgs(
  contractDescriptor: ContractDescriptor,
  hre: HardhatRuntimeEnvironment,
  factoryAddr: string,
  inputedInitCallData?: string
) {
  const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
  const factory = factoryBase.attach(factoryAddr);
  const logicContract: ContractFactory = await hre.ethers.getContractFactory(
    contractDescriptor.name
  );
  const logicFactory = await hre.ethers.getContractFactory(
    contractDescriptor.name
  );
  const deployTxReq = logicContract.getDeployTransaction(
    ...contractDescriptor.constructorArgs
  );
  let initCallData = "0x";
  if (inputedInitCallData) {
    initCallData = inputedInitCallData;
  }
  if (contractDescriptor.initializerArgs.length > 0)
    initCallData = logicFactory.interface.encodeFunctionData(
      INITIALIZER,
      contractDescriptor.initializerArgs
    );
  const salt = await getBytes32Salt(
    contractDescriptor.name,
    hre.artifacts,
    hre.ethers
  );
  const nonce = await hre.ethers.provider.getTransactionCount(factory.address);
  const logicAddress = hre.ethers.utils.getContractAddress({
    from: factory.address,
    nonce: nonce,
  });

  // encode deploy create
  const deployCreateCallData: BytesLike =
    factoryBase.interface.encodeFunctionData(DEPLOY_CREATE, [deployTxReq.data]);
  // encode the deployProxy function call with Salt as arg
  const deployProxyCallData: BytesLike =
    factoryBase.interface.encodeFunctionData(DEPLOY_PROXY, [salt]);
  // encode upgrade proxy multicall
  const upgradeProxyCallData: BytesLike =
    factoryBase.interface.encodeFunctionData(UPGRADE_PROXY, [
      salt,
      logicAddress,
      initCallData,
    ]);
  const deployCreate = encodeMultiCallArgs(
    factory.address,
    0,
    deployCreateCallData
  );
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
  const multiCallArgs = [deployCreate, deployProxy, upgradeProxy];
  return multiCallArgs;
}

export async function deployContractsMulticall(
  contracts: ContractDescriptor[],
  hre: HardhatRuntimeEnvironment,
  factoryAddr: string,
  txCount: number,
  inputFolder?: string,
  outputFolder?: string
) {
  const factoryBase = await hre.ethers.getContractFactory(ALICENET_FACTORY);
  const factory = factoryBase.attach(factoryAddr);
  let proxyData: ProxyData;
  let multiCallArgsArray = Array();

  for (let i = 0; i < contracts.length; i++) {
    const contract = contracts[i];
    if (true) {
      const deployType = contract.deployType;
      switch (deployType) {
        case UPGRADEABLE_DEPLOYMENT: {
          let multiCallArgsArray = Array();
          let [deployCreate, deployProxy, upgradeProxy] =
            await getDeployUpgradeableMultiCallArgs(
              contract,
              hre,
              factory.address
            );
          multiCallArgsArray.push(deployCreate);
          multiCallArgsArray.push(deployProxy);
          multiCallArgsArray.push(upgradeProxy);
          const txResponse = await factory.multiCall(multiCallArgsArray);
          const receipt = await txResponse.wait();
          const address = getEventVar(receipt, DEPLOYED_PROXY, CONTRACT_ADDR);
          console.log(contract.name, "deployed at:", address);
          break;
        }
        case ONLY_PROXY: {
          const name = extractName(contract.fullyQualifiedName);
          const salt: BytesLike = await getBytes32Salt(
            name,
            hre.artifacts,
            hre.ethers
          );
          const factoryAddress = factory.address;
          proxyData = await hre.run("deployProxy", {
            factoryAddress,
            salt,
          });
          break;
        }
        default: {
          break;
        }
      }
    }
  }
  return multiCallArgsArray;
}
