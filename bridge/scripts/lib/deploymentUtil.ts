import { artifacts, ethers, run } from "hardhat";
import {
  getDeploymentConstructorArgs,
  getDeploymentInitializerArgs,
} from "./deployArgUtil";

export interface InitData {
  constructorArgs: { [key: string]: any };
  initializerArgs: { [key: string]: any };
}
export interface ArgTemplate {
  [key: string]: any;
  type: string;
}

//function to deploy the factory
export async function deployFactory() {
  return await run("deployFactory", { factoryName: "MadnetFactory" });
}

export async function deployStatic(fullyQualifiedName: string) {
  let initializerArgs: Array<string> = [];
  let initCallData = "0x";
  //check if contract needs to be initialized
  let initAble = await isInitializable(fullyQualifiedName);
  if (initAble) {
    initializerArgs = await getDeploymentInitializerArgs(fullyQualifiedName);
    initCallData = await getEncodedInitCallData(
      fullyQualifiedName,
      initializerArgs
    );
  }

  let hasConArgs = await hasConstructorArgs(fullyQualifiedName);
  let constructorArgs = hasConArgs
    ? await getDeploymentConstructorArgs(fullyQualifiedName)
    : [];
  return await run("deployMetamorphic", {
    contractName: extractName(fullyQualifiedName),
    initCallData: initCallData,
    constructorArgs: constructorArgs,
  });
}

export async function deployUpgradeableProxy(fullyQualifiedName: string) {
  let name: string = extractName(fullyQualifiedName);
  let initializerArgs: Array<string> = [];
  let initCallData = "0x";
  let initAble = await isInitializable(fullyQualifiedName);
  if (initAble) {
    initializerArgs = await getDeploymentInitializerArgs(fullyQualifiedName);
    initCallData = await getEncodedInitCallData(
      fullyQualifiedName,
      initializerArgs
    );
  }
  let hasConArgs = await hasConstructorArgs(fullyQualifiedName);
  let constructorArgs = hasConArgs
    ? await getDeploymentConstructorArgs(fullyQualifiedName)
    : [];
  return run("deployUpgradeableProxy", {
    contractName: extractName(fullyQualifiedName),
    initCallData: initCallData,
    constructorArgs: constructorArgs,
  });
}

export async function isInitializable(fullyQualifiedName: string) {
  let buildInfo: any = await artifacts.getBuildInfo(fullyQualifiedName);
  let path = extractPath(fullyQualifiedName);
  let name = extractName(fullyQualifiedName);
  let methods = buildInfo.output.contracts[path][name].abi;
  for (let method of methods) {
    if (method.name === "initialize") {
      return true;
    }
  }
  return false;
}

export async function hasConstructorArgs(fullName: string) {
  let buildInfo: any = await artifacts.getBuildInfo(fullName);
  let path = extractPath(fullName);
  let name = extractName(fullName);
  let methods = buildInfo.output.contracts[path][name].abi;
  for (let method of methods) {
    if (method.type === "constructor") {
      return method.inputs.length > 0 ? true : false;
    }
  }
  return false;
}

export async function getEncodedInitCallData(
  fullName: string,
  args: Array<any>
) {
  let name = extractName(fullName);
  let contractFactory = await ethers.getContractFactory(name);
  return contractFactory.interface.encodeFunctionData("initialize", args);
}

export async function getContract(name: string) {
  let artifactPaths = await artifacts.getAllFullyQualifiedNames();
  for (let i = 0; i < artifactPaths.length; i++) {
    if (artifactPaths[i].split(":")[1] === name) {
      return String(artifactPaths[i]);
    }
  }
}

export async function getAllContracts() {
  //get a list with all the contract names
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
  tagName: string
): Promise<string> {
  let buildInfo = await artifacts.getBuildInfo(fullyQaulifiedContractName);
  if (buildInfo !== undefined) {
    let name = extractName(fullyQaulifiedContractName);
    let path = extractPath(fullyQaulifiedContractName);
    let info: any = buildInfo?.output.contracts[path][name];
    return info["devdoc"][`custom:${tagName}`];
  } else {
    throw new Error(`Failed to get natspec tag`);
  }
}

export async function getSalt(fullName: string) {
  return await getCustomNSTag(fullName, "salt");
}

export async function getBytes32Salt(contractName: string) {
  let salt: string = await getSalt(contractName);
  return ethers.utils.formatBytes32String(salt);
}

export async function getDeployType(fullName: string) {
  return await getCustomNSTag(fullName, "deploy-type");
}

export async function getDeployGroup(fullName: string) {
  return await getCustomNSTag(fullName, "deploy-group");
}

export async function getDeployGroupIndex(fullName: string) {
  return await getCustomNSTag(fullName, "deploy-group-index");
}
