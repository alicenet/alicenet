import toml from "@iarna/toml";
import fs from "fs";
import { readDeploymentArgs } from "./deploymentConfigUtil";
import { DEPLOYMENT_ARG_PATH } from "../constants";
import { ArgData, DeploymentArgs, extractName, extractPath } from "./deploymentUtil";
import { Artifacts } from "hardhat/types";


export async function writeDeploymentArgs(deploymentArgs: DeploymentArgs, usrPath?: string){
  let path = usrPath === undefined ? DEPLOYMENT_ARG_PATH : usrPath;
  if(deploymentArgs !== undefined){
    let config:any = deploymentArgs
    // let config:any = readDeploymentConfig();
    // if(deploymentArgs.initializer !== undefined){
    //   config.initializer = deploymentArgs.initializer;
    // }
    // if(deploymentArgs.constructor !== undefined){
    //   config.constructor = deploymentArgs.constructor;
    // }
    let data = toml.stringify(config)
    fs.writeFileSync(path, data)
    // let output = fs.readFileSync(path).toString().split("\n");
    // output.unshift(
    //   "# WARNING: DO NOT CHANGE THE GENERATED DEFAULT LIST \n# TO ADD A CUSTOM LIST COPY THE FORMAT OF THE DEFAULT LIST WITH DIFFERENT FIELD NAME"
    // );
    // fs.writeFileSync(path, output.join("\n"));
  }
}

export async function generateDeployArgTemplate(list: Array<string>, artifacts: Artifacts): Promise<DeploymentArgs>{
  let deploymentArgs:DeploymentArgs = {
    constructor: {},
    initializer: {}
  };
  for (let contract of list) {
    //check each contract for a constructor and
    let cArgs: Array<ArgData> = await getConstructorArgsABI(contract, artifacts);
    let iArgs: Array<ArgData> = await getInitializerArgsABI(contract, artifacts);
    if (cArgs.length != 0) {
      deploymentArgs.constructor[contract] = cArgs;
    }
    if (iArgs.length != 0) {
      deploymentArgs.initializer[contract] = iArgs;
    }
  }
  return deploymentArgs
}

export async function getConstructorArgsABI(fullName: string, artifacts: Artifacts) {
  return await parseArgsArray(fullName, "constructor", artifacts);
}

export async function getInitializerArgsABI(fullName: string, artifacts: Artifacts) { 
  return await parseArgsArray(fullName, "initialize", artifacts);
}

export async function getMethodArgCount(fullName: string, methodName: string, artifacts: Artifacts) {
  let methods = await getContractABI(fullName, artifacts);
  for (let method of methods) {
    let target = methodName === "constructor" ? method.type : method.name
    if (target === "constructor") {
      return method.inputs.length;
    }
  }
  return 0;
}

export async function getInitializerArgCount(fullName: string, artifacts: Artifacts) {
  let methods = await getContractABI(fullName, artifacts);
  for (let method of methods) {
    if (method.name === "initialize") {
      return method.inputs.length;
    }
  }
  return 0;
}

export async function parseArgsArray(fullName: string, methodName: string, artifacts: Artifacts) {
  let args: Array<ArgData> = [];
  let methods = await getContractABI(fullName, artifacts)
  for (let method of methods) {
    let target = methodName === "constructor" ? method.type : method.name
    if (target === methodName) {
      for (let input of method.inputs) {
        let argData: ArgData = {
          [input.name]: "UNDEFINED"
        };
        args.push(argData);
      };
    };
  };
  return args;
}

async function getContractABI(fullName: string, artifacts: Artifacts){
  let buildInfo =  await artifacts.getBuildInfo(fullName);
  let path = extractPath(fullName);
  let name = extractName(fullName);
  if (buildInfo !== undefined){
    return buildInfo.output.contracts[path][name].abi;
  }else {
    throw new Error(`failed to fetch ${fullName} abi`)
  }
}

export async function getBuildInfo (fullName: string, artifacts: Artifacts){
  return await artifacts.getBuildInfo(fullName);
}


