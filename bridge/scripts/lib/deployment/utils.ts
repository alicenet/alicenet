import { ContractFactory } from "ethers";
import fs from "fs";
import { ethers } from "hardhat";
import { Artifacts, HardhatRuntimeEnvironment } from "hardhat/types";
import { exit } from "process";
import readline from "readline";
import { DEFAULT_CONFIG_FILE_PATH } from "../constants";
import {
  ArgData,
  DeploymentConfig,
  DeploymentConfigWrapper,
  DeploymentList,
  InitializerArgsError,
} from "./interfaces";

type Ethers = typeof ethers;

export async function encodeInitCallData(
  implementationBase: ContractFactory,
  initializerArgs?: any[]
) {
  try {
    return implementationBase.interface.encodeFunctionData(
      "initialize",
      initializerArgs
    );
  } catch (err: any) {
    if (err.reason === "no matching function" && err.value === "initialize") {
      return "0x";
    } else if (err.reason === "types/values length mismatch") {
      throw new InitializerArgsError(
        "Initializer args provided do not match the initializer function"
      );
    } else {
      throw err;
    }
  }
}

export async function promptCheckDeploymentArgs(message: string) {
  let missingInput = true;
  if (process.env.silencer === "true") {
    missingInput = false;
  }

  let dynamicSuggestion = message;
  const defaultSuggestion = dynamicSuggestion;
  while (missingInput) {
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });
    const prompt = (query: any) =>
      new Promise((resolve) => rl.question(query, resolve));
    const answer = await prompt(dynamicSuggestion);
    if (
      answer === "y" ||
      answer === "Y" ||
      answer === "yes" ||
      answer === "Yes" ||
      answer === "YES"
    ) {
      missingInput = false;
      break;
    } else if (
      answer === "n" ||
      answer === "N" ||
      answer === "no" ||
      answer === "No" ||
      answer === "NO"
    ) {
      missingInput = false;
      exit();
    } else {
      if (dynamicSuggestion === defaultSuggestion) {
        dynamicSuggestion =
          "invalid input, enter one of the following: Y, y, yes, Yes, YES, N, n, no, No, NO";
      }
    }
    rl.close();
  }
}

