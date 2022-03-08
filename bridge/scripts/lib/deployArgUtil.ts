import toml from "@iarna/toml";
import fs from "fs";
import { artifacts } from "hardhat";
import { readDeploymentConfig } from "./baseConfigUtil";
import { DEPLOYMENT_CONFIG_PATH } from "./constants";
import { extractName, extractPath } from "./deploymentUtil";
export interface DeploymentArgs {
   constructor: ContractArgs;
   initializer: ContractArgs;
}

export interface ContractArgs {
  [key: string]: Array<ArgData>;
}
export interface ArgData {
  [key: string]: string;
}


// return a list of constructor inputs for each contract
export async function getDeploymentConstructorArgs(fullName: string) {
  let output: Array<string> = [];
  //get the deployment args
  let deploymentConfig:any = readDeploymentConfig();
  if(deploymentConfig !== undefined){
    let deploymentArgs: DeploymentArgs = {
      constructor: deploymentConfig.constructor,
      initializer: deploymentConfig.initializer
    };
    if(deploymentArgs.constructor[fullName] !== undefined){
      output = extractArgs(deploymentArgs.constructor[fullName]) 
    }
  } else {
    output = [];
  }
  return output;
}

export function extractArgs(input: Array<ArgData>){
  let output : Array<string> = [];
  for(let i = 0; i < input.length; i++){
    let argName = Object.keys(input[i])[0]
    let argData = input[i] 
    output.push(argData[argName])
  }
  return output
}

// return a list of initializer inputs for each contract 
export async function getDeploymentInitializerArgs(fullName: string) {
  let output: Array<string> = [];
  //get the deployment args
  let deploymentConfig:any = readDeploymentConfig();
  if(deploymentConfig !== undefined){
    let deploymentArgs: DeploymentArgs = {
      constructor: deploymentConfig.constructor,
      initializer: deploymentConfig.initializer
    };
    if(deploymentArgs.initializer[fullName] !== undefined){
      output = extractArgs(deploymentArgs.initializer[fullName]) 
    }
  } else {
    output = [];
  }
  return output;
}


export function writeDeploymentArgs(deploymentArgs: DeploymentArgs){
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
    fs.writeFileSync(DEPLOYMENT_CONFIG_PATH, data)
  }
}

export async function getConstructorArgsABI(fullName: string) {
  return await parseArgsArray(fullName, "constructor");
}

export async function getInitializerArgsABI(fullName: string) { 
  return await parseArgsArray(fullName, "initialize");
}

export async function getMethodArgCount(fullName: string, methodName: string) {
  let methods = await getContractABI(fullName);
  for (let method of methods) {
    let target = methodName === "constructor" ? method.type : method.name
    if (target === "constructor") {
      return method.inputs.length;
    }
  }
  return 0;
}

export async function getInitializerArgCount(fullName: string) {
  let methods = await getContractABI(fullName);
  for (let method of methods) {
    if (method.name === "initialize") {
      return method.inputs.length;
    }
  }
  return 0;
}

export async function parseArgsArray(fullName: string, methodName: string) {
  let args: Array<ArgData> = [];
  let methods = await getContractABI(fullName)
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

async function getContractABI(fullName: string){
  let buildInfo =  await artifacts.getBuildInfo(fullName);
  let path = extractPath(fullName);
  let name = extractName(fullName);
  if (buildInfo !== undefined){
    return buildInfo.output.contracts[path][name].abi;
  }else {
    throw new Error(`failed to fetch ${fullName} abi`)
  }
}

export async function getBuildInfo (fullName: string){
  return await artifacts.getBuildInfo(fullName);
}


