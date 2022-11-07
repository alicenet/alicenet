import toml from "@iarna/toml";
import fs from "fs";
import { Artifacts, HardhatEthersHelpers } from "hardhat/types";
import {
  DEPLOYMENT_ARGS_TEMPLATE_FPATH,
  DEPLOYMENT_ARG_PATH,
} from "../constants";
import {
  ArgData,
  DeploymentConfigWrapper,
  extractName,
  extractPath,
  getFullyQualifiedName,
} from "./deploymentUtil";

type Ethers = typeof import("../../../node_modules/ethers/lib/ethers") &
  HardhatEthersHelpers;

export async function writeDeploymentArgs(
  deploymentArgs: DeploymentConfigWrapper,
  configDirPath?: string
) {
  const path =
    configDirPath === undefined
      ? DEPLOYMENT_ARG_PATH + DEPLOYMENT_ARGS_TEMPLATE_FPATH
      : configDirPath + DEPLOYMENT_ARGS_TEMPLATE_FPATH;
  const config: any = deploymentArgs;
  const data = toml.stringify(config);
  fs.writeFileSync(path, data);
  // also write to json file
  const jsonData = JSON.stringify(deploymentArgs, null, 2);
  fs.writeFileSync(path + ".json", jsonData);
}

export async function generateDeployArgTemplate(
  list: Array<string>,
  artifacts: Artifacts,
  ethers: Ethers
): Promise<DeploymentConfigWrapper> {
  const deploymentArgs: DeploymentConfigWrapper = {};
  const factoryName = await getFullyQualifiedName("AliceNetFactory", artifacts);
  const contracts = [factoryName, ...list];
  for (const contract of contracts) {
    console.log(contract);

    const contractInfo = await extractFullContractInfo(
      contract,
      artifacts,
      ethers
    );
    if (contractInfo !== undefined) {
      deploymentArgs[contract] = contractInfo;
    }
  }
  return deploymentArgs;
}

export async function getConstructorArgsABI(
  fullName: string,
  artifacts: Artifacts
) {
  return await parseArgsArray(fullName, "constructor", artifacts);
}

export async function getInitializerArgsABI(
  fullName: string,
  artifacts: Artifacts
) {
  return await parseArgsArray(fullName, "initialize", artifacts);
}

export async function getMethodArgCount(
  fullName: string,
  methodName: string,
  artifacts: Artifacts
) {
  const methods = await getContractABI(fullName, artifacts);
  for (const method of methods) {
    const target = methodName === "constructor" ? method.type : method.name;
    if (target === "constructor") {
      return method.inputs.length;
    }
  }
  return 0;
}

export async function getInitializerArgCount(
  fullName: string,
  artifacts: Artifacts
) {
  const methods = await getContractABI(fullName, artifacts);
  for (const method of methods) {
    if (method.name === "initialize") {
      return method.inputs.length;
    }
  }
  return 0;
}

export async function parseArgsArray(
  fullName: string,
  methodName: string,
  artifacts: Artifacts
) {
  const args: ArgData = {};
  const methods = await getContractABI(fullName, artifacts);
  for (const method of methods) {
    const target = methodName === "constructor" ? method.type : method.name;
    if (target === methodName) {
      for (const input of method.inputs) {
        args[input.name] = "UNDEFINED";
      }
    }
  }
  return args;
}

export async function extractFullContractInfo(
  fullName: string,
  artifacts: Artifacts,
  ethers: Ethers
) {
  let constructorArgs: ArgData = {};
  let initializerArgs: ArgData = {};

  const buildInfo = await artifacts.getBuildInfo(fullName);
  if (buildInfo !== undefined) {
    const name = extractName(fullName);
    const path = extractPath(fullName);
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

    let salt =
      info.devdoc["custom:salt"] !== undefined
        ? ethers.utils.formatBytes32String(info.devdoc["custom:salt"])
        : "";
    const deployGroup = info.devdoc["custom:deploy-group"];
    const deployGroupIndex = info.devdoc["custom:deploy-group-index"];
    const deployType = info.devdoc["custom:deploy-type"];
    return {
      name: name,
      fullyQualifiedName: fullName,
      salt,
      deployGroup,
      deployGroupIndex,
      deployType,
      constructorArgs,
      initializerArgs,
    };
  }

  return undefined;
}

async function getContractABI(fullName: string, artifacts: Artifacts) {
  const buildInfo = await artifacts.getBuildInfo(fullName);
  const path = extractPath(fullName);
  const name = extractName(fullName);
  if (buildInfo !== undefined) {
    return buildInfo.output.contracts[path][name].abi;
  } else {
    throw new Error(`failed to fetch ${fullName} abi`);
  }
}

export async function getBuildInfo(fullName: string, artifacts: Artifacts) {
  return await artifacts.getBuildInfo(fullName);
}