export async function checkUserDirPath(path: string) {
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

export async function hasConstructorArgs(
  fullName: string,
  artifacts: Artifacts
) {
  const buildInfo: any = await artifacts.getBuildInfo(fullName);
  const path = extractPathFromFullyQualifiedName(fullName);
  const name = extractNameFromFullyQualifiedName(fullName);
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

export function extractPathFromFullyQualifiedName(fullName: string) {
  return fullName.split(":")[0];
}

export function extractNameFromFullyQualifiedName(fullName: string) {
  return fullName.split(":")[1];
}

export async function getCustomNSTag(
  fullyQaulifiedContractName: string,
  tagName: string,
  artifacts: Artifacts
): Promise<string> {
  const buildInfo = await artifacts.getBuildInfo(fullyQaulifiedContractName);
  if (buildInfo !== undefined) {
    const name = extractNameFromFullyQualifiedName(fullyQaulifiedContractName);
    const path = extractPathFromFullyQualifiedName(fullyQaulifiedContractName);
    const info: any = buildInfo?.output.contracts[path][name];
    return info.devdoc[`custom:${tagName}`];
  } else {
    throw new Error(`Failed to get natspec tag ${tagName}`);
  }
}

/**
 * @description gets the salt specified in the contracts head with a custom natspec tag @custom:salt
 * @param contractName name of the contract
 * @param artifacts artifacts object from hardhat artifacts
 * @param ethers ethersjs object
 * @returns bytes32 formatted salt
 */
export async function getBytes32SaltFromContractNSTag(
  contractName: string,
  artifacts: Artifacts,
  ethers: Ethers,
  fullyQualifiedName?: string
): Promise<string> {
  fullyQualifiedName =
    fullyQualifiedName === undefined
      ? await getFullyQualifiedName(contractName, artifacts)
      : fullyQualifiedName;
  const salt = await getSalt(fullyQualifiedName, artifacts);
  const saltType: string = await getSaltType(fullyQualifiedName, artifacts);
  return calculateSalt(salt, saltType, ethers);
}

export async function getSalt(fullName: string, artifacts: Artifacts) {
  return await getCustomNSTag(fullName, "salt", artifacts);
}

export async function getSaltType(fullName: string, artifacts: Artifacts) {
  return await getCustomNSTag(fullName, "salt-type", artifacts);
}

export async function getBytes32Salt(
  contractName: string,
  artifacts: Artifacts,
  ethers: Ethers
) {
  const fullName = await getFullyQualifiedName(contractName, artifacts);
  const salt: string = await getSalt(fullName, artifacts);
  const saltType: string = await getSaltType(fullName, artifacts);
  return calculateSalt(salt, saltType, ethers);
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

export async function getGasPrices(ethers: Ethers) {
  // get the latest block
  const latestBlock = await ethers.provider.getBlock("latest");
  // get the previous basefee from the latest block
  const _blockBaseFee = latestBlock.baseFeePerGas;
  if (_blockBaseFee === undefined || _blockBaseFee === null) {
    throw new Error("undefined block base fee per gas");
  }
  const blockBaseFee = _blockBaseFee.toBigInt();
  // miner tip
  let maxPriorityFeePerGas: bigint;
  const network = await ethers.provider.getNetwork();
  const minValue = ethers.utils.parseUnits("2.0", "gwei").toBigInt();
  if (network.chainId === 1337) {
    maxPriorityFeePerGas = minValue;
  } else {
    maxPriorityFeePerGas = BigInt(
      await ethers.provider.send("eth_maxPriorityFeePerGas", [])
    );
  }
  maxPriorityFeePerGas = (maxPriorityFeePerGas * 125n) / 100n;
  maxPriorityFeePerGas =
    maxPriorityFeePerGas < minValue ? minValue : maxPriorityFeePerGas;
  const maxFeePerGas = 2n * blockBaseFee + maxPriorityFeePerGas;
  return { maxPriorityFeePerGas, maxFeePerGas };
}

export async function verifyContract(
  hre: HardhatRuntimeEnvironment,
  deployedContractAddress: string,
  constructorArgs: Array<any>
) {
  let result;
  try {
    result = await hre.run("verify", {
      network: hre.network.name,
      address: deployedContractAddress,
      constructorArgsParams: constructorArgs,
    });
  } catch (error) {
    console.log(
      `Failed to automatically verify ${deployedContractAddress} please do it manually!`
    );
    console.log(error);
  }
  return result;
}

export const showState = async (message: string): Promise<void> => {
  if (process.env.silencer === undefined || process.env.silencer === "false") {
    console.log(message);
  }
};

export async function generateDeployConfigTemplate(
  list: DeploymentList,
  artifacts: Artifacts,
  ethers: Ethers
): Promise<DeploymentConfigWrapper> {
  const deploymentArgs: DeploymentConfigWrapper = {};
  const factoryName = await getFullyQualifiedName("AliceNetFactory", artifacts);

  const factoryContractInfo = await extractFullContractInfo(
    factoryName,
    artifacts,
    ethers
  );
  deploymentArgs[factoryName] = factoryContractInfo;

  // iterate over the deployment list
  for (const key of Object.keys(list)) {
    const arrayOfConfigs = list[key];
    for (const deploymentConfig of arrayOfConfigs) {
      deploymentArgs[deploymentConfig.fullyQualifiedName] = deploymentConfig;
    }
  }

  return deploymentArgs;
}

export async function getSortedDeployList(
  contracts: Array<string>,
  artifacts: Artifacts,
  ethers: Ethers
) {
  const deploymentList: DeploymentList = {};
  for (const contract of contracts) {
    const contractInfo = await extractFullContractInfo(
      contract,
      artifacts,
      ethers
    );

    const deployType: string | undefined = contractInfo?.deployType;
    let group: string | undefined = contractInfo?.deployGroup;

    if (group !== undefined) {
      if (contractInfo?.deployGroupIndex === undefined) {
        throw new Error(
          "If deploy-group-index is specified a deploy-group-index also should be!"
        );
      }
      try {
        // check deploy group index exists
        parseInt(contractInfo?.deployGroupIndex);
      } catch (error) {
        throw new Error(
          `Failed to convert deploy-group-index for contract ${contract}! deploy-group-index should be an integer!`
        );
      }
    } else {
      group = "general";
    }
    if (deployType !== undefined) {
      if (deploymentList[group] === undefined) {
        deploymentList[group] = [];
      }
      deploymentList[group].push(contractInfo);
    }
  }
  for (const key in deploymentList) {
    if (key !== "general") {
      deploymentList[key].sort((contractA, contractB) => {
        const indexA = parseInt(contractA.deployGroupIndex);
        const indexB = parseInt(contractB.deployGroupIndex);
        return indexA - indexB;
      });
    }
  }
  return deploymentList;
}

export async function extractFullContractInfoByContractName(
  contractName: string,
  artifacts: Artifacts,
  ethers: Ethers
): Promise<DeploymentConfig> {
  const fullyQualifiedName = await getFullyQualifiedName(
    contractName,
    artifacts
  );
  return extractFullContractInfo(fullyQualifiedName, artifacts, ethers);
}
export async function extractFullContractInfo(
  fullName: string,
  artifacts: Artifacts,
  ethers: Ethers
): Promise<DeploymentConfig> {
  let constructorArgs: ArgData = {};
  let initializerArgs: ArgData = {};

  const buildInfo = await artifacts.getBuildInfo(fullName);
  if (buildInfo !== undefined) {
    const name = extractNameFromFullyQualifiedName(fullName);
    const path = extractPathFromFullyQualifiedName(fullName);
    const info: any = buildInfo?.output.contracts[path][name];

    const methods = buildInfo.output.contracts[path][name].abi;
    for (const method of methods) {
      if (method.type === "constructor" || method.name === "initialize") {
        const args: ArgData = {};
        for (const input of method.inputs) {
          args[input.name] = "UNDEFINED";
        }

        if (method.type === "constructor") {
          constructorArgs = args;
        } else {
          initializerArgs = args;
        }
      }
    }

    const devDocSalt: string = info.devdoc["custom:salt"];
    const devDocSaltType: string = info.devdoc["custom:salt-type"];

    const salt =
      devDocSalt !== undefined
        ? calculateSalt(devDocSalt, devDocSaltType, ethers)
        : "";
    const deployGroup = info.devdoc["custom:deploy-group"];
    const deployGroupIndex = info.devdoc["custom:deploy-group-index"];
    const deployType = info.devdoc["custom:deploy-type"];
    const deploymentConfig: DeploymentConfig = {
      name,
      fullyQualifiedName: fullName,
      salt,
      deployGroup,
      deployGroupIndex,
      deployType,
      constructorArgs,
      initializerArgs,
    };
    return deploymentConfig;
  } else {
    throw new Error(`failed to fetch ${fullName} info`);
  }
}

export function writeDeploymentConfig(
  deploymentArgs: DeploymentConfigWrapper,
  configFile?: string
) {
  const file = configFile === undefined ? DEFAULT_CONFIG_FILE_PATH : configFile;
  const jsonData = JSON.stringify(deploymentArgs, null, 2);
  fs.writeFileSync(file, jsonData);
  return file;
}

export function readDeploymentConfig(file: string): DeploymentConfigWrapper {
  return readJSON(file);
}

export function readJSON(file: string) {
  const rawData = fs.readFileSync(file);
  return JSON.parse(rawData.toString("utf8"));
}

export function populateInitializerArgs(
  initializerArgs: string[],
  deploymentConfigForContract: DeploymentConfig
) {
  const initializerArgsArray = initializerArgs;
  const initializerArgsKeys = Object.keys(
    deploymentConfigForContract.initializerArgs
  );
  if (initializerArgsArray.length !== initializerArgsKeys.length) {
    throw new Error(
      `Incorrect number of initializer arguments provided. Expected ${initializerArgsKeys.length} but got ${initializerArgsArray.length}`
    );
  }

  for (let i = 0; i < initializerArgsArray.length; i++) {
    const arg = initializerArgsArray[i];
    const key = initializerArgsKeys[i];
    deploymentConfigForContract.initializerArgs[key] = arg;
  }
}

export function populateConstructorArgs(
  constructorArgs: string[],
  deploymentConfigForContract: DeploymentConfig
) {
  const constructorArgsArray = constructorArgs;
  const constructorArgsKeys = Object.keys(
    deploymentConfigForContract.constructorArgs
  );
  if (constructorArgsArray.length !== constructorArgsKeys.length) {
    throw new Error(
      `Incorrect number of constructor arguments provided. Expected ${constructorArgsKeys.length} but got ${constructorArgsArray.length}`
    );
  }

  for (let i = 0; i < constructorArgsArray.length; i++) {
    const arg = constructorArgsArray[i];
    const key = constructorArgsKeys[i];
    deploymentConfigForContract.constructorArgs[key] = arg;
  }
}

function calculateSalt(salt: string, saltType: string, ethers: Ethers): string {
  return saltType === undefined
    ? ethers.utils.formatBytes32String(salt)
    : ethers.utils.keccak256(
        ethers.utils
          .keccak256(ethers.utils.formatBytes32String(salt))
          .concat(
            ethers.utils
              .keccak256(ethers.utils.formatBytes32String(saltType))
              .slice(2)
          )
      );
}
