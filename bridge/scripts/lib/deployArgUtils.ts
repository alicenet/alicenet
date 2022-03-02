import fs from "fs";
import { env } from "./constants";
export interface DeployArgs {
  [key: string]: any;
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
