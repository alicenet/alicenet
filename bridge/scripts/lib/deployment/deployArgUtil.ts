import toml from "@iarna/toml";
import fs from "fs";
import { Artifacts } from "hardhat/types";
import {
  DEPLOYMENT_ARGS_TEMPLATE_FPATH,
  DEPLOYMENT_ARG_PATH,
} from "../constants";
import {
  ArgData,
  DeploymentArgs,
  extractName,
  extractPath,
} from "./deploymentUtil";

export async function writeDeploymentArgs(
  deploymentArgs: DeploymentArgs,
  configDirPath?: string
) {
  const path =
    configDirPath === undefined
      ? DEPLOYMENT_ARG_PATH + DEPLOYMENT_ARGS_TEMPLATE_FPATH
      : configDirPath + DEPLOYMENT_ARGS_TEMPLATE_FPATH;
  const config: any = deploymentArgs;
  const data = toml.stringify(config);
  fs.writeFileSync(path, data);
  // let output = fs.readFileSync(path).toString().split("\n");
  // output.unshift(
  //   "# WARNING: DO NOT CHANGE THE GENERATED DEFAULT LIST \n# TO ADD A CUSTOM LIST COPY THE FORMAT OF THE DEFAULT LIST WITH DIFFERENT FIELD NAME"
  // );
  // fs.writeFileSync(path, output.join("\n"));
}

export async function generateDeployArgTemplate(
  list: Array<string>,
  artifacts: Artifacts
): Promise<DeploymentArgs> {
  const deploymentArgs: DeploymentArgs = {
    constructor: {},
    initializer: {},
  };
  for (const contract of list) {
    // check each contract for a constructor and
    const cArgs: Array<ArgData> = await getConstructorArgsABI(
      contract,
      artifacts
    );
    const iArgs: Array<ArgData> = await getInitializerArgsABI(
      contract,
      artifacts
    );
    if (cArgs.length !== 0) {
      deploymentArgs.constructor[contract] = cArgs;
    }
    if (iArgs.length !== 0) {
      deploymentArgs.initializer[contract] = iArgs;
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
  const args: Array<ArgData> = [];
  const methods = await getContractABI(fullName, artifacts);
  for (const method of methods) {
    const target = methodName === "constructor" ? method.type : method.name;
    if (target === methodName) {
      for (const input of method.inputs) {
        const argData: ArgData = {
          [input.name]: "UNDEFINED",
        };
        args.push(argData);
      }
    }
  }
  return args;
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
