import fs from "fs";
import { artifacts } from "hardhat";
import { env } from "./constants";
import { extractPath, extractName, ArgTemplate } from "./deploymentUtils";
export interface DeployArgs {
  [key: string]: any;
}
export interface ArgData {
  name: string;
}


export async function readDeploymentArgs() {
  //this output object allows dynamic addition of fields
  let output: DeployArgs = {};
  let rawData = fs.readFileSync(`./deployments/${env()}/deploymentArgs.json`);
  output = await JSON.parse(rawData.toString("utf8"));
  return output;
}

export async function getDeploymentConstructorArgs(fullName: string) {
  let output: Array<string> = [];
  //get the deployment args
  let deploymentArgs: DeployArgs = await readDeploymentArgs();
  let args = deploymentArgs.constructorArgs[fullName];
  if (args !== undefined) {
    for (let arg of args) {
      let name: string = Object.keys(arg)[0];
      output.push(arg[name]);
    }
  } else {
    output = [];
  }
  return output;
}

export async function getDeploymentInitializerArgs(fullName: string) {
  let output: Array<any> = [];
  //get the deployment args
  let deploymentArgs: DeployArgs = await readDeploymentArgs();
  let args = deploymentArgs.initializerArgs[fullName];
  if (args !== undefined) {
    for (let arg of args) {
      let name: string = Object.keys(arg)[0];
      output.push(arg[name]);
    }
  } else {
    output = [];
  }
  return output;
}

export async function getConstructorArgsABI(fullName: string) {
  let args: Array<ArgData> = [];
  let buildInfo: any = await artifacts.getBuildInfo(fullName);
  let path = extractPath(fullName);
  let name = extractName(fullName);
  let methods = buildInfo.output.contracts[path][name].abi;
  for (let method of methods) {
    if (method.type === "constructor") {
      for (let input of method.inputs) {
        let argData = <ArgData>{};
        argData.name = input.name;
        argData.type = input.type;
        args.push(argData);
      }
    }
  }
  return args;
}

export async function getInitializerArgsABI(fullName: string) {
  let args: Array<ArgData> = [];
  let buildInfo: any = await artifacts.getBuildInfo(fullName);
  let path = extractPath(fullName);
  let name = extractName(fullName);
  let methods = buildInfo.output.contracts[path][name].abi;
  for (let method of methods) {
    if (method.name === "initialize") {
      for (let input of method.inputs) {
        let argData: ArgData = {
          name: input.name,
        };
        args.push(argData);
      }
    }
  }
  return args;
}

export function parseArgsArray(args: ArgData[]) {
  let output: Array<ArgTemplate> = [];
  //console.log(args)
  for (let i = 0; i < args.length; i++) {
    let template = <ArgTemplate>{};
    template[args[i].name] = "UNDEFINED";
    template.type = args[i].type;
    output.push(template);
  }
  return output;
}

export async function getConstructorArgCount(fullName: string) {
  let buildInfo: any = await artifacts.getBuildInfo(fullName);
  let path = extractPath(fullName);
  let name = extractName(fullName);
  let methods = buildInfo.output.contracts[path][name].abi;
  for (let method of methods) {
    if (method.type === "constructor") {
      return method.inputs.length;
    }
  }
  return 0;
}

export async function getInitializerArgCount(fullName: string) {
  let buildInfo: any = await artifacts.getBuildInfo(fullName);
  let path = extractPath(fullName);
  let name = extractName(fullName);
  let methods = buildInfo.output.contracts[path][name].abi;
  for (let method of methods) {
    if (method.name === "initialize") {
      return method.inputs.length;
    }
  }
  return 0;
}