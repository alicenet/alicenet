import fs from "fs";
import { env } from "./lib/constants";
import { ArgData, ContractArgs, DeploymentArgs, getConstructorArgsABI, getInitializerArgsABI, parseArgsArray, writeDeploymentArgs } from "./lib/deployArgUtil";
import { getDeploymentList } from "./lib/deploymentListUtil";
import {
  getAllContracts,
  getDeployType, 
  InitData,
} from "./lib/deploymentUtil";

async function main() {
  //get an array of all contracts in the artifacts
  let contracts = await getDeploymentList();
  let deploymentArgs:DeploymentArgs = {
    constructor: {},
    initializer: {}
  };
  for (let contract of contracts) {
    //check each contract for a constructor and
    let cArgs: Array<ArgData> = await getConstructorArgsABI(contract);
    let iArgs: Array<ArgData> = await getInitializerArgsABI(contract);
    if (cArgs.length != 0) {
      deploymentArgs.constructor[contract] = cArgs;
    }
    if (iArgs.length != 0) {
      deploymentArgs.initializer[contract] = iArgs;
    }
  }
  writeDeploymentArgs(deploymentArgs)

}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
